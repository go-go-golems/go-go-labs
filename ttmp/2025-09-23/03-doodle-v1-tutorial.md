## Doodle-v1 YAML DSL Tutorial (SQLite + CLI)

This tutorial shows how to run the doodle-clone CLI against a YAML actions file, verify results in SQLite, and iterate.

### Prereqs
- Go 1.21+
- SQLite CLI (`sqlite3`) recommended for verification

### 1) Quick check the CLI is available
```bash
go run ./go-go-labs/cmd/experiments/2025-09-23/doodle-clone --help
```

### 2) Create a demo YAML
Save to `go-go-labs/ttmp/2025-09-23/doodle-demo.yaml`:

```yaml
version: "doodle-v1"
actions:
  - id: poll1
    action: create_poll
    use_tz: "Europe/Berlin"
    title: "Team sync"
    participants:
      - email: "a@example.com"
      - email: "b@example.com"
    duration: "45m"
    candidate_windows:
      - start: "2025-09-25T09:00"
        end:   "2025-09-25T17:00"
      - start: "2025-09-26T09:00"
        end:   "2025-09-26T17:00"
    strategy: "approval"
    quorum: 1
    notes: "Remote"

  - id: slots1
    action: add_slots
    poll_ref: "poll1"
    slots:
      - start: "2025-09-25T10:00"
        end:   "2025-09-25T10:45"
      - start: "2025-09-26T14:00"
        end:   "2025-09-26T14:45"

  - id: votes1
    action: vote_slot
    poll_ref: "poll1"
    votes:
      - slot_ref: "2025-09-25T10:00"
        vote: yes
        email: "a@example.com"
      - slot_ref: "2025-09-26T14:00"
        vote: maybe
        email: "b@example.com"

  - id: finalize
    action: finalize_poll
    poll_ref: "poll1"
```

### 3) Apply the actions (fresh DB) and view output
```bash
rm -f go-go-labs/ttmp/2025-09-23/doodle.db && \
go run ./go-go-labs/cmd/experiments/2025-09-23/doodle-clone apply \
  -f go-go-labs/ttmp/2025-09-23/doodle-demo.yaml \
  --db go-go-labs/ttmp/2025-09-23/doodle.db \
  -v
```

Expected: a summary listing each action and a created event after finalize.

### 4) Verify results in SQLite
```bash
sqlite3 go-go-labs/ttmp/2025-09-23/doodle.db "SELECT id,title,status,event_id FROM polls;"
sqlite3 go-go-labs/ttmp/2025-09-23/doodle.db "SELECT id,poll_id,start_ts,end_ts FROM slots ORDER BY start_ts;"
sqlite3 go-go-labs/ttmp/2025-09-23/doodle.db "SELECT id,slot_id,email,vote FROM votes ORDER BY created_at;"
sqlite3 go-go-labs/ttmp/2025-09-23/doodle.db "SELECT id,title,start_ts,end_ts FROM events;"
```

You should see:
- One poll with status=finalized and a non-null `event_id`
- Two slots
- Two votes
- One event covering the winning slot

### 5) Optional: Propose times
Save to `go-go-labs/ttmp/2025-09-23/doodle-propose.yaml`:
```yaml
version: "doodle-v1"
actions:
  - id: propose
    action: propose_times
    use_tz: "Europe/Berlin"
    duration: "30m"
    candidate_windows:
      - start: "2025-09-27T09:00"
        end:   "2025-09-27T17:00"
    max_candidates: 6
```

Run:
```bash
go run ./go-go-labs/cmd/experiments/2025-09-23/doodle-clone apply \
  -f go-go-labs/ttmp/2025-09-23/doodle-propose.yaml \
  --db go-go-labs/ttmp/2025-09-23/doodle.db \
  -v
```

Expected: printed JSON candidates; no writes to DB.

### Notes
- Naive times are resolved via `use_tz` and stored in UTC.
- `slot_ref` may be a slot id or the slot start timestamp.


