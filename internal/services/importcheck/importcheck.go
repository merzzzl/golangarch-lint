package importcheck

import (
	"fmt"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/helpers"
)

func (s *Service) importCheckSingle(importerRel string, imp dto.ImportEntry, rep *dto.Report) {
	if s.importCheckIsStdlib(imp.Path) {
		return
	}

	if s.importCheckIsRuleIgnored(imp.File, s.cfg.Ignore) {
		return
	}

	for _, idx := range s.importCheckSelectRules(imp.File, importerRel) {
		r := &s.cfg.Rules[idx]

		if r.Modules == nil {
			continue
		}

		if s.importCheckIsRuleIgnored(imp.File, r.Ignore) {
			continue
		}

		if !s.importCheckIsListed(r.Modules, imp.Path) {
			rep.Violations = append(rep.Violations, dto.Violation{
				Check:   "import-denied",
				Rule:    r.Path,
				Path:    imp.File,
				Pos:     imp.Pos,
				Message: fmt.Sprintf("import %q is not in modules list for rule %q", imp.Path, r.Path),
			})

			return
		}
	}
}

// importCheckSelectRules returns indices of rules applicable to a file: rules
// whose path matches the file's relative path take priority over rules
// matching the importer package directory.
func (s *Service) importCheckSelectRules(rel, importerRel string) []int {
	var fileRules, dirRules []int

	for i := range s.cfg.Rules {
		r := &s.cfg.Rules[i]

		switch {
		case helpers.GlobMatch(r.Path, rel):
			fileRules = append(fileRules, i)
		case helpers.GlobMatch(r.Path, importerRel):
			dirRules = append(dirRules, i)
		default:
		}
	}

	if len(fileRules) > 0 {
		return fileRules
	}

	return dirRules
}

func (*Service) importCheckIsListed(modules []string, importPath string) bool {
	for _, pattern := range modules {
		if helpers.GlobMatch(pattern, importPath) {
			return true
		}
	}

	return false
}

func (*Service) importCheckIsRuleIgnored(rel string, ignore []string) bool {
	for _, p := range ignore {
		if helpers.GlobMatch(p, rel) {
			return true
		}
	}

	return false
}

func (*Service) importCheckIsStdlib(p string) bool {
	firstSlash := strings.Index(p, "/")

	first := p
	if firstSlash > 0 {
		first = p[:firstSlash]
	}

	return !strings.Contains(first, ".")
}
