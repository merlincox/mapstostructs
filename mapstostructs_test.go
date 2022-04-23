package mapstostructs_test

import (
	"mapstostructs"
	"testing"
)

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender"`
	Age    int    `json:"age"`
}

func TestMapsToStructsSimple(t *testing.T) {

	maps := []map[string]interface{}{
		{"Id": 213, "name": "Zhaoliu", "gender": "male", "age": 19},
		{"Id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"Id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"Id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users)

	if err != nil {
		t.Fail()
	}

}
