//go:build ignore

#include <linux/bpf.h>
#include <bpf/bpf_helpers.h>

#define MAX_COMM_LEN 16
#define MAX_CONTENT_LEN 4096
#define MAX_CHUNKS 32

struct event {
    __u32 pid;
    __s32 fd;
    char comm[MAX_COMM_LEN];
    __u32 path_hash; // 32-bit hash of the path for cache lookup
    __u32 type; // 0 = open, 1 = read, 2 = write, 3 = close
    __u64 write_size; // Total size of write operation
    __u64 file_offset; // File offset where the operation occurs
    __u32 content_len; // Actual content captured in this chunk
    __u32 chunk_seq; // Sequence number for chunked events (0-based)
    __u32 total_chunks; // Total number of chunks for this operation
    char content[MAX_CONTENT_LEN]; // Write/read content
};

// Temporary structure to store read buffer info between enter/exit
struct read_info {
    __u64 buf_addr; // Store as address instead of pointer
    __u64 count;
    __u64 offset;
    __s32 fd;
};

struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 24); // 16 MiB
} events SEC(".maps");

// Only store fd and path hash for userspace lookup
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64); // pid << 32 | fd
    __type(value, __u32); // path hash
} fd_to_hash SEC(".maps");

// Store read buffer info between enter/exit
struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 1024);
    __type(key, __u64); // pid << 32 | tid
    __type(value, struct read_info);
} read_buffers SEC(".maps");

// Control map for content capture (single entry: key=0, value=1 means enabled)
struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u32);
} content_capture_enabled SEC(".maps");



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
    for (int i = 0; i < 256; i++) path[i] = 0;
    
    // Get path from current task's memory (this is a simplified approach)
    // In production, you'd want to store the path from enter and retrieve here
    // For now, we'll compute hash of empty string and let userspace resolve
    __u32 path_hash = hash_path("");
    
    // Reserve ring buffer space
    struct event *e = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
    if (!e) return 0;
    
    e->pid = pid;
    e->fd = fd;
    e->type = 0; // open
    e->path_hash = path_hash;
    e->write_size = 0;
    e->file_offset = 0;
    e->content_len = 0;
    e->chunk_seq = 0;
    e->total_chunks = 1;
    bpf_get_current_comm(e->comm, sizeof(e->comm));
    
    bpf_ringbuf_submit(e, 0);
    
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
    void *buf = (void *)ctx->args[1];
    __u64 count = (__u64)ctx->args[2];
    
    // Skip obvious non-file descriptors (stdin, stdout, stderr, sockets, pipes)
    if (fd <= 2) return 0;
    
    // Store read info for the exit handler
    __u64 tid_key = pid_tgid; // Full pid_tgid as key
    struct read_info info = {
        .buf_addr = (__u64)buf,
        .count = count,
        .offset = 0, // Will be updated if we can determine it
        .fd = fd,
    };
    
    bpf_map_update_elem(&read_buffers, &tid_key, &info, BPF_ANY);
    
    return 0;
}

static inline void emit_read_chunk(struct sys_exit_ctx *ctx, __u32 pid, __s32 fd,
                                  void *buf, __u64 total_size, __u64 offset,
                                  __u32 chunk_size, __u32 chunk_seq, __u32 total_chunks,
                                  __u32 path_hash) {
    struct event *e = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
    if (!e) return;
    
    e->pid = pid;
    e->fd = fd;
    e->type = 1; // read
    e->write_size = total_size; // For reads, this is the bytes read
    e->file_offset = offset;
    e->content_len = chunk_size;
    e->chunk_seq = chunk_seq;
    e->total_chunks = total_chunks;
    e->path_hash = path_hash;
    bpf_get_current_comm(e->comm, sizeof(e->comm));
    
    if (chunk_size > 0) {
        bpf_probe_read_user(e->content, chunk_size, buf);
    }
    
    bpf_ringbuf_submit(e, 0);
}

SEC("tracepoint/syscalls/sys_exit_read")
int trace_read_exit(struct sys_exit_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    long ret = ctx->ret;
    
    if (ret <= 0) {
        // Clean up on error or EOF
        bpf_map_delete_elem(&read_buffers, &pid_tgid);
        return 0;
    }
    
    // Look up stored read info
    struct read_info *info = bpf_map_lookup_elem(&read_buffers, &pid_tgid);
    if (!info) return 0;
    
    __s32 fd = info->fd;
    __u64 bytes_read = (__u64)ret;
    
    // Get path hash from our tracking map
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    __u32 *path_hash_ptr = bpf_map_lookup_elem(&fd_to_hash, &fd_key);
    __u32 path_hash = path_hash_ptr ? *path_hash_ptr : 0;
    
    // Check if content capture is enabled
    __u32 capture_key = 0;
    __u32 *enabled = bpf_map_lookup_elem(&content_capture_enabled, &capture_key);
    
    if (!enabled || !*enabled) {
        // Emit event without content
        emit_read_chunk(ctx, pid, fd, (void *)info->buf_addr, bytes_read, info->offset, 0, 0, 1, path_hash);
        bpf_map_delete_elem(&read_buffers, &pid_tgid);
        return 0;
    }
    
    // Calculate total chunks needed
    __u32 total_chunks = (bytes_read + MAX_CONTENT_LEN - 1) / MAX_CONTENT_LEN;
    if (total_chunks > MAX_CHUNKS) {
        total_chunks = MAX_CHUNKS;
    }
    
    // Emit chunks with content
    #pragma unroll
    for (__u32 chunk = 0; chunk < MAX_CHUNKS; chunk++) {
        if (chunk >= total_chunks) break;
        
        __u64 chunk_offset = chunk * MAX_CONTENT_LEN;
        __u32 chunk_size = bytes_read - chunk_offset;
        if (chunk_size > MAX_CONTENT_LEN) {
            chunk_size = MAX_CONTENT_LEN;
        }
        if (chunk_size == 0) break;
        
        emit_read_chunk(ctx, pid, fd, (void *)(info->buf_addr + chunk_offset), bytes_read, 
                       info->offset + chunk_offset, chunk_size, chunk, total_chunks, path_hash);
    }
    
    // Clean up
    bpf_map_delete_elem(&read_buffers, &pid_tgid);
    return 0;
}

static inline void emit_write_chunk(struct sys_enter_ctx *ctx, __u32 pid, __s32 fd, 
                                    void *buf, __u64 total_size, __u64 offset,
                                    __u32 chunk_size, __u32 chunk_seq, __u32 total_chunks,
                                    __u32 path_hash) {
    struct event *e = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
    if (!e) return;
    
    e->pid = pid;
    e->fd = fd;
    e->type = 2; // write
    e->write_size = total_size;
    e->file_offset = offset;
    e->content_len = chunk_size;
    e->chunk_seq = chunk_seq;
    e->total_chunks = total_chunks;
    e->path_hash = path_hash;
    bpf_get_current_comm(e->comm, sizeof(e->comm));
    
    if (chunk_size > 0) {
        bpf_probe_read_user(e->content, chunk_size, buf);
    }
    
    bpf_ringbuf_submit(e, 0);
}

SEC("tracepoint/syscalls/sys_enter_write")
int trace_write_enter(struct sys_enter_ctx *ctx) {
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    __u32 pid = pid_tgid >> 32;
    __s32 fd = (__s32)ctx->args[0];
    void *buf = (void *)ctx->args[1];
    __u64 count = (__u64)ctx->args[2];
    
    // Skip obvious non-file descriptors (stdin, stdout, stderr, sockets, pipes)
    if (fd <= 2) return 0;
    
    // Try to get file offset using lseek(fd, 0, SEEK_CUR)
    // Note: This is a simplified approach, in production you'd want to track offsets more accurately
    __u64 offset = 0; // We'll set this to 0 for now, could be enhanced with file position tracking
    
    // Get path hash from our tracking map
    __u64 fd_key = ((__u64)pid << 32) | (__u32)fd;
    __u32 *path_hash_ptr = bpf_map_lookup_elem(&fd_to_hash, &fd_key);
    __u32 path_hash = path_hash_ptr ? *path_hash_ptr : 0;
    
    // Check if content capture is enabled
    __u32 capture_key = 0;
    __u32 *enabled = bpf_map_lookup_elem(&content_capture_enabled, &capture_key);
    
    if (!enabled || !*enabled || count == 0) {
        // Emit event without content
        emit_write_chunk(ctx, pid, fd, buf, count, offset, 0, 0, 1, path_hash);
        return 0;
    }
    
    // Calculate total chunks needed
    __u32 total_chunks = (count + MAX_CONTENT_LEN - 1) / MAX_CONTENT_LEN;
    if (total_chunks > MAX_CHUNKS) {
        total_chunks = MAX_CHUNKS;
    }
    
    // Emit chunks with content
    #pragma unroll
    for (__u32 chunk = 0; chunk < MAX_CHUNKS; chunk++) {
        if (chunk >= total_chunks) break;
        
        __u64 chunk_offset = chunk * MAX_CONTENT_LEN;
        __u32 chunk_size = count - chunk_offset;
        if (chunk_size > MAX_CONTENT_LEN) {
            chunk_size = MAX_CONTENT_LEN;
        }
        if (chunk_size == 0) break;
        
        emit_write_chunk(ctx, pid, fd, buf + chunk_offset, count, offset + chunk_offset,
                        chunk_size, chunk, total_chunks, path_hash);
    }
    
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
        struct event *e = bpf_ringbuf_reserve(&events, sizeof(struct event), 0);
        if (e) {
            e->pid = pid;
            e->fd = fd;
            e->type = 3; // close
            e->path_hash = *path_hash;
            e->write_size = 0;
            e->file_offset = 0;
            e->content_len = 0;
            e->chunk_seq = 0;
            e->total_chunks = 1;
            bpf_get_current_comm(e->comm, sizeof(e->comm));
            
            bpf_ringbuf_submit(e, 0);
        }
    }
    
    // Clean up our tracking regardless
    bpf_map_delete_elem(&fd_to_hash, &fd_key);
    
    return 0;
}

char _license[] SEC("license") = "GPL";
