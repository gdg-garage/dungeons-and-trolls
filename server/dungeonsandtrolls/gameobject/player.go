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
	Character      *api.Character              `json:"character"`
	BaseAttributes *api.Attributes             `json:"-"`
	ItemAttributes *api.Attributes             `json:"-"`
	MaxStats       *api.Attributes             `json:"-"`
	Skills         map[string]*api.Skill       `json:"-"`
	IsAdmin        bool                        `json:"admin"`
}

func CreatePlayer(name string) *Player {
	p := &Player{
		Character: &api.Character{
			Name: name,
			Id:   GetNewId(),
		},
		Equipped: map[api.Item_Type]*api.Item{},
	}
	p.InitAttributes()
	p.ResetAttributes()
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
	err = MaxAllAttributes(p.MaxStats, p.ItemAttributes, true)
	if err != nil {
		return err
	}
	// TODO This operation is not "healing" using newly added base attributes (life, stamina, mana) just setting the max values.
	p.Character.Attributes.Life = currentAttributes.Life
	p.Character.Attributes.Mana = currentAttributes.Mana
	p.Character.Attributes.Stamina = currentAttributes.Stamina
	p.updateAttributesUsingEffects()
	return nil
}

func (p *Player) GetId() string {
	return p.Character.Id
}

func (p *Player) GetName() string {
	return p.Character.Name
}

func (p *Player) GetPosition() *api.Coordinates {
	return p.Position
}

func (p *Player) SetPosition(c *api.Coordinates) {
	p.Position = c
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
