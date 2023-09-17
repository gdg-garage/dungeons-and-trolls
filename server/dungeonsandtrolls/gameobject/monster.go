package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/solarlune/paths"
)

// TODO this should be same class as player (or use same parent)

type Monster struct {
	Position *api.Coordinates `json:"position"`
	MovingTo *paths.Path      `json:"-"`
	Monster  *api.Monster     `json:"-"`
}

func CreateMonster(m *api.Monster, p *api.Coordinates) *Monster {
	return &Monster{
		Position: p,
		Monster:  m,
	}
}

func (m *Monster) GetId() string {
	return m.Monster.Id
}
