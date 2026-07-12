package helpers

import (
	"strings"
	"unicode"
)

// TokenizeCamel splits a CamelCase identifier into lowercase tokens.
// Consecutive uppercase letters form an abbreviation token.
// Example: "HTTPServer" -> ["http", "server"].
func TokenizeCamel(s string) []string {
	var tokens []string

	runes := []rune(s)
	start := 0

	for i := 1; i < len(runes); i++ {
		cur := runes[i]
		prev := runes[i-1]

		split := false

		switch {
		case unicode.IsUpper(cur) && unicode.IsLower(prev):
			split = true
		case unicode.IsUpper(cur) && unicode.IsUpper(prev) && i+1 < len(runes) && unicode.IsLower(runes[i+1]):
			split = true
		default:
		}

		if split {
			tokens = append(tokens, strings.ToLower(string(runes[start:i])))
			start = i
		}
	}

	if start < len(runes) {
		tokens = append(tokens, strings.ToLower(string(runes[start:])))
	}

	return tokens
}
