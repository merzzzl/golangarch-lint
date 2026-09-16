package config

type Directories struct {
	Flat    *bool `yaml:"flat,omitempty"`
	Subdirs *bool `yaml:"subdirs,omitempty"`
}
