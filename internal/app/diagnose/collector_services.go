package diagnose

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// ServicesCollector checks common services
type ServicesCollector struct {
	// Known services with their ports for enhanced info
	knownPorts map[string]int
}

func NewServicesCollector() *ServicesCollector {
	return &ServicesCollector{
		knownPorts: map[string]int{
			"nginx":        80,
			"apache2":      80,
			"httpd":        80,
			"mysql":        3306,
			"mysqld":       3306,
			"mariadb":      3306,
			"redis":        6379,
			"redis-server": 6379,
			"postgresql":   5432,
			"postgres":     5432,
			"mongodb":      27017,
			"mongod":       27017,
			"sshd":         22,
			"docker":       0,
			"containerd":   0,
		},
	}
}

func (c *ServicesCollector) Name() string {
	return "services"
}

func (c *ServicesCollector) Collect(ctx context.Context, info *SystemInfo) error {
	if runtime.GOOS == "windows" {
		c.collectWindows(info)
	} else {
		c.collectUnix(info)
	}
	return nil
}

func (c *ServicesCollector) collectUnix(info *SystemInfo) {
	// Auto-detect ALL running services via systemctl
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--state=running", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to checking known services if systemctl not available or fails
		c.collectUnixFallback(info)
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			// Service name format: nginx.service
			svcName := strings.TrimSuffix(fields[0], ".service")
			port := c.knownPorts[svcName]
			info.Services = append(info.Services, ServiceInfo{
				Name:   svcName,
				Status: "running",
				Port:   port,
			})
		}
	}
}

// collectUnixFallback checks a predefined list of services using older methods (systemctl is-active, pgrep)
func (c *ServicesCollector) collectUnixFallback(info *SystemInfo) {
	for svc := range c.knownPorts { // Iterate over known services
		status := c.checkServiceUnixFallback(svc)
		if status != "" {
			info.Services = append(info.Services, ServiceInfo{
				Name:   svc,
				Status: status,
				Port:   c.knownPorts[svc],
			})
		}
	}
}

// checkServiceUnixFallback checks a single service using systemctl is-active or pgrep
func (c *ServicesCollector) checkServiceUnixFallback(service string) string {
	// Try systemctl first
	cmd := exec.Command("systemctl", "is-active", service)
	output, err := cmd.Output()
	if err == nil {
		status := strings.TrimSpace(string(output))
		if status == "active" {
			return "running"
		}
		return "stopped"
	}

	// Try pgrep as fallback
	cmd = exec.Command("pgrep", "-x", service)
	if err := cmd.Run(); err == nil {
		return "running"
	}

	return "" // Not installed or not running
}

func (c *ServicesCollector) collectWindows(info *SystemInfo) {
	// For Windows, iterate over known services and check their status
	for svc := range c.knownPorts {
		status := c.checkServiceWindows(svc)
		if status != "" {
			info.Services = append(info.Services, ServiceInfo{
				Name:   svc,
				Status: status,
				Port:   c.getServicePort(svc),
			})
		}
	}
}

func (c *ServicesCollector) checkServiceWindows(service string) string {
	// Check if process is running
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-Process -Name '%s' -ErrorAction SilentlyContinue | Select-Object -First 1", service))
	if err := cmd.Run(); err == nil {
		return "running"
	}
	return ""
}

func (c *ServicesCollector) getServicePort(service string) int {
	ports := map[string]int{
		"nginx":      80,
		"mysql":      3306,
		"redis":      6379,
		"postgresql": 5432,
		"mongodb":    27017,
	}
	if port, ok := ports[service]; ok {
		return port
	}
	return 0
}

// ServiceRule checks if critical services are running
type ServiceRule struct{}

func NewServiceRule() *ServiceRule {
	return &ServiceRule{}
}

func (r *ServiceRule) Name() string {
	return "services"
}

func (r *ServiceRule) Evaluate(info *SystemInfo) (issues []Issue, checks []Check) {
	for _, svc := range info.Services {
		if svc.Status == "running" {
			portInfo := ""
			if svc.Port > 0 {
				portInfo = fmt.Sprintf(" (port %d)", svc.Port)
			}
			checks = append(checks, Check{
				Category:    "service",
				Description: svc.Name,
				Value:       "running" + portInfo,
			})
		} else if svc.Status == "stopped" {
			issues = append(issues, Issue{
				Category:    "service",
				Description: fmt.Sprintf("%s not running", svc.Name),
				Severity:    "warning",
				FixCommand:  fmt.Sprintf("sudo systemctl start %s", svc.Name),
			})
		}
	}
	return
}
