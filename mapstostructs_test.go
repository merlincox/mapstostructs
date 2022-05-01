package mapstostructs_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/merlincox/mapstostructs"
)

type User struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Gender   string   `json:"gender"`
	Age      int      `json:"age"`
	Sports   []string `json:"sports"`
	Location Location `json:"location"`
}

type Location struct {
	Country string `json:"country"`
	City    string `json:"city"`
}

type UserWithTags struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Gender string `json:"gender" alias:"sex"`
	Age    int
}

type UserWithPointers struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Gender   string    `json:"gender"`
	Age      *int      `json:"age"`
	Sports   *[]string `json:"sports"`
	Location *Location `json:"location"`
}

type Recursor1 struct {
	Simple1  Recursor2   `tag1:"simple1"`
	Pointer1 *Recursor2  `tag1:"pointer1,omitempty"`
	Slice1   []Recursor2 `tag1:"slice1,omitempty"`
}

type Recursor2 struct {
	Simple2  Recursor3   `tag2:"simple2"`
	Pointer2 *Recursor3  `tag2:"pointer2,omitempty"`
	Slice2   []Recursor3 `tag2:"slice2,omitempty"`
}

type Recursor3 struct {
	Field3 string `tag3:"field3"`
}

func TestRecurse(t *testing.T) {
	in := Recursor1{
		Simple1: Recursor2{
			Simple2: Recursor3{
				Field3: "val1",
			},
			Pointer2: &Recursor3{
				Field3: "val2",
			},
			Slice2: []Recursor3{
				{
					Field3: "val3",
				},
				{
					Field3: "val4",
				},
			},
		},
		Pointer1: &Recursor2{
			Simple2: Recursor3{
				Field3: "val11",
			},
			Pointer2: &Recursor3{
				Field3: "val12",
			},
			Slice2: []Recursor3{
				{
					Field3: "val13",
				},
				{
					Field3: "val14",
				},
			},
		},
		Slice1: []Recursor2{
			{
				Simple2: Recursor3{
					Field3: "val15",
				},
				Pointer2: &Recursor3{
					Field3: "val16",
				},
				Slice2: []Recursor3{
					{
						Field3: "val17",
					},
					{
						Field3: "val18",
					},
				},
			},
		},
	}

	slice := make(map[string]interface{})
	_ = json.Unmarshal(jsonMarshal(in), &slice)

	var out Recursor1

	err := mapstostructs.MapToStruct(slice, &out, "tag1", "tag2", "tag3")

	if assert.Nil(t, err) {
		assert.Equal(t, in.Simple1.Simple2.Field3, out.Simple1.Simple2.Field3)
		assert.Equal(t, in.Simple1.Pointer2.Field3, out.Simple1.Pointer2.Field3)
		if assert.Equal(t, len(in.Simple1.Slice2), len(out.Simple1.Slice2)) {
			assert.Equal(t, in.Simple1.Slice2[0].Field3, out.Simple1.Slice2[0].Field3)
			assert.Equal(t, in.Simple1.Slice2[1].Field3, out.Simple1.Slice2[1].Field3)
		}
		assert.Equal(t, in.Pointer1.Simple2.Field3, out.Pointer1.Simple2.Field3)
		assert.Equal(t, in.Pointer1.Pointer2.Field3, out.Pointer1.Pointer2.Field3)
		if assert.Equal(t, len(in.Pointer1.Slice2), len(out.Pointer1.Slice2)) {
			assert.Equal(t, in.Pointer1.Slice2[0].Field3, out.Pointer1.Slice2[0].Field3)
			assert.Equal(t, in.Pointer1.Slice2[1].Field3, out.Pointer1.Slice2[1].Field3)
		}
		if assert.Equal(t, len(in.Slice1), len(out.Slice1)) {
			assert.Equal(t, in.Slice1[0].Simple2.Field3, out.Slice1[0].Simple2.Field3)
			if assert.Equal(t, len(in.Slice1[0].Slice2), len(out.Slice1[0].Slice2)) {
				assert.Equal(t, in.Slice1[0].Slice2[0].Field3, out.Slice1[0].Slice2[0].Field3)
				assert.Equal(t, in.Slice1[0].Slice2[1].Field3, out.Slice1[0].Slice2[1].Field3)
			}
		}
	}
}

func jsonMarshal(in interface{}) []byte {
	bytes, _ := json.Marshal(in)
	return bytes
}

func TestMapsToStructsSimple(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": Location{
				Country: "UK",
				City:    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call") {
		if assert.Equal(t, 4, len(users), "all rows should be returned") {
			assert.Equal(t, 19, users[0].Age, "values should be correctly set at start")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start")
			if assert.Equal(t, 2, len(users[0].Sports), "slices should be the right size") {
				assert.Equal(t, "football", users[0].Sports[0], "values should be correctly set at start")
			}
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end")
		}
	}
}

func TestMapsToStructsInnerMap(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": map[string]interface{}{
				"country": "UK",
				"city":    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call") {
		if assert.Equal(t, 4, len(users), "all rows should be returned") {
			assert.Equal(t, 19, users[0].Age, "values should be correctly set at start")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start")
			if assert.Equal(t, 2, len(users[0].Sports), "slices should be the right size") {
				assert.Equal(t, "football", users[0].Sports[0], "values should be correctly set at start")
			}
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end")
		}
	}
}

func TestMapsToStructsInnerMapWithPointers(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": map[string]interface{}{
				"country": "UK",
				"city":    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []UserWithPointers

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call") {
		if assert.Equal(t, 4, len(users), "all rows should be returned") {
			assert.Equal(t, 19, *users[0].Age, "values should be correctly set at start")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start")
			if assert.Equal(t, 2, len(*users[0].Sports), "slices should be the right size") {
				sports := users[0].Sports
				assert.Equal(t, "football", (*sports)[0], "values should be correctly set at start")
			}
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end")
		}
	}
}

func TestMapsToStructsMapWithPointers(t *testing.T) {
	male := "male"
	nineteen := float64(19)
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": &male, "age": &nineteen,
			"sports": []string{"football", "tennis"},
			"location": Location{
				Country: "UK",
				City:    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call") {
		if assert.Equal(t, 4, len(users), "all rows should be returned") {
			assert.Equal(t, 19, users[0].Age, "values should be correctly set at start")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start")
			if assert.Equal(t, 2, len(users[0].Sports), "slices should be the right size") {
				assert.Equal(t, "football", users[0].Sports[0], "values should be correctly set at start")
			}
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end")
		}
	}
}

func TestMapsToStructsWithConvert(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": float64(213), "name": "Zhaoliu", "gender": "male", "age": 19,
			"sports": []string{"football", "tennis"},
			"location": Location{
				Country: "UK",
				City:    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call") {
		if assert.Equal(t, 4, len(users), "all rows should be returned") {
			assert.Equal(t, 19, users[0].Age, "values should be correctly set at start")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start")
			if assert.Equal(t, 2, len(users[0].Sports), "slices should be the right size") {
				assert.Equal(t, "football", users[0].Sports[0], "values should be correctly set at start")
			}
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end")
		}
	}
}

func TestMapsToStructsWithPointers(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19,
			"sports": &[]string{"football", "tennis"},
			"location": &Location{
				Country: "UK",
				City:    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []UserWithPointers

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call with struct containing pointer") {
		if assert.Equal(t, 4, len(users), "all rows should be returned with struct containing pointer") {
			assert.Equal(t, 19, *users[0].Age, "values should be correctly set at start with struct containing pointer")
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end with struct containing pointer")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start with struct containing pointer")
			if assert.Equal(t, 2, len(*users[0].Sports), "slices should be the right size") {
				sports := users[0].Sports
				assert.Equal(t, "football", (*sports)[0], "values should be correctly set at start with struct containing pointer")
			}
		}
	}
}

func TestMapsToStructsWithPointersAndConvert(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": float64(19),
			"sports": &[]string{"football", "tennis"},
			"location": &Location{
				Country: "UK",
				City:    "London",
			}},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []UserWithPointers

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.Nil(t, err, "error should be nil for valid call with struct containing pointer") {
		if assert.Equal(t, 4, len(users), "all rows should be returned with struct containing pointer") {
			assert.Equal(t, 19, *users[0].Age, "values should be correctly set at start with struct containing pointer")
			assert.Equal(t, 978, users[3].ID, "values should be correctly set at end with struct containing pointer")
			assert.Equal(t, "UK", users[0].Location.Country, "values should be correctly set at start with struct containing pointer")
			if assert.Equal(t, 2, len(*users[0].Sports), "slices should be the right size") {
				sports := users[0].Sports
				assert.Equal(t, "football", (*sports)[0], "values should be correctly set at start with struct containing pointer")
			}
		}
	}
}

func TestMapsToStructsUsingTags(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "sex": "male", "age": 19},
		{"id": 56, "name": "Zhangsan", "sex": "male", "age": 37},
		{"id": 7, "name": "Lisi", "sex": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "sex": "male", "age": 28},
	}

	var users []UserWithTags

	err := mapstostructs.MapsToStructs(maps, &users, "alias")

	if assert.Nil(t, err, "error should be nil for valid call") {
		if assert.Equal(t, 4, len(users), "all rows should be returned") {
			assert.Equal(t, "male", users[0].Gender, "values should be correctly set from tags")
			assert.Equal(t, 28, users[3].Age, "values with any tags should be correctly set")
		}
	}
}

func TestMapsToStructsBadMap(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": "7", "name": "Lisi", "gender": "female", "age": 54},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []User

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.NotNil(t, err, "error should not be nil with invalid data") {
		expected := "the ID field for a struct of type User must be of type int but received a value of type string in row 3"
		assert.Equal(t, expected, err.Error(), "the error string should identify the bad data location")
	}
}

func TestMapsToStructsPointerBadMap(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": "54"},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	var users []UserWithPointers

	err := mapstostructs.MapsToStructs(maps, &users)

	if assert.NotNil(t, err, "error should not be nil with invalid data") {
		expected := "the Age field for a struct of type UserWithPointers must be of type int but received a value of type string in row 3"
		assert.Equal(t, expected, err.Error(), "the error string should identify the bad data location")
	}
}

func TestMapsToStructsPointerBadReceiver1(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": "54"},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}

	err := mapstostructs.MapsToStructs(maps, "test")

	if assert.NotNil(t, err, "error should not be nil with an invalid receiver") {
		expected := "the receivers argument must be a ptr to a slice of struct but a string was given"
		assert.Equal(t, expected, err.Error(), "the error string should identify the bad data location")
	}
}

func TestMapsToStructsPointerBadReceiver2(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": "54"},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}
	test := "test"

	err := mapstostructs.MapsToStructs(maps, &test)

	if assert.NotNil(t, err, "error should not be nil with an invalid receiver") {
		expected := "the receivers argument must be a ptr to a slice of struct but a ptr to a string was given"
		assert.Equal(t, expected, err.Error(), "the error string should identify the bad data location")
	}
}

func TestMapsToStructsPointerBadReceiver3(t *testing.T) {
	maps := []map[string]interface{}{
		{"id": 213, "name": "Zhaoliu", "gender": "male", "age": 19},
		{"id": 56, "name": "Zhangsan", "gender": "male", "age": 37},
		{"id": 7, "name": "Lisi", "gender": "female", "age": "54"},
		{"id": 978, "name": "Wangwu", "gender": "male", "age": 28},
	}
	test := []string{"test"}

	err := mapstostructs.MapsToStructs(maps, &test)

	if assert.NotNil(t, err, "error should not be nil with an invalid receiver") {
		expected := "the receivers argument must be a ptr to a slice of struct but a ptr to a slice of string was given"
		assert.Equal(t, expected, err.Error(), "the error string should identify the bad data location")
	}
}

func TestMapsToStructsSingle(t *testing.T) {
	map1 := map[string]interface{}{
		"id":     213,
		"name":   "Zhaoliu",
		"gender": "male",
		"age":    19,
		"sports": []string{"football", "tennis"},
		"location": Location{
			Country: "UK",
			City:    "London",
		}}

	var user User

	err := mapstostructs.MapToStruct(map1, &user)

	if assert.Nil(t, err, "error should be nil for valid call") {
		assert.Equal(t, 19, user.Age, "values should be correctly set")
		assert.Equal(t, "UK", user.Location.Country, "values should be correctly set")
		if assert.Equal(t, 2, len(user.Sports), "slices should be the right size") {
			assert.Equal(t, "football", user.Sports[0], "values should be correctly set")
		}
	}
}
