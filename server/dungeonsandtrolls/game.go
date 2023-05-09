package dungeonsandtrolls

import (
	"dungeons-and-trolls/dungeonsandtrolls/gameobject"
	"time"
)

const LoopTime = time.Second

type Game struct {
	Map     *Map                  `json:"map"`
	Inputs  map[string][]CommandI `json:"-"`
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
