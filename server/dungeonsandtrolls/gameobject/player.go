package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
)

type Player struct {
	GameObject `json:",inline"`
	Position   api.Coordinates             `json:"position"`
	MovingTo   api.Coordinates             `json:"-"`
	Equipped   map[api.Item_Type]*api.Item `json:"equipped"`
	Character  api.Character               `json:"character"`
}

func CreatePlayer(name string) *Player {
	return &Player{
		Character: api.Character{
			Name: name,
			Id:   GetNewId(),
		},
		GameObject: GameObject{
			Type: "Player",
		},
		Equipped: map[api.Item_Type]*api.Item{},
	}
}

func (p *Player) GetId() string {
	return p.Character.Id
}

func (p *Player) generateEquip() {
	p.Character.Equip = []*api.Item{}
	for _, item := range p.Equipped {
		p.Character.Equip = append(p.Character.Equip, item)
	}
}

func (p *Player) Equip(item *api.Item) {
	p.Equipped[item.Slot] = item
	p.generateEquip()
}
