package gameobject

type Position struct {
	Level, X, Y int
}

type Player struct {
	GameObject `json:",inline"`
	Name       string   `json:"name"`
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
