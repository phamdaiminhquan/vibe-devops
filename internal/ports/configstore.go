package ports

import "github.com/phamdaiminhquan/vibe-devops/pkg/config"

// ConfigStore is the outbound port for reading/writing the on-disk configuration.
// Today this is backed by `.vibe.yaml`, but this abstraction allows future
// config backends (XDG paths, env vars, remote stores) without changing use-cases.
type ConfigStore interface {
	Load(dir string) (*config.Config, error)
	Write(dir string, cfg *config.Config) error
	Default() *config.Config
}
