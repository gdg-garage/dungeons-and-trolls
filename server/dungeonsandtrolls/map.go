package dungeonsandtrolls

import (
	"encoding/json"
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/gameobject"
	"github.com/rs/zerolog/log"
	"github.com/solarlune/paths"
	"go.openly.dev/pointy"
	"google.golang.org/protobuf/encoding/protojson"
)

type LevelCache struct {
	SpawnPoint         *api.Coordinates
	Width              int32
	Height             int32
	Objects            map[int32]map[int32]*api.MapObjects
	Grid               *paths.Grid
	GeneratedTick      int32
	LastInteractedTick int32
	Fow                map[int32]map[int32]bool
}

type MapCache struct {
	Level map[int32]*LevelCache
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
	// Ground effects
	return l, nil
}

func nonNil(d *api.Droppable) {
	switch i := d.Data.(type) {
	case *api.Droppable_Item:
		nonNilItem(i.Item)
	case *api.Droppable_Monster:
		nonNilMonster(i.Monster)
	}
}

func nonNilMonster(m *api.Monster) {
	for _, i := range m.EquippedItems {
		nonNilItem(i)
	}
	if m.Attributes == nil {
		m.Attributes = &api.Attributes{}
	}
	m.Stun = &api.Stun{}
	for _, d := range m.OnDeath {
		nonNil(d)
	}
}

func nonNilItem(i *api.Item) {
	if i.Attributes == nil {
		i.Attributes = &api.Attributes{}
	}
	if i.Requirements == nil {
		i.Requirements = &api.Attributes{}
	}
	for _, s := range i.Skills {
		nonNilISkill(s)
	}
}

func nonNilISkill(s *api.Skill) {
	if s.Cost == nil {
		s.Cost = &api.Attributes{}
	}
	if s.Range == nil {
		s.Range = &api.Attributes{}
	}
	if s.Radius == nil {
		s.Radius = &api.Attributes{}
	}
	if s.Duration == nil {
		s.Duration = &api.Attributes{}
	}
	if s.DamageAmount == nil {
		s.DamageAmount = &api.Attributes{}
	}
	if s.CasterEffects == nil {
		s.CasterEffects = &api.SkillEffect{Attributes: &api.SkillAttributes{}, Flags: &api.SkillSpecificFlags{}}
	}
	if s.TargetEffects == nil {
		s.TargetEffects = &api.SkillEffect{Attributes: &api.SkillAttributes{}, Flags: &api.SkillSpecificFlags{}}
	}
	if s.CasterEffects.Attributes == nil {
		s.CasterEffects.Attributes = &api.SkillAttributes{}
	}
	if s.TargetEffects.Attributes == nil {
		s.TargetEffects.Attributes = &api.SkillAttributes{}
	}
	if s.CasterEffects.Flags == nil {
		s.TargetEffects.Flags = &api.SkillSpecificFlags{}
	}
	if s.TargetEffects.Flags == nil {
		s.TargetEffects.Flags = &api.SkillSpecificFlags{}
	}
	if s.Flags == nil {
		s.Flags = &api.SkillGenericFlags{}
	}
	for _, sum := range s.TargetEffects.Summons {
		nonNil(sum)
	}
	for _, sum := range s.CasterEffects.Summons {
		nonNil(sum)
	}
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
		Position: &api.Position{
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
	case "door":
		o.IsDoor = true
		o.IsFree = false
	case "decoration", "monster", "chest", "waypoint":
		// We do not care, the specifics are defined in the data field.
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
		d := &api.Droppable{}
		err = protojson.Unmarshal(j, d)
		if err != nil {
			fmt.Println(string(j))
			return fmt.Errorf("tile data is malformed %v", err)
		}
		nonNil(d)

		switch i := d.Data.(type) {
		case *api.Droppable_Item:
			o.Items = append(o.Items, i.Item)
		case *api.Droppable_Monster:
			o.Monsters = append(o.Monsters, i.Monster)
		case *api.Droppable_Decoration:
			o.Decorations = append(o.Decorations, i.Decoration)
		case *api.Droppable_Waypoint:
			o.Portal = i.Waypoint
		case *api.Droppable_Skill:
			// Spawn ground effects - using dummy monster
			m := &api.Monster{
				Name:       "dummy - ground effect",
				Faction:    "none",
				Attributes: &api.Attributes{Life: pointy.Float32(0), Constant: pointy.Float32(1)},
				OnDeath:    []*api.Droppable{d}}
			nonNilMonster(m)
			o.Monsters = append(o.Monsters, m)
		default:
			log.Info().Msgf("I found something(%T) %v", i, i)
		}
	}
	return nil
}

func addIds(mp *api.Map) {
	for _, l := range mp.Levels {
		for _, o := range l.Objects {
			for _, m := range o.Monsters {
				// Items are not needed
				m.Id = gameobject.GetNewId()
				for _, i := range m.EquippedItems {
					for _, s := range i.Skills {
						s.Id = gameobject.GetNewId()
					}
				}
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
			return gameobject.PositionToCoordinates(o.Position, l.Level), nil
		}
	}
	return nil, fmt.Errorf("spawn point not found in the level")
}

func (lc *LevelCache) CacheObjectsOnPosition(p *api.Coordinates, mo *api.MapObjects) *api.MapObjects {
	if _, ok := lc.Objects[p.PositionX]; !ok {
		lc.Objects[p.PositionX] = map[int32]*api.MapObjects{}
	}
	v, ok := lc.Objects[p.PositionX][p.PositionY]
	if ok {
		return v
	}
	// Does not exist yet
	if mo != nil {
		lc.Objects[p.PositionX][p.PositionY] = mo
		c := lc.Grid.Get(int(p.PositionX), int(p.PositionY))
		if !mo.IsFree {
			c.Walkable = false
		}
		if mo.Portal != nil || mo.IsStairs {
			c.Cost = 5
		}
	} else {
		if _, ok := lc.Objects[p.PositionX][p.PositionY]; !ok {
			return nil
			// todo return error
		}
	}
	return lc.Objects[p.PositionX][p.PositionY]
}

func (m *MapCache) ClearLevelCache(l int32) {
	delete(m.Level, l)
}

func (m *MapCache) createLevelCache(l int32) *LevelCache {
	if _, ok := m.Level[l]; !ok {
		m.Level[l] = &LevelCache{
			Objects: map[int32]map[int32]*api.MapObjects{},
		}
	}
	m.Level[l].Fow = map[int32]map[int32]bool{}
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
		log.Info().Msgf("generated level %d", l.Level)
		// Strip items from the first level.
		if l.Level == gameobject.ZeroLevel {
			for _, s := range g.Game.ShopItems {
				g.Unregister(s)
			}
			g.Game.ShopItems = []*api.Item{}
			for _, o := range l.Objects {
				for _, i := range o.Items {
					g.AddItem(i)
				}
				o.Items = []*api.Item{}
			}
		}

		for _, o := range l.Objects {
			for _, i := range o.Items {
				g.Register(i)
			}
			for _, m := range o.Monsters {
				levelPos := &api.Coordinates{
					PositionX: o.Position.PositionX,
					PositionY: o.Position.PositionY,
					Level:     l.Level,
				}
				g.Register(gameobject.CreateMonster(m, levelPos))
			}
		}

		lc := mapCache.createLevelCache(l.Level)
		lc.Width = l.Width
		lc.Height = l.Height
		lc.GeneratedTick = g.Game.Tick
		lc.LastInteractedTick = g.Game.Tick
		lc.Grid = paths.NewGrid(int(l.Width), int(l.Height), 1, 1)

		var x, y int32
		for x = 0; x < l.Width; x++ {
			lc.Fow[x] = map[int32]bool{}
			for y = 0; y < l.Height; y++ {
				// TODO enable FOW here + special handling for 0 level
				lc.Fow[x][y] = false
			}
		}

		spawn, err := findLevelSpawnPoint(l)
		if err != nil {
			return err
		}
		lc.SpawnPoint = spawn

		for _, o := range l.Objects {
			lc.CacheObjectsOnPosition(gameobject.PositionToCoordinates(o.Position, l.Level), o)
		}
	}
	return nil
}
