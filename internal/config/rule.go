package config

type Rule struct {
	Scope        Scope        `yaml:"scope"`
	Layout       Layout       `yaml:"layout,omitempty"`
	Declarations Declarations `yaml:"declarations,omitempty"`
	Imports      Filter       `yaml:"imports,omitempty"`
	Signatures   Signatures   `yaml:"signatures,omitempty"`
}
