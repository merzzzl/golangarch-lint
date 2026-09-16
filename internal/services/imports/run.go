package imports

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/config"
	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/helpers"
	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/vta"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// Run checks imported packages and direct call targets against one policy.
func (s *Service) Run(input *dto.ASTInput, scope *dto.Scope, rep *dto.Report) error {
	if err := s.checkImports(input, scope, rep); err != nil {
		return err
	}

	return s.checkCalls(input, scope, rep)
}

func (s *Service) checkImports(input *dto.ASTInput, scope *dto.Scope, rep *dto.Report) error {
	for _, pkg := range input.Packages {
		for j, file := range pkg.Files {
			rel, err := filepath.Rel(scope.Root, pkg.GoFiles[j])
			if err != nil {
				return fmt.Errorf("imports path: %w", err)
			}

			rel = filepath.ToSlash(rel)

			for _, imp := range file.Imports {
				name, err := strconv.Unquote(imp.Path.Value)
				if err != nil {
					return fmt.Errorf("import path: %w", err)
				}

				for _, idx := range scope.Files[rel] {
					r := s.rules[idx]
					if !s.isAllowed(r.Options, name, "") {
						pos := input.Fset.Position(imp.Pos())
						rep.Violations = append(rep.Violations, dto.Violation{Severity: "error", Check: "import-denied", Rule: r.Path, Path: rel, Pos: fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column), Message: fmt.Sprintf("import %q is not allowed by imports for rule %q", name, r.Path)})

						break
					}
				}
			}
		}
	}

	return nil
}

// checkCalls checks call sites, not transitive reachability. SSA wrappers are resolved
// only to recover the source-level target of a method value or interface call.
func (s *Service) checkCalls(input *dto.ASTInput, scope *dto.Scope, rep *dto.Report) error {
	enabled := false

	for i := range s.rules {
		r := &s.rules[i]
		if r.Options.Allow != nil || len(r.Options.Deny) > 0 {
			enabled = true

			break
		}
	}

	if !enabled {
		return nil
	}

	prog, _ := ssautil.AllPackages(input.LoadedPackages, ssa.InstantiateGenerics)
	prog.Build()
	funcs := ssautil.AllFunctions(prog)
	graph := vta.CallGraph(funcs, nil)
	seen := map[string]bool{}

	for fn := range funcs {
		if fn.Syntax() == nil && fn.Synthetic != "package initializer" {
			continue
		}

		caller := s.callerPackage(fn)

		for _, block := range fn.Blocks {
			for _, instruction := range block.Instrs {
				site, ok := instruction.(ssa.CallInstruction)
				if !ok || !site.Pos().IsValid() {
					continue
				}

				if _, builtin := site.Common().Value.(*ssa.Builtin); builtin {
					continue
				}

				pos := input.Fset.Position(site.Pos())

				rel, err := filepath.Rel(input.Root, pos.Filename)
				if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
					continue
				}

				rel = filepath.ToSlash(rel)
				if strings.HasSuffix(rel, "_test.go") {
					continue
				}

				indices := scope.Files[rel]

				targets := s.targets(graph, fn, site)
				for _, idx := range indices {
					r := &s.rules[idx]
					if r.Options.Allow == nil && len(r.Options.Deny) == 0 {
						continue
					}

					if len(targets) == 0 {
						key := fmt.Sprintf("%s:%d:%d:%d:unresolved", rel, pos.Line, pos.Column, idx)
						if !seen[key] {
							seen[key] = true

							rep.Violations = append(rep.Violations, dto.Violation{Severity: "warning", Check: "call-unresolved", Rule: r.Path, Path: rel, Pos: fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column), Message: "cannot resolve the target of this call; imports policy could not be fully checked"})
						}
					}

					for target, name := range targets {
						if s.isAllowed(r.Options, target, caller) {
							continue
						}

						key := fmt.Sprintf("%s:%d:%d:%d:%s", rel, pos.Line, pos.Column, idx, target)
						if seen[key] {
							continue
						}

						seen[key] = true

						rep.Violations = append(rep.Violations, dto.Violation{
							Severity: "error", Check: "call-denied", Rule: r.Path, Path: rel,
							Pos:     fmt.Sprintf("%s:%d:%d", rel, pos.Line, pos.Column),
							Message: fmt.Sprintf("possible direct call to %s (package %q) is not allowed by imports", name, target),
						})
					}
				}
			}
		}
	}

	return nil
}

func (s *Service) targets(graph *callgraph.Graph, fn *ssa.Function, site ssa.CallInstruction) map[string]string {
	targets := map[string]string{}

	visited := map[*ssa.Function]bool{}
	if callee := site.Common().StaticCallee(); callee != nil {
		s.resolve(graph, callee, visited, targets)
	}

	if node := graph.Nodes[fn]; node != nil {
		for _, edge := range node.Out {
			if edge.Site == site {
				s.resolve(graph, edge.Callee.Func, visited, targets)
			}
		}
	}

	return targets
}

func (s *Service) resolve(graph *callgraph.Graph, fn *ssa.Function, visited map[*ssa.Function]bool, targets map[string]string) {
	if fn == nil || visited[fn] {
		return
	}

	visited[fn] = true
	if object := fn.Object(); object != nil && object.Pkg() != nil {
		s.addTarget(targets, object.Pkg().Path(), fn.String())

		return
	}

	if fn.Synthetic != "" {
		if node := graph.Nodes[fn]; node != nil {
			for _, edge := range node.Out {
				s.resolve(graph, edge.Callee.Func, visited, targets)
			}
		}

		return
	}

	for current := fn; current != nil; current = current.Parent() {
		if current.Pkg != nil {
			s.addTarget(targets, current.Pkg.Pkg.Path(), fn.String())

			return
		}
	}
}

func (*Service) addTarget(targets map[string]string, pkg, name string) {
	if previous, ok := targets[pkg]; !ok || name < previous {
		targets[pkg] = name
	}
}

// isAllowed gives explicit deny priority over stdlib and same-package defaults.
func (*Service) isAllowed(filter config.Filter, name, caller string) bool {
	for _, pattern := range filter.Deny {
		if helpers.GlobMatch(pattern, name) {
			return false
		}
	}

	return (caller != "" && name == caller) || helpers.IsStdlibPackage(name) || filter.Matches(name)
}

func (*Service) callerPackage(fn *ssa.Function) string {
	for current := fn; current != nil; current = current.Parent() {
		if current.Pkg != nil {
			return current.Pkg.Pkg.Path()
		}

		if object := current.Object(); object != nil && object.Pkg() != nil {
			return object.Pkg().Path()
		}
	}

	return ""
}
