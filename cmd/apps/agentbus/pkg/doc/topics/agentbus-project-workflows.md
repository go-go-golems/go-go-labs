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
- announce
- await
- satisfy
Flags:
- agent
- channel
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
2. Agree on _communication channels_ (`speak` / `overhear`)
3. Define the _coordination flags_ (`announce` / `await` / `satisfy`)
4. Show how to _store & retrieve knowledge_ (`jot` / `recall`)
5. Walk through an **end-to-end workflow script** so you can copy-paste and adapt.

**Storing large knowledge snippets:** If your knowledge data is sizable or non-transient—think diagrams, lengthy specs, protocol files—commit the data to a file and set the jot’s `--value` to that **file path**. Agents can later do `cat $(agentbus recall --key <snippet-key>)` to retrieve the content without bloating Redis.

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

### Channels
* `status` – high-level progress
* `design` – architecture discussions
* `code` – granular coding updates
* `test` – QA results
* `deploy` – environment changes
* `alerts` – build / security failures

### Coordination Flags
| Flag | Meaning |
|------|---------|
| `pos-design` | Architecture in progress |
| `pos-backend-build` | Backend compiling |
| `pos-frontend-build` | Frontend build |
| `pos-tests` | End-to-end tests running |
| `pos-release` | Deployment lock |

### Knowledge Snippets
```bash
# Architecture diagrams (stored as file reference)
agentbus jot --key "pos-service-diagram" --value "diagrams/services.puml" --tag "pos,design"

# Coding conventions
agentbus jot --key "pos-go-style" --value "https://github.com/uber-go/guide" --tag "pos,backend,style"
```

### End-to-End Workflow
```bash
# 1. Architecture Agent kicks off
AGENT_ID=pos-arch \
agentbus announce pos-design && \
agentbus speak --channel design --msg "Creating service boundaries"

# 2. Backend & Frontend wait for design
AGENT_ID=pos-backend agentbus await pos-design --timeout 1800
AGENT_ID=pos-frontend agentbus await pos-design --timeout 1800

# 3. Backend builds in parallel with Frontend
AGENT_ID=pos-backend \
agentbus announce pos-backend-build && \
agentbus speak --channel code --msg "Implementing gRPC inventory service"

AGENT_ID=pos-frontend \
agentbus announce pos-frontend-build && \
agentbus speak --channel code --msg "Building HTMX cash-drawer UI"

# 4. QA waits for both builds
AGENT_ID=pos-qa \
agentbus await pos-backend-build && \
agentbus await pos-frontend-build && \
agentbus announce pos-tests && \
agentbus speak --channel test --msg "Running playwright suite"

# 5. Ops waits for green tests before deploying
AGENT_ID=pos-ops \
agentbus await pos-tests --timeout 1200 && \
agentbus announce pos-release && \
agentbus speak --channel deploy --msg "Deploying v1.3.0 to staging"

# 6. Release flag satisfied after deploy
AGENT_ID=pos-ops agentbus satisfy pos-release
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

### Channels
* `firmware` – code pushes
* `simulation` – HIL run status
* `safety` – risk findings
* `docs` – manuals
* `alerts` – watchdog failures

### Coordination Flags
| Flag | Purpose |
|------|---------|
| `robot-pinmap` | GPIO assignments frozen |
| `robot-motion-build` | Motor layer compiling |
| `robot-assay-tests` | Protocol regression tests |
| `robot-safety-signoff` | Safety review approved |

### Workflow Highlights
1. `robot-arch` publishes the pin map as a jot and `satisfy robot-pinmap` once done.
2. `robot-motion` and `robot-assay` `await robot-pinmap`, then `announce` their respective build flags so they can proceed in parallel.
3. `robot-hil` waits for both build flags, runs simulation suites, `satisfy robot-assay-tests`.
4. `robot-safe` `await robot-assay-tests`, inspects hazardous moves, then `satisfy robot-safety-signoff`.
5. Firmware is flashed only when `robot-safety-signoff` is satisfied.

```bash
# Example safety-officer check
AGENT_ID=robot-safe \
agentbus await robot-assay-tests && \
agentbus speak --channel safety --msg "Reviewing HIL logs" && \
agentbus jot --key "robot-safety-report-$(date +%F)" --value "$(cat report.md)" --tag "robot,safety" && \
agentbus satisfy robot-safety-signoff
```

---

## 3. Web-Scrape Orchestrator

A data-engineering team scrapes hundreds of websites daily. We will coordinate micro-agents that each own part of the ETL.

### Agent Roles
| Role | AGENT_ID | Function |
|------|----------|----------|
| Scheduler | `scrape-sched` | Cron & queue management |
| Scraper Bots (×N) | `scraper-*` | Fetch & parse HTML |
| Parser Agent | `scrape-parse` | Normalise data |
| Storage Agent | `scrape-store` | Writes to S3/DB |
| Quality Agent | `scrape-qa` | Validates schema |
| Monitoring Agent | `scrape-mon` | Alerts & dashboards |

### Channels
* `queue` – job assignments
* `scrape` – per-site status
* `parse` – parsing pipeline
* `storage` – load status
* `alerts` – extraction errors

### Coordination Pattern
Instead of long-lived flags we use **short-lived per-job flags** for each batch ID so shards can run truly in parallel.

```bash
# Scheduler publishes a new batch id
BATCH_ID=$(date +%s)
AGENT_ID=scrape-sched \
agentbus announce "batch-${BATCH_ID}" && \
agentbus speak --channel queue --msg "New batch ${BATCH_ID} published"

# Each scraper bot waits for the batch flag and then releases its own flag when done
for SITE in $(cat sites.txt); do
  AGENT_ID="scraper-${SITE}" \
  agentbus await "batch-${BATCH_ID}" --timeout 3600 && \
  agentbus speak --channel scrape --msg "${SITE} scraping started" && \
  # ... scrape ...
  agentbus speak --channel scrape --msg "${SITE} scraping finished" && \
  agentbus satisfy "scrape-${BATCH_ID}-${SITE}"
done

# Parser waits for all site flags (with wildcard support via script)
AGENT_ID=scrape-parse ./await_all_site_flags.sh ${BATCH_ID}

# Storage waits for parser, then loads
AGENT_ID=scrape-store agentbus await "parse-${BATCH_ID}" && ...
```

`jot` is used to keep reusable XPath/CSS selector snippets tagged by domain name so scrapers can self-service:
```bash
agentbus jot --key "selector-example.com-price" --value "div.price > span" --tag "scraper,example.com"
agentbus recall --tag "scraper,example.com"
```

---

# Best-Practice Checklist

* **Namespace everything** with project prefix (`pos-`, `robot-`, `scrape-`) to avoid collisions.
* **Always set reasonable `--timeout`** on `await`.
* **Release flags** with `satisfy` even on failure (use traps or CI post-steps).
* **Log human-readable messages** with `speak` so dashboards are helpful.
* **Store conventions & diagrams** in `jot` for on-boarding new agents quickly.
* **Automate clean-up** of expired flags and old chat streams via cron.

---

# Next Steps
1. Copy one of the examples and run it locally with `docker compose up redis && agentbus ...`.
2. Expand agent lists or split further (e.g., reviewer agents, security scanning bots).
3. Integrate with CI systems by wrapping `agentbus` calls inside your build pipelines.
4. Share your own patterns back via PR to this documentation! 