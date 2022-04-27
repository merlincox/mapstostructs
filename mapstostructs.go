package mapstostructs

import (
	"fmt"
	"reflect"
	"strings"
)

const (
	badReceiversMsg = "the receivers argument must be a ptr to a slice of struct but a %s was given"
	badReceiverMsg  = "the receiver argument must be a ptr to a struct but a %s was given"
	badFieldMsg     = "the %s field for a struct of type %s must be of type %s but received a value of type %s in row %d"
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
		return fmt.Errorf(badReceiversMsg, reflect.ValueOf(receivers).Kind().String())
	}
	receivingValues := reflect.Indirect(reflect.ValueOf(receivers))
	if receivingValues.Kind() != reflect.Slice {
		return fmt.Errorf(badReceiversMsg, "ptr to a "+receivingValues.Kind().String())
	}
	structType := receivingValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(badReceiversMsg, "ptr to a slice of "+structType.Kind().String())
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
		thisReceiver := reflect.Indirect(reflect.New(structType))
		for key, value := range thisMap {
			if fieldName, ok := tagMap[strings.ToLower(key)]; ok {
				err := setStructField(thisReceiver.Addr().Interface(), fieldName, value, i+1)
				if err != nil {
					return err
				}
			}
		}
		receivingValues = reflect.Append(receivingValues, thisReceiver)
	}
	reflect.ValueOf(receivers).Elem().Set(receivingValues)

	return nil
}

func MapToStruct(inputMap map[string]interface{}, receiver interface{}, tags ...string) error {
	if reflect.ValueOf(receiver).Kind() != reflect.Ptr {
		return fmt.Errorf(badReceiverMsg, reflect.ValueOf(receiver).Kind().String())
	}
	structType := reflect.Indirect(reflect.ValueOf(receiver)).Type()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(badReceiverMsg, "ptr to a slice of "+structType.Kind().String())
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

	thisReceiver := reflect.Indirect(reflect.New(structType))
	for key, value := range inputMap {
		if fieldName, ok := tagMap[strings.ToLower(key)]; ok {
			err := setStructField(thisReceiver.Addr().Interface(), fieldName, value, 0) //row..
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func setStructField(object interface{}, fieldName string, mapValue interface{}, line int) error {
	field := reflect.ValueOf(object).Elem().FieldByName(fieldName)
	value := reflect.ValueOf(mapValue)
	structName := reflect.ValueOf(object).Elem().Type().Name()

	return innerSetStructField(field, value, structName, fieldName, line)
}

func innerSetStructField(field reflect.Value, value reflect.Value, structName, fieldName string, line int) error {
	if value.Type().Kind() == reflect.Ptr {

		return innerSetStructField(field, value.Elem(), structName, fieldName, line)
	}
	have := value.Kind().String()
	wantType := field.Type()
	if field.Type().Kind() == reflect.Ptr {
		wantType = field.Type().Elem()
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

	return fmt.Errorf(badFieldMsg, fieldName, structName, want, have, line)
}

func setField(field reflect.Value, value reflect.Value) {
	if field.Type().Kind() == reflect.Ptr {
		field.Set(reflect.New(field.Type().Elem()))
		field.Elem().Set(value)
	} else {
		field.Set(value)
	}
}
