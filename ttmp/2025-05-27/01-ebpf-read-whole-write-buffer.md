Below is a practical “recipe” for grabbing **every byte a process passes through `read(2)` or `write(2)`** with eBPF on a 6.8-series kernel, plus a discussion of why you sometimes need to *chunk* the data and how the new `bpf_*_dynptr()` helpers (merged for 6.9 but easy to cherry-pick) let you lift many of the old size limits.

---

### 1  Pick the right probe‐points

| syscall     | best attach-point                     | why?                                                                          |
| ----------- | ------------------------------------- | ----------------------------------------------------------------------------- |
| **`write`** | `tracepoint:syscalls:sys_enter_write` | buffer contents are still untouched; size is known                            |
| **`read`**  | `tracepoint:syscalls:sys_exit_read`   | only on exit do we know how many bytes the kernel copied into the user buffer |

`tracepoint:` is kinder to the verifier than kprobes and survives across minor kernel releases without symbol changes. Probe arguments are already flattened for you (see the `__data_loc` structs in `/sys/kernel/debug/tracing/events/syscalls` for the exact layout).

---

### 2  Recorder topology

```
          ┌──────────────────┐               ┌─────────────────────────────┐
          │  BPF prog #1     │               │  BPF prog #2                │
          │  tp/sys_enter_write────────────┐ │  tp/sys_exit_read           │
          │  (grabs buf & len)             │ │  (looks up buf len, copies)│
          └──────────────────┘             │ └──────────────┬──────────────┘
                │ map[tid] (buf, len)      │                │
                └──────────────────────────┘                │
                                                            ▼
                                                  ┌─────────────────┐  ring buffer
                                                  │ struct io_evt { │  → user land
                                                  │   u64 pid,fd;   │
                                                  │   u32 rw;       │  (perf/ring)
                                                  │   u32 sz;       │
                                                  │   char data[N]; │
                                                  └─────────────────┘
```

* A **per-PID hash map** lets the entry probe pass the user‐space buffer pointer to the exit probe.
* A **ring buffer** is the most memory-efficient way to stream variable-length events to user space. Perf buffers work too but cost more per-event overhead.

---

### 3  Hard eBPF limits you must respect

| Limit (6.8)                    | Consequence                                                       |
| ------------------------------ | ----------------------------------------------------------------- |
| 512 B stack                    | You can’t copy an arbitrarily large read into the BPF stack.      |
| < \~10 k insn/verifier         | Unbounded loops are rejected; long byte-by-byte copies need care. |
| Ring-buffer event size == 2 MB | Still not “infinite”; huge I/O must be chunked.                   |

So in vanilla 6.8 **capture at most N bytes per event** (4096 is conventional—Falco/Sysdig do exactly that).

From 6.9, new helpers let you stream directly from user memory into a dynptr backed by the ring buffer:

```c
struct bpf_dynptr dp;
if (!bpf_ringbuf_reserve_dynptr(&events, count, 0, &dp))
        return 0;
bpf_copy_from_user_dynptr(&dp, 0, count, buf, 0);
bpf_ringbuf_submit_dynptr(&dp, 0);
```

Patchset: “**bpf\_copy\_from\_user\_dynptr helper**” (Jan 2025) .
If you use this, the only real ceiling is the ring buffer page size you reserve.

---

### 4  Skeleton BPF program (libbpf, CO-RE)

```c
// io_capture.bpf.c
#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#define MAX_PER_EVENT 4096

struct io_key { u32 pid; u32 tid; };
struct { __uint(type, BPF_MAP_TYPE_HASH);
         __uint(max_entries, 10240);
         __type(key, struct io_key);
         __type(value, void *);
} buf_map SEC(".maps");

struct io_evt {
        u64 pid;
        u64 fd;
        u32 rw;   // 0 = write, 1 = read
        u32 size; // bytes captured in this chunk
        char data[MAX_PER_EVENT];
};

struct { __uint(type, BPF_MAP_TYPE_RINGBUF);
         __uint(max_entries, 1 << 24);   // 16 MiB
} events SEC(".maps");

/* ---------------- write ---------------- */
SEC("tp/syscalls/sys_enter_write")
int handle_enter_write(struct trace_event_raw_sys_enter *ctx)
{
        struct io_evt *evt;
        size_t len = ctx->args[2];
        const char *buf = (const char *)ctx->args[1];
        if (!len) return 0;

        len = len > MAX_PER_EVENT ? MAX_PER_EVENT : len;
        evt = bpf_ringbuf_reserve(&events, sizeof(*evt), 0);
        if (!evt) return 0;

        evt->pid  = bpf_get_current_pid_tgid() >> 32;
        evt->fd   = ctx->args[0];
        evt->rw   = 0;
        evt->size = len;
        bpf_probe_read_user(evt->data, len, buf);
        bpf_ringbuf_submit(evt, 0);
        return 0;
}

/* ---------------- read ---------------- */
SEC("tp/syscalls/sys_enter_read")
int save_read_buf(struct trace_event_raw_sys_enter *ctx)
{
        struct io_key k = { .pid = bpf_get_current_pid_tgid() >> 32,
                            .tid = bpf_get_current_pid_tgid() };
        void *ptr = (void *)ctx->args[1];
        bpf_map_update_elem(&buf_map, &k, &ptr, BPF_ANY);
        return 0;
}

SEC("tp/syscalls/sys_exit_read")
int handle_exit_read(struct trace_event_raw_sys_exit *ctx)
{
        ssize_t ret = ctx->ret;
        struct io_key k = { .pid = bpf_get_current_pid_tgid() >> 32,
                            .tid = bpf_get_current_pid_tgid() };
        void **pp = bpf_map_lookup_elem(&buf_map, &k);
        if (!pp) return 0;
        if (ret > 0) {
                size_t len = ret > MAX_PER_EVENT ? MAX_PER_EVENT : ret;
                struct io_evt *evt = bpf_ringbuf_reserve(&events,
                                                         sizeof(struct io_evt), 0);
                if (evt) {
                        evt->pid  = k.pid;
                        evt->fd   = ctx->args[0];   /* fd from _enter_, still in regs */
                        evt->rw   = 1;
                        evt->size = len;
                        bpf_probe_read_user(evt->data, len, *pp);
                        bpf_ringbuf_submit(evt, 0);
                }
        }
        bpf_map_delete_elem(&buf_map, &k);
        return 0;
}

char LICENSE[] SEC("license") = "GPL";
```

*Compile with `clang -O2 -g -target bpf ...` and load with the libbpf skeleton/`cilium/ebpf` in Go.*

The write-side fits in one probe; the read path uses an extra map to pass the user buffer between enter/exit phases.

---

### 5  User-space side (Go + `cilium/ebpf`)

```go
type IOEvt struct {
    Pid  uint64
    Fd   uint64
    RW   uint32
    Size uint32
    Data [4096]byte
}

rd, _ := ebpf.NewRingBufReader(objs.Events)
for {
    record, err := rd.Read()
    if err != nil { continue }
    var evt IOEvt
    binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &evt)
    dir := map[uint32]string{0: "WRITE", 1: "READ"}[evt.RW]
    fmt.Printf("[%d] %s fd=%d %q\n", evt.Pid, dir, evt.Fd,
               evt.Data[:evt.Size])
}
```

---

### 6  Capturing **truly huge** I/O (> 4 KiB)

1. **Chunk**: split `count` into ≤ `MAX_PER_EVENT` slices and emit several events.
   *Verifier*: wrap the copy in a fixed-unroll loop (`#pragma unroll`) with `MAX_CHUNKS` iterations (each event *may* early-return once bytes\_left==0).
2. **Dynptr helper (6.9)**: one `bpf_copy_from_user_dynptr()` handles the entire buffer in a single call; no manual loop, verifier accepts because the helper guarantees bounds.

---

### 7  Performance & safety knobs

* **Filter early**: restrict by PID/comm or fd type to avoid saturating the ring buffer.
* **Ring size**: allocate a power-of-two number of pages (`perf_event_attr::rb_page_count`) to prevent drops; user land must drain fast.
* **Security**: you need `CAP_BPF` + `CAP_SYS_ADMIN` or kernel lockdown disabled. Avoid shipping data off-box unencrypted—this is effectively a wire-tap.

---

### 8  Alternatives

* **LSM BPF** for mandatory logging (post-6.8 you can attach to `file_read_iter` and `file_write_iter` LSM hooks—avoids the extra map glue).
* **uprobes** on library wrappers (e.g., `SSL_write`) for decrypted data streams.
* **Existing tools**: Falco’s “I/O capture” and Datadog’s blog show production-grade patterns you can crib from.

---

#### TL;DR

1. Attach at `sys_enter_write` and `sys_exit_read`.
2. Copy ≤ 4 KiB with `bpf_probe_read_user()`, send via ring buffer.
3. For bigger payloads, chunk—or jump to 6.9’s `bpf_copy_from_user_dynptr`.
4. Drain events in user space and stitch chunks back together.

With those pieces you can faithfully record **every byte** that flows through the kernel’s read/write syscalls, bounded only by the ring-buffer memory you’re willing to dedicate.
