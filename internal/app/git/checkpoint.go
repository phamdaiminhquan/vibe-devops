package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	// CheckpointPrefix is used to identify vibe checkpoints in git log
	CheckpointPrefix = "[vibe-checkpoint]"
	// MaxCheckpoints is the maximum number of checkpoints to keep
	MaxCheckpoints = 10
)

// CheckpointInfo represents a vibe checkpoint
type CheckpointInfo struct {
	Hash      string
	Message   string
	Timestamp time.Time
	Relative  string // e.g., "5 minutes ago"
}

// HasUncommittedChanges checks if there are uncommitted changes in the working directory
func HasUncommittedChanges(workDir string) bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}

// IsGitRepo checks if the directory is a git repository
func IsGitRepo(workDir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

// CreateCheckpoint creates a checkpoint commit with uncommitted changes
// Returns the commit hash of the checkpoint
func CreateCheckpoint(workDir string) (string, error) {
	// Stage all changes
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = workDir
	if err := addCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create checkpoint commit
	message := fmt.Sprintf("%s Before AI session at %s", CheckpointPrefix, time.Now().Format("15:04:05"))
	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = workDir
	if err := commitCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create checkpoint: %w", err)
	}

	// Get the commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = workDir
	output, err := hashCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	return strings.TrimSpace(string(output))[:7], nil
}

// GetRecentCheckpoints returns recent vibe checkpoints from git log
func GetRecentCheckpoints(workDir string, limit int) ([]CheckpointInfo, error) {
	// Get recent commits with vibe-checkpoint prefix
	cmd := exec.Command("git", "log", "--oneline", "--format=%h|%s|%ar", fmt.Sprintf("-n%d", limit*2))
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git log: %w", err)
	}

	var checkpoints []CheckpointInfo
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}

		hash := parts[0]
		message := parts[1]
		relative := parts[2]

		// Only include vibe checkpoints
		if strings.Contains(message, CheckpointPrefix) {
			checkpoints = append(checkpoints, CheckpointInfo{
				Hash:     hash,
				Message:  message,
				Relative: relative,
			})
		}

		if len(checkpoints) >= limit {
			break
		}
	}

	return checkpoints, nil
}

// RestoreCheckpoint restores the working directory to a checkpoint
// Uses git reset --hard to restore, then creates a new commit to preserve history
func RestoreCheckpoint(workDir, commitHash string) error {
	// First, save current state
	if HasUncommittedChanges(workDir) {
		// Stash current changes
		stashCmd := exec.Command("git", "stash", "push", "-m", "[vibe] Stashed before undo")
		stashCmd.Dir = workDir
		_ = stashCmd.Run() // Ignore if stash fails (no changes)
	}

	// Reset to the checkpoint
	resetCmd := exec.Command("git", "reset", "--hard", commitHash)
	resetCmd.Dir = workDir
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to checkpoint: %w", err)
	}

	return nil
}

// UndoLastCheckpoint undoes the last vibe checkpoint (restores to before checkpoint)
func UndoLastCheckpoint(workDir string) error {
	checkpoints, err := GetRecentCheckpoints(workDir, 1)
	if err != nil {
		return err
	}

	if len(checkpoints) == 0 {
		return fmt.Errorf("no vibe checkpoints found")
	}

	// Get the parent of the checkpoint
	parentCmd := exec.Command("git", "rev-parse", checkpoints[0].Hash+"^")
	parentCmd.Dir = workDir
	output, err := parentCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get parent commit: %w", err)
	}

	parentHash := strings.TrimSpace(string(output))[:7]
	return RestoreCheckpoint(workDir, parentHash)
}

// CountUncommittedFiles returns the number of uncommitted files
func CountUncommittedFiles(workDir string) int {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = workDir
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return 0
	}
	return len(lines)
}
