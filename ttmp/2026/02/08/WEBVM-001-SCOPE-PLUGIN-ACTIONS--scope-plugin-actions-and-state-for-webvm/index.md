---
Title: Scope Plugin Actions and State for WebVM
Ticket: WEBVM-001-SCOPE-PLUGIN-ACTIONS
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/WidgetRenderer.tsx
      Note: UINode tree to React component renderer
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/pluginManager.ts
      Note: In-process plugin execution engine
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiTypes.ts
      Note: Canonical UINode type contract
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: Main playground page orchestrating plugin lifecycle
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: Redux store and plugin action routing
    - Path: ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/02-quickjs-isolation-architecture-and-mock-runtime-removal-plan.md
      Note: Architecture rationale with reality-check and handoff
    - Path: ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/index.md
      Note: Execution follow-on ticket for implementation and test gates
ExternalSources: []
Summary: Landing page for WEBVM-001 with links to the simplified v1 plugin scoping model and QuickJS isolation/removal design documents.
LastUpdated: 2026-02-08T13:26:00-05:00
WhatFor: Track the plugin identity/action/state scoping architecture investigation and implementation strategy.
WhenToUse: Use as the landing page for WEBVM-001 design docs, decisions, and deliverables.
---



# Scope Plugin Actions and State for WebVM

## Overview

This ticket investigates plugin identity, action scoping, and state scoping in the `plugin-playground` system. The current decision is a simplified v1 API (`selectPluginState`, `selectGlobalState`, plugin/global dispatch actions with `dispatchId`) plus a second design focused on real QuickJS isolation and removal of mock runtime paths.

## Key Links

- Design doc 01 (state/action scoping): `design-doc/01-plugin-action-and-state-scoping-architecture-review.md`
- Design doc 02 (QuickJS isolation): `design-doc/02-quickjs-isolation-architecture-and-mock-runtime-removal-plan.md`
- Design doc 03 (vision explainer): `design-doc/03-webvm-plugin-playground-vision-and-architecture-explainer.md`
- Design doc 04 (Phase 3-4 brief): `design-doc/04-phase-3-4-design-brief-multi-instance-identity-and-capability-model.md`
- Follow-on execution ticket (WEBVM-002): `../../WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/index.md`
- reMarkable upload (doc 01): `/ai/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS/01-plugin-action-and-state-scoping-architecture-review`
- reMarkable upload (bundle docs 01+02): `/ai/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS/WEBVM-001-scoping-and-quickjs-review`
- reMarkable upload (doc 03): `/ai/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS/03-webvm-plugin-playground-vision-and-architecture-explainer`
- reMarkable upload (WEBVM-002 bundle incl. updated doc 02 + execution guide/diary/tasks): `/ai/2026/02/08/WEBVM-002-QUICKJS-MIGRATION/WEBVM-002-quickjs-migration-guide-and-diary`
- Changelog: `changelog.md`
- Tasks: `tasks.md`

## Status

Current status: **active**

Latest outcome:

- Completed a detailed architecture assessment and migration plan.
- Updated the assessment with the simplified v1 selector/action model.
- Added a dedicated QuickJS isolation and mock-runtime removal plan.
- Added a comprehensive vision and architecture explainer (doc 03) for newcomers.
- Uploaded all docs to reMarkable.

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
