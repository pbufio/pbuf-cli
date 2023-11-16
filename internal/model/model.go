package model

type Config struct {
	Version  string `yaml:"version"`
	Name     string `yaml:"name"`
	Registry struct {
		Addr     string `yaml:"addr"`
		Insecure bool   `yaml:"insecure"`
	} `yaml:"registry"`
	Export struct {
		Paths []string `yaml:"paths"`
	} `yaml:"export"`
	Modules []*Module `yaml:"modules"`
}

type Module struct {
	Name                 string `yaml:"name"`
	Repository           string `yaml:"repository"`
	Path                 string `yaml:"path"`
	Branch               string `yaml:"branch"`
	Tag                  string `yaml:"tag"`
	OutputFolder         string `yaml:"out"`
	GenerateOutputFolder string `yaml:"gen_out"`
}

func (c *Config) HasRegistry() bool {
	return c.Registry.Addr != ""
}
