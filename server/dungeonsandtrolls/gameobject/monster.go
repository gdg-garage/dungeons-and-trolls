package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

// TODO this should be same class as player (or use same parent)
// TODO player and monster are items on the map and players may interact with them, therfore they should have an ID.

type Monster struct {
	GameObject `json:",inline"`
	Position   api.Coordinates `json:"-"`
	MovingTo   api.Coordinates `json:"-"`
	Monster    api.Monster     `json:"-"`
}

func CreateMonster(name string) *Monster {
	return &Monster{
		Monster: api.Monster{
			Name: name,
			Id:   GetNewId(),
		},
		GameObject: GameObject{
			Type: "Monster",
		},
	}
}
