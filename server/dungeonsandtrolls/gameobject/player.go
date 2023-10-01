package gameobject

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/rs/zerolog/log"
	"github.com/solarlune/paths"
	"go.openly.dev/pointy"
	"google.golang.org/protobuf/proto"
)

const baseStat float32 = 100

type Player struct {
	MovingTo       *paths.Path                 `json:"-"`
	Equipped       map[api.Item_Type]*api.Item `json:"-"`
	Character      *api.Character              `json:"character"`
	BaseAttributes *api.Attributes             `json:"-"`
	ItemAttributes *api.Attributes             `json:"-"`
	MaxStats       *api.Attributes             `json:"-"`
	Skills         map[string]*api.Skill       `json:"-"`
	IsAdmin        bool                        `json:"admin"`
	Stun           Stun                        `json:"-"`
	TeleportedTo   TeleportPosition            `json:"-"`
}

func CreatePlayer(name string) *Player {
	p := &Player{
		Character: &api.Character{
			Name:            name,
			Id:              GetNewId(),
			LastDamageTaken: 10,
		},
		Equipped: map[api.Item_Type]*api.Item{},
	}
	p.InitAttributes()
	p.UpdateAttributes()
	return p
}

func (p *Player) InitAttributes() {
	p.BaseAttributes = &api.Attributes{
		Life:    pointy.Float32(baseStat),
		Mana:    pointy.Float32(baseStat),
		Stamina: pointy.Float32(baseStat),

		Strength:     pointy.Float32(0),
		Dexterity:    pointy.Float32(0),
		Intelligence: pointy.Float32(0),
		Willpower:    pointy.Float32(0),
		Constitution: pointy.Float32(0),

		SlashResist:    pointy.Float32(0),
		PierceResist:   pointy.Float32(0),
		FireResist:     pointy.Float32(0),
		PoisonResist:   pointy.Float32(0),
		ElectricResist: pointy.Float32(0),

		// necessary because scalar part in the skills would be zero or skipped
		Constant: pointy.Float32(1),
	}
	p.MaxStats = &api.Attributes{
		Life:    pointy.Float32(baseStat),
		Mana:    pointy.Float32(baseStat),
		Stamina: pointy.Float32(baseStat),
	}
	p.Character.Attributes = &api.Attributes{
		Life:    pointy.Float32(baseStat),
		Mana:    pointy.Float32(baseStat),
		Stamina: pointy.Float32(baseStat),
	}
}

func (p *Player) ResetAttributes() error {
	p.ItemAttributes = &api.Attributes{}
	//var ok bool
	//p.ItemAttributes, ok = proto.Clone(p.BaseAttributes).(*api.Attributes)
	//if !ok {
	//	return fmt.Errorf("cloning base attributes failed")
	//}

	a := proto.Clone(p.ItemAttributes).(*api.Attributes)
	p.Character.Attributes = a
	return MergeAllAttributes(p.Character.Attributes, p.BaseAttributes, false)
}

func (p *Player) updateAttributesUsingEffects() {
	for _, e := range p.Character.Effects {
		MergeAllAttributes(p.Character.Attributes, e.Effects, false)
	}
}

func (p *Player) UpdateAttributes() error {
	log.Info().Msgf("%s (%s) current attributes %+v", p.GetId(), p.GetName(), p.GetAttributes())
	currentAttributes := proto.Clone(p.Character.Attributes).(*api.Attributes)
	err := p.ResetAttributes()
	log.Info().Msgf("%s (%s) base attributes %+v", p.GetId(), p.GetName(), p.BaseAttributes)
	if err != nil {
		return err
	}
	for _, i := range p.Equipped {
		err := MergeAllAttributes(p.ItemAttributes, i.Attributes, false)
		if err != nil {
			return err
		}
	}
	log.Info().Msgf("%s (%s) item attributes %+v", p.GetId(), p.GetName(), p.ItemAttributes)
	err = MergeAllAttributes(p.GetAttributes(), p.ItemAttributes, false)
	log.Info().Msgf("%s (%s) max attributes %+v", p.GetId(), p.GetName(), p.MaxStats)
	if err != nil {
		return err
	}
	lastMax, ok := proto.Clone(p.MaxStats).(*api.Attributes)
	if !ok {
		return fmt.Errorf("cloning max stats failed")
	}
	err = MaxAllAttributes(p.MaxStats, p.GetAttributes(), true)
	if err != nil {
		return err
	}
	log.Info().Msgf("%s (%s) new max attributes %+v", p.GetId(), p.GetName(), p.MaxStats)
	added, ok := proto.Clone(p.MaxStats).(*api.Attributes)
	if !ok {
		return fmt.Errorf("cloning max stats failed")
	}
	p.Character.MaxAttributes = p.MaxStats
	SubtractAllAttributes(added, lastMax, true)
	if currentAttributes != nil {
		p.Character.Attributes.Life = currentAttributes.Life
		p.Character.Attributes.Mana = currentAttributes.Mana
		p.Character.Attributes.Stamina = currentAttributes.Stamina
	}
	*p.Character.Attributes.Life += *added.Life
	*p.Character.Attributes.Mana += *added.Mana
	*p.Character.Attributes.Stamina += *added.Stamina
	log.Info().Msgf("%s (%s) final attributes without effects %+v", p.GetId(), p.GetName(), p.GetAttributes())
	p.updateAttributesUsingEffects()
	log.Info().Msgf("%s (%s) final attributes with effects %+v", p.GetId(), p.GetName(), p.GetAttributes())
	return nil
}

func (p *Player) GetId() string {
	return p.Character.Id
}

func (p *Player) GetName() string {
	return p.Character.Name
}

func (p *Player) GetPosition() *api.Coordinates {
	return p.Character.Coordinates
}

func (p *Player) SetPosition(c *api.Coordinates) {
	p.Character.Coordinates = c
}

func (p *Player) GetMovingTo() *paths.Path {
	return p.MovingTo
}

func (p *Player) SetMovingTo(m *paths.Path) {
	p.MovingTo = m
}

func (p *Player) GetSkill(id string) (*api.Skill, bool) {
	skill, ok := p.Skills[id]
	return skill, ok
}

func (p *Player) GetAttributes() *api.Attributes {
	return p.Character.Attributes
}

func (p *Player) IsStunned() bool {
	return p.Stun.IsStunned
}

func (p *Player) GetLastDamageTaken() int32 {
	return p.Character.LastDamageTaken
}

func (p *Player) DamageTaken() {
	p.Character.LastDamageTaken = -1
}

func (p *Player) GetTeleportTo() *TeleportPosition {
	return &p.TeleportedTo
}

func (p *Player) AddEffect(e *api.Effect) {
	p.Character.Effects = append(p.Character.Effects, e)
}

func (p *Player) Stunned() {
	if !p.Stun.IsImmune {
		p.Stun.IsStunned = true
		// cancel movement
		p.SetMovingTo(nil)
	}
}

func (p *Player) generateSkills() {
	p.Skills = map[string]*api.Skill{}
	for _, i := range p.Equipped {
		for _, s := range i.Skills {
			p.Skills[s.Id] = s
		}
	}
}

func (p *Player) generateEquip() {
	p.Character.Equip = []*api.Item{}
	for _, item := range p.Equipped {
		p.Character.Equip = append(p.Character.Equip, item)
	}
	p.generateSkills()
}

func (p *Player) Equip(item *api.Item) error {
	p.Equipped[item.Slot] = item
	p.generateEquip()
	return p.UpdateAttributes()
}
