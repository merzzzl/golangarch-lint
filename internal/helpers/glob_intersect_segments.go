package helpers

func GlobIntersectSegments(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}

	if len(a) == 0 || len(b) == 0 {
		rest := a
		if len(a) == 0 {
			rest = b
		}

		for _, seg := range rest {
			if seg != "**" {
				return false
			}
		}

		return true
	}

	if a[0] == "**" && b[0] == "**" {
		return true
	}

	if a[0] == "**" {
		return GlobIntersectSegments(a[1:], b) || GlobIntersectSegments(a, b[1:])
	}

	if b[0] == "**" {
		return GlobIntersectSegments(a, b[1:]) || GlobIntersectSegments(a[1:], b)
	}

	if !GlobSegmentsCanOverlap(a[0], b[0]) {
		return false
	}

	return GlobIntersectSegments(a[1:], b[1:])
}
