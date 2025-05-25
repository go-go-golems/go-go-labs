# eBPF File Monitoring Tool: Complete Technical Guide

## Overview

This tool monitors file operations using eBPF (Extended Berkeley Packet Filter) to track when processes open, read, write, or close files in specific directories. Let me break down every component and technique used.

## eBPF Fundamentals

### What is eBPF?

eBPF is a technology that allows you to run sandboxed programs in kernel space without changing kernel source code. Think of it as a "virtual machine" inside the Linux kernel that can:

1. **Hook into kernel events** (syscalls, tracepoints, kprobes)
2. **Process data in real-time** with minimal overhead
3. **Communicate with userspace** via maps and perf events
4. **Filter and aggregate data** before sending to userspace

### Key eBPF Components

1. **Programs**: C code compiled to eBPF bytecode that runs in kernel
2. **Maps**: Data structures shared between kernel and userspace
3. **Verifier**: Ensures eBPF programs are safe to run
4. **Helper functions**: Kernel APIs available to eBPF programs

## Our Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Syscalls      │    │   eBPF Programs  │    │  Go Userspace   │
│                 │    │                  │    │                 │
│ openat()        │───▶│ trace_openat_*   │───▶│ Event Parser    │
│ read()          │───▶│ trace_read_*     │───▶│ Filtering       │
│ write()         │───▶│ trace_write_*    │───▶│ Output Format   │
│ close()         │───▶│ trace_close_*    │───▶│ CLI Interface   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────┐
                       │  eBPF Maps   │
                       │              │
                       │ • events     │
                       │ • fd_mapping │
                       │ • temp_paths │
                       │ • scratch    │
                       └──────────────┘
```

## Syscall Monitoring Strategy

### Why These Syscalls?

1. **`openat()`**: Modern way to open files (replaces `open()`)
2. **`read()`**: Read data from file descriptors
3. **`write()`**: Write data to file descriptors  
4. **`close()`**: Close file descriptors

### Tracepoints vs Kprobes

We use **tracepoints** because they:
- Are stable kernel ABI (won't break with kernel updates)
- Have predefined argument structures
- Are more efficient than kprobes
- Provide both enter/exit events

## eBPF Code Deep Dive

### Data Structures

```c
// Event sent to userspace
struct event {
    __u32 pid;           // Process ID
    __s32 fd;            // File descriptor
    char comm[16];       // Process name
    char filename[64];   // Filename (truncated)
    __u32 type;          // Operation type (0=open, 1=read, 2=write, 3=close)
};
```

### Map Types and Their Purposes

```c
// 1. Perf event array - sends events to userspace
struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

// 2. Hash map - tracks fd -> filename mapping
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64);     // (pid << 32) | fd
    __type(value, char[64]);
} fd_to_filename SEC(".maps");

// 3. Temporary storage for openat correlation
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64);     // (pid << 32) | comm_hash
    __type(value, char[64]);
} temp_paths SEC(".maps");

// 4. Per-CPU scratch space to avoid stack overflow
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct event);
} scratch_event SEC(".maps");
```

### The File Descriptor Tracking Problem

**Challenge**: When we see a `read()` or `write()` syscall, we only get a file descriptor number, not the filename.

**Solution**: Track the mapping between file descriptors and filenames:

1. **openat_enter**: Capture the filename being opened
2. **openat_exit**: If successful, store `fd -> filename` mapping
3. **read/write**: Look up filename from fd mapping
4. **close**: Clean up the fd mapping

### Stack Overflow Issues

eBPF has a 512-byte stack limit. Large structures cause "invalid write to stack" errors.

**Solutions**:
1. **Per-CPU arrays**: Allocate structures in maps instead of stack
2. **Smaller structures**: Limit filename to 64 bytes
3. **Careful memory management**: Use `__builtin_memset()` and `__builtin_memcpy()`

## Go Userspace Code

### eBPF Integration with cilium/ebpf

```go
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 sniffwrites sniff_writes.c
```

This generates:
- `sniffwrites_bpfel.go`: Go bindings for eBPF programs/maps
- `sniffwrites_bpfel.o`: Compiled eBPF bytecode

### Loading and Attaching eBPF Programs

```go
// 1. Load eBPF bytecode into kernel
spec, err := loadSniffwrites()
coll, err := ebpf.NewCollection(spec)

// 2. Attach to tracepoints
link.Tracepoint("syscalls", "sys_enter_openat", coll.Programs["trace_openat_enter"], nil)
```

### Event Processing Pipeline

```go
// 1. Create perf event reader
rd, err := perf.NewReader(coll.Maps["events"], os.Getpagesize())

// 2. Read events in loop
for {
    record, err := rd.Read()
    parseEvent(record.RawSample, &event)
    
    // 3. Filter and process
    if shouldProcessEvent(&event) {
        outputEvent(&event, outputWriter)
    }
}
```

## Data Flow Analysis

### 1. openat() Syscall Flow

```
Process calls openat("/path/to/file", flags) 
    ↓
sys_enter_openat tracepoint fires
    ↓
eBPF captures: PID, comm, filename
    ↓
Send event to userspace via perf buffer
    ↓
Store in temp_paths map: (pid|comm_hash) -> filename
    ↓
sys_exit_openat tracepoint fires  
    ↓
If successful (ret >= 0): fd_to_filename[pid|fd] = filename
```

### 2. read()/write() Syscall Flow

```
Process calls read(fd, buffer, size)
    ↓
sys_enter_read tracepoint fires
    ↓
eBPF looks up: fd_to_filename[pid|fd] -> filename
    ↓
Send event with filename to userspace
    ↓
Go code applies filters and formats output
```

## Current Issues and Limitations

### 1. Filename Truncation (64 bytes)

**Problem**: Long paths get cut off, making debugging difficult.

**Causes**:
- eBPF stack size limitations
- Fixed-size structures for performance

**Solutions**:
- Use longer paths in per-CPU arrays (done)
- Implement filename resolution in userspace via `/proc/PID/fd/FD`
- Hash-based filename storage

### 2. Directory Filtering Issues

**Problem**: Seeing events for files outside target directory.

**Root Causes**:
1. **Relative vs Absolute Paths**: eBPF captures paths as-is from syscall
2. **Symlinks**: Real path may differ from captured path
3. **Working Directory**: Relative paths depend on process CWD
4. **Race Conditions**: Fast-changing file operations

**Current Filtering Logic**:
```go
// Convert to absolute paths for comparison
absFilename := makeAbs(filename)
absTargetDir := makeAbs(config.Directory)

// Check if file is within target directory
relPath, err := filepath.Rel(absTargetDir, absFilename)
if err != nil || strings.HasPrefix(relPath, "..") {
    return false // Outside target directory
}
```

### 3. Missing File Descriptor Mappings

**Problem**: read/write events with empty filenames.

**Causes**:
1. Files opened before monitoring started
2. File descriptors inherited from parent processes
3. Memory mapped files
4. Special file types (sockets, pipes, devices)

**Mitigation**: Userspace filename resolution via `/proc`:
```go
func resolveFilenameFromFd(pid uint32, fd int32) string {
    procPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
    return os.Readlink(procPath) // Returns actual file path
}
```

## Debugging Techniques

### 1. Enable Debug Mode

```bash
sudo ./sniff-writes monitor --debug -v
```

Shows all events regardless of filters with detailed info:
```
[DEBUG] PID=12345 FD=3 COMM=myprocess FILE=/path/to/file TYPE=1
```

### 2. Check eBPF Program Loading

```bash
# List loaded eBPF programs
sudo bpftool prog list

# Show program details
sudo bpftool prog show id 123

# Dump program instructions
sudo bpftool prog dump xlated id 123
```

### 3. Monitor eBPF Maps

```bash
# List eBPF maps
sudo bpftool map list

# Dump map contents
sudo bpftool map dump id 456
```

### 4. Kernel Logs

eBPF verifier errors appear in kernel logs:
```bash
sudo dmesg | grep -i bpf
journalctl -k | grep -i bpf
```

### 5. Test with Simple Cases

```bash
# Test with a specific process
sudo ./sniff-writes monitor -p "myprocess" -v

# Test with current directory
sudo ./sniff-writes monitor -d . -f table

# Test specific operations only
sudo ./sniff-writes monitor -o write -f json
```

## Performance Considerations

### 1. Event Volume

**Problem**: High-frequency syscalls can overwhelm the system.

**Mitigations**:
- Filter in eBPF (skip stdin/stdout/stderr)
- Process-specific filtering
- Directory-specific filtering
- Sampling (future enhancement)

### 2. Memory Usage

**Maps size limits**:
- `fd_to_filename`: 1024 entries max
- `temp_paths`: 1024 entries max
- Can cause event loss if exceeded

### 3. CPU Overhead

eBPF programs must be:
- **Fast**: No loops, limited instructions
- **Safe**: Pass verifier checks
- **Bounded**: Predictable execution time

## Advanced Debugging Scenarios

### Scenario 1: "Why am I seeing events outside my directory?"

1. **Check the actual filename**: Use debug mode to see raw filenames
2. **Verify path resolution**: Check if symlinks are involved
3. **Process working directory**: Different processes may have different CWDs
4. **Inherited file descriptors**: Child processes inherit parent FDs

### Scenario 2: "Why are filenames truncated?"

1. **eBPF structure size**: Currently limited to 64 bytes
2. **Stack overflow prevention**: Larger structures cause kernel errors
3. **Userspace resolution**: Use `/proc/PID/fd/FD` for full paths

### Scenario 3: "Why do I see empty filenames?"

1. **Missing FD mapping**: File opened before monitoring
2. **Special file types**: Sockets, pipes, devices
3. **Memory mapped files**: Not tracked by read/write syscalls
4. **Race conditions**: FD closed before we could track it

## Extending the Tool

### 1. Add More Syscalls

Monitor additional operations:
```c
// Add openat2, statx, etc.
SEC("tracepoint/syscalls/sys_enter_openat2")
int trace_openat2_enter(struct sys_enter_ctx *ctx) { ... }
```

### 2. Better Filename Handling

Use dynamic memory or string maps:
```c
// Variable-length string storage
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __type(key, __u64);
    __type(value, char[256]);  // Larger buffer
} long_filenames SEC(".maps");
```

### 3. Process Tree Tracking

Track parent-child relationships:
```c
struct process_info {
    __u32 pid;
    __u32 ppid;
    char comm[16];
};
```

### 4. File Content Inspection

Capture read/write data (be careful with privacy/performance):
```c
struct io_event {
    __u32 pid;
    __s32 fd;
    char data[64];  // First N bytes
    __u32 size;     // Total operation size
};
```

This guide should give you a comprehensive understanding of how the tool works and how to debug and extend it. The key is understanding the flow from syscall → eBPF → userspace and the limitations at each stage.