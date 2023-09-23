package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/google/uuid"
)

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

func GetNewId() string {
	return uuid.New().String()
}
