package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/solarlune/paths"
)

var ZeroLevel int32 = 0

type Positioner interface {
	Ider
	GetPosition() *api.Coordinates
	SetPosition(c *api.Coordinates)
	GetMovingTo() *paths.Path
	SetMovingTo(m *paths.Path)
}

type Alive interface {
	Positioner
	IsStunned() bool
}

type Skiller interface {
	Alive
	GetSkill(id string) (*api.Skill, bool)
	GetAttributes() *api.Attributes
}
