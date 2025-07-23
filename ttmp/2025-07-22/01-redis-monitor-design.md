┌──────────────────────────────────────────────────────────────────────────────┐
│                       Redis Streams Monitor (top-like)                      │
│                       Uptime: 12d 5h 34m   Refresh: 5s                      │
├─────────┬──────────────┬──────────┬───────────────┬───────────┬──────────────┤
│ Stream  │   Entries    │   Size   │   Groups      │  Last ID  │  Memory RSS  │
├─────────┼──────────────┼──────────┼───────────────┼───────────┼──────────────┤
│ orders  │  1,243,592   │ 120.5MB  │  3            │ 160123-7  │  64MB        │
│         │ message/s: ▄▄▄▄▂▂▂▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ │           │
├─────────┼──────────────┼──────────┼───────────────┼───────────┼──────────────┤
│ events  │    98,234    │  9.2MB   │  5            │ 160123-3  │  12MB        │
│         │ message/s: ▁ ▁ ▁▁▂▂▃▄▄▇▇▇▆▅▅▄▄▂▂▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁ ▁▁ ▁ ▁ │           │
├─────────┼──────────────┼──────────┼───────────────┼───────────┼──────────────┤
│ logs    │  5,432,100   │ 512.1MB  │  1            │ 160122-9  │ 256MB        │
│         │ message/s: ▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▇▆▅▅▅▄▄▂▁ ▁ ▁ ▁ ▁ ▁│           │
└─────────┴──────────────┴──────────┴───────────────┴───────────┴──────────────┘

Groups Detail:
┌─────────┬─────────┬───────────────┬──────────┬──────────────┐
│ Group   │ Stream  │ Consumers     │ Pending  │ Idle Time    │
├─────────┼─────────┼───────────────┼──────────┼──────────────┤
│ cg-1    │ orders  │ Alice(3) Bob(2)│  12      │ 00:00:05     │
│ cg-2    │ orders  │ Charlie(5)    │  0       │ 00:00:01     │
│ cg-3    │ orders  │ Dave(1)       │  28      │ 00:02:12     │
│ cg-A    │ events  │ Eve(4) Frank(4)│  3       │ 00:00:00     │
│ cg-B    │ events  │ Grace(2)      │  47      │ 00:01:23     │
│ cg-logs │ logs    │ Heidi(10)     │  0       │ 00:00:00     │
└─────────┴─────────┴───────────────┴──────────┴──────────────┘

Trim / Memory Alerts:
  • orders: maxlen=500k (approx 1.2m entries) → trim rate: 50/s
  • events: maxlen=100k (98k entries) → within threshold
  • logs: no maxlen

Global Throughput ──▁▂▃▄▅▆▇▇▇▇▆▅▄▃▂▁   1023 msg/s
Memory Usage    ■■■■■■■■■■■■■■■■■■■■■■■  332MB / 1GB

Commands:
 [R]efresh  [G]roup filter  [S]ort  [Q]uit


---

```yaml
# Redis Streams Monitor – Command Specification

## 1. Stream Discovery
# — find all keys of type “stream”
- SCAN 0 TYPE stream MATCH *  

## 2. Server/Uptime
# — get server uptime in seconds (convert to days/hours/minutes)
- INFO SERVER  
  • parse field `uptime_in_seconds`

## 3. Per‑Stream Metrics
# For each stream key `<stream>`:

### 3.1 Entry Count
- XLEN <stream>  
  • returns total number of entries

### 3.2 Stream Details
- XINFO STREAM <stream>  
  • returns:
    – length  
    – radix-tree-keys / nodes  
    – groups count  
    – first-entry ID  
    – last-generated ID (→ “Last ID”)

### 3.3 Memory Usage (per key)
- MEMORY USAGE <stream>  
  • returns bytes allocated for this stream

## 4. Consumer‑Group Metrics
# For each `<stream>`:

### 4.1 List Groups
- XINFO GROUPS <stream>  
  • returns one record per group:
    – name  
    – consumers count  
    – pending messages count  

### 4.2 List Consumers (per group)
- XINFO CONSUMERS <stream> <group>  
  • returns one record per consumer:
    – consumer name  
    – pending messages  
    – idle time (ms)

## 5. Throughput Metrics
# — derive messages/sec by comparing XADD/XREADGROUP counts between intervals:

- INFO commandstats  
  • parse:
    – `cmdstat_xadd:calls`  
    – `cmdstat_xreadgroup:calls`  
  • Δ(calls) / Δ(time) → messages/sec sparkline

## 6. Global Memory Usage
- INFO MEMORY  
  • parse:
    – `used_memory` (total bytes)  
    – `used_memory_rss` (RSS bytes)  
    – `total_system_memory` (from INFO MEMORY or INFO)
  • compute “Memory Usage” bar

## 7. Trim / MaxLen Monitoring
# — detect maxlen overflows & trim events:

### 7.1 MaxLen Enforcement (configured externally)
# (the application’s `XTRIM … MAXLEN` setting is known; not read from Redis)

### 7.2 Trim Events via Keyspace Notifications
- CONFIG GET notify-keyspace-events  
  • ensure `E` and `x` flags include key‑event for XTRIM  
- PSUBSCRIBE __keyevent@<db>__:xtrim  
  • count incoming XTRIM events → trim rate

## 8. Commands Legend (for UI hotkeys)
- R → full refresh (re‑run above commands)  
- G → filter by group (re‑run XINFO GROUPS/XINFO CONSUMERS)  
- S → sort streams (by XLEN, memory, throughput)  
- Q → quit
```

