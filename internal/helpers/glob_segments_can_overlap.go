package helpers

import (
	"path"
	"strings"
)

func GlobSegmentsCanOverlap(a, b string) bool {
	if a == "*" || b == "*" {
		return true
	}

	if strings.ContainsAny(a, "*?[") && strings.ContainsAny(b, "*?[") {
		return true
	}

	if strings.ContainsAny(a, "*?[") {
		matched, _ := path.Match(a, b)

		return matched
	}

	if strings.ContainsAny(b, "*?[") {
		matched, _ := path.Match(b, a)

		return matched
	}

	return a == b
}
