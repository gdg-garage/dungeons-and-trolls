package dungeonsandtrolls

import (
	"errors"
	"path/filepath"
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
	Inputs map[string][]CommandI `json:"-"` // TODO deprecated
	// Gained after kill (may be used in the next run)
	Score           float32                       `json:"-"`
	Players         map[string]*gameobject.Player `json:"-"`
	IdToName        map[string]string             `json:"-"`
	IdToItem        map[string]*api.Item          `json:"-"`
	ApiKeyToPlayer  map[string]*gameobject.Player `json:"player_api_keys"`
	MaxLevelReached int                           `json:"max_reached_level"`
	Game            api.GameState                 `json:"-"`

	generatorLock sync.RWMutex
	gameLock      sync.RWMutex
	gameStorage   *storage.Storage
	userStorage   *storage.Storage
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
		Inputs:          map[string][]CommandI{},
		Players:         map[string]*gameobject.Player{},
		IdToName:        map[string]string{},
		ApiKeyToPlayer:  map[string]*gameobject.Player{},
		IdToItem:        map[string]*api.Item{},
		gameStorage:     gameStorage,
		userStorage:     userStorage,
		MaxLevelReached: 1,
		Game: api.GameState{
			Map: &api.Map{},
		},
	}

	err = gameStorage.ReadTo(gameStorageKey, g)
	if err != nil {
		log.Warn().Msgf("Game was not loaded from the storage %v", err)
	} else {
		g.handleStoredPlayers()
	}
	err = gameStorage.ReadTo(gameTickStorageKey, &g.Game.Tick)
	if err != nil {
		log.Warn().Msgf("Game tick was not loaded from the storage %v", err)
	}

	return g
}

func CreateGame() (*Game, error) {
	g := NewGame()

	// place test player
	playerName := "player 1"
	var p *gameobject.Player
	var testPlayerFound bool
	if p, testPlayerFound = g.Players[playerName]; !testPlayerFound {
		testKey := "test"
		p = gameobject.CreatePlayer(playerName)
		g.AddPlayer(p, &api.Registration{ApiKey: &testKey})
		// (*g.Map)[0][4][4].SetChildren(append((*g.Map)[0][4][4].GetChildren(), p))
		var levelZero int32 = 0
		p.Position = api.Coordinates{Level: &levelZero, PositionX: 4, PositionY: 4}
	}

	// g.Game.Map.Levels = []*api.Level{
	// 	{
	// 		Level: 0,
	// 		Free:  []*structpb.ListValue{{Values: []*structpb.Value{{Kind: &structpb.Value_NumberValue{NumberValue: 1}}}}},
	// 		Objects: []*api.MapObjects{
	// 			&api.MapObjects{
	// 				Position: &api.Coordinates{
	// 					PositionY: 4,
	// 					PositionX: 4,
	// 				},
	// 				Players: []*api.Character{&p.Character},
	// 			}},
	// 	},
	// }

	m, err := ParseMap(g.generateLevel(0, 1))
	if err != nil {
		log.Fatal().Err(err).Msg("Parsing map failed")
	}
	log.Info().Msgf("map: %v", m)

	g.Game.Map = m
	g.Game.Map.Levels[0].Objects = append(g.Game.Map.Levels[0].Objects, &api.MapObjects{
		Position: &api.Coordinates{
			PositionY: 1,
			PositionX: 1,
		},
		Players: []*api.Character{&p.Character},
	})

	// TODO place the player on spawn

	// Create some items
	g.AddItem(gameobject.CreateWeapon("axe", 12, 42))
	g.AddItem(gameobject.CreateWeapon("sword", 11, 20))

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

func (g *Game) generateLevel(start int, end int) string {
	g.generatorLock.Lock()
	defer g.generatorLock.Unlock()
	return generator.Generate_level(start, end, g.MaxLevelReached)
}

func (g *Game) gameLoop() {
	for {
		startTime := time.Now()
		g.processCommands()
		g.Game.Tick++
		g.storeGameState()
		g.Game.Events = []*api.Event{}
		time.Sleep(LoopTime - (time.Since(startTime)))
	}
}

func (g *Game) MarkVisitedLevel(level int) {
	g.gameLock.Lock()
	defer g.gameLock.Unlock()
	g.MaxLevelReached = utils.Max(g.MaxLevelReached, level)
}

func (g *Game) AddPlayer(player *gameobject.Player, registration *api.Registration) {
	g.Players[player.Character.Name] = player
	g.IdToName[player.GetId()] = player.Character.Name
	g.ApiKeyToPlayer[*registration.ApiKey] = player
}

func (g *Game) AddItem(item *api.Item) {
	g.Game.Items = append(g.Game.Items, item)
	g.IdToItem[item.Id] = item
}

func (g *Game) LogEvent(event *api.Event) {
	g.Game.Events = append(g.Game.Events, event)
	log.Info().Msgf(event.String())
}

func (g *Game) processCommands() {
	//for playerId, commands := range g.Inputs {
	//	p := g.Players[playerId]
	//	for _, command := range commands {
	//		switch command.GetType() {
	//		case "Move":
	//			// TODO check boundaries, do not remove other children (children is probably deprecated anyway)
	//			(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{})
	//			switch command.(MoveCommand).Direction {
	//			case Up:
	//				p.Position.Y--
	//				(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
	//			case Down:
	//				p.Position.Y++
	//				(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
	//			case Left:
	//				p.Position.X--
	//				(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
	//			case Right:
	//				p.Position.X++
	//				(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
	//			}
	//		}
	//	}
	//}
	//g.Inputs = make(map[string][]CommandI)
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
