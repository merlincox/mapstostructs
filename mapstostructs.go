package mapstostructs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	notStructSliceReceiversMsg = "the receiver argument must be a ptr to a slice of struct but a %s was given"
	notStructReceiverMsg       = "the receiver argument must be a ptr to a struct but a %s was given"
	notMapReceiverMsg          = "the receiver argument must be a ptr to a map but a %s was given"

	badValueMsg    = "must be or be convertible to %s type, but received '%v'"
	structPrefix   = "the %s field for a struct of type %s "
	rowSuffix      = " in row %d"
	mapKeyPrefix   = "the map key for a %s "
	mapValuePrefix = "the map value for a %s "

	jsonTag = "json"
)

// MapsToStructs provides functionality for a slice of structs to be populated from a slice of map[string]interface{}
// with the option of passing alternative struct tags to use as map keys. If no tags are specified the json tag is
// used and if that is not present, the lowercase value of the struct field is assumed.
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
		return fmt.Errorf(notStructSliceReceiversMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structValues := reflect.Indirect(reflect.ValueOf(receiver))
	if structValues.Kind() != reflect.Slice {
		return fmt.Errorf(notStructSliceReceiversMsg, "ptr to a "+structValues.Kind().String())
	}
	structType := structValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(notStructSliceReceiversMsg, "ptr to a slice of "+structType.Kind().String())
	}

	return setSlice(reflect.ValueOf(receiver).Elem(), reflect.ValueOf(input), tags)
}

// MapToStruct provides functionality for a struct to be populated from a map[string]interface{} with the option of
// passing alternative struct tags to use as map keys. If no tags are specified the json tag is used and if that is not
// present, the lowercase value of the struct field is assumed.
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
// tag is used and if that is not present, the lowercase value of the struct field is assumed.
//
// The receiver argument must be a pointer to a map.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
//
// Conversion of map[string]interface() to struct embedded within the map[string]interface{} is permitted.
//
// Maps with numeric keys will accept string representations of numeric values.
func MapToMap(input map[string]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(notMapReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	mapValue := reflect.Indirect(reflect.ValueOf(receiver))
	if mapValue.Kind() != reflect.Map {
		return fmt.Errorf(notMapReceiverMsg, "ptr to a "+mapValue.Kind().String())
	}

	return setMap(reflect.ValueOf(receiver).Elem(), reflect.ValueOf(input), tags)
}

func makeTagMap(structType reflect.Type, tags []string) map[string]string {
	numFields := structType.NumField()
	tagMap := make(map[string]string, numFields)
	tags = append(tags, jsonTag)
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		var tagged bool
		for _, tagName := range tags {
			tag, ok := field.Tag.Lookup(tagName)
			if ok {
				parts := strings.Split(tag, ",")
				tagMap[parts[0]] = field.Name
				tagged = true
				break
			}
		}
		if !tagged {
			tagMap[strings.ToLower(field.Name)] = field.Name
		}
	}

	return tagMap
}

func setSlice(receiver reflect.Value, input reflect.Value, tags []string) error {
	if input.Len() == 0 {

		return nil
	}
	elementType := receiver.Type().Elem()
	newSliceValue := reflect.MakeSlice(reflect.SliceOf(elementType), 0, input.Len())
	for i := 0; i < input.Len(); i++ {
		newElement := reflect.Indirect(reflect.New(elementType))
		if err := setRecursively(newElement, input.Index(i), tags); err != nil {

			return fmt.Errorf(err.Error()+rowSuffix, i+1)
		}
		newSliceValue = reflect.Append(newSliceValue, newElement)
	}
	setValue(receiver, newSliceValue)

	return nil
}

func setStructFromMap(receiver reflect.Value, input reflect.Value, tags []string) error {
	if input.Len() == 0 {
		return nil
	}
	wantType := receiver.Type()
	if receiver.Kind() == reflect.Ptr {
		wantType = receiver.Type().Elem()
	}
	tagMap := makeTagMap(wantType, tags)
	newStructValue := reflect.Indirect(reflect.New(wantType))
	mapRange := input.MapRange()
	for mapRange.Next() {
		if fieldName, ok := tagMap[mapRange.Key().String()]; ok {
			receivingField := newStructValue.FieldByName(fieldName)
			inputField := mapRange.Value().Elem()
			if err := setRecursively(receivingField, inputField, tags); err != nil {

				return fmt.Errorf(structPrefix+err.Error(), fieldName, receiver.Type().Name())
			}
		}
	}
	setValue(receiver, newStructValue)

	return nil
}

func setMap(receiver reflect.Value, input reflect.Value, tags []string) error {
	if input.Len() == 0 {

		return nil
	}
	wantType := receiver.Type()
	wantKeyType := wantType.Key()
	newMapValue := reflect.MakeMap(wantType)
	mapRange := input.MapRange()

	for mapRange.Next() {
		key, ok := convertToType(mapRange.Key(), wantKeyType, true)
		if !ok {
			want := wantKeyType.String()
			have := mapRange.Key().Interface()

			return fmt.Errorf(mapKeyPrefix+badValueMsg, wantType.String(), want, have)
		}

		newElement := reflect.Indirect(reflect.New(wantType.Elem()))
		if err := setRecursively(newElement, mapRange.Value(), tags); err != nil {

			return fmt.Errorf(mapValuePrefix+err.Error(), wantType.String())
		}
		newMapValue.SetMapIndex(key, newElement)
	}
	setValue(receiver, newMapValue)

	return nil
}

func setRecursively(receiver reflect.Value, input reflect.Value, tags []string) error {
	if input.Kind() == reflect.Ptr || input.Kind() == reflect.Interface {

		return setRecursively(receiver, input.Elem(), tags)
	}
	have := input.Interface()
	wantType := receiver.Type()
	if wantType.Kind() == reflect.Ptr {
		wantType = wantType.Elem()
	}
	want := wantType.String()

	if valueToSet, ok := convertToType(input, wantType, false); ok {
		setValue(receiver, valueToSet)

		return nil
	}

	if wantType.Kind() == reflect.Struct && input.Kind() == reflect.Map && input.Type().Key().Kind() == reflect.String {

		return setStructFromMap(receiver, input, tags)
	}

	if wantType.Kind() == reflect.Slice && input.Kind() == reflect.Slice {

		return setSlice(receiver, input, tags)
	}

	if wantType.Kind() == reflect.Map && input.Kind() == reflect.Map {

		return setMap(receiver, input, tags)
	}

	return fmt.Errorf(badValueMsg, want, have)
}

func convertToType(input reflect.Value, wantType reflect.Type, convertMapIndexes bool) (reflect.Value, bool) {
	if input.IsValid() {
		if input.Type() == wantType {

			return input, true
		}

		// number to string conversions will produce ASCII values and are not wanted
		if wantType.Kind() != reflect.String && input.CanConvert(wantType) {

			return input.Convert(wantType), true
		}

		if convertMapIndexes && input.Kind() == reflect.String {
			// support reverse string to number conversions for maps with numeric keys converted to strings
			// in JSON representations, etc
			var convertedVal reflect.Value
			stringVar := input.String()
			switch wantType.Kind() {

			case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
				int64Var, err := strconv.ParseInt(stringVar, 10, 64)

				if err != nil {

					return reflect.Value{}, false
				}

				convertedVal = reflect.ValueOf(int64Var)

			case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
				uint64Var, err := strconv.ParseUint(stringVar, 10, 64)

				if err != nil {

					return reflect.Value{}, false
				}

				convertedVal = reflect.ValueOf(uint64Var)

			case reflect.Float64, reflect.Float32:
				float64Var, err := strconv.ParseFloat(stringVar, 64)

				if err != nil {

					return reflect.Value{}, false
				}

				convertedVal = reflect.ValueOf(float64Var)

			default:

				return reflect.Value{}, false
			}

			return convertToType(convertedVal, wantType, false)
		}
	}

	return reflect.Value{}, false
}

func setValue(receiver reflect.Value, input reflect.Value) {
	if receiver.Kind() == reflect.Ptr {
		receiver.Set(reflect.New(receiver.Type().Elem()))
		receiver.Elem().Set(input)
	} else {
		receiver.Set(input)
	}
}
