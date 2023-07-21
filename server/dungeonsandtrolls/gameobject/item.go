package gameobject

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

//type Weapon struct {
//	*Item  `json:",inline"`
//	Damage float32 `json:"damage"`
//	Weight float32 `json:"weight"`
//}

//func CreateWeapon(name string, damage, weight float32) *Weapon {
//	return &Weapon{
//		Item:   CreateItem(name),
//		Damage: damage,
//		Weight: weight,
//	}
//}
