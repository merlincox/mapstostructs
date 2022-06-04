# mapstostructs

Simple utility functions to convert from `[]map[string]interface{}` into a slice of structs or from a `map[string]interface{}` into a struct, or between two maps, with the option to specify alternative tags for the mapping keys.

Type conversions to the struct type are performed where permitted by the `reflect` library. This helps with the situation where integer values have been JSON-unmarshalled into `float64` values in a map.

There is support for `map[string]interface{}` to struct conversions embedded within the map(s).

There is also support for `map[string]interface{}` to `map[`{numeric}`]interface{}` where numeric is of `int`, `int64`, `int32`, `int16`, `int8`, `uint`, `uint64`, `uint32`, `uint16`, `uint8`, `float64` or `float32` type.

This is to support the situation where a map with numeric keys has been converted by JSON unmarshalling into a map with string keys.

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
    // Outputs: 
    // Zhaoliu is male and 19, lives in London and plays football
    // Zhaoliu is male and 19, lives in London and plays football
```

Acknowledgement: the starting point for this code is to be found here (hence the test names):

https://developpaper.com/question/golang-the-method-of-converting-a-map-array-to-a-structure-array-using-reflection-the-code-is-as-follows-how-to-add-the-structure-generated-by-reflection-to-the-array/
