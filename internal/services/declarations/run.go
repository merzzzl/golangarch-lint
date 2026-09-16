package declarations

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"

	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
)

func (s *Service) Run(input *dto.ASTInput, scope *dto.Scope, rep *dto.Report) error {
	for i := range input.Packages {
		pkg := &input.Packages[i]
		for j, file := range pkg.Files {
			rel, err := filepath.Rel(scope.Root, pkg.GoFiles[j])
			if err != nil {
				return fmt.Errorf("declarations path: %w", err)
			}

			rel = filepath.ToSlash(rel)
			for _, idx := range scope.Files[rel] {
				s.checkFile(input.Fset, file, rel, idx, rep)
			}
		}
	}

	return nil
}

func (s *Service) checkFile(fset *token.FileSet, file *ast.File, rel string, idx int, rep *dto.Report) {
	r := s.rules[idx].Options

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			s.checkDeclaration(fset, d.Name, r.Funcs, "func", rel, idx, rep)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch v := spec.(type) {
				case *ast.TypeSpec:
					s.checkDeclaration(fset, v.Name, r.Types, "type", rel, idx, rep)
				case *ast.ValueSpec:
					filter := r.Vars
					kind := "var"

					if d.Tok == token.CONST {
						filter = r.Consts
						kind = "const"
					}

					for _, name := range v.Names {
						s.checkDeclaration(fset, name, filter, kind, rel, idx, rep)
					}
				default:
				}
			}
		default:
		}
	}
}

func (s *Service) checkDeclaration(fset *token.FileSet, name *ast.Ident, visibility config.Visibility, kind, rel string, idx int, rep *dto.Report) {
	filter := visibility.Unexported
	if name.IsExported() {
		filter = visibility.Exported
	}

	if !filter.Matches(name.Name) {
		s.reportViolation(fset, name.Pos(), rel, idx, kind+"-decl-denied", fmt.Sprintf("%s %q is not allowed by declarations", kind, name.Name), rep)
	}
}

func (s *Service) reportViolation(fset *token.FileSet, pos token.Pos, rel string, idx int, check, message string, rep *dto.Report) {
	p := fset.Position(pos)
	rep.Violations = append(rep.Violations, dto.Violation{
		Severity: "error", Check: check, Rule: s.rules[idx].Path, Path: rel,
		Pos: fmt.Sprintf("%s:%d:%d", rel, p.Line, p.Column), Message: message,
	})
}
