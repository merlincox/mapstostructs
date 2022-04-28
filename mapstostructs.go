package mapstostructs

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	badReceiversMsg    = "the receivers argument must be a ptr to a slice of struct but a %s was given"
	badReceiverMsg     = "the receiver argument must be a ptr to a struct but a %s was given"
	badFieldMsg        = "the %s field for a struct of type %s must be of type %s but received a value of type %s"
	badFieldMsgWithRow = "%s in row %d"
)

// MapsToStructs provides functionality for a slice of structs to be populated from a slice of map[string]interface{}
// with the option of passing alternative struct tags to use as map keys. If no tags are specified the json tag is
// used and if that is not present, the lowercase value of the struct field is assumed.
//
// The receivers argument must be a pointer to a slice of structs.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
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
	if len(inputMaps) == 0 {
		return nil
	}
	tagMap := makeTagMap(structType, tags)

	for i, inputMap := range inputMaps {
		structValue, err := getStructValue(structType, inputMap, tagMap, tags)
		if err != nil {
			return fmt.Errorf(badFieldMsgWithRow, err.Error(), i+1)
		}
		structValues = reflect.Append(structValues, structValue)
	}
	reflect.ValueOf(receivers).Elem().Set(structValues)

	return nil
}

// MapToStruct provides functionality for a struct to be populated from a map[string]interface{} with the option of
// passing alternative struct tags to use as map keys. If no tags are specified the json tag is used and if that is not
// present, the lowercase value of the struct field is assumed.
//
// The receiver argument must be a pointer to a struct.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
func MapToStruct(inputMap map[string]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(badReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structType := reflect.Indirect(reflect.ValueOf(receiver)).Type()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(badReceiverMsg, "ptr to a "+structType.Kind().String())
	}
	tagMap := makeTagMap(structType, tags)
	structValue, err := getStructValue(structType, inputMap, tagMap, tags)
	if err != nil {

		return err
	}
	reflect.ValueOf(receiver).Elem().Set(structValue)

	return nil
}

func getStructValue(structType reflect.Type, inputMap map[string]interface{}, tagMap map[string]string, tags []string) (reflect.Value, error) {
	structValue := reflect.Indirect(reflect.New(structType))
	for key, mapValue := range inputMap {
		if fieldName, ok := tagMap[strings.ToLower(key)]; ok {
			err := setStructField(structValue.Addr().Interface(), fieldName, mapValue, tags)
			if err != nil {
				return reflect.Value{}, err
			}
		}
	}
	return structValue, nil
}

func makeTagMap(structType reflect.Type, tags []string) map[string]string {
	numFields := structType.NumField()
	tagMap := make(map[string]string, numFields)
	tags = append(tags, "json")
	for i := 0; i < numFields; i++ {
		field := structType.Field(i)
		var tagged bool
		for _, tagName := range tags {
			tag, ok := field.Tag.Lookup(tagName)
			if ok {
				tagMap[tag] = field.Name
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

func setStructField(object interface{}, fieldName string, mapValue interface{}, tags []string) error {
	field := reflect.ValueOf(object).Elem().FieldByName(fieldName)
	value := reflect.ValueOf(mapValue)
	structName := reflect.ValueOf(object).Elem().Type().Name()

	return setStructFieldRecursively(field, value, structName, fieldName, tags)
}

func setStructFieldRecursively(field reflect.Value, value reflect.Value, structName, fieldName string, tags []string) error {
	if value.Type().Kind() == reflect.Ptr {

		return setStructFieldRecursively(field, value.Elem(), structName, fieldName, tags)
	}
	have := value.Kind().String()
	wantType := field.Type()
	if wantType.Kind() == reflect.Ptr {
		wantType = wantType.Elem()
	}
	want := wantType.Kind().String()

	if value.Type() == wantType {
		setField(field, value)

		return nil
	}

	if value.CanConvert(wantType) {
		setField(field, value.Convert(wantType))

		return nil
	}

	if wantType.Kind() == reflect.Struct {
		if innerMap, ok := value.Interface().(map[string]interface{}); ok {
			tagMap := makeTagMap(wantType, tags)
			structValue, err := getStructValue(wantType, innerMap, tagMap, tags)
			if err != nil {

				return err
			}
			setField(field, structValue)

			return nil
		}
	}

	return fmt.Errorf(badFieldMsg, fieldName, structName, want, have)
}

func setField(field reflect.Value, value reflect.Value) {
	if field.Type().Kind() == reflect.Ptr {
		field.Set(reflect.New(field.Type().Elem()))
		field.Elem().Set(value)
	} else {
		field.Set(value)
	}
}
