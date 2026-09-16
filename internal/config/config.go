package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"slices"
	"strings"

	"github.com/merzzzl/golangarch-lint/internal/migrations"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version int    `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

// Parse strictly decodes v2 or migrates v1 in memory, returning migration notes.
func (c *Config) Parse(data []byte) (*Config, []string, error) {
	const currentVersion = 2

	var header struct {
		Version int `yaml:"version"`
	}
	if err := yaml.Unmarshal(data, &header); err != nil {
		return nil, nil, fmt.Errorf("parsing config: %w", err)
	}

	switch header.Version {
	case 1:
		migrated, notes, err := migrations.Migrate(data)
		if err != nil {
			return nil, nil, fmt.Errorf("migrating config: %w", err)
		}

		cfg, _, err := c.Parse(migrated)

		return cfg, notes, err
	case currentVersion:
		var cfg Config
		if err := c.parseStrict(data, &cfg); err != nil {
			return nil, nil, err
		}

		if err := cfg.Validate(); err != nil {
			return nil, nil, err
		}

		return &cfg, nil, nil
	default:
		return nil, nil, fmt.Errorf("%w %d: expected 1 or 2", ErrUnsupportedVersion, header.Version)
	}
}

// Migrate produces v2 YAML without expanding $module. It never writes the source.
func (c *Config) Migrate(data []byte) ([]byte, []string, error) {
	cfg, notes, err := c.Parse(data)
	if err != nil {
		return nil, nil, err
	}

	output, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding v2 config: %w", err)
	}

	return output, notes, nil
}

func (c *Config) Validate() error {
	const currentVersion = 2

	if c.Version != currentVersion {
		return fmt.Errorf("%w %d: expected 2", ErrUnsupportedVersion, c.Version)
	}

	for i := range c.Rules {
		r := &c.Rules[i]
		if r.Scope.Path == "" {
			return fmt.Errorf("rule #%d: %w", i+1, ErrPathRequired)
		}

		if err := c.validateGlob(r.Scope.Path); err != nil {
			return err
		}

		if r.Layout.Filenames != nil {
			for _, pattern := range *r.Layout.Filenames {
				if pattern == "" || strings.ContainsAny(pattern, "/\\") {
					return fmt.Errorf("%w: rule %q: %q must be a basename glob", ErrInvalidFilenames, r.Scope.Path, pattern)
				}

				if _, err := path.Match(pattern, ""); err != nil {
					return fmt.Errorf("rule %q: invalid filenames pattern %q: %w", r.Scope.Path, pattern, err)
				}
			}
		}

		patterns := slices.Clone(r.Scope.Ignore)

		filters := []Filter{r.Imports}
		for _, v := range []Visibility{r.Declarations.Types, r.Declarations.Vars, r.Declarations.Consts, r.Declarations.Funcs} {
			filters = append(filters, v.Exported, v.Unexported)
		}

		for _, f := range filters {
			if f.Allow != nil {
				patterns = append(patterns, (*f.Allow)...)
			}

			patterns = append(patterns, f.Deny...)
		}

		for _, s := range []Signature{r.Signatures.Exported, r.Signatures.Unexported} {
			if s.Inputs != nil {
				patterns = append(patterns, (*s.Inputs)...)
			}

			if s.Outputs != nil {
				patterns = append(patterns, (*s.Outputs)...)
			}
		}

		for _, p := range patterns {
			if err := c.validateGlob(p); err != nil {
				return fmt.Errorf("rule %q: %w", r.Scope.Path, err)
			}
		}
	}

	for i := range c.Rules {
		for j := i + 1; j < len(c.Rules); j++ {
			if c.Rules[i].Scope.Overlaps(c.Rules[j].Scope) {
				return fmt.Errorf("%w: %q and %q", ErrOverlappingPaths, c.Rules[i].Scope.Path, c.Rules[j].Scope.Path)
			}
		}
	}

	return nil
}

func (*Config) parseStrict(data []byte, target any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	var extra any
	if err := dec.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrMultipleDocuments
	}

	return nil
}

func (*Config) validateGlob(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty glob pattern: %w", ErrInvalidMode)
	}

	for seg := range strings.SplitSeq(pattern, "/") {
		if seg == "**" {
			continue
		}

		if _, err := path.Match(seg, "test"); err != nil {
			return fmt.Errorf("invalid glob %q: %w", pattern, err)
		}
	}

	return nil
}
