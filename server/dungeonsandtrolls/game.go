package dungeonsandtrolls

import "dungeons-and-trolls/dungeonsandtrolls/gameobject"

type Game struct {
	Map *Map `json:"map"`
}

func CreateGame() (*Game, error) {
	g := Game{}
	m, err := CreateMap()
	if err != nil {
		return nil, err
	}
	g.Map = m
	// place player
	(*g.Map)[0][4][4].Children = append((*g.Map)[0][4][4].Children, gameobject.CreatePlayer("player"))
	return &g, nil
}
