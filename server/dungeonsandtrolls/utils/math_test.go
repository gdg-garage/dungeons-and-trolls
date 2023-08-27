package utils

import (
	"testing"
)

func TestMaxInt(t *testing.T) {
	if Max(1, 2) != 2 {
		t.Fatal("wrong int comparison result")
	}
	if Max(2, 1) != 2 {
		t.Fatal("wrong int comparison result")
	}
}

func TestMaxfloat(t *testing.T) {
	if Max(1.2, 2.1) != 2.1 {
		t.Fatal("wrong float comparison result")
	}
	if Max(2.1, 1.2) != 2.1 {
		t.Fatal("wrong float comparison result")
	}

}
