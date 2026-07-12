package config

import (
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/helpers"
)

// Validate checks the whole config: globs, enum values and rule path overlaps.
func (c *Config) Validate() error {
	validModes := []string{"any", "flat", "subdirs-only"}
	validAllowModes := []string{"all", "local", "exported", "none"}

	validateGlob := func(pattern string) error {
		for seg := range strings.SplitSeq(pattern, "/") {
			if seg == "**" {
				continue
			}

			if _, err := path.Match(seg, "test"); err != nil {
				return fmt.Errorf("invalid glob segment %q: %w", seg, err)
			}
		}

		return nil
	}

	for _, ig := range c.Ignore {
		if err := validateGlob(ig); err != nil {
			return fmt.Errorf("global ignore: invalid glob %q: %w", ig, err)
		}
	}

	for i := range c.Rules {
		r := &c.Rules[i]

		if r.Path == "" {
			return fmt.Errorf("rule #%d: %w", i+1, ErrPathRequired)
		}

		if err := validateGlob(r.Path); err != nil {
			return fmt.Errorf("rule %q: invalid path glob: %w", r.Path, err)
		}

		for _, ig := range r.Ignore {
			if err := validateGlob(ig); err != nil {
				return fmt.Errorf("rule %q: invalid ignore glob %q: %w", r.Path, ig, err)
			}
		}

		if r.Mode != "" && !slices.Contains(validModes, r.Mode) {
			return fmt.Errorf("rule %q: %w: %q, must be one of: %v", r.Path, ErrInvalidMode, r.Mode, validModes)
		}

		allowFields := map[string]string{
			"allow-types":      r.AllowTypes,
			"allow-vars":       r.AllowVars,
			"allow-funcs":      r.AllowFuncs,
			"require-receiver": r.RequireReceiver,
		}

		for field, v := range allowFields {
			if v != "" && !slices.Contains(validAllowModes, v) {
				return fmt.Errorf("rule %q: %w: %s: %q, must be one of: %v", r.Path, ErrInvalidAllowMode, field, v, validAllowModes)
			}
		}

		for _, m := range r.Modules {
			if err := validateGlob(m); err != nil {
				return fmt.Errorf("rule %q: invalid modules glob %q: %w", r.Path, m, err)
			}
		}

		excludeFields := map[string][]string{
			"exclude-types": r.ExcludeTypes,
			"exclude-vars":  r.ExcludeVars,
			"exclude-funcs": r.ExcludeFuncs,
		}

		for field, patterns := range excludeFields {
			for _, p := range patterns {
				if err := validateGlob(p); err != nil {
					return fmt.Errorf("rule %q: invalid %s glob %q: %w", r.Path, field, p, err)
				}
			}
		}
	}

	for i := range c.Rules {
		for j := i + 1; j < len(c.Rules); j++ {
			if helpers.GlobIntersects(c.Rules[i].Path, c.Rules[j].Path) {
				return fmt.Errorf("%w: rules %q and %q", ErrOverlappingPaths, c.Rules[i].Path, c.Rules[j].Path)
			}
		}
	}

	return nil
}
