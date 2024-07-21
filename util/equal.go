package util

import "reflect"

// comparing every single field for 2 objects
func AreStructsEqual(a, b interface{}) bool {
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)

	// different type, return false
	if valA.Type() != valB.Type() {
		return false
	}

	// not struct type, return false
	if valA.Kind() != reflect.Struct || valB.Kind() != reflect.Struct {
		return false
	}

	// comparing every single filed if equals
	for i := 0; i < valA.NumField(); i++ {
		fieldA := valA.Field(i)
		fieldB := valB.Field(i)

		if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) {
			return false
		}
	}

	return true
}
