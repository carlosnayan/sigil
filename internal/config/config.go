package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config espelha vault.yaml do projeto.
type Config struct {
	Project string            `yaml:"project"`
	Env     string            `yaml:"env"`
	Vaults  []string          `yaml:"vaults"`
	Inject  map[string]string `yaml:"inject"`
}

// Load lê vault.yaml do disco. Stub: retorna config vazia se path vazio ou erro de leitura nas fases seguintes.
func Load(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{}, nil
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return &Config{}, err
	}
	return &c, nil
}

// Save grava vault.yaml. Stub mínimo para compilação; comportamento completo nas fases seguintes.
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
