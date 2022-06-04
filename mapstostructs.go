package mapstostructs

import (
	"fmt"
	"reflect"
)

const (
	notStructSliceReceiverMsg = "the receiver argument must be a ptr to a slice of struct but a %s was given"
	notStructReceiverMsg      = "the receiver argument must be a ptr to a struct but a %s was given"
	notMapReceiverMsg         = "the receiver argument must be a ptr to a map but a %s was given"
	notMapInputMsg            = "the input argument must be a map but a %s was given"
)

// MapsToStructs provides functionality for a slice of structs to be populated from a slice of map[string]interface{}
// with the option of passing alternative struct tags to use as map keys. If no tags are specified the json tag is used
// and if that is not present, the struct field is assumed. Keys are not case-sensitive.
//
// The receiver argument must be a pointer to a slice of structs.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
//
// Conversion of map[string]interface() to struct embedded within the slice of map[string]interface{} is permitted.
//
// Maps with numeric keys will accept string representations of numeric values.
func MapsToStructs(input []map[string]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(notStructSliceReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structValues := reflect.Indirect(reflect.ValueOf(receiver))
	if structValues.Kind() != reflect.Slice {
		return fmt.Errorf(notStructSliceReceiverMsg, "ptr to a "+structValues.Kind().String())
	}
	structType := structValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(notStructSliceReceiverMsg, "ptr to a slice of "+structType.Kind().String())
	}

	return setSlice(reflect.ValueOf(receiver).Elem(), reflect.ValueOf(input), tags)
}

// MapToStruct provides functionality for a struct to be populated from a map[string]interface{} with the option of
// passing alternative struct tags to use as map keys. If no tags are specified the json tag is used and if that is not
// present, the struct field is assumed. Keys are not case-sensitive.
//
// The receiver argument must be a pointer to a struct.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
//
// Conversion of map[string]interface() to struct embedded within the map[string]interface{} is permitted.
//
// Maps with numeric keys will accept string representations of numeric values.
func MapToStruct(input map[string]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(notStructReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structType := reflect.Indirect(reflect.ValueOf(receiver)).Type()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(notStructReceiverMsg, "ptr to a "+structType.Kind().String())
	}

	return setStructFromMap(reflect.ValueOf(receiver).Elem(), reflect.ValueOf(input), tags)
}

// MapToMap allows a map to be populated from another map, allowing key and value conversions where these are
// possible with the option of passing alternative struct tags to use as map keys. If no tags are specified the json
// tag is used and if that is not present, the struct field is assumed. Keys are not case-sensitive.
//
// The receiver argument must be a pointer to a map.
//
// The input argument must be a map.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
//
// Conversion of map[string]interface() to struct embedded within the map[string]interface{} is permitted.
//
// Maps with numeric keys will accept string representations of numeric values.
func MapToMap(input interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(notMapReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	mapValue := reflect.Indirect(reflect.ValueOf(receiver))
	if mapValue.Kind() != reflect.Map {
		return fmt.Errorf(notMapReceiverMsg, "ptr to a "+mapValue.Kind().String())
	}
	inputValue := reflect.ValueOf(input)
	if inputValue.Kind() != reflect.Map {
		return fmt.Errorf(notMapInputMsg, inputValue.Type().String())
	}

	return setMap(reflect.ValueOf(receiver).Elem(), inputValue, tags)
}
