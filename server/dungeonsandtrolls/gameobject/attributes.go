package gameobject

import (
	"fmt"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/api"
	"github.com/gdg-garage/dungeons-and-trolls/server/dungeonsandtrolls/utils"
	"reflect"
)

func AttributesOperation(receiver *api.Attributes, booster *api.Attributes, skipped map[string]struct{}, mustBePresent bool, operation func(r, b *float32)) error {
	fields := reflect.VisibleFields(reflect.ValueOf(booster).Elem().Type())
	for _, field := range fields {
		v := reflect.ValueOf(booster).Elem().FieldByName(field.Name)
		if !field.IsExported() {
			continue
		}
		if _, ok := skipped[field.Name]; ok {
			continue
		}
		if !v.IsNil() {
			vv, ok := v.Interface().(*float32)
			if !ok {
				return fmt.Errorf("attribute field %s is unexpected type %s (instead of *float32)", field.Name, field.Type)
			}
			r, ok := reflect.ValueOf(receiver).Elem().FieldByName(field.Name).Interface().(*float32)
			if !ok {
				return fmt.Errorf("attribute field %s is unexpected type %s (instead of *float32)", field.Name, reflect.ValueOf(receiver).Elem().FieldByName(field.Name).Type())
			}
			if r != nil {
				operation(r, vv)
			} else if !mustBePresent {
				reflect.ValueOf(receiver).Elem().FieldByName(field.Name).Set(v)
			}
		}
	}
	return nil
}

func MaxAllAttributes(receiver *api.Attributes, booster *api.Attributes, mustBePresent bool) error {
	return AttributesOperation(receiver, booster, map[string]struct{}{}, mustBePresent, func(r, b *float32) { *r = utils.Max(*r, *b) })
}

func MergeAttributes(receiver *api.Attributes, booster *api.Attributes, skipped map[string]struct{}, mustBePresent bool) error {
	return AttributesOperation(receiver, booster, skipped, mustBePresent, func(r, b *float32) { *r += *b })
}

func SubtractAllAttributes(receiver *api.Attributes, booster *api.Attributes, mustBePresent bool) error {
	// TODO how to set negative values for missing attributes?
	return AttributesOperation(receiver, booster, map[string]struct{}{}, mustBePresent, func(r, b *float32) { *r -= *b })
}

func MergeAllAttributes(receiver *api.Attributes, booster *api.Attributes, mustBePresent bool) error {
	return MergeAttributes(receiver, booster, map[string]struct{}{}, mustBePresent)
}

func SatisfyingAttributes(attributes *api.Attributes, requirements *api.Attributes) (bool, error) {
	fields := reflect.VisibleFields(reflect.ValueOf(requirements).Elem().Type())
	for _, field := range fields {
		v := reflect.ValueOf(requirements).Elem().FieldByName(field.Name)
		if !field.IsExported() {
			continue
		}
		if !v.IsNil() {
			vv, ok := v.Interface().(*float32)
			if !ok {
				return false, fmt.Errorf("attribute field %s is unexpected type %s (instead of *float32)", field.Name, field.Type)
			}
			r, ok := reflect.ValueOf(attributes).Elem().FieldByName(field.Name).Interface().(*float32)
			if !ok {
				return false, fmt.Errorf("attribute field %s is unexpected type %s (instead of *float32)", field.Name, reflect.ValueOf(attributes).Elem().FieldByName(field.Name).Type())
			}
			if r != nil {
				if *r < *vv {
					return false, nil
				}
			} else {
				return false, nil
			}
		}
	}
	return true, nil
}
