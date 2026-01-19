package safety

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	BackupDir     = ".vibe/backups"
	MaxBackupDays = 7
)

// DangerLevel indicates how dangerous a command is
type DangerLevel int

const (
	Safe DangerLevel = iota
	Warning
	Dangerous
	Blocked
)

// DangerousPattern represents a pattern that indicates danger
type DangerousPattern struct {
	Pattern     *regexp.Regexp
	Level       DangerLevel
	Description string
	Alternative string
}

// ProtectedPath represents a system path that needs extra protection
type ProtectedPath struct {
	Path        string
	Description string
}

// BackupManifest tracks what was backed up
type BackupManifest struct {
	Timestamp   time.Time         `json:"timestamp"`
	Command     string            `json:"command"`
	BackupPaths map[string]string `json:"backup_paths"` // original -> backup
}

// Dangerous command patterns
var dangerousPatterns = []DangerousPattern{
	// Blocked - never allow
	{regexp.MustCompile(`rm\s+-rf?\s+/\s*$`), Blocked, "Delete entire filesystem", ""},
	{regexp.MustCompile(`rm\s+-rf?\s+/\*`), Blocked, "Delete root contents", ""},
	{regexp.MustCompile(`dd\s+.*of=/dev/[sh]d`), Blocked, "Overwrite disk", ""},
	{regexp.MustCompile(`mkfs\s+/dev/`), Blocked, "Format disk", ""},
	{regexp.MustCompile(`:\(\)\s*\{\s*:\|:\s*&\s*\}`), Blocked, "Fork bomb", ""},

	// Dangerous - require backup
	{regexp.MustCompile(`rm\s+-rf?\s+/(etc|var|usr|bin|sbin|root|home)`), Dangerous, "Delete system directory", "Be more specific about what to delete"},
	{regexp.MustCompile(`rm\s+-rf?\s+/opt/`), Dangerous, "Delete application directory", ""},
	{regexp.MustCompile(`>\s*/(etc|var)/`), Dangerous, "Overwrite system file", "Use tee or proper editor"},
	{regexp.MustCompile(`chmod\s+-R\s+777\s+/`), Dangerous, "Dangerous permissions", "Use appropriate permissions"},
	{regexp.MustCompile(`chown\s+-R\s+.*\s+/(etc|var|usr)`), Dangerous, "Change system ownership", ""},

	// Warning - just warn
	{regexp.MustCompile(`rm\s+-rf?`), Warning, "Recursive delete", ""},
	{regexp.MustCompile(`>\s*[^|]`), Warning, "File overwrite", "Consider using >> for append"},
	{regexp.MustCompile(`kill\s+-9`), Warning, "Force kill process", "Try SIGTERM first"},
	{regexp.MustCompile(`systemctl\s+(stop|disable)`), Warning, "Stop/disable service", ""},
	{regexp.MustCompile(`docker\s+(rm|rmi|system\s+prune)`), Warning, "Remove Docker resources", ""},
}

// Protected paths
var protectedPaths = []ProtectedPath{
	{"/etc", "System configuration"},
	{"/var", "Variable data"},
	{"/usr", "User programs"},
	{"/bin", "Essential binaries"},
	{"/sbin", "System binaries"},
	{"/root", "Root home"},
	{"/home", "User homes"},
	{"/opt", "Optional applications"},
}

// CheckResult contains the result of command safety check
type CheckResult struct {
	Level         DangerLevel
	Description   string
	Alternative   string
	AffectedPaths []string
	SuggestBackup bool
}

// CheckCommand analyzes a command for safety
func CheckCommand(cmd string) CheckResult {
	result := CheckResult{Level: Safe}

	// Check against dangerous patterns
	for _, dp := range dangerousPatterns {
		if dp.Pattern.MatchString(cmd) {
			if dp.Level > result.Level {
				result.Level = dp.Level
				result.Description = dp.Description
				result.Alternative = dp.Alternative
			}
		}
	}

	// Check for protected paths
	for _, pp := range protectedPaths {
		if strings.Contains(cmd, pp.Path) {
			result.AffectedPaths = append(result.AffectedPaths, pp.Path+" ("+pp.Description+")")
			if result.Level < Warning {
				result.Level = Warning
			}
		}
	}

	// Suggest backup for dangerous commands
	result.SuggestBackup = result.Level >= Warning && len(result.AffectedPaths) > 0

	return result
}

// ExtractPaths extracts file/directory paths from a command
func ExtractPaths(cmd string) []string {
	var paths []string

	// Common patterns for paths in commands
	pathPatterns := []*regexp.Regexp{
		regexp.MustCompile(`rm\s+-rf?\s+(.+)`),
		regexp.MustCompile(`cp\s+.+\s+(.+)`),
		regexp.MustCompile(`mv\s+.+\s+(.+)`),
		regexp.MustCompile(`>\s*(.+)`),
	}

	for _, pattern := range pathPatterns {
		matches := pattern.FindStringSubmatch(cmd)
		if len(matches) > 1 {
			for _, path := range strings.Fields(matches[1]) {
				if strings.HasPrefix(path, "/") {
					paths = append(paths, path)
				}
			}
		}
	}

	return paths
}

// CreateBackup creates a backup of the specified paths
func CreateBackup(cmd string, paths []string) (string, error) {
	if len(paths) == 0 {
		return "", nil
	}

	// Create backup directory
	homeDir, _ := os.UserHomeDir()
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(homeDir, BackupDir, timestamp)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	manifest := BackupManifest{
		Timestamp:   time.Now(),
		Command:     cmd,
		BackupPaths: make(map[string]string),
	}

	// Backup each path
	for _, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue // Skip non-existent paths
		}

		// Create backup name (replace / with _)
		backupName := strings.ReplaceAll(strings.TrimPrefix(path, "/"), "/", "_")
		backupDest := filepath.Join(backupPath, backupName)

		// Copy using cp -r
		cpCmd := exec.Command("cp", "-r", path, backupDest)
		if err := cpCmd.Run(); err != nil {
			fmt.Printf("⚠️ Failed to backup %s: %v\n", path, err)
			continue
		}

		manifest.BackupPaths[path] = backupDest
	}

	// Write manifest
	manifestPath := filepath.Join(backupPath, "manifest.json")
	manifestData, _ := json.MarshalIndent(manifest, "", "  ")
	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return "", fmt.Errorf("failed to write manifest: %w", err)
	}

	return backupPath, nil
}

// GetRecentBackups returns recent backups
func GetRecentBackups(limit int) ([]BackupManifest, error) {
	homeDir, _ := os.UserHomeDir()
	backupRoot := filepath.Join(homeDir, BackupDir)

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return nil, nil // No backups yet
	}

	var backups []BackupManifest

	// Read in reverse order (newest first)
	for i := len(entries) - 1; i >= 0 && len(backups) < limit; i-- {
		entry := entries[i]
		if !entry.IsDir() {
			continue
		}

		manifestPath := filepath.Join(backupRoot, entry.Name(), "manifest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			continue
		}

		var manifest BackupManifest
		if err := json.Unmarshal(data, &manifest); err != nil {
			continue
		}

		backups = append(backups, manifest)
	}

	return backups, nil
}

// RestoreBackup restores files from a backup
func RestoreBackup(backupPath string) error {
	manifestPath := filepath.Join(backupPath, "manifest.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest BackupManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	for original, backup := range manifest.BackupPaths {
		fmt.Printf("Restoring %s...\n", original)
		cpCmd := exec.Command("cp", "-r", backup, original)
		if err := cpCmd.Run(); err != nil {
			fmt.Printf("⚠️ Failed to restore %s: %v\n", original, err)
		}
	}

	return nil
}

// CleanupOldBackups removes backups older than MaxBackupDays
func CleanupOldBackups() {
	homeDir, _ := os.UserHomeDir()
	backupRoot := filepath.Join(homeDir, BackupDir)

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		return
	}

	cutoff := time.Now().AddDate(0, 0, -MaxBackupDays)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Parse timestamp from directory name
		t, err := time.Parse("2006-01-02_15-04-05", entry.Name())
		if err != nil {
			continue
		}

		if t.Before(cutoff) {
			path := filepath.Join(backupRoot, entry.Name())
			os.RemoveAll(path)
			fmt.Printf("Cleaned up old backup: %s\n", entry.Name())
		}
	}
}

// PromptBackupChoice asks user what to do with dangerous command
func PromptBackupChoice(cmd string, result CheckResult) (action string, doBackup bool) {
	fmt.Println()

	if result.Level == Blocked {
		fmt.Println("⚠️  BLOCKED: This command is too dangerous to execute:")
		fmt.Printf("   %s\n", cmd)
		fmt.Printf("   Reason: %s\n", result.Description)
		return "cancel", false
	}

	if result.Level == Dangerous {
		fmt.Println("⚠️  DANGEROUS COMMAND:")
	} else {
		fmt.Println("⚠️  WARNING:")
	}

	fmt.Printf("   %s\n", cmd)
	fmt.Printf("   Reason: %s\n", result.Description)

	if len(result.AffectedPaths) > 0 {
		fmt.Println("\n   Affected system paths:")
		for _, p := range result.AffectedPaths {
			fmt.Printf("   • %s\n", p)
		}
	}

	if result.Alternative != "" {
		fmt.Printf("\n   Suggestion: %s\n", result.Alternative)
	}

	fmt.Println("\n   Options:")
	fmt.Println("   [b] Create backup first, then run")
	fmt.Println("   [r] Run without backup")
	fmt.Println("   [c] Cancel")
	fmt.Print("\n   Choice: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	switch input {
	case "b":
		return "run", true
	case "r":
		return "run", false
	default:
		return "cancel", false
	}
}
