package mapstostructs

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestXX(t *testing.T) {

	var var1 interface{}
	var var2 string

	var1 = "test"

	val1 := reflect.ValueOf(var1)
	val2 := reflect.ValueOf(var2)

	typ1 := val1.Type()
	typ2 := val2.Type()

	if typ1 == typ2 {
		fmt.Println("a")
		return
	}

	if val1.CanConvert(typ2) {
		fmt.Println("b")
		return
	}

	fmt.Println("c")
}

func TestSetMap(t *testing.T) {

	var receiver map[int]string

	input := make(map[string]interface{})

	err := MapToMap(input, &receiver)

	assert.Nil(t, err)

	input["5"] = "test1"

	err = MapToMap(input, &receiver)

	if assert.Nil(t, err) {
		assert.Equal(t, input["5"], receiver[5])
	}

	input["dog"] = "test2"

	err = MapToMap(input, &receiver)

	if assert.NotNil(t, err) {
		assert.Equal(t, "the map key must be convertible to int but received 'dog'", err.Error())
	}

	delete(input, "dog")

	err = MapToMap(input, &receiver)

	assert.Nil(t, err)

	input["10"] = 65

	err = MapToMap(input, &receiver)

	if assert.NotNil(t, err) {
		assert.Equal(t, "the %s field for a struct of type %s must be of type string but received a value of type interface {}", err.Error())
	}

}
