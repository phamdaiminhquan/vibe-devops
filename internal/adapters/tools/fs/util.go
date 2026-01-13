package fs

import (
	"fmt"
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

	baseWithSep := baseAbs
	if !strings.HasSuffix(baseWithSep, string(filepath.Separator)) {
		baseWithSep += string(filepath.Separator)
	}

	// Allow exactly baseDir as well.
	if abs != baseAbs && !strings.HasPrefix(abs, baseWithSep) {
		return "", fmt.Errorf("path escapes workspace root")
	}

	return abs, nil
}
