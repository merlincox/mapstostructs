package mapstostructs

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	badReceiverMsg = "the receivers argument must be a ptr to a slice of struct but a %s was given"
	badFieldMsg    = "the %s field for a struct of type %s must be of type %s but received a value of type %s in row %d"
)

// MapsToStructs provides functionality for a slice of structs to be generated from a slice of map[string]interface{}
// with the option of passing alternative struct tags to use as map keys. If no tags are specified the json tag is
// used and if that is not present, the lowercase value of the struct field is assumed.
//
// The receivers argument must be a pointer to a slice of structs.
//
// Type conversions to the struct type are performed where permitted by the reflect library. This helps with the
// situation where integer values have been JSON-unmarshalled into float64 values in a map.
func MapsToStructs(inputMaps []map[string]interface{}, receivers interface{}, tags ...string) error {
	if reflect.ValueOf(receivers).Kind() != reflect.Ptr {
		return fmt.Errorf(badReceiverMsg, reflect.ValueOf(receivers).Kind().String())
	}
	receivingValues := reflect.Indirect(reflect.ValueOf(receivers))
	if receivingValues.Kind() != reflect.Slice {
		return fmt.Errorf(badReceiverMsg, "ptr to a "+receivingValues.Kind().String())
	}
	structType := receivingValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(badReceiverMsg, "ptr to a slice of "+structType.Kind().String())
	}
	if len(inputMaps) == 0 {
		return nil
	}
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

	for i, thisMap := range inputMaps {
		thisValue := reflect.Indirect(reflect.New(structType))
		for key, value := range thisMap {
			if fieldName, ok := tagMap[strings.ToLower(key)]; ok {
				err := setStructField(thisValue.Addr().Interface(), fieldName, value, i+1)
				if err != nil {
					return err
				}
			}
		}
		receivingValues = reflect.Append(receivingValues, thisValue)
	}
	reflect.ValueOf(receivers).Elem().Set(receivingValues)

	return nil
}

func setStructField(object interface{}, fieldName string, mapValue interface{}, line int) error {
	field := reflect.ValueOf(object).Elem().FieldByName(fieldName)
	value := reflect.ValueOf(mapValue)
	want := field.Kind().String()
	have := value.Kind().String()
	structName := reflect.ValueOf(object).Elem().Type().Name()

	if field.Type().Kind() != reflect.Ptr {
		if field.Type() == value.Type() {
			field.Set(value)

			return nil
		}
		if value.CanConvert(field.Type()) {
			field.Set(value.Convert(field.Type()))

			return nil
		}
	} else {
		want = field.Type().Elem().Kind().String()
		if value.Type().Kind() == reflect.Ptr {
			have = value.Elem().Kind().String()
			if field.Type().Elem() == value.Elem().Type() {
				field.Set(reflect.New(field.Type().Elem()))
				field.Elem().Set(value.Elem())

				return nil
			}
			if value.Elem().CanConvert(field.Type().Elem()) {
				field.Set(reflect.New(field.Type().Elem()))
				field.Elem().Set(value.Elem().Convert(field.Type().Elem()))

				return nil
			}
		} else {
			if field.Type().Elem() == value.Type() {
				field.Set(reflect.New(field.Type().Elem()))
				field.Elem().Set(value)

				return nil
			}
			if value.CanConvert(field.Type().Elem()) {
				field.Set(reflect.New(field.Type().Elem()))
				field.Elem().Set(value.Convert(field.Type().Elem()))

				return nil
			}
		}
	}

	return fmt.Errorf(badFieldMsg, fieldName, structName, want, have, line)
}
