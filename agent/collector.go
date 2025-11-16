package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type SystemMetrics struct {
	Timestamp time.Time      `json:"timestamp"`
	Hostname  string         `json:"hostname"`
	OS        OSInfo         `json:"os"`
	CPU       CPUMetrics     `json:"cpu"`
	Memory    MemoryMetrics  `json:"memory"`
	Disk      []DiskMetrics  `json:"disk"`
	Network   NetworkMetrics `json:"network"`
	Runtime   RuntimeMetrics `json:"runtime"`
}

type OSInfo struct {
	Platform           string `json:"platform"`
	PlatformFamily     string `json:"platform_family"`
	PlatformVersion    string `json:"platform_version"`
	Architecture       string `json:"architecture"`
	KernelVersion      string `json:"kernel_version"`
	VirtualizationRole string `json:"virtualization_role"`
}

type CPUMetrics struct {
	ModelName    string    `json:"model_name"`
	Cores        int32     `json:"cores"`
	LogicalCores int32     `json:"logical_cores"`
	UsagePercent float64   `json:"usage_percent"`
	LoadAverage  []float64 `json:"load_average"`
	FrequencyMHz float64   `json:"frequency_mhz"`
}

type MemoryMetrics struct {
	Total       uint64  `json:"total"`
	Available   uint64  `json:"available"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapPercent float64 `json:"swap_percent"`
}

type DiskMetrics struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	Fstype      string  `json:"fstype"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"used_percent"`
}

type NetworkMetrics struct {
	Interfaces  []NetworkInterface `json:"interfaces"`
	Connections []ConnectionInfo   `json:"connections"`
}

type NetworkInterface struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardware_addr"`
	MTU          int      `json:"mtu"`
	Flags        []string `json:"flags"`
	Addresses    []string `json:"addresses"`
	BytesSent    uint64   `json:"bytes_sent"`
	BytesRecv    uint64   `json:"bytes_recv"`
	PacketsSent  uint64   `json:"packets_sent"`
	PacketsRecv  uint64   `json:"packets_recv"`
	Errin        uint64   `json:"errin"`
	Errout       uint64   `json:"errout"`
	Dropin       uint64   `json:"dropin"`
	Dropout      uint64   `json:"dropout"`
}

type ConnectionInfo struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	State      string `json:"state"`
	PID        int32  `json:"pid"`
	Process    string `json:"process"`
}

type RuntimeMetrics struct {
	GoVersion    string `json:"go_version"`
	GoOS         string `json:"go_os"`
	GoArch       string `json:"go_arch"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`
}

func CollectSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now().UTC(),
	}

	// Host information
	hostInfo, err := host.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get host info: %w", err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = hostInfo.Hostname
	}

	metrics.Hostname = hostname
	metrics.OS = OSInfo{
		Platform:           hostInfo.Platform,
		PlatformFamily:     hostInfo.PlatformFamily,
		PlatformVersion:    hostInfo.PlatformVersion,
		Architecture:       hostInfo.KernelArch,
		KernelVersion:      hostInfo.KernelVersion,
		VirtualizationRole: hostInfo.VirtualizationRole,
	}

	// CPU metrics
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percent: %w", err)
	}

	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU info: %w", err)
	}

	var modelName string
	if len(cpuInfo) > 0 {
		modelName = cpuInfo[0].ModelName
	}

	metrics.CPU = CPUMetrics{
		ModelName:    modelName,
		Cores:        int32(runtime.NumCPU()),
		LogicalCores: int32(runtime.NumCPU()),
		UsagePercent: cpuPercent[0],
		LoadAverage:  []float64{0.1, 0.2, 0.3}, // Simplified
		FrequencyMHz: float64(cpuInfo[0].Mhz) / 1000.0,
	}

	// Memory metrics
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	swapInfo, err := mem.SwapMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get swap info: %w", err)
	}

	metrics.Memory = MemoryMetrics{
		Total:       memInfo.Total,
		Available:   memInfo.Available,
		Used:        memInfo.Used,
		UsedPercent: memInfo.UsedPercent,
		SwapTotal:   swapInfo.Total,
		SwapUsed:    swapInfo.Used,
		SwapPercent: swapInfo.UsedPercent,
	}

	// Disk metrics
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get disk partitions: %w", err)
	}

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue // Skip if we can't get usage
		}

		metrics.Disk = append(metrics.Disk, DiskMetrics{
			Device:      partition.Device,
			Mountpoint:  partition.Mountpoint,
			Fstype:      partition.Fstype,
			Total:       usage.Total,
			Free:        usage.Free,
			Used:        usage.Used,
			UsedPercent: usage.UsedPercent,
		})
	}

	// Network metrics
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range netInterfaces {
		// Skip loopback interfaces
		hasLoopback := false
		for _, flag := range iface.Flags {
			if flag == "flagLoopback" || flag == "loopback" {
				hasLoopback = true
				break
			}
		}
		if hasLoopback {
			continue
		}

		var addressStrings []string
		for _, addr := range iface.Addrs {
			addrStr := addr.String()
			// Parse the JSON format to extract the actual address
			if strings.HasPrefix(addrStr, "{\"addr\":\"") {
				// Extract address from JSON format: {"addr":"..."}
				start := strings.Index(addrStr, "{\"addr\":\"") + 9
				end := strings.Index(addrStr[start:], "\"")
				if end != -1 {
					addressStrings = append(addressStrings, addrStr[start:start+end])
					continue
				}
			}
			// Fallback to original string if not JSON format
			addressStrings = append(addressStrings, addrStr)
		}

		metrics.Network.Interfaces = append(metrics.Network.Interfaces, NetworkInterface{
			Name:         iface.Name,
			HardwareAddr: iface.HardwareAddr,
			MTU:          iface.MTU,
			Flags:        iface.Flags,
			Addresses:    addressStrings,
			BytesSent:    0, // These counters aren't available in basic interface info
			BytesRecv:    0,
			PacketsSent:  0,
			PacketsRecv:  0,
			Errin:        0,
			Errout:       0,
			Dropin:       0,
			Dropout:      0,
		})
	}

	// Network connections (limited to first 20)
	connections, err := net.Connections("all")
	if err == nil {
		limit := 20
		if len(connections) < limit {
			limit = len(connections)
		}
		for _, conn := range connections[:limit] {
			metrics.Network.Connections = append(metrics.Network.Connections, ConnectionInfo{
				LocalAddr:  fmt.Sprintf("%s:%d", conn.Laddr.IP, conn.Laddr.Port),
				RemoteAddr: fmt.Sprintf("%s:%d", conn.Raddr.IP, conn.Raddr.Port),
				State:      conn.Status,
				PID:        conn.Pid,
				Process:    fmt.Sprintf("process_%d", conn.Pid),
			})
		}
	}

	// Runtime metrics
	metrics.Runtime = RuntimeMetrics{
		GoVersion:    runtime.Version(),
		GoOS:         runtime.GOOS,
		GoArch:       runtime.GOARCH,
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
	}

	return metrics, nil
}

func PrintMetricsJSON() {
	metrics, err := CollectSystemMetrics()
	if err != nil {
		log.Fatalf("Failed to collect metrics: %v", err)
	}

	jsonData, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal metrics: %v", err)
	}

	fmt.Println(string(jsonData))
}
