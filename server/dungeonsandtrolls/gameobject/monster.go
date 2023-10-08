package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/rs/zerolog/log"
	"github.com/solarlune/paths"
	"go.openly.dev/pointy"
	"google.golang.org/protobuf/proto"
)

type Monster struct {
	Position     *api.Coordinates      `json:"position"`
	MovingTo     *paths.Path           `json:"-"`
	Monster      *api.Monster          `json:"-"`
	Skills       map[string]*api.Skill `json:"-"`
	TeleportedTo TeleportPosition      `json:"-"`
	KillCounter  *int32                `json:"-"`
}

func CreateMonster(mon *api.Monster, p *api.Coordinates) *Monster {
	mon.Attributes.Constant = pointy.Float32(1)
	for _, i := range mon.EquippedItems {
		MergeAllAttributes(mon.Attributes, i.Attributes, true)
	}

	// TODO check
	m := &Monster{
		Position: p,
		Monster:  mon,
	}
	maxAttributes, ok := proto.Clone(mon.Attributes).(*api.Attributes)
	if !ok {
		log.Warn().Msgf("cloning monster attributes failed")
	}
	m.Monster.MaxAttributes = maxAttributes
	m.Monster.LastDamageTaken = pointy.Int32(10)
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
	return m.Monster.Stun.IsStunned
}

func (m *Monster) GetLastDamageTaken() int32 {
	return *m.Monster.LastDamageTaken
}

func (m *Monster) DamageTaken() {
	m.Monster.LastDamageTaken = pointy.Int32(-1)
}

func (m *Monster) GetTeleportTo() *TeleportPosition {
	return &m.TeleportedTo
}

func (m *Monster) ResetTeleportTo() {
	m.TeleportedTo = TeleportPosition{}
}

func (m *Monster) Stunned() {
	// TODO log stun?
	if !m.Monster.Stun.IsImmune {
		m.Monster.Stun.IsStunned = true
		// cancel movement
		m.SetMovingTo(nil)
	}
}

func (m *Monster) Stun() *api.Stun {
	return m.Monster.Stun
}

func (m *Monster) AddEffect(e *api.Effect) {
	m.Monster.Effects = append(m.Monster.Effects, e)
}

func (m *Monster) generateSkills() {
	m.Skills = map[string]*api.Skill{}
	for _, i := range m.Monster.EquippedItems {
		for _, s := range i.Skills {
			m.Skills[s.Id] = s
		}
	}
}

func (m *Monster) GetSkills() map[string]*api.Skill {
	return m.Skills
}

func (m *Monster) UpdateAttributes() {
	currentAttributes := proto.Clone(m.GetAttributes()).(*api.Attributes)
	m.Monster.Attributes = proto.Clone(m.Monster.MaxAttributes).(*api.Attributes)
	m.GetAttributes().Life = currentAttributes.Life
	m.GetAttributes().Mana = currentAttributes.Mana
	m.GetAttributes().Stamina = currentAttributes.Stamina
	for _, e := range m.Monster.Effects {
		MergeAllAttributes(m.GetAttributes(), e.Effects, false)
	}
}
