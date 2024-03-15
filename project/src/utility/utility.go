package utility

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"project/types"
	"reflect"
)

func MarshalJson(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshaling JSON:", err)
		os.Exit(1)
	}
	return jsonData
}

func UnmarshalJson(data []byte, v interface{}) (reflect.Type, error) {
	err := json.Unmarshal(data, v)
	if err != nil {
		log.Println("Error unmarshaling JSON:", err)
		return nil, err
	}

	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		// If v is a pointer, get the type it points to
		return t.Elem(), nil
	}
	return t, nil
}

func SlicesAreEqual(a, b interface{}) bool {
	sliceA := reflect.ValueOf(a)
	sliceB := reflect.ValueOf(b)

	if sliceA.Kind() != reflect.Slice || sliceB.Kind() != reflect.Slice {
		return false
	}

	if sliceA.Len() != sliceB.Len() {
		return false
	}

	for i := 0; i < sliceA.Len(); i++ {
		elementA := sliceA.Index(i)
		elementB := sliceB.Index(i)

		if !reflect.DeepEqual(elementA.Interface(), elementB.Interface()) {
			return false
		}
	}

	return true
}

func StringSlicesAreEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func DetermineStructTypeAndUnmarshal(data []byte) (interface{}, error) {
	var tempMap map[string]interface{}
	if err := json.Unmarshal(data, &tempMap); err != nil {
		return nil, err
	}

	if _, ok := tempMap["elevators"]; ok {
		var ml types.MasterList
		if err := json.Unmarshal(data, &ml); err != nil {
			return nil, err
		}
		return ml, nil
	} else if _, ok := tempMap["ip"]; ok {
		var el types.Elevator
		if err := json.Unmarshal(data, &el); err != nil {
			return nil, err
		}
		return el, nil
	} else if _, ok := tempMap["direction"]; ok {
		var es types.ElevStatus
		if err := json.Unmarshal(data, &es); err != nil {
			return nil, err
		}
		return es, nil
	}

	return nil, fmt.Errorf("unable to determine struct type from JSON keys")
}

func IsIPInMasterList(ip string, masterList types.MasterList) bool {
	for _, elevator := range masterList.Elevators {
		if elevator.Ip == ip {
			return true // IP found in the list
		}
	}
	return false // IP not found in the list
}

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func CombineOrders(a [][]bool, b [][]bool) [][]bool {
	for i, a_row := range a {
		for j, value := range a_row {
			if b[i][j] || value {
				a[i][j] = true
			}
		}
	}

	return a

}
