package utils


import "strings"

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

