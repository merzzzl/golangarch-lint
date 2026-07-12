package astcheck

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

func (s *Service) astCheckFile(
	fset *token.FileSet,
	pkg *dto.ASTPackage,
	file *ast.File,
	rel string,
	pkgDir string,
	rep *dto.Report,
) {
	ruleIdxs := s.astCheckSelectRules(rel, pkgDir)

	s.astCheckBindToFile(fset, file, rel, ruleIdxs, rep)

	for _, idx := range ruleIdxs {
		if s.astCheckIsRuleIgnored(rel, s.cfg.Rules[idx].Ignore) {
			continue
		}

		s.astCheckDecls(fset, file, rel, idx, rep)

		if s.cfg.Rules[idx].Exported != nil {
			s.astCheckExported(fset, pkg, file, rel, idx, rep)
		}
	}
}

// astCheckPackageName enforces that the package name is exactly equal to its
// directory name. Package main is exempt.
func (*Service) astCheckPackageName(
	fset *token.FileSet,
	pkg *dto.ASTPackage,
	file *ast.File,
	rel string,
	rep *dto.Report,
) {
	pkgName := file.Name.Name
	if pkgName == "main" || pkg.Dir == "." {
		return
	}

	expected := filepath.Base(pkg.Dir)
	if pkgName == expected {
		return
	}

	pos := fset.Position(file.Name.Pos())

	rep.Violations = append(rep.Violations, dto.Violation{
		Check:   "package-dir-binding",
		Rule:    "",
		Path:    rel,
		Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
		Message: fmt.Sprintf("package %q does not match directory %q (expected %q)", pkgName, pkg.Dir, expected),
	})
}

// astCheckSelectRules returns indices of rules applicable to a file: rules
// whose path matches the file's relative path take priority over rules
// matching the package directory.
func (s *Service) astCheckSelectRules(rel, pkgDir string) []int {
	var fileRules, dirRules []int

	for i := range s.cfg.Rules {
		r := &s.cfg.Rules[i]

		switch {
		case helpers.GlobMatch(r.Path, rel):
			fileRules = append(fileRules, i)
		case helpers.GlobMatch(r.Path, pkgDir):
			dirRules = append(dirRules, i)
		default:
		}
	}

	if len(fileRules) > 0 {
		return fileRules
	}

	return dirRules
}

// astCheckIsExcluded reports whether a declaration name matches any exclude pattern.
func (*Service) astCheckIsExcluded(name string, patterns []string) bool {
	for _, p := range patterns {
		if helpers.GlobMatch(p, name) {
			return true
		}
	}

	return false
}

// astCheckIsExcludedByRules reports whether a func name is excluded by any applicable rule.
func (s *Service) astCheckIsExcludedByRules(name string, ruleIdxs []int) bool {
	for _, idx := range ruleIdxs {
		if s.astCheckIsExcluded(name, s.cfg.Rules[idx].ExcludeFuncs) {
			return true
		}
	}

	return false
}

func (s *Service) astCheckDecls(
	fset *token.FileSet,
	file *ast.File,
	rel string,
	ruleIdx int,
	rep *dto.Report,
) {
	r := &s.cfg.Rules[ruleIdx]

	if r.AllowFuncs == "" && r.AllowVars == "" && r.AllowTypes == "" && r.RequireReceiver == "" {
		return
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if s.astCheckIsExcluded(d.Name.Name, r.ExcludeFuncs) {
				continue
			}

			if s.astCheckDeclDenied(r.AllowFuncs, d.Name.IsExported()) {
				pos := fset.Position(d.Pos())

				rep.Violations = append(rep.Violations, dto.Violation{
					Check:   "func-decl-denied",
					Rule:    r.Path,
					Path:    rel,
					Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
					Message: fmt.Sprintf("func %q is not allowed: allow-funcs is %q", d.Name.Name, r.AllowFuncs),
				})
			}

			if d.Recv == nil && s.astCheckReceiverRequired(r.RequireReceiver, d.Name.IsExported()) {
				pos := fset.Position(d.Pos())

				rep.Violations = append(rep.Violations, dto.Violation{
					Check:   "receiver-required",
					Rule:    r.Path,
					Path:    rel,
					Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
					Message: fmt.Sprintf("func %q must be a method (require-receiver is %q)", d.Name.Name, r.RequireReceiver),
				})
			}
		case *ast.GenDecl:
			s.astCheckGenDecl(fset, d, rel, ruleIdx, rep)
		default:
		}
	}
}

// astCheckReceiverRequired reports whether a plain function must be a method:
// "all" — every function; "exported" — exported only; "local" — unexported
// only; "none" or empty — not required.
func (*Service) astCheckReceiverRequired(mode string, exported bool) bool {
	switch mode {
	case "all":
		return true
	case "exported":
		return exported
	case "local":
		return !exported
	default:
		return false
	}
}

// astCheckDeclDenied reports whether a declaration is denied by the allow mode:
// "all" or empty — everything allowed; "local" — only unexported allowed;
// "exported" — only exported allowed; "none" — nothing allowed.
func (*Service) astCheckDeclDenied(mode string, exported bool) bool {
	switch mode {
	case "none":
		return true
	case "local":
		return exported
	case "exported":
		return !exported
	default:
		return false
	}
}

func (s *Service) astCheckGenDecl(
	fset *token.FileSet,
	d *ast.GenDecl,
	rel string,
	ruleIdx int,
	rep *dto.Report,
) {
	r := &s.cfg.Rules[ruleIdx]

	if d.Tok == token.VAR || d.Tok == token.CONST {
		for _, spec := range d.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for _, name := range vs.Names {
				if s.astCheckIsExcluded(name.Name, r.ExcludeVars) {
					continue
				}

				if !s.astCheckDeclDenied(r.AllowVars, name.IsExported()) {
					continue
				}

				pos := fset.Position(name.Pos())

				rep.Violations = append(rep.Violations, dto.Violation{
					Check:   "var-decl-denied",
					Rule:    r.Path,
					Path:    rel,
					Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
					Message: fmt.Sprintf("%s %q is not allowed: allow-vars is %q", d.Tok, name.Name, r.AllowVars),
				})
			}
		}
	}

	if d.Tok == token.TYPE {
		for _, spec := range d.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || s.astCheckIsExcluded(ts.Name.Name, r.ExcludeTypes) {
				continue
			}

			if !s.astCheckDeclDenied(r.AllowTypes, ts.Name.IsExported()) {
				continue
			}

			pos := fset.Position(ts.Pos())

			rep.Violations = append(rep.Violations, dto.Violation{
				Check:   "type-decl-denied",
				Rule:    r.Path,
				Path:    rel,
				Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
				Message: fmt.Sprintf("type %q is not allowed: allow-types is %q", ts.Name.Name, r.AllowTypes),
			})
		}
	}
}

func (s *Service) astCheckBindToFile(
	fset *token.FileSet,
	file *ast.File,
	rel string,
	ruleIdxs []int,
	rep *dto.Report,
) {
	fileName := filepath.Base(rel)
	normFile := helpers.Normalize(fileName)

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if s.astCheckIsExcludedByRules(fn.Name.Name, ruleIdxs) {
			continue
		}

		normFunc := helpers.Normalize(fn.Name.Name)

		if fn.Name.IsExported() {
			if normFunc != normFile {
				pos := fset.Position(fn.Pos())
				suggested := s.astCheckSuggestFileName(fn.Name.Name)

				rep.Violations = append(rep.Violations, dto.Violation{
					Check: "method-file-binding",
					Rule:  "",
					Path:  rel,
					Pos:   fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
					Message: fmt.Sprintf(
						"exported func %q must live in file named after it (e.g. %q), found in %q",
						fn.Name.Name, suggested, fileName,
					),
				})
			}

			continue
		}

		if normFunc != normFile && !strings.HasPrefix(normFunc, normFile) {
			pos := fset.Position(fn.Pos())

			rep.Violations = append(rep.Violations, dto.Violation{
				Check: "method-file-binding",
				Rule:  "",
				Path:  rel,
				Pos:   fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
				Message: fmt.Sprintf(
					"private func %q does not belong in %q, normalized name must start with %q",
					fn.Name.Name, fileName, normFile,
				),
			})
		}
	}
}

func (s *Service) astCheckExported(
	fset *token.FileSet,
	pkg *dto.ASTPackage,
	file *ast.File,
	rel string,
	ruleIdx int,
	rep *dto.Report,
) {
	r := &s.cfg.Rules[ruleIdx]

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !fn.Name.IsExported() {
			continue
		}

		if s.astCheckIsExcluded(fn.Name.Name, r.ExcludeFuncs) {
			continue
		}

		if r.Exported.Inputs != nil {
			s.astCheckFieldList(fset, pkg, fn, rel, fn.Type.Params, "inputs", ruleIdx, rep)
		}

		if r.Exported.Outputs != nil {
			s.astCheckFieldList(fset, pkg, fn, rel, fn.Type.Results, "outputs", ruleIdx, rep)
		}
	}
}

func (s *Service) astCheckFieldList(
	fset *token.FileSet,
	pkg *dto.ASTPackage,
	fn *ast.FuncDecl,
	rel string,
	fields *ast.FieldList,
	section string,
	ruleIdx int,
	rep *dto.Report,
) {
	if fields == nil {
		return
	}

	r := &s.cfg.Rules[ruleIdx]

	allow := r.Exported.Inputs
	checkName := "exported-input-type"

	if section == "outputs" {
		allow = r.Exported.Outputs
		checkName = "exported-output-type"
	}

	for _, field := range fields.List {
		typeObj := pkg.TypesInfo.TypeOf(field.Type)
		if typeObj == nil {
			continue
		}

		baseType := s.astCheckUnwrapType(typeObj)

		// Maps, channels, funcs — only outer type matters, which is builtin
		if s.astCheckIsBuiltinContainer(baseType) {
			continue
		}

		typeStr := types.TypeString(baseType, nil)

		if helpers.IsBuiltinType(typeStr) || s.astCheckIsStdlibType(typeStr) {
			continue
		}

		if !s.astCheckIsAllowed(typeStr, allow) {
			pos := fset.Position(field.Pos())

			rep.Violations = append(rep.Violations, dto.Violation{
				Check: checkName,
				Rule:  r.Path,
				Path:  rel,
				Pos:   fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
				Message: fmt.Sprintf(
					"type %q in %s of func %q is not in allow list",
					typeStr, section, fn.Name.Name,
				),
			})
		}
	}
}

func (*Service) astCheckIsBuiltinContainer(t types.Type) bool {
	switch t.(type) {
	case *types.Map, *types.Chan, *types.Signature:
		return true
	default:
		return false
	}
}

func (*Service) astCheckSuggestFileName(name string) string {
	tokens := helpers.TokenizeCamel(name)

	return strings.Join(tokens, "_") + ".go"
}

func (*Service) astCheckUnwrapType(t types.Type) types.Type {
	for {
		switch v := t.(type) {
		case *types.Pointer:
			t = v.Elem()
		case *types.Slice:
			t = v.Elem()
		case *types.Array:
			t = v.Elem()
		default:
			return t
		}
	}
}

func (s *Service) astCheckIsAllowed(typeStr string, allow []string) bool {
	for _, pattern := range allow {
		if s.astCheckMatchType(pattern, typeStr) {
			return true
		}
	}

	return false
}

// astCheckMatchType matches a type against a pattern:
//   - "pkg/path.Name" — exact type;
//   - "pkg/path" — any type from the package;
//   - "pkg/*" — any type from packages one level under pkg;
//   - "pkg/**" — any type from any package under pkg.
func (*Service) astCheckMatchType(pattern, typeStr string) bool {
	if pattern == typeStr {
		return true
	}

	dot := strings.LastIndex(typeStr, ".")
	if dot < 0 {
		return false
	}

	return helpers.GlobMatch(pattern, typeStr[:dot])
}

func (*Service) astCheckIsStdlibType(typeStr string) bool {
	if !strings.Contains(typeStr, ".") {
		return true
	}

	return helpers.IsStdlibType(typeStr)
}

func (*Service) astCheckIsIgnored(rel string, patterns []string) bool {
	for _, p := range patterns {
		if helpers.GlobMatch(p, rel) {
			return true
		}
	}

	return false
}

func (*Service) astCheckIsRuleIgnored(rel string, ignore []string) bool {
	for _, p := range ignore {
		if helpers.GlobMatch(p, rel) {
			return true
		}
	}

	return false
}
