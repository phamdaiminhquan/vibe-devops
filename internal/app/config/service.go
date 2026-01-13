package config

import (
	"fmt"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	cfgpkg "github.com/phamdaiminhquan/vibe-devops/pkg/config"
)

type Service struct {
	store ports.ConfigStore
}

func NewService(store ports.ConfigStore) *Service {
	return &Service{store: store}
}

func (s *Service) Load(dir string) (*cfgpkg.Config, error) {
	return s.store.Load(dir)
}

func (s *Service) SetProvider(dir, providerName string) (*cfgpkg.Config, error) {
	providerName = strings.ToLower(strings.TrimSpace(providerName))
	if providerName == "" {
		return nil, fmt.Errorf("provider name is empty")
	}

	cfg, err := s.store.Load(dir)
	if err != nil {
		return nil, err
	}

	cfg.AI.Provider = providerName
	if err := s.store.Write(dir, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) SetGeminiAPIKey(dir, apiKey string) (*cfgpkg.Config, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("api key is empty")
	}

	cfg, err := s.store.Load(dir)
	if err != nil {
		return nil, err
	}

	cfg.AI.Gemini.APIKey = apiKey
	if err := s.store.Write(dir, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (s *Service) SetGeminiModel(dir, model string) (*cfgpkg.Config, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return nil, fmt.Errorf("model is empty")
	}

	cfg, err := s.store.Load(dir)
	if err != nil {
		return nil, err
	}

	cfg.AI.Gemini.Model = model
	if err := s.store.Write(dir, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
