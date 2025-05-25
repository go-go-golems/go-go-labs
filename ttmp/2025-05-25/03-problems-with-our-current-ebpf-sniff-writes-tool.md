# Problems with Our Current eBPF sniff-writes Tool

This document exhaustively catalogs all the issues we encountered during development and testing of our eBPF file monitoring tool, including both resolved and ongoing problems.

## 1. Initial Setup and Build Issues

### 1.1 Missing Dependencies

**Problem**: Tool failed to compile due to missing system dependencies.

**Symptoms**:
```bash
Error: compile: exec: "clang": executable file not found in $PATH
fatal error: 'bpf/bpf_helpers.h' file not found
fatal error: 'asm/types.h' file not found
```

**Root Cause**: Missing eBPF development tools and kernel headers.

**Resolution**: 
- Added dependency checking to Makefile
- Required packages: `clang`, `llvm`, `libbpf-dev`, `linux-headers-$(uname -r)`

**Status**: ✅ RESOLVED

### 1.2 bpf2go Compilation Issues

**Problem**: Go code generation from eBPF C code failed with include path issues.

**Symptoms**:
```bash
In file included from /usr/include/linux/types.h:5:
/usr/include/linux/types.h:5:10: fatal error: 'asm/types.h' file not found
```

**Resolution**: Added architecture-specific include paths:
```go
//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -cflags "-I/usr/include/x86_64-linux-gnu -mllvm -bpf-stack-size=8192" sniffwrites sniff_writes.c
```

**Status**: ✅ RESOLVED

## 2. eBPF Stack Overflow Issues

### 2.1 Stack Size Limit Exceeded

**Problem**: eBPF verifier rejected programs due to stack usage exceeding 512-byte limit.

**Symptoms**:
```bash
Error: failed to create eBPF collection: program trace_close_enter: load program: permission denied: invalid write to stack R10 off=-528 size=8
```

**Root Cause**: Large structures allocated on stack in eBPF programs.

**Timeline of Attempts**:
1. **First attempt**: Increased stack size with `-mllvm -bpf-stack-size=8192`
2. **Second attempt**: Reduced filename buffer from 256 → 128 → 64 bytes
3. **Final solution**: Used per-CPU arrays instead of stack allocation

**Current Solution**:
```c
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct event);
} scratch_event SEC(".maps");

// Instead of: struct event e = {};
// Use: struct event *e = bpf_map_lookup_elem(&scratch_event, &key);
```

**Status**: ✅ RESOLVED (but with consequences - see filename truncation)

## 3. Filename Truncation Issues

### 3.1 64-Byte Filename Limit

**Problem**: Filenames truncated at 64 characters, making debugging impossible.

**Symptoms**:
```json
{"filename":"/home/manuel/code/wesen/corpo"}
{"filename":"/home/manuel/code/wesen/corporate-headquarters/go-go-labs/cmd/ex"}
```

**Root Cause**: Stack overflow prevention forced us to limit filename buffer size.

**Impact**:
- Can't see which specific files are being accessed
- Directory filtering becomes unreliable
- Debugging information is lost
- User experience severely degraded

**Attempted Solutions**:
1. Increase buffer size → Stack overflow
2. Use dynamic memory → eBPF limitations
3. Userspace filename resolution via `/proc/PID/fd/FD` → Partial success

**Status**: 🔄 PARTIALLY RESOLVED (userspace resolution helps but not complete)

### 3.2 Path Display Issues

**Problem**: Even when full paths are available, display is not user-friendly.

**Issues**:
- Long absolute paths dominate output
- Important filename parts cut off in table format
- Need relative paths from target directory
- Should elide beginning, not end of paths

**Attempted Solution**: `formatFilename()` function to:
- Convert to relative paths when possible
- Elide beginning for long paths: `"...path/to/important/file.txt"`

**Status**: 🔄 IN PROGRESS

## 4. Directory Filtering Problems

### 4.1 False Positives: Events Outside Target Directory

**Problem**: Tool shows file operations for files clearly outside the monitored directory.

**Symptoms**:
```bash
# Monitoring cmd/n8n-cli but seeing:
Process kitty reading from: /home/manuel/code/wesen/corpo...
Process chrome writing to: /tmp/some-unrelated-file
```

**Root Causes**:

#### 4.1.1 Relative vs Absolute Path Confusion
- eBPF captures paths exactly as passed to syscalls
- Some processes use relative paths, others absolute
- Working directory affects relative path resolution

#### 4.1.2 Symlink Resolution Issues
- Target directory might be a symlink
- Files accessed through different symlink paths
- Real path vs symlink path mismatches

#### 4.1.3 Process Working Directory Variations
- Different processes have different current working directories
- Relative paths resolve differently per process
- Our filtering assumes uniform working directory

#### 4.1.4 Inherited File Descriptors
- Child processes inherit parent file descriptors
- FD might point to file in target directory but accessed from different process context
- No easy way to track FD inheritance

**Current Filtering Logic Issues**:
```go
// This logic has several flaws:
absTargetDir := filepath.Join(cwd, config.Directory)
relPath, err := filepath.Rel(absTargetDir, absFilename)
if err != nil || strings.HasPrefix(relPath, "..") {
    return false // May incorrectly filter valid files
}
```

**Status**: ❌ UNRESOLVED - Major issue affecting tool reliability

### 4.2 False Negatives: Missing Valid Events

**Problem**: Tool might miss file operations that should be captured.

**Potential Causes**:
- Overly strict filtering logic
- Race conditions in filename resolution
- Symlink handling edge cases

**Status**: ❓ UNKNOWN EXTENT

## 5. Event Volume and Noise Issues

### 5.1 Overwhelming Number of Events

**Problem**: Tool generates too many events, making useful information hard to find.

**Symptoms**:
- Hundreds of close events per second
- Events from unrelated system processes
- Output scrolls too fast to read

**Sources of Noise**:
1. **System processes**: systemd, kernel threads, etc.
2. **stdin/stdout/stderr**: FDs 0, 1, 2 operations
3. **Temporary files**: `/tmp`, cache files
4. **Socket operations**: Network I/O appearing as file operations
5. **Memory mapped files**: Shared libraries, executables

**Attempted Mitigations**:
```c
// Skip obvious non-file descriptors
if (fd <= 2) return 0;

// Only send close events if we have filename info
char *filename = bpf_map_lookup_elem(&fd_to_filename, &fd_key);
if (filename) {
    // Send event
}
```

**Status**: 🔄 PARTIALLY RESOLVED (but still noisy)

### 5.2 Process Filtering Limitations

**Problem**: Process name filtering is too simple and ineffective.

**Issues**:
- Process names truncated to 16 characters in kernel
- Substring matching too broad (`"python"` matches many processes)
- No parent-child process relationship tracking
- Can't filter by process tree or user

**Status**: ❌ NEEDS IMPROVEMENT

## 6. File Descriptor Tracking Issues

### 6.1 Missing FD→Filename Mappings

**Problem**: Read/write events often have empty filenames.

**Symptoms**:
```bash
[DEBUG] PID=12345 FD=7 COMM=myprocess FILE= TYPE=2
```

**Root Causes**:

#### 6.1.1 Files Opened Before Monitoring
- Process opened file before we started monitoring
- No way to retroactively discover filename
- Common with long-running processes

#### 6.1.2 Inherited File Descriptors
- Child processes inherit parent FDs
- We don't track inheritance relationships
- FD mapping is per-process but inheritance crosses processes

#### 6.1.3 Memory Mapped Files
- mmap() doesn't go through read/write syscalls we monitor
- Shared libraries loaded via mmap
- Memory-mapped I/O bypasses our tracking

#### 6.1.4 Special File Types
- Sockets (network connections)
- Pipes (inter-process communication)
- Device files (/dev/*)
- Pseudo-filesystems (/proc, /sys)

**Mitigation Attempt**: Userspace resolution via `/proc/PID/fd/FD`
```go
func resolveFilenameFromFd(pid uint32, fd int32) string {
    procPath := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
    return os.Readlink(procPath)
}
```

**Limitations of /proc Resolution**:
- Process might exit before we can read /proc
- Permission issues (can't read other users' /proc entries)
- Race conditions
- Performance overhead

**Status**: 🔄 PARTIALLY RESOLVED (but unreliable)

### 6.2 FD Mapping Race Conditions

**Problem**: Race conditions between openat_enter and openat_exit tracking.

**Scenario**:
1. openat_enter fires, stores temp path
2. openat_exit fires, but process already exited
3. Temp path never gets moved to fd_to_filename map
4. Later read/write events have no filename

**Status**: ❌ UNRESOLVED

## 7. Data Structure and Parsing Issues

### 7.1 Struct Size Mismatches

**Problem**: Changes to eBPF struct size cause parsing errors in Go.

**Symptoms**:
```bash
2025/05/25 18:04:13 parsing event: data too short: got 72 bytes, expected 280 bytes
```

**Root Cause**: C struct and Go struct must have identical memory layout.

**When This Occurs**:
- Changing filename buffer size
- Adding new fields to event struct
- Compiler padding differences

**Resolution**: Careful struct synchronization and size validation.

**Status**: ✅ RESOLVED (but fragile)

### 7.2 String Handling Issues

**Problem**: Converting between C char arrays and Go strings.

**Issues**:
- Null termination handling
- UTF-8 vs ASCII assumptions
- Buffer overflow protection
- Memory copying performance

**Current Implementation**:
```go
func cString(data []int8) string {
    var buf []byte
    for _, b := range data {
        if b == 0 {
            break
        }
        buf = append(buf, byte(b))
    }
    return string(buf)
}
```

**Status**: ✅ WORKS but could be optimized

## 8. Output and User Interface Issues

### 8.1 Table Format Problems

**Problem**: Table format doesn't show filenames properly.

**Issues**:
- Filename column gets cut off
- Fixed-width formatting inadequate for variable-length paths
- No intelligent wrapping or truncation

**Status**: 🔄 PARTIALLY RESOLVED (improved but not perfect)

### 8.2 JSON Output Truncation

**Problem**: Even JSON output shows truncated filenames due to eBPF limitations.

**Impact**: Makes tool unsuitable for programmatic use.

**Status**: ❌ UNRESOLVED (fundamental limitation)

### 8.3 Debug Information Overload

**Problem**: Debug mode produces too much information to be useful.

**Need**: Structured debug levels:
- Level 1: Summary statistics
- Level 2: Filtered events with context
- Level 3: All events
- Level 4: eBPF map contents

**Status**: ❌ NEEDS IMPROVEMENT

## 9. Performance and Scalability Issues

### 9.1 High CPU Usage

**Problem**: Tool can consume significant CPU on busy systems.

**Causes**:
- High-frequency syscalls
- Userspace processing overhead
- String operations and path resolution
- /proc filesystem access

**Status**: ❓ NOT FULLY CHARACTERIZED

### 9.2 Memory Usage

**Problem**: eBPF maps have fixed size limits.

**Limitations**:
- fd_to_filename: 1024 entries max
- temp_paths: 1024 entries max
- LRU eviction not implemented

**Consequences**:
- Map full → events lost
- No visibility into map utilization
- No graceful degradation

**Status**: ❌ NEEDS MONITORING AND LIMITS

### 9.3 Event Loss

**Problem**: High event rates can cause perf buffer overflows.

**Symptoms**:
```bash
lost 156 samples
```

**Mitigation**: Larger perf buffers, but increases memory usage.

**Status**: 🔄 PARTIALLY ADDRESSED

## 10. Architecture and Design Issues

### 10.1 Filtering Strategy Problems

**Current Approach**: Send all events to userspace, filter there.

**Problems**:
- Wastes kernel→userspace bandwidth
- Higher CPU usage
- More event loss under load

**Alternative**: More filtering in eBPF, but limited by:
- 64-byte filename truncation
- Complex path resolution logic
- eBPF program complexity limits

**Status**: ❌ FUNDAMENTAL DESIGN TRADE-OFF

### 10.2 Single-Threaded Event Processing

**Problem**: Event processing is single-threaded in userspace.

**Limitations**:
- Can't keep up with high event rates
- No parallel processing of events
- Blocking I/O affects all events

**Status**: ❌ SCALABILITY LIMITATION

## 11. Testing and Reliability Issues

### 11.1 Lack of Comprehensive Tests

**Problem**: Tool difficult to test systematically.

**Missing Tests**:
- Unit tests for filtering logic
- Integration tests with known file operations
- Performance benchmarks
- Edge case handling

**Status**: ❌ MAJOR GAP

### 11.2 Error Handling

**Problem**: Many error conditions not handled gracefully.

**Examples**:
- eBPF program loading failures
- Permission denied scenarios
- Out of memory conditions
- Kernel version compatibility

**Status**: ❌ NEEDS IMPROVEMENT

## Summary of Current Status

### ✅ Resolved Issues
- Build system and dependencies
- Basic eBPF compilation and loading
- Stack overflow (with trade-offs)
- Basic event capture and parsing

### 🔄 Partially Resolved Issues
- Filename truncation (userspace resolution helps)
- Event noise (basic filtering in place)
- Directory filtering (works sometimes)
- Output formatting (improved but not perfect)

### ❌ Major Unresolved Issues
- **Directory filtering reliability**: Core functionality doesn't work properly
- **Filename truncation**: Severely impacts usability
- **Event volume**: Tool produces too much noise
- **FD tracking reliability**: Missing too many filename mappings
- **Performance characteristics**: Not well understood
- **Testing coverage**: Insufficient for reliability

### ❓ Unknown Issues
- Full extent of false positives/negatives
- Performance impact on production systems
- Compatibility across different kernel versions
- Memory leak potential in long-running scenarios

## Recommendations for Next Steps

1. **Fix directory filtering**: This is the highest priority issue
2. **Implement better filename handling**: Consider alternative approaches
3. **Add comprehensive testing**: Especially for edge cases
4. **Improve performance monitoring**: Add metrics and limits
5. **Better error handling**: Graceful degradation and user feedback
6. **Consider architectural changes**: Maybe eBPF + userspace hybrid approach

The tool shows promise but has significant reliability and usability issues that need to be addressed before it can be considered production-ready or even useful for debugging purposes.