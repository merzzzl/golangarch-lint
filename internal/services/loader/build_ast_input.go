package loader

import (
	"github.com/merzzzl/golangarch-lint/internal/dto"
	"golang.org/x/tools/go/packages"
)

// BuildASTInput converts loaded packages into the AST check input.
func (s *Service) buildASTInput(root string, pkgs []*packages.Package) *dto.ASTInput {
	input := &dto.ASTInput{
		Root: root,
	}

	for _, pkg := range pkgs {
		pkgDir := s.packageDir(root, pkg)
		if pkgDir == "" {
			continue
		}

		ap := dto.ASTPackage{
			Dir:       pkgDir,
			TypesInfo: pkg.TypesInfo,
		}

		for _, file := range pkg.Syntax {
			fp := pkg.Fset.File(file.Pos()).Name()
			ap.GoFiles = append(ap.GoFiles, fp)
			ap.Files = append(ap.Files, file)
		}

		input.Packages = append(input.Packages, ap)
	}

	if len(pkgs) > 0 {
		input.Fset = pkgs[0].Fset
	}

	return input
}
