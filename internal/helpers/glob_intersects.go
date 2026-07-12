package helpers

// GlobIntersects checks if two glob patterns can potentially match the same path.
func GlobIntersects(a, b string) bool {
	return GlobIntersectSegments(GlobSplitPath(a), GlobSplitPath(b))
}
