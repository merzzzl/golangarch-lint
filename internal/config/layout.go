package config

type Layout struct {
	Filenames   *[]string   `yaml:"filenames,omitempty"`
	Directories Directories `yaml:"directories,omitempty"`
	Binding     Binding     `yaml:"binding,omitempty"`
}
