# sniff-writes
:

A Go port of the bpftrace script for monitoring file reads/writes to cmd/n8n-cli files.

## Prerequisites

This tool requires:
- Linux kernel with eBPF support (4.1+)
- clang compiler
- LLVM tools
- libbpf headers

On Ubuntu/Debian:
```bash
sudo apt install clang llvm libbpf-dev linux-headers-$(uname -r)
```

## Building

```bash
make build
```

This will:
1. Generate eBPF bytecode from the C source
2. Generate Go bindings
3. Build the final binary

## Running

The program requires root privileges to load eBPF programs:

```bash
sudo make run
```

Or manually:
```bash
sudo ./sniff-writes
```

## What it does

This tool monitors system calls and tracks:
- Files opened in the `cmd/n8n-cli` directory
- Read operations on those files
- Write operations on those files
- When file descriptors are closed

The output shows the process name, PID, and file being accessed.

## Original bpftrace script

This is a Go port of the equivalent bpftrace script:

```bash
sudo bpftrace -e '
// Track file descriptors associated with cmd/n8n-cli
BEGIN { printf("Monitoring reads/writes to cmd/n8n-cli...\n"); }

// Store file descriptors for files in cmd/n8n-cli
tracepoint:syscalls:sys_enter_openat /strncmp(str(args->filename), "cmd/n8n-cli", 11) == 0/ { 
  printf("Process %s (PID %d) opening cmd/n8n-cli file: %s\n", comm, pid, str(args->filename));
  @paths[pid, comm] = str(args->filename);
}

// Capture the returned fd from openat
tracepoint:syscalls:sys_exit_openat /@paths[pid, comm] != NULL/ {
  if (args->ret >= 0) {
    @fds[pid, args->ret] = @paths[pid, comm];
    printf("File descriptor %d assigned for %s\n", args->ret, @paths[pid, comm]);
    delete(@paths[pid, comm]);
  }
}

// Track read operations on our tracked file descriptors
tracepoint:syscalls:sys_enter_read {
  $fd = args->fd;
  $filename = @fds[pid, $fd];
  if ($filename != "") {
    printf("Process %s (PID %d) reading from cmd/n8n-cli file: %s (fd: %d)\n", comm, pid, $filename, $fd);
  }
}

// Track write operations on our tracked file descriptors
tracepoint:syscalls:sys_enter_write {
  $fd = args->fd;
  $filename = @fds[pid, $fd];
  if ($filename != "") {
    printf("Process %s (PID %d) writing to cmd/n8n-cli file: %s (fd: %d)\n", comm, pid, $filename, $fd);
  }
}

// Clean up tracking when files are closed
tracepoint:syscalls:sys_enter_close {
  delete(@fds[pid, args->fd]);
}
'
```
