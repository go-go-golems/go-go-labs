---
Title: Multi-Agent Project Workflows with AgentBus
Slug: project-workflows
Short: End-to-end examples of coordinating multi-agent software projects with AgentBus
Topics:
- coordination
- agents
- redis
- workflows
- examples
Commands:
- speak
- overhear
- jot
- recall
- list
- announce
- await
- satisfy
Flags:
- agent
- topic
- timeout
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

# Introduction

This guide presents **three concrete, real-world software projects** and shows how you can use **AgentBus** to coordinate dozens of specialised coding agents that build, test, document and deploy them **in parallel**.

For each project we will:

1. Define the _agent roles_ involved
2. Use the _single communication channel_ with topics (`speak` / `overhear`)
3. Define the _coordination flags_ (`announce` / `await` / `satisfy`)
4. Show how to _store & retrieve knowledge_ (`jot` / `recall` / `list`)
5. Walk through an **end-to-end workflow script** so you can copy-paste and adapt.

**Key improvements in this version:**
- **Single shared channel**: All agents communicate through one Redis stream with optional topic filtering
- **Auto-published coordination events**: `announce` and `satisfy` automatically publish coordination updates with 🚩 and ✅ emojis
- **Enhanced metadata**: Agent ID and timestamp are automatically included in all messages
- **Knowledge discovery**: Use `list` command to discover existing knowledge snippets by tag
- **Improved overhear**: Tracks read position per agent, shows new vs. seen messages

**Storing large knowledge snippets:** If your knowledge data is sizable or non-transient—think diagrams, lengthy specs, protocol files—commit the data to a file and set the jot's `--value` to that **file path**. Agents can later do `cat $(agentbus recall --key <snippet-key>)` to retrieve the content without bloating Redis.

---

## 1. Retail Pawn-Shop POS System

A small chain of pawn shops needs a modern Point-Of-Sale application covering inventory, customer management and regulatory reporting.

### Agent Roles
| Role | AGENT_ID | Purpose |
|------|----------|---------|
| Product Owner | `pos-po` | Owns requirements & priorities |
| Architecture Agent | `pos-arch` | Splits system into micro-services |
| Backend Agent | `pos-backend` | Golang gRPC services |
| Frontend Agent | `pos-frontend` | React/HTMX UI |
| QA Agent | `pos-qa` | Test plans & playwright scripts |
| Docs Agent | `pos-docs` | End-user manual |
| DevOps Agent | `pos-ops` | CI/CD & Helm charts |
| Security Agent | `pos-sec` | Dependency scanning |

### Communication Topics
All agents use the single shared channel with these topic categories:
* `status` – high-level progress updates
* `design` – architecture discussions
* `code` – granular coding updates
* `test` – QA results and coverage reports
* `deploy` – environment changes and releases
* `alerts` – build failures, security issues
* `coordination` – auto-published announce/satisfy events

### Coordination Flags
| Flag | Meaning |
|------|---------|
| `pos-design` | Architecture in progress |
| `pos-backend-build` | Backend compiling |
| `pos-frontend-build` | Frontend build |
| `pos-tests` | End-to-end tests running |
| `pos-release` | Deployment lock |

### Knowledge Management
```bash
# Architecture diagrams (stored as file reference)
agentbus jot --key "pos-service-diagram" --value "diagrams/services.puml" --tag "pos,design"

# Coding conventions
agentbus jot --key "pos-go-style" --value "https://github.com/uber-go/guide" --tag "pos,backend,style"

# Discover available docs by tag
agentbus list --tag "pos,design" --latest 10
agentbus list --tag "backend" --latest 5
```

### End-to-End Workflow
```bash
# 1. Architecture Agent kicks off (auto-publishes 🚩 coordination event)
AGENT_ID=pos-arch \
agentbus announce --flag pos-design && \
agentbus speak --topic design --msg "Creating service boundaries"

# 2. Backend & Frontend monitor for design completion and wait
AGENT_ID=pos-backend agentbus await --flag pos-design --timeout 1800
AGENT_ID=pos-frontend agentbus await --flag pos-design --timeout 1800

# 3. Backend builds in parallel with Frontend (both auto-publish 🚩 events)
AGENT_ID=pos-backend \
agentbus announce --flag pos-backend-build && \
agentbus speak --topic code --msg "Implementing gRPC inventory service"

AGENT_ID=pos-frontend \
agentbus announce --flag pos-frontend-build && \
agentbus speak --topic code --msg "Building HTMX cash-drawer UI"

# 4. QA monitors all build activity and waits for completion
AGENT_ID=pos-qa \
agentbus overhear --topic coordination --follow &  # Monitor coordination events
AGENT_ID=pos-qa \
agentbus await --flag pos-backend-build && \
agentbus await --flag pos-frontend-build && \
agentbus announce --flag pos-tests && \
agentbus speak --topic test --msg "Running playwright suite"

# 5. Ops waits for green tests before deploying
AGENT_ID=pos-ops \
agentbus await --flag pos-tests --timeout 1200 && \
agentbus announce --flag pos-release && \
agentbus speak --topic deploy --msg "Deploying v1.3.0 to staging"

# 6. Release flag satisfied after deploy (auto-publishes ✅ coordination event)
AGENT_ID=pos-ops agentbus satisfy --flag pos-release

# 7. Check what knowledge was created during this workflow
agentbus list --tag "pos" --latest 20
```

---

## 2. Genomics Lab Robot Control Firmware

A research lab is updating the firmware of its pipetting robot to support new genetic assays.

### Agent Roles
| Role | AGENT_ID | Purpose |
|------|----------|---------|
| Firmware Architect | `robot-arch` | Define RTOS tasks |
| Motion Control Dev | `robot-motion` | Kinematics & stepper drivers |
| Assay Logic Dev | `robot-assay` | High-level protocol scripts |
| Hardware-In-Loop QA | `robot-hil` | Simulated hardware tests |
| Safety Officer | `robot-safe` | Reviews hazardous moves |
| Documentation Agent | `robot-docs` | Lab technician manuals |

### Communication Topics
* `firmware` – code pushes and compilation status
* `simulation` – HIL run status and results
* `safety` – risk findings and approvals
* `docs` – manual updates and API changes
* `alerts` – watchdog failures and critical errors
* `coordination` – auto-published workflow events

### Coordination Pattern
1. `robot-arch` publishes the pin map as a jot and `satisfy robot-pinmap` once done.
2. `robot-motion` and `robot-assay` `await robot-pinmap`, then `announce` their respective build flags so they can proceed in parallel.
3. `robot-hil` waits for both build flags, runs simulation suites, `satisfy robot-assay-tests`.
4. `robot-safe` `await robot-assay-tests`, inspects hazardous moves, then `satisfy robot-safety-signoff`.
5. Firmware is flashed only when `robot-safety-signoff` is satisfied.

```bash
# Example safety-officer workflow with enhanced discovery
AGENT_ID=robot-safe \
agentbus await --flag robot-assay-tests && \
agentbus speak --topic safety --msg "Reviewing HIL simulation logs" && \
agentbus list --tag "robot,simulation" --latest 5 && \  # Discover recent simulation data
agentbus jot --key "robot-safety-report-$(date +%F)" --value "$(cat safety-report.md)" --tag "robot,safety" && \
agentbus satisfy --flag robot-safety-signoff

# Monitor all robot coordination in real-time
AGENT_ID=robot-monitor agentbus overhear --topic coordination --follow

# Check recent robot documentation
agentbus list --tag "robot,docs" --latest 10
```

---

## 3. Web Research Orchestrator

Imagine a knowledge-gathering pipeline where you give the system a high-level TOPIC (e.g. *"battery recycling supply-chain"*). A **scheduler agent** seeds the initial task list, and a swarm of **researcher sub-agents** explore the web, summarise findings, and can create *new subtasks* (flags) whenever they uncover interesting sub-topics.

### Agent Roles
| Role | AGENT_ID | Function |
|------|----------|----------|
| Scheduler | `research-sched` | Creates root topic flag & monitors overall progress |
| Researcher Bots (×N) | `researcher-*` | Crawl web, extract key facts |
| Summariser | `research-sum` | Merges raw notes into concise bullet lists |
| Curator | `research-cur` | Filters low-quality sources, dedupes |
| Task Splitter | `research-split` | Reads notes & spawns new subtasks for uncovered sub-topics |
| Knowledge Librarian | `research-lib` | Stores evergreen insights as jots |

### Communication Topics
* `tasks` – new task announcements and assignments
* `research` – raw findings, URLs, and extracted data
* `summary` – curated summaries and insights
* `alerts` – scraping failures, rate-limit issues
* `coordination` – workflow state changes

### Coordination Pattern
1. **Scheduler** receives input TOPIC and `announce topic:<UUID>`.
2. Each **Researcher Bot** `await`s that flag, then `announce research:<SUBID>` for its own crawl and `satisfy research:<SUBID>` when done.
3. **Summariser** continuously `await`s batches of `research:*` flags, jots interim summaries, and `announce summary:<UUID>`.
4. **Task Splitter** reviews summaries, detects emergent sub-topics, and for each creates a new flag `announce topic:<NEWUUID>` — feeding back into step 2.
5. When no new tasks appear for *N* minutes, **Scheduler** marks the original topic flag satisfied, signalling completion.

```bash
# 1. User kicks off a topic (with auto-coordination publishing)
TOPIC="battery recycling supply-chain"
ROOT_ID=$(date +%s)
AGENT_ID=research-sched \
agentbus announce --flag "topic-${ROOT_ID}" && \
agentbus speak --topic tasks --msg "New research topic: ${TOPIC} (id=${ROOT_ID})"

# 2. Researchers discover and pick up available tasks
agentbus list --tag "research,active" --latest 20  # See what's currently being worked on

while read TASK; do
  SUBID=$(uuidgen)
  AGENT_ID="researcher-${SUBID}" \
  agentbus announce --flag "research-${SUBID}" && \
  agentbus speak --topic research --msg "Collecting sources for ${TASK}" && \
  # ... gather URLs, notes into notes/${SUBID}.md ...
  agentbus jot --key "notes-${SUBID}" --value "notes/${SUBID}.md" --tag "research,raw" && \
  agentbus satisfy --flag "research-${SUBID}"
done < <(agentbus overhear --topic tasks --since "1${ROOT_ID}-0")

# 3. Summariser processes finished research tasks
AGENT_ID=research-sum ./process_research_batch.sh "${ROOT_ID}"

# 4. Task Splitter spawns follow-up topics based on summaries
AGENT_ID=research-split ./spawn_new_topics.sh "${ROOT_ID}"

# 5. Monitor entire research pipeline in real-time
AGENT_ID=research-monitor agentbus overhear --topic coordination --follow
```

`jot` and `list` examples for knowledge management:
```bash
# Store an evergreen insight with proper tagging
agentbus jot --key "recycling-milestone-2025" --value "docs/milestones/2025.md" --tag "research,battery,forecast,milestone"

# Discover all battery-related research
agentbus list --tag "research,battery" --latest 50

# Find recent forecasting data
agentbus list --tag "forecast" --latest 10

# Recall specific insights
agentbus recall --tag "research,battery,forecast"
agentbus recall --key "recycling-milestone-2025"
```

---

# Best-Practice Checklist

* **Namespace everything** with project prefix (`pos-`, `robot-`, `research-`) to avoid collisions.
* **Use meaningful topics** for filtering communication (`speak --topic`, `overhear --topic`).
* **Always set reasonable `--timeout`** on `await`.
* **Release flags** with `satisfy` even on failure (use traps or CI post-steps).
* **Tag your jots** consistently for easy discovery with `list`.
* **Monitor coordination events** with `overhear --topic coordination --follow`.
* **Use the auto-published coordination messages** to track workflow progress.
* **Leverage agent ID and timestamp** metadata that's automatically included.
* **Check existing knowledge** with `list` before creating duplicate jots.
* **Automate clean-up** of expired flags and old chat streams via cron.

---

# New Features Summary

## Single Communication Channel
- All agents share one Redis stream instead of multiple channels
- Use `--topic` for categorization and filtering
- `overhear --topic coordination` to monitor workflow events

## Auto-Published Coordination Events
- `announce` automatically publishes 🚩 "Announced working on 'flag'" 
- `satisfy` automatically publishes ✅ "Satisfied 'flag'"
- Monitor these with `overhear --topic coordination`

## Enhanced Metadata
- Agent ID automatically included in all messages
- Timestamps automatically added to all communications
- Read position tracking per agent in `overhear`

## Knowledge Discovery
- `list` command to browse available knowledge snippets
- Filter by tag: `list --tag "docker,api"`
- Limit results: `list --latest 10`

## Improved overhear
- Shows "NEW:" prefix for unread messages
- Tracks read position per agent to avoid re-reading
- Summary header shows total vs. new message counts

---

# Next Steps
1. Copy one of the examples and run it locally with `docker compose up redis && agentbus ...`.
2. Expand agent lists or split further (e.g., reviewer agents, security scanning bots).
3. Integrate with CI systems by wrapping `agentbus` calls inside your build pipelines.
4. Use `overhear --topic coordination --follow` to monitor your workflows in real-time.
5. Share your own patterns back via PR to this documentation!
