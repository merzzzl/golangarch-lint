package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidFilenames   = errors.New("invalid filenames pattern")
	ErrUnsupportedVersion = errors.New("unsupported config version")
	ErrMultipleDocuments  = errors.New("config must contain exactly one YAML document")
	ErrConfigNotFound     = errors.New("config not found")
	ErrPathRequired       = errors.New("path is required")
	ErrOverlappingPaths   = errors.New("overlapping paths")
	ErrInvalidFileBinding = errors.New("invalid file-binding")
	ErrInvalidMode        = errors.New("invalid mode")
	ErrInvalidAllowMode   = errors.New("invalid allow mode")
	ErrModuleNotFound     = errors.New("module path not found in go.mod")
)

// Load discovers (when configPath is empty), reads, substitutes $module,
// parses and validates the config for the project at root.
func Load(root, configPath string) (*Config, error) {
	if configPath == "" {
		names := []string{".golangarch.yml", ".golangarch.yaml"}

		for _, name := range names {
			p := filepath.Join(root, name)
			if _, err := os.Stat(p); err == nil {
				configPath = p

				break
			}
		}

		if configPath == "" {
			return nil, fmt.Errorf("%w: tried %v in %s", ErrConfigNotFound, names, root)
		}
	}

	modData, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}

	modulePath := ""

	for line := range strings.SplitSeq(string(modData), "\n") {
		if mod, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			modulePath = strings.TrimSpace(mod)

			break
		}
	}

	if modulePath == "" {
		return nil, ErrModuleNotFound
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	content := strings.ReplaceAll(string(data), "$module", modulePath)

	var parser Config

	cfg, notes, err := parser.Parse([]byte(content))
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		_, _ = fmt.Fprintf(os.Stderr, "migration warning: %s\n", note)
	}

	return cfg, nil
}
