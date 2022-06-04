package mapstostructs_test

import (
	"encoding/json"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/merlincox/mapstostructs"
)

type Outer struct {
	IntVal     int
	StringVal  string
	PStringVal *string
	PIntVal    *int
	IFaceVal   interface{}
	SliceVal   []string
	PSliceVal  []*string
	StructVal  Inner
	PStructVal *Inner
	StringMap  map[string]Inner
	IntMap     map[int]Inner
}

func (o Outer) Equals(o2 Outer) bool {
	if o.IntVal != o2.IntVal {
		return false
	}
	if o.StringVal != o2.StringVal {
		return false
	}
	if (o.PStringVal == nil) != (o2.PStringVal == nil) {
		return false
	}
	if o.PStringVal != nil && *o.PStringVal != *o2.PStringVal {
		return false
	}
	if (o.PIntVal == nil) != (o2.PIntVal == nil) {
		return false
	}
	if o.PIntVal != nil && *o.PIntVal != *o2.PIntVal {
		return false
	}
	if len(o.SliceVal) != len(o2.SliceVal) {
		return false
	}
	for i := range o.SliceVal {
		if o.SliceVal[i] != o2.SliceVal[i] {
			return false
		}
	}
	if len(o.PSliceVal) != len(o2.PSliceVal) {
		return false
	}
	for i := range o.PSliceVal {
		if *o.PSliceVal[i] != *o2.PSliceVal[i] {
			return false
		}
	}
	if !o.StructVal.Equals(o2.StructVal) {
		return false
	}
	if !o.PStructVal.Equals(*o2.PStructVal) {
		return false
	}
	if len(o.StringMap) != len(o2.StringMap) {
		return false
	}
	for key, val := range o.StringMap {
		if val2, ok := o2.StringMap[key]; !ok || !val.Equals(val2) {
			return false
		}
	}
	if len(o.IntMap) != len(o2.IntMap) {
		return false
	}
	for key, val := range o.IntMap {
		if val2, ok := o2.IntMap[key]; !ok || !val.Equals(val2) {
			return false
		}
	}
	//JSON coding converts int to float, which we will not treat as an error..
	int1, ok1 := o.IFaceVal.(int)
	int2, ok2 := o2.IFaceVal.(int)
	float2, ok2a := o2.IFaceVal.(float64)
	if ok1 != ok2 && ok1 != ok2a {
		return false
	}
	if ok2a {
		int2 = int(float2)
	}
	if ok1 && (int1 != int2) {
		return false
	}

	string1, ok1 := o.IFaceVal.(string)
	string2, ok2 := o2.IFaceVal.(string)
	if ok1 != ok2 {
		return false
	}
	if ok1 && (string1 != string2) {
		return false
	}

	return true
}

func (o Inner) Equals(o2 Inner) bool {
	if o.StringVal1 != o2.StringVal1 {
		return false
	}
	if o.StringVal2 != o2.StringVal2 {
		return false
	}
	return true
}

type Inner struct {
	StringVal1 string
	StringVal2 string
}

var (
	outerStringMapSlice []map[string]interface{}
	outerStringMap      map[string]interface{}
	outerIntMap         map[int]interface{}
	outerIntStringMap   map[string]interface{}
	outers              []Outer
	outer               Outer
)

func init() {
	rand.Seed(time.Now().UnixNano())
	outers = make([]Outer, 10000)
	outerIntStringMap = make(map[string]interface{})
	outerIntMap = make(map[int]interface{})
	for i := 0; i < len(outers); i++ {
		outers[i] = randomOuter()
		outerIntMap[i] = outers[i]
		outerIntStringMap[strconv.Itoa(i)] = outers[i]
	}
	data, _ := json.Marshal(outers)
	_ = json.Unmarshal(data, &outerStringMapSlice)
	outer = outers[0]
	data, _ = json.Marshal(outer)
	_ = json.Unmarshal(data, &outerStringMap)
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func randomStrings(n1, n2 int) []string {
	out := make([]string, rand.Intn(n1))
	for i := 0; i < len(out); i++ {
		out[i] = randomString(n2)
	}
	return out
}

func randomPStrings(n1, n2 int) []*string {
	out := make([]*string, rand.Intn(n1))
	for i := 0; i < len(out); i++ {
		r := randomString(n2)
		out[i] = &r
	}
	return out
}

func randomInner() Inner {
	return Inner{
		StringVal1: randomString(10),
		StringVal2: randomString(10),
	}
}

func randomOuter() Outer {

	intMap := make(map[int]Inner, 20)
	stringMap := make(map[string]Inner, 20)

	for i := 0; i < 20; i++ {
		intMap[rand.Intn(100)] = randomInner()
		stringMap[randomString(10)] = randomInner()
	}

	pStringVal := randomString(10)
	pIntVal := rand.Intn(10000000)
	pStructVal := randomInner()

	var ifaceVal interface{}
	if rand.Intn(1) == 0 {
		ifaceVal = rand.Intn(10)
	} else {
		ifaceVal = randomString(10)
	}

	return Outer{
		IntVal:     rand.Intn(10000000),
		StringVal:  randomString(10),
		PStringVal: &pStringVal,
		PIntVal:    &pIntVal,
		IFaceVal:   ifaceVal,
		SliceVal:   randomStrings(10, 10),
		PSliceVal:  randomPStrings(10, 10),
		StructVal:  randomInner(),
		PStructVal: &pStructVal,
		IntMap:     intMap,
		StringMap:  stringMap,
	}
}

type slicesFunc func([]map[string]interface{}, interface{}, ...string) error

func innerBenchmarkMapsToStructs(b *testing.B, fn slicesFunc) {
	var receiver []Outer
	for i := 0; i < b.N; i++ {
		if err := fn(outerStringMapSlice, &receiver); err != nil {
			b.Fail()
		}
		if len(outers) != len(receiver) {
			b.Fail()
		}
		for i := range outers {
			if !outers[i].Equals(receiver[i]) {
				b.Fail()
			}
		}
	}
}

func BenchmarkMapsToStructs(b *testing.B) {
	innerBenchmarkMapsToStructs(b, mapstostructs.MapsToStructs)
}

func BenchmarkMapsToStructsJSON(b *testing.B) {
	innerBenchmarkMapsToStructs(b, jsonMapsToStructs)
}

type structFunc func(map[string]interface{}, interface{}, ...string) error

func innerBenchmarkMapToStruct(b *testing.B, fn structFunc) {
	var receiver Outer
	for i := 0; i < b.N; i++ {
		if err := fn(outerStringMap, &receiver); err != nil {
			b.Fail()
		}
		if !outer.Equals(receiver) {
			b.Fail()
		}
	}
}

func BenchmarkMapToStruct(b *testing.B) {
	innerBenchmarkMapToStruct(b, mapstostructs.MapToStruct)
}

func BenchmarkMapToStructJSON(b *testing.B) {
	innerBenchmarkMapToStruct(b, jsonMapToStruct)
}

type mapFunc func(interface{}, interface{}, ...string) error

func innerBenchmarkMapToMap(b *testing.B, fn mapFunc) {
	var receiver map[int]Outer
	for i := 0; i < b.N; i++ {
		if err := fn(outerIntStringMap, &receiver); err != nil {
			b.Fail()
		}
		if len(receiver) != len(outerIntStringMap) {
			b.Fail()
		}
		for key := range outerIntMap {
			val, ok := outerIntMap[key].(Outer)
			if !ok {
				b.Fail()
			}
			if val2, ok := receiver[key]; !ok || !val.Equals(val2) {
				b.Fail()
			}
		}
	}
}

func BenchmarkMapToMap(b *testing.B) {
	innerBenchmarkMapToMap(b, mapstostructs.MapToMap)
}

func BenchmarkMapToMapJSON(b *testing.B) {
	innerBenchmarkMapToMap(b, jsonMapToMap)
}

func jsonMapsToStructs(input []map[string]interface{}, receiver interface{}, tags ...string) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, receiver)
}

func jsonMapToStruct(input map[string]interface{}, receiver interface{}, tags ...string) error {
	raw, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, receiver)
}

func jsonMapToMap(intStringMap interface{}, receiver interface{}, tags ...string) error {
	data, err := json.Marshal(intStringMap)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, receiver)
}
