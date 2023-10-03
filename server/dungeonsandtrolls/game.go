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
	TickCond        *sync.Cond                    `json:"-"`

	generatorLock sync.RWMutex

	gameStorage *storage.Storage
	userStorage *storage.Storage

	mapCache MapCache

	idToObject map[string]gameobject.Ider

	Commands map[string]*api.CommandsBatch `json:"-"`

	Respawns []*gameobject.Player `json:"-"`
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
		Commands:   map[string]*api.CommandsBatch{},
		idToObject: map[string]gameobject.Ider{},
		Score:      0,
		TickCond:   sync.NewCond(&sync.Mutex{}),
	}

	return g
}

func CreateGame() (*Game, error) {
	g := NewGame()

	// TODO this needs to be properly thought out

	err := g.gameStorage.ReadTo(gameTickStorageKey, &g.Game.Tick)
	if err != nil {
		log.Warn().Msgf("Game tick was not loaded from the storage %v", err)
	}
	err = g.gameStorage.ReadTo(gameStorageKey, g)
	if err != nil {
		log.Warn().Msgf("Game was not loaded from the storage %v", err)
	} else {
		g.AddLevels(0, g.MaxLevelReached)
		g.handleStoredPlayers()
	}

	go g.gameLoop()

	return g, nil
}

func (g *Game) handleStoredPlayers() {
	for key, p := range g.ApiKeyToPlayer {
		if p.IsAdmin {
			// Do not show admin users in the game
			continue
		}
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
		g.GameLock.Lock()
		g.Game.Events = []*api.Event{}

		for _, r := range g.Respawns {
			g.Respawn(r, true)
		}
		g.Respawns = []*gameobject.Player{}

		g.processCommands()

		g.TickCond.L.Lock()
		g.Game.Tick++
		g.TickCond.L.Unlock()
		g.TickCond.Broadcast()

		// Copy score - for storage reasons
		// TODO maybe use the same solution as for tick or find something more elegant
		g.Game.Score = g.Score

		// mark active levels
		for _, p := range g.Players {
			lc, err := g.GetCachedLevel(p.GetPosition().Level)
			if err != nil {
				log.Warn().Msgf("Getting level cache for %d failed", p.GetPosition().Level)
			} else {
				lc.LastInteractedTick = g.Game.Tick
			}
		}

		// regenerate levels
		for l, lc := range g.mapCache.Level {
			if IsMapDeprecated(lc, g.Game.Tick, l) {
				// map garbage collection
				log.Info().Msgf("Garbage collecting level %d", l)
				for _, i := range lc.Objects {
					for _, o := range i {
						for _, p := range o.Players {
							if l != 0 {
								log.Warn().Msgf("Player %s (%s) is on a dead level (%d) for some reason, respawning", p.GetId(), p.GetName(), l)
								// TODO this is dumb
								for _, gop := range g.Players {
									if gop.GetId() != p.GetId() {
										g.Respawn(gop, false)
									}
								}
							}
						}
						for _, j := range o.Items {
							g.Unregister(j)
						}
						for _, j := range o.Monsters {
							g.Unregister(j)
						}

					}
				}
				g.mapCache.ClearLevelCache(l)
				// TODO check all valid
				// - go through objects and remove empty ones
				// - sort by position
				// - update the cache
				// - unregister IDs
				for i, lvl := range g.Game.Map.Levels {
					if lvl.Level == l {
						g.Game.Map.Levels[i] = g.Game.Map.Levels[len(g.Game.Map.Levels)-1]
						g.Game.Map.Levels = g.Game.Map.Levels[:len(g.Game.Map.Levels)-1]
						break
					}
				}
				if l == 0 {
					g.AddLevel(0)
					// respawn players on level 0
					for _, p := range g.Players {
						if p.GetPosition().Level == 0 {
							g.Respawn(p, false)
						}
					}
				}
			}
		}
		g.SortMaps()
		g.GameLock.Unlock()
		g.storeGameState()

		// TODO generate new levels (based on all the skips)
		//log.Debug().Msgf("sleeping for %v", LoopTime-time.Since(startTime))
		time.Sleep(LoopTime - time.Since(startTime))
	}
}

func IsMapDeprecated(mm *LevelCache, t int32, l int32) bool {
	if l == 0 {
		if (t - mm.GeneratedTick) > 30 {
			log.Info().Msgf("first level is deprecated")
			return true
		}
	} else {
		if (t - mm.GeneratedTick) > 4*60 {
			log.Info().Msgf("%d level is deprecated due to age", l)
			return true
		}
		if (t - mm.LastInteractedTick) > 60 {
			log.Info().Msgf("%d level is deprecated due to inactivity", l)
			return true
		}
	}
	return false
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
	g.SortMaps()
}

func (g *Game) SortMaps() {
	sort.Slice(g.Game.Map.Levels, func(i, j int) bool {
		return g.Game.Map.Levels[i].Level < g.Game.Map.Levels[j].Level
	})
}

func (g *Game) AddLevel(level int32) {
	g.AddLevels(level, level)
}

func (g *Game) MarkVisitedLevel(level int32) {
	g.MaxLevelReached = utils.Max(g.MaxLevelReached, level)
}

func (g *Game) Respawn(player *gameobject.Player, markDeath bool) {
	// TODO mark death if appropriate
	if markDeath {
		deathEvent := api.Event_DEATH
		g.LogEvent(&api.Event{
			Type:        &deathEvent,
			Message:     fmt.Sprintf("%s (%s) died", player.GetId(), player.Character.Name),
			Coordinates: player.GetPosition(),
		})
	}

	if player.GetPosition() != nil {
		o, err := g.GetObjectsOnPosition(player.GetPosition())
		if err != nil {
			log.Warn().Err(err).Msg("")
		} else if o != nil {
			RemovePlayerFromTile(o, player)
		}
	}
	g.SpawnPlayer(player, gameobject.ZeroLevel)
	player.InitAttributes()
	player.UpdateAttributes()
	player.Character.Money = g.GetMoney()
	player.Character.SkillPoints = float32(g.MaxLevelReached)
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
	g.Game.Events = append(g.Game.Events, event)
	log.Info().Msgf(event.String())
}

func (g *Game) GetMapObjectsOrCreateDefault(c *api.Coordinates) *api.MapObjects {
	lc, err := g.mapCache.CachedLevel(c.Level)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}
	return lc.CacheObjectsOnPosition(c, &api.MapObjects{
		Position: gameobject.CoordinatesToPosition(c),
		IsFree:   true,
	})
}

func (g *Game) processCommands() {
	errorEvent := api.Event_ERROR
	deathEvent := api.Event_DEATH
	scoreEvent := api.Event_SCORE

	for _, i := range g.idToObject {
		switch c := i.(type) {
		case *gameobject.Monster:
			if c.Stun.IsStunned {
				c.Stun.IsStunned = false
				c.Stun.IsImmune = true
			}
			if c.Stun.IsImmune {
				c.Stun.IsImmune = false
			}
		case *gameobject.Player:
			if c.Stun.IsStunned {
				c.Stun.IsStunned = false
				c.Stun.IsImmune = true
			}
			if c.Stun.IsImmune {
				c.Stun.IsImmune = false
			}
		}
	}

	for pId, c := range g.Commands {
		maybePlayer, err := g.GetObjectById(pId)
		if err != nil {
			log.Warn().Err(err).Msg("")
			continue
		}
		skiller, ok := maybePlayer.(gameobject.Skiller)
		if !ok {
			log.Warn().Err(err).Msg("object retrieved by ID is not a skiller")
			continue
		}

		if c.Yell != nil {
			err = ExecuteYell(g, skiller, c.Yell)
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to yell: %s", skiller.GetId(), skiller.GetName(), err.Error()),
					Coordinates: skiller.GetPosition(),
				})
			}
		}

		if c.Skill != nil {
			skiller.SetMovingTo(nil)
			err = ExecuteSkill(g, skiller, c.Skill)
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to use skill: %s", skiller.GetId(), skiller.GetName(), err.Error()),
					Coordinates: skiller.GetPosition(),
				})
			}
		}

		p, ok := maybePlayer.(*gameobject.Player)
		if !ok {
			//log.Warn().Err(err).Msg("object retrieved by ID is not a player")
			continue
		}

		if c.AssignSkillPoints != nil {
			err = ExecuteAssignSkillPoints(p, c.AssignSkillPoints)
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to assign skill point %s", p.GetId(), p.GetName(), err.Error()),
					Coordinates: p.GetPosition(),
				})
			}
		}

		if c.PickUp != nil {
			err = ExecutePickUp(g, p, c.PickUp)
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to pick up %s: %s", p.GetId(), p.GetName(), c.PickUp, err.Error()),
					Coordinates: p.GetPosition(),
				})
			}
		}

		if c.Buy != nil {
			err = ExecuteBuy(g, p, c.Buy)
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("%s (%s): failed to buy: %s", p.GetId(), p.GetName(), err.Error()),
					Coordinates: p.GetPosition(),
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
		//log.Info().Msgf("player is at (%d, %d), moving to (%d, %d)", p.Positioner.PositionX, p.Positioner.PositionY, p.MovingTo.Current().X, p.MovingTo.Current().Y)
		g.MoveCharacter(p, &api.Coordinates{
			PositionX: int32(p.MovingTo.Current().X),
			PositionY: int32(p.MovingTo.Current().Y),
			Level:     int32(p.GetPosition().Level),
		})
		// TODO log errors
		if p.MovingTo.AtEnd() {
			p.MovingTo = nil
		}
		// check stairs
		o, err := g.GetObjectsOnPosition(p.GetPosition())
		if err != nil {
			log.Warn().Err(err).Msg("")
			continue
		}
		if o.IsStairs {
			// spawn in the next level.
			g.SpawnPlayer(p, p.GetPosition().Level+1)
			// cancel currently invalid path
			p.MovingTo = nil
			// TODO log level traverse stats
			// TODO log newly discovered levels
		}
		if o.Portal != nil {
			// spawn in the next level.
			g.SpawnPlayer(p, o.Portal.DestinationFloor)
			// cancel currently invalid path
			p.MovingTo = nil
			// TODO log level traverse stats
			// TODO log newly discovered levels

		}
	}

	// move monsters based on move to
	for _, o := range g.idToObject {
		switch m := o.(type) {
		case *gameobject.Monster:
			if m.MovingTo == nil {
				continue
			}
			m.MovingTo.Advance()
			g.MoveCharacter(m, &api.Coordinates{
				PositionX: int32(m.MovingTo.Current().X),
				PositionY: int32(m.MovingTo.Current().Y),
				Level:     int32(m.GetPosition().Level),
			})
			// TODO log errors
			if m.MovingTo.AtEnd() {
				m.MovingTo = nil
			}
		default:
			continue
		}
	}

	// TODO passive skills

	// TODO ground effects - pass to players and monsters

	for _, i := range g.idToObject {
		switch c := i.(type) {
		case *gameobject.Monster:
			e, err := EvaluateEffects(g, c.Monster.Effects, c.Monster.Attributes, c)
			*c.Monster.LastDamageTaken += 1
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("failed to evaluate effects for monster %s: %s", c.GetId(), err.Error()),
					Coordinates: c.Position,
				})
			} else {
				c.Monster.Effects = e
			}
		case *gameobject.Player:
			e, err := EvaluateEffects(g, c.Character.Effects, c.Character.Attributes, c)
			c.Character.LastDamageTaken += 1
			if err != nil {
				g.LogEvent(&api.Event{
					Type:        &errorEvent,
					Message:     fmt.Sprintf("failed to evaluate effects for player %s: %s", c.GetId(), err.Error()),
					Coordinates: c.GetPosition(),
				})
			} else {
				c.Character.Effects = e
			}
		}
	}

	// Kill what is dead
	// TODO solve kill stats (all players who interacted)
	for _, i := range g.idToObject {
		switch c := i.(type) {
		case *gameobject.Monster:
			// move kill count
			if c.KillCounter != nil {
				*c.KillCounter -= 1
			}
			if (c.Monster.Attributes.Life != nil && *c.Monster.Attributes.Life <= 0) || (c.KillCounter != nil && *c.KillCounter <= 0) {
				o, err := g.GetObjectsOnPosition(c.GetPosition())
				if err != nil {
					g.LogEvent(&api.Event{
						Type:        &errorEvent,
						Message:     fmt.Sprintf("failed to evaluate effects for monster %s: %s", c.GetId(), err.Error()),
						Coordinates: c.Position,
					})
				} else {
					g.LogEvent(&api.Event{
						Type:        &deathEvent,
						Message:     fmt.Sprintf("monster %s (%s) died", c.GetId(), c.Monster.Name),
						Coordinates: c.Position,
					})
					if c.Monster.Score != nil {
						g.Score += *c.Monster.Score
						g.LogEvent(&api.Event{
							Type:        &scoreEvent,
							Message:     fmt.Sprintf("monster %s (%s) provided %f score", c.GetId(), c.Monster.Name, *c.Monster.Score),
							Coordinates: c.Position,
						})
					}
					RemoveMonsterFromTile(o, c)
				}
				po := g.GetMapObjectsOrCreateDefault(c.GetPosition())
				for _, d := range c.Monster.OnDeath {
					switch o := d.Data.(type) {
					case *api.Droppable_Skill:
						// TODO
					case *api.Droppable_Item:
						o.Item.Id = gameobject.GetNewId()
						po.Items = append(po.Items, o.Item)
					case *api.Droppable_Monster:
						o.Monster.Id = gameobject.GetNewId()
						po.Monsters = append(po.Monsters, o.Monster)
						g.Register(gameobject.CreateMonster(o.Monster, c.GetPosition()))
					case *api.Droppable_Decoration:
						po.Decorations = append(po.Decorations, o.Decoration)
					case *api.Droppable_Waypoint:
						po.Portal = o.Waypoint
					case *api.Droppable_Key:
						for _, door := range o.Key.Doors {
							lc, err := g.mapCache.CachedLevel(c.GetPosition().Level)
							if err != nil {
								log.Warn().Err(err).Msg("")
							}
							// find door and remove it
							dp := lc.CacheObjectsOnPosition(gameobject.PositionToCoordinates(door, c.GetPosition().Level), nil)
							if dp != nil {
								dp.IsDoor = false
								dp.IsFree = !dp.IsWall
								gr := lc.Grid.Get(int(door.PositionX), int(door.PositionY))
								gr.Walkable = dp.IsFree
							}
						}
					}
				}
				g.Unregister(c)
			}
		case *gameobject.Player:
			if c.Character.Attributes.Life != nil && *c.Character.Attributes.Life <= 0 {
				g.Respawn(c, true)
			}
		default:
			continue
		}
	}

	// resolve teleports and knockbacks
	for _, i := range g.idToObject {
		switch c := i.(type) {
		case gameobject.Skiller:
			// TODO add knockback
			if c.GetTeleportTo().Move != nil {
				log.Info().Msgf("moving %s (%s) based on casts %+v", c.GetId(), c.GetName(), c.GetTeleportTo().Move)
				g.ForceMoveCharacter(c, c.GetTeleportTo().Move)
			}
		}
	}

	g.Commands = map[string]*api.CommandsBatch{}
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

	c := lc.SpawnPoint
	err = g.MoveCharacter(p, c)
	if err != nil {
		log.Warn().Err(err).Msg("")
	}
}

// Todo more generic version?
func (g *Game) removePlayerFromPosition(p *gameobject.Player) {
	o, err := g.GetObjectsOnPosition(p.GetPosition())
	if err != nil {
		// maybe destroyed level
		log.Warn().Err(err).Msg("")
	}
	if o != nil {
		RemovePlayerFromTile(o, p)
	}
}

func (g *Game) removeMonsterFromPosition(m *gameobject.Monster) {
	o, err := g.GetObjectsOnPosition(m.Position)
	if err != nil {
		// maybe destroyed level
		log.Warn().Err(err).Msg("")
	}
	if o != nil {
		RemoveMonsterFromTile(o, m)
	}
}

func (g *Game) findLevel(lvl int32) *api.Level {
	for _, l := range g.Game.Map.Levels {
		if l.Level == lvl {
			return l
		}
	}
	return nil
}

// Todo more generic version?
func (g *Game) addPlayerToNewPosition(o *api.MapObjects, p *api.Character, c *api.Coordinates, lc *LevelCache) {
	if o != nil {
		o.Players = append(o.Players, p)
	} else {
		mo := &api.MapObjects{
			Position: gameobject.CoordinatesToPosition(c),
			Players: []*api.Character{
				p,
			},
			IsFree: true,
		}
		l := g.findLevel(c.Level)
		if l != nil {
			l.Objects = append(l.Objects, mo)
			lc.CacheObjectsOnPosition(c, mo)
		} else {
			log.Warn().Msgf("player moved to level %d which does not exist", c.Level)
		}

	}
}

func (g *Game) addMonsterToNewPosition(o *api.MapObjects, m *api.Monster, c *api.Coordinates, lc *LevelCache) {
	if o != nil {
		o.Monsters = append(o.Monsters, m)
	} else {
		mo := &api.MapObjects{
			Position: gameobject.CoordinatesToPosition(c),
			Monsters: []*api.Monster{
				m,
			},
			IsFree: true,
		}
		l := g.findLevel(c.Level)
		if l != nil {
			l.Objects = append(l.Objects, mo)
			lc.CacheObjectsOnPosition(c, mo)
		} else {
			log.Warn().Msgf("player moved to level %d which does not exist", c.Level)
		}
	}
}

// MoveCharacter The coordinates must include level.
func (g *Game) MoveCharacter(p gameobject.Positioner, c *api.Coordinates) error {
	equipEvent := api.Event_MOVE
	if p.GetPosition() != nil {
		switch pt := p.(type) {
		case *gameobject.Player:
			// remove player from the previous position
			g.removePlayerFromPosition(pt)
		case *gameobject.Monster:
			g.removeMonsterFromPosition(pt)
		}
		g.LogEvent(&api.Event{
			Type: &equipEvent,
			Message: fmt.Sprintf("Character %s (%s) moved from (%d, %d) to (%d, %d)",
				p.GetId(), p.GetName(), p.GetPosition().PositionX, p.GetPosition().PositionY, c.PositionX, c.PositionY),
			Coordinates: p.GetPosition()})
	}
	lc, err := g.mapCache.CachedLevel(c.Level)
	if err != nil {
		log.Warn().Err(err).Msg("")
	} else {
		o := lc.CacheObjectsOnPosition(c, nil)
		switch pt := p.(type) {
		case *gameobject.Player:
			g.addPlayerToNewPosition(o, pt.Character, c, lc)
		case *gameobject.Monster:
			g.addMonsterToNewPosition(o, pt.Monster, c, lc)
		}

	}
	p.SetPosition(c)
	return nil
}

func (g *Game) ForceMoveCharacter(p gameobject.Positioner, c *api.Coordinates) error {
	err := g.MoveCharacter(p, c)
	if err != nil {
		return err
	}
	// cancel movement
	p.SetMovingTo(nil)
	return nil
}

func (g *Game) GetCachedLevel(level int32) (*LevelCache, error) {
	return g.mapCache.CachedLevel(level)
}

func (g *Game) GetObjectsOnPosition(c *api.Coordinates) (*api.MapObjects, error) {
	lc, err := g.mapCache.CachedLevel(c.Level)
	if err != nil {
		return nil, err
	}
	return lc.CacheObjectsOnPosition(c, nil), nil
}

func (g *Game) GetCurrentPlayer(token string) (*gameobject.Player, error) {
	return g.GetPlayerByKey(token)
}

func (g *Game) GetCommands(pId string) *api.CommandsBatch {
	if pc, ok := g.Commands[pId]; ok {
		return pc
	}
	g.Commands[pId] = &api.CommandsBatch{}
	return g.Commands[pId]
}

func (g *Game) Register(o gameobject.Ider) {
	// TODO lock
	g.idToObject[o.GetId()] = o
}

func (g *Game) Unregister(o gameobject.Ider) {
	delete(g.idToObject, o.GetId())
}

func (g *Game) GetObjectById(id string) (gameobject.Ider, error) {
	if o, ok := g.idToObject[id]; ok {
		return o, nil
	}
	return nil, fmt.Errorf("object with id %s not found", id)
}

func (g *Game) WaitForNextTick(tick int32) {
	g.TickCond.L.Lock()
	defer g.TickCond.L.Unlock()
	for g.Game.Tick <= tick {
		g.TickCond.Wait()
	}
}

func HideNonPublicMonsterFields(g *Game, m *api.Monster) {
	// Propagate partial info
	for _, i := range m.EquippedItems {
		m.Items = append(m.Items, &api.SimpleItem{
			Name: i.Name,
			Slot: i.Slot,
			Icon: i.Icon,
		})
	}
	for _, e := range m.Effects {
		gameobject.FilterEffect(e)
	}

	m.LifePercentage = float32(math.Round(float64(*m.Attributes.Life) / float64(*m.MaxAttributes.Life) * 100))

	// Hide the rest
	m.EquippedItems = []*api.Item{}
	m.Score = nil
	m.Algorithm = nil
	m.OnDeath = []*api.Droppable{}
	m.Attributes = nil
	m.MaxAttributes = nil
	m.LastDamageTaken = nil
}

func RemovePlayerFromTile(o *api.MapObjects, p *gameobject.Player) {
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

func RemoveMonsterFromTile(o *api.MapObjects, m *gameobject.Monster) {
	for pi, pd := range o.Monsters {
		if pd.Id == m.GetId() {
			// move last element to removed position
			o.Monsters[pi] = o.Monsters[len(o.Monsters)-1]
			// shorten the slice
			o.Monsters = o.Monsters[:len(o.Monsters)-1]
			break
		}
	}
}
