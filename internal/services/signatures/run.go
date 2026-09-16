package signatures

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/helpers"
)

func (s *Service) Run(input *dto.ASTInput, scope *dto.Scope, rep *dto.Report) error {
	for i := range input.Packages {
		pkg := &input.Packages[i]
		for j, file := range pkg.Files {
			rel, err := filepath.Rel(scope.Root, pkg.GoFiles[j])
			if err != nil {
				return fmt.Errorf("signatures path: %w", err)
			}

			rel = filepath.ToSlash(rel)
			for _, idx := range scope.Files[rel] {
				s.checkFile(input.Fset, pkg, file, rel, idx, rep)
			}
		}
	}

	return nil
}

func (s *Service) checkFile(fset *token.FileSet, pkg *dto.ASTPackage, file *ast.File, rel string, idx int, rep *dto.Report) {
	settings := s.rules[idx].Options

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		sig := settings.Unexported
		visibility := "unexported"

		if fn.Name.IsExported() {
			sig = settings.Exported
			visibility = "exported"
		}

		if sig.Receiver != nil && *sig.Receiver != (fn.Recv != nil) {
			s.reportViolation(fset, fn.Pos(), rel, idx, "receiver-required", fmt.Sprintf("func %q must have receiver=%t", fn.Name.Name, *sig.Receiver), rep)
		}

		s.checkFields(fset, pkg, fn.Type.Params, sig.Inputs, rel, idx, visibility+"-input-type", rep)
		s.checkFields(fset, pkg, fn.Type.Results, sig.Outputs, rel, idx, visibility+"-output-type", rep)
	}
}

func (s *Service) reportViolation(fset *token.FileSet, pos token.Pos, rel string, idx int, check, message string, rep *dto.Report) {
	p := fset.Position(pos)
	rep.Violations = append(rep.Violations, dto.Violation{
		Severity: "error", Check: check, Rule: s.rules[idx].Path, Path: rel,
		Pos: fmt.Sprintf("%s:%d:%d", rel, p.Line, p.Column), Message: message,
	})
}

func (s *Service) checkFields(fset *token.FileSet, pkg *dto.ASTPackage, fields *ast.FieldList, allow *[]string, rel string, idx int, check string, rep *dto.Report) {
	if fields == nil || allow == nil || pkg.TypesInfo == nil {
		return
	}

	for _, field := range fields.List {
		t := pkg.TypesInfo.TypeOf(field.Type)
		if t == nil {
			continue
		}

		base := helpers.UnwrapType(t)
		if helpers.IsBuiltinContainer(base) {
			continue
		}

		name := types.TypeString(base, nil)
		if helpers.IsBuiltinType(name) || s.isStdlibType(name) {
			continue
		}

		if !s.isAllowed(name, *allow) {
			s.reportViolation(fset, field.Pos(), rel, idx, check, fmt.Sprintf("type %q is not allowed by signatures", name), rep)
		}
	}
}

func (s *Service) isAllowed(typeStr string, allow []string) bool {
	for _, pattern := range allow {
		if s.matchType(pattern, typeStr) {
			return true
		}
	}

	return false
}

// matchType matches a type against a pattern:
//   - "pkg/path.Name" — exact type;
//   - "pkg/path" — any type from the package;
//   - "pkg/*" — any type from packages one level under pkg;
//   - "pkg/**" — any type from any package under pkg.
func (*Service) matchType(pattern, typeStr string) bool {
	if pattern == typeStr {
		return true
	}

	// Type arguments may contain package-qualified names; they are not part of
	// the declaring package path of the outer named type.
	if bracket := strings.IndexByte(typeStr, '['); bracket >= 0 {
		typeStr = typeStr[:bracket]
		if pattern == typeStr {
			return true
		}
	}

	dot := strings.LastIndex(typeStr, ".")
	if dot < 0 {
		return false
	}

	return helpers.GlobMatch(pattern, typeStr[:dot])
}

func (*Service) isStdlibType(typeStr string) bool {
	if !strings.Contains(typeStr, ".") {
		return true
	}

	return helpers.IsStdlibType(typeStr)
}
