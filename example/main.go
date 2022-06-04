package main

import (
	"fmt"

	"github.com/merlincox/mapstostructs"
)

type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Gender   string   `json:"gender" alias:"sex"`
	Age      int      `json:"age"`
	Sports   []string `json:"sports"`
	Location Location `json:"location"`
}

type Location struct {
	Country string `json:"country" alias:"nation"`
	City    string `json:"city"`
}

func main() {

	// example with embedded map and aliased field names
	withMap := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "sex": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": map[string]interface{}{
				"nation": "UK",
				"city":   "London",
			}},
		{"id": 56, "name": "Zhangsan", "sex": "male", "age": 37},
		{"id": 7, "name": "Lisi", "sex": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "sex": "male", "age": 28},
	}

	// example with embedded struct and aliased field name
	withStruct := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "sex": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": Location{
				Country: "UK",
				City:    "London",
			}},
		{"id": 56, "name": "Zhangsan", "sex": "male", "age": 37},
		{"id": 7, "name": "Lisi", "sex": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "sex": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(withMap, &users, "alias")
	if err == nil {
		fmt.Printf("%s is %s and %d, lives in %s and plays %s\n", users[0].Name, users[0].Gender, users[0].Age, users[0].Location.City, users[0].Sports[0])
	}
	err = mapstostructs.MapsToStructs(withStruct, &users, "alias")
	if err == nil {
		fmt.Printf("%s is %s and %d, lives in %s and plays %s\n", users[0].Name, users[0].Gender, users[0].Age, users[0].Location.City, users[0].Sports[0])
	}
}
