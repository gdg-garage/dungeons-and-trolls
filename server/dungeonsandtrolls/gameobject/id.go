package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

type Ider interface {
	GetId() string
	GetName() string
	// TODO may contain name method too
}

type Positioner interface {
	Ider
	GetPosition() *api.Coordinates
	SetPosition(c *api.Coordinates)
}
