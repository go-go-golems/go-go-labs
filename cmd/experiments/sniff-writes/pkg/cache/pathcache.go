package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

// PathCache stores full paths by hash for quick lookup
type PathCache struct {
	mu    sync.RWMutex
	cache map[uint32]string
}

func New() *PathCache {
	return &PathCache{
		cache: make(map[uint32]string),
	}
}

func (pc *PathCache) Set(hash uint32, path string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.cache[hash] = path
}

func (pc *PathCache) Get(hash uint32) (string, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	path, exists := pc.cache[hash]
	return path, exists
}

// Hash function that matches the eBPF implementation
func HashPath(path string) uint32 {
	hash := uint32(0)
	for i, c := range []byte(path) {
		if i >= 256 { // Match eBPF limit
			break
		}
		hash = hash*31 + uint32(c)
	}
	return hash
}

func ResolvePath(event *models.Event, pc *PathCache) string {
	// For open events, we need to resolve the path from /proc since eBPF doesn't provide it
	if event.Type == 0 { // open
		filename := resolveFilenameFromFd(event.Pid, event.Fd)
		if filename != "" {
			// Store in cache with the hash
			hash := HashPath(filename)
			pc.Set(hash, filename)
			return filename
		}
		return ""
	}

	// For other events, try cache first
	if event.PathHash != 0 {
		if path, exists := pc.Get(event.PathHash); exists {
			return path
		}
	}

	// Fallback to /proc resolution
	return resolveFilenameFromFd(event.Pid, event.Fd)
}

func resolveFilenameFromFd(pid uint32, fd int32) string {
	// Try to resolve filename from /proc/PID/fd/FD
	procPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)

	// Use readlink to get the actual file path
	if link, err := os.Readlink(procPath); err == nil {
		// Clean the path to make it more readable
		if abs, err := filepath.Abs(link); err == nil {
			return abs
		}
		return link
	}

	return ""
}
