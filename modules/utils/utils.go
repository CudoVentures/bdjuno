package utils

import "strings"

func SanitizeUTF8(s string) string {
	return strings.ToValidUTF8(s, "")
}
