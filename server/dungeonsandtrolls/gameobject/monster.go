package gameobject

// TODO this should be same class as player (or use same parent)
// TODO player and monster are items on the map and players may interact with them, therfore they should have an ID.
type Monster struct {
	GameObject `json:",inline"`
	Name       string  `json:"name"`
	Health     float32 `json:"health"`
	Items      []Item  `json:"items"`
}

func CreateMonster(name string) *Monster {
	return &Monster{
		Name: name,
		GameObject: GameObject{
			Type: "Monster",
		},
	}
}
