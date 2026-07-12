package helpers

import "strings"

// TokenizeFilename splits a filename (without .go extension) by underscores.
func TokenizeFilename(name string) []string {
	name = strings.TrimSuffix(name, ".go")

	parts := strings.Split(name, "_")

	result := make([]string, 0, len(parts))

	for _, p := range parts {
		if p != "" {
			result = append(result, strings.ToLower(p))
		}
	}

	return result
}
