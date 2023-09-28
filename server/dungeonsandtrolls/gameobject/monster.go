package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/rs/zerolog/log"
	"github.com/solarlune/paths"
	"google.golang.org/protobuf/proto"
)

type Monster struct {
	Position *api.Coordinates      `json:"position"`
	MovingTo *paths.Path           `json:"-"`
	MaxStats *api.Attributes       `json:"-"`
	Monster  *api.Monster          `json:"-"`
	Skills   map[string]*api.Skill `json:"-"`
	Stun     Stun                  `json:"-"`
}

func CreateMonster(mon *api.Monster, p *api.Coordinates) *Monster {
	for _, i := range mon.EquippedItems {
		MergeAllAttributes(mon.Attributes, i.Attributes, true)
	}

	maxAttributes, ok := proto.Clone(mon.Attributes).(*api.Attributes)
	if !ok {
		log.Warn().Msgf("cloning monster attributes failed")
	}
	// TODO check
	m := &Monster{
		Position: p,
		Monster:  mon,
		MaxStats: maxAttributes,
	}
	m.generateSkills()
	return m
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

func (m *Monster) GetSkill(id string) (*api.Skill, bool) {
	skill, ok := m.Skills[id]
	return skill, ok
}

func (m *Monster) GetAttributes() *api.Attributes {
	return m.Monster.Attributes
}

func (m *Monster) IsStunned() bool {
	return m.Stun.IsStunned
}

func (m *Monster) generateSkills() {
	m.Skills = map[string]*api.Skill{}
	for _, i := range m.Monster.EquippedItems {
		for _, s := range i.Skills {
			m.Skills[s.Id] = s
		}
	}
}

func (m *Monster) UpdateAttributes() {
	currentAttributes := proto.Clone(m.GetAttributes()).(*api.Attributes)
	m.Monster.Attributes = proto.Clone(m.MaxStats).(*api.Attributes)
	m.GetAttributes().Life = currentAttributes.Life
	m.GetAttributes().Mana = currentAttributes.Mana
	m.GetAttributes().Stamina = currentAttributes.Stamina
	for _, e := range m.Monster.Effects {
		MergeAllAttributes(m.GetAttributes(), e.Effects, false)
	}
}
