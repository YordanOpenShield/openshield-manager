package config

import (
	"os"
	"runtime"

	"gopkg.in/yaml.v2"
)

var ConfigPath string = "config"
var ScriptsPath string = "scripts"
var CertsPath string = "certs"

func init() {
	switch runtime.GOOS {
	case "windows":
		ConfigPath = "C:\\ProgramData\\openshield\\config"
		ScriptsPath = "C:\\ProgramData\\openshield\\scripts"
		CertsPath = "C:\\ProgramData\\openshield\\certs"
	default:
		ConfigPath = "/etc/openshield/config"
		ScriptsPath = "/etc/openshield/scripts"
		CertsPath = "/etc/openshield/certs"
	}
}

var GlobalConfig Config

type Config struct {
	ENVIRONMENT string `yaml:"ENVIRONMENT"`
	DB_HOST     string `yaml:"DB_HOST"`
	DB_PORT     string `yaml:"DB_PORT"`
	DB_USER     string `yaml:"DB_USER"`
	DB_PASSWORD string `yaml:"DB_PASSWORD"`
	DB_NAME     string `yaml:"DB_NAME"`
	DB_SSLMODE  string `yaml:"DB_SSLMODE"`
	HTTP_PORT   string `yaml:"HTTP_PORT"`
	TLS_ENABLED bool   `yaml:"TLS_ENABLED"`
	JWT_SECRET  string `yaml:"JWT_SECRET"`
}

func GenerateConfig(opts Config) *Config {
	return &Config{
		ENVIRONMENT: func() string {
			if opts.ENVIRONMENT != "" {
				return opts.ENVIRONMENT
			} else {
				return "development"
			}
		}(),
		DB_HOST: func() string {
			if opts.DB_HOST != "" {
				return opts.DB_HOST
			} else {
				return "localhost"
			}
		}(),
		DB_PORT: func() string {
			if opts.DB_PORT != "" {
				return opts.DB_PORT
			} else {
				return "5432"
			}
		}(),
		DB_USER: func() string {
			if opts.DB_USER != "" {
				return opts.DB_USER
			} else {
				return "user"
			}
		}(),
		DB_PASSWORD: func() string {
			if opts.DB_PASSWORD != "" {
				return opts.DB_PASSWORD
			} else {
				return "pass"
			}
		}(),
		DB_NAME: func() string {
			if opts.DB_NAME != "" {
				return opts.DB_NAME
			} else {
				return "openshield"
			}
		}(),
		DB_SSLMODE: func() string {
			if opts.DB_SSLMODE != "" {
				return opts.DB_SSLMODE
			} else {
				return "disable"
			}
		}(),
		HTTP_PORT: func() string {
			if opts.HTTP_PORT != "" {
				return opts.HTTP_PORT
			} else {
				return "9000"
			}
		}(),
		TLS_ENABLED: opts.TLS_ENABLED,
		JWT_SECRET: func() string {
			if opts.JWT_SECRET != "" {
				return opts.JWT_SECRET
			} else {
				return "change-me-in-production"
			}
		}(),
	}
}

func LoadConfig(configPath string) (*Config, error) {
	configFile := configPath + string(os.PathSeparator) + "config.yml"
	data, err := os.ReadFile(configFile)
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
