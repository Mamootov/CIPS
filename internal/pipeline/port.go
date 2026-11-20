package pipeline

import (
	"fmt"
	"net"
	"sync"
)

// PortManager manages ephemeral local port allocation and tracking.
type PortManager struct {
	mu            sync.Mutex
	reservedPorts map[int]bool
}

// NewPortManager creates a thread-safe port manager.
func NewPortManager() *PortManager {
	return &PortManager{
		reservedPorts: make(map[int]bool),
	}
}

// AllocatePort finds an available ephemeral port on 127.0.0.1 and marks it reserved.
func (pm *PortManager) AllocatePort() (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for attempts := 0; attempts < 200; attempts++ {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			continue
		}
		port := l.Addr().(*net.TCPAddr).Port
		_ = l.Close()

		if !pm.reservedPorts[port] {
			pm.reservedPorts[port] = true
			return port, nil
		}
	}

	return 0, fmt.Errorf("unable to find free ephemeral port after 200 attempts")
}

// AllocateBatch allocates N distinct available ports.
func (pm *PortManager) AllocateBatch(count int) ([]int, error) {
	ports := make([]int, 0, count)
	for i := 0; i < count; i++ {
		port, err := pm.AllocatePort()
		if err != nil {
			pm.ReleaseBatch(ports)
			return nil, err
		}
		ports = append(ports, port)
	}
	return ports, nil
}

// ReleasePort frees a previously reserved port.
func (pm *PortManager) ReleasePort(port int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.reservedPorts, port)
}

// ReleaseBatch frees a collection of ports.
func (pm *PortManager) ReleaseBatch(ports []int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	for _, p := range ports {
		delete(pm.reservedPorts, p)
	}
}

// DefaultPortManager global singleton instance for pipeline execution.
var DefaultPortManager = NewPortManager()
