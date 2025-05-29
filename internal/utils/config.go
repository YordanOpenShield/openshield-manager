package utils

import (
	"log"
	"openshield-manager/internal/config"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func CreateConfig(configDir string, opts config.Config) {
	configFile := filepath.Join(configDir, "config.yml")

	// Check if the config directory exists, if not, create it
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.MkdirAll(configDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create config directory: %v", err)
		}
	}

	// Check if the config file exists, if not, create it with default values
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := config.GenerateConfig(opts)
		yamlBytes, err := yaml.Marshal(defaultConfig)
		if err != nil {
			log.Fatalf("Failed to marshal default config: %v", err)
		}
		err = os.WriteFile(configFile, yamlBytes, 0644)
		if err != nil {
			log.Fatalf("Failed to create default config.yml: %v", err)
		}
		log.Printf("Created default config at %s", configFile)
	}
}

func CreateScriptsDir(scriptsDir string) {
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		err := os.MkdirAll(scriptsDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create scripts directory: %v", err)
		}
		log.Printf("Created scripts directory at %s", scriptsDir)
	}
}

func CreateCertsDir(certsDir string) {
	if _, err := os.Stat(certsDir); os.IsNotExist(err) {
		err := os.MkdirAll(certsDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create certs directory: %v", err)
		}
		log.Printf("Created scripts directory at %s", certsDir)
	}
}
