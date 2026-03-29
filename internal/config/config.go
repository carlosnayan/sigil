package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Project string            `yaml:"project"`
	Env     string            `yaml:"env"`
	Secret  string            `yaml:"secret"`
	Vaults  []string          `yaml:"vaults"`
	Inject  map[string]string `yaml:"inject"`
	Links   map[string]string `yaml:"links,omitempty"`
}

func Load(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	return &c, nil
}

func Save(path string, c *Config) error {
	if c == nil {
		c = &Config{}
	}
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}
