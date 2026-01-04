---
Title: Zigbee low-level ZNP/ZCL orchestrator guide
Ticket: ADD-ZIGBEE-CONTROL-001
Status: active
Topics:
    - zigbee
    - znp
    - zcl
    - ti-zstack
    - zbdongle-p
    - python
    - zigpy-znp
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/zigbee/diary.md
      Note: Operational diary for the Zigbee experiments
    - Path: cmd/experiments/zigbee/scripts
      Note: Runnable low-level Zigbee scripts (ZNP/ZDO/ZCL)
ExternalSources: []
Summary: ""
LastUpdated: 2026-01-04T14:34:31.650488933-05:00
WhatFor: ""
WhenToUse: ""
---


# Zigbee low-level ZNP/ZCL orchestrator guide

## Overview

This ticket captures an end-to-end “low-level Zigbee” onboarding path for developers building a coordinator/orchestrator on a **Sonoff ZBDongle-P** (TI Z-Stack) using **ZNP** over serial and **ZCL** for application control.

The primary artifact is the guide in `reference/01-zigbee-orchestrator-guide-zbdongle-p-znp-zcl.md`, which documents:

- Zigbee fundamentals (roles, addressing, endpoints/clusters, security)
- ZNP framing and the concrete SYS/ZDO/AF commands used
- ZCL framing and the concrete On/Off + Read Attributes flows used
- How to run and extend the repo scripts under `cmd/experiments/zigbee/scripts/`

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Main Guide**: [Zigbee Orchestrator Guide](./reference/01-zigbee-orchestrator-guide-zbdongle-p-znp-zcl.md)

## Status

Current status: **active**

## Topics

- zigbee
- znp
- zcl
- ti-zstack
- zbdongle-p
- python
- zigpy-znp

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
