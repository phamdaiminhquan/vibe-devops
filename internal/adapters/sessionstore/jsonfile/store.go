package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

type Store struct {
	baseDir string
}

func New(baseDir string) *Store {
	return &Store{baseDir: baseDir}
}

func (s *Store) Load(sessionName string) (*ports.SessionState, error) {
	path, err := s.sessionPath(sessionName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &ports.SessionState{Version: 1}, nil
		}
		return nil, err
	}

	var st ports.SessionState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, err
	}
	if st.Version == 0 {
		st.Version = 1
	}
	return &st, nil
}

func (s *Store) Save(sessionName string, state *ports.SessionState) error {
	if state == nil {
		return fmt.Errorf("nil session state")
	}
	if state.Version == 0 {
		state.Version = 1
	}

	path, err := s.sessionPath(sessionName)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

var safeNameRE = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func (s *Store) sessionPath(sessionName string) (string, error) {
	name := strings.TrimSpace(sessionName)
	if name == "" {
		name = "default"
	}
	name = safeNameRE.ReplaceAllString(name, "_")
	if len(name) > 80 {
		name = name[:80]
	}

	base := s.baseDir
	if strings.TrimSpace(base) == "" {
		base = "."
	}
	return filepath.Join(base, "sessions", name+".json"), nil
}

var _ ports.SessionStore = (*Store)(nil)
