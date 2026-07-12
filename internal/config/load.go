package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
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

	var cfg Config
	if err := yaml.Unmarshal([]byte(content), &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}
