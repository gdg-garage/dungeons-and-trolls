package utils

import (
	"math"
	"testing"
)

func TestManhattanDistance(t *testing.T) {
	if ManhattanDistance(0, 0, 0, 0) != 0 {
		t.Fatal("invalid distance")
	}
	if ManhattanDistance(0, 0, 1, 1) != 2 {
		t.Fatal("invalid distance")
	}
	if ManhattanDistance(0, 0, 3, 0) != 3 {
		t.Fatal("invalid distance")
	}
	if ManhattanDistance(4, 4, 2, 1) != 5 {
		t.Fatal("invalid distance")
	}
}

func TestVectorFromPoints(t *testing.T) {
	v := VectorFromPoints(0, 0, 1, 1)
	if v.X != 1 || v.Y != 1 {
		t.Fatal("invalid vector from start")
	}
	v = VectorFromPoints(1, 1, 1, 1)
	if v.X != 0 || v.Y != 0 {
		t.Fatal("invalid null vector")
	}
	v = VectorFromPoints(1, 1, 8, 3)
	if v.X != 7 || v.Y != 2 {
		t.Fatal("invalid rando vector")
	}
}

func TestInverseVector(t *testing.T) {
	v := &V{
		X: 0,
		Y: 0,
	}
	InverseVector(v)
	if v.X != 0 || v.Y != 0 {
		t.Fatal("invalid vector from start")
	}
	v = &V{
		X: 10,
		Y: 2,
	}
	InverseVector(v)
	if v.X != -10 || v.Y != -2 {
		t.Fatal("invalid vector from start")
	}
}

func TestVectorLen(t *testing.T) {
	l := VectorLen(&V{
		X: 0,
		Y: 0,
	})
	if l != 0 {
		t.Fatal("invalid vector len")
	}
	l = VectorLen(&V{
		X: 0,
		Y: 2,
	})
	if l != 2 {
		t.Fatal("invalid vector len")
	}
	l = VectorLen(&V{
		X: 1,
		Y: 1,
	})
	if math.Abs(l-1.414) > 0.01 {
		t.Fatal("invalid vector len")
	}
	l = VectorLen(&V{
		X: 4,
		Y: 3,
	})
	if math.Abs(l-5) > 0.01 {
		t.Fatal("invalid vector len")
	}
}

func TestNormalizeVector(t *testing.T) {
	v := &V{
		X: 0,
		Y: 0,
	}
	NormalizeVector(v)
	if v.X-0 > 0.01 || v.Y-0 > 0.01 {
		t.Fatal("invalid normalized vector")
	}
	v = &V{
		X: 2,
		Y: 2,
	}
	NormalizeVector(v)
	if v.X-1 > 0.01 || v.Y-1 > 0.01 {
		t.Fatal("invalid normalized vector")
	}
}

func TestAddVectors(t *testing.T) {
	v1 := &V{
		X: 0,
		Y: 0,
	}
	v2 := &V{
		X: 0,
		Y: 0,
	}
	r := AddVectors(v1, v2)
	if r.X-0 > 0.01 || r.Y-0 > 0.01 {
		t.Fatal("invalid added vector")
	}
	v2 = &V{
		X: 1,
		Y: 2,
	}
	r = AddVectors(v1, v2)
	if r.X-1 > 0.01 || r.Y-2 > 0.01 {
		t.Fatal("invalid added vector")
	}
	v2 = &V{
		X: -4,
		Y: 2,
	}
	r = AddVectors(v1, v2)
	if r.X+4 > 0.01 || r.Y-2 > 0.01 {
		t.Fatal("invalid added vector")
	}
}
