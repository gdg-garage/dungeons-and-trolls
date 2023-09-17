package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/solarlune/paths"
	"go.openly.dev/pointy"
	"google.golang.org/protobuf/proto"
)

const baseStat float32 = 100

type Player struct {
	Position       *api.Coordinates            `json:"position"`
	MovingTo       *paths.Path                 `json:"-"`
	Equipped       map[api.Item_Type]*api.Item `json:"-"`
	Character      api.Character               `json:"character"`
	ItemAttributes *api.Attributes             `json:"-"`
	MaxStats       *api.Attributes             `json:"-"`
	Skills         map[string]*api.Skill       `json:"-"`
	IsAdmin        bool                        `json:"admin"`
}

func CreatePlayer(name string) *Player {
	p := &Player{
		Character: api.Character{
			Name: name,
			Id:   GetNewId(),
		},
		Equipped: map[api.Item_Type]*api.Item{},
	}
	p.ResetAttributes()
	return p
}

func (p *Player) ResetAttributes() error {
	p.ItemAttributes = &api.Attributes{
		Life:    pointy.Float32(0),
		Mana:    pointy.Float32(0),
		Stamina: pointy.Float32(0),

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
		Scalar: pointy.Float32(1),
	}
	p.MaxStats = &api.Attributes{
		Life:    pointy.Float32(baseStat),
		Mana:    pointy.Float32(baseStat),
		Stamina: pointy.Float32(baseStat),
	}
	a := proto.Clone(p.ItemAttributes).(*api.Attributes)
	p.Character.Attributes = a
	return MergeAllAttributes(p.Character.Attributes, p.MaxStats, false)
}

func (p *Player) updateAttributesUsingEffects() {
	// TODO affect only p.Character.Attributes
}

func (p *Player) updateAttributes() error {
	currentAttributes := proto.Clone(p.Character.Attributes).(*api.Attributes)
	err := p.ResetAttributes()
	if err != nil {
		return err
	}
	for _, i := range p.Equipped {
		err := MergeAllAttributes(p.ItemAttributes, i.Attributes, false)
		if err != nil {
			return err
		}
	}
	err = MergeAllAttributes(p.Character.Attributes, p.ItemAttributes, false)
	if err != nil {
		return err
	}
	err = MergeAllAttributes(p.MaxStats, p.ItemAttributes, true)
	if err != nil {
		return err
	}
	// This operation is not "healing" using newly added base attributes (life, stamina, mana) just setting the max values.
	p.Character.Attributes.Life = currentAttributes.Life
	p.Character.Attributes.Mana = currentAttributes.Mana
	p.Character.Attributes.Stamina = currentAttributes.Stamina
	p.updateAttributesUsingEffects()
	return nil
}

func (p *Player) GetId() string {
	return p.Character.Id
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
	return p.updateAttributes()
}
