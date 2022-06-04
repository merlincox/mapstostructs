package mapstostructs

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const (
	badValueMsg    = "must be or be convertible to %s type, but received '%v'"
	structPrefix   = "the %s field for a struct of type %s "
	rowSuffix      = " in row %d"
	mapKeyPrefix   = "the map key for a %s "
	mapValuePrefix = "the map value for a %s "
	jsonTag        = "json"
)

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
				tagMap[strings.ToLower(parts[0])] = field.Name
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
		if fieldName, ok := tagMap[strings.ToLower(mapRange.Key().String())]; ok {
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
