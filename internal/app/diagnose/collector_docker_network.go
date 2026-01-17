package diagnose

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// DockerCollector collects Docker info
type DockerCollector struct{}

func NewDockerCollector() *DockerCollector {
	return &DockerCollector{}
}

func (c *DockerCollector) Name() string {
	return "docker"
}

func (c *DockerCollector) Collect(ctx context.Context, info *SystemInfo) error {
	// Check if Docker is running
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("docker", "info")
	} else {
		cmd = exec.Command("docker", "info")
	}

	if err := cmd.Run(); err != nil {
		info.DockerRunning = false
		return nil
	}
	info.DockerRunning = true

	// Count running containers
	countCmd := exec.Command("docker", "ps", "-q")
	output, err := countCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			info.DockerContainers = 0
		} else {
			info.DockerContainers = len(lines)
		}
	}

	// Count images
	imgCmd := exec.Command("docker", "images", "-q")
	imgOutput, err := imgCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(imgOutput)), "\n")
		if len(lines) == 1 && lines[0] == "" {
			info.DockerImages = 0
		} else {
			info.DockerImages = len(lines)
		}
	}

	return nil
}

// NetworkCollector collects network/port info
type NetworkCollector struct{}

func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{}
}

func (c *NetworkCollector) Name() string {
	return "network"
}

func (c *NetworkCollector) Collect(ctx context.Context, info *SystemInfo) error {
	if runtime.GOOS == "linux" {
		c.collectLinux(info)
	} else if runtime.GOOS == "darwin" {
		c.collectDarwin(info)
	} else if runtime.GOOS == "windows" {
		c.collectWindows(info)
	}
	return nil
}

func (c *NetworkCollector) collectLinux(info *SystemInfo) {
	// ss -tlnp | grep LISTEN
	cmd := exec.Command("sh", "-c", "ss -tlnp 2>/dev/null | grep LISTEN | head -20")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	c.parseSSOutput(info, string(output))
}

func (c *NetworkCollector) collectDarwin(info *SystemInfo) {
	// lsof -i -P -n | grep LISTEN
	cmd := exec.Command("sh", "-c", "lsof -i -P -n 2>/dev/null | grep LISTEN | head -20")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 9 {
			// Parse port from address field (e.g., *:22)
			addrField := fields[8]
			if idx := strings.LastIndex(addrField, ":"); idx != -1 {
				portStr := addrField[idx+1:]
				if port, err := strconv.Atoi(portStr); err == nil {
					info.ListeningPorts = append(info.ListeningPorts, PortInfo{
						Port:    port,
						Process: fields[0],
						State:   "LISTEN",
					})
				}
			}
		}
	}
}

func (c *NetworkCollector) collectWindows(info *SystemInfo) {
	// netstat -an | findstr LISTENING
	cmd := exec.Command("powershell", "-Command",
		"netstat -an | Select-String 'LISTENING' | Select-Object -First 20")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			// Parse port from local address (e.g., 0.0.0.0:22)
			addrField := fields[1]
			if idx := strings.LastIndex(addrField, ":"); idx != -1 {
				portStr := addrField[idx+1:]
				if port, err := strconv.Atoi(portStr); err == nil {
					info.ListeningPorts = append(info.ListeningPorts, PortInfo{
						Port:    port,
						Process: "unknown",
						State:   "LISTEN",
					})
				}
			}
		}
	}
}

func (c *NetworkCollector) parseSSOutput(info *SystemInfo, output string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			// Parse port from Local Address:Port (e.g., 0.0.0.0:22)
			addrField := fields[3]
			if idx := strings.LastIndex(addrField, ":"); idx != -1 {
				portStr := addrField[idx+1:]
				if port, err := strconv.Atoi(portStr); err == nil {
					process := ""
					if len(fields) >= 6 {
						process = fields[5]
					}
					info.ListeningPorts = append(info.ListeningPorts, PortInfo{
						Port:    port,
						Process: process,
						State:   "LISTEN",
					})
				}
			}
		}
	}
}
