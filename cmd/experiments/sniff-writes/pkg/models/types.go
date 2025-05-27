package models

import "time"

type Event struct {
	Pid         uint32
	Fd          int32
	Comm        [16]int8
	PathHash    uint32     // 32-bit hash of the path for cache lookup
	Type        uint32     // 0 = open, 1 = read, 2 = write, 3 = close
	WriteSize   uint64     // Total size of write/read operation
	FileOffset  uint64     // File offset where the operation occurs
	ContentLen  uint32     // Actual content captured in this chunk
	ChunkSeq    uint32     // Sequence number for chunked events (0-based)
	TotalChunks uint32     // Total number of chunks for this operation
	Content     [4096]int8 // Write/read content
}

type Config struct {
	Directory          string
	OutputFormat       string
	Operations         []string
	ProcessFilter      string
	Duration           time.Duration
	Verbose            bool
	ShowFd             bool
	OutputFile         string
	Debug              bool
	ShowAllFiles       bool     // Show pipes, sockets, etc. (default: false)
	CaptureContent     bool     // Capture write content (default: false)
	ContentSize        int      // Max content bytes to capture (default: 4096)
	GlobPatterns       []string // Include patterns for file filtering
	GlobExclude        []string // Exclude patterns for file filtering
	ProcessGlob        []string // Include patterns for process name filtering
	ProcessGlobExclude []string // Exclude patterns for process name filtering
	SqliteDB           string   // Path to SQLite database for logging
	WebUI              bool     // Enable web UI
	WebPort            int      // Web UI port
}

type EventOutput struct {
	Timestamp   string `json:"timestamp"`
	Pid         uint32 `json:"pid"`
	Process     string `json:"process"`
	Operation   string `json:"operation"`
	Filename    string `json:"filename"`
	Fd          int32  `json:"fd,omitempty"`
	WriteSize   uint64 `json:"write_size,omitempty"`
	FileOffset  uint64 `json:"file_offset,omitempty"`
	Content     string `json:"content,omitempty"`
	ChunkSeq    uint32 `json:"chunk_seq,omitempty"`
	TotalChunks uint32 `json:"total_chunks,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
}
