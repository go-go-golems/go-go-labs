---
Title: "Phase 3-4 Design Brief: Multi-Instance Identity and Capability Model"
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
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: Runtime slice and reducer routing target
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts
      Note: QuickJS lifecycle + VM bootstrap API changes
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts
      Note: Worker RPC client contract migration
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts
      Note: Request/response and dispatch intent type changes
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: Preset package manifests and API migration to shared-domain dispatch
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: Multi-instance load/unload UX + per-instance rendering context
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/workers/quickjsRuntime.worker.ts
      Note: Worker request routing keyed by instanceId
ExternalSources: []
Summary: "Resolved architecture for Phase 3 and 4: host-authoritative package/instance identity, package reducer registry, shared domain model, per-instance capability grants, and a concrete migration/test plan."
LastUpdated: 2026-02-09T00:00:00Z
WhatFor: "Implementation-ready design for multi-instance plugins and capability-governed shared state in the QuickJS playground runtime."
WhenToUse: "Use as the source of truth while implementing and reviewing Phase 3 and Phase 4 changes."
---

# Phase 3-4 Design Brief: Multi-Instance Identity and Capability Model

## Status

This document resolves all open design decisions from the prior brief. It is intentionally concrete and implementation-oriented.

## Scope

- Phase 3: split identity into `packageId` and `instanceId`, enable multi-instance loading, and remove plugin-ID-coupled reducer routing.
- Phase 4: replace flat globals with shared domains + capability grants enforced per instance.

Out of scope stays unchanged: plugin marketplace, persistence across sessions, hot-reload state preservation, server-side sandbox.

## Decision Summary

1. Identity is split into host-defined `packageId` (plugin type) and host-generated `instanceId` (runtime instance).
2. Worker/runtime/store contracts move to `instanceId` for instance operations; `packageId` is carried where routing is required.
3. Reducer routing is package-based through a host-owned registry, not by instance ID strings and not by code inside the VM.
4. Unknown packages use a small generic reducer contract (`state/replace`, `state/merge`) so custom plugins are functional.
5. Flat `globals` is replaced by named shared domains with host-owned reducers.
6. Capabilities are declared in host manifests (trusted source) and stored as per-instance grants.
7. Enforcement is defense-in-depth: pre-dispatch helper checks + reducer checks + filtered state projection to VM.
8. Denied writes are dropped and traced (not thrown), so unauthorized plugins cannot break UX loops.
9. Plugin VM API gains `dispatchSharedAction(domain, actionType, payload)`; `dispatchGlobalAction` remains as a deprecated compatibility alias during migration.
10. System commands are represented in schema but intentionally deferred in Phase 4 implementation.

## Phase 3: Multi-Instance Identity Design

### Identity Model

```ts
export type PackageId = string;   // stable plugin type id, e.g. "counter"
export type InstanceId = string;  // runtime instance id, e.g. "counter@a1b2c3d4"

export interface RuntimeIdentity {
  packageId: PackageId;
  instanceId: InstanceId;
}
```

`instanceId` generation is centralized in a host utility (`runtimeIdentity.ts`), used by `Playground` before load:

```ts
export function createInstanceId(packageId: PackageId): InstanceId {
  return `${packageId}@${nanoid(8)}`;
}
```

Rationale:
- Generation stays host-authoritative.
- Prefix preserves operator readability in logs/UI.
- Central utility avoids format drift.

### Worker + Runtime Contract Changes

All runtime operations are keyed by `instanceId`.

```ts
export interface LoadPluginRequest {
  id: number;
  type: "loadPlugin";
  packageId: PackageId;
  instanceId: InstanceId;
  code: string;
}

export interface RenderRequest {
  id: number;
  type: "render";
  instanceId: InstanceId;
  widgetId: string;
  pluginState: unknown;
  globalState: unknown;
}

export interface EventRequest {
  id: number;
  type: "event";
  instanceId: InstanceId;
  widgetId: string;
  handler: string;
  args?: unknown;
  pluginState: unknown;
  globalState: unknown;
}

export interface DisposePluginRequest {
  id: number;
  type: "disposePlugin";
  instanceId: InstanceId;
}

export interface LoadedPlugin {
  packageId: PackageId;
  instanceId: InstanceId;
  declaredId?: string;
  title: string;
  description?: string;
  initialState?: unknown;
  widgets: string[];
}
```

`pluginId` is removed from contracts (clean break). There is no backward-compat layer in transport types to avoid dual-schema complexity.

### Dispatch Intent Attribution

```ts
export interface DispatchIntent {
  scope: "plugin" | "shared";
  actionType: string;
  payload?: unknown;
  instanceId?: InstanceId; // required for scope=plugin
  domain?: string;         // required for scope=shared
}
```

`validateDispatchIntents()` stamps `instanceId` for plugin-scoped intents using runtime context.

### Runtime Slice Data Model

```ts
export interface RuntimePluginInstance {
  instanceId: InstanceId;
  packageId: PackageId;
  title: string;
  description?: string;
  widgets: string[];
  enabled: boolean;
  status: "loaded" | "error";
  error?: string;
}

export interface CapabilityGrants {
  readShared: string[];
  writeShared: string[];
  systemCommands: string[];
}

interface RuntimeState {
  instances: Record<InstanceId, RuntimePluginInstance>;
  pluginStateByInstance: Record<InstanceId, unknown>;
  grantsByInstance: Record<InstanceId, CapabilityGrants>;
  shared: Record<string, unknown>;
  dispatchTrace: {
    count: number;
    lastDispatchId: string | null;
    lastScope: "plugin" | "shared" | null;
    lastActionType: string | null;
    lastOutcome: "applied" | "denied" | "ignored" | null;
    lastReason: string | null;
  };
}
```

### Reducer Routing Model

Reducer routing is package-centric.

```ts
type LocalReducer = (
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload: unknown
) => "applied" | "ignored";

const localReducerRegistry: Record<PackageId, LocalReducer> = {
  counter: reduceCounterLocal,
  calculator: reduceCalculatorLocal,
  greeter: reduceGreeterLocal,
};
```

Algorithm for plugin-scoped dispatch:
1. Resolve instance by `instanceId`.
2. Resolve `packageId` from `instances[instanceId]`.
3. Lookup `localReducerRegistry[packageId]`.
4. If found, execute.
5. If not found, execute `reduceGenericLocal`:

```ts
// Generic fallback for custom plugins
// state/replace: payload becomes full local state
// state/merge: payload object shallow-merges into local state
```

This directly solves current behavior where custom plugin local actions are dropped silently.

### Global Mirroring Decision

Manual mirroring into flat globals does not survive Phase 3/4. It is replaced by Phase 4 shared domains. During migration, existing mirror behavior is preserved only until shared-domain reducers are in place.

### Playground UX for Multi-Instance

- Clicking a preset always creates a new instance.
- Preset list no longer uses "loaded" checkmark by `packageId`; instead show a badge count per package (`xN`).
- Loaded panel shows entries by `instanceId` with title and package.
- Unload removes only that instance.
- Optional guardrail: max 10 instances per package (soft limit with UI error).

### Selector Changes

```ts
selectPluginState(state, instanceId)
selectAllPluginState(state)
selectInstances(state)
selectInstancesByPackage(state, packageId)
selectInstanceCountByPackage(state)
```

## Phase 4: Capability Model and Shared Domains

### Shared Domain Registry

Shared state becomes a registry of named domains with host-owned reducers.

```ts
type DomainName = string;

type SharedDomainReducer = (
  current: unknown,
  actionType: string,
  payload: unknown,
  context: { instanceId: InstanceId; packageId: PackageId }
) => { next: unknown; outcome: "applied" | "ignored"; reason?: string };

interface SharedDomainDefinition {
  name: DomainName;
  initialState: unknown;
  reducer: SharedDomainReducer;
  publicProjection?: (state: unknown) => unknown;
}

const sharedDomainRegistry: Record<DomainName, SharedDomainDefinition> = {
  "counter-summary": { ... },
  "greeter-profile": { ... },
  "runtime-metrics": { ... },
  "runtime-registry": { ... },
};
```

Domain reducers are host-defined only. Plugins never provide executable reducers.

### Manifest and Capability Declaration

Capabilities are defined in trusted host metadata (preset manifest + host policy for custom plugins).

```ts
interface PackageManifest {
  packageId: PackageId;
  title: string;
  description?: string;
  capabilities: {
    readShared?: DomainName[];
    writeShared?: DomainName[];
    systemCommands?: string[];
  };
}
```

Grant model:
- Requested capabilities come from host manifest.
- Host policy computes effective grant at load time.
- Effective grant is stored at `grantsByInstance[instanceId]`.

Default policy for custom plugins:
- `readShared: []`
- `writeShared: []`
- `systemCommands: []`

### Capability Enforcement

Enforcement happens at three layers.

1. `dispatchSharedAction()` helper (fast deny before reducer dispatch).
2. `sharedActionDispatched` reducer (authoritative deny; logs outcome).
3. Per-instance `globalState` projection (plugins cannot read domains they lack).

Denied write behavior:
- Intent is dropped.
- Dispatch trace stores `lastOutcome: "denied"` and reason like `missing-write-grant:<domain>`.
- No throw into app UI loop.

### Shared Action Shape

Action transport includes explicit domain:

```ts
{
  scope: "shared";
  domain: "counter-summary";
  actionType: "set" | "increment" | string;
  payload?: unknown;
}
```

Using explicit `domain` avoids brittle string parsing and allows concise action names.

### Per-Instance Global View

Plugins no longer receive ungoverned global objects. They receive a projected view:

```ts
interface PluginGlobalView {
  self: {
    instanceId: InstanceId;
    packageId: PackageId;
  };
  shared: Record<DomainName, unknown>; // readShared filtered
}
```

`Playground`/store selector builds this view before each `render()` and `event()` call.

### System Commands

In Phase 4 design schema: included.
In Phase 4 implementation: deferred.

- Keep `systemCommands` in grants and manifest.
- No runtime command executor added yet.
- Unknown system command intents are denied and traced.

### VM Bootstrap API Changes

Final API:
- `dispatchPluginAction(actionType, payload?)`
- `dispatchSharedAction(domain, actionType, payload?)`

Transition compatibility (one phase):
- Keep `dispatchGlobalAction(actionType, payload?)` alias mapped to `dispatchSharedAction("legacy-global", actionType, payload)`.
- Migrate presets off alias in same PR.
- Remove alias after all presets/tests use shared-domain calls.

## Updated Preset Package Plan

1. `counter`
- Local reducer keeps `value` per instance.
- Shared writes move to `dispatchSharedAction("counter-summary", "set-instance", { value })`.
- Domain reducer computes aggregate (`totalValue`, `instanceCount`, `lastUpdatedInstanceId`).

2. `calculator`
- Local-only plugin. No shared write grant.

3. `greeter`
- Local reducer keeps name per instance.
- Optional shared write to `greeter-profile` domain if enabled in manifest.

4. `status-dashboard`
- Read-only grants: `runtime-metrics`, `counter-summary`, `runtime-registry`.

5. `runtime-monitor`
- Read-only grant: `runtime-registry`.

6. `greeter-shared-state`
- Read-only grant: `greeter-profile`.

## Migration Plan (File-by-File, Safe Order)

1. `client/src/lib/runtimeIdentity.ts` (new)
- Add `createInstanceId(packageId)` utility and exported ID types.

2. `client/src/lib/quickjsContracts.ts`
- Replace `pluginId` transport with `instanceId` + `packageId` where needed.
- Add `scope: "shared"` + `domain` to dispatch intents.

3. `client/src/lib/quickjsRuntimeService.ts`
- `PluginVm` carries `{ instanceId, packageId }`.
- VM map keyed by `instanceId`.
- Remove implicit replacement-on-load behavior.
- Update bootstrap dispatch API.

4. `client/src/workers/quickjsRuntime.worker.ts`
- Route requests by `instanceId`.

5. `client/src/lib/quickjsSandboxClient.ts`
- Update method signatures to `instanceId`.

6. `client/src/lib/dispatchIntent.ts`
- Stamp `instanceId` on plugin intents.
- Validate shared intents include `domain`.

7. `client/src/store/reducers/localReducers.ts` (new)
- Implement package reducers + generic fallback reducer.

8. `client/src/store/reducers/sharedDomainReducers.ts` (new)
- Implement domain registry and reducers.

9. `client/src/store/capabilities.ts` (new)
- Manifest/grant types + policy evaluation helpers.

10. `client/src/store/store.ts`
- Replace plugin maps with instance/grant/shared model.
- Add plugin/shared action reducers with outcome tracing.
- Add selectors for per-instance projections.

11. `client/src/lib/presetPlugins.ts`
- Add host-side package manifests with capability declarations.
- Update plugin code to `dispatchSharedAction` where relevant.

12. `client/src/pages/Playground.tsx`
- Generate `instanceId` on load.
- Load/unload/render/event by `instanceId`.
- Show multi-instance UI and package counts.

13. Tests
- Update unit/integration/e2e in the same sequence (details below).

Implementation sequence guardrail:
- Land Phase 3 identity + reducer routing first with legacy global behavior still intact.
- Land Phase 4 shared domains/capabilities second.
- Remove legacy global alias third.

## Test Strategy

### Unit Tests

1. `quickjsContracts` and `dispatchIntent`
- Reject missing `domain` for shared intents.
- Stamp `instanceId` correctly for plugin intents.

2. Store reducer tests (new)
- Two instances of same package evolve independently.
- Unknown package reducer fallback handles `state/replace` and `state/merge`.
- Shared write denied without grant.
- Shared read projection excludes ungranted domains.

3. Domain reducer tests
- `counter-summary` aggregation semantics across add/remove/update instance events.

### Integration Tests (`quickjsRuntimeService.integration.test.ts`)

1. Load two `counter` instances and verify separate VM handles.
2. `event()` returns intents with `instanceId` attribution.
3. Disposing one instance does not affect the other.
4. Backward compatibility alias test for temporary `dispatchGlobalAction` path.

### E2E Tests (`tests/e2e/quickjs-runtime.spec.ts`)

1. Multi-instance counter
- Load counter twice.
- Increment first instance.
- Assert second instance unchanged.

2. Capability enforcement
- Load plugin without `writeShared` grant that emits shared action.
- Assert no shared-state mutation and no app crash.
- Assert denial surfaced in runtime metrics panel.

3. Shared read filtering
- Plugin without read grants sees empty `globalState.shared`.

## Resolved Open Questions (Explicit)

1. Where is `instanceId` generated?
- Host utility (`runtimeIdentity.ts`), called by `Playground` at load initiation.

2. Are contracts rename-only or dual ID?
- Runtime operation fields are `instanceId`; `packageId` is explicit where lifecycle/routing requires.

3. How are package reducers registered?
- Host-side registry map keyed by `packageId`.

4. What for custom plugin reducers?
- Generic fallback reducer contract (`state/replace`, `state/merge`).

5. Do flat globals survive?
- No. Replaced by shared domains.

6. How are domains registered?
- Host-side registry; plugin code cannot register executable reducers.

7. Where does manifest live?
- Host-owned manifest (`presetPlugins` metadata + custom default policy).

8. Where is enforcement?
- Dispatch helper + reducer + read projection (defense in depth).

9. Denial behavior?
- Drop + trace, not throw.

10. Is system command execution in scope?
- Schema yes, executor no (deferred).

11. Bootstrap API shape?
- New `dispatchSharedAction(domain, actionType, payload)` with one-phase alias.

## Risks and Mitigations

1. Risk: silent behavior changes during dual API period.
- Mitigation: explicit deprecation logs when `dispatchGlobalAction` alias is used.

2. Risk: large store refactor regressions.
- Mitigation: add reducer unit tests before wiring UI changes.

3. Risk: capability policy drift between presets and runtime.
- Mitigation: centralize manifest/grant types and import from one module.

4. Risk: state leak across instances.
- Mitigation: enforce `instanceId` keying in all selectors and action creators; add regression tests for two-instance isolation.

## Acceptance Criteria

1. Loading same preset multiple times creates independent widgets and state.
2. Unloading one instance never unloads siblings.
3. Local action routing depends on `packageId`, not `instanceId` string literals.
4. Shared writes are denied without explicit per-instance grant.
5. Plugins only read granted shared domains.
6. Existing runtime timeout and error handling remain intact.
7. Unit, integration, and e2e suites pass with updated expectations.
