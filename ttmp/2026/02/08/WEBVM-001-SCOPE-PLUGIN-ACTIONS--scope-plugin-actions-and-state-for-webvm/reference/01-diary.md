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
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: Preset capabilities and dispatchSharedAction migration (commit 709df40)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts
      Note: Identity contract migration to packageId/instanceId (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts
      Note: VM map and lifecycle now keyed by instanceId (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts
      Note: Client API updated to send packageId+instanceId (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/runtimeIdentity.ts
      Note: Central instance ID generation utility (commit 96c6225)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: |-
        Multi-instance load/render/unload orchestration (commit 96c6225)
        Per-instance shared dispatch and filtered global projection (commit 709df40)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: |-
        Instance-keyed state and package-based reducer routing (commit 96c6225)
        Shared-domain state model
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/workers/quickjsRuntime.worker.ts
      Note: Worker request routing updated for instance identity (commit 414b68a)
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/tests/e2e/quickjs-runtime.spec.ts
      Note: Added multi-instance and capability-enforcement e2e cases (commit 709df40)
    - Path: ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/04-phase-3-4-design-brief-multi-instance-identity-and-capability-model.md
      Note: Source-of-truth design spec implemented in this diary
    - Path: ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/tasks.md
      Note: Step-by-step execution checklist for this implementation
ExternalSources: []
Summary: Implementation diary for WEBVM-001, with commit-by-commit notes, failures, and validation instructions.
LastUpdated: 2026-02-09T06:08:24Z
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

## Step 4: Shared Domains, Capability Grants, and API Migration

I implemented the Phase 4 capability model end-to-end: plugin intents can now target shared domains, per-instance grants are stored in runtime state, and shared writes are denied when grants are missing. I also migrated preset plugins to `dispatchSharedAction(...)` and updated tests to cover multi-instance behavior and capability enforcement.

This step replaces the previous flat global model with a governed shared-domain projection while preserving host-controlled runtime metadata for observability.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete the remaining implementation phases with commit-by-commit execution and diary tracking.

**Inferred user intent:** Finish WEBVM-001 with real capability-governed shared state and verification coverage.

**Commit (code):** 709df40 — "feat(webvm-001): add shared-domain capability model and migrate presets"

### What I did

- Updated transport/runtime intent model:
  - `DispatchIntent.scope` migrated from `"plugin" | "global"` to `"plugin" | "shared"`.
  - Added `domain` field for shared intents (`quickjsContracts.ts`).
  - Updated VM bootstrap to expose `dispatchSharedAction(domain, actionType, payload)` and keep `dispatchGlobalAction(...)` as a compatibility alias to `legacy-global`.
- Updated intent validation:
  - Shared intents now require non-empty `domain` (`dispatchIntent.ts`).
  - Expanded unit tests for shared intent validation (`dispatchIntent.test.ts`).
- Replaced runtime store model (`store.ts`):
  - Added per-instance `CapabilityGrants` storage.
  - Added shared domain state for `counter-summary` and `greeter-profile`.
  - Added dispatch outcome tracing (`applied`/`denied`/`ignored` + reason).
  - Added package reducer fallback for custom plugins (`state/replace`, `state/merge`).
  - Added `dispatchSharedAction(...)` helper and `selectGlobalStateForInstance(...)` filtered projection.
- Updated Playground orchestration (`Playground.tsx`):
  - Registers per-instance grants from preset manifests.
  - Uses per-instance filtered global state for render/event.
  - Dispatches shared intents through `dispatchSharedAction`.
- Migrated presets (`presetPlugins.ts`):
  - Added capability manifests per preset.
  - Migrated counter/greeter to shared domain writes.
  - Migrated monitor/dashboard/shared-state viewer to domain-based reads.
- Expanded tests:
  - Added runtime integration test for loading same package with two instance IDs.
  - Added e2e tests for multi-instance counter independence and shared write denial for custom plugin without grants.

### Why

- Phase 4 requires explicit capability boundaries and per-instance data visibility rules.
- Preset migration and tests ensure the new API is exercised through real user flows, not only type-level changes.

### What worked

- Typecheck passed: `pnpm check`.
- Unit tests passed: `pnpm test:unit`.
- Integration tests passed: `pnpm test:integration`.
- E2E tests passed: `pnpm test:e2e` (4/4 passing after adding new tests).

### What didn't work

- First e2e run failed in sandbox due local webserver bind restrictions:
  - Command: `pnpm test:e2e`
  - Error: `Error: Process from config.webServer was not able to start. Exit code: 1`
  - Follow-up direct command showed: `Error: listen EPERM: operation not permitted 127.0.0.1:4173`
- Resolved by rerunning e2e with escalated permissions.

### What I learned

- The capability model is easiest to reason about when read filtering (`selectGlobalStateForInstance`) and write enforcement (`sharedActionDispatched`) both live in store-level host logic.

### What was tricky to build

- The sharp edge was sequencing API migration without breaking custom/legacy plugin code. I handled this by adding `dispatchSharedAction` as the primary path while keeping a controlled `dispatchGlobalAction` alias that maps into normalized shared actions.

### What warrants a second pair of eyes

- `legacy-global` alias behavior and denial semantics for non-migrated custom plugins should be explicitly accepted (or removed quickly) to avoid long-lived ambiguity.
- Shared domain schemas are currently runtime-validated ad hoc; if domain count grows, central schema validators should be added.

### What should be done in the future

- Remove `dispatchGlobalAction` alias after confirming no remaining plugin code depends on it.
- Add dedicated reducer-level unit tests for grant denial reasons and domain reducer edge cases.

### Code review instructions

- Start with `client/src/store/store.ts` (core model and enforcement).
- Then review `client/src/lib/presetPlugins.ts` + `client/src/pages/Playground.tsx` for host manifest wiring and plugin API migration.
- Finally review `client/src/lib/quickjsRuntimeService.ts` and tests.
- Validate with:
  - `pnpm check`
  - `pnpm test:unit`
  - `pnpm test:integration`
  - `pnpm test:e2e`

### Technical details

- Shared write denial is traced as `missing-write-grant:<domain>` and does not throw into the UI event loop.
- Per-instance filtered global context shape now includes:
  - `self: { instanceId, packageId }`
  - `shared: { ...grantedDomainsOnly }`
  - `system: runtime metrics + registry`

## Step 5: Prune Unused UI Wrapper Components

I removed dead wrapper files under `client/src/components/ui` and kept only the wrappers currently used by the app. This shrinks maintenance surface area and sets up a cleaner dependency-pruning pass next.

This change is intentionally constrained to file removals and validation; no runtime behavior or API paths were changed.

### Prompt Context

**User prompt (verbatim):** "yes, go ahead."

**Assistant interpretation:** Continue the cleanup pass by removing unused wrapper components before dependency pruning.

**Inferred user intent:** Keep the codebase lean and reduce stale abstraction layers that are no longer used.

**Commit (code):** d38e329 — `refactor(playground): remove unused UI wrapper components`

### What I did

- Audited direct imports from app code and identified active wrappers:
  - `button`, `input`, `badge`, `card`, `sonner`, `tooltip`
  - plus `dialog` (required by `input` via `useDialogComposition`)
- Deleted all other wrapper files in `client/src/components/ui`.
- Re-ran static and automated checks.

### Why

- Most shadcn wrapper files were no longer referenced by the runtime/playground UI.
- Keeping dead wrappers creates confusion and inflates future migration work.

### What worked

- Typecheck passed: `pnpm check`.
- Unit tests passed: `pnpm test:unit`.
- Integration tests passed: `pnpm test:integration`.

### What didn't work

- N/A

### What I learned

- The only non-obvious retained dependency was `dialog.tsx`, required indirectly by `input.tsx`.

### What was tricky to build

- Ensuring no hidden imports existed outside `client/src/components/ui` before bulk deleting wrappers.

### What warrants a second pair of eyes

- Whether `input.tsx` should keep `useDialogComposition` coupling or be simplified in a later pass.

### What should be done in the future

- Prune now-unused UI dependencies from `package.json` and rerun checks/tests.

### Code review instructions

- Confirm only these wrapper files remain:
  - `client/src/components/ui/button.tsx`
  - `client/src/components/ui/input.tsx`
  - `client/src/components/ui/badge.tsx`
  - `client/src/components/ui/card.tsx`
  - `client/src/components/ui/sonner.tsx`
  - `client/src/components/ui/tooltip.tsx`
  - `client/src/components/ui/dialog.tsx`
- Validate with:
  - `pnpm check`
  - `pnpm test:unit`
  - `pnpm test:integration`

## Step 6: Prune Unused NPM Dependencies

I removed stale runtime and UI library dependencies that became dead after the component cleanup and runtime simplification. The package manifest now tracks only the libraries still referenced by source code, CSS imports, and active scripts.

This step keeps behavior unchanged and focuses on reducing dependency surface and long-term maintenance overhead.

### Prompt Context

**User prompt (verbatim):** "yes, go ahead."

**Assistant interpretation:** Continue cleanup after dead-component removal by trimming unused dependencies.

**Inferred user intent:** Keep the playground lean and avoid carrying unused transitive risk.

**Commit (code):** `chore(playground): prune unused npm dependencies`

### What I did

- Removed unused runtime dependencies from `package.json`, including:
  - legacy Radix wrappers no longer present (`accordion`, `alert-dialog`, `avatar`, etc.)
  - old form/editor/chart stacks (`react-hook-form`, `@hookform/resolvers`, `monaco`, `recharts`, `zod`)
  - unused utility/UI libs (`axios`, `cmdk`, `embla-carousel-react`, `framer-motion`, `vaul`, `streamdown`)
- Removed unused dev dependencies:
  - `@tailwindcss/typography`
  - `@types/google.maps`
  - `add`
  - `autoprefixer`
  - `postcss`
  - `pnpm`
  - `tsx`
- Updated `pnpm-lock.yaml` via `pnpm remove`.

### Why

- Unused packages increase install time, vulnerability surface, and upgrade burden.
- The current runtime/playground path uses a much smaller subset than the historical manifest.

### What worked

- Typecheck passed: `pnpm check`.
- Unit tests passed: `pnpm test:unit`.
- Integration tests passed: `pnpm test:integration`.
- Build passed: `pnpm build`.

### What didn't work

- N/A (build produced warnings only; no failures).

### What I learned

- The active dependency set is now mostly QuickJS runtime, Redux runtime, minimal Radix primitives, and routing/toast/theme support.

### What was tricky to build

- Distinguishing true runtime requirements from old package history and deleted wrapper references.

### What warrants a second pair of eyes

- `vite-plugin-jsx-loc` currently reports a peer warning with Vite 7 (`expects ^4 || ^5`), which predates this cleanup but remains visible.

### What should be done in the future

- Decide whether to keep or replace `vite-plugin-jsx-loc` to eliminate peer-version drift warnings.

### Code review instructions

- Review `cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json` for removed dependencies.
- Confirm lockfile regeneration in `cmd/experiments/2026-02-08--simulated-communication/plugin-playground/pnpm-lock.yaml`.
- Validate with:
  - `pnpm check`
  - `pnpm test:unit`
  - `pnpm test:integration`
  - `pnpm build`
