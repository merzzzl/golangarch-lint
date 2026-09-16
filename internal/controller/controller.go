package controller

import (
	"fmt"
	"path/filepath"

	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/services/declarations"
	"github.com/merzzzl/golangarch-lint/internal/services/imports"
	"github.com/merzzzl/golangarch-lint/internal/services/layout"
	"github.com/merzzzl/golangarch-lint/internal/services/scope"
	"github.com/merzzzl/golangarch-lint/internal/services/signatures"
	"github.com/merzzzl/golangarch-lint/internal/services/templater"
	"golang.org/x/tools/go/packages"
)

type Controller struct {
	analyzeCalls bool
	scope        *scope.Service
	layout       *layout.Service
	declarations *declarations.Service
	imports      *imports.Service
	signatures   *signatures.Service
	templater    *templater.Service
}

func (s *Controller) Lint(root string, rep *dto.Report) error {
	selected, err := s.scope.Run(root, rep)
	if err != nil {
		return fmt.Errorf("scope: %w", err)
	}

	// Load one shared project snapshot for all checks.
	mode := packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports
	if s.analyzeCalls {
		mode |= packages.NeedDeps
	}

	pkgs, err := packages.Load(&packages.Config{Dir: selected.Root, Mode: mode, Tests: false}, "./...")
	if err != nil {
		return fmt.Errorf("loading project: loading packages: %w", err)
	}

	if len(pkgs) == 0 {
		return fmt.Errorf("loading project: %w", ErrNoPackages)
	}

	var packageErr error

	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if len(pkg.Errors) > 0 && packageErr == nil {
			packageErr = fmt.Errorf("package error: %w", pkg.Errors[0])
		}
	})

	if packageErr != nil {
		return fmt.Errorf("loading project: %w", packageErr)
	}

	input := &dto.ASTInput{Root: selected.Root, LoadedPackages: pkgs}
	for _, pkg := range pkgs {
		if len(pkg.CompiledGoFiles) == 0 {
			continue
		}

		dir, err := filepath.Rel(selected.Root, filepath.Dir(pkg.CompiledGoFiles[0]))
		if err != nil {
			return fmt.Errorf("loading project: package directory: %w", err)
		}

		ap := dto.ASTPackage{Dir: filepath.ToSlash(dir), TypesInfo: pkg.TypesInfo}
		for _, file := range pkg.Syntax {
			ap.GoFiles = append(ap.GoFiles, pkg.Fset.Position(file.Pos()).Filename)
			ap.Files = append(ap.Files, file)
		}

		input.Packages = append(input.Packages, ap)
	}

	input.Fset = pkgs[0].Fset

	if err := s.layout.Run(input, selected, rep); err != nil {
		return fmt.Errorf("checking layout: %w", err)
	}

	if err := s.declarations.Run(input, selected, rep); err != nil {
		return fmt.Errorf("checking declarations: %w", err)
	}

	if err := s.imports.Run(input, selected, rep); err != nil {
		return fmt.Errorf("checking imports: %w", err)
	}

	if err := s.signatures.Run(input, selected, rep); err != nil {
		return fmt.Errorf("checking signatures: %w", err)
	}

	return nil
}

// Docs renders the AI-oriented architecture instruction (GOLANGARCH.md content).
func (s *Controller) Docs() string {
	return s.templater.Run()
}
