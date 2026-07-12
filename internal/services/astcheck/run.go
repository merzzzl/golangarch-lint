package astcheck

import (
	"path/filepath"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
)

func (s *Service) Run(input *dto.ASTInput, rep *dto.Report) {
	for i := range input.Packages {
		pkg := &input.Packages[i]

		pkgChecked := false

		for j, file := range pkg.Files {
			filePath := pkg.GoFiles[j]

			rel, err := filepath.Rel(input.Root, filePath)
			if err != nil {
				continue
			}

			rel = filepath.ToSlash(rel)

			if strings.HasSuffix(rel, "_test.go") || s.astCheckIsIgnored(rel, s.cfg.Ignore) {
				continue
			}

			if !pkgChecked {
				s.astCheckPackageName(input.Fset, pkg, file, rel, rep)

				pkgChecked = true
			}

			s.astCheckFile(input.Fset, pkg, file, rel, pkg.Dir, rep)
		}
	}
}
