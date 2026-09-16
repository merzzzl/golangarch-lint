package config

type Binding struct {
	Name     *bool `yaml:"name,omitempty"`
	Receiver *bool `yaml:"receiver,omitempty"`
}

// UnmarshalYAML accepts false as shorthand for explicitly disabled grouping.
// An omitted binding or an empty mapping still disables binding checks entirely.
func (b *Binding) UnmarshalYAML(unmarshal func(any) error) error {
	var enabled bool
	if err := unmarshal(&enabled); err == nil && !enabled {
		name, receiver := false, false
		*b = Binding{Name: &name, Receiver: &receiver}

		return nil
	}

	type full Binding

	var decoded full
	if err := unmarshal(&decoded); err != nil {
		return err
	}

	*b = Binding(decoded)

	return nil
}

// Modes distinguishes disabled checking from explicitly disabled grouping.
// Once either flag is specified, an omitted partner does not enable grouping.
func (b *Binding) Modes() []string {
	if b.Name == nil && b.Receiver == nil {
		return nil
	}

	var modes []string
	if b.Name != nil && *b.Name {
		modes = append(modes, "name")
	}

	if b.Receiver != nil && *b.Receiver {
		modes = append(modes, "receiver")
	}

	if len(modes) == 0 {
		return []string{"single"}
	}

	return modes
}
