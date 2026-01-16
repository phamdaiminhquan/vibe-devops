package file

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/phamdaiminhquan/vibe-devops/internal/ports"
)

// Provider provides file content as context
type Provider struct {
	baseDir string
}

// NewProvider creates a new file context provider
func NewProvider(baseDir string) *Provider {
	return &Provider{baseDir: baseDir}
}

// Description returns the provider's metadata
func (p *Provider) Description() ports.ContextProviderDescription {
	return ports.ContextProviderDescription{
		Name:         "file",
		DisplayTitle: "File Content",
		Description:  "Read file content to provide as context. Use @file path/to/file",
		Type:         ports.ContextTypeFile,
	}
}

// GetContextItems retrieves file content as context
func (p *Provider) GetContextItems(ctx context.Context, query string, extras ports.ContextExtras) ([]ports.ContextItem, error) {
	filePath := strings.TrimSpace(query)
	if filePath == "" {
		return nil, fmt.Errorf("file path is required")
	}

	// Resolve path
	absPath := filePath
	if !filepath.IsAbs(filePath) {
		baseDir := p.baseDir
		if extras.WorkDir != "" {
			baseDir = extras.WorkDir
		}
		absPath = filepath.Join(baseDir, filePath)
	}

	// Check if file exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if info.IsDir() {
		return p.getDirectoryContext(absPath, filePath)
	}

	return p.getFileContext(absPath, filePath, info.Size())
}

func (p *Provider) getFileContext(absPath, displayPath string, size int64) ([]ports.ContextItem, error) {
	// Limit file size to 256KB
	const maxSize = 256 * 1024
	if size > maxSize {
		return nil, fmt.Errorf("file too large (%d bytes), max is %d bytes", size, maxSize)
	}

	f, err := os.Open(absPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var content strings.Builder
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 512*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		content.WriteString(fmt.Sprintf("%4d: %s\n", lineNo, scanner.Text()))
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []ports.ContextItem{
		{
			Name:        filepath.Base(displayPath),
			Description: fmt.Sprintf("Contents of %s (%d lines)", displayPath, lineNo),
			Content:     content.String(),
			URI:         "file://" + absPath,
		},
	}, nil
}

func (p *Provider) getDirectoryContext(absPath, displayPath string) ([]ports.ContextItem, error) {
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("Directory: %s\n\n", displayPath))

	for _, entry := range entries {
		kind := "file"
		if entry.IsDir() {
			kind = "dir "
		}
		info, _ := entry.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		content.WriteString(fmt.Sprintf("[%s] %-40s %10d bytes\n", kind, entry.Name(), size))
	}

	return []ports.ContextItem{
		{
			Name:        filepath.Base(displayPath),
			Description: fmt.Sprintf("Directory listing of %s (%d entries)", displayPath, len(entries)),
			Content:     content.String(),
			URI:         "file://" + absPath,
		},
	}, nil
}

var _ ports.ContextProvider = (*Provider)(nil)
