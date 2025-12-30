package utils

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// DeviceInfo contains device information
type DeviceInfo struct {
	Hostname      string
	OSType        string
	OSVersion     string
	Architecture  string
	CPUModel      string
	CPUCores      int
	MemoryTotal   uint64
	DiskTotal     uint64
	IPAddresses    []string
	HardwareInfo  map[string]string
}

// CollectDeviceInfo collects comprehensive device information
func CollectDeviceInfo() (*DeviceInfo, error) {
	info := &DeviceInfo{
		OSType:       runtime.GOOS,
		Architecture: runtime.GOARCH,
		HardwareInfo: make(map[string]string),
	}

	// Get hostname
	hostname, err := host.Info()
	if err == nil {
		info.Hostname = hostname.Hostname
		info.OSVersion = fmt.Sprintf("%s %s", hostname.Platform, hostname.PlatformVersion)
		info.HardwareInfo["host_id"] = hostname.HostID
		info.HardwareInfo["platform"] = hostname.Platform
		info.HardwareInfo["platform_family"] = hostname.PlatformFamily
		info.HardwareInfo["platform_version"] = hostname.PlatformVersion
		info.HardwareInfo["kernel_version"] = hostname.KernelVersion
		info.HardwareInfo["kernel_arch"] = hostname.KernelArch
	}

	// Get CPU info
	cpuInfo, err := cpu.Info()
	if err == nil && len(cpuInfo) > 0 {
		info.CPUModel = cpuInfo[0].ModelName
		info.CPUCores = int(cpuInfo[0].Cores)
		info.HardwareInfo["cpu_model"] = cpuInfo[0].ModelName
		info.HardwareInfo["cpu_cores"] = fmt.Sprintf("%d", cpuInfo[0].Cores)
		info.HardwareInfo["cpu_family"] = cpuInfo[0].Family
		info.HardwareInfo["cpu_vendor_id"] = cpuInfo[0].VendorID
	}

	// Get memory info
	memInfo, err := mem.VirtualMemory()
	if err == nil {
		info.MemoryTotal = memInfo.Total
		info.HardwareInfo["memory_total"] = fmt.Sprintf("%d", memInfo.Total)
	}

	// Get disk info
	diskInfo, err := disk.Usage("/")
	if err == nil {
		info.DiskTotal = diskInfo.Total
		info.HardwareInfo["disk_total"] = fmt.Sprintf("%d", diskInfo.Total)
		info.HardwareInfo["disk_fstype"] = diskInfo.Fstype
	}

	// Get network interfaces
	interfaces, err := net.Interfaces()
	if err == nil {
		var ipAddresses []string
		for _, iface := range interfaces {
			for _, addr := range iface.Addrs {
				// Filter out loopback and link-local addresses
				addrStr := addr.Addr
				if !strings.Contains(addrStr, "::1") &&
					!strings.Contains(addrStr, "127.0.0.1") &&
					!strings.HasPrefix(addrStr, "fe80:") {
					// Remove CIDR notation
					if idx := strings.Index(addrStr, "/"); idx > 0 {
						addrStr = addrStr[:idx]
					}
					ipAddresses = append(ipAddresses, addrStr)
				}
			}
		}
		info.IPAddresses = ipAddresses
		if len(ipAddresses) > 0 {
			info.HardwareInfo["primary_ip"] = ipAddresses[0]
		}
	}

	return info, nil
}

// GetDeviceSerial attempts to get a unique device serial number
func GetDeviceSerial() string {
	hostInfo, err := host.Info()
	if err == nil && hostInfo.HostID != "" {
		return hostInfo.HostID
	}
	
	// Fallback: use hostname + architecture
	hostname, _ := host.Info()
	if hostname != nil {
		return fmt.Sprintf("%s-%s-%s", hostname.Hostname, runtime.GOOS, runtime.GOARCH)
	}
	
	return fmt.Sprintf("unknown-%s-%s", runtime.GOOS, runtime.GOARCH)
}

// GetDeviceID generates a unique device ID
func GetDeviceID() string {
	hostInfo, err := host.Info()
	if err == nil && hostInfo.HostID != "" {
		return hostInfo.HostID
	}
	
	// Fallback
	hostname, _ := host.Info()
	if hostname != nil {
		return fmt.Sprintf("%s-%s", hostname.Hostname, runtime.GOARCH)
	}
	
	return fmt.Sprintf("device-%s-%s", runtime.GOOS, runtime.GOARCH)
}

