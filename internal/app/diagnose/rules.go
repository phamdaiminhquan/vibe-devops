package diagnose

import "fmt"

// DiskRule checks disk usage
type DiskRule struct {
	threshold float64
}

func NewDiskRule(threshold float64) *DiskRule {
	return &DiskRule{threshold: threshold}
}

func (r *DiskRule) Name() string {
	return "disk"
}

func (r *DiskRule) Evaluate(info *SystemInfo) (issues []Issue, checks []Check) {
	if info.DiskUsagePercent == 0 {
		return // No data
	}

	if info.DiskUsagePercent >= r.threshold {
		issues = append(issues, Issue{
			Category:    "disk",
			Description: fmt.Sprintf("Disk %s: %.1f%%", info.DiskPath, info.DiskUsagePercent),
			Value:       fmt.Sprintf("%.1f%%", info.DiskUsagePercent),
			Threshold:   fmt.Sprintf("%.0f%%", r.threshold),
			Severity:    r.getSeverity(info.DiskUsagePercent),
			FixCommand:  r.getFixCommand(info.DiskUsagePercent),
		})
	} else {
		checks = append(checks, Check{
			Category:    "disk",
			Description: fmt.Sprintf("Disk %s", info.DiskPath),
			Value:       fmt.Sprintf("%.1f%%", info.DiskUsagePercent),
		})
	}
	return
}

func (r *DiskRule) getSeverity(pct float64) string {
	if pct >= 95 {
		return "error"
	}
	return "warning"
}

func (r *DiskRule) getFixCommand(pct float64) string {
	if pct >= 90 {
		return "docker system prune -af && sudo apt-get clean"
	}
	return "docker system prune -af"
}

// MemoryRule checks memory usage
type MemoryRule struct {
	threshold float64
}

func NewMemoryRule(threshold float64) *MemoryRule {
	return &MemoryRule{threshold: threshold}
}

func (r *MemoryRule) Name() string {
	return "memory"
}

func (r *MemoryRule) Evaluate(info *SystemInfo) (issues []Issue, checks []Check) {
	if info.MemoryUsagePercent == 0 {
		return // No data
	}

	if info.MemoryUsagePercent >= r.threshold {
		issues = append(issues, Issue{
			Category:    "memory",
			Description: fmt.Sprintf("RAM: %.1f%%", info.MemoryUsagePercent),
			Value:       fmt.Sprintf("%.1f%%", info.MemoryUsagePercent),
			Threshold:   fmt.Sprintf("%.0f%%", r.threshold),
			Severity:    r.getSeverity(info.MemoryUsagePercent),
			FixCommand:  "top -o %MEM | head -15",
		})
	} else {
		checks = append(checks, Check{
			Category:    "memory",
			Description: "RAM",
			Value:       fmt.Sprintf("%.1f%%", info.MemoryUsagePercent),
		})
	}
	return
}

func (r *MemoryRule) getSeverity(pct float64) string {
	if pct >= 95 {
		return "error"
	}
	return "warning"
}

// DockerRule checks Docker status
type DockerRule struct{}

func NewDockerRule() *DockerRule {
	return &DockerRule{}
}

func (r *DockerRule) Name() string {
	return "docker"
}

func (r *DockerRule) Evaluate(info *SystemInfo) (issues []Issue, checks []Check) {
	if info.DockerRunning {
		checks = append(checks, Check{
			Category:    "docker",
			Description: "Docker",
			Value:       fmt.Sprintf("running (%d containers)", info.DockerContainers),
		})
	} else {
		issues = append(issues, Issue{
			Category:    "docker",
			Description: "Docker not running",
			Severity:    "warning",
			FixCommand:  "sudo systemctl start docker",
		})
	}
	return
}

// PortRule checks for common ports
type PortRule struct {
	requiredPorts map[int]string // port -> service name
}

func NewPortRule() *PortRule {
	return &PortRule{
		requiredPorts: map[int]string{
			22: "SSH",
		},
	}
}

func (r *PortRule) Name() string {
	return "ports"
}

func (r *PortRule) Evaluate(info *SystemInfo) (issues []Issue, checks []Check) {
	// Check if SSH is listening
	openPorts := make(map[int]bool)
	for _, p := range info.ListeningPorts {
		openPorts[p.Port] = true
	}

	for port, service := range r.requiredPorts {
		if openPorts[port] {
			checks = append(checks, Check{
				Category:    "network",
				Description: service,
				Value:       fmt.Sprintf("port %d open", port),
			})
		}
	}

	return
}
