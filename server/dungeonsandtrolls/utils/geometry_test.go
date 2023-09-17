package utils

import "testing"

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
