package helpers

func GlobMatch(pattern, name string) bool {
	return GlobMatchSegments(GlobSplitPath(pattern), GlobSplitPath(name))
}
