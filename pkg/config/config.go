package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = ".vibe.yaml"
)

// Config holds the application's configuration.
type Config struct {
	AI struct {
		Provider string `yaml:"provider"`
		Gemini   struct {
			APIKey string `yaml:"apiKey"`
			Model  string `yaml:"model"`
		} `yaml:"gemini"`
	} `yaml:"ai"`
}

// Load loads the configuration from the .vibe.yaml file in the specified directory.
func Load(dir string) (*Config, error) {
	configFile := filepath.Join(dir, ConfigFileName)
	
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

// GetDefaultConfig returns a default configuration object.
func GetDefaultConfig() *Config {
	var cfg Config
	cfg.AI.Provider = "gemini"
	cfg.AI.Gemini.APIKey = "YOUR_API_KEY_HERE"
	cfg.AI.Gemini.Model = "gemini-pro"
	return &cfg
}
