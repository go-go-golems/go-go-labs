---
Title: 'Deep Pass Refresh: Current Codebase Audit and UI/Runtime/Docs Roadmap'
Ticket: WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/index.html
      Note: |-
        Stale analytics placeholders and template comment leftovers
        Template/analytics cleanup target
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/WidgetRenderer.tsx
      Note: |-
        Debug leftovers and weak typing in renderer path
        Debug leftovers and type hardening targets
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: Preset catalog currently coupled to Redux store types
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts
      Note: Shared intent typing and domain validation boundary
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts
      Note: |-
        Worker transport client with pending-request lifecycle gaps
        Request lifecycle hardening target
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: |-
        Current orchestration hotspot and primary UI-overhaul target
        Orchestration hotspot and UI workbench split boundary
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: |-
        Monolithic runtime policy + reducer + selector module
        Monolithic reducer/policy/selector module for runtime extraction
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vite.config.ts
      Note: Vite plugin composition and current peer-version warning context
ExternalSources: []
Summary: Current-state deep audit for WEBVM-003 with concrete removals, refactor plan, doc backlog, and developer-first UI roadmap using one runtime package plus one playground app.
LastUpdated: 2026-02-09T12:24:00Z
WhatFor: Convert WEBVM-003 from high-level intent to an implementation-ready cleanup and architecture roadmap grounded in current code.
WhenToUse: Read before implementing WEBVM-003 steps; use as source of truth for deprecations, docs scope, and UI/runtime restructuring.
---


# Deep Pass Refresh: Current Codebase Audit and UI/Runtime/Docs Roadmap

## Executive Summary

The codebase is now materially cleaner (dead components and most unused UI wrappers/dependencies were removed), but the runtime remains app-centric. The core issue is not correctness; tests pass. The issue is shape: orchestration, transport, policy, and developer UX are still tightly coupled to a single page component and a single Redux file.

This document updates WEBVM-003 to current reality and defines the next moves:
1. Finish cleanup of remaining confusing/deprecated artifacts.
2. Extract one reusable `plugin-runtime` package (with internal modules) while keeping one `plugin-playground` app.
3. Overhaul the UI into a developer workbench with timeline/inspector workflows.
4. Add first-class documentation for plugin authors and runtime embedders.

## Problem Statement

We need one system that simultaneously supports:
- A developer-focused playground UI for rapid plugin iteration.
- A reusable runtime package for non-UI hosts.
- Clear, stable documentation for authoring and embedding.

Current structure still blocks this:
- `Playground.tsx` owns too much runtime coordination.
- `store.ts` combines reducers, policy, shared projection, and selectors.
- Runtime contract/domain typing is partially weak and partially cast-based.
- Some stale/debug/template artifacts remain and confuse maintainers.

## Runtime and Architecture Map (Current)

Flow today:
1. UI loads preset/custom code in `Playground.tsx`.
2. `quickjsSandboxClient.ts` sends worker RPC requests.
3. `quickjsRuntime.worker.ts` delegates into `quickjsRuntimeService.ts`.
4. `quickjsRuntimeService.ts` executes plugin handlers/renders and returns intents/UI trees.
5. `store.ts` applies plugin/shared actions and builds per-instance global projections.
6. `Playground.tsx` re-renders all widgets after state changes.

Hotspots by size and responsibility:
- `client/src/store/store.ts` (~574 LOC): state + policy + selectors + helper dispatchers.
- `client/src/pages/Playground.tsx` (~353 LOC): UI + lifecycle + runtime orchestration.
- `client/src/lib/quickjsRuntimeService.ts` (~373 LOC): runtime engine + bootstrap + validation boundary.
- `client/src/lib/presetPlugins.ts` (~331 LOC): plugin catalog + capability metadata + embedded code strings.

## Findings and Cleanup Sketches

## 1) Runtime and Transport

### 1.1 Worker request lifecycle has no timeout/cancel path

Problem:
`QuickJSSandboxClient` stores pending requests but does not time out or cancel unresolved requests.

Where to look:
- `client/src/lib/quickjsSandboxClient.ts:39`
- `client/src/lib/quickjsSandboxClient.ts:71`
- `client/src/lib/quickjsSandboxClient.ts:151`

Example:
```ts
this.pending.set(id, { resolve, reject });
this.worker.postMessage(requestWithId);
```

Why it matters:
- Potential pending-promise buildup after worker stalls/crashes.
- Hard to reason about teardown and flaky e2e behavior.

Cleanup sketch:
```ts
postRequest(req, opts?: { timeoutMs?: number; signal?: AbortSignal }) {
  const timeout = setTimeout(() => reject(new Error("worker request timeout")), timeoutMs);
  signal?.addEventListener("abort", () => reject(new Error("worker request aborted")));
  // always clear pending + timeout in finally path
}
```

### 1.2 Playground render effect recomputes all trees on broad dependency changes

Problem:
`Playground.tsx` triggers full render loop whenever `rootState` changes, even when only unrelated fields changed.

Where to look:
- `client/src/pages/Playground.tsx:85`
- `client/src/pages/Playground.tsx:126`
- `client/src/pages/Playground.tsx:188`

Example:
```tsx
React.useEffect(() => { ... }, [loadedPlugins, pluginMetaById, pluginStateById, rootState]);
```

Why it matters:
- O(instances * widgets) render calls per broad state change.
- Harder to scale instance count and timeline-heavy debugging.

Cleanup sketch:
```ts
// move runtime orchestration into controller
controller.computeRenderInputs(instanceId)
controller.renderDirtyWidgets(instanceId)
// track dirty keys by action outcome and affected domains
```

## 2) State/Policy Boundaries

### 2.1 `store.ts` is still a monolith

Problem:
One file contains plugin reducers, shared reducers, grants policy, projection selectors, action preparation, and dispatch helper wrappers.

Where to look:
- `client/src/store/store.ts:129`
- `client/src/store/store.ts:303`
- `client/src/store/store.ts:331`
- `client/src/store/store.ts:523`

Example:
```ts
function reducePluginScopedAction(...) { ... }
function reduceSharedScopedAction(...) { ... }
export const selectGlobalState = createSelector(...)
```

Why it matters:
- Slows migration to reusable runtime package.
- Prevents focused unit tests for policy-only logic.
- Increases merge conflict probability.

Cleanup sketch:
```txt
packages/plugin-runtime/src/redux-adapter/
  state.ts
  reducers/pluginReducers.ts
  reducers/sharedReducers.ts
  policy/capabilityPolicy.ts
  selectors/projections.ts
  actions.ts
```

### 2.2 Shared-domain typing is weak across boundaries

Problem:
`DispatchIntent.domain` is `string`, then cast to `SharedDomainName` in UI event handling.

Where to look:
- `client/src/lib/quickjsContracts.ts:17`
- `client/src/lib/dispatchIntent.ts:33`
- `client/src/pages/Playground.tsx:215`

Example:
```tsx
dispatchSharedAction(dispatch, instanceId, intent.domain as SharedDomainName, ...)
```

Why it matters:
- Compile-time protection is bypassed at a critical policy boundary.
- Unknown domains degrade to ignored/denied behavior instead of early typed failure.

Cleanup sketch:
```ts
// plugin-runtime/contracts
export const sharedDomains = ["counter-summary", "greeter-profile", ...] as const;
export type SharedDomainName = (typeof sharedDomains)[number];

function parseSharedDomain(v: unknown): SharedDomainName {
  if (!sharedDomains.includes(v as SharedDomainName)) throw new Error("unsupported domain");
  return v as SharedDomainName;
}
```

### 2.3 Preset catalog is coupled to Redux-store-local types

Problem:
`presetPlugins.ts` imports `SharedDomainName` from `store.ts`.

Where to look:
- `client/src/lib/presetPlugins.ts:1`
- `client/src/store/store.ts:6`

Why it matters:
- Inverts ownership: runtime contract types depend on app store implementation.
- Blocks moving preset authoring/runtime contracts into reusable package cleanly.

Cleanup sketch:
```txt
packages/plugin-runtime/src/contracts/sharedDomains.ts  // canonical type
apps/plugin-playground/src/lib/presetPlugins.ts         // imports from runtime package
```

## 3) UI/DX Findings

### 3.1 Developer workbench capabilities are missing

Problem:
Current 3-column layout handles loading/rendering, but lacks timeline, shared-domain diff, and capability inspection workflows.

Where to look:
- `client/src/pages/Playground.tsx:236`
- `client/src/pages/Playground.tsx:290`

Why it matters:
- Debugging denied shared writes and intent behavior is slow.
- Multi-instance behavior is difficult to inspect at scale.

Cleanup sketch:
```txt
Left: Package/Instance Explorer (+ capability badges)
Center: Widget canvas + custom code editor + run controls
Right tabs: Intent Timeline | Shared Domains | Runtime Health
Bottom: Error stream + raw intent payload inspector
```

### 3.2 WidgetRenderer still contains debug leftovers and weak `any` typing

Problem:
Button clicks still emit console noise and write globals.

Where to look:
- `client/src/components/WidgetRenderer.tsx:73`
- `client/src/components/WidgetRenderer.tsx:74`
- `client/src/components/WidgetRenderer.tsx:12`

Example:
```ts
console.log("[WidgetRenderer] Button clicked:", label, onClick);
(window as any).__lastButtonClick = ...
```

Why it matters:
- Global side effects in production path.
- `any` typing masks event payload shape issues.

Cleanup sketch:
```ts
// remove debug writes
type EventPayload = Record<string, unknown> | string | number | boolean | null;
onEvent: (ref: UIEventRef, eventPayload?: EventPayload) => void;
```

### 3.3 Theme stack is inconsistent (custom theme provider + next-themes hook)

Problem:
App wraps with a custom `ThemeProvider`, while `sonner.tsx` reads theme from `next-themes` context.

Where to look:
- `client/src/App.tsx:26`
- `client/src/contexts/ThemeContext.tsx:19`
- `client/src/components/ui/sonner.tsx:1`

Why it matters:
- Two theme systems increase confusion and maintenance cost.
- Toast theming behavior can diverge from app theme intent.

Cleanup sketch:
- Choose one:
  - Replace custom provider with `next-themes` provider, or
  - Stop using `next-themes` in `sonner.tsx` and use local theme context only.

### 3.4 UI consistency mismatch remains in 404 page

Problem:
NotFound page uses a different visual language than the main “technical brutalism” playground.

Where to look:
- `client/src/pages/NotFound.tsx:14`

Why it matters:
- Visual inconsistency makes app feel stitched from templates.

Cleanup sketch:
- Re-theme NotFound using the same token system and type palette as the main app.

## 4) Deprecated/Confusing Artifacts Still to Remove

### 4.1 `index.html` still has stale template comments and placeholder analytics script

Where to look:
- `client/index.html:10`
- `client/index.html:22`

Why it matters:
- Build warnings continue.
- Placeholder script appears production-like but is not wired safely.

Cleanup sketch:
```html
<!-- remove stale comment block -->
<!-- inject analytics script only when env vars are defined -->
```

### 4.2 Unused hook/module leftovers

Where to look:
- `client/src/hooks/useMobile.tsx:5` (currently unused)
- `vite.config.ts:4` (`fs` import unused)

Why it matters:
- Small, but signals stale code pathways and weak hygiene.

Cleanup sketch:
- Remove unused hook/module imports or wire them intentionally with tests.

## 5) Documentation Backlog (What to Write and Why)

1. `cmd/experiments/2026-02-08--simulated-communication/plugin-playground/README.md`
- Why: first-run instructions and architecture map are currently missing.

2. `docs/architecture/runtime-flow.md`
- Why: explain event->intent->reducer->render flow with instance identity and capability gates.

3. `docs/architecture/capability-model.md`
- Why: define read/write grant semantics, denial outcomes, and domain contracts.

4. `docs/plugin-authoring/quickstart.md`
- Why: stable author contract for `definePlugin`, widgets, handlers, and dispatch APIs.

5. `docs/plugin-authoring/capabilities-reference.md`
- Why: one place to document each shared domain and action schema.

6. `docs/integration/runtime-package-usage.md`
- Why: show non-UI embedding patterns for hosts outside this app.

7. `docs/testing/strategy.md`
- Why: map unit/integration/e2e responsibilities and add-test checklist.

8. `docs/migration/changelog-vm-api.md`
- Why: breaking changes log for identity model and capability API evolution.

## Proposed Solution: One Runtime Package + One Playground App

## Package direction

```txt
packages/
  plugin-runtime/
    src/core/
      contracts.ts
      runtimeService.ts
      dispatchIntent.ts
      uiSchema.ts
      runtimeIdentity.ts
    src/worker/
      runtime.worker.ts
      sandboxClient.ts
    src/redux-adapter/
      state.ts
      reducers/
      selectors/
      policy/
    src/index.ts
apps/
  plugin-playground/
    client/
    server/
    tests/
```

Design decisions:
1. Keep exactly one runtime package and one app package.
2. Keep core runtime independent from React and Redux.
3. Keep Redux integration as internal runtime submodule (`redux-adapter`) instead of separate package sprawl.
4. Move canonical domain/action typing into runtime contracts.
5. Make the app consume runtime APIs; app stops owning runtime internals.

## Implementation Plan (Phased)

Phase A: Remaining cleanup and hygiene
1. Remove `WidgetRenderer` debug globals/logging and tighten event payload types.
2. Remove stale `index.html` comment block and guard analytics injection.
3. Resolve theme-provider split (single theme source of truth).
4. Remove unused `useIsMobile` and unused `fs` import in Vite config.

Phase B: Runtime extraction
1. Scaffold `packages/plugin-runtime`.
2. Move contracts + validators + runtime service into `core/`.
3. Move worker + client transport into `worker/`.
4. Move Redux-specific policy/reducers/selectors into `redux-adapter/`.
5. Update playground imports to runtime package APIs.

Phase C: Developer-workbench UI overhaul
1. Split `Playground` into feature modules (`catalog`, `workspace`, `inspector`).
2. Add Intent Timeline panel (filter by `instanceId`, `scope`, `domain`, `outcome`).
3. Add Shared Domain inspector with before/after snapshots.
4. Add instance capability badges and quick diagnostics.

Phase D: Documentation completion
1. Publish architecture docs.
2. Publish authoring + embedding docs.
3. Publish migration notes and troubleshooting runbook.

## Acceptance Criteria

1. Runtime package is consumable by at least one non-playground host module.
2. Playground imports runtime APIs instead of runtime internals from app-local files.
3. Remaining stale/debug artifacts identified in this doc are removed.
4. UI includes timeline + shared-domain inspector workflows.
5. Documentation set is sufficient for first-time plugin authoring and runtime embedding without oral transfer.

## Open Questions

1. Should capability grants for custom plugins remain deny-by-default, or support an explicit approval UI flow?
2. Should shared domain contracts become schema-validated (e.g., zod/valibot) at runtime boundaries?
3. Do we keep embedded preset code strings in app source, or move presets to versioned files loaded at runtime?

## References

- Prior baseline doc: `design-doc/01-deep-pass-ui-overhaul-runtime-packaging-and-docs-plan.md`
- WEBVM-001 implementation diary: `ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/reference/01-diary.md`
