package utils

import "math"

func ManhattanDistance(fromX, fromY, toX, toY int32) int32 {
	return int32(math.Abs(float64(fromX-toX)) + math.Abs(float64(fromY-toY)))
}
