package config

type Filter struct {
	Allow *[]string `yaml:"allow,omitempty"`
	Deny  []string  `yaml:"deny,omitempty"`
}

// UnmarshalYAML accepts an allow-list shorthand as well as allow/deny mappings.
// The decoder callback preserves KnownFields checking, aliases and merge keys.
func (f *Filter) UnmarshalYAML(unmarshal func(any) error) error {
	var allow []string
	if err := unmarshal(&allow); err == nil {
		*f = Filter{Allow: &allow}

		return nil
	}

	type full Filter

	var decoded full
	if err := unmarshal(&decoded); err != nil {
		return err
	}

	*f = Filter(decoded)

	return nil
}

func (f *Filter) Matches(name string) bool {
	for _, pattern := range f.Deny {
		if (Scope{Path: pattern}).Matches(name) {
			return false
		}
	}

	if f.Allow == nil {
		return true
	}

	for _, pattern := range *f.Allow {
		if (Scope{Path: pattern}).Matches(name) {
			return true
		}
	}

	return false
}
