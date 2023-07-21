package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

type Position struct {
	Level, X, Y int
}

type Player struct {
	GameObject `json:",inline"`
	Position   Position         `json:"-"`
	MovingTo   *api.Coordinates `json:"-"`
	Character  api.Character    `json:"-"`
}

func (p *Player) GetId() string {
	return p.Character.Id
}

func CreatePlayer(name string) *Player {
	return &Player{
		Character: api.Character{
			Name: name,
			Id:   GetNewId(),
		},
		GameObject: GameObject{
			Type: "Player",
		},
	}
}
