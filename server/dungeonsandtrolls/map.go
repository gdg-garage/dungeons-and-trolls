package dungeonsandtrolls

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
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

func ParseMap(mapJson string) (*api.Map, error) {
	var parsedMap map[string]interface{}
	err := json.Unmarshal([]byte(mapJson), &parsedMap)
	if err != nil {
		return nil, err
	}
	maybeFloors, ok := parsedMap["floors"]
	if !ok {
		return nil, fmt.Errorf("floors not found in the generated map")
	}
	floors, ok := maybeFloors.([]interface{})
	if !ok {
		return nil, fmt.Errorf("floors are malformed in generated map")
	}
	m := &api.Map{}
	for _, f := range floors {
		l, err := parseLevel(f)
		if err != nil {
			return nil, err
		}
		m.Levels = append(m.Levels, l)
	}
	return m, nil
}

func parseLevel(level interface{}) (*api.Level, error) {
	l := &api.Level{}
	levelJson, err := json.Marshal(level)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(levelJson, &l)
	if err != nil {
		return nil, err
	}
	levelMap, ok := level.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("level is not a map")
	}
	maybeTiles, ok := levelMap["tiles"]
	if !ok {
		return nil, fmt.Errorf("tile not found in the level")
	}
	tiles, ok := maybeTiles.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tiles are malformed")
	}
	for _, maybeTile := range tiles {
		tile, ok := maybeTile.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("tile is malformed")
		}
		maybeType, ok := tile["type"]
		if !ok {
			return nil, fmt.Errorf("type not found in the tile")
		}
		t, ok := maybeType.(string)
		if !ok {
			return nil, fmt.Errorf("tile type is not string")
		}

		// This silently assumes that positions exist.
		o := &api.MapObjects{
			Position: &api.Coordinates{
				PositionX: int32(tile["x"].(float64)),
				PositionY: int32(tile["y"].(float64)),
			},
			Decorations: &api.MapDecorations{},
		}

		switch t {
		case "wall":
			o.Walkable = false
			o.Decorations.Wall = true
		case "spawn":
			// TODO use this
		default:
			log.Warn().Msgf("unknown terrain type %s", t)
		}

		l.Objects = append(l.Objects, o)

		// TODO parse data
		// field data is array
	}
	return l, nil
}
