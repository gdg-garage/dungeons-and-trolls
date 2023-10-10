package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"math"
)

type MapCellExt struct {
	MapObjects  *api.MapObjects
	Distance    int
	LineOfSight bool
}

type PlainPos struct {
	PositionX int32
	PositionY int32
}

func PlainPosFromApiPos(position *api.Position) PlainPos {
	return PlainPos{
		PositionX: position.PositionX,
		PositionY: position.PositionY,
	}
}

func CalculateDistanceAndLineOfSight(currentMap *api.Level, currentPosition *api.Position) map[PlainPos]MapCellExt {
	// Distance to obstacles used for line of sight
	distanceToFirstObstacle := make(map[float32]float32)
	currentPlainPos := PlainPosFromApiPos(currentPosition)

	// map for resulting map positions with Distance and line of sight
	resultMap := make(map[PlainPos]MapCellExt)
	// fill map with map objects
	for _, objects := range currentMap.Objects {
		resultMap[PlainPosFromApiPos(objects.Position)] = MapCellExt{
			MapObjects:  objects,
			Distance:    -1,
			LineOfSight: false,
		}
	}

	//log.Info().Msgf("Original map -> (player: A, no data / free: ' ', wall: w, spawn: *, stairs: s, unknown: ?)")
	//for y := int32(0); y < currentMap.Height; y++ {
	//	row := ""
	//	for x := int32(0); x < currentMap.Width; x++ {
	//		cell, found := resultMap[makePosition(x, y)]
	//		if currentPosition.PositionX == x && currentPosition.PositionY == y {
	//			row += "A"
	//		} else if !found {
	//			row += " "
	//		} else if cell.MapObjects.IsSpawn != nil && *cell.MapObjects.IsSpawn {
	//			row += "*"
	//		} else if cell.MapObjects.IsStairs {
	//			row += "s"
	//		} else if cell.MapObjects.IsFree {
	//			row += " "
	//		} else if cell.MapObjects.IsWall {
	//			row += "w"
	//		} else {
	//			row += "?"
	//		}
	//	}
	//	log.Info().Msgf("Map row: %s (y = %d)", row, y)
	//}

	// standard BFS stuff
	visited := make(map[PlainPos]bool)
	var queue []PlainPos

	// start from player
	// add current node to queue and add its Distance to final map
	queue = append(queue, currentPlainPos)
	cell, found := resultMap[currentPlainPos]
	mapObjects := &api.MapObjects{}
	if found {
		mapObjects = cell.MapObjects
	}
	resultMap[currentPlainPos] = MapCellExt{
		MapObjects:  mapObjects,
		Distance:    0,
		LineOfSight: true,
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		nodeVisited, found := visited[node]
		if !found || !nodeVisited {
			visited[node] = true

			// Enqueue all unvisited neighbors
			for _, neighbor := range getNeighbors(node) {
				// neighbors can be out of the map,
				cell, found := resultMap[neighbor]
				// must be in bounds
				// must not be visited
				// must be free
				if isInBounds(currentMap, &neighbor) && !visited[neighbor] && (!found || cell.MapObjects.IsFree) {
					mapObjects := &api.MapObjects{
						IsFree: true,
					}
					if found {
						mapObjects = cell.MapObjects
					}
					distance := resultMap[node].Distance + 1
					lineOfSight := GetLoS(currentMap, resultMap, distanceToFirstObstacle, currentPosition, neighbor)
					resultMap[neighbor] = MapCellExt{
						MapObjects:  mapObjects,
						Distance:    distance,
						LineOfSight: lineOfSight,
					}
					queue = append(queue, neighbor)
				}
			}
		}
	}

	//log.Info().Msgf("Map with distances -> (player: A, no data: !, not reachable: ~, Distance < 10: 0-9, Distance >= 10: +)")
	//for y := int32(0); y < currentMap.Height; y++ {
	//	row := ""
	//	for x := int32(0); x < currentMap.Width; x++ {
	//		cell, found := resultMap[makePosition(x, y)]
	//		if makePosition(x, y) == currentPosition {
	//			row += "A"
	//		} else if !found {
	//			row += "!"
	//		} else if cell.Distance < 10 {
	//			row += fmt.Sprintf("%d", cell.Distance)
	//		} else if cell.Distance == math.MaxInt32 {
	//			row += "~"
	//		} else {
	//			row += "+"
	//		}
	//	}
	//	//b.Logger.Debugf("Map row: %s (y = %d)", row, y)
	//}

	//log.Info().Msgf("Map with line of sight -> (player: A, no data: !, line of sight: ' ', wall: w, no line of sight: ~)")
	//for y := int32(0); y < currentMap.Height; y++ {
	//	row := ""
	//	for x := int32(0); x < currentMap.Width; x++ {
	//		cell, found := resultMap[makePosition(x, y)]
	//		if currentPosition.PositionX == x && currentPosition.PositionY == y {
	//			row += "A"
	//		} else if !found {
	//			row += "!"
	//		} else if cell.LineOfSight {
	//			row += " "
	//		} else if cell.MapObjects.IsWall {
	//			row += "w"
	//		} else {
	//			row += "~"
	//		}
	//	}
	//	log.Info().Msgf("Map row: %s (y = %d)", row, y)
	//}

	return resultMap
}

func makePosition(x int32, y int32) PlainPos {
	return PlainPos{
		PositionX: x,
		PositionY: y,
	}
}

func getNeighbors(pos PlainPos) []PlainPos {
	return []PlainPos{
		makePosition(pos.PositionX-1, pos.PositionY),
		makePosition(pos.PositionX+1, pos.PositionY),
		makePosition(pos.PositionX, pos.PositionY-1),
		makePosition(pos.PositionX, pos.PositionY+1),
	}
}

func isInBounds(currentMap *api.Level, pos *PlainPos) bool {
	return pos.PositionX >= 0 && pos.PositionX < currentMap.Width && pos.PositionY >= 0 && pos.PositionY < currentMap.Height
}

func GetLoS(currentLevel *api.Level, resultMap map[PlainPos]MapCellExt, distanceToFirstObstacle map[float32]float32, pos1 *api.Position, pos2 PlainPos) bool {
	// get the center of the cell
	x1 := float32(pos1.PositionX) + 0.5
	y1 := float32(pos1.PositionY) + 0.5
	x2 := float32(pos2.PositionX) + 0.5
	y2 := float32(pos2.PositionY) + 0.5

	// slope := float32(y2-y1) / float32(x2-x1)
	distance := math.Sqrt(float64((x2-x1)*(x2-x1) + (y2-y1)*(y2-y1)))
	// angle in radians
	slope := float32(math.Atan2(float64(y2-y1), float64(x2-x1)))
	// angleDegrees := angleRadians * 180 / math.Pi

	// TODO: somehow round the value to prevent cache misses
	losDist, found := distanceToFirstObstacle[slope]
	if found {
		//	log.Info().Msgf("LoS: found in cache",
		//		"playerPosition", pos1,
		//		"position", pos2,
		//		"slope", slope,
		//		"Distance", Distance,
		//		"lineOfSightDistance", losDist,
		//		"LineOfSight", Distance < float64(losDist),
		//	)
		return distance < float64(losDist)
	}
	losDist = rayTrace(currentLevel, resultMap, slope, x1, y1, x2, y2)
	distanceToFirstObstacle[slope] = losDist
	//log.Info().Msgf("LoS: calculated",
	//	"playerPosition", pos1,
	//	"position", pos2,
	//	"slope", slope,
	//	"Distance", Distance,
	//	"lineOfSightDistance", losDist,
	//	"LineOfSight", Distance < float64(losDist),
	//)
	return distance < float64(losDist)
}

func rayTrace(currentLevel *api.Level, resultMap map[PlainPos]MapCellExt, slope float32, x1 float32, y1 float32, x2 float32, y2 float32) float32 {
	dx := x2 - x1
	dy := y2 - y1

	// Calculate absolute values of dx and dy
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}

	// Determine the sign of movement along x and y
	sx := float32(1)
	sy := float32(1)
	if x1 > x2 {
		sx = -1
	}
	if y1 > y2 {
		sy = -1
	}

	// Initialize error variables
	e := dx - dy
	x := x1
	y := y1

	for {
		// TODO: any mapping needed here?
		pos := getPositionsForFloatCoords(x, y)
		// Check the current cell for obstacles or objects
		cell, found := resultMap[pos]

		// obstacle hit if end of map OR not free
		if !isInBounds(currentLevel, &pos) || (found && !cell.MapObjects.IsFree) {
			dist := math.Sqrt(float64((x-x1)*(x-x1) + (y-y1)*(y-y1)))
			return float32(dist)
		}

		// Calculate the next step
		e2 := 2 * e
		if e2 > -dy {
			e -= dy
			x += sx
		}
		if e2 < dx {
			e += dx
			y += sy
		}
	}
}

func getPositionsForFloatCoords(x float32, y float32) PlainPos {
	// I was worried about what position to return if the float values are exactly on the border between two positions.
	// But it looks like this works fine.
	// NOTE: This might be something to adjust if we see weird line of sight.
	//			 E.g. if we see different LoS on right and left side of player or obstacle.
	return PlainPos{
		PositionX: int32(x),
		PositionY: int32(y),
	}
}
