package dungeonsandtrolls

import (
	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
)

type ObsoleteMap [][][]gameobject.Interface

type LevelCache struct {
	SpawnPoint *api.Coordinates
	Objects    map[int32]map[int32]*api.MapObjects
}

type MapCache struct {
	Level map[int32]*LevelCache
}

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

	addIds(m)

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
		err = parseTile(maybeTile, l)
		if err != nil {
			return nil, err
		}
	}
	return l, nil
}

func parseTile(maybeTile interface{}, l *api.Level) error {
	tile, ok := maybeTile.(map[string]interface{})
	if !ok {
		return fmt.Errorf("tile is malformed")
	}
	maybeType, ok := tile["type"]
	if !ok {
		return fmt.Errorf("type not found in the tile")
	}
	t, ok := maybeType.(string)
	if !ok {
		return fmt.Errorf("tile type is not string")
	}

	// This silently assumes that positions exist.
	o := &api.MapObjects{
		Position: &api.Coordinates{
			PositionX: int32(tile["x"].(float64)),
			PositionY: int32(tile["y"].(float64)),
		},
	}

	o.IsFree = true

	// TODO this is just a droppable.
	switch t {
	case "wall":
		o.IsFree = false
		o.IsWall = true
	case "spawn":
		spawn := true
		o.IsSpawn = &spawn
	case "stairs":
		o.IsStairs = true
	case "decoration", "monster":
		// We do not care, the specific decoration is defined in the data field.
	default:
		log.Warn().Msgf("unknown terrain type %s", t)
	}

	err := parseMapObjects(tile, o)
	if err != nil {
		return err
	}

	l.Objects = append(l.Objects, o)

	return nil
}

func parseMapObjects(tile map[string]interface{}, o *api.MapObjects) error {
	maybeData, ok := tile["data"]
	if !ok {
		return nil
	}
	data, ok := maybeData.([]interface{})
	if !ok {
		return fmt.Errorf("tile data is malformed")
	}
	for _, d := range data {
		j, err := json.Marshal(d)
		if err != nil {
			return fmt.Errorf("tile data serialization failed %v", err)
		}
		d := &api.Dropable{}
		err = protojson.Unmarshal(j, d)
		if err != nil {
			fmt.Println(string(j))
			return fmt.Errorf("tile data is malformed %v", err)
		}

		switch i := d.Data.(type) {
		case *api.Dropable_Item:
			o.Items = append(o.Items, i.Item)
			// log.Info().Msgf("I found item %v", i)
		case *api.Dropable_Monster:
			o.Monsters = append(o.Monsters, i.Monster)
			// log.Info().Msgf("I found monster %v", i)
		case *api.Dropable_Decoration:
			o.Decorations = append(o.Decorations, i.Decoration)
			// log.Info().Msgf("I found decoration %v", i)
		default:
			log.Info().Msgf("I found something(%T) %v", i, i)
		}
	}
	return nil

	// maybeClass, ok := d["class"]
	// if !ok {
	// 	return fmt.Errorf("type not found in the tile data")
	// }
	// c, ok := maybeClass.(string)
	// if !ok {
	// 	return fmt.Errorf("tile data type is not string")
	// }
	// switch c {
	// case "decoration":
	// 	maybeType, ok := d["type"]
	// 	if !ok {
	// 		return fmt.Errorf("type not found in the tile data decoration")
	// 	}
	// 	t, ok := maybeType.(string)
	// 	if !ok {
	// 		return fmt.Errorf("tile data decoration type is not string")
	// 	}

	// 	o.Decoration = t
	// case "item":
	// 	delete(d, "class")

	// 	// maybeSkills, ok := d["skills"]
	// 	// if !ok {
	// 	// 	return fmt.Errorf("type not found in the tile data decoration")
	// 	// }
	// 	// skills, ok := maybeSkills.([]interface{})
	// 	// if !ok {
	// 	// 	return fmt.Errorf("type not found in the tile data decoration")
	// 	// }

	// 	// log.Info().Msgf("item: %v", d)
	// 	// log.Info().Msgf("skills: %v", skills)
	// 	for _, skill := range d["skills"].([]interface{}) {
	// 		s, _ := skill.(map[string]interface{})
	// 		delete(d, "class")
	// 		if cfs, ok := s["casterFlags"]; ok {

	// 			log.Info().Msgf("casterFlags %v %s", cfs, reflect.TypeOf(cfs))
	// 			switch v := cfs.(type) {
	// 			case []interface{}:
	// 				for _, cf := range v {

	// 					log.Info().Msgf("flag %s", cf)
	// 					a := api.SkillFlag{}
	// 					sfj, _ := json.Marshal(map[string]string{
	// 						"flag": cf.(string),
	// 					})
	// 					log.Info().Msgf("flag %s", string(sfj))
	// 					protojson.Unmarshal(sfj, &a)
	// 					log.Info().Msgf("%d", a.String())
	// 				}
	// 			default:
	// 				log.Info().Msgf("casterFlags %d", cfs)
	// 			}

	// 		}
	// 		delete(s, "casterFlags")
	// 		delete(s, "targetFlags")
	// 	}

	// 	// TODO use branch restructure

	// 	log.Info().Msgf("item: %v", d)

	// 	j, err := json.Marshal(d)
	// 	if err != nil {
	// 		return fmt.Errorf("item serialization failed %v", err)
	// 	}
	// 	i := &api.Item{}

	// 	// d, _ := protojson.Marshal(&api.SkillFlag{Data: &api.SkillFlag_Flag{Flag: "aaaa"}})
	// 	// log.Info().Msg(string(d))
	// 	// d, _ = protojson.Marshal(&api.SkillFlag{
	// 	// 	Data: &api.SkillFlag_Dropables{Dropables: &api.Dropables{Data: &api.Dropables_Decoration{Decoration: &api.Decoration{Name: "bb"}}}},
	// 	// })
	// 	// log.Info().Msg(string(d))

	// 	// err = protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(j, i)
	// 	err = protojson.UnmarshalOptions{}.Unmarshal(j, i)
	// 	// err = protojson.UnmarshalOptions{}.Unmarshal(j, i)
	// 	if err != nil {
	// 		log.Info().Msgf("%v", string(j))
	// 		return fmt.Errorf("item deserialization failed %v", err)
	// 	}
	// 	i.Id = gameobject.GetNewId()
	// 	for _, skill := range i.Skills {
	// 		skill.Id = gameobject.GetNewId()

	// 		if skill.Cost != nil {
	// 			skill.Cost = &api.Attributes{}
	// 		}
	// 	}
	// 	if i.Requirements == nil {
	// 		i.Requirements = &api.Attributes{}
	// 	}
	// 	if i.Attributes == nil {
	// 		i.Attributes = &api.Attributes{}
	// 	}
	// 	o.Items = append(o.Items, i)
	// 	log.Info().Msg(string(j))
	// default:
	// 	log.Warn().Msgf("Unknown data tile class: %s", c)
	// 	log.Info().Msgf("%v", d)
	// }

	// }
	// return nil
}

func addIds(mp *api.Map) {
	for _, l := range mp.Levels {
		for _, o := range l.Objects {
			for _, m := range o.Monsters {
				// when to handle on death items
				// skills and items are not needed
				m.Id = gameobject.GetNewId()
			}
			for _, i := range o.Items {
				i.Id = gameobject.GetNewId()
				for _, s := range i.Skills {
					s.Id = gameobject.GetNewId()
				}
			}
		}
	}
}

func findLevelSpawnPoint(l *api.Level) (*api.Coordinates, error) {
	for _, o := range l.Objects {
		if o.IsSpawn != nil && *o.IsSpawn {
			return o.Position, nil
		}
	}
	return nil, fmt.Errorf("spawn point not found in the level")
}

func (lc *LevelCache) CacheObjectsOnPosition(p *api.Coordinates, mo *api.MapObjects) *api.MapObjects {
	if _, ok := lc.Objects[p.PositionX]; !ok {
		lc.Objects[p.PositionX] = map[int32]*api.MapObjects{}
	}
	if mo != nil {
		lc.Objects[p.PositionX][p.PositionY] = mo
	} else {
		if _, ok := lc.Objects[p.PositionX][p.PositionY]; !ok {
			return nil
			// todo return error
		}
	}
	return lc.Objects[p.PositionX][p.PositionY]
}

func (m *MapCache) CacheLevel(l int32) *LevelCache {
	if _, ok := m.Level[l]; !ok {
		m.Level[l] = &LevelCache{
			Objects: map[int32]map[int32]*api.MapObjects{},
		}
	}
	return m.Level[l]
}

func (m *MapCache) CachedLevel(l int32) (*LevelCache, error) {
	if _, ok := m.Level[l]; !ok {
		return nil, fmt.Errorf("cache for level %d not found", l)
	}
	return m.Level[l], nil
}

func LevelsPostProcessing(g *Game, m *api.Map, mapCache *MapCache) error {
	for _, l := range m.Levels {
		// Strip items from the first level.
		if l.Level == gameobject.ZeroLevel {
			for _, o := range l.Objects {
				// TODO use game.AddItem()
				g.Game.Items = append(g.Game.Items, o.Items...)
				o.Items = []*api.Item{}
			}
		}

		lc := mapCache.CacheLevel(l.Level)

		spawn, err := findLevelSpawnPoint(l)
		if err != nil {
			return err
		}
		lc.SpawnPoint = spawn

		for _, o := range l.Objects {
			lc.CacheObjectsOnPosition(o.Position, o)
		}
	}
	return nil
}
