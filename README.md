# mapstostructs

A simple utility function to convert from `[]map[string]interface{}` into a slice of structs or from `map[string]interface{}` into a struct, with the option to specify alternative tags for the mapping keys.

Type conversions to the struct type are performed where permitted by the `reflect` library. This helps with the situation where integer values have been JSON-unmarshalled into `float64` values in a map.

There is support for `map[string]interface{}` to struct conversions embedded within the map(s), limited by the depth of the embedding.

```go
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
	
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "sex": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": map[string]interface{}{
				"nation": "UK",
				"city":    "London",
			}},
		{"id": 56, "name": "Zhangsan", "sex": "male", "age": 37},
		{"id": 7, "name": "Lisi", "sex": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "sex": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users, "alias")
	if err == nil {
		fmt.Printf("%s is %s and %d, lives in %s and plays %s\n", users[0].Name, users[0].Gender, users[0].Age, users[0].Location.City, users[0].Sports[0])
	}

}
	// Outputs: Zhaoliu is male and 19, lives in London and plays football

```

Acknowledgement: the starting point for this code is to be found here (hence the test names):

https://developpaper.com/question/golang-the-method-of-converting-a-map-array-to-a-structure-array-using-reflection-the-code-is-as-follows-how-to-add-the-structure-generated-by-reflection-to-the-array/
