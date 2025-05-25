//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#define MAX_FILENAME_LEN 64
#define MAX_COMM_LEN 16

struct event {
    __u32 pid;
    __s32 fd;
    char comm[MAX_COMM_LEN];
    char filename[MAX_FILENAME_LEN];
    __u32 type; // 0 = open, 1 = read, 2 = write, 3 = close
};

struct {
    __uint(type, BPF_MAP_TYPE_PERF_EVENT_ARRAY);
} events SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64); // pid << 32 | fd
    __type(value, char[MAX_FILENAME_LEN]);
} fd_to_filename SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64); // pid << 32 | comm_hash
    __type(value, char[MAX_FILENAME_LEN]);
} temp_paths SEC(".maps");

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

static inline __u32 hash_comm(const char *comm) {
    __u32 hash = 0;
    #pragma unroll
    for (int i = 0; i < MAX_COMM_LEN && comm[i]; i++) {
        hash = hash * 31 + comm[i];
    }
    return hash;
}

SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    
    char comm[MAX_COMM_LEN];
    bpf_get_current_comm(comm, sizeof(comm));
    
    char filename[MAX_FILENAME_LEN];
    bpf_probe_read_user_str(filename, sizeof(filename), (void *)ctx->args[1]);
    
    // Send all openat events - filtering will be done in userspace
    struct event e = {};
    e.pid = pid;
    e.type = 0; // open
    __builtin_memcpy(e.comm, comm, MAX_COMM_LEN);
    __builtin_memcpy(e.filename, filename, MAX_FILENAME_LEN);
    
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &e, sizeof(e));
    
    // Store temp path for matching with exit
    __u64 key = ((__u64)pid << 32) | hash_comm(comm);
    bpf_map_update_elem(&temp_paths, &key, filename, BPF_ANY);
    
    return 0;
}

SEC("tracepoint/syscalls/sys_exit_openat")
int trace_openat_exit(struct sys_exit_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    
    char comm[MAX_COMM_LEN];
    bpf_get_current_comm(comm, sizeof(comm));
    
    __u64 temp_key = ((__u64)pid << 32) | hash_comm(comm);
    char *filename = bpf_map_lookup_elem(&temp_paths, &temp_key);
    
    if (filename && ctx->ret >= 0) {
        __u64 fd_key = ((__u64)pid << 32) | (__u32)ctx->ret;
        bpf_map_update_elem(&fd_to_filename, &fd_key, filename, BPF_ANY);
        bpf_map_delete_elem(&temp_paths, &temp_key);
    }
    
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_read")
int trace_read_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    char *filename = bpf_map_lookup_elem(&fd_to_filename, &fd_key);
    
    if (filename) {
        struct event e = {};
        e.pid = pid;
        e.fd = fd;
        e.type = 1; // read
        bpf_get_current_comm(e.comm, sizeof(e.comm));
        __builtin_memcpy(e.filename, filename, MAX_FILENAME_LEN);
        
        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &e, sizeof(e));
    }
    
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_write")
int trace_write_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    char *filename = bpf_map_lookup_elem(&fd_to_filename, &fd_key);
    
    if (filename) {
        struct event e = {};
        e.pid = pid;
        e.fd = fd;
        e.type = 2; // write
        bpf_get_current_comm(e.comm, sizeof(e.comm));
        __builtin_memcpy(e.filename, filename, MAX_FILENAME_LEN);
        
        bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, &e, sizeof(e));
    }
    
    return 0;
}

SEC("tracepoint/syscalls/sys_enter_close")
int trace_close_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    bpf_map_delete_elem(&fd_to_filename, &fd_key);
    
    return 0;
}

char _license[] SEC("license") = "GPL";