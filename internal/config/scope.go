package config

import (
	"path"

	"github.com/merzzzl/golangarch-lint/internal/helpers"
)

type Scope struct {
	Path   string   `yaml:"path"`
	Ignore []string `yaml:"ignore,omitempty"`
}

// Ignored includes descendants of ignored directories. Patterns are root-relative.
func (s Scope) Ignored(name string) bool {
	for {
		for _, pattern := range s.Ignore {
			if (Scope{Path: pattern}).Matches(name) {
				return true
			}
		}

		if name == "." || name == "/" {
			return false
		}

		name = path.Dir(name)
	}
}

func (s Scope) Matches(name string) bool  { return helpers.GlobMatch(s.Path, name) }
func (s Scope) Overlaps(other Scope) bool { return helpers.GlobIntersects(s.Path, other.Path) }
