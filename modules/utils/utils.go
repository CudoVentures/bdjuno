package utils

import (
	"encoding/json"
	"strings"
)

func SanitizeUTF8(s string) string {
	return strings.ToValidUTF8(s, "")
}

func IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func GetData(data string) (string, string) {
	dataText := data
	dataJSON := "{}"

	if data != "" && IsJSON(data) {
		dataJSON = data
		dataText = ""
	}

	return dataJSON, dataText
}
