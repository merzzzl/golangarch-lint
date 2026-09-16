package config

type Declarations struct {
	Types  Visibility `yaml:"types,omitempty"`
	Vars   Visibility `yaml:"vars,omitempty"`
	Consts Visibility `yaml:"consts,omitempty"`
	Funcs  Visibility `yaml:"funcs,omitempty"`
}
