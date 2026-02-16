
Here’s a tight, implementation-oriented sketch you can build from. I’ve kept it focused: core architecture, the AI-chat flow (as real tools), and a minimal, expressive YAML DSL for actions.

# System overview

* **Clients**: Web (React) + Mobile (optional). Realtime over WebSocket.
* **API**: Go/TS service exposing REST+WebSocket; serves auth, polls, proposals, scheduling actions.
* **Sync workers**: Per-provider connectors (Google Calendar, Outlook, CalDAV). Event-driven, idempotent.
* **Scheduler**: Computes candidate times, resolves constraints, finalizes events, writes back to calendars.
* **AI Orchestrator**: Chat endpoint that calls tool functions (the same REST actions) and emits YAML actions (below).
* **Storage**: PostgreSQL (+ Redis for jobs/locks).
* **Queue**: e.g., Redis Streams / RabbitMQ for sync + scheduler jobs.
* **Secrets/OAuth**: per-provider tokens (rotated), PKCE/OIDC for user login.

# Core data model (essentials)

* `users(id, name, email, tz, locale)`
* `calendars(id, user_id, provider, external_id, access_token, refresh_token, sync_state)`
* `events(id, organizer_id, calendar_id, title, description, location, start_ts, end_ts, status{draft,pending,final,deleted}, external_id)`
* `participants(id, event_id, email, role{organizer,required,optional}, response{unknown,yes,no,maybe})`
* `polls(id, event_id, strategy{rank,approval,first-fit}, quorum, deadline_ts, status)`
* `slots(id, poll_id, start_ts, end_ts, votes_yes, votes_no, votes_maybe)`
* `constraints(id, scope{user,event,poll}, kind{window,avoid,capacity,travel}, payload jsonb)`
* `sync_jobs(id, calendar_id, kind{pull,push}, cursor, status, retries, last_error)`
* `messages(id, thread_id, user_id, role{user,assistant,tool}, content, action_yaml jsonb)`
* `webhooks(id, provider, secret, last_seen_ts)` (for provider push)

# Sync model

* **Pull**: initial full sync (time-bounded window), then incremental via provider cursors; de-dupe by `(provider, external_id)`.
* **Push**: only for **finalized** events; drafts/polls live only in your DB.
* **Conflict resolution**: write-ahead record version; if provider changed externally, re-open poll or re-optimize.
* **Locking**: per `event_id` mutex for finalize/writeback.

# Scheduling/availability sketch

1. **Collect availability**: own calendars + optional free/busy pulls of invitees (when authorized) + constraints.
2. **Generate candidates**: sliding window over organizer’s preferred ranges; prune by hard constraints; score by soft prefs.
3. **Poll** (optional): create N slots; collect votes; early-stop when quorum satisfied.
4. **Finalize**: pick top feasible slot; push to organizer’s calendar and send ICS to others (and to provider if connected).

# API surface (selected)

* `POST /auth/provider/{google|microsoft|caldav}/init` → OAuth URL
* `POST /polls` `{ title, participants[], candidate_windows[], duration, strategy, quorum, deadline }`
* `POST /polls/{id}/slots` `{ slots[] }`
* `POST /polls/{id}/vote` `{ slot_id, vote{yes|no|maybe} }`
* `POST /polls/{id}/finalize`
* `POST /constraints` `{ scope, kind, payload }`
* `POST /events` `{ title, participants[], start_ts, end_ts, location }`
* `POST /events/{id}/reschedule` `{ candidate_windows[] | slot_id }`
* `POST /sync/run` `{ calendar_id, kind }`
* WebSocket: `subscribe: {event_id|poll_id}` → server pushes slot updates, votes, finalize.

# AI chat—tool calling (what the model can do)

Expose these **verifiable tools** (the assistant never writes directly; it *calls tools* that hit the API):

* `create_poll`, `add_slots`, `vote_slot`, `finalize_poll`
* `create_event`, `reschedule_event`, `cancel_event`
* `set_constraints`, `remove_constraints`
* `propose_times` (server runs candidate generator and returns ranked slots)
* `sync_now`

The assistant **also** emits/accepts the YAML DSL (below) so users/agents can compose multi-step changes in one message.

---

# YAML DSL for scheduling actions

## Design goals

* **Minimal**: actions are explicit and map 1:1 to API/tool calls.
* **Deterministic defaults**: server fills defaults when fields omitted (e.g., duration).
* **Staging**: multiple actions in a single doc; each action has an `id` so later steps can reference earlier outputs.

## Top-level schema

```yaml
version: "doodle-v1"
actions:
  - id: <string>            # optional; for cross references
    when: now|on_approve    # optional; default now
    use_tz: "America/New_York"  # optional; per-action override
    action: <OneOfBelow>
```

## Action variants

### 1) Create poll

```yaml
action: create_poll
title: "Team sync"
participants:
  - email: "a@example.com"     # role defaults to required
  - email: "b@example.com"
duration: "45m"
candidate_windows:             # organizer windows to search within
  - start: "2025-09-24T09:00"
    end:   "2025-09-24T17:00"
  - start: "2025-09-25T09:00"
    end:   "2025-09-25T17:00"
strategy: "approval"           # rank|approval|first-fit
quorum: 2
deadline: "2025-09-24T23:00"
notes: "Remote, Zoom"
constraints:
  - kind: window               # hard allow list (also supports avoid)
    scope: poll
    payload:
      weekdays: [Mon, Tue, Wed, Thu]
      earliest: "09:00"
      latest: "17:30"
  - kind: avoid
    scope: poll
    payload:
      dates: ["2025-09-24"]
```

### 2) Add explicit slots to poll

```yaml
action: add_slots
poll_ref: "<id|external>"      # id of poll or previous action id
slots:
  - start: "2025-09-24T10:00"
    end:   "2025-09-24T10:45"
  - start: "2025-09-25T14:00"
    end:   "2025-09-25T14:45"
```

### 3) Vote on a slot

```yaml
action: vote_slot
poll_ref: "<id>"
votes:
  - slot_ref: "<slot-id|iso-ts>"  # slot id or start time
    vote: yes|no|maybe
```

### 4) Finalize a poll

```yaml
action: finalize_poll
poll_ref: "<id>"
preferred_order: ["slot-123", "slot-456"]  # optional; overrides server ranking if feasible
```

### 5) Create an event directly (no poll)

```yaml
action: create_event
title: "Design review"
participants:
  - email: "c@example.com"
  - email: "d@example.com"
start: "2025-09-26T11:00"
end:   "2025-09-26T12:00"
location: "Zoom"
notes: "Focus on API"
calendar_ref: "primary"        # or specific connected calendar id
```

### 6) Propose times (server generates candidates)

```yaml
action: propose_times
duration: "30m"
candidate_windows:
  - start: "2025-09-27T09:00"
    end:   "2025-09-27T17:00"
participants:
  - email: "e@example.com"
include_freebusy: true          # if authorized
max_candidates: 8
```

### 7) Reschedule an existing event

```yaml
action: reschedule_event
event_ref: "<id|external>"
candidate_windows:
  - start: "2025-09-28T13:00"
    end:   "2025-09-28T18:00"
keep_participants: true
```

### 8) Set / remove constraints

```yaml
action: set_constraints
scope: user|event|poll
scope_ref: "<optional id>"      # required for event|poll
constraints:
  - kind: window
    payload:
      weekdays: [Mon, Tue, Wed, Thu, Fri]
      earliest: "10:00"
      latest: "16:00"
  - kind: avoid
    payload:
      locations: ["travel"]
```

```yaml
action: remove_constraints
constraint_ids: ["constr-1","constr-2"]
```

### 9) Trigger sync

```yaml
action: sync_now
calendar_ref: "primary"
direction: pull|push|both
```

## Batch example (multi-step, referenced)

```yaml
version: "doodle-v1"
actions:
  - id: poll1
    action: create_poll
    title: "Quarterly planning"
    participants:
      - email: "ops@example.com"
      - email: "pm@example.com"
    duration: "1h"
    candidate_windows:
      - start: "2025-09-29T09:00"
        end:   "2025-10-01T17:00"
    strategy: "rank"
    quorum: 2

  - id: slots1
    action: add_slots
    poll_ref: "poll1"
    slots:
      - start: "2025-09-30T10:00"
        end:   "2025-09-30T11:00"
      - start: "2025-10-01T15:00"
        end:   "2025-10-01T16:00"

  - id: finalize
    action: finalize_poll
    poll_ref: "poll1"
```

---

# AI chat loop (tool wiring)

1. User says: “find a 45m slot with Alice/Bob next Tue/Wed.”
2. NLU → extract participants, windows, duration.
3. Assistant emits **YAML** with `propose_times` (or `create_poll`+`add_slots`) and calls tool(s).
4. Server returns candidates; assistant summarizes and, on “yes”, emits `finalize_poll`.
5. On finalize, the **Scheduler** creates the event and the **Sync worker** writes it to the organizer’s calendar.

**Tools JSON (example)**

* `propose_times(params) → {candidates:[{start,end,score}]}`
* `create_poll(params) → {poll_id, slots[]?}`
* `finalize_poll({poll_id, slot_id}) → {event_id, external_id}`
* `create_event(params) → {event_id, external_id}`
* `set_constraints(params) → {constraint_ids[]}`

---

# Implementation notes (just the essentials)

* **Idempotency keys** on create/finalize to survive retries.
* **TZ handling**: store UTC + `original_tz`; DSL `use_tz` resolves naive times.
* **Free/busy**: cache short-lived; fall back to “unknown” for non-authorized invitees.
* **Webhooks**: register per provider to receive updates; re-sync on change notifications.

If you want, I can turn this into:

* an OpenAPI stub for the API,
* a JSON Schema for the YAML DSL,
* or a minimal Go service skeleton with the tool endpoints wired to the scheduler.
