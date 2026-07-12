package loader

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
	"golang.org/x/tools/go/packages"
)

// CollectImports gathers direct imports per package directory.
func (s *Service) collectImports(root string, pkgs []*packages.Package) map[string][]dto.ImportEntry {
	result := make(map[string][]dto.ImportEntry)

	for _, pkg := range pkgs {
		pkgDir := s.packageDir(root, pkg)
		if pkgDir == "" {
			continue
		}

		for _, file := range pkg.Syntax {
			filePath := pkg.Fset.File(file.Pos()).Name()

			rel, err := filepath.Rel(root, filePath)
			if err != nil {
				continue
			}

			rel = filepath.ToSlash(rel)

			if strings.HasSuffix(rel, "_test.go") {
				continue
			}

			for _, imp := range file.Imports {
				importPath := strings.Trim(imp.Path.Value, `"`)
				pos := pkg.Fset.Position(imp.Pos())

				result[pkgDir] = append(result[pkgDir], dto.ImportEntry{
					Path: importPath,
					Pos:  fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
					File: rel,
				})
			}
		}
	}

	return result
}
