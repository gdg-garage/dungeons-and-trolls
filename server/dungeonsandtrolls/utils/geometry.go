package utils

import (
	"math"
)

func ManhattanDistance(fromX, fromY, toX, toY int32) int32 {
	return int32(math.Abs(float64(fromX-toX)) + math.Abs(float64(fromY-toY)))
}

type V struct {
	X, Y float64
}

func VectorFromPoints(fromX, fromY, toX, toY int32) *V {
	return &V{
		X: float64(toX - fromX),
		Y: float64(toY - fromY),
	}
}

func InverseVector(v *V) {
	v.X *= float64(-1)
	v.Y *= float64(-1)
}

func VectorLen(v *V) float64 {
	return math.Sqrt(math.Pow(v.X, 2) + math.Pow(v.Y, 2))
}

func NormalizeVector(v *V) {
	l := VectorLen(v)
	v.X /= l
	v.Y /= l
}

func AddVectors(v1 *V, v2 *V) *V {
	return &V{
		X: v1.X + v2.X,
		Y: v1.Y + v2.Y,
	}
}
