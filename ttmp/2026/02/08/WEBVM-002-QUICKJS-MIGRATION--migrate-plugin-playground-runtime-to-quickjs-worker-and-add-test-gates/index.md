---
Title: Migrate Plugin Playground Runtime to QuickJS Worker and Add Test Gates
Ticket: WEBVM-002-QUICKJS-MIGRATION
Status: active
Topics:
    - architecture
    - plugin
    - state-management
    - testing
    - quickjs
    - playwright
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/pluginManager.ts
      Note: Legacy in-process runtime scheduled for removal
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: Main runtime orchestration entrypoint for cutover
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: Host dispatch pipeline and scoped action behavior
    - Path: ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/design-doc/01-quickjs-migration-implementation-guide-and-test-strategy.md
      Note: Primary migration runbook
    - Path: ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/reference/01-diary.md
      Note: Planning and troubleshooting diary
ExternalSources: []
Summary: Execution ticket for implementing QuickJS worker isolation and adding unit/integration/Playwright test gates for plugin-playground.
LastUpdated: 2026-02-08T19:05:00-05:00
WhatFor: Track concrete migration execution from in-process plugin runtime to QuickJS worker runtime.
WhenToUse: Use as the landing page for WEBVM-002 deliverables, tasks, and progress updates.
---


# Migrate Plugin Playground Runtime to QuickJS Worker and Add Test Gates

## Overview

This ticket contains the concrete implementation path to migrate plugin execution from main-thread `new Function(...)` to a dedicated QuickJS worker runtime, while preserving the current plugin/global state/action contract model. It also defines the test gates required for safe rollout, including Playwright e2e scenarios.

## Key Links

- Parent architecture rationale: `../WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/02-quickjs-isolation-architecture-and-mock-runtime-removal-plan.md`
- Implementation guide: `design-doc/01-quickjs-migration-implementation-guide-and-test-strategy.md`
- Diary: `reference/01-diary.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`
- reMarkable bundle: `/ai/2026/02/08/WEBVM-002-QUICKJS-MIGRATION/WEBVM-002-quickjs-migration-guide-and-diary`

## Status

Current status: **active**

Latest outcome:

- Created a dedicated WEBVM-002 execution ticket.
- Authored a detailed migration and test strategy guide.
- Added a step-by-step implementation task breakdown.
- Added a detailed diary with research findings, command failures, and design rationale.

## Topics

- architecture
- plugin
- state-management
- testing
- quickjs
- playwright

## Tasks

See [tasks.md](./tasks.md) for the full migration checklist.

## Changelog

See [changelog.md](./changelog.md) for chronological updates.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
