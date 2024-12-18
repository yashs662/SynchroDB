package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Address     string `yaml:"address"`
		Password    string `yaml:"password"`
		AuthEnabled bool   `yaml:"auth_enabled"`
	} `yaml:"server"`
	Log struct {
		File  string `yaml:"file"`
		Debug bool   `yaml:"debug"`
	} `yaml:"log"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("config/server.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
