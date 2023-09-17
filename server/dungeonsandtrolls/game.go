package dungeonsandtrolls

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/storage"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
	"github.com/gdg-garage/dungeons-and-trolls/server/generator"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/proto"
)

const LoopTime = time.Second

const storageBasePath = "data/"
const userStorageFile = "users.json"
const gameStorageFile = "game.json"

const gameStorageKey = "game"
const gameTickStorageKey = "game_tick"

type Game struct {
	// Gained after kill (may be used in the next run)
	Score           float32                       `json:"score"`
	Players         map[string]*gameobject.Player `json:"-"`
	ApiKeyToPlayer  map[string]*gameobject.Player `json:"player_api_keys"`
	MaxLevelReached int32                         `json:"max_reached_level"`
	Game            api.GameState                 `json:"-"`
	GameLock        sync.RWMutex                  `json:"-"`

	generatorLock sync.RWMutex

	gameStorage *storage.Storage
	userStorage *storage.Storage

	mapCache MapCache

	idToObject map[string]gameobject.Id

	// todo create player cache

	playerCommands map[string]*api.CommandsBatch

	// TODO last action in level probably in map cache
	// TODO time since generated level
	// - if level is too old regen
	// TODO regen 0 level and respawn players there based on the rules described above.
}

func NewGame() *Game {
	gameStorage, err := storage.NewStorage(filepath.Join(storageBasePath, gameStorageFile))
	if err != nil {
		log.Fatal().Msgf("Game storage init failed %v", err)
	}
	userStorage, err := storage.NewStorage(filepath.Join(storageBasePath, userStorageFile))
	if err != nil {
		log.Fatal().Msgf("User storage init failed %v", err)
	}

	g := &Game{
		Players:         map[string]*gameobject.Player{},
		ApiKeyToPlayer:  map[string]*gameobject.Player{},
		gameStorage:     gameStorage,
		userStorage:     userStorage,
		MaxLevelReached: 1,
		Game: api.GameState{
			Map: &api.Map{},
		},
		mapCache: MapCache{
			Level: map[int32]*LevelCache{},
		},
		playerCommands: map[string]*api.CommandsBatch{},
		idToObject:     map[string]gameobject.Id{},
		Score:          0,
	}

	return g
}

func CreateGame() (*Game, error) {
	g := NewGame()

	// TODO this needs to be properly thought out

	err := g.gameStorage.ReadTo(gameStorageKey, g)
	if err != nil {
		log.Warn().Msgf("Game was not loaded from the storage %v", err)
	} else {
		g.AddLevels(0, g.MaxLevelReached)
		g.handleStoredPlayers()
	}
	err = g.gameStorage.ReadTo(gameTickStorageKey, &g.Game.Tick)
	if err != nil {
		log.Warn().Msgf("Game tick was not loaded from the storage %v", err)
	}

	go g.gameLoop()

	return g, nil
}

func (g *Game) handleStoredPlayers() {
	for key, p := range g.ApiKeyToPlayer {
		g.AddPlayer(p, &api.Registration{ApiKey: &key})
	}
}

func (g *Game) storeGameState() {
	g.gameStorage.Write(gameStorageKey, g)
	g.gameStorage.Write(gameTickStorageKey, g.Game.Tick)
}

func (g *Game) generateLevels(start int32, end int32) string {
	startGen := time.Now()
	defer func(start time.Time) { log.Info().Msgf("Map generation took %s", time.Since(start)) }(startGen)
	g.generatorLock.Lock()
	defer g.generatorLock.Unlock()
	return generator.GenerateLevel(start, end, g.MaxLevelReached)
}

func (g *Game) gameLoop() {
	for {
		startTime := time.Now()
		// TODO lock
		g.Game.Events = []*api.Event{}
		g.processCommands()
		// TODO map garbage collection
		// - go through objects and remove empty ones
		// - sort by position
		// - update the cache
		// - unregister IDs
		g.Game.Tick++
		// Copy score - for storage reasons
		// TODO maybe use the same solution as for tick or find something more elegant
		g.Game.Score = g.Score
		g.storeGameState()
		// TODO regenerate levels
		// TODO generate new levels (based on all the skips)
		//log.Debug().Msgf("sleeping for %v", LoopTime-time.Since(startTime))
		time.Sleep(LoopTime - time.Since(startTime))
	}
}

func (g *Game) AddLevels(start int32, end int32) {
	m, err := ParseMap(g.generateLevels(start, end))
	if err != nil {
		log.Fatal().Err(err).Msg("Parsing map failed")
	}
	err = LevelsPostProcessing(g, m, &g.mapCache)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}

	g.Game.Map.Levels = append(g.Game.Map.Levels, m.Levels...)

	sort.Slice(g.Game.Map.Levels, func(i, j int) bool {
		return g.Game.Map.Levels[i].Level < g.Game.Map.Levels[j].Level
	})
}

func (g *Game) AddLevel(level int32) {
	g.AddLevels(level, level)
}

func (g *Game) MarkVisitedLevel(level int32) {
	// TODO maybe this lock will be more global.
	g.GameLock.Lock()
	defer g.GameLock.Unlock()
	g.MaxLevelReached = utils.Max(g.MaxLevelReached, level)
}

func (g *Game) Respawn(player *gameobject.Player, markDeath bool) {
	// TODO mark death if appropriate

	g.SpawnPlayer(player, gameobject.ZeroLevel)
	player.ResetAttributes()
	player.Character.Money = g.GetMoney()
	player.Character.Equip = []*api.Item{}
	player.Equipped = map[api.Item_Type]*api.Item{}

	g.Register(player)
}

func (g *Game) AddPlayer(player *gameobject.Player, registration *api.Registration) {
	g.Players[player.Character.Name] = player
	g.ApiKeyToPlayer[*registration.ApiKey] = player
	g.Respawn(player, false)
}

func (g *Game) AddItem(item *api.Item) {
	g.Game.ShopItems = append(g.Game.ShopItems, item)

	// This should imho fail because items does not implement id
	g.Register(item)
}

func (g *Game) LogEvent(event *api.Event) {
	g.GameLock.Lock()
	defer g.GameLock.Unlock()

	g.Game.Events = append(g.Game.Events, event)
	log.Info().Msgf(event.String())
}

func (g *Game) processCommands() {
	for pId, c := range g.playerCommands {
		maybePlayer, err := g.GetObjectById(pId)
		if err != nil {
			log.Warn().Err(err).Msg("")
			continue
		}
		p, ok := maybePlayer.(*gameobject.Player)
		if !ok {
			log.Warn().Err(err).Msg("object retrieved by ID is not a player")
			continue
		}

		if c.PickUp != nil {
			err = ExecutePickUp(g, p, c.PickUp)
			if err != nil {
				errorEvent := api.Event_ERROR
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to pick up %s: %s", p.Character.Id, p.Character.Name, c.PickUp, err.Error()),
					Coordinates: p.Position,
				})
			}
		}

		if c.Yell != nil {
			err = ExecuteYell(g, p, c.Yell)
			if err != nil {
				errorEvent := api.Event_ERROR
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to yell: %s", p.Character.Id, p.Character.Name, err.Error()),
					Coordinates: p.Position,
				})
			}
		}

		if c.Buy != nil {
			err = ExecuteBuy(g, p, c.Buy)
			if err != nil {
				errorEvent := api.Event_ERROR
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to buy: %s", p.Character.Id, p.Character.Name, err.Error()),
					Coordinates: p.Position,
				})
			}
		}

		//TODO skill on newly bought (picked up) items?
		if c.Skill != nil {
			err = ExecuteSkill(g, p, c.Skill)
			if err != nil {
				errorEvent := api.Event_ERROR
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to use skill: %s", p.Character.Id, p.Character.Name, err.Error()),
					Coordinates: p.Position,
				})
			}
		}
	}

	// move players based on move to
	for _, p := range g.Players {
		if p.MovingTo == nil {
			continue
		}
		p.MovingTo.Advance()
		//log.Info().Msgf("player is at (%d, %d), moving to (%d, %d)", p.Position.PositionX, p.Position.PositionY, p.MovingTo.Current().X, p.MovingTo.Current().Y)
		g.MovePlayer(p, &api.Coordinates{
			PositionX: int32(p.MovingTo.Current().X),
			PositionY: int32(p.MovingTo.Current().Y)})
		// TODO log errors
		if p.MovingTo.AtEnd() {
			p.MovingTo = nil
		}
		// check stairs
		o, err := g.GetObjectsOnPosition(p.Position)
		if err != nil {
			log.Warn().Err(err).Msg("")
			continue
		}
		if o.IsStairs {
			// spawn in the next level.
			g.SpawnPlayer(p, *p.Position.Level+1)
			// cancel currently invalid path
			p.MovingTo = nil
			// TODO log level traverse stats
			// TODO log newly discovered levels
		}
		if o.Portal != nil {
			// spawn in the next level.
			g.SpawnPlayer(p, *p.Position.Level+1)
			// cancel currently invalid path
			p.MovingTo = nil
			// TODO log level traverse stats
			// TODO log newly discovered levels
		}
	}

	for _, i := range g.idToObject {
		switch c := i.(type) {
		case *gameobject.Monster:
			e, err := gameobject.EvaluateEffects(c.Monster.Effects, c.Monster.Attributes)
			if err != nil {
				errorEvent := api.Event_ERROR
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("failed to evaluate effects for monster %s: %s", c.GetId(), err.Error()),
					Coordinates: c.Position,
				})
			} else {
				c.Monster.Effects = e
			}
		case *gameobject.Player:
			e, err := gameobject.EvaluateEffects(c.Character.Effects, c.Character.Attributes)
			if err != nil {
				errorEvent := api.Event_ERROR
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("failed to evaluate effects for player %s: %s", c.GetId(), err.Error()),
					Coordinates: c.Position,
				})
			} else {
				c.Character.Effects = e
			}
		default:
			// TODO items?
			continue
		}
	}

	// TODO kill what is dead

	g.playerCommands = map[string]*api.CommandsBatch{}
}

func (g *Game) GetPlayerByKey(apiKey string) (*gameobject.Player, error) {
	player, ok := g.ApiKeyToPlayer[apiKey]
	if !ok {
		return nil, errors.New("API key is not valid")
	}
	return player, nil
}

func (g *Game) GetMoney() int32 {
	//  TODO edit this formula
	return int32(math.Sqrt(float64(g.Score)) + float64(420))
}

func (g *Game) SpawnPlayer(p *gameobject.Player, level int32) {
	g.MarkVisitedLevel(level)
	lc, err := g.mapCache.CachedLevel(level)
	if err != nil {
		log.Warn().Msgf("New level %d discovered", level)
		g.AddLevel(level)
		lc, err = g.mapCache.CachedLevel(level)
		if err != nil {
			log.Warn().Err(err).Msgf("Newly generated level is missing in the cache")
		}
	}

	c := proto.Clone(lc.SpawnPoint).(*api.Coordinates)
	c.Level = &level
	err = g.MovePlayer(p, c)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}
}

func (g *Game) removePlayerFromPosition(p *gameobject.Player) {
	o, err := g.GetObjectsOnPosition(p.Position)
	if err != nil {
		// maybe destroyed level
		log.Warn().Err(err).Msg("")
	}
	if o != nil {
		for pi, pd := range o.Players {
			if pd.Id == p.Character.Id {
				// move last element to removed position
				o.Players[pi] = o.Players[len(o.Players)-1]
				// shorten the slice
				o.Players = o.Players[:len(o.Players)-1]
				break
			}
		}
	}
}

// MovePlayer The coordinates must include level.
func (g *Game) MovePlayer(p *gameobject.Player, c *api.Coordinates) error {
	// TODO log move event
	// we can assume the same level if not provided
	if c.Level == nil {
		c.Level = p.Position.Level
	}
	if p.Position != nil {
		// remove player from the previous position
		g.removePlayerFromPosition(p)
	}
	lc, err := g.mapCache.CachedLevel(*c.Level)
	if err != nil {
		log.Warn().Err(err).Msg("")
	} else {
		o := lc.CacheObjectsOnPosition(c, nil)
		if o != nil {
			o.Players = append(o.Players, &p.Character)
			p.Position = c
		} else {
			coord := proto.Clone(c).(*api.Coordinates)
			coord.Level = nil
			mo := &api.MapObjects{
				Position: coord,
				Players: []*api.Character{
					&p.Character,
				},
				IsFree: true,
			}
			g.Game.Map.Levels[*c.Level].Objects = append(g.Game.Map.Levels[*c.Level].Objects, mo)
			lc.CacheObjectsOnPosition(c, mo)
		}
	}
	p.Position = c
	return nil
}

func (g *Game) GetCachedLevel(level int32) (*LevelCache, error) {
	return g.mapCache.CachedLevel(level)
}

func (g *Game) GetObjectsOnPosition(c *api.Coordinates) (*api.MapObjects, error) {
	lc, err := g.mapCache.CachedLevel(*c.Level)
	if err != nil {
		return nil, err
	}
	return lc.CacheObjectsOnPosition(c, nil), nil
}

func (g *Game) GetCurrentPlayer(token string) (*gameobject.Player, error) {
	return g.GetPlayerByKey(token)
}

func (g *Game) GetPlayerCommands(pId string) *api.CommandsBatch {
	if pc, ok := g.playerCommands[pId]; ok {
		return pc
	}
	g.playerCommands[pId] = &api.CommandsBatch{}
	return g.playerCommands[pId]
}

func (g *Game) Register(o gameobject.Id) {
	// TODO lock
	g.idToObject[o.GetId()] = o
}

func (g *Game) Unregister(o gameobject.Id) {
	delete(g.idToObject, o.GetId())
}

func (g *Game) GetObjectById(id string) (gameobject.Id, error) {
	if o, ok := g.idToObject[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("object with id %s not found", id)
}

func HideNonPublicMonsterFields(g *Game, m *api.Monster) {
	// Propagate partial info
	for _, i := range m.EquippedItems {
		m.Items = append(m.Items, i.Name)
	}
	o, err := g.GetObjectById(m.GetId())
	if err != nil {
		log.Warn().Err(err).Msg("")
	}
	mo, ok := o.(*gameobject.Monster)
	if !ok {
		log.Warn().Msg("not a monster")
	}
	m.LifePercentage = float32(math.Round(float64(*m.Attributes.Life) / float64(*mo.MaxStats.Life) * 100))

	// Hide the rest
	m.Attributes = nil
	m.EquippedItems = []*api.Item{}
	m.Score = nil
	m.Algorithm = nil
	m.Faction = nil
	m.OnDeath = []*api.Droppable{}
}
