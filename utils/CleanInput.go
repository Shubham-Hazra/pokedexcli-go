package utils

import "strings"

func CleanInput(text string) []string {
	parts := strings.Fields(text)
	for i, s := range parts {
		parts[i] = strings.ToLower(strings.TrimSpace(s))
	}
	return parts
}
