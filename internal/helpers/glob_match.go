package helpers

// GlobMatch checks if a slash-separated path matches a glob pattern with ** support.
// ** matches zero or more path segments.
func GlobMatch(pattern, name string) bool {
	return GlobMatchSegments(GlobSplitPath(pattern), GlobSplitPath(name))
}
