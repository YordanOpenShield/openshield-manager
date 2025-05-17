package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

var GlobalConfig Config

type Config struct {
	ENVIRONMENT string `yaml:"ENVIRONMENT"`
	DB_HOST     string `yaml:"DB_HOST"`
	DB_PORT     string `yaml:"DB_PORT"`
	DB_USER     string `yaml:"DB_USER"`
	DB_PASSWORD string `yaml:"DB_PASSWORD"`
	DB_NAME     string `yaml:"DB_NAME"`
	DB_SSLMODE  string `yaml:"DB_SSLMODE"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func LoadAndSetConfig(path string) error {
	cfg, err := LoadConfig(path)
	if err != nil {
		return err
	}
	GlobalConfig = *cfg
	return nil
}
