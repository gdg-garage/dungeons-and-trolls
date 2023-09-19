package gameobject

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
	"math"
	"math/rand"
	"reflect"
)

func AttributesValue(playerAttributes *api.Attributes, a *api.Attributes) (float64, error) {
	var value float64
	fields := reflect.VisibleFields(reflect.ValueOf(a).Elem().Type())
	for _, field := range fields {
		v := reflect.ValueOf(a).Elem().FieldByName(field.Name)
		if !field.IsExported() {
			continue
		}
		if !v.IsNil() {
			vv, ok := v.Interface().(*float32)
			if !ok {
				return value, fmt.Errorf("attribute field %s is unexpected type %s (instead of *float32)", field.Name, field.Type)
			}
			r, ok := reflect.ValueOf(playerAttributes).Elem().FieldByName(field.Name).Interface().(*float32)
			if !ok {
				return value, fmt.Errorf("attribute field %s is unexpected type %s (instead of *float32)", field.Name, reflect.ValueOf(playerAttributes).Elem().FieldByName(field.Name).Type())
			}
			if r != nil {
				value += float64(*r) * float64(*vv)
			}
		}
	}
	return value, nil
}

func RoundRange(r float64) float64 {
	return math.Floor(r)
}

func RoundSkill(r float64) float64 {
	whole := math.Floor(r)
	rest := r - whole
	if rest >= rand.Float64() {
		return math.Ceil(r)
	}
	return whole
}

func EvaluateDamage(power float64, t api.DamageType, a *api.Attributes) {
	var resist float64
	switch t {
	case api.DamageType_slash:
		if a.SlashResist != nil {
			resist = float64(*a.SlashResist)
		}
	case api.DamageType_pierce:
		if a.PierceResist != nil {
			resist = float64(*a.PierceResist)
		}
	case api.DamageType_fire:
		if a.FireResist != nil {
			resist = float64(*a.FireResist)
		}
	case api.DamageType_poison:
		if a.PoisonResist != nil {
			resist = float64(*a.PoisonResist)
		}
	case api.DamageType_electric:
		if a.ElectricResist != nil {
			resist = float64(*a.ElectricResist)
		}
	}
	*a.Life -= float32(RoundSkill(power) * 10 / (10 + utils.Max(resist, -5)))
}

func EvaluateSkillAttributes(sa *api.SkillAttributes, casterAttributes *api.Attributes) (*api.Attributes, error) {
	a := &api.Attributes{}
	fields := reflect.VisibleFields(reflect.ValueOf(sa).Elem().Type())
	for _, field := range fields {
		v := reflect.ValueOf(sa).Elem().FieldByName(field.Name)
		if !field.IsExported() {
			continue
		}
		if !v.IsNil() {
			vv, ok := v.Interface().(*api.Attributes)
			if !ok {
				return a, fmt.Errorf("attribute field %s is unexpected type %s (instead of *api.Attributes)", field.Name, field.Type)
			}
			rv, err := AttributesValue(casterAttributes, vv)
			frv := float32(rv)
			if err != nil {
				return a, err
			}
			reflect.ValueOf(a).Elem().FieldByName(field.Name).Set(reflect.ValueOf(&frv))
		}
	}
	return a, nil
}

func EvaluateEffects(effects []*api.Effect, a *api.Attributes) ([]*api.Effect, error) {
	var keptEffects []*api.Effect
	for _, e := range effects {
		err := MergeAllAttributes(a, e.Effects, false)
		EvaluateDamage(float64(e.DamageAmount), e.DamageType, a)
		if err != nil {
			return keptEffects, err
		}
		e.Duration--
		if e.Duration > 0 {
			keptEffects = append(keptEffects, e)
		}
	}
	return keptEffects, nil
}
