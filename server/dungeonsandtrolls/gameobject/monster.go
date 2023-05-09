package gameobject

type Monster struct {
	GameObject `json:",inline"`
	Name       string `json:"name"`
}

func CreateMonster(name string) *Monster {
	return &Monster{
		Name: name,
		GameObject: GameObject{
			Type: "Monster",
		},
	}
}
