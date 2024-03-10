package utility

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

func MarshalJson(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		os.Exit(1)
	}
	return jsonData
}

func UnmarshalJson(data []byte, v interface{}) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		fmt.Println("Data:", data)
		fmt.Println("String:", string(data))
		return err
	}
	return nil
}

func SlicesAreEqual(a, b interface{}) bool {
	sliceA := reflect.ValueOf(a)
	sliceB := reflect.ValueOf(b)

	if sliceA.Kind() != reflect.Slice || sliceB.Kind() != reflect.Slice {
		fmt.Println("SlicesAreEqual: Invalid input, both arguments must be slices")
		return false
	}

	if sliceA.Len() != sliceB.Len() {
		// for debug purposes
		if sliceA.Len() > 5 || sliceB.Len() > 5 {
			fmt.Println("SlicesAreEqual: Slices have different lengths")
		}
		return false
	}

	for i := 0; i < sliceA.Len(); i++ {
		elementA := sliceA.Index(i)
		elementB := sliceB.Index(i)

		if !reflect.DeepEqual(elementA.Interface(), elementB.Interface()) {
			fmt.Printf("SlicesAreEqual: Elements at index %d are not equal\n", i)
			return false
		}
	}

	return true
}
