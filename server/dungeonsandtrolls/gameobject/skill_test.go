package gameobject

import (
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"go.openly.dev/pointy"
	"testing"
)

func TestSingleAttributesValue(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Willpower: pointy.Float32(1),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(2),
		Dexterity: pointy.Float32(3),
	}

	v, err := AttributesValue(a, b)
	if err != nil {
		t.Error(err)
	}

	if v != 2 {
		t.Fatalf("attributes value incorrect")
	}
}

func TestMultipleAttributesValue(t *testing.T) {
	a := &api.Attributes{
		Strength:  pointy.Float32(1),
		Dexterity: pointy.Float32(2),
		Willpower: pointy.Float32(1),
	}
	b := &api.Attributes{
		Strength:  pointy.Float32(0.5),
		Dexterity: pointy.Float32(3),
	}

	v, err := AttributesValue(a, b)
	if err != nil {
		t.Error(err)
	}

	if v != 6.5 {
		t.Fatalf("attributes value incorrect")
	}
}

func TestEvaluateSkillAttributes(t *testing.T) {
	a := &api.SkillAttributes{
		Strength: &api.Attributes{
			Strength:  pointy.Float32(1),
			Dexterity: pointy.Float32(3),
		},
		Dexterity: &api.Attributes{
			Strength: pointy.Float32(0.5),
			Scalar:   pointy.Float32(3),
		},
		Willpower: &api.Attributes{
			Willpower: pointy.Float32(1),
			Dexterity: pointy.Float32(1),
		},
	}
	caster := &api.Attributes{
		Strength: pointy.Float32(2),
		Scalar:   pointy.Float32(1),
	}

	v, err := EvaluateSkillAttributes(a, caster)
	if err != nil {
		t.Error(err)
	}

	if *v.Strength != 2 {
		t.Fatalf("attributes value incorrect")
	}
	if *v.Dexterity != 4 {
		t.Fatalf("attributes value incorrect")
	}
	if *v.Willpower != 0 {
		t.Fatalf("attributes value incorrect")
	}
}
