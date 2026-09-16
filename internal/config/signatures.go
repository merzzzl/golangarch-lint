package config

type Signatures struct {
	Exported   Signature `yaml:"exported,omitempty"`
	Unexported Signature `yaml:"unexported,omitempty"`
}
