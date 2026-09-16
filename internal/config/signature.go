package config

type Signature struct {
	Receiver *bool     `yaml:"receiver,omitempty"`
	Inputs   *[]string `yaml:"inputs,omitempty"`
	Outputs  *[]string `yaml:"outputs,omitempty"`
}
