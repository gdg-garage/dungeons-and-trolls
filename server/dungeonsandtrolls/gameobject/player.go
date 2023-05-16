package gameobject

type Position struct {
	Level, X, Y int
}

type Player struct {
	GameObject `json:",inline"`
	Name       string   `json:"name"`
	Health     float32  `json:"health"`
	Items      []Item   `json:"items"`
	Money      float32  `json:"money"`
	Experience float32  `json:"experience"`
	Position   Position `json:"-"`
}

func CreatePlayer(name string) *Player {
	return &Player{
		Name: name,
		GameObject: GameObject{
			Type: "Player",
		},
	}
}
