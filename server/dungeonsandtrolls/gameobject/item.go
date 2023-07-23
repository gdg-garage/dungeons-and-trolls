package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

type Id interface {
	GetId() string
}

type Item struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func CreateItem(name string) *Item {
	return &Item{
		Name: name,
		Id:   GetNewId(),
	}
}

func (i *Item) GetId() string {
	return i.Id
}

func CreateWeapon(name string, minDamage, maxDamage int32) *api.Item {
	return &api.Item{
		//Item:   CreateItem(name),
		Type: api.Item_WEAPON,
		Id:   GetNewId(),
		Name: name,
		Skills: []*api.Skill{
			{
				Id:          GetNewId(),
				Name:        "attack",
				MaxDistance: 1,
				Cost: &api.Stats{
					Stamina: 5,
				},
				Effects: []*api.Effect{
					{
						MinDamage: &api.Elements{
							Physical: minDamage,
						},
						MaxDamage: &api.Elements{
							Physical: maxDamage,
						},
					},
				},
			},
		},
	}
}
