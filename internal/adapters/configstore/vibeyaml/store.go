package vibeyaml

import (
	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
	"github.com/phamdaiminhquan/vibe-devops/pkg/config"
)

// Store is a `.vibe.yaml` implementation of ports.ConfigStore.
// It intentionally delegates to pkg/config for now to keep behavior stable
// while the architecture is being migrated.
type Store struct{}

func New() *Store { return &Store{} }

func (s *Store) Load(dir string) (*config.Config, error) { return config.Load(dir) }

func (s *Store) Write(dir string, cfg *config.Config) error { return config.Write(dir, cfg) }

func (s *Store) Default() *config.Config { return config.GetDefaultConfig() }

var _ ports.ConfigStore = (*Store)(nil)
