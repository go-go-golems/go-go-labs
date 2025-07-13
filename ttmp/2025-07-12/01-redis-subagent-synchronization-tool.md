Below is a **redis‑backed CLI surface** that coding sub‑agents can invoke with `shell_exec()` without risking accidental misuse.
The binary name is shown as `agentbus`, but you can rename it.

---

## 0. Global conventions

| Convention                         | Purpose                                                                                                                  |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `AGENT_ID` **(required)**          | Unique name for the calling sub‑agent; all state is namespaced with this key. Pass via `--agent <id>` or env `AGENT_ID`. |
| `PROJECT_PREFIX` **(required)**    | Project isolation prefix for Redis keys. Pass via env `PROJECT_PREFIX`. All keys use format `<PROJECT_PREFIX>:<suffix>`. |
| Output                             | Single UTF‑8 line per record, JSON by default (`--format text` for plain).                                               |
| Exit codes                         | `0` = success, `1` = transient redis/network error, `2` = usage error, `3` = timeout (for blocking verbs).               |
| Redis keys                         | Use project prefix plus logical sub‑keys so multiple projects can safely share one redis instance.                       |
| Debug logging                      | All operations logged to `/tmp/agentbus.log` for troubleshooting hanging or timeout issues.                              |

---

## 1. Realtime chat stream

Minimal wrapper around a Redis Stream so agents can "speak" and "overhear" each other.

```bash
agentbus speak   --channel <topic>  --msg "Unit tests green ✅"
agentbus overhear --channel <topic>  [--since <offset>|--follow]  [--max <n>]
```

| Verb         | Why this name?            | Behaviour                                                                                                                                                                                                                                                               |
| ------------ | ------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **speak**    | "Say something out loud." | `XADD` to `<PROJECT_PREFIX>:ch:<topic>` with the caller's `AGENT_ID`, timestamp, and message.                                                                                                                                                                           |
| **overhear** | "Passively receive."      | *Pull* model: <br>`--since <id>` ⇒ one‑shot read after the given redis stream ID (default: last ID the same agent read, stored in `<PROJECT_PREFIX>:last:<agent>:<topic>`).<br>`--follow` ⇒ block until new messages arrive, then print and exit (good for cron‑style polling). |

*Semantics* – No auto‑fan‑out; each agent holds its own offset, so no agent starves another.

---

## 2. Long‑lived knowledge snippets ("Docs / TIL")

Key/value store plus lightweight tagging.

```bash
agentbus jot   --key <title> --value "$(cat README.md)"   [--tag til,docs]
agentbus recall --tag til                           [--latest n]
agentbus recall --key <title>
```

| Verb       | Behaviour                                                                                                                                         |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| **jot**    | `HSET <PROJECT_PREFIX>:jot:<title>` stores the blob, author, timestamp, and comma‑separated tags. Existing key is overwritten unless `--if‑absent` given. |
| **recall** | Fetches one note (`--key`) or a reverse‑chronological list filtered by tag(s).                                                                    |

Why "jot/recall"? They are unambiguous, map to write/read, and differ from chat verbs.

---

## 3. Cross‑agent coordination flags

A tiny abstraction over redis keys that lets agents advertise progress, wait for dependencies, and mark them resolved.

```bash
# Coding agent declares it is working
agentbus announce building

# …later, marks done
agentbus satisfy  building

# Review agent blocks until coding agent has satisfied
agentbus await    building   [--timeout 900]
```

| Verb         | Behaviour                                                                                                                  |
| ------------ | -------------------------------------------------------------------------------------------------------------------------- |
| **announce** | `SETNX <PROJECT_PREFIX>:flag:<name> "<AGENT_ID> @ <timestamp>"` (returns error if flag already present unless `--force`). |
| **await**    | Polls/blocks until key exists. `--timeout` (sec) gives up with exit 3. Option `--delete` removes the flag after detection. |
| **satisfy**  | Deletes the key (`DEL`). Keeps your agents from leaving stale flags around.                                                |

The noun "flag" is implicit in the names; "await" and "satisfy" read naturally in English sentences, making incorrect usage stick out.

---

## 4. Utility commands

Additional commands for debugging, monitoring, and cleanup.

```bash
# Monitor real-time activity across all channels and flags
agentbus monitor  [--interval 2]

# List active channels, flags, and recent activity
agentbus list     [--channels|--flags|--agents]  [--max 10]

# Clean up all project data from Redis
agentbus clear    [--force]
```

| Verb        | Behaviour                                                                                                                    |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------- |
| **monitor** | Real-time display of new messages, flag changes, and agent activity. Press Ctrl+C to exit.                                  |
| **list**    | Shows active channels, flags, and agents. Useful for debugging coordination issues.                                         |
| **clear**   | Removes ALL project data from Redis (using `PROJECT_PREFIX`). Requires `--force` flag for safety.                          |

---

## 5. Putting it together (sample `shell_exec` calls)

```python
# publish a chat message
shell_exec("agentbus speak --channel build --msg 'compile step complete'")

# wait until build flag is satisfied (15‑minute cap)
shell_exec("agentbus await build --timeout 900")

# share a TIL after finishing
shell_exec("agentbus jot --key 'cli‑pattern' --value 'Use verb‑noun naming…' --tag til,docs")

# monitor for debugging coordination issues
shell_exec("agentbus monitor --interval 3")

# clean up after project completion
shell_exec("agentbus clear --force")
```

---

## 6. Environment setup

Before using agentbus, ensure these environment variables are set:

```bash
export AGENT_ID="my-coding-agent"
export PROJECT_PREFIX="myproject"
export REDIS_HOST="localhost"  # optional, defaults to localhost:6379
```

---

## 7. Internal redis schema cheat‑sheet

```
<PROJECT_PREFIX>:ch:<topic>             # Redis Stream of chat messages
<PROJECT_PREFIX>:last:<agent>:<topic>   # Last stream ID pulled by agent
<PROJECT_PREFIX>:jot:<title>            # Redis Hash {body,author,timestamp,tags}
<PROJECT_PREFIX>:jots_by_tag:<tag>      # Sorted‑set of jot keys by timestamp
<PROJECT_PREFIX>:flag:<name>            # Simple string = holder|timestamp
```

All verbs are single‑purpose, human‑readable, and grouped by mental model (chat, notes, flags).
Because each action's name encodes its intent, agents (and humans debugging them) are much less likely to call the wrong command.

The `PROJECT_PREFIX` isolation ensures multiple projects can safely share the same Redis instance without data conflicts.
Debug logging to `/tmp/agentbus.log` helps troubleshoot coordination issues, especially timeout problems with `announce` and `await` commands.
