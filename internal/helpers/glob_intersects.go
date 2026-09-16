package helpers

func GlobIntersects(a, b string) bool {
	return GlobIntersectSegments(GlobSplitPath(a), GlobSplitPath(b))
}
