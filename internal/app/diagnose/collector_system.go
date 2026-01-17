package diagnose

import (
	"context"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// SystemCollector collects OS, disk, and memory info
type SystemCollector struct{}

func NewSystemCollector() *SystemCollector {
	return &SystemCollector{}
}

func (c *SystemCollector) Name() string {
	return "system"
}

func (c *SystemCollector) Collect(ctx context.Context, info *SystemInfo) error {
	// Collect disk usage
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		c.collectDiskUnix(info)
		c.collectMemoryUnix(info)
	} else if runtime.GOOS == "windows" {
		c.collectDiskWindows(info)
		c.collectMemoryWindows(info)
	}
	return nil
}

func (c *SystemCollector) collectDiskUnix(info *SystemInfo) {
	// df -h / | tail -1 | awk '{print $5}'
	cmd := exec.Command("sh", "-c", "df -h / | tail -1 | awk '{print $5}'")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// Parse percentage (e.g., "87%")
	pctStr := strings.TrimSpace(string(output))
	pctStr = strings.TrimSuffix(pctStr, "%")
	if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
		info.DiskUsagePercent = pct
		info.DiskPath = "/"
	}
}

func (c *SystemCollector) collectMemoryUnix(info *SystemInfo) {
	// free -m | grep Mem | awk '{print $3/$2 * 100}'
	cmd := exec.Command("sh", "-c", "free -m | grep Mem | awk '{print $3/$2 * 100}'")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	pctStr := strings.TrimSpace(string(output))
	if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
		info.MemoryUsagePercent = pct
	}
}

func (c *SystemCollector) collectDiskWindows(info *SystemInfo) {
	// PowerShell: Get-WmiObject Win32_LogicalDisk -Filter "DeviceID='C:'" | Select-Object FreeSpace, Size
	cmd := exec.Command("powershell", "-Command",
		"$disk = Get-WmiObject Win32_LogicalDisk -Filter \"DeviceID='C:'\"; "+
			"[math]::Round((1 - $disk.FreeSpace / $disk.Size) * 100, 1)")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	pctStr := strings.TrimSpace(string(output))
	if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
		info.DiskUsagePercent = pct
		info.DiskPath = "C:\\"
	}
}

func (c *SystemCollector) collectMemoryWindows(info *SystemInfo) {
	// PowerShell: Get memory usage percentage
	cmd := exec.Command("powershell", "-Command",
		"$os = Get-WmiObject Win32_OperatingSystem; "+
			"[math]::Round(($os.TotalVisibleMemorySize - $os.FreePhysicalMemory) / $os.TotalVisibleMemorySize * 100, 1)")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	pctStr := strings.TrimSpace(string(output))
	if pct, err := strconv.ParseFloat(pctStr, 64); err == nil {
		info.MemoryUsagePercent = pct
	}
}
