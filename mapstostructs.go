package mapstostructs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	badReceiversMsg    = "the receivers argument must be a ptr to a slice of struct but a %s was given"
	badReceiverMsg     = "the receiver argument must be a ptr to a struct but a %s was given"
	badFieldMsg        = "the %%s field for a struct of type %%s must be of type %s but received a value of type %s"
	badFieldMsgWithRow = "%s in row %d"

	jsonTag = "json"
)

// MapsToStructs provides functionality for a slice of structs to be populated from a slice of map[string]interface{}
// with the option of passing alternative struct tags to use as map keys. If no tags are specified the json tag is
// used and if that is not present, the lowercase value of the struct field is assumed.
//
// The receivers argument must be a pointer to a slice of structs.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
//
// Conversion of map[string]interface() to struct embedded within the slice of map[string]interface{} is permitted.
//
// Maps with numeric keys will accept string representations of numeric values.
func MapsToStructs(inputMaps []map[string]interface{}, receivers interface{}, tags ...string) error {
	if reflect.ValueOf(receivers).Kind() != reflect.Ptr {
		return fmt.Errorf(badReceiversMsg, reflect.ValueOf(receivers).Kind().String())
	}
	structValues := reflect.Indirect(reflect.ValueOf(receivers))
	if structValues.Kind() != reflect.Slice {
		return fmt.Errorf(badReceiversMsg, "ptr to a "+structValues.Kind().String())
	}
	structType := structValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(badReceiversMsg, "ptr to a slice of "+structType.Kind().String())
	}

	return setSlice(reflect.ValueOf(receivers).Elem(), structType, inputMaps, tags)
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
func MapToStruct(inputMap map[string]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(badReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structType := reflect.Indirect(reflect.ValueOf(receiver)).Type()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(badReceiverMsg, "ptr to a "+structType.Kind().String())
	}

	return setStruct(reflect.ValueOf(receiver).Elem(), structType, inputMap, tags)
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

func setStruct(receivingValue reflect.Value, wantType reflect.Type, inputMap map[string]interface{}, tags []string) error {
	tagMap := makeTagMap(wantType, tags)
	structValue := reflect.Indirect(reflect.New(wantType))
	for key, mapValue := range inputMap {
		if mapValue != nil {
			if fieldName, ok := tagMap[key]; ok {
				err := setStructField(structValue.Addr().Interface(), fieldName, mapValue, tags)
				if err != nil {
					return err
				}
			}
		}
	}
	setValue(receivingValue, structValue)

	return nil
}

func setStructField(object interface{}, fieldName string, mapValue interface{}, tags []string) error {
	field := reflect.ValueOf(object).Elem().FieldByName(fieldName)
	value := reflect.ValueOf(mapValue)
	structName := reflect.ValueOf(object).Elem().Type().Name()

	if err := setRecursively(field, value, tags); err != nil {

		return fmt.Errorf(err.Error(), fieldName, structName)
	}

	return nil
}

func setRecursively(receivingValue reflect.Value, value reflect.Value, tags []string) error {
	if value.Kind() == reflect.Invalid {

		return nil
	}
	if value.Type().Kind() == reflect.Ptr {

		return setRecursively(receivingValue, value.Elem(), tags)
	}
	have := value.Type().String()
	wantType := receivingValue.Type()
	if wantType.Kind() == reflect.Ptr {
		wantType = wantType.Elem()
	}
	want := wantType.String()

	if valueToSet, ok := convertToType(value, wantType, false); ok {
		setValue(receivingValue, valueToSet)

		return nil
	}

	if wantType.Kind() == reflect.Struct {
		if inputMap, ok := value.Interface().(map[string]interface{}); ok {

			return setStruct(receivingValue, wantType, inputMap, tags)
		}
	}

	if wantType.Kind() == reflect.Slice {
		if inputMaps, ok := interfacesToMapInterfaces(value); ok {

			return setSlice(receivingValue, wantType.Elem(), inputMaps, tags)
		}
	}

	if wantType.Kind() == reflect.Map && value.Type().Kind() == reflect.Map {

		return setMap(receivingValue, wantType, value, tags)
	}

	return fmt.Errorf(badFieldMsg, want, have)
}

func setMap(receivingValue reflect.Value, wantType reflect.Type, inputValue reflect.Value, tags []string) error {
	wantKeyType := wantType.Key()
	if inputValue.Len() == 0 {

		return nil
	}
	mapToSet := reflect.MakeMap(wantType)
	mapRange := inputValue.MapRange()

	for mapRange.Next() {
		keyToSet, ok := convertToType(mapRange.Key(), wantKeyType, true)
		if !ok {
			want := wantType.String()
			have := inputValue.Type().String()

			return fmt.Errorf(badFieldMsg, want, have)
		}

		valueToSet := reflect.Indirect(reflect.New(wantType.Elem()))
		if err := setRecursively(valueToSet, mapRange.Value(), tags); err != nil {

			return err
		}
		mapToSet.SetMapIndex(keyToSet, valueToSet)
	}
	setValue(receivingValue, mapToSet)

	return nil
}

func convertToType(value reflect.Value, wantType reflect.Type, mapIndex bool) (reflect.Value, bool) {
	if value.Type() == wantType {

		return value, true
	}
	if value.CanConvert(wantType) {

		return value.Convert(wantType), true
	}

	if mapIndex && value.Type().Kind() == reflect.String {
		// support reverse string to number conversions for maps with numeric keys converted to strings in JSON representations
		var (
			convertedVal reflect.Value
			err          error
			int64val     int64
			uint64val    uint64
			float64val   float64
		)

		if isInt(wantType) {
			int64val, err = strconv.ParseInt(value.Interface().(string), 10, 64)

			if err == nil {
				convertedVal = reflect.ValueOf(int64val)
			}
		}

		if isUint(wantType) {
			uint64val, err = strconv.ParseUint(value.Interface().(string), 10, 64)

			if err == nil {
				convertedVal = reflect.ValueOf(uint64val)
			}
		}

		if isFloat(wantType) {
			float64val, err = strconv.ParseFloat(value.Interface().(string), 64)

			if err == nil {
				convertedVal = reflect.ValueOf(float64val)
			}
		}

		if err != nil {

			return reflect.Value{}, false
		}

		return convertToType(convertedVal, wantType, false)
	}

	return reflect.Value{}, false
}

func isFloat(wantType reflect.Type) bool {

	return wantType.Kind() >= reflect.Float32 && wantType.Kind() <= reflect.Float64
}

func isInt(wantType reflect.Type) bool {

	return wantType.Kind() >= reflect.Int && wantType.Kind() <= reflect.Int64
}

func isUint(wantType reflect.Type) bool {

	return wantType.Kind() >= reflect.Uint && wantType.Kind() <= reflect.Uint64
}

func setSlice(receivingValue reflect.Value, wantType reflect.Type, inputMaps []map[string]interface{}, tags []string) error {
	sliceValues := reflect.MakeSlice(reflect.SliceOf(wantType), 0, len(inputMaps))
	for i, inputMap := range inputMaps {
		valueToSet := reflect.Indirect(reflect.New(wantType))
		if err := setRecursively(valueToSet, reflect.ValueOf(inputMap), tags); err != nil {

			return fmt.Errorf(badFieldMsgWithRow, err.Error(), i+1)
		}
		sliceValues = reflect.Append(sliceValues, valueToSet)
	}
	setValue(receivingValue, sliceValues)

	return nil
}

func interfacesToMapInterfaces(value reflect.Value) ([]map[string]interface{}, bool) {
	inputs, ok := value.Interface().([]interface{})
	if !ok {
		return nil, false
	}
	outMaps := make([]map[string]interface{}, len(inputs))
	for i := range inputs {
		outMaps[i], ok = inputs[i].(map[string]interface{})
		if !ok {
			return nil, false
		}
	}
	return outMaps, true
}

func setValue(receivingValue reflect.Value, value reflect.Value) {
	if receivingValue.Type().Kind() == reflect.Ptr {
		receivingValue.Set(reflect.New(receivingValue.Type().Elem()))
		receivingValue.Elem().Set(value)
	} else {
		receivingValue.Set(value)
	}
}
