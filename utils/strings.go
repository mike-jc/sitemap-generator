package utils

import "encoding/json"

func InJSON(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

func StringSliceUnique(input []string) []string {
	keys := make(map[string]bool)
	result := make([]string, 0)

	for _, v := range input {
		if _, exists := keys[v]; !exists {
			keys[v] = true
			result = append(result, v)
		}
	}
	return result
}
