package loader

import (
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

// PackageDir returns the package directory relative to root.
func (*Service) packageDir(root string, pkg *packages.Package) string {
	var first string

	switch {
	case len(pkg.GoFiles) > 0:
		first = pkg.GoFiles[0]
	case len(pkg.CompiledGoFiles) > 0:
		first = pkg.CompiledGoFiles[0]
	default:
		return ""
	}

	dir := filepath.Dir(first)

	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return ""
	}

	return filepath.ToSlash(rel)
}
