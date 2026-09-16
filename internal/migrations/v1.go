package migrations

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// V1 is retained only for reading and migrating legacy configuration.
type V1 struct {
	Version int      `yaml:"version"`
	Ignore  []string `yaml:"ignore"`
	Rules   []struct {
		Path            string   `yaml:"path"`
		Ignore          []string `yaml:"ignore"`
		Mode            string   `yaml:"mode"`
		FileBinding     string   `yaml:"file-binding"`
		AllowTypes      string   `yaml:"allow-types"`
		AllowVars       string   `yaml:"allow-vars"`
		AllowFuncs      string   `yaml:"allow-funcs"`
		ExcludeTypes    []string `yaml:"exclude-types"`
		ExcludeVars     []string `yaml:"exclude-vars"`
		ExcludeFuncs    []string `yaml:"exclude-funcs"`
		RequireReceiver string   `yaml:"require-receiver"`
		Exported        *struct {
			Inputs  *[]string `yaml:"inputs"`
			Outputs *[]string `yaml:"outputs"`
		} `yaml:"exported"`
		Modules *[]string `yaml:"modules"`
	} `yaml:"rules"`
}

func (c *V1) Migrate() ([]byte, []string, error) {
	const targetVersion = 2

	rules := make([]map[string]any, 0, len(c.Rules))

	notes := []string{"imports now constrains direct call targets, including interface implementations, as well as imported packages; review dependencies that v1 did not check", "binding now also checks types, variables and constants; review files that grouped these declarations without filename restrictions in v1"}

	for i := range c.Rules {
		old := &c.Rules[i]
		for _, mode := range []string{old.AllowTypes, old.AllowVars, old.AllowFuncs, old.RequireReceiver} {
			if !slices.Contains([]string{"", "all", "none", "local", "exported"}, mode) {
				return nil, nil, fmt.Errorf("rule %q: %w %q", old.Path, ErrInvalidMode, mode)
			}
		}

		binding := old.FileBinding
		if binding == "" {
			binding = "name"
		}

		scope := map[string]any{"path": old.Path}

		ignores := append(slices.Clone(c.Ignore), old.Ignore...)
		if len(ignores) > 0 {
			scope["ignore"] = ignores
		}

		if !slices.Contains([]string{"name", "type"}, binding) {
			return nil, nil, fmt.Errorf("%w: file-binding %q", ErrInvalidMode, binding)
		}

		if !slices.Contains([]string{"", "any", "flat", "subdirs-only"}, old.Mode) {
			return nil, nil, fmt.Errorf("%w: mode %q", ErrInvalidMode, old.Mode)
		}

		layout := map[string]any{"binding": map[string]bool{"name": binding == "name", "receiver": binding == "type"}}
		if old.Mode != "" {
			layout["directories"] = map[string]bool{"flat": old.Mode != "subdirs-only", "subdirs": old.Mode != "flat"}
		}

		r := map[string]any{"scope": scope, "layout": layout}
		declarations := map[string]any{}

		for _, entry := range []struct {
			kind, mode string
			except     []string
		}{
			{"types", old.AllowTypes, old.ExcludeTypes},
			{"vars", old.AllowVars, old.ExcludeVars},
			{"consts", old.AllowVars, old.ExcludeVars},
			{"funcs", old.AllowFuncs, old.ExcludeFuncs},
		} {
			visibility := map[string]any{}

			allow := slices.Clone(entry.except)
			if allow == nil {
				allow = []string{}
			}

			if entry.mode == "none" || entry.mode == "local" {
				visibility["exported"] = map[string]any{"allow": allow}
			}

			if entry.mode == "none" || entry.mode == "exported" {
				visibility["unexported"] = map[string]any{"allow": allow}
			}

			if len(visibility) > 0 {
				declarations[entry.kind] = visibility
			}
		}

		if len(declarations) > 0 {
			r["declarations"] = declarations
		}

		exported, unexported := map[string]any{}, map[string]any{}
		if old.RequireReceiver == "all" || old.RequireReceiver == "exported" {
			exported["receiver"] = true
		}

		if old.RequireReceiver == "all" || old.RequireReceiver == "local" {
			unexported["receiver"] = true
		}

		if old.Exported != nil {
			if old.Exported.Inputs != nil {
				exported["inputs"] = *old.Exported.Inputs
			}

			if old.Exported.Outputs != nil {
				exported["outputs"] = *old.Exported.Outputs
			}
		}

		signatures := map[string]any{}
		if len(exported) > 0 {
			signatures["exported"] = exported
		}

		if len(unexported) > 0 {
			signatures["unexported"] = unexported
		}

		if len(signatures) > 0 {
			r["signatures"] = signatures
		}

		if old.Modules != nil {
			r["imports"] = map[string]any{"allow": *old.Modules}
		}

		if len(old.ExcludeFuncs) > 0 {
			notes = append(notes, fmt.Sprintf("rule %q: exclude-funcs %v now permits declarations only; review layout.binding and signatures, whose old exemptions cannot be preserved", old.Path, old.ExcludeFuncs))
		}

		if len(old.ExcludeTypes) > 0 && binding == "type" {
			notes = append(notes, fmt.Sprintf("rule %q: exclude-types no longer exempts type file binding", old.Path))
		}

		if len(old.Ignore) > 0 {
			notes = append(notes, fmt.Sprintf("rule %q: scope.ignore now skips all checks of the selected rule, including file binding", old.Path))
		}

		rules = append(rules, r)
	}

	if len(c.Ignore) > 0 {
		notes = append(notes, "global ignore copied to each scope.ignore; ignored paths outside all scopes still require coverage")
	}

	output, err := yaml.Marshal(map[string]any{"version": targetVersion, "rules": rules})
	if err != nil {
		return nil, nil, fmt.Errorf("encoding migrated config: %w", err)
	}

	return output, notes, nil
}
