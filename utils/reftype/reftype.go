package reftype

import "reflect"

// IsNil verifies whether the provided data is equal to nil of not.
func IsNil(instance any) bool {
	if instance == nil {
		return true
	}
	nilType := reflect.TypeOf((*reflect.Type)(nil)).Elem()
	if reflect.TypeOf(instance).Implements(nilType) {
		return true
	}
	value := reflect.ValueOf(instance)
	switch value.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array, reflect.Chan, reflect.Map:
		return value.IsNil()
	}
	return false
}

// IsTypeOf checks whether the provided data have the exact same type.
// The function will return true if the types of both values are identical, false otherwise.
func IsTypeOf(instance any, ref any) bool {
	if !IsNil(instance) {
		instance = reflect.TypeOf(instance)
	}
	if !IsNil(ref) {
		ref = reflect.TypeOf(ref)
	}
	return instance == ref
}

// IsStructEmbeds verifies whether the provided instance or type directly embeds the embed struct.
// The function will return true if the instance directly embeds the struct, false otherwise.
func IsStructEmbeds(instance any, embed any) bool {
	var insType reflect.Type
	var refType reflect.Type
	if !IsNil(instance) {
		insType = reflect.TypeOf(instance)
	} else {
		insType = instance.(reflect.Type)
	}
	if !IsNil(embed) {
		refType = reflect.TypeOf(embed)
	} else {
		refType = embed.(reflect.Type)
	}
	if insType.Kind() != reflect.Struct {
		return false
	}
	for i := 0; i < insType.NumField(); i++ {
		fType := insType.Field(i)
		if fType.Anonymous && fType.Type == refType {
			return true
		}
	}
	return false
}
