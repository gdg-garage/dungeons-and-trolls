package gameobject

type Player struct {
	GameObject `json:",inline"`
	Name       string `json:"name"`
}

func CreatePlayer(name string) (player *Player) {
	return &Player{
		Name: name,
		GameObject: GameObject{
			Type: "Player",
		},
	}
}
