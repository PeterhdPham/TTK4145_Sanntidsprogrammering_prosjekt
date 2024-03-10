package utility

import (
	"encoding/json"
	"fmt"
	"os"
)

func MarshalJson(data interface{}) []byte {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		os.Exit(1)
	}
	return jsonData
}

func UnmarshalJson(data []byte, v interface{}) {
	err := json.Unmarshal(data, v)
	if err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		os.Exit(1)
	}
}

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

