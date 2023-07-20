package dungeonsandtrolls

import (
	"time"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
)

const LoopTime = time.Second

func CreateWeapon(name string, damage, weight float32) *Item {
	return &Item{
		//Item:   CreateItem(name),
		//MaxDamage: damage,
		Weight: &weight,
	}
}

type Game struct {
	Map    *ObsoleteMap          `json:"map"`
	Items  []*Item               `json:"items"`
	Inputs map[string][]CommandI `json:"-"`
	// Gained after kill (may be used in the next run)
	Experience float32 `json:"-"`
	// Gained after kill (may be used in the next run)
	Money   float32 `json:"-"`
	players map[string]*gameobject.Player
}

func CreateGame() (*Game, error) {
	g := Game{
		Inputs:  map[string][]CommandI{},
		players: map[string]*gameobject.Player{},
	}
	m, err := CreateMap()
	if err != nil {
		return nil, err
	}
	g.Map = m
	// place player
	p := gameobject.CreatePlayer("player 1")
	g.players["player 1"] = p
	(*g.Map)[0][4][4].SetChildren(append((*g.Map)[0][4][4].GetChildren(), p))
	p.Position = gameobject.Position{Level: 0, X: 4, Y: 4}

	// Create some items
	g.Items = append(g.Items, CreateWeapon("axe", 1.2, 4.2))
	g.Items = append(g.Items, CreateWeapon("sword", 1.1, 2))

	go g.gameLoop()

	return &g, nil
}

func (g *Game) gameLoop() {
	for {
		startTime := time.Now()
		g.processCommands()
		time.Sleep(LoopTime - (time.Now().Sub(startTime)))
	}
}

func (g *Game) processCommands() {
	for playerId, commands := range g.Inputs {
		p := g.players[playerId]
		for _, command := range commands {
			switch command.GetType() {
			case "Move":
				// TODO check boundaries, do not remove other children (children is probably deprecated anyway)
				(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{})
				switch command.(MoveCommand).Direction {
				case Up:
					p.Position.Y--
					(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
				case Down:
					p.Position.Y++
					(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
				case Left:
					p.Position.X--
					(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
				case Right:
					p.Position.X++
					(*g.Map)[p.Position.Level][p.Position.Y][p.Position.X].SetChildren([]gameobject.Interface{p})
				}
			}
		}
	}
	g.Inputs = make(map[string][]CommandI)
}
