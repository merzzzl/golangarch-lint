package helpers

import "strings"

func IsStdlibPackage(p string) bool {
	firstSlash := strings.Index(p, "/")

	first := p
	if firstSlash > 0 {
		first = p[:firstSlash]
	}

	return !strings.Contains(first, ".")
}
