package fscheck

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/dto"
	"github.com/merzzzl/golangarch-lint/internal/helpers"
)

func (s *Service) fsCheckDir(rel string, rep *dto.Report) {
	for i := range s.cfg.Rules {
		r := &s.cfg.Rules[i]

		if r.Mode == "" || r.Mode == "any" {
			continue
		}

		parent := filepath.ToSlash(filepath.Dir(rel))
		if !helpers.GlobMatch(r.Path, parent) {
			continue
		}

		if s.fsCheckIsRuleIgnored(rel, r.Ignore) {
			continue
		}

		if r.Mode == "flat" {
			rep.Violations = append(rep.Violations, dto.Violation{
				Check:   "fs-mode",
				Rule:    r.Path,
				Path:    rel,
				Pos:     rel,
				Message: fmt.Sprintf("directory %q is not allowed: rule requires flat structure (no subdirectories)", rel),
			})
		}
	}
}

func (s *Service) fsCheckFile(rel string, rep *dto.Report) {
	name := filepath.Base(rel)
	if strings.HasSuffix(name, "_test.go") || !strings.HasSuffix(name, ".go") {
		return
	}

	dir := filepath.ToSlash(filepath.Dir(rel))

	if !s.fsCheckIsCovered(dir) {
		rep.Violations = append(rep.Violations, dto.Violation{
			Check:   "fs-coverage",
			Rule:    "",
			Path:    rel,
			Pos:     rel,
			Message: fmt.Sprintf("directory %q is not covered by any fs rule", dir),
		})

		return
	}

	for i := range s.cfg.Rules {
		r := &s.cfg.Rules[i]

		if r.Mode == "subdirs-only" && helpers.GlobMatch(r.Path, dir) && !s.fsCheckIsRuleIgnored(rel, r.Ignore) {
			rep.Violations = append(rep.Violations, dto.Violation{
				Check:   "fs-mode",
				Rule:    r.Path,
				Path:    rel,
				Pos:     rel,
				Message: fmt.Sprintf("file %q is not allowed: rule requires subdirs-only structure (no files)", name),
			})
		}
	}
}

func (s *Service) fsCheckIsCovered(dir string) bool {
	for i := range s.cfg.Rules {
		if helpers.GlobMatch(s.cfg.Rules[i].Path, dir) {
			return true
		}
	}

	return false
}

func (*Service) fsCheckIsHidden(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (*Service) fsCheckIsIgnored(rel string, patterns []string) bool {
	for _, p := range patterns {
		if helpers.GlobMatch(p, rel) {
			return true
		}
	}

	return false
}

func (*Service) fsCheckIsRuleIgnored(rel string, ignore []string) bool {
	for _, p := range ignore {
		if helpers.GlobMatch(p, rel) {
			return true
		}
	}

	return false
}
