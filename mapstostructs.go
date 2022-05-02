package mapstostructs

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	badReceiversMsg    = "the receivers argument must be a ptr to a slice of struct but a %s was given"
	badReceiverMsg     = "the receiver argument must be a ptr to a struct but a %s was given"
	badFieldMsg        = "the %%s field for a struct of type %%s must be of type %s but received a value of type %s"
	badFieldMsgOld     = "the %s field for a struct of type %s must be of type %s but received a value of type %s"
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
// Conversion of map[string]interface() to struct embedded within the slice of map[string]interface{} is permitted to a
// limited depth.
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
// Conversion of map[string]interface() to struct embedded within the map[string]interface{} is permitted to a limited
// depth.
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
		if mapValue != nil {
			if fieldName, ok := tagMap[strings.ToLower(key)]; ok {
				err := setStructField(structValue.Addr().Interface(), fieldName, mapValue, tags)
				if err != nil {
					return reflect.Value{}, err
				}
			}
		}
	}
	return structValue, nil
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

func setStructField(object interface{}, fieldName string, mapValue interface{}, tags []string) error {
	field := reflect.ValueOf(object).Elem().FieldByName(fieldName)
	value := reflect.ValueOf(mapValue)
	structName := reflect.ValueOf(object).Elem().Type().Name()

	if err := setRecursively(field, value, tags); err != nil {

		return fmt.Errorf(err.Error(), fieldName, structName)
	}

	return nil
}

func setStructFieldOld(object interface{}, fieldName string, mapValue interface{}, tags []string) error {
	field := reflect.ValueOf(object).Elem().FieldByName(fieldName)
	value := reflect.ValueOf(mapValue)
	structName := reflect.ValueOf(object).Elem().Type().Name()

	return setStructFieldRecursively(field, value, structName, fieldName, tags)
}

func setStructFieldRecursively(field reflect.Value, value reflect.Value, structName, fieldName string, tags []string) error {
	if value.Kind() == reflect.Invalid {

		return nil
	}
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
		setValue(field, value)

		return nil
	}

	if value.CanConvert(wantType) {
		setValue(field, value.Convert(wantType))

		return nil
	}

	if wantType.Kind() == reflect.Struct {
		if innerMap, ok := value.Interface().(map[string]interface{}); ok {
			tagMap := makeTagMap(wantType, tags)
			structValue, err := getStructValue(wantType, innerMap, tagMap, tags)
			if err != nil {

				return err
			}
			setValue(field, structValue)

			return nil
		}
	}

	if wantType.Kind() == reflect.Slice {
		if inputMaps, ok := interfacesToMapInterfaces(value); ok {

			return setSlice(field, wantType.Elem(), inputMaps, tags)
		}
	}

	if wantType.Kind() == reflect.Map && wantType.Key().Kind() == reflect.String {
		mapToSet := reflect.MakeMap(wantType)

		if innerMap, ok := value.Interface().(map[string]interface{}); ok {
			for key, mapValue := range innerMap {

				valueToSet := reflect.Indirect(reflect.New(wantType.Elem()))

				setStructFieldRecursively(valueToSet, reflect.ValueOf(mapValue), structName, fieldName, tags)

				mapToSet.SetMapIndex(reflect.ValueOf(key), valueToSet)
			}
			setValue(field, mapToSet)
		}

		return nil
	}

	return fmt.Errorf(badFieldMsgOld, fieldName, structName, want, have)
}

func setRecursively(field reflect.Value, value reflect.Value, tags []string) error {
	if value.Kind() == reflect.Invalid {

		return nil
	}
	if value.Type().Kind() == reflect.Ptr {

		return setRecursively(field, value.Elem(), tags)
	}
	have := value.Kind().String()
	wantType := field.Type()
	if wantType.Kind() == reflect.Ptr {
		wantType = wantType.Elem()
	}
	want := wantType.Kind().String()

	if value.Type() == wantType {
		setValue(field, value)

		return nil
	}

	if value.CanConvert(wantType) {
		setValue(field, value.Convert(wantType))

		return nil
	}

	if wantType.Kind() == reflect.Struct {
		if innerMap, ok := value.Interface().(map[string]interface{}); ok {
			tagMap := makeTagMap(wantType, tags)
			structValue, err := getStructValue(wantType, innerMap, tagMap, tags)
			if err != nil {

				return err
			}
			setValue(field, structValue)

			return nil
		}
	}

	if wantType.Kind() == reflect.Slice {
		if inputMaps, ok := interfacesToMapInterfaces(value); ok {

			return setSlice(field, wantType.Elem(), inputMaps, tags)
		}
	}

	if wantType.Kind() == reflect.Map && wantType.Key().Kind() == reflect.String {
		mapToSet := reflect.MakeMap(wantType)

		if innerMap, ok := value.Interface().(map[string]interface{}); ok {
			for key, mapValue := range innerMap {

				valueToSet := reflect.Indirect(reflect.New(wantType.Elem()))

				if err := setRecursively(valueToSet, reflect.ValueOf(mapValue), tags); err != nil {

					return err
				}

				mapToSet.SetMapIndex(reflect.ValueOf(key), valueToSet)
			}
			setValue(field, mapToSet)
		}

		return nil
	}

	return fmt.Errorf(badFieldMsg, want, have)
}

func setSlice(receivingValue reflect.Value, wantType reflect.Type, inputMaps []map[string]interface{}, tags []string) error {
	if len(inputMaps) == 0 {
		return nil
	}
	tagMap := makeTagMap(wantType, tags)
	sliceValues := reflect.MakeSlice(reflect.SliceOf(wantType), 0, len(inputMaps))

	for i, inputMap := range inputMaps {
		structValue, err := getStructValue(wantType, inputMap, tagMap, tags)
		if err != nil {
			return fmt.Errorf(badFieldMsgWithRow, err.Error(), i+1)
		}
		sliceValues = reflect.Append(sliceValues, structValue)
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
