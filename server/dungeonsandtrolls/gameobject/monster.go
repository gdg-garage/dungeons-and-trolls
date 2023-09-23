package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/rs/zerolog/log"
	"github.com/solarlune/paths"
	"google.golang.org/protobuf/proto"
)

type Monster struct {
	Position *api.Coordinates `json:"position"`
	MovingTo *paths.Path      `json:"-"`
	MaxStats *api.Attributes  `json:"-"`
	Monster  *api.Monster     `json:"-"`
}

func CreateMonster(m *api.Monster, p *api.Coordinates) *Monster {
	maxAttributes, ok := proto.Clone(m.Attributes).(*api.Attributes)
	if !ok {
		log.Warn().Msgf("cloning monster attributes failed")
	}
	// TODO check
	return &Monster{
		Position: p,
		Monster:  m,
		MaxStats: maxAttributes,
	}
}

func (m *Monster) GetId() string {
	return m.Monster.Id
}

func (m *Monster) GetName() string {
	return m.Monster.Name
}

func (m *Monster) GetPosition() *api.Coordinates {
	return m.Position
}

func (m *Monster) SetPosition(c *api.Coordinates) {
	m.Position = c
}

func (m *Monster) GetMovingTo() *paths.Path {
	return m.MovingTo
}

func (m *Monster) SetMovingTo(p *paths.Path) {
	m.MovingTo = p
}
