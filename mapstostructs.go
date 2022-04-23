package mapstostructs

import (
	"fmt"
	"reflect"
	"strings"
)

// MapsToStructs provides functionality for a slice of structs to generated from a slice of map[string]interface{}
func MapsToStructs(maps []map[string]interface{}, receivers interface{}, tags ...string) error {
	errorMsg := "receivers argument must be a ptr to a slice of struct but a %s was given"
	if reflect.ValueOf(receivers).Kind() != reflect.Ptr {
		return fmt.Errorf(errorMsg, reflect.ValueOf(receivers).Kind().String())
	}
	receivingValues := reflect.Indirect(reflect.ValueOf(receivers))
	if receivingValues.Kind() != reflect.Slice {
		return fmt.Errorf(errorMsg, "ptr to a "+receivingValues.Kind().String())
	}
	structType := receivingValues.Type().Elem()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf(errorMsg, "ptr to a slice of "+structType.Kind().String())
	}
	tagMap := make(map[string]string, structType.NumField())
	tags = append(tags, "json")
	for i := 0; i < structType.NumField(); i++ {
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

	for i, thisMap := range maps {
		thisVal := reflect.Indirect(reflect.New(structType))
		for key, value := range thisMap {
			if fieldName, ok := tagMap[strings.ToLower(key)]; ok {
				err := setStructField(thisVal.Addr().Interface(), fieldName, value, i+1)
				if err != nil {
					return err
				}
			}
		}
		receivingValues = reflect.Append(receivingValues, thisVal)
	}
	reflect.ValueOf(receivers).Elem().Set(receivingValues)

	return nil
}

func setStructField(obj interface{}, fieldName string, value interface{}, line int) error {
	rField := reflect.ValueOf(obj).Elem().FieldByName(fieldName)
	rValue := reflect.ValueOf(value)

	if rField.Type() != rValue.Type() {
		return fmt.Errorf(
			"the %s field for a %s must be of type %s but a value of type %s was given on row %d",
			fieldName,
			reflect.ValueOf(obj).Elem().Type().Name(),
			rField.Kind().String(),
			rValue.Kind().String(),
			line,
		)
	}
	rField.Set(rValue)

	return nil
}
