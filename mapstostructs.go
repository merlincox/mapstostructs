package mapstostructs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	notStructSliceReceiversMsg = "the receivers argument must be a ptr to a slice of struct but a %s was given"
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
		return fmt.Errorf(notStructSliceReceiversMsg, reflect.ValueOf(receivers).Kind().String())
	}
	structValues := reflect.Indirect(reflect.ValueOf(receivers))
	if structValues.Kind() != reflect.Slice {
		return fmt.Errorf(notStructSliceReceiversMsg, "ptr to a "+structValues.Kind().String())
	}
	structType := structValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(notStructSliceReceiversMsg, "ptr to a slice of "+structType.Kind().String())
	}

	return setSlice(reflect.ValueOf(receivers).Elem(), structType, reflect.ValueOf(inputMaps), tags)
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
		return fmt.Errorf(notStructReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structType := reflect.Indirect(reflect.ValueOf(receiver)).Type()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(notStructReceiverMsg, "ptr to a "+structType.Kind().String())
	}

	return setStruct(reflect.ValueOf(receiver).Elem(), structType, inputMap, tags)
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
func MapToMap(inputMap map[interface{}]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(notMapReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	mapValue := reflect.Indirect(reflect.ValueOf(receiver))
	if mapValue.Kind() != reflect.Map {
		return fmt.Errorf(notMapReceiverMsg, "ptr to a "+mapValue.Kind().String())
	}

	return setMap(reflect.ValueOf(receiver).Elem(), mapValue.Type(), reflect.ValueOf(inputMap), tags)
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

		return fmt.Errorf(structPrefix+err.Error(), fieldName, structName)
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
	have := value.Interface()
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

	if wantType.Kind() == reflect.Slice && value.Type().Kind() == reflect.Slice {

		return setSlice(receivingValue, wantType.Elem(), value, tags)
	}

	if wantType.Kind() == reflect.Map && value.Type().Kind() == reflect.Map {

		return setMap(receivingValue, wantType, value, tags)
	}

	return fmt.Errorf(badValueMsg, want, have)
}

func setMap(receivingValue reflect.Value, wantType reflect.Type, inputValue reflect.Value, tags []string) error {
	if inputValue.Len() == 0 {

		return nil
	}
	wantKeyType := wantType.Key()
	mapToSet := reflect.MakeMap(wantType)
	mapRange := inputValue.MapRange()

	for mapRange.Next() {
		keyToSet, ok := convertToType(mapRange.Key(), wantKeyType, true)
		if !ok {
			want := wantKeyType.String()
			have := mapRange.Key().Interface()

			return fmt.Errorf(mapKeyPrefix+badValueMsg, wantType.String(), want, have)
		}

		valueToSet := reflect.Indirect(reflect.New(wantType.Elem()))
		if err := setRecursively(valueToSet, mapRange.Value(), tags); err != nil {

			return fmt.Errorf(mapValuePrefix+err.Error(), wantType.String())
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

	// number to string conversions will produce ASCII values and are not wanted
	if wantType.Kind() != reflect.String && value.CanConvert(wantType) {

		return value.Convert(wantType), true
	}

	if value.Type().Kind() == reflect.Interface {

		return convertToType(value.Elem(), wantType, mapIndex)
	}

	if mapIndex && value.Type().Kind() == reflect.String {
		// support reverse string to number conversions for maps with numeric keys converted to strings
		// in JSON representations
		var convertedVal reflect.Value
		stringVar := value.Interface().(string)
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

	return reflect.Value{}, false
}

func setSlice(receivingValue reflect.Value, wantType reflect.Type, inputValue reflect.Value, tags []string) error {
	if inputValue.Len() == 0 {

		return nil
	}
	sliceValues := reflect.MakeSlice(reflect.SliceOf(wantType), 0, inputValue.Len())
	for i := 0; i < inputValue.Len(); i++ {
		valueToSet := reflect.Indirect(reflect.New(wantType))
		if err := setRecursively(valueToSet, inputValue.Index(i), tags); err != nil {

			return fmt.Errorf(err.Error()+rowSuffix, i+1)
		}
		sliceValues = reflect.Append(sliceValues, valueToSet)
	}
	setValue(receivingValue, sliceValues)

	return nil
}

func setValue(receivingValue reflect.Value, value reflect.Value) {
	if receivingValue.Type().Kind() == reflect.Ptr {
		receivingValue.Set(reflect.New(receivingValue.Type().Elem()))
		receivingValue.Elem().Set(value)
	} else {
		receivingValue.Set(value)
	}
}
