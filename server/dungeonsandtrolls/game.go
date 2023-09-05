package dungeonsandtrolls

import (
	"errors"
	"fmt"
	"path/filepath"
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
	MaxLevelReached int                           `json:"max_reached_level"`
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
	}

	return g
}

func CreateGame() (*Game, error) {
	g := NewGame()

	m, err := ParseMap(g.generateLevels(0, 1))
	if err != nil {
		log.Fatal().Err(err).Msg("Parsing map failed")
	}
	// TODO register all ids
	err = LevelsPostProcessing(g, m, &g.mapCache)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}

	g.Game.Map = m

	// TODO this needs to be properly thought out

	err = g.gameStorage.ReadTo(gameStorageKey, g)
	if err != nil {
		log.Warn().Msgf("Game was not loaded from the storage %v", err)
	} else {
		g.handleStoredPlayers()
	}
	err = g.gameStorage.ReadTo(gameTickStorageKey, &g.Game.Tick)
	if err != nil {
		log.Warn().Msgf("Game tick was not loaded from the storage %v", err)
	}

	// place test player
	playerName := "player 1"
	var testPlayerFound bool
	if _, testPlayerFound = g.Players[playerName]; !testPlayerFound {
		testKey := "test"
		p := gameobject.CreatePlayer(playerName)
		g.AddPlayer(p, &api.Registration{ApiKey: &testKey})
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

func (g *Game) generateLevels(start int, end int) string {
	g.generatorLock.Lock()
	defer g.generatorLock.Unlock()
	return generator.Generate_level(start, end, g.MaxLevelReached)
}

func (g *Game) gameLoop() {
	for {
		startTime := time.Now()
		g.processCommands()
		// TODO map garbage collection
		// - go through objects and remove empty ones
		// - sort by postion
		// - update the cache
		// - unregister IDs
		g.Game.Tick++
		g.storeGameState()
		g.Game.Events = []*api.Event{}
		time.Sleep(LoopTime - (time.Since(startTime)))
	}
}

func (g *Game) MarkVisitedLevel(level int) {
	// TODO maybe this lock will be more global.
	g.GameLock.Lock()
	defer g.GameLock.Unlock()
	g.MaxLevelReached = utils.Max(g.MaxLevelReached, level)
}

func (g *Game) AddPlayer(player *gameobject.Player, registration *api.Registration) {
	g.Players[player.Character.Name] = player
	g.ApiKeyToPlayer[*registration.ApiKey] = player
	g.SpawnPlayer(player)

	g.Register(player)
}

func (g *Game) AddItem(item *api.Item) {
	g.Game.Items = append(g.Game.Items, item)

	// This shuld imho fail because items does not implement id
	g.Register(item)
}

func (g *Game) LogEvent(event *api.Event) {
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
		p.MovingTo = c.Move
	}

	// move players based on move to
	for _, p := range g.Players {
		if p.MovingTo == nil {
			continue
		}
		// TODO this is teleporting, we need to figure out path and plan steps
		g.MovePlayer(p, p.MovingTo)
		if p.MovingTo == p.Position {
			p.MovingTo = nil
		}
	}

	// TODO handle stairs
	// - mark visited level

	g.playerCommands = map[string]*api.CommandsBatch{}
}

func (g *Game) GetPlayerByKey(apiKey string) (*gameobject.Player, error) {
	player, ok := g.ApiKeyToPlayer[apiKey]
	if !ok {
		return nil, errors.New("API key is not valid")
	}
	return player, nil
}

func (g *Game) GetMoney() float32 {
	//  TODO edit this formula
	return g.Score * 4.2
}

func (g *Game) SpawnPlayer(p *gameobject.Player) {
	lc, err := g.mapCache.CachedLevel(gameobject.ZeroLevel)
	if err != nil {
		log.Warn().Err(err).Msg("")
	} else {
		c := proto.Clone(lc.SpawnPoint).(*api.Coordinates)
		groundLevel := gameobject.ZeroLevel
		c.Level = &groundLevel
		err := g.MovePlayer(p, c)
		if err != nil {
			log.Warn().Err(err).Msg("")
		}
	}
}

// The coordinates must include level
func (g *Game) MovePlayer(p *gameobject.Player, c *api.Coordinates) error {
	// TODO log move event
	// we can assume the same level if not provided
	if c.Level == nil {
		c.Level = p.Position.Level
	}
	if p.Position != nil {
		// remove player from the previous position
		lc, err := g.mapCache.CachedLevel(*p.Position.Level)
		if err != nil {
			// maybe destroyed level
			log.Warn().Err(err).Msg("")
		} else {
			o := lc.CacheObjectsOnPosition(p.Position, nil)
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
			}
			g.Game.Map.Levels[*c.Level].Objects = append(g.Game.Map.Levels[*c.Level].Objects, mo)
			lc.CacheObjectsOnPosition(c, mo)
		}
	}
	return nil
}

func (g *Game) GetCurrentPlayer() (*gameobject.Player, error) {
	// TODO implement this
	return g.GetPlayerByKey("test")
}

func (g *Game) GetPlayerCommands(pId string) *api.CommandsBatch {
	if pc, ok := g.playerCommands[pId]; ok {
		return pc
	}
	g.playerCommands[pId] = &api.CommandsBatch{}
	return g.playerCommands[pId]
}

func (g *Game) Register(o gameobject.Id) {
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
