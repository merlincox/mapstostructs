# mapstostructs

A simple utility function to convert from `map[string]interface{}` to a slice of structs, with the option to specify tags for the map keys.

```go

	type User struct {
		ID       int      `json:"id"`
		Name     string   `json:"name"`
		Gender   string   `json:"gender" alias:"sex"`
		Age      int      `json:"age"`
		Sports   []string `json:"sports"`
		Location Location `json:"location"`
	}

	type Location struct {
		Country string `json:"country"`
		City    string `json:"city"`
	}

	maps := []map[string]interface{}{
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

	err := mapstostructs.MapsToStructs(maps, &users)
	if err == nil {
        fmt.Printf("%s is %s and %d, lives in %s and plays %s\n", users[0].Name, users[0].Gender, users[0].Age, users[0].Location.City, users[0].Sports[0])
	}
	
	Outputs: Zhaoliu is male and 19, lives in London and plays football

```

Acknowledgement: the starting point for this code is to be found here:

https://developpaper.com/question/golang-the-method-of-converting-a-map-array-to-a-structure-array-using-reflection-the-code-is-as-follows-how-to-add-the-structure-generated-by-reflection-to-the-array/
