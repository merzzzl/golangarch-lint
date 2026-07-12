package helpers

import "strings"

func IsStdlibType(t string) bool {
	dot := strings.LastIndex(t, ".")
	if dot < 0 {
		return false
	}

	pkgPath := t[:dot]

	for seg := range strings.SplitSeq(pkgPath, "/") {
		if strings.Contains(seg, ".") {
			return false
		}
	}

	return true
}
