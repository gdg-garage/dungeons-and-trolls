package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"go.openly.dev/pointy"
	"testing"
)

func TestAttributesMerge(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Willpower: pointy.Float32(1),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(3),
	}

	err := MergeAllAttributes(a, b, false)
	if err != nil {
		t.Error(err)
	}

	if a.Strength == nil {
		t.Fatalf("field is removed")
	}
	if *a.Strength != 2 {
		t.Errorf("field is not merged")
	}

	if a.Willpower == nil {
		t.Fatalf("field is removed")
	}
	if *a.Willpower != 1 {
		t.Errorf("field is not merged")
	}

	if a.Dexterity == nil {
		t.Fatalf("field is not propagated")
	}
	if *a.Dexterity != 3 {
		t.Errorf("field is not propagated")
	}
}

func TestPresentAttributesMerge(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Willpower: pointy.Float32(1),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(3),
	}

	err := MergeAllAttributes(a, b, true)
	if err != nil {
		t.Error(err)
	}

	if a.Strength == nil {
		t.Fatalf("field is removed")
	}
	if *a.Strength != 2 {
		t.Errorf("field is not merged")
	}

	if a.Willpower == nil {
		t.Fatalf("field is removed")
	}
	if *a.Willpower != 1 {
		t.Errorf("field is not merged")
	}

	if a.Dexterity != nil {
		t.Fatalf("field is propagated")
	}
}

func TestAttributesMergeSkipped(t *testing.T) {
	a := &api.Attributes{
		Life:    pointy.Float32(1),
		Stamina: pointy.Float32(1),
	}
	b := &api.Attributes{
		Life:    pointy.Float32(1),
		Stamina: pointy.Float32(1),
	}

	err := MergeAttributes(a, b, map[string]struct{}{"Life": {}}, false)
	if err != nil {
		t.Error(err)
	}

	if a.Life == nil {
		t.Fatalf("field is removed")
	}
	if *a.Life != 1 {
		t.Errorf("field is merged but should be ignored")
	}

	if a.Stamina == nil {
		t.Fatalf("field is removed")
	}
	if *a.Stamina != 2 {
		t.Errorf("field is not merged")
	}
}

func TestAttributesMax(t *testing.T) {
	a := &api.Attributes{
		Life:    pointy.Float32(1),
		Stamina: pointy.Float32(1),
	}
	b := &api.Attributes{
		Life:      pointy.Float32(2),
		Dexterity: pointy.Float32(4),
	}

	err := MaxAllAttributes(a, b, false)
	if err != nil {
		t.Error(err)
	}

	if a.Life == nil {
		t.Fatalf("field is removed")
	}
	if *a.Life != 2 {
		t.Errorf("field max incorrect")
	}

	if a.Stamina == nil {
		t.Fatalf("field is removed")
	}
	if *a.Stamina != 1 {
		t.Errorf("field max incorrect")
	}

	if a.Dexterity == nil {
		t.Fatalf("field is removed")
	}
	if *a.Dexterity != 4 {
		t.Errorf("field max incorrect")
	}
}

func TestRequirementsOk(t *testing.T) {
	a := &api.Attributes{
		Strength:     pointy.Float32(1),
		Dexterity:    pointy.Float32(3),
		Constitution: pointy.Float32(1),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(2),
	}

	s, err := SatisfyingAttributes(a, b)
	if err != nil {
		t.Error(err)
	}
	if !s {
		t.Errorf("requirements are not satisfied")
	}
}

func TestRequirementsFailMissing(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(3),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(2),
		Willpower: pointy.Float32(2),
	}

	s, err := SatisfyingAttributes(a, b)
	if err != nil {
		t.Error(err)
	}
	if s {
		t.Errorf("requirements are satisfied")
	}
}

func TestRequirementsFailLow(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(3),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(2),
		Dexterity: pointy.Float32(2),
	}

	s, err := SatisfyingAttributes(a, b)
	if err != nil {
		t.Error(err)
	}
	if s {
		t.Errorf("requirements are satisfied")
	}
}

func TestSum(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(3),
	}

	s, err := SumAttributes(a)
	if err != nil {
		t.Error(err)
	}
	if s != 4 {
		t.Errorf("sum is not correct")
	}
}
