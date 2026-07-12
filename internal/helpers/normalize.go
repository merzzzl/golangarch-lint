package helpers

import "strings"

// Normalize removes underscores, strips .go, and lowercases for bindToFile comparison.
func Normalize(s string) string {
	s = strings.ReplaceAll(s, "_", "")
	s = strings.TrimSuffix(s, ".go")

	return strings.ToLower(s)
}
