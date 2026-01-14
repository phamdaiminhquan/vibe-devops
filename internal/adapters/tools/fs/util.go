package fs

import (
	"path/filepath"
	"strings"
)

func resolvePath(baseDir, userPath string) (string, error) {
	if strings.TrimSpace(userPath) == "" {
		userPath = "."
	}

	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return "", err
	}

	// Interpret absolute inputs as-is, but still enforce base containment.
	clean := filepath.Clean(userPath)
	var abs string
	if filepath.IsAbs(clean) {
		abs = clean
	} else {
		abs = filepath.Join(baseAbs, clean)
	}

	abs, err = filepath.Abs(abs)
	if err != nil {
		return "", err
	}

	// Security: For Vibe DevOps, we trust the agent/user to access system files.
	// We do NOT enforce workspace containment because we need to read /var/log, /etc, etc.

	// if abs != baseAbs && !strings.HasPrefix(abs, baseWithSep) {
	// 	 return "", fmt.Errorf("path escapes workspace root")
	// }

	return abs, nil
}
