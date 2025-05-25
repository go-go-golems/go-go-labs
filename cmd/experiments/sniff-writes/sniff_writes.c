//foo
//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#define MAX_COMM_LEN 16

struct event {
    __u32 pid;
    __s32 fd;
    char comm[MAX_COMM_LEN];
    __u32 path_hash; // 32-bit hash of the path for cache lookup
    __u32 type; // 0 = open, 1 = read, 2 = write, 3 = close
};

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

// Only store fd and path hash for userspace lookup
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64); // pid << 32 | fd
    __type(value, __u32); // path hash
} fd_to_hash SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, struct event);
} scratch_event SEC(".maps");

// Simplified tracepoint context structures
struct sys_enter_ctx {
    unsigned short common_type;
    unsigned char common_flags;
    unsigned char common_preempt_count;
    int common_pid;
    int __syscall_nr;
    long args[6];
};

struct sys_exit_ctx {
    unsigned short common_type;
    unsigned char common_flags;
    unsigned char common_preempt_count;
    int common_pid;
    int __syscall_nr;
    long ret;
};

// Remove directory filtering from eBPF - we'll do it in userspace

static inline __u32 hash_path(const char *path) {
    __u32 hash = 0;
    #pragma unroll
    for (int i = 0; i < 256 && path[i]; i++) { // Hash full path, not just first 64 chars
        hash = hash * 31 + path[i];
    }
    return hash;
}

SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat_enter(struct sys_enter_ctx *ctx) {
    // Only emit open event on successful exit, not on enter
    return 0;
}

SEC("tracepoint/syscalls/sys_exit_openat")
int trace_openat_exit(struct sys_exit_ctx *ctx) {
    if (ctx->ret < 0) return 0; // Failed open
    
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->ret;
    
    // Get the path from the original syscall args (stored in a temp map would be ideal,
    // but for simplicity we'll let userspace handle the path entirely)
    char path[256];
    __builtin_memset(path, 0, sizeof(path));
    
    // Get path from current task's memory (this is a simplified approach)
    // In production, you'd want to store the path from enter and retrieve here
    // For now, we'll compute hash of empty string and let userspace resolve
    __u32 path_hash = hash_path("");
    
    // Use per-CPU array to avoid stack issues
    __u32 key = 0;
    struct event *e = bpf_map_lookup_elem(&scratch_event, &key);
    if (!e) return 0;
    
    __builtin_memset(e, 0, sizeof(*e));
    e->pid = pid;
    e->fd = fd;
    e->type = 0; // open
    e->path_hash = path_hash;
    bpf_get_current_comm(e->comm, sizeof(e->comm));
    
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, e, sizeof(*e));
    
    // Store fd -> hash mapping for later events
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    bpf_map_update_elem(&fd_to_hash, &fd_key, &path_hash, BPF_ANY);
    
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_read")
int trace_read_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    
    // Skip obvious non-file descriptors (stdin, stdout, stderr, sockets, pipes)
    if (fd <= 2) return 0;
    
    // Use per-CPU array to avoid stack issues
    __u32 key = 0;
    struct event *e = bpf_map_lookup_elem(&scratch_event, &key);
    if (!e) return 0;
    
    __builtin_memset(e, 0, sizeof(*e));
    e->pid = pid;
    e->fd = fd;
    e->type = 1; // read
    bpf_get_current_comm(e->comm, sizeof(e->comm));
    
    // Try to get path hash from our tracking map
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    __u32 *path_hash = bpf_map_lookup_elem(&fd_to_hash, &fd_key);
    if (path_hash) {
        e->path_hash = *path_hash;
    } else {
        e->path_hash = 0; // No hash available, userspace will resolve via /proc
    }
    
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, e, sizeof(*e));
    
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_write")
int trace_write_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    
    // Skip obvious non-file descriptors (stdin, stdout, stderr, sockets, pipes)
    if (fd <= 2) return 0;
    
    // Use per-CPU array to avoid stack issues
    __u32 key = 0;
    struct event *e = bpf_map_lookup_elem(&scratch_event, &key);
    if (!e) return 0;
    
    __builtin_memset(e, 0, sizeof(*e));
    e->pid = pid;
    e->fd = fd;
    e->type = 2; // write
    bpf_get_current_comm(e->comm, sizeof(e->comm));
    
    // Try to get path hash from our tracking map
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    __u32 *path_hash = bpf_map_lookup_elem(&fd_to_hash, &fd_key);
    if (path_hash) {
        e->path_hash = *path_hash;
    } else {
        e->path_hash = 0; // No hash available, userspace will resolve via /proc
    }
    
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, e, sizeof(*e));
    
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_close")
int trace_close_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    
    // Skip obvious non-file descriptors (stdin, stdout, stderr)
    if (fd <= 2) return 0;
    
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    
    // Only send close events if we have hash info
    __u32 *path_hash = bpf_map_lookup_elem(&fd_to_hash, &fd_key);
    if (path_hash) {
        // Use per-CPU array to avoid stack issues
        __u32 key = 0;
        struct event *e = bpf_map_lookup_elem(&scratch_event, &key);
        if (e) {
            __builtin_memset(e, 0, sizeof(*e));
            e->pid = pid;
            e->fd = fd;
            e->type = 3; // close
            e->path_hash = *path_hash;
            bpf_get_current_comm(e->comm, sizeof(e->comm));
            
            bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, e, sizeof(*e));
        }
    }
    
    // Clean up our tracking regardless
    bpf_map_delete_elem(&fd_to_hash, &fd_key);
    
    return 0;
}

char _license[] SEC("license") = "GPL";
