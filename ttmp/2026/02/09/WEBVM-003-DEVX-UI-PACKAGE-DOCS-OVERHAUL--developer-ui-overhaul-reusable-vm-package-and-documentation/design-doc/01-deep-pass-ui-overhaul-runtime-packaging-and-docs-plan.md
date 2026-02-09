---
Title: 'Deep Pass: UI Overhaul, Runtime Packaging, and Docs Plan'
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
        Analytics placeholders causing runtime warnings; cleanup target
        Analytics placeholders and HTML cleanup target
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: Preset catalog and capability metadata currently coupled to app runtime
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts
      Note: |-
        QuickJS VM API and bootstrap boundary to extract into package
        QuickJS core extraction boundary
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts
      Note: |-
        Worker transport client; candidate adapter boundary for package reuse
        Worker transport adapter extraction boundary
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: |-
        Current orchestration hotspot and primary UI-overhaul target
        UI orchestration hotspot and developer-workbench redesign input
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: |-
        Monolithic runtime/store logic that should be split into reusable modules
        Monolithic runtime policy/reducer/selector module to split
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json
      Note: |-
        Dependency and component-template bloat to prune during packaging
        Dependency/template bloat evidence
ExternalSources: []
Summary: Deep codebase pass focused on developer-facing UX overhaul, extraction of a reusable VM runtime package, deprecation/removal candidates, and a concrete documentation plan.
LastUpdated: 2026-02-09T00:00:00Z
WhatFor: Guide the next implementation ticket(s) for UI redesign, runtime modularization, and docs quality uplift.
WhenToUse: Read before implementing WEBVM-003; use as checklist for cleanup, packaging, and DX documentation work.
---


# Deep Pass: UI Overhaul, Runtime Packaging, and Docs Plan

## Executive Summary

The current system works and tests pass, but the architecture is still app-centric rather than package-centric. Runtime orchestration, policy, view models, and UI concerns are tightly coupled inside the React app, making reuse in other contexts hard. The codebase also carries template/dead artifacts that increase cognitive load and dependency footprint.

Main recommendation:
1. Extract runtime engine + contracts + validators into reusable packages.
2. Reduce `Playground` to a composition layer over package APIs.
3. Redesign UI around developer workflows (load, inspect, trace, diff, debug) rather than current static three-panel layout.
4. Remove dead/template artifacts and publish explicit docs for plugin authors + runtime integrators.

## Problem Statement

We want one codebase that supports:
- A developer-focused playground UI for experimentation.
- A reusable VM runtime package that can run in other UIs/hosts.
- Clear docs for plugin authors, runtime embedders, and maintainers.

Current implementation does not separate these concerns enough.

## Runtime/Code Map (Current)

Core runtime flow:
- `Playground.tsx` loads plugin code, tracks local UI caches, computes per-instance global projection, renders widgets, routes events.
- `quickjsSandboxClient.ts` wraps worker RPC.
- `quickjsRuntime.worker.ts` delegates to `quickjsRuntimeService.ts`.
- `quickjsRuntimeService.ts` runs QuickJS VMs + bootstrap API.
- `store.ts` handles instance registry, local reducers, shared domains, grants, and projections.

Primary hotspots:
- `Playground.tsx` (orchestration complexity)
- `store.ts` (monolithic state+policy+projection logic)
- `presetPlugins.ts` (embedded code strings + capability metadata coupled to app package)

## Findings and Cleanup Sketches

## Messaging + Runtime Boundaries

### Issue 1: UI orchestrator is doing runtime coordinator work

Problem: `Playground.tsx` handles loading, policy mapping, render scheduling, event routing, and VM lifecycle in one component.

Where to look:
- `client/src/pages/Playground.tsx:39`
- `client/src/pages/Playground.tsx:85`
- `client/src/pages/Playground.tsx:188`

Example:
```tsx
const globalState = selectGlobalStateForInstance(rootState, instanceId);
const tree = await quickjsSandboxClient.render(instanceId, widgetId, pluginState, globalState);
...
const intents = await quickjsSandboxClient.event(...);
```

Why it matters:
- Hard to reuse runtime in non-React contexts.
- High chance of UI-driven regressions in runtime behavior.
- Makes testing orchestration logic difficult without full component tests.

Cleanup sketch:
```ts
// new app-facing controller package
interface RuntimeController {
  loadFromPreset(packageId: string): Promise<InstanceHandle>;
  renderAll(state: RuntimeSnapshot): Promise<RenderBatch>;
  dispatchEvent(input: EventInput): Promise<DispatchResult>;
}
```

### Issue 2: Worker client has unbounded pending request map and no timeout/cancel API

Problem: pending RPC promises can accumulate indefinitely if worker never responds.

Where to look:
- `client/src/lib/quickjsSandboxClient.ts:39`
- `client/src/lib/quickjsSandboxClient.ts:71`

Example:
```ts
private pending = new Map<number, PendingRequest>();
...
this.pending.set(id, { resolve, reject });
this.worker.postMessage(requestWithId);
```

Why it matters:
- Memory leak risk during runaway or crashy plugin sessions.
- Poor debuggability for flaky runtime behavior.

Cleanup sketch:
```ts
postRequest(req, { timeoutMs = 2000, signal })
  -> reject on timeout
  -> reject on AbortSignal
  -> clear pending entry in finally
```

## Runtime State + Policy

### Issue 3: `store.ts` is monolithic (state, reducers, domain policies, selectors, projections)

Problem: one file owns too many responsibilities.

Where to look:
- `client/src/store/store.ts:1`
- `client/src/store/store.ts:129`
- `client/src/store/store.ts:303`
- `client/src/store/store.ts:460`

Example:
```ts
function reducePluginScopedAction(...) { ... }
function reduceSharedScopedAction(...) { ... }
function buildSharedForInstance(...) { ... }
export function dispatchSharedAction(...) { ... }
```

Why it matters:
- Hard to unit-test policies independently.
- Difficult to extract runtime as a standalone package.
- Higher merge conflict rate for ongoing feature work.

Cleanup sketch:
```txt
client/src/runtime/
  state/types.ts
  reducers/localReducers.ts
  reducers/sharedReducers.ts
  policy/capabilities.ts
  selectors/projections.ts
  store.ts
```

### Issue 4: Shared domain model is partially typed and partially stringly-typed

Problem: domain names are typed union, but action payload contracts are mostly `unknown` + runtime casts.

Where to look:
- `client/src/store/store.ts:282`
- `client/src/store/store.ts:303`
- `client/src/lib/quickjsContracts.ts:12`

Example:
```ts
const value = Number((payload as { value?: unknown }).value ?? 0);
```

Why it matters:
- Weak compile-time guarantees.
- Runtime failures become behavior-level bugs instead of type errors.

Cleanup sketch:
```ts
type SharedAction =
  | { domain: "counter-summary"; actionType: "set-instance"; payload: { value: number } }
  | { domain: "greeter-profile"; actionType: "set-name"; payload: string };
```

## UI/UX and DX

### Issue 5: Current UI is functional but not optimized for developer workflows

Problem: no explicit runtime timeline, no per-instance capability inspector, no structured event trace panel, no clear package/instance grouping.

Where to look:
- `client/src/pages/Playground.tsx:236`
- `client/src/pages/Playground.tsx:255`
- `client/src/pages/Playground.tsx:290`

Example:
```tsx
<div className="grid grid-cols-1 lg:grid-cols-3 ...">
```

Why it matters:
- Developers debugging capability denials and state propagation have low observability.
- UI does not scale to many plugin instances.

Cleanup sketch:
```txt
A. Left rail: package catalog + templates + capability badges
B. Center: editor + run controls + current instance inspector
C. Right: tabs (Rendered UI | Dispatch Timeline | Shared Domains | VM Health)
D. Bottom tray: raw intents/errors/log stream
```

### Issue 6: Debug leftovers and visual inconsistency in production path

Problem:
- Debug global writes in renderer.
- NotFound page style mismatches app theme.
- Index HTML contains placeholder analytics script and stale comment block.

Where to look:
- `client/src/components/WidgetRenderer.tsx:73`
- `client/src/components/WidgetRenderer.tsx:74`
- `client/src/pages/NotFound.tsx:14`
- `client/index.html:10`
- `client/index.html:20`

Example:
```tsx
console.log("[WidgetRenderer] Button clicked:", label, onClick);
(window as any).__lastButtonClick = ...
```

Why it matters:
- Noise in console and global namespace.
- Inconsistent UX polish.
- Avoidable warning spam during e2e/dev runs.

Cleanup sketch:
- Remove debug globals/logs.
- Re-theme 404 to same design system.
- Gate analytics script injection by defined env vars.

## Deprecated/Confusing Artifacts to Remove

### Issue 7: Dead components and template baggage inflate maintenance surface

Problem: several components exist but are not referenced in app flow.

Where to look:
- `client/src/components/PluginEditor.tsx`
- `client/src/components/Map.tsx`
- `client/src/components/ManusDialog.tsx`
- `client/src/const.ts`
- `shared/const.ts`

Evidence:
- No imports outside definitions for above files.

Why it matters:
- Raises onboarding cost.
- Encourages accidental coupling to unrelated template code.
- Pulls in dependencies not needed for runtime playground.

Cleanup sketch:
- Remove dead feature files.
- Prune dependencies after removal.
- If some are future plans, move to `archive/` or separate spike branch with clear ownership.

### Issue 8: UI component library and dependency set is much larger than active usage

Problem: dozens of Radix/UI wrappers are present but unused by current app routes.

Where to look:
- `client/src/components/ui/*.tsx`
- `package.json:18`

Why it matters:
- Dependency drift and higher attack/upgrade surface.
- Slower installs and noisier updates.

Cleanup sketch:
- Keep only components referenced by current routes.
- Split reusable UI kit into optional package only if truly reused.

## Documentation Gaps and What to Write

No focused project README exists in `plugin-playground/` and several existing markdown files are stale (`test-log.md`, `DEBUG_FINDINGS.md`, and older ticket docs now referencing removed files).

Required docs:
1. `README.md` (root of plugin-playground)
- Why: fast onboarding for running, testing, and architecture orientation.

2. `docs/architecture/runtime-flow.md`
- Why: formal data flow from UI event -> VM intent -> store reducers -> render.

3. `docs/architecture/capability-model.md`
- Why: make grants/denials and shared domain semantics explicit.

4. `docs/plugin-authoring/quickstart.md`
- Why: plugin authors need a stable contract for `definePlugin`, render/handler context, and shared dispatch.

5. `docs/plugin-authoring/capabilities-reference.md`
- Why: list each shared domain, read/write semantics, payload schema, denial behavior.

6. `docs/integration/runtime-package-usage.md`
- Why: target deliverable for running VMs in other contexts (CLI, other UI, tests).

7. `docs/testing/strategy.md`
- Why: clarify responsibility boundaries across unit/integration/e2e and how to add new cases.

8. `docs/migration/changelog-vm-api.md`
- Why: record breaking changes like removal of `dispatchGlobalAction` alias and identity model shifts.

## Proposed Package Restructure (Reusable VM Runtime)

### Target package layout

```txt
packages/
  plugin-runtime-contracts/
    src/index.ts         // ids, intents, worker contracts, errors
  plugin-runtime-core/
    src/quickjsRuntimeService.ts
    src/dispatchIntent.ts
    src/uiSchema.ts
    src/runtimeIdentity.ts
  plugin-runtime-worker/
    src/quickjsRuntime.worker.ts
  plugin-runtime-redux/
    src/types.ts
    src/reducers/
    src/selectors/
    src/policies/
apps/
  plugin-playground/
    client/...           // UI only
```

### Key design decisions

1. Contracts package has zero React/Redux dependencies.
2. Core runtime package exposes deterministic APIs usable from browser and tests.
3. Redux package provides optional store integration, not required for runtime use.
4. UI app imports packages instead of owning runtime internals.

### Minimal host API sketch

```ts
interface PluginRuntimeHost {
  load(input: { packageId: string; instanceId?: string; code: string }): Promise<LoadedPlugin>;
  render(input: { instanceId: string; widgetId: string; pluginState: unknown; globalState: unknown }): Promise<UINode>;
  event(input: { instanceId: string; widgetId: string; handler: string; args?: unknown; pluginState: unknown; globalState: unknown }): Promise<DispatchIntent[]>;
  dispose(instanceId: string): Promise<boolean>;
}
```

## UI Overhaul Direction (Developer-Centric)

Goals:
- Make debugging plugin behavior first-class.
- Keep multi-instance operations understandable at a glance.
- Surface capability/read-write boundaries in context.

Core UX upgrades:
1. Instance Explorer
- Group by package, expand into instances, show capabilities and health badges.

2. Runtime Timeline
- Chronological intent trace with filters (`scope`, `domain`, `outcome`, `instanceId`).

3. Shared Domain Inspector
- Diff view before/after each shared action.

4. Plugin Workbench
- Preset/custom authoring with snippets, manifest editor, validation panel.

5. Failure-First Diagnostics
- Structured error cards: VM error, schema error, capability denial, reducer ignored.

## Alternatives Considered

1. Keep current single-app structure and only clean UI
- Rejected: does not solve reusable runtime requirement.

2. Extract only `quickjsRuntimeService.ts`
- Rejected: contracts/policies/selectors remain app-coupled; reuse still painful.

3. Full rewrite from scratch
- Rejected: high risk; existing behavior and tests are valuable and should be incrementally migrated.

## Implementation Plan

Phase A (cleanup and baseline)
1. Remove dead files and debug leftovers.
2. Add missing README and architecture docs.
3. Add lint/check rule preventing `window as any` debug escapes.

Phase B (package extraction)
1. Create `plugin-runtime-contracts` and migrate types.
2. Move runtime service + validators into `plugin-runtime-core`.
3. Move worker wrapper into `plugin-runtime-worker`.
4. Update playground imports and tests.

Phase C (UI overhaul)
1. Split `Playground` into feature modules (`catalog`, `workspace`, `runtime-inspector`).
2. Introduce runtime timeline and domain inspector.
3. Replace ad hoc local metadata caches with selector-driven view models.

Phase D (documentation completion)
1. Plugin authoring docs.
2. Runtime embedding docs.
3. Migration notes and troubleshooting runbook.

## Immediate Deletion/Deprecation Candidates

1. `client/src/components/PluginEditor.tsx` (unused)
2. `client/src/components/Map.tsx` (unused, unrelated)
3. `client/src/components/ManusDialog.tsx` (unused, unrelated)
4. `client/src/const.ts` and `shared/const.ts` (unused auth leftovers)
5. `DEBUG_FINDINGS.md` and `test-log.md` (historical logs; move to `archive/`)

## Acceptance Criteria for WEBVM-003

1. Playground runtime logic is mostly in packages, not page components.
2. A separate host app can run VM plugins using extracted package APIs.
3. Dead/template artifacts are removed or archived with rationale.
4. Developer-centric UI adds timeline/inspector workflows.
5. New docs are sufficient for first-time plugin authoring and runtime embedding.
