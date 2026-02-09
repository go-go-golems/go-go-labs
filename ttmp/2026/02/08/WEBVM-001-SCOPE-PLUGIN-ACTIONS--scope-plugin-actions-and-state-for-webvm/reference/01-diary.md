---
Title: Diary
Ticket: WEBVM-001-SCOPE-PLUGIN-ACTIONS
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts
      Note: Identity contract migration to packageId/instanceId (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts
      Note: VM map and lifecycle now keyed by instanceId (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts
      Note: Client API updated to send packageId+instanceId (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/runtimeIdentity.ts
      Note: Central instance ID generation utility (commit 96c6225)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: Multi-instance load/render/unload orchestration (commit 96c6225)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: Instance-keyed state and package-based reducer routing (commit 96c6225)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/workers/quickjsRuntime.worker.ts
      Note: Worker request routing updated for instance identity (commit 414b68a)
    - Path: ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/04-phase-3-4-design-brief-multi-instance-identity-and-capability-model.md
      Note: Source-of-truth design spec implemented in this diary
    - Path: ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/tasks.md
      Note: Step-by-step execution checklist for this implementation
ExternalSources: []
Summary: Implementation diary for WEBVM-001, with commit-by-commit notes, failures, and validation instructions.
LastUpdated: 2026-02-09T00:00:00Z
WhatFor: Track implementation progress from design brief to running multi-instance + capability model runtime.
WhenToUse: Read during development and review to understand what changed in each step and how to validate it.
---



# Diary

## Goal

Capture the implementation of WEBVM-001 as small, reviewable steps with explicit commands, outcomes, and risks.

## Step 1: Task Setup and Execution Baseline

I converted the ticket from a placeholder task list to a concrete implementation checklist and created this diary so every upcoming code step is documented in context. I also kept the new Phase 3-4 design brief as the active spec for execution.

This step intentionally focuses on scaffolding and traceability before code refactors start. The outcome is a clear work queue and a structured diary artifact that can be updated after every commit.

### Prompt Context

**User prompt (verbatim):** "Ok, add tasks to the ticket, and then implement step by step, committing at each step, and keeping a frequent diary."

**Assistant interpretation:** Add actionable tasks to WEBVM-001, then execute implementation incrementally with frequent diary updates and per-step commits.

**Inferred user intent:** Ensure implementation progress is disciplined, auditable, and easy to review.

**Commit (code):** N/A — documentation setup step.

### What I did

- Added concrete tasks to `tasks.md` via `docmgr task add`.
- Created a dedicated ticket diary document with `docmgr doc add`.
- Initialized diary structure and step logging format.

### Why

- The ticket needed an execution checklist before starting refactors.
- Frequent diary updates require a stable document and section structure from the start.

### What worked

- `docmgr` task and document operations succeeded.
- Task list now reflects the implementation order from the design brief.

### What didn't work

- N/A

### What I learned

- The ticket already had partial updates in `index.md`/`changelog.md`, so commits should stay scoped to intentional changes.

### What was tricky to build

- Keeping the diary format strict enough for future step-by-step updates while not overfitting to one commit pattern.

### What warrants a second pair of eyes

- Task sequencing alignment with the Phase 3/4 design brief before deep refactors begin.

### What should be done in the future

- Append a new diary step after each implementation commit with exact command/test evidence.

### Code review instructions

- Start with: `ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/tasks.md`.
- Then review: `ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/reference/01-diary.md`.

### Technical details

- Commands used:
  - `docmgr task add --ticket WEBVM-001-SCOPE-PLUGIN-ACTIONS --text "..."`
  - `docmgr doc add --ticket WEBVM-001-SCOPE-PLUGIN-ACTIONS --doc-type reference --title "Diary"`

## Step 2: Contract and Runtime Identity Migration

I migrated the QuickJS boundary contracts and runtime internals from `pluginId` identity to explicit `packageId` plus `instanceId`, with runtime operations keyed by `instanceId`. This establishes the foundation for true multi-instance support while keeping current behavior stable.

I kept this step scoped to transport/runtime identity plumbing and associated tests so later refactors (store routing and capability domains) can build on a stable API boundary.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement WEBVM-001 incrementally, committing each slice and recording the details.

**Inferred user intent:** Move from design to running code with minimal-risk, reviewable increments.

**Commit (code):** 414b68a — "feat(webvm-001): migrate quickjs contracts and runtime to instance identity"

### What I did

- Updated `quickjsContracts.ts`:
  - Added `PackageId` and `InstanceId` aliases.
  - Replaced request/response fields from `pluginId` to `instanceId`, and `loadPlugin` now carries both `packageId` + `instanceId`.
  - Updated `LoadedPlugin` to carry `{ packageId, instanceId }`.
  - Updated `DispatchIntent` plugin attribution field to `instanceId`.
- Updated `quickjsRuntimeService.ts`:
  - VM map keyed by `instanceId`.
  - `loadPlugin(packageId, instanceId, code)` now checks for duplicate instance IDs instead of replacing existing runtimes by key.
  - Runtime metadata validation now returns `{ packageId, instanceId }`.
- Updated worker/client plumbing (`quickjsRuntime.worker.ts`, `quickjsSandboxClient.ts`) to use new request fields.
- Updated `dispatchIntent.ts` + `dispatchIntent.test.ts` to stamp/expect `instanceId`.
- Updated integration tests in `quickjsRuntimeService.integration.test.ts` for new load signature and intent attribution.
- Updated `Playground.tsx` load/register calls to use new `LoadedPlugin` fields and the new sandbox client signature.

### Why

- Identity split is a prerequisite for loading multiple instances of the same package without key collisions.
- Moving this boundary first isolates downstream reducer/UI changes from transport-layer churn.

### What worked

- Typecheck passed: `pnpm check`.
- Unit tests passed: `pnpm test:unit`.
- Integration tests passed: `pnpm test:integration`.

### What didn't work

- Initial `git commit` attempt failed inside sandbox due worktree lock path permissions:
  - Command: `git add ... && git commit ...`
  - Error: `fatal: Unable to create '/home/manuel/code/wesen/corporate-headquarters/go-go-labs/.git/worktrees/go-go-labs39/index.lock': Permission denied`
- Resolved by rerunning commit command with escalated permissions.

### What I learned

- This repo’s `.git` worktree metadata sits outside the writable sandbox path, so commit operations need elevated execution in this environment.

### What was tricky to build

- The main edge was avoiding accidental behavior changes while replacing identity fields across contracts, runtime service, worker, and tests in one slice. I constrained the step to boundary-layer identity changes and verified with both unit and integration suites before committing.

### What warrants a second pair of eyes

- `QuickJSRuntimeService.loadPlugin(...)` duplicate-instance behavior change (now throws on existing `instanceId`) should be explicitly validated against expected UX before multi-instance UI lands.

### What should be done in the future

- Implement store-level package-based reducer routing and true multi-instance ID generation (`packageId@...`) in the next step.

### Code review instructions

- Start with `client/src/lib/quickjsContracts.ts` for type contract changes.
- Then inspect `client/src/lib/quickjsRuntimeService.ts`, `client/src/workers/quickjsRuntime.worker.ts`, and `client/src/lib/quickjsSandboxClient.ts` for boundary propagation.
- Validate with:
  - `pnpm check`
  - `pnpm test:unit`
  - `pnpm test:integration`

### Technical details

- Key API shift:
  - `loadPlugin(packageId, instanceId, code)`
  - `render(instanceId, ...)`
  - `event(instanceId, ...)`
  - `disposePlugin(instanceId)`

## Step 3: Store Routing and Multi-Instance UI

I completed the Phase 3 behavioral shift: Redux local-state routing is now package-based per instance, and the Playground now creates unique instance IDs when loading presets/custom plugins. This removes the old single-instance key collision behavior for repeated preset loads.

I kept legacy global action semantics intact in this step so Phase 4 (shared-domain capability model) can be introduced as an isolated follow-up change.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Continue implementation in commit-sized slices with diary updates and progress bookkeeping.

**Inferred user intent:** Reach full Phase 3/4 implementation with clear incremental checkpoints.

**Commit (code):** 96c6225 — "feat(webvm-001): add instance-based store routing and multi-instance playground"

### What I did

- Added `runtimeIdentity.ts` with `createInstanceId(packageId)`.
- Refactored `store.ts`:
  - `RuntimePlugin` now stores `{ instanceId, packageId, ... }`.
  - Plugin/local-state maps are keyed by instance ID.
  - `pluginActionDispatched` now carries `instanceId`.
  - Local reducer routing uses `packageId` from the instance registry instead of string matching on ID.
  - Added mirror recompute helpers for `counter`/`greeter` when removing an instance.
  - Updated selectors and dispatch helper signatures to instance-centric parameters.
- Refactored `Playground.tsx`:
  - Generates unique IDs per load (`counter@...`, etc.).
  - Calls runtime load with `(packageId, instanceId, code)`.
  - Registers/removes plugins by instance ID + package ID.
  - Displays preset load counts and loaded entries by instance ID.
  - Renders and dispatches widget events per instance.

### Why

- Multi-instance support requires unique runtime identities at load time and package-aware reducer dispatching.
- Store routing by package resolves the old bug where non-literal IDs (e.g., `counter@abc`) failed reducer matching.

### What worked

- Typecheck passed: `pnpm check`.
- Unit tests passed: `pnpm test:unit`.
- Integration tests passed: `pnpm test:integration`.

### What didn't work

- N/A

### What I learned

- The existing `selectLoadedPluginIds` shape was reusable for instance IDs, so the migration stayed contained without broad selector API churn.

### What was tricky to build

- Global mirror cleanup with multi-instance removal needed explicit recompute logic; naive reset-on-remove would incorrectly zero shared mirrors while sibling instances still exist.

### What warrants a second pair of eyes

- Mirror semantics (`counterValue`, `greeterName`) currently use last-instance heuristics, which are transitional until Phase 4 shared domains fully replace these fields.

### What should be done in the future

- Implement shared domain registry + capability grants and migrate plugin API to domain-scoped shared dispatch.

### Code review instructions

- Review `client/src/store/store.ts` first (routing and state shape).
- Then review `client/src/pages/Playground.tsx` for load/render/event/unload lifecycle updates.
- Validate with:
  - `pnpm check`
  - `pnpm test:unit`
  - `pnpm test:integration`

### Technical details

- Instance ID format: `${packageId}@${nanoid(8)}`.
- Plugin routing switch now targets `plugin.packageId` rather than instance key string literals.
