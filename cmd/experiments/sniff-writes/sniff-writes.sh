#!/usr/bin/env bash
#
sudo bpftrace -e '
// Track file descriptors associated with cmd/n8n-cli
BEGIN { printf("Monitoring reads/writes to cmd/n8n-cli...\n"); }

// Store file descriptors for files in cmd/n8n-cli
tracepoint:syscalls:sys_enter_openat /strncmp(str(args->filename), "cmd/n8n-cli", 11) == 0/ { 
  printf("Process %s (PID %d) opening cmd/n8n-cli file: %s\n", comm, pid, str(args->filename));
  // Store the mapping between returned fd and filename (will be populated on exit)
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
}'
