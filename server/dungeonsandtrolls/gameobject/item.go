package gameobject

import "github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"

func CreateWeapon(name string, minDamage, maxDamage int32) *api.Item {
	return &api.Item{
		//Item:   CreateItem(name),
		Name:  name,
		Type:  api.Item_WEAPON,
		Id:    GetNewId(),
		Price: 1,
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
