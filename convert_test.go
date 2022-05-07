package mapstostructs

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertFloat64(t *testing.T) {
	var number float64 = math.MaxFloat64
	numberString := fmt.Sprintf("%f", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, val.Float())
	}

	number = -number
	numberString = fmt.Sprintf("%f", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, val.Float())
	}
}

func TestConvertFloat32(t *testing.T) {
	var number float32 = math.MaxFloat32
	numberString := fmt.Sprintf("%f", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, float32(val.Float()))
	}

	number = -number
	numberString = fmt.Sprintf("%f", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, float32(val.Float()))
	}
}

func TestConvertInt64(t *testing.T) {

	var number int64 = math.MaxInt64
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int64(val.Int()))
	}

	number = math.MinInt64
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int64(val.Int()))
	}
}

func TestConvertInt32(t *testing.T) {

	var number int32 = math.MaxInt32
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int32(val.Int()))
	}

	number = math.MinInt32
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int32(val.Int()))
	}
}

func TestConvertInt16(t *testing.T) {

	var number int16 = math.MaxInt16
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int16(val.Int()))
	}

	number = math.MinInt16
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int16(val.Int()))
	}
}

func TestConvertInt8(t *testing.T) {

	var number int8 = math.MaxInt8
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int8(val.Int()))
	}

	number = math.MinInt8
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int8(val.Int()))
	}
}

func TestConvertInt(t *testing.T) {

	var number int = math.MaxInt
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int(val.Int()))
	}

	number = math.MinInt
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, int(val.Int()))
	}
}

func TestConvertUint64(t *testing.T) {

	var number uint64 = math.MaxUint64
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint64(val.Uint()))
	}

	number = 0
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint64(val.Uint()))
	}
}

func TestConvertUint(t *testing.T) {

	var number uint = math.MaxUint
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint(val.Uint()))
	}

	number = 0
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint(val.Uint()))
	}
}

func TestConvertUint32(t *testing.T) {

	var number uint32 = math.MaxUint32
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint32(val.Uint()))
	}

	number = 0
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint32(val.Uint()))
	}
}

func TestConvertUint16(t *testing.T) {

	var number uint16 = math.MaxUint16
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint16(val.Uint()))
	}

	number = 0
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint16(val.Uint()))
	}
}

func TestConvertUint8(t *testing.T) {

	var number uint8 = math.MaxUint8
	numberString := fmt.Sprintf("%d", number)
	numberType := reflect.TypeOf(number)

	val, ok := convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint8(val.Uint()))
	}

	number = 0
	numberString = fmt.Sprintf("%d", number)

	val, ok = convertToType(reflect.ValueOf(numberString), numberType, true)

	if assert.True(t, ok) {
		assert.Equal(t, number, uint8(val.Uint()))
	}
}
