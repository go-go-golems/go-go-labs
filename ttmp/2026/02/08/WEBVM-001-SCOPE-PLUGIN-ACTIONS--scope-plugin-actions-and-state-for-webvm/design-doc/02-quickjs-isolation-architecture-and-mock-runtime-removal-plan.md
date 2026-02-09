---
Title: QuickJS Isolation Architecture and Mock Runtime Removal Plan
Ticket: WEBVM-001-SCOPE-PLUGIN-ACTIONS
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/pluginManager.ts
      Note: Current main-thread new-Function runtime slated for removal
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiTypes.ts
      Note: Canonical UINode contract to preserve during runtime replacement
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: Active runtime entrypoint that must migrate to QuickJS worker client
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json
      Note: Declares quickjs-emscripten dependency used for real isolation
    - Path: ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/design-doc/01-quickjs-migration-implementation-guide-and-test-strategy.md
      Note: Execution-level migration and testing guide spun out into WEBVM-002 ticket
ExternalSources: []
Summary: Design for replacing in-process mock plugin execution with real QuickJS WASM isolation in a dedicated worker, including phased removal of mock runtime paths.
LastUpdated: 2026-02-08T18:18:00Z
WhatFor: Define a production-realistic QuickJS runtime architecture and a concrete plan to remove the current mock/new-Function execution paths.
WhenToUse: Use when implementing QuickJS worker isolation, runtime hardening, or removing legacy plugin execution code paths.
---



# QuickJS Isolation Architecture and Mock Runtime Removal Plan

## Executive Summary

The current plugin playground presents itself as QuickJS-based, but active plugin execution is still mock/in-process:

- `new Function(...)` in `pluginManager` (main thread)
- `new Function(...)` in worker path (`pluginSandbox.worker.ts`)
- global `window.definePlugin` wiring in `pluginSandboxClient.ts`

This is not real VM isolation. It leaves plugins with host access, weak fault isolation, and inconsistent runtime behavior.

This document proposes a real QuickJS isolation architecture using `quickjs-emscripten` in a dedicated Web Worker and a migration plan to remove the mock runtime path.

Key outcomes of this plan:

1. Plugin JS executes inside QuickJS VM, not browser/global JS engine.
2. Host interaction is only via an explicit bridge API.
3. Time/memory limits become enforceable at runtime level.
4. Mock paths are removed so there is one runtime model to debug and maintain.

## Problem Statement

### What Is Broken Today

1. Isolation claims vs implementation mismatch.
- UI and comments reference QuickJS/WASM isolation.
- Execution currently happens through `new Function` on host JS engine.

2. Split runtime implementations.
- Active path (`Playground` + `pluginManager`) is in-process.
- Alternate path (`pluginSandboxClient` + worker) is partially wired and contract-drifted.

3. Security and stability risk.
- Plugin code can touch host globals in in-process mode.
- Infinite loops run on main thread and can freeze UI.
- Action/state boundaries are convention-based rather than runtime-enforced.

4. Hard-to-debug drift.
- Different node schemas (`kind` vs `type`) across paths.
- Different event signatures and loading semantics.

### Repository Reality Check (2026-02-08)

The active implementation in this repository currently has:

- `pluginManager.ts` + `Playground.tsx` on the critical path.
- No `client/src/workers/` directory yet.
- No `pluginSandboxClient.ts` file in the active `plugin-playground` path.

This means the migration work starts from an in-process baseline and must introduce worker/runtime files rather than only refactor existing ones.

### Why QuickJS Isolation Matters

Real QuickJS isolation provides:

- Separate JS heap/runtime for plugin code.
- No implicit access to browser `window`, DOM, fetch, etc.
- Runtime-level controls (`setMemoryLimit`, `setMaxStackSize`, interrupt handler).
- Cleaner host/plugin contract via explicit bridge functions.

## Current Architecture vs Target Architecture

### Current Runtime (Simplified)

```text
React UI (main thread)
  -> pluginManager.loadPlugin(code)
    -> new Function("definePlugin", code)
      -> plugin object in host memory
  -> widget.render({state}) in host
  -> handler({dispatch,state}) in host
```

### Target Runtime (Real Isolation)

```text
React UI (main thread)
  -> QuickJSSandboxClient (RPC)
    -> Web Worker (QuickJSRuntimeService)
      -> QuickJS runtime/context
        -> plugin code eval inside QuickJS
      <- serializable UINode tree
      <- dispatch events (plugin/global actions with dispatchId)
```

Key boundary: plugin code never runs in browser main-thread global context.

## Proposed Solution

## 1) Make Worker + QuickJS the Only Execution Path

Create one canonical runtime service in worker:

- initialize QuickJS via `getQuickJS()`
- create runtime/context per plugin instance (or per trust zone)
- load plugin code in VM
- handle render/event calls via RPC
- emit host dispatch intents back to main thread

Main thread only keeps:

- UI rendering
- Redux store
- RPC client + dispatch gateway

## 2) Define a Strict Runtime Bridge API

Inside QuickJS context, expose minimal host bridge functions only.

### Proposed Bridge (inside VM)

- `hostDispatchPlugin(type, payload)`
- `hostDispatchGlobal(type, payload)`
- `hostNow()`
- optional `hostLog(level, msg)` for debug mode

No direct host object references. No dynamic eval APIs from host beyond plugin load.

## 3) Keep Plugin API Compatible with Simplified V1 Scoping

Align with your current v1 direction:

- State selectors on host:
  - `selectPluginState(pluginId)`
  - `selectGlobalState()`
- Actions from plugin bridge:
  - plugin-scoped
  - global
- Every dispatch gets global `dispatchId`.

QuickJS isolation and simplified scoping are compatible and complementary.

## 4) Enforce Runtime Resource Limits

Per QuickJS runtime:

- `setMemoryLimit(limitBytes)`
- `setMaxStackSize(stackBytes)`
- `setInterruptHandler(...)` with deadline/cycle budget

Use interrupt handler to stop runaway code and return structured timeout errors.

## 5) Remove Mock Runtime Code Paths Completely

After migration:

- Remove `pluginManager` `new Function` path.
- Remove in-process `window.definePlugin` path.
- Remove worker `new Function` fallback.
- Keep one runtime implementation and one contract.

## Design Decisions

## D1: Run QuickJS in Worker, Not Main Thread

Reason:

- Prevent UI freezes due to plugin compute.
- Isolate plugin VM failures from render thread.
- Keep RPC boundary explicit.

Tradeoff:

- Slightly higher complexity and serialization overhead.

## D2: One Runtime per Plugin Instance (Default)

Reason:

- Strongest isolation between plugins.
- Easy teardown and memory accounting per plugin.
- Simpler fault containment.

Tradeoff:

- Higher memory/CPU overhead than shared runtime.

Note:

- If performance requires, can move to pooled runtime model later.

## D3: No Direct Host State in VM

Reason:

- Keep data flow explicit and serializable.
- Avoid hidden coupling and accidental mutation.

Mechanism:

- Host sends state snapshots (`pluginState`, `globalState`) to render/event calls.

## D4: Action Egress Is Intent-Based, Not Raw Dispatch

Reason:

- VM should not invoke Redux directly.
- Host owns action envelope and dispatch pipeline.

Mechanism:

- VM emits dispatch intent -> host stamps `dispatchId`, validates, dispatches.

## D5: Remove Dual Runtime Paths

Reason:

- Split-brain architecture is causing drift and confusion.
- One canonical path improves reliability and onboarding.

## QuickJS Runtime Architecture (Detailed)

## Worker Service Responsibilities

1. Runtime lifecycle
- initialize QuickJS module once
- create/dispose plugin runtimes

2. RPC handling
- `loadPlugin`
- `render`
- `event`
- `disposePlugin`
- `health`

3. Resource enforcement
- timeouts, memory limits, stack limits
- runaway script interruption

4. Serialization and validation
- parse/validate plugin return structure
- ensure output conforms to canonical `UINode` schema

## VM Bootstrap Pattern

On plugin load:

1. Create runtime + context.
2. Install host bridge functions into VM global.
3. Install bootstrap script that defines:
   - `definePlugin`
   - plugin registry holder
   - safe wrappers for render/event calls
4. Eval plugin code.
5. Validate plugin definition.
6. Return metadata to host.

### Bootstrap Pseudocode

```ts
const QuickJS = await getQuickJS();
const runtime = QuickJS.newRuntime();
runtime.setMemoryLimit(MEM_LIMIT_BYTES);
runtime.setMaxStackSize(STACK_LIMIT_BYTES);
runtime.setInterruptHandler(makeDeadlineInterrupt(deadlineRef));

const vm = runtime.newContext();
installBridge(vm, bridgeFns);
vm.unwrapResult(vm.evalCode(BOOTSTRAP_SOURCE));

const loadResult = vm.evalCode(pluginCode, "plugin.js");
if (loadResult.error) throw toHostError(vm, loadResult.error);

const meta = callVmHelper(vm, "__pluginHost.getMeta");
```

## Render/Event Calling Pattern

### Render RPC

Host -> worker:

```json
{ "type": "render", "pluginId": "...", "widgetId": "...", "pluginState": {...}, "globalState": {...} }
```

Worker actions:

1. deadline starts.
2. invoke VM helper `__pluginHost.render(widgetId, state)`.
3. validate output tree schema (`kind` contract).
4. return serializable tree to host.

### Event RPC

Host -> worker:

```json
{ "type": "event", "pluginId": "...", "widgetId": "...", "handler": "...", "event": {...}, "pluginState": {...}, "globalState": {...} }
```

Worker actions:

1. deadline starts.
2. invoke VM handler.
3. VM may emit dispatch intents via bridge.
4. worker forwards intents to host.

## Bridge Function Implementation Notes

Use `vm.newFunction(...)` to define bridge functions exposed to plugin code.

Example host bridge callback responsibilities:

- decode VM arguments (`type`, `payload`, optional metadata)
- return only primitives/serializable values
- never expose host objects directly
- keep bridge narrow and versioned

## Security Model

## What QuickJS Isolation Gives

1. No implicit browser global access in VM.
2. Host bridge is explicit and auditable.
3. Runtime CPU/memory controls available.
4. Safer execution for buggy/malicious plugin code.

## What It Does Not Automatically Solve

1. Denial-of-service if limits are too loose.
2. Unsafe host bridge design.
3. Business-logic misuse via allowed global actions.
4. Network/storage side effects if bridge exposes them.

Isolation is necessary but not sufficient; bridge design still matters.

## Runtime Limits and Fault Handling

## Suggested Initial Limits

- Memory limit per plugin runtime: `16MB` to `32MB`.
- Stack limit: `512KB` to `1MB`.
- Render timeout: `30ms` to `100ms`.
- Event timeout: `30ms` to `100ms`.

Tune with measurements from real presets.

## Timeout Strategy

- On each RPC call, set deadline.
- interrupt handler checks deadline and returns interrupt signal.
- worker catches interruption and returns structured timeout error.
- host marks plugin status as degraded/error.

## Crash/Leak Strategy

- Any VM exception returns structured error and keeps runtime alive if safe.
- Repeated fatal errors trigger plugin runtime recycle.
- On unload, always `context.dispose()` then `runtime.dispose()`.

## Contract Standardization

## UINode Contract

Canonical shape must remain `kind`-based (as used by `WidgetRenderer`).

Reject or transform any `type`-based output from old code.

## Event Contract

Use one event payload format across runtime:

```ts
interface PluginEvent {
  name: string;
  args?: unknown;
  value?: unknown;
}
```

Avoid multiple handler signatures across paths.

## Dispatch Intent Contract

```ts
interface DispatchIntent {
  scope: "plugin" | "global";
  pluginId?: string;
  type: string;
  payload?: unknown;
}
```

Host adds final metadata including global `dispatchId`.

## Concrete Removal Plan (Rip Out Mock Version)

## Phase 1: Introduce Real QuickJS Worker Runtime

Add new files:

- `client/src/workers/quickjsRuntime.worker.ts`
- `client/src/lib/quickjsSandboxClient.ts`
- `client/src/lib/quickjsContracts.ts`

Do not delete old code yet; feature-flag the new path.

## Phase 2: Route Playground Through New Client

Update `Playground` to use `quickjsSandboxClient` for:

- plugin load
- render
- event
- unload

Ensure UI remains unchanged from user perspective.

## Phase 3: Remove `new Function` Paths

Delete or archive:

- `client/src/lib/pluginManager.ts`
- in-process execution parts of `client/src/lib/pluginSandboxClient.ts`
- `new Function` in `client/src/workers/pluginSandbox.worker.ts`

Replace references in imports and store types.

## Phase 4: Remove Dead Components and Drifted Contracts

- Remove or refactor components that depend on legacy contracts.
- Enforce one runtime contract in `uiTypes.ts` and worker contracts.

## Phase 5: Hardening and Test Gates

- Add VM timeout tests.
- Add memory pressure tests.
- Add malicious script tests.
- Add contract compatibility tests.

## File-by-File Change Map

### `client/src/pages/Playground.tsx`

- Replace direct `pluginManager` usage with sandbox client RPC calls.
- Keep loaded plugin registry keyed by authoritative `pluginId`.

### `client/src/lib/pluginManager.ts`

- Remove file after migration.

### `client/src/lib/pluginSandboxClient.ts`

- Either delete or rewrite into thin wrapper around QuickJS worker client.
- Remove `window.definePlugin` and `new Function` usage.

### `client/src/workers/pluginSandbox.worker.ts`

- Replace with new QuickJS worker runtime implementation.
- Remove dynamic `new Function` execution.

### `client/src/lib/uiTypes.ts`

- Keep canonical `kind` contract.
- Add schema validator utility for runtime responses.

### `client/src/store/store.ts`

- Keep simplified action/state model, but consume worker dispatch intents via host action wrappers.

## Test Strategy for “Real” Isolation

## Unit Tests

1. Bridge function behavior:
- plugin intent -> host event conversion
- invalid args handling

2. Contract validation:
- UINode schema accept/reject
- event payload schema accept/reject

3. Limit handling:
- interrupt on timeout
- memory limit exception mapping

## Integration Tests

1. Load plugin in QuickJS and render successfully.
2. Dispatch plugin-scoped action from VM and verify host receives it with `dispatchId`.
3. Dispatch global action from VM and verify allowlist behavior.
4. Infinite loop script is interrupted and plugin marked error.
5. Plugin unload disposes runtime/context with no leaks.

## Adversarial Tests

1. `while(true){}` in render.
2. deep recursion stack blow.
3. large allocation loops.
4. attempts to access `window`/`document`.
5. malformed UI tree returns.

## Operational Diagnostics

For each plugin runtime call, log:

- `pluginId`
- operation (`load`, `render`, `event`)
- elapsed ms
- interrupted (bool)
- memory error (bool)
- error type/code if failed

For each dispatch intent, log:

- `dispatchId`
- source `pluginId`
- scope (`plugin|global`)
- action type
- validation result

## Risks and Mitigations

## Risk 1: Performance Regression from Worker + VM

Mitigation:

- Batch or memoize render calls.
- Keep payloads small.
- Benchmark per-widget latency before/after.

## Risk 2: QuickJS Handle Leaks in Worker

Mitigation:

- strict handle disposal discipline.
- wrapper helpers for `unwrapResult` + dispose patterns.
- leak tests on repetitive load/render/unload cycles.

## Risk 3: Migration Breaks Existing Preset Plugins

Mitigation:

- compatibility layer for plugin API shape.
- migrate presets one-by-one with contract tests.

## Risk 4: False Sense of Security

Mitigation:

- document trust model and bridge attack surface explicitly.
- keep global action allowlist and strict schema validation.

## Alternatives Considered

## A) Keep Current Mock Runtime, Add Minor Guards

Why rejected:

- still not real isolation
- still split runtime drift
- still `new Function` risk surface

## B) Use Only Native Worker Without QuickJS

Why rejected:

- plugin code still runs in browser JS engine
- weaker isolation semantics
- harder to enforce deterministic runtime constraints

## C) Full Server-Side Sandbox Instead of Browser QuickJS

Why deferred:

- stronger isolation, but higher latency and operational complexity
- not necessary for current local playground goals

## Implementation Plan (Detailed Milestones)

## Milestone 0: Architecture Guardrail (1 day)

- Freeze new `new Function` additions.
- Add TODO banners pointing to this ticket.

## Milestone 1: QuickJS Worker Skeleton (2 days)

- Worker boot with `getQuickJS()`.
- runtime/context create/dispose flows.
- basic RPC framing.

## Milestone 2: Plugin Load + Meta Extraction (2 days)

- VM bootstrap script.
- `definePlugin` capture.
- metadata return path.

## Milestone 3: Render + Event Execution (2 days)

- host state snapshot input (`pluginState`, `globalState`).
- UINode output + validation.
- handler invocation path.

## Milestone 4: Dispatch Intent Bridge (1 day)

- plugin/global dispatch intents from VM.
- host wrappers attach `dispatchId`.

## Milestone 5: Hard Limits and Error Policy (1 day)

- interrupt deadlines
- memory/stack limits
- structured error envelope

## Milestone 6: Cutover and Removal (2 days)

- switch `Playground` to new runtime.
- delete old mock runtime files/paths.
- clean imports and dead types.

## Milestone 7: Tests + Docs + Rollout (2 days)

- full integration and adversarial tests.
- update ticket docs and runbook.

## Execution Handoff

Implementation and test execution has been split into follow-on ticket:

- `WEBVM-002-QUICKJS-MIGRATION`
- Path: `ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates`
- Primary guide: `design-doc/01-quickjs-migration-implementation-guide-and-test-strategy.md`

This document remains the architecture/rationale source. WEBVM-002 tracks concrete file-by-file implementation sequencing, acceptance criteria, and Playwright-backed test gates.

## Success Criteria

1. No plugin execution via `new Function` in production path.
2. All plugin code executes in QuickJS worker runtime.
3. Infinite loop plugin cannot freeze UI thread.
4. Plugin dispatches include global `dispatchId`.
5. Single runtime contract for load/render/event.
6. Old mock runtime files removed or archived outside active build.

## Open Questions

1. One runtime/context per plugin instance vs pooled runtimes for performance?
2. Required timeout values for acceptable UX with complex widgets?
3. Do we need plugin module loading support (`import`) in v1?
4. Should debug logging bridge be enabled only in dev builds?
5. Should runtime recycle automatically after N errors?

## References

- `client/src/lib/pluginManager.ts`
- `client/src/lib/pluginSandboxClient.ts`
- `client/src/workers/pluginSandbox.worker.ts`
- `client/src/pages/Playground.tsx`
- `client/src/lib/uiTypes.ts`
- `client/src/components/WidgetRenderer.tsx`
- `package.json` (`quickjs-emscripten` dependency)
- `node_modules/.pnpm/quickjs-emscripten@0.23.0/node_modules/quickjs-emscripten/README.md`
- `node_modules/.pnpm/quickjs-emscripten@0.23.0/node_modules/quickjs-emscripten/dist/runtime.d.ts`
- `node_modules/.pnpm/quickjs-emscripten@0.23.0/node_modules/quickjs-emscripten/dist/context.d.ts`

## Appendix A: Example Worker RPC Types

```ts
type Req =
  | { id: number; type: "loadPlugin"; pluginId: string; code: string }
  | { id: number; type: "render"; pluginId: string; widgetId: string; pluginState: unknown; globalState: unknown }
  | { id: number; type: "event"; pluginId: string; widgetId: string; handler: string; event: unknown; pluginState: unknown; globalState: unknown }
  | { id: number; type: "disposePlugin"; pluginId: string };

type Res = { id: number; ok: boolean; result?: unknown; error?: { code: string; message: string; details?: unknown } };

type DispatchIntentEvent = {
  type: "dispatchIntent";
  pluginId: string;
  scope: "plugin" | "global";
  actionType: string;
  payload?: unknown;
};
```

## Appendix B: Example Host Dispatch Stamping

```ts
function stampDispatch(intent: DispatchIntentEvent) {
  return {
    type: intent.actionType,
    payload: intent.payload,
    meta: {
      dispatchId: crypto.randomUUID(),
      scope: intent.scope,
      pluginId: intent.pluginId,
      ts: Date.now(),
      source: "quickjs-worker",
    },
  };
}
```
