---
Title: Developer UI Overhaul, Reusable VM Package, and Documentation
Ticket: WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Deep-pass ticket to redesign the plugin playground for developer workflows, extract reusable VM runtime packages, and produce production-quality docs.
LastUpdated: 2026-02-09T00:23:20.806201038-05:00
WhatFor: Plan and execute the next architectural step after WEBVM-001 completion.
WhenToUse: Use when prioritizing UI overhaul, package extraction, and docs work for the plugin runtime.
---

# Developer UI Overhaul, Reusable VM Package, and Documentation

## Overview

This ticket covers three linked goals:

- Rework the playground into a developer-first workbench (debuggability, observability, instance management).
- Extract the VM runtime and contracts into reusable packages for use outside this app.
- Write explicit architecture/authoring/integration docs so the system can be adopted without oral context.

## Key Links

- Deep pass design document: `design-doc/01-deep-pass-ui-overhaul-runtime-packaging-and-docs-plan.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **active**

## Topics

- architecture
- plugin
- state-management

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
