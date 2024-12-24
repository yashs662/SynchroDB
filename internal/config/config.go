package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Address            string `yaml:"address"`
		Password           string `yaml:"password"`
		AuthEnabled        bool   `yaml:"auth_enabled"`
		PersistentAOFPath  string `yaml:"persistent_aof_path"`
		ReplayAOFOnStartup bool   `yaml:"replay_aof_on_startup"`
		MaxConnections     int    `yaml:"max_connections"`
		RateLimit          int    `yaml:"rate_limit"`
	} `yaml:"server"`
	Log struct {
		File  string `yaml:"file"`
		Debug bool   `yaml:"debug"`
	} `yaml:"log"`
}

func LoadConfig() (*Config, error) {
	return LoadConfigFromPath("config/server.yaml")
}

func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return &config, nil
}
