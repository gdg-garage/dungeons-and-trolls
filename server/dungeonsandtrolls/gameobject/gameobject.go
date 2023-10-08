package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
	"github.com/solarlune/paths"
)

const ZeroLevel int32 = 0

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
	DamageTaken()
}

type Skiller interface {
	Alive
	GetSkill(id string) (*api.Skill, bool)
	GetAttributes() *api.Attributes
	GetLastDamageTaken() int32
	GetTeleportTo() *TeleportPosition
	ResetTeleportTo()
	Stunned()
	Stun() *api.Stun
	AddEffect(e *api.Effect)
	GetSkills() map[string]*api.Skill
}

type TeleportPosition struct {
	Move *api.Coordinates
	// TODO add knockbacks - to have main influence
}

func TeleportMoveTo(s Skiller, c *api.Coordinates) {
	if s.GetTeleportTo().Move == nil {
		s.GetTeleportTo().Move = c
		// use one which is further
	} else if utils.ManhattanDistance(s.GetPosition().PositionX, s.GetPosition().PositionY, c.PositionX, c.PositionY) >
		utils.ManhattanDistance(s.GetTeleportTo().Move.PositionX, s.GetTeleportTo().Move.PositionY, c.PositionX, c.PositionY) {
		s.GetTeleportTo().Move = c
	}
}
