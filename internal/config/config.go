package config

type Config struct {
	Version int      `yaml:"version"`
	Ignore  []string `yaml:"ignore"`
	Rules   []struct {
		Path            string   `yaml:"path"`
		Ignore          []string `yaml:"ignore"`
		Mode            string   `yaml:"mode"`
		AllowTypes      string   `yaml:"allow-types"`
		AllowVars       string   `yaml:"allow-vars"`
		AllowFuncs      string   `yaml:"allow-funcs"`
		ExcludeTypes    []string `yaml:"exclude-types"`
		ExcludeVars     []string `yaml:"exclude-vars"`
		ExcludeFuncs    []string `yaml:"exclude-funcs"`
		RequireReceiver string   `yaml:"require-receiver"`
		Exported        *struct {
			Inputs  []string `yaml:"inputs"`
			Outputs []string `yaml:"outputs"`
		} `yaml:"exported"`
		Modules []string `yaml:"modules"`
	} `yaml:"rules"`
}
