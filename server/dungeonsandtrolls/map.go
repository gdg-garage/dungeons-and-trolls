package dungeonsandtrolls

import (
	"dungeons-and-trolls/dungeonsandtrolls/gameobject"
	"math/rand"
)

type Map [][][]gameobject.GameObject

func CreateMap() (*Map, error) {
	var m Map
	m = append(m, baseFloor())
	return &m, nil
}

func baseFloor() [][]gameobject.GameObject {
	var floor [][]gameobject.GameObject
	stairsXPos := rand.Intn(9) + 1
	stairsYPos := rand.Intn(9) + 1
	for i := 0; i < 10; i++ {
		var row []gameobject.GameObject
		for j := 0; j < 10; j++ {
			if i == stairsYPos && j == stairsXPos {
				row = append(row, gameobject.CreateStairs())
			} else if j == 0 || j == 9 || i == 0 || i == 9 {
				row = append(row, gameobject.CreateWall())
			} else {
				row = append(row, gameobject.CreateEmpty())
			}
		}
		floor = append(floor, row)
	}
	return floor
}
