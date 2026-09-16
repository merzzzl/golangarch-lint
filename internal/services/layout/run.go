package layout

import (
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/helpers"
)

func (s *Service) Run(input *dto.ASTInput, scope *dto.Scope, rep *dto.Report) error {
	s.checkDirectories(scope, rep)
	s.checkFilenames(scope, rep)

	for i := range input.Packages {
		pkg := &input.Packages[i]
		checked := false

		for j, file := range pkg.Files {
			rel, err := filepath.Rel(scope.Root, pkg.GoFiles[j])
			if err != nil {
				return fmt.Errorf("layout path: %w", err)
			}

			rel = filepath.ToSlash(rel)

			indices := scope.Files[rel]
			if len(indices) == 0 {
				continue
			}

			if !checked {
				s.checkPackageName(input.Fset, pkg, file, rel, rep)

				checked = true
			}

			for _, idx := range indices {
				s.checkBinding(input.Fset, pkg, file, rel, idx, rep)
			}
		}

		s.checkScattered(input, pkg, scope, rep)
	}

	return nil
}

// checkFilenames checks the scoped filesystem, including empty directories and
// files excluded by build tags. Directory constraints survive file overrides.
func (s *Service) checkFilenames(scope *dto.Scope, rep *dto.Report) {
	type group struct {
		dir string
		idx int
	}

	groups := map[group][]string{}

	for dir, indices := range scope.Directories {
		for _, idx := range indices {
			groups[group{dir, idx}] = nil
		}
	}

	for file, indices := range scope.FileDirectories {
		if len(scope.Files[file]) == 0 {
			continue
		}

		all := slices.Clone(indices)
		for _, idx := range scope.Files[file] {
			if !slices.Contains(all, idx) {
				all = append(all, idx)
			}
		}

		for _, idx := range all {
			key := group{path.Dir(file), idx}
			groups[key] = append(groups[key], file)
		}
	}

	keys := make([]group, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}

	slices.SortFunc(keys, func(a, b group) int {
		if a.dir < b.dir {
			return -1
		}

		if a.dir > b.dir {
			return 1
		}

		return a.idx - b.idx
	})

	for _, key := range keys {
		rule := s.rules[key.idx]
		if rule.Options.Filenames == nil {
			continue
		}

		patterns := *rule.Options.Filenames
		matched := make([]bool, len(patterns))
		files := groups[key]
		slices.Sort(files)

		for _, file := range files {
			allowed := false

			for i, pattern := range patterns {
				if ok, _ := path.Match(pattern, path.Base(file)); ok {
					allowed, matched[i] = true, true
				}
			}

			if !allowed {
				rep.Violations = append(rep.Violations, dto.Violation{
					Severity: "error", Check: "layout-filename", Rule: rule.Path, Path: file, Pos: file,
					Message: fmt.Sprintf("filename %q does not match any required filename pattern", path.Base(file)),
				})
			}
		}

		for i, pattern := range patterns {
			if !matched[i] {
				rep.Violations = append(rep.Violations, dto.Violation{
					Severity: "error", Check: "layout-filename-missing", Rule: rule.Path, Path: key.dir, Pos: key.dir,
					Message: fmt.Sprintf("no Go source file matches required filename pattern %q", pattern),
				})
			}
		}
	}
}

func (s *Service) reportViolation(fset *token.FileSet, pos token.Pos, rel string, idx int, check, message string, rep *dto.Report) {
	p := fset.Position(pos)
	rep.Violations = append(rep.Violations, dto.Violation{
		Severity: "error", Check: check, Rule: s.rules[idx].Path, Path: rel,
		Pos: fmt.Sprintf("%s:%d:%d", rel, p.Line, p.Column), Message: message,
	})
}

func (*Service) checkPackageName(fset *token.FileSet, pkg *dto.ASTPackage, file *ast.File, rel string, rep *dto.Report) {
	if file.Name.Name == "main" || pkg.Dir == "." || file.Name.Name == filepath.Base(pkg.Dir) {
		return
	}

	pos := fset.Position(file.Name.Pos())
	rep.Violations = append(rep.Violations, dto.Violation{
		Severity: "error", Check: "package-dir-binding", Path: rel,
		Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
		Message: fmt.Sprintf("package %q does not match directory %q", file.Name.Name, pkg.Dir),
	})
}

func (*Service) suggestFileName(name string) string {
	tokens := helpers.TokenizeCamel(name)

	return strings.Join(tokens, "_") + ".go"
}

// Every candidate is evaluated against the entire file. The candidate with the
// fewest errors/warnings is used for diagnostics; declarations never mix modes.
func (s *Service) checkBinding(fset *token.FileSet, pkg *dto.ASTPackage, file *ast.File, rel string, idx int, rep *dto.Report) {
	modes := s.rules[idx].Options.Binding.Modes()

	if len(modes) == 0 {
		return
	}

	var best *dto.Report

	for _, mode := range modes {
		candidate := &dto.Report{}

		switch mode {
		case "single":
			s.checkSingleBinding(fset, file, rel, idx, candidate)
		case "name":
			s.checkNameBinding(fset, file, rel, idx, candidate)
		case "receiver":
			s.checkReceiverBinding(fset, pkg, file, rel, idx, candidate)
		default:
		}

		if best == nil || len(candidate.Violations) < len(best.Violations) {
			best = candidate
		}
	}

	rep.Violations = append(rep.Violations, best.Violations...)
}

func (s *Service) checkNameBinding(fset *token.FileSet, file *ast.File, rel string, idx int, rep *dto.Report) {
	norm := helpers.Normalize(filepath.Base(rel))

	for _, ident := range s.declarationNames(file) {
		name := helpers.Normalize(ident.Name)
		if name == norm || (!ident.IsExported() && strings.HasPrefix(name, norm)) {
			continue
		}

		s.reportViolation(fset, ident.Pos(), rel, idx, "declaration-file-binding", fmt.Sprintf("declaration %q does not belong in %q in name binding", ident.Name, filepath.Base(rel)), rep)
	}
}

func (s *Service) checkReceiverBinding(fset *token.FileSet, pkg *dto.ASTPackage, file *ast.File, rel string, idx int, rep *dto.Report) {
	norm := helpers.Normalize(filepath.Base(rel))
	owner := s.typeOwner(pkg, file)
	// An actual type declaration fixes ownership before methods are inspected.
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}

		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if owner == "" {
				owner = ts.Name.Name
			}

			if helpers.Normalize(ts.Name.Name) != norm || ts.Name.Name != owner {
				s.reportViolation(fset, ts.Pos(), rel, idx, "type-file-binding", fmt.Sprintf("type %q must live in %q", ts.Name.Name, s.suggestFileName(ts.Name.Name)), rep)
			}
		}
	}

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if fn.Recv == nil {
			// Plain functions keep name binding in type layout.
			single := &ast.File{Decls: []ast.Decl{fn}}
			s.checkNameBinding(fset, single, rel, idx, rep)

			continue
		}

		receiver := helpers.ReceiverName(fn.Recv.List[0].Type)
		if owner == "" {
			owner = receiver
		}

		if receiver != owner || !s.hasType(pkg, receiver) {
			s.reportViolation(fset, fn.Pos(), rel, idx, "method-file-binding", fmt.Sprintf("method %q must belong to the file's type %q, found receiver %q", fn.Name.Name, owner, receiver), rep)
		}
	}

	for _, name := range s.valueNames(file) {
		if owner != "" && s.hasType(pkg, owner) && strings.HasPrefix(helpers.Normalize(name.Name), helpers.Normalize(owner)) {
			continue
		}

		s.reportViolation(fset, name.Pos(), rel, idx, "declaration-file-binding", fmt.Sprintf("variable or constant %q must start with the owning type name %q", name.Name, owner), rep)
	}
}

func (*Service) hasType(pkg *dto.ASTPackage, name string) bool {
	for j, file := range pkg.Files {
		if helpers.Normalize(filepath.Base(pkg.GoFiles[j])) != helpers.Normalize(name) {
			continue
		}

		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok && ts.Name.Name == name {
					return true
				}
			}
		}
	}

	return false
}

func (s *Service) checkScattered(input *dto.ASTInput, pkg *dto.ASTPackage, scope *dto.Scope, rep *dto.Report) {
	// Group only files within the same effective rule and build constraints.
	for idx := range s.rules {
		r := &s.rules[idx]

		modes := r.Options.Binding.Modes()

		typeGroups := map[string][]string{}
		nameFiles := map[string]*ast.File{}

		for j, file := range pkg.Files {
			rel, err := filepath.Rel(input.Root, pkg.GoFiles[j])
			if err != nil {
				continue
			}

			rel = filepath.ToSlash(rel)
			if strings.HasSuffix(rel, "_test.go") || !slices.Contains(scope.Files[rel], idx) {
				continue
			}

			if slices.Contains(modes, "receiver") {
				test := &dto.Report{}
				s.checkReceiverBinding(input.Fset, pkg, file, rel, idx, test)

				if len(test.Violations) == 0 {
					if len(s.valueNames(file)) > 0 {
						owner := s.typeOwner(pkg, file)
						key := owner + "|" + s.buildConstraintKey(file, rel)
						typeGroups[key] = append(typeGroups[key], rel)
					}

					for _, decl := range file.Decls {
						if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil {
							receiver := helpers.ReceiverName(fn.Recv.List[0].Type)
							key := receiver + "|" + s.buildConstraintKey(file, rel)
							typeGroups[key] = append(typeGroups[key], rel)
						}

						if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
							for _, spec := range gen.Specs {
								if ts, ok := spec.(*ast.TypeSpec); ok {
									key := ts.Name.Name + "|" + s.buildConstraintKey(file, rel)
									typeGroups[key] = append(typeGroups[key], rel)
								}
							}
						}
					}
				}
			}

			if slices.Contains(modes, "name") {
				test := &dto.Report{}
				s.checkNameBinding(input.Fset, file, rel, idx, test)

				if len(test.Violations) == 0 {
					nameFiles[rel] = file
				}
			}
		}

		keys := make([]string, 0, len(typeGroups))
		for key := range typeGroups {
			keys = append(keys, key)
		}

		slices.Sort(keys)

		for _, key := range keys {
			files := typeGroups[key]
			slices.Sort(files)

			files = slices.Compact(files)
			if len(files) > 1 {
				s.reportScattered(files, idx, "type "+strings.Split(key, "|")[0], rep)
			}
		}

		names := make([]string, 0, len(nameFiles))
		for rel := range nameFiles {
			names = append(names, rel)
		}

		slices.Sort(names)

		for i, rel := range names {
			// Compare each pair once. At least one function must move, and every
			// declaration in either file must be legal in the proposed target file.
			for _, other := range names[i+1:] {
				if s.buildConstraintKey(nameFiles[rel], rel) != s.buildConstraintKey(nameFiles[other], other) {
					continue
				}

				combined := &ast.File{Decls: append(slices.Clone(nameFiles[rel].Decls), nameFiles[other].Decls...)}
				if len(s.declarationNames(nameFiles[rel])) == 0 || len(s.declarationNames(nameFiles[other])) == 0 {
					continue
				}

				a, b := &dto.Report{}, &dto.Report{}
				s.checkNameBinding(input.Fset, combined, rel, idx, a)
				s.checkNameBinding(input.Fset, combined, other, idx, b)

				if len(a.Violations) == 0 || len(b.Violations) == 0 {
					s.reportScattered([]string{rel, other}, idx, "name layout", rep)
				}
			}
		}
	}
}

func (*Service) buildConstraintKey(file *ast.File, rel string) string {
	var tags []string

	for _, group := range file.Comments {
		for _, comment := range group.List {
			if strings.HasPrefix(comment.Text, "//go:build ") || strings.HasPrefix(comment.Text, "// +build ") {
				tags = append(tags, comment.Text)
			}
		}
	}
	// Preserve implicit filename build constraints when suggesting consolidation.
	parts := strings.Split(strings.TrimSuffix(filepath.Base(rel), ".go"), "_")

	platforms := []string{"aix", "android", "darwin", "dragonfly", "freebsd", "illumos", "ios", "js", "linux", "netbsd", "openbsd", "plan9", "solaris", "wasip1", "windows", "386", "amd64", "arm", "arm64", "loong64", "mips", "mipsle", "mips64", "mips64le", "ppc64", "ppc64le", "riscv64", "s390x", "wasm"}
	for _, part := range parts[1:] {
		if slices.Contains(platforms, part) {
			tags = append(tags, part)
		}
	}

	return strings.Join(tags, ";")
}

func (s *Service) reportScattered(files []string, idx int, group string, rep *dto.Report) {
	rep.Violations = append(rep.Violations, dto.Violation{
		Severity: "warning", Check: "layout-scattered", Rule: s.rules[idx].Path,
		Path: files[0], Pos: files[0], Message: fmt.Sprintf("%s declarations can be consolidated; spread across %s", group, strings.Join(files, ", ")),
	})
}

func (s *Service) checkSingleBinding(fset *token.FileSet, file *ast.File, rel string, idx int, rep *dto.Report) {
	names := s.declarationNames(file)
	for _, name := range names[min(1, len(names)):] {
		s.reportViolation(fset, name.Pos(), rel, idx, "file-declaration-count", "file may contain at most one top-level function, method, type, variable or constant when binding grouping is disabled", rep)
	}
}

// names counts every declared name, including names in grouped declarations,
// but excludes imports, struct fields, interface methods and local declarations.
func (*Service) declarationNames(file *ast.File) []*ast.Ident {
	var names []*ast.Ident

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			names = append(names, d.Name)
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch v := spec.(type) {
				case *ast.TypeSpec:
					names = append(names, v.Name)
				case *ast.ValueSpec:
					names = append(names, v.Names...)
				default:
				}
			}
		default:
		}
	}

	return names
}

// typeOwner uses declared types and receivers first. For a values-only
// file, choose the longest type-name prefix shared by all values in the file.
func (s *Service) typeOwner(pkg *dto.ASTPackage, file *ast.File) string {
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
			for _, spec := range gen.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					return ts.Name.Name
				}
			}
		}
	}

	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv != nil {
			return helpers.ReceiverName(fn.Recv.List[0].Type)
		}
	}

	values := s.valueNames(file)
	if len(values) == 0 {
		return ""
	}

	owner := ""

	for _, candidate := range pkg.Files {
		for _, decl := range candidate.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}

			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !s.hasType(pkg, ts.Name.Name) {
					continue
				}

				prefix := helpers.Normalize(ts.Name.Name)
				matches := true

				for _, value := range values {
					if !strings.HasPrefix(helpers.Normalize(value.Name), prefix) {
						matches = false

						break
					}
				}

				if matches && (len(prefix) > len(helpers.Normalize(owner)) || (len(prefix) == len(helpers.Normalize(owner)) && ts.Name.Name < owner)) {
					owner = ts.Name.Name
				}
			}
		}
	}

	return owner
}

func (s *Service) valueNames(file *ast.File) []*ast.Ident {
	var names []*ast.Ident

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if ok && (gen.Tok == token.VAR || gen.Tok == token.CONST) {
			names = append(names, s.declarationNames(&ast.File{Decls: []ast.Decl{gen}})...)
		}
	}

	return names
}

func (s *Service) checkDirectories(scope *dto.Scope, rep *dto.Report) {
	for rel, indices := range scope.Subdirs {
		for _, idx := range indices {
			r := s.rules[idx]
			if r.Options.Directories.Subdirs != nil && !*r.Options.Directories.Subdirs {
				rep.Violations = append(rep.Violations, dto.Violation{Severity: "error", Check: "fs-mode", Rule: r.Path, Path: rel, Pos: rel, Message: fmt.Sprintf("directory %q is not allowed: layout.directories.subdirs is false", rel)})
			}
		}
	}

	for rel, indices := range scope.FileDirectories {
		for _, idx := range indices {
			r := s.rules[idx]
			if r.Options.Directories.Flat != nil && !*r.Options.Directories.Flat {
				rep.Violations = append(rep.Violations, dto.Violation{Severity: "error", Check: "fs-mode", Rule: r.Path, Path: rel, Pos: rel, Message: fmt.Sprintf("file %q is not allowed: layout.directories.flat is false", rel)})
			}
		}
	}
}
