package dungeonsandtrolls

import (
	"errors"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"google.golang.org/protobuf/types/known/structpb"
	"time"
)

const LoopTime = time.Second

type Game struct {
	Map    *ObsoleteMap          `json:"map"`
	Inputs map[string][]CommandI `json:"-"` // TODO deprecated
	// Gained after kill (may be used in the next run)
	Experience float32 `json:"-"`
	// Gained after kill (may be used in the next run)
	Money          float32                       `json:"-"`
	Players        map[string]*gameobject.Player `json:"-"`
	IdToName       map[string]string             `json:"-"`
	IdToItem       map[string]*api.Item          `json:"-"`
	ApiKeyToPlayer map[string]*gameobject.Player `json:"-"`
	Game           api.GameState
}

func NewGame() *Game {
	return &Game{
		Inputs:         map[string][]CommandI{},
		Players:        map[string]*gameobject.Player{},
		IdToName:       map[string]string{},
		ApiKeyToPlayer: map[string]*gameobject.Player{},
		IdToItem:       map[string]*api.Item{},
		Game: api.GameState{
			Map: &api.Map{},
		},
	}
}

func CreateGame() (*Game, error) {
	g := NewGame()
	m, err := CreateMap()
	if err != nil {
		return nil, err
	}
	g.Map = m

	// place player
	playerName := "player 1"
	testKey := "test"
	p := gameobject.CreatePlayer(playerName)
	g.AddPlayer(p, &api.Registration{ApiKey: &testKey})
	(*g.Map)[0][4][4].SetChildren(append((*g.Map)[0][4][4].GetChildren(), p))
	var levelZero int32 = 0
	p.Position = api.Coordinates{Level: &levelZero, PositionX: 4, PositionY: 4}
	g.Game.Map.Levels = []*api.Level{
		{
			Level: 0,
			Free:  []*structpb.ListValue{{Values: []*structpb.Value{{Kind: &structpb.Value_NumberValue{NumberValue: 1}}}}},
			Objects: &api.MapObjects{
				Position: &api.Coordinates{
					PositionY: 4,
					PositionX: 4,
				},
				Players: []*api.Character{&p.Character},
			},
		},
	}

	// Create some items
	g.AddItem(gameobject.CreateWeapon("axe", 12, 42))
	g.AddItem(gameobject.CreateWeapon("sword", 11, 20))

	go g.gameLoop()

	return g, nil
}

func (g *Game) gameLoop() {
	for {
		startTime := time.Now()
		g.processCommands()
		g.Game.Tick++
		time.Sleep(LoopTime - (time.Now().Sub(startTime)))
	}
}

func (g *Game) AddPlayer(player *gameobject.Player, registration *api.Registration) {
	g.Players[player.Character.Name] = player
	g.IdToName[player.Character.Id] = player.GetId()
	g.ApiKeyToPlayer[*registration.ApiKey] = player
}

func (g *Game) AddItem(item *api.Item) {
	g.Game.Items = append(g.Game.Items, item)
	g.IdToItem[item.Id] = item
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
