package diagnose

import "context"

// SystemInfo holds collected system information
type SystemInfo struct {
	OS   string
	Arch string

	// Disk info
	DiskUsagePercent float64
	DiskTotal        uint64
	DiskFree         uint64
	DiskPath         string

	// Memory info
	MemoryUsagePercent float64
	MemoryTotal        uint64
	MemoryFree         uint64

	// Docker info
	DockerRunning    bool
	DockerContainers int
	DockerImages     int

	// Network info
	ListeningPorts []PortInfo

	// Services info
	Services []ServiceInfo
}

// PortInfo represents a listening port
type PortInfo struct {
	Port    int
	Process string
	State   string
}

// ServiceInfo represents a running service
type ServiceInfo struct {
	Name   string
	Status string // running, stopped, unknown
	Port   int
}

// Collector interface for collecting system info
type Collector interface {
	Name() string
	Collect(ctx context.Context, info *SystemInfo) error
}

// Rule interface for evaluating system info
type Rule interface {
	Name() string
	Evaluate(info *SystemInfo) (issues []Issue, checks []Check)
}
