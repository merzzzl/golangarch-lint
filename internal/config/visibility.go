package config

import (
	"slices"
)

type Visibility struct {
	Exported   Filter `yaml:"exported,omitempty"`
	Unexported Filter `yaml:"unexported,omitempty"`
}

// UnmarshalYAML applies a list to both exported and unexported declarations.
func (v *Visibility) UnmarshalYAML(unmarshal func(any) error) error {
	var allow []string
	if err := unmarshal(&allow); err == nil {
		private := slices.Clone(allow)
		*v = Visibility{Exported: Filter{Allow: &allow}, Unexported: Filter{Allow: &private}}

		return nil
	}

	type full Visibility

	var decoded full
	if err := unmarshal(&decoded); err != nil {
		return err
	}

	*v = Visibility(decoded)

	return nil
}
