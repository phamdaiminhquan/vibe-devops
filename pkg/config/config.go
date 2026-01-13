package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = ".vibe.yaml"
	DefaultAPIKeyPlaceholder = "YOUR_GEMINI_API_KEY_HERE"
)

type GeminiConfig struct {
	APIKey string `yaml:"apiKey"`
	Model  string `yaml:"model"`
}

type AIConfig struct {
	Provider string       `yaml:"provider"`
	Gemini   GeminiConfig `yaml:"gemini"`
}

// Config holds the application's configuration.
type Config struct {
	AI AIConfig `yaml:"ai"`
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

// Write writes the configuration to the .vibe.yaml file in the specified directory.
func Write(dir string, cfg *Config) error {
	configFile := filepath.Join(dir, ConfigFileName)

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configFile, data, 0644)
}

// GetDefaultConfig returns a default configuration object.
func GetDefaultConfig() *Config {
	var cfg Config
	cfg.AI.Provider = "gemini"
	cfg.AI.Gemini.APIKey = DefaultAPIKeyPlaceholder
	cfg.AI.Gemini.Model = "gemini-pro"
	return &cfg
}
