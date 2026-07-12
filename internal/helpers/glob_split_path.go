package helpers

import "strings"

func GlobSplitPath(p string) []string {
	p = strings.TrimPrefix(p, "/")
	if p == "" {
		return nil
	}

	return strings.Split(p, "/")
}
