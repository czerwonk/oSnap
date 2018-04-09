package config

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config represents the configuration for the application
type Config struct {
	Keep        int        `yaml:"keep"`
	Description string     `yaml:"description"`
	Cluster     string     `yaml:"cluster"`
	API         *APIConfig `yaml:"api"`
	Includes    []string   `yaml:"includes"`
	Excludes    []string   `yaml:"excludes"`
}

// APIConfig contains all settings to communicate with the oVirt API
type APIConfig struct {
	URL      string `yaml:"url"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Insecure bool   `yaml:"insecure"`
}

// Load reads config from reader
func Load(reader io.Reader) (*Config, error) {
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	err = yaml.Unmarshal(b, cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
