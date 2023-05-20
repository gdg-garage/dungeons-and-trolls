package dungeonsandtrolls

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"math/rand"
)

type ObsoleteMap [][][]gameobject.Interface

func CreateMap() (*ObsoleteMap, error) {
	var m ObsoleteMap
	m = append(m, baseFloor())
	monsterCounter := 1
	for l := 1; l < 3; l++ {
		m = append(m, generateFloor(l, 10, 10))
		// add monsters
		for i := 0; i < 1+rand.Intn(l); i++ {
			mYPos := rand.Intn(len(m[l])-1) + 1
			mXPos := rand.Intn(len(m[l][0])-1) + 1
			m[l][mYPos][mXPos].SetChildren(append(m[l][mYPos][mXPos].GetChildren(), gameobject.CreateMonster(fmt.Sprintf("monster %d", monsterCounter))))
			monsterCounter++
		}
	}
	return &m, nil
}

func baseFloor() [][]gameobject.Interface {
	var floor [][]gameobject.Interface
	stairsXPos := rand.Intn(9) + 1
	stairsYPos := rand.Intn(9) + 1
	for i := 0; i < 10; i++ {
		var row []gameobject.Interface
		for j := 0; j < 10; j++ {
			if i == stairsYPos && j == stairsXPos {
				row = append(row, gameobject.CreateStairs(1))
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

func generateFloor(level int, maxSizeX int, maxSizeY int) [][]gameobject.Interface {
	var floor [][]gameobject.Interface
	xSize := 10 + rand.Intn(maxSizeX)
	ySize := 10 + rand.Intn(maxSizeY)
	stairsXPos := rand.Intn(xSize-1) + 1
	stairsYPos := rand.Intn(ySize-1) + 1
	// TODO this may end up on the same place as the stair up
	stairsDownXPos := rand.Intn(xSize-1) + 1
	stairsDownYPos := rand.Intn(ySize-1) + 1
	for i := 0; i < ySize; i++ {
		var row []gameobject.Interface
		for j := 0; j < xSize; j++ {
			if i == stairsYPos && j == stairsXPos {
				row = append(row, gameobject.CreateStairs(level-1))
			} else if i == stairsDownYPos && j == stairsDownXPos {
				row = append(row, gameobject.CreateStairs(level+1))
			} else if j == 0 || j == xSize-1 || i == 0 || i == ySize-1 {
				row = append(row, gameobject.CreateWall())
			} else {
				row = append(row, gameobject.CreateEmpty())
			}
		}
		floor = append(floor, row)
	}
	return floor
}
