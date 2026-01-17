package diagnose

import (
	"context"
	"fmt"
	"runtime"
)

// DiagnoseResult represents the result of a diagnose operation
type DiagnoseResult struct {
	Warnings []Issue
	Errors   []Issue
	OK       []Check
	Summary  string
}

// Issue represents a problem found during diagnosis
type Issue struct {
	Category    string // disk, memory, network, service
	Description string
	Value       string // current value
	Threshold   string // threshold value
	Severity    string // warning, error
	FixCommand  string // suggested fix command
}

// Check represents a passed check
type Check struct {
	Category    string
	Description string
	Value       string
}

// Service is the diagnose service
type Service struct {
	collectors []Collector
	rules      []Rule
}

// NewService creates a new diagnose service
func NewService() *Service {
	return &Service{
		collectors: []Collector{
			NewSystemCollector(),
			NewDockerCollector(),
			NewNetworkCollector(),
			NewServicesCollector(),
		},
		rules: []Rule{
			NewDiskRule(85),   // warn at 85%
			NewMemoryRule(80), // warn at 80%
			NewDockerRule(),
			NewPortRule(),
			NewServiceRule(),
		},
	}
}

// Run executes the diagnosis
func (s *Service) Run(ctx context.Context) (*DiagnoseResult, error) {
	result := &DiagnoseResult{
		Warnings: []Issue{},
		Errors:   []Issue{},
		OK:       []Check{},
	}

	// Collect system info
	info := &SystemInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}

	for _, collector := range s.collectors {
		if err := collector.Collect(ctx, info); err != nil {
			// Log error but continue
			continue
		}
	}

	// Run rules
	for _, rule := range s.rules {
		issues, checks := rule.Evaluate(info)
		for _, issue := range issues {
			if issue.Severity == "error" {
				result.Errors = append(result.Errors, issue)
			} else {
				result.Warnings = append(result.Warnings, issue)
			}
		}
		result.OK = append(result.OK, checks...)
	}

	// Generate summary
	result.Summary = fmt.Sprintf(
		"Found %d warnings, %d errors, %d OK",
		len(result.Warnings), len(result.Errors), len(result.OK),
	)

	return result, nil
}
