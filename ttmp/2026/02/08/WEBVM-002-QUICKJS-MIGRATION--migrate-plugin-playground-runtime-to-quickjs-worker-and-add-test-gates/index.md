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
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/dispatchIntent.test.ts
      Note: Unit tests for dispatch intent validation
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/dispatchIntent.ts
      Note: Runtime intent validation
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts
      Note: Runtime message and result contracts
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.integration.test.ts
      Note: Integration coverage for runtime lifecycle and timeout handling
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts
      Note: Reusable runtime service extracted for integration testing
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts
      Note: Worker RPC client
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiSchema.test.ts
      Note: Unit tests for runtime UI schema validation
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiSchema.ts
      Note: Runtime UI tree validation
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: |-
        Main runtime orchestration entrypoint for cutover
        Main runtime orchestration cutover
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: Host dispatch pipeline and scoped action behavior
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/workers/quickjsRuntime.worker.ts
      Note: QuickJS runtime worker implementation
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json
      Note: Migration test scripts and Playwright dependency
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/playwright.config.ts
      Note: E2E harness configuration
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/tests/e2e/quickjs-runtime.spec.ts
      Note: Playwright E2E runtime assertions
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vitest.config.ts
      Note: Unit test runner config
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vitest.integration.config.ts
      Note: Integration test runner config
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
