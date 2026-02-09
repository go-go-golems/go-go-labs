---
Title: Plugin Action and State Scoping Architecture Review
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
    - Path: client/src/components/WidgetRenderer.tsx
      Note: Canonical renderer contract for UINode kind-based trees
    - Path: client/src/lib/pluginManager.ts
      Note: In-process plugin execution and ID storage behavior
    - Path: client/src/lib/pluginSandboxClient.ts
      Note: Alternate sandbox path and dispatch policy gap
    - Path: client/src/lib/presetPlugins.ts
      Note: Active preset plugin API samples and action semantics
    - Path: client/src/pages/Playground.tsx
      Note: Primary active runtime and plugin load/render/event orchestration
    - Path: client/src/store/store.ts
      Note: Current reducer model and plugin action matcher behavior
    - Path: client/src/workers/pluginSandbox.worker.ts
      Note: Worker contract drift and dispatch/event path
ExternalSources: []
Summary: Deep architecture analysis of plugin identity, action scoping, and state isolation in plugin-playground, updated with a simplified v1 contract using selectPluginState/selectGlobalState and plugin/global action dispatch.
LastUpdated: 2026-02-08T18:15:00Z
WhatFor: Diagnose plugin ID and scope issues and provide an implementation strategy centered on a pragmatic v1 API with optional future hardening.
WhenToUse: Use when implementing plugin runtime selectors, action wrappers, dispatch tracing, or plugin lifecycle changes in plugin-playground/WebVM.
---


# Plugin Action and State Scoping Architecture Review

## Executive Summary

This review analyzes the current `plugin-playground` system with specific focus on the problem you called out: plugin IDs, action scoping, and state scoping in a multi-plugin environment where some state/actions must be shared.

Current behavior works for a constrained demo but does not provide enforceable scope boundaries:

- Plugin identity authority is ambiguous.
- Action namespaces are cooperative rather than enforced.
- Plugin state is effectively global and hardcoded.
- Two runtime architectures coexist (one active, one mostly dormant), and they disagree on contracts.

The system is currently in a **“functional prototype” stage**, not a **“safe multi-tenant plugin runtime” stage**.

Primary recommendation (updated to the simplified v1 model):

1. Keep host-assigned plugin identity (`pluginId`) authoritative.
2. Expose exactly two state selectors:
   - `selectPluginState(pluginId)`
   - `selectGlobalState()`
3. Expose exactly two action scopes:
   - plugin-scoped action (`dispatchPluginAction`)
   - global action (`dispatchGlobalAction`)
4. Require a global `dispatchId` on every dispatched action (both scopes).
5. Defer capability granularity until after v1 contracts are stable.

This document includes:

- Architecture map and control/data-flow trace.
- Concrete code-level findings with file references.
- A target architecture that supports plugin-local scoping with a small global surface.
- A phased migration/implementation plan with test strategy and risk controls.

## Simplified V1 Contract (Authoritative for This Ticket)

The initial implementation should stay intentionally small.

### State API

```ts
selectPluginState(pluginId: string): unknown
selectGlobalState(): unknown
```

Rules:

- `selectPluginState(pluginId)` is the default selector for plugin UI and handlers.
- `selectGlobalState()` is for intentionally shared state only.
- Do not pass full Redux root state directly to plugin code.

### Action API

```ts
dispatchPluginAction(pluginId: string, type: string, payload?: unknown): void
dispatchGlobalAction(type: string, payload?: unknown): void
```

Both functions stamp:

```ts
meta: {
  dispatchId: string; // globally unique id
  scope: "plugin" | "global";
  pluginId?: string;
  ts: number;
}
```

### Does This Work?

Yes. For current scope, this is the right level of complexity.

### Pros

1. Easy to implement and reason about.
2. Immediate separation between per-plugin and shared/global behavior.
3. Fast migration path from current reducer and handler code.
4. `dispatchId` gives useful tracing without heavy policy machinery.
5. Leaves room to add capabilities later without replacing the basic API.

### Cons

1. `dispatchGlobalAction` is still broad unless you add an allowlist.
2. `selectGlobalState()` can become a coupling point if unmanaged.
3. No fine-grained permission controls yet.
4. Not sufficient for untrusted third-party plugins by itself.

### Minimal Guardrails for V1

1. Maintain a host-side allowlist of global action types.
2. Keep `selectGlobalState()` mapped to a curated object (not raw root state).
3. Reject plugin-scoped actions whose `pluginId` is unknown.

## Scope and Method

### Repositories and Files Reviewed

The analysis focused on the active runtime and adjacent plugin infrastructure:

- `client/src/pages/Playground.tsx`
- `client/src/lib/pluginManager.ts`
- `client/src/store/store.ts`
- `client/src/lib/presetPlugins.ts`
- `client/src/lib/presets.ts`
- `client/src/lib/pluginSandboxClient.ts`
- `client/src/workers/pluginSandbox.worker.ts`
- `client/src/components/WidgetRenderer.tsx`
- `client/src/components/PluginWidget.tsx`
- `client/src/components/PluginList.tsx`
- `client/src/lib/uiTypes.ts`
- `client/src/App.tsx`
- `client/src/pages/Home.tsx`

### Runtime Validation Performed

- `pnpm install --frozen-lockfile`
- `pnpm -s check`
- `pnpm -s build`

Build/typecheck pass after dependency install. This confirms current code compiles, but compile success does not imply runtime scope safety.

## System Overview (How It Works Today)

### High-Level Stack

- UI shell: React + Wouter.
- State: Redux Toolkit (`store.ts`).
- Plugin execution (active path): `pluginManager` using `new Function` directly in main thread.
- Plugin UI contract: data-only nodes (`UINode`) rendered by `WidgetRenderer`.

### Active Route and Runtime Path

`client/src/App.tsx` routes `/` to `Home`, which returns `Playground`.

`Playground` performs all plugin lifecycle operations locally:

- Loads preset code from `presetPlugins.ts`.
- Calls `pluginManager.loadPlugin(...)`.
- Keeps loaded plugin IDs in local component state (`loadedPlugins`).
- Calls plugin widget handlers through `pluginManager.callHandler(...)`.
- Re-renders widgets by reading Redux state and invoking plugin `render(...)` directly.

### Dormant/Alternate Runtime Path

There is a second architecture that looks intended for sandboxed execution:

- `PluginSandboxClient`
- `pluginSandbox.worker.ts`
- `PluginWidget`
- `PluginList`

This path is largely not wired from `Playground`. Contracts in this path drift from active path.

## Runtime Flow Trace (Current)

### 1. Plugin Load Flow (Preset)

From `client/src/pages/Playground.tsx`:

```ts
const preset = presetPlugins.find((p) => p.id === presetId);
await pluginManager.loadPlugin(preset.code, { ui: uiBuilder, createActions });
setLoadedPlugins((prev) => (!prev.includes(presetId) ? [...prev, presetId] : prev));
```

Key points:

- Preset selected by `preset.id`.
- Plugin manager stores plugin by plugin-defined `id` from code output.
- UI tracks loaded plugins by preset ID, not necessarily plugin-defined ID.

### 2. Render Flow

From `Playground`:

```ts
const plugin = pluginManager.getPlugin(pluginId);
const tree = widget.render({ state });
<WidgetRenderer tree={tree} onEvent={(eventRef) => handleEvent(pluginId, widgetId, eventRef)} />
```

Key points:

- Plugin render receives full Redux state.
- No scope filtering on what part of state is visible.

### 3. Event/Dispatch Flow

```ts
pluginManager.callHandler(pluginId, widgetId, eventRef.handler, dispatch, eventRef.args, state);
```

In manager:

```ts
handler({ dispatch, state }, args);
```

Key points:

- Handler gets unrestricted Redux dispatch and full state.
- No host validation of action type, namespace, or target scope.

### 4. Reducer Processing

`store.ts` uses a matcher:

```ts
(action) => action.type.startsWith("plugin.")
```

Then hardcoded if/else branches for specific action types such as:

- `plugin.counter/incremented`
- `plugin.calculator/equals`
- `plugin.greeter/nameChanged`

Key points:

- “Plugin actions” are global by convention.
- Reducer behavior is static and host-hardcoded.
- No plugin instance awareness.

## Current State and Identity Model

### Current State Shape

`plugins` slice contains:

- `plugins: Record<string, LoadedPlugin>` (metadata/lifecycle)
- `counter`, `calculator`, `greeter` (plugin domain states)

Store root also has a separate top-level `counter` reducer, creating duplicate conceptual domain names.

### Current Plugin Identity Inputs

There are multiple identity sources:

1. Preset catalog ID (`preset.id`) from `presetPlugins.ts`.
2. Plugin-declared ID (`plugin.id`) returned by plugin code.
3. Loader input ID (`pluginId`) in `PluginSandboxClient.loadPlugin(pluginId, code)`.

No single canonical identity model is enforced across all paths.

## Detailed Findings

## 1) Ambiguous Plugin Identity Authority

### Problem

Plugin IDs come from multiple channels, and authority is inconsistent between runtime paths. This creates collision, mismatch, and spoofing potential.

### Where to Look

- `client/src/pages/Playground.tsx`
- `client/src/lib/pluginManager.ts`
- `client/src/lib/pluginSandboxClient.ts`

### Example

```ts
// pluginManager stores by plugin-defined id
this.plugins.set((plugin as PluginInstance).id, plugin as PluginInstance);

// Playground tracks loaded by preset id
setLoadedPlugins((prev) => (!prev.includes(presetId) ? [...prev, presetId] : prev));
```

### Why It Matters

- If preset ID and plugin-declared ID diverge, lookups fail or become non-deterministic.
- Two plugins can claim same ID and overwrite each other.
- Host cannot reliably bind state/action scopes to runtime instance.

### Cleanup Sketch

```ts
// host-owned identity
interface PluginInstanceRef {
  instanceId: string;      // authoritative runtime key
  packageId: string;       // logical plugin type/catalog id
  declaredId?: string;     // metadata only
}

registry.add({ instanceId, packageId, declaredId, ... });
```

## 2) Action Namespace Is Cooperative, Not Enforced

### Problem

`createActions(namespace, names)` produces action types but does not enforce namespace ownership. Plugins can dispatch arbitrary action types.

### Where to Look

- `client/src/pages/Playground.tsx`
- `client/src/lib/pluginSandboxClient.ts`
- `client/src/store/store.ts`

### Example

```ts
const createActions = (namespace: string, actionNames: string[]) => {
  actions[name] = (payload?: any) => ({ type: `${namespace}/${name}`, payload });
};
```

And dispatch is raw Redux dispatch:

```ts
const dispatch = (action: any) => {
  this.store.dispatch(action);
};
```

### Why It Matters

- No defense against plugins writing into unrelated plugin domains.
- No way to audit local vs shared intent cleanly.
- Shared-state access can become accidental coupling.

### Cleanup Sketch

```ts
dispatchGateway(instanceId, action) {
  const verdict = policyEngine.validate(instanceId, action);
  if (!verdict.ok) throw new Error(verdict.reason);
  store.dispatch(attachMeta(instanceId, verdict.scope, action));
}
```

## 3) Plugin State Is Global/Hardcoded, Not Instance-Scoped

### Problem

Plugin domain states (`counter`, `calculator`, `greeter`) are hardcoded in host reducer. This prevents multiple instances and dynamic plugin domains.

### Where to Look

- `client/src/store/store.ts`

### Example

```ts
interface PluginsState {
  plugins: Record<string, LoadedPlugin>;
  counter: number;
  calculator: { display: string; accumulator: number; operation: string | null };
  greeter: { name: string };
}
```

### Why It Matters

- Cannot run two independent instances of the same plugin type.
- New plugins require host reducer modifications.
- “Local plugin state” and “shared state” are indistinguishable in structure.

### Cleanup Sketch

```ts
interface PluginRuntimeState {
  instances: Record<string, {
    packageId: string;
    local: unknown;
    ui: { enabled: boolean; status: PluginStatus };
  }>;
  shared: {
    counter: CounterState;
    workspace: WorkspaceState;
  };
}
```

## 4) Full Root State Is Exposed to Every Plugin

### Problem

Plugins receive complete Redux root state during render/event handling.

### Where to Look

- `client/src/pages/Playground.tsx`
- `client/src/lib/pluginManager.ts`

### Example

```ts
const tree = widget.render({ state });
handler({ dispatch, state }, args);
```

### Why It Matters

- No principle-of-least-privilege boundary.
- Plugins can accidentally depend on internal host details.
- Refactors become risky due undocumented plugin reads.

### Cleanup Sketch

```ts
const pluginView = {
  local: selectLocalState(rootState, instanceId),
  shared: selectAllowedSharedState(rootState, capabilities.read),
  system: selectSystemMeta(rootState),
};
widget.render({ state: pluginView });
```

## 5) Active Runtime Is Not Sandboxed

### Problem

Plugins execute in main thread with `new Function`, no isolation, no resource governance.

### Where to Look

- `client/src/lib/pluginManager.ts`

### Example

```ts
const fn = new Function("definePlugin", code);
fn(definePlugin);
```

### Why It Matters

- Plugin code can touch `window`, network, storage, and globals.
- A bad plugin can freeze UI thread.
- “QuickJS VM” messaging in UI does not match actual active path.

### Cleanup Sketch

- Either harden worker/isolated runtime and make it primary.
- Or explicitly document “trusted plugin only” and remove sandbox claims.

## 6) Alternate Runtime Path Has Contract Drift

### Problem

Worker path uses node objects with `type` field, but renderer expects `kind`.

### Where to Look

- `client/src/workers/pluginSandbox.worker.ts`
- `client/src/lib/uiTypes.ts`
- `client/src/components/WidgetRenderer.tsx`

### Example

```ts
// worker
text: (text: string) => ({ type: "text", text })

// renderer
switch (node.kind) {
  case "text":
```

### Why It Matters

- If worker path is re-enabled, UI nodes will fail to render without translation.
- Indicates parallel evolution without contract tests.

### Cleanup Sketch

```ts
// single canonical UI contract
type UINode = { kind: ... };
// all runtimes must emit exact UINode schema (validated)
```

## 7) Unused Security Hook (`allowDispatch`) Indicates Incomplete Policy Layer

### Problem

`PluginSandboxClient` constructor accepts `allowDispatch` but never uses it.

### Where to Look

- `client/src/lib/pluginSandboxClient.ts`

### Example

```ts
constructor({ store, workerUrl, allowDispatch }: { ... allowDispatch?: (...) => boolean }) {
  this.store = store;
}
```

### Why It Matters

- Intended dispatch policy gate exists conceptually but not implemented.
- Gives false confidence that action control exists.

### Cleanup Sketch

```ts
if (allowDispatch && !allowDispatch(instanceId, action)) {
  throw new Error(`Dispatch denied for ${instanceId}: ${action.type}`);
}
store.dispatch(action);
```

## 8) Two Preset Catalogs and Two Event Signatures Increase Drift

### Problem

`presets.ts` and `presetPlugins.ts` both define plugin samples but with different assumptions about handler args and state shape.

### Where to Look

- `client/src/lib/presets.ts`
- `client/src/lib/presetPlugins.ts`

### Why It Matters

- Developer confusion about canonical plugin API.
- Harder debugging when sample code implies incompatible patterns.

### Cleanup Sketch

- Keep one canonical preset catalog.
- Add schema tests to validate all presets against runtime contracts.

## 9) Plugin Lifecycle Source of Truth Is Split

### Problem

`Playground` uses component-local `loadedPlugins` while store has plugin lifecycle slice (`plugins.plugins`) that active route mostly bypasses.

### Where to Look

- `client/src/pages/Playground.tsx`
- `client/src/store/store.ts`

### Why It Matters

- UI registry and state registry can diverge.
- Hard to support cross-page plugin management later.

### Cleanup Sketch

- Move authoritative plugin registry to Redux runtime slice.
- Components render from selectors only.

## 10) No Stable Local-vs-Shared Contract for Plugin Authors

### Problem

Plugin authors currently infer state shape ad hoc (examples read from different paths).

### Where to Look

- `client/src/lib/presetPlugins.ts`
- `client/src/lib/presets.ts`

### Why It Matters

- High accidental coupling.
- Breakage when host state layout changes.

### Cleanup Sketch

Expose explicit plugin context:

```ts
render({ state: { local, shared, system } })
```

with declared capabilities controlling available `shared` domains.

## Assessment: Why You Are Seeing Plugin ID and Scope Problems

Your issue is structural, not a one-off bug.

The current architecture mixes:

- A prototype global reducer model.
- A mostly local component registry.
- Two different runtime strategies.
- Conventions instead of enforceable scope policy.

This combination naturally creates plugin ID drift and scope leakage. The system can “work” for demos but will become brittle as plugin count/complexity grows.

## Target Architecture

## Design Goals

1. Host-authoritative identity.
2. Strict local isolation by default.
3. Explicit shared domains and shared actions.
4. Clear plugin author contract.
5. Auditable dispatch and state access.
6. Single runtime path and contract.

## 1) Identity Model

Use three IDs with clear authority:

- `packageId`: logical plugin kind (catalog key), e.g. `counter`.
- `instanceId`: host-generated runtime ID, e.g. `counter@a1b2c3d4`.
- `declaredId`: plugin self-declared metadata, informational only.

Rules:

- All internal maps keyed by `instanceId`.
- All action metadata stamped with `instanceId`.
- Widgets keyed by `{instanceId, widgetKey}`.
- `declaredId` cannot overwrite registry keys.

## 2) State Partition Model

```ts
interface RootState {
  pluginRuntime: {
    instances: Record<string, {
      packageId: string;
      status: "loading" | "loaded" | "error";
      enabled: boolean;
      local: unknown;
      capabilities: PluginCapabilities;
      widgets: string[];
      errors?: string[];
    }>;
    shared: {
      counter: CounterSharedState;
      workspace: WorkspaceSharedState;
      // more host-defined shared domains
    };
  };
  app: {
    theme: string;
    router: { path: string };
  };
}
```

Policy:

- `instances[instanceId].local` is private to that plugin instance.
- `shared.*` is accessible only if declared + approved capability.
- Host `app` internals are never directly exposed.

## 3) Action Scope Model

Action envelope:

```ts
interface PluginActionEnvelope<T = unknown> {
  type: string;
  payload?: T;
  meta: {
    source: "plugin";
    instanceId: string;
    packageId: string;
    scope: "local" | "shared" | "system";
    domain?: string;  // required for shared
    timestamp: number;
  };
}
```

Action creation APIs exposed to plugins:

- `ctx.actions.local("setName")` -> type auto-prefixed to local scope.
- `ctx.actions.shared("counter", "increment")` -> only if capability allows.
- No raw arbitrary `dispatch` from plugin code.

## 4) Capability Model

Plugin manifest includes requested capabilities:

```ts
interface PluginCapabilities {
  readShared?: string[];   // domains
  writeShared?: string[];  // domains
  dispatchSystem?: string[]; // explicit host commands
}
```

Host policy decides grant set (could be static for now).

At runtime:

- Selector composer only exposes granted shared domains.
- Dispatch gateway validates write scopes against grant set.

## 5) Plugin Context Contract

```ts
interface PluginRuntimeContext {
  ui: UIBuilder;
  state: {
    local: unknown;
    shared: Record<string, unknown>;
    system: { instanceId: string; packageId: string; now: number };
  };
  actions: {
    local: (name: string) => (payload?: unknown) => PluginActionEnvelope;
    shared: (domain: string, name: string) => (payload?: unknown) => PluginActionEnvelope;
    command: (name: string) => (payload?: unknown) => PluginActionEnvelope;
  };
}
```

No unscoped `dispatch` should be directly available.

## 6) Reducer Architecture

Host reducers:

- `pluginRuntime.instances` reducer for lifecycle and local state updates.
- Shared-domain reducers registered by host per domain.

Flow:

1. Plugin emits envelope action.
2. Gateway validates scope and permissions.
3. Reducer routes:
   - `scope=local` -> instance local reducer.
   - `scope=shared` -> domain reducer.
   - `scope=system` -> command handler.

Pseudo-router:

```ts
function pluginActionRouter(state: RootState, action: PluginActionEnvelope) {
  if (action.meta.scope === "local") {
    const inst = state.pluginRuntime.instances[action.meta.instanceId];
    inst.local = localReducers[inst.packageId](inst.local, action);
    return;
  }
  if (action.meta.scope === "shared") {
    const d = action.meta.domain!;
    state.pluginRuntime.shared[d] = sharedReducers[d](state.pluginRuntime.shared[d], action);
    return;
  }
  handleSystemCommand(state, action);
}
```

## 7) Rendering API

Each plugin render receives curated state view:

```ts
const renderCtx = {
  state: {
    local: selectLocal(instanceId),
    shared: selectGrantedShared(instanceId),
    system: { instanceId, packageId, now: Date.now() },
  },
};
```

No direct root store object leakage.

## Concrete Implementation Plan

## Phase 0: Contract Clarification (No Behavior Change)

1. Pick one canonical plugin API doc and sample set.
2. Mark `presets.ts` or `presetPlugins.ts` as canonical, deprecate other.
3. Add comments/README clarifying current trusted-runtime status.

Deliverables:

- `docs/plugin-api.md` (or ticket design doc appendix).
- One source of sample truth.

## Phase 1: Host Identity Refactor

1. Add `instanceId` generation at load.
2. Store registry keyed by `instanceId`.
3. Keep plugin-declared ID as metadata.
4. Convert widget rendering to use `instanceId` keys.

Suggested modules:

- `client/src/lib/pluginIdentity.ts`
- `client/src/lib/pluginRegistry.ts`

## Phase 2: Dispatch Gateway and Envelope

1. Introduce `dispatchPluginAction(instanceId, action)`.
2. Stamp metadata and validate against capability grants.
3. Remove direct raw dispatch from plugin context.

Suggested modules:

- `client/src/lib/pluginDispatchGateway.ts`
- `client/src/lib/pluginPolicy.ts`

## Phase 3: State Partitioning

1. Create `pluginRuntime` slice:
   - `instances[instanceId].local`
   - `shared` domain object
2. Migrate hardcoded plugin states (`counter`, `calculator`, `greeter`) into:
   - local domain when purely plugin-owned
   - shared domain when intentionally shared
3. Replace `action.type.startsWith("plugin.")` matcher with envelope router.

Suggested modules:

- `client/src/store/pluginRuntimeSlice.ts`
- `client/src/store/pluginSharedReducers.ts`
- `client/src/store/pluginLocalReducers.ts`

## Phase 4: Curated Plugin State View

1. Add selectors that build plugin view by capabilities.
2. Render with `{ local, shared, system }` shape.
3. Remove root state passthrough to plugin render handlers.

Suggested modules:

- `client/src/lib/pluginSelectors.ts`

## Phase 5: Runtime Unification

1. Choose one execution path:
   - If trusted runtime: keep in-process and drop dormant worker code.
   - If sandbox needed: finish worker path and enforce UI contract tests.
2. Delete incompatible path or gate behind experimental flag.

## Phase 6: Compatibility Layer and Migration

1. Add adapter for old plugins that dispatch string action types.
2. Translate old action types into envelope model with warnings.
3. Sunset adapter after migration period.

## File-by-File Recommended Changes

### `client/src/pages/Playground.tsx`

- Replace local `loadedPlugins` source-of-truth with selectors from runtime slice.
- Render plugin cards by instance records.
- Remove direct `pluginManager.getPlugin(id)` keyed by non-authoritative IDs.

### `client/src/lib/pluginManager.ts`

- Change API:
  - `loadPlugin(code, packageId?) -> instanceRecord`
  - internal map keyed by `instanceId`
- Pass `dispatch` via gateway only.
- Remove raw state exposure; use selector-produced view.

### `client/src/store/store.ts`

- Move plugin runtime logic into dedicated slice module.
- Remove hardcoded plugin action matcher.
- Remove duplicate top-level `counter` reducer unless intentionally host-owned.

### `client/src/lib/presetPlugins.ts`

- Add manifest block with capability declarations.
- Move handlers to scoped action API.

### `client/src/lib/pluginSandboxClient.ts` and `client/src/workers/pluginSandbox.worker.ts`

- Either finish and promote to primary runtime, or archive/remove.
- If kept:
  - use same `UINode` schema (`kind`)
  - ensure `allowDispatch` is implemented
  - add strict message schema validation.

## Proposed Type Contracts (Reference)

### Manifest and Definition

```ts
interface PluginManifest {
  packageId: string;
  title: string;
  version?: string;
  capabilities?: PluginCapabilities;
}

interface PluginDefinition {
  manifest: PluginManifest;
  widgets: Record<string, WidgetDefinition>;
  initLocalState?: () => unknown;
  reduceLocal?: (state: unknown, action: PluginActionEnvelope) => unknown;
}
```

### Action Helper API

```ts
function createScopedActions(instance: {
  instanceId: string;
  packageId: string;
  caps: PluginCapabilities;
}) {
  return {
    local(name: string) {
      return (payload?: unknown): PluginActionEnvelope => ({
        type: `plugin.local/${name}`,
        payload,
        meta: { source: "plugin", instanceId: instance.instanceId, packageId: instance.packageId, scope: "local", timestamp: Date.now() },
      });
    },
    shared(domain: string, name: string) {
      if (!instance.caps.writeShared?.includes(domain)) throw new Error(`shared domain denied: ${domain}`);
      return (payload?: unknown): PluginActionEnvelope => ({
        type: `plugin.shared.${domain}/${name}`,
        payload,
        meta: { source: "plugin", instanceId: instance.instanceId, packageId: instance.packageId, scope: "shared", domain, timestamp: Date.now() },
      });
    },
  };
}
```

## Local + Shared Example

### Counter Plugin (Local State)

- Local domain: `instances[instanceId].local = { value: number }`
- Actions:
  - local/increment
  - local/decrement
  - local/reset

### Dashboard Plugin (Shared Read)

- Declares `readShared: ["counter"]`
- Renders aggregate values from shared counter domain.
- No write permission unless explicitly granted.

### Collaborative Plugin (Shared Write)

- Declares `writeShared: ["workspace"]`
- Can emit scoped shared actions in `workspace` domain only.
- Cannot mutate unrelated shared domains.

## Security and Operational Considerations

### In-Process Runtime Risk

If plugin code remains in-process with `new Function`, isolation guarantees do not exist. If this is acceptable (trusted plugins only), document it explicitly and remove sandbox language in UI.

If untrusted plugins are intended, move execution to isolated runtime and block direct global APIs.

### Telemetry and Auditing

Add plugin action logs:

- `instanceId`
- `packageId`
- `scope`
- `type`
- decision (allowed/denied)

This will shorten debugging loops for scope issues.

### Versioning Strategy

Support plugin API versions:

- `manifest.apiVersion = 1`
- runtime provides compatibility adapters per major version.

Prevents silent breakage during contract evolution.

## Testing Strategy

## Unit Tests

1. Identity:
   - Generates unique `instanceId`.
   - Registry keyed by instance, not declared ID.
2. Dispatch policy:
   - Deny ungranted shared domains.
   - Deny system commands without capability.
3. Local reducers:
   - Local action only affects matching instance state.
4. Shared reducers:
   - Shared action affects only declared domain.

## Integration Tests

1. Load two instances of same package and verify isolated local state.
2. Verify shared read-only plugin cannot write shared domain.
3. Verify plugin cannot read ungranted shared domain.
4. Verify plugin action envelope metadata always attached.

## Contract Tests

1. Validate all preset plugins against runtime schema.
2. Validate all runtime outputs use `UINode.kind` contract.
3. Validate handler signatures and event payload shape consistency.

## Migration Risks and Mitigations

### Risk: Existing Plugins Break

Mitigation:

- Provide compatibility adapter translating old action style.
- Emit runtime warnings and migration hints.

### Risk: Increased Boilerplate for Plugin Authors

Mitigation:

- Provide tiny helper SDK (`createScopedActions`, typed context).
- Provide sample templates.

### Risk: Feature Freeze During Refactor

Mitigation:

- Phase rollout with adapter layer.
- Keep existing demos functioning while migrating one preset at a time.

### Risk: Shared State Becomes New Global Dumping Ground

Mitigation:

- Strict domain ownership and capability checks.
- Require review for adding new shared domain.

## Alternatives Considered

## A) Keep Current Model, Just Prefix Better

Description:

- Keep global state and free-form action dispatch.
- Encourage naming conventions like `plugin.<id>/<action>`.

Why Rejected:

- Conventions do not solve authority or enforcement.
- Still allows cross-plugin mutation and accidental coupling.

## B) One Redux Slice per Plugin Package

Description:

- Host creates static slice per package.

Why Rejected:

- Fails with dynamic plugin loading.
- Does not handle multiple instances cleanly without additional indexing.

## C) Fully Isolated Store per Plugin + Event Bus Only

Description:

- Each plugin has private store; shared interactions via bus.

Why Deferred:

- Strong isolation but higher complexity and tooling overhead for current stage.
- Proposed envelope model gives sufficient control first with lower migration cost.

## D) Complete Worker/QuickJS Rewrite Immediately

Description:

- Replace active path with worker/QuickJS now.

Why Deferred:

- Valuable long-term for untrusted plugins, but not prerequisite for fixing identity/scope contracts.
- Scope correctness should be solved first at architecture contract level.

## Decision Summary

Chosen path (v1): **Host-authoritative plugin identity + dual selector model (`selectPluginState`, `selectGlobalState`) + dual action model (`dispatchPluginAction`, `dispatchGlobalAction`) + global `dispatchId` tracing**.

Rationale:

- Directly addresses your stated pain points.
- Keeps implementation effort low while fixing the core scope model.
- Still allows a future upgrade to capability-gated permissions.

## Practical “First 10 Days” Execution Plan

1. Day 1-2: Add authoritative `pluginId` handling and registry cleanup.
2. Day 3-4: Add `selectPluginState(pluginId)` and `selectGlobalState()` selectors.
3. Day 5-6: Add `dispatchPluginAction` and `dispatchGlobalAction` wrappers with `dispatchId`.
4. Day 7: Add global action allowlist and unknown-plugin rejection.
5. Day 8: Migrate one preset plugin (counter) to the new API.
6. Day 9: Migrate remaining presets.
7. Day 10: Remove dead paths and finalize docs/tests.

## Assessment of Current System Maturity

### What Is Strong Today

- Plugin UI DSL approach is practical and understandable.
- React rendering bridge is straightforward.
- Demo presets communicate concept quickly.
- Build and typecheck baseline is healthy once deps are installed.

### What Is Limiting

- Identity model is not authoritative.
- Scoping model is implicit and unenforced.
- Runtime boundaries are overstated (sandbox language vs main-thread eval).
- Architecture drift from duplicate pathways and duplicate preset sources.

### Overall Assessment

The system is a good proof-of-concept foundation with clear potential, but scope control and identity consistency need a contract-first refactor before scaling plugin count, adding third-party plugins, or depending on shared-state correctness.

## Open Questions

1. Are plugins intended to be trusted internal code only, or eventually untrusted/third-party?
2. Should shared domains be static (host-defined) or dynamically registered by plugins under governance?
3. Do you need persistence for plugin local state across sessions?
4. Should plugin instances be user-created multiple times (same package, many instances)?
5. Is hot-reload/edit loop expected to preserve local state or reset on code change?

## References (Reviewed Sources)

- `client/src/pages/Playground.tsx`
- `client/src/lib/pluginManager.ts`
- `client/src/store/store.ts`
- `client/src/lib/presetPlugins.ts`
- `client/src/lib/presets.ts`
- `client/src/lib/pluginSandboxClient.ts`
- `client/src/workers/pluginSandbox.worker.ts`
- `client/src/components/WidgetRenderer.tsx`
- `client/src/components/PluginWidget.tsx`
- `client/src/components/PluginList.tsx`
- `client/src/lib/uiTypes.ts`
- `client/src/App.tsx`
- `client/src/pages/Home.tsx`

## Appendix A: Concrete Migration Example (Counter)

### Current Style (Representative)

```ts
const actions = createActions("plugin.counter", ["incremented", "decremented", "reset"]);
increment({ dispatch }) {
  dispatch(actions.incremented());
}
```

### Target Style

```ts
const actions = ctx.actions;
increment({ dispatch }) {
  dispatch(actions.local("increment")());
}
```

Local reducer for package `counter`:

```ts
function reduceCounterLocal(state = { value: 0 }, action: PluginActionEnvelope) {
  if (action.meta.scope !== "local") return state;
  if (action.type === "plugin.local/increment") return { value: state.value + 1 };
  if (action.type === "plugin.local/decrement") return { value: state.value - 1 };
  if (action.type === "plugin.local/reset") return { value: 0 };
  return state;
}
```

## Appendix B: Suggested Directory Layout

```text
client/src/
  lib/
    pluginIdentity.ts
    pluginRegistry.ts
    pluginDispatchGateway.ts
    pluginPolicy.ts
    pluginSelectors.ts
    pluginContracts.ts
  store/
    pluginRuntimeSlice.ts
    pluginLocalReducers.ts
    pluginSharedReducers.ts
  plugins/
    presets/
      counter.ts
      dashboard.ts
```

## Appendix C: Minimal Acceptance Criteria

1. Two instances of same plugin package can run concurrently with isolated local state.
2. Shared writes are blocked unless capability explicitly granted.
3. Plugins receive only `{ local, shared(granted), system }` view; no full root state.
4. All plugin actions include `meta.instanceId` and `meta.scope`.
5. Runtime uses one canonical UI node contract and one primary execution path.

## Deep Dive: Current Runtime Semantics by Concern

### Identity Semantics in Active Path

In the active path, plugin identity is effectively derived from whichever string was used in the current step, not a stable identity object:

1. The preset list is keyed by `preset.id` in `presetPlugins.ts`.
2. The plugin code returns its own `id` field inside `definePlugin` return object.
3. `pluginManager` stores plugin by plugin return `id`.
4. `Playground` UI keeps an array of IDs from the load path (`presetId` for presets, `plugin.id` for custom code).

This works only if values happen to stay aligned by convention. There is no authority hierarchy.

### Action Semantics in Active Path

Actions are plain Redux actions with string types.

- `createActions` is a helper only.
- Plugins can dispatch direct literal types bypassing helper.
- Store matcher routes by string compare.

Action “scope” is therefore inferred from the string prefix and reducer implementation, not from validated metadata.

### State Semantics in Active Path

Every render and handler sees the full root state object. There is no explicit plugin-local vs shared split.

The practical consequences:

- Any plugin can read all host state fields.
- Plugin code can become coupled to incidental host internals.
- Local and shared concerns cannot be audited from interface alone.

### Lifecycle Semantics in Active Path

There are two lifecycle representations:

- Local React state in `Playground` (`loadedPlugins`).
- Redux plugin metadata slice (`plugins.plugins`) intended for lifecycle, largely not driving the active route.

This implies the project does not yet have a single runtime source of truth for plugin lifecycle.

## Failure Mode Matrix: Plugin IDs

| Failure Mode | Trigger | Current Behavior | Impact | Detectability |
|---|---|---|---|---|
| Preset ID != plugin-declared ID | Plugin code changed but preset id not updated | UI stores one id, manager stores another | Widget not found / stale UI entries | Medium |
| Duplicate plugin-declared IDs | Two plugins return same `id` | Map overwrite in `pluginManager` | Existing plugin silently replaced | Low |
| Caller-supplied ID mismatch in sandbox client | `loadPlugin(pluginId, code)` uses different internal `def.id` | stored by caller ID but metadata returns def.id | follow-up render/event mismatch | Medium |
| Plugin spoofing another plugin ID | Malicious or mistaken `id` value | accepted and stored | action/state attribution confusion | Low |
| Multiple instances same package | user loads same package twice | currently dedup by id conventions | no independent state instances | High |

The key pattern is that identity collisions are silent and structural rather than explicit failures.

## Failure Mode Matrix: Action Scope

| Failure Mode | Trigger | Current Behavior | Impact |
|---|---|---|---|
| Cross-plugin write | Plugin dispatches another domain action type | accepted if reducer handles type | plugin isolation broken |
| Host internals write | Plugin dispatches host/system action type | likely accepted unless reducer ignores | unintended host mutations |
| Shared domain overwrite | Plugin dispatches shared action without authorization | no capability checks | shared state corruption |
| Action type typos | String mismatch | silently ignored or misrouted | debugging pain |
| Replay/stale actions | no instance metadata | hard to correlate origin | audit and rollback difficult |

## Failure Mode Matrix: State Scope

| Failure Mode | Trigger | Current Behavior | Impact |
|---|---|---|---|
| Hidden coupling to host schema | Plugin reads arbitrary root fields | allowed | host refactors break plugins |
| Privacy leak between plugins | Plugin reads shared root fields not intended for it | allowed | data exposure |
| No instance-locality | same package loaded twice | shared hardcoded state fields | state collisions |
| Hardcoded domain growth | add new plugin | host reducer edited each time | scaling friction |

## Recommended Contract: Scope Taxonomy

To avoid ambiguity, scope should be formalized as a first-class runtime concept.

### Scope Types

1. `local`
- State owned by one plugin instance.
- Read/write only by that instance.

2. `shared:<domain>`
- State shared across approved plugin instances.
- Read and write governed by capabilities per domain.

3. `system:<command>`
- Commands targeting host/runtime functions.
- Allowlist only, explicit capability required.

### Scope Naming Rules

- Never encode scope only in free-form string prefixes.
- Keep scope in `meta.scope`, domain in `meta.domain`.
- Keep raw `type` for domain operation semantics only.

## Proposed Runtime Data Model (Detailed)

```ts
interface PluginInstanceRuntime {
  instanceId: string;
  packageId: string;
  declaredId?: string;
  title: string;
  status: "loading" | "loaded" | "error";
  enabled: boolean;
  widgets: string[];
  localState: unknown;
  capabilities: PluginCapabilities;
  diagnostics: {
    loadErrors: string[];
    deniedActions: number;
    lastRenderMs?: number;
  };
}

interface PluginRuntimeRoot {
  instances: Record<string, PluginInstanceRuntime>;
  packageIndex: Record<string, string[]>; // packageId -> instanceIds
  shared: Record<string, unknown>;
  telemetry: {
    actionCount: number;
    deniedActionCount: number;
    lastActionAt?: number;
  };
}
```

This shape directly supports:

- multiple instances per package,
- local state isolation,
- governed shared domains,
- action attribution and observability.

## Reducer/Router Design (Detailed)

### Why a Router

A router reducer centralizes policy decisions and avoids dozens of unstructured matcher branches.

### Router Algorithm

1. Assert action envelope validity.
2. Resolve instance record.
3. Validate capability for scope/domain/command.
4. Route to local/shared/system handler.
5. Emit diagnostics counters.

Pseudo:

```ts
function reducePluginEnvelope(state: PluginRuntimeRoot, action: PluginActionEnvelope) {
  if (!isPluginEnvelope(action)) return;

  const instance = state.instances[action.meta.instanceId];
  if (!instance) return recordDenied(state, action, "unknown-instance");

  const decision = policyDecision(instance, action);
  if (!decision.ok) return recordDenied(state, action, decision.reason);

  switch (action.meta.scope) {
    case "local":
      state.instances[instance.instanceId].localState = reduceLocal(instance.packageId, instance.localState, action);
      break;
    case "shared":
      state.shared[action.meta.domain!] = reduceShared(action.meta.domain!, state.shared[action.meta.domain!], action);
      break;
    case "system":
      reduceSystem(state, instance, action);
      break;
  }

  state.telemetry.actionCount += 1;
  state.telemetry.lastActionAt = Date.now();
}
```

### Local Reducer Registry

Use package ID keyed reducer registry:

```ts
const localReducers: Record<string, (s: unknown, a: PluginActionEnvelope) => unknown> = {
  counter: reduceCounterLocal,
  calculator: reduceCalculatorLocal,
  greeter: reduceGreeterLocal,
};
```

This allows dynamic plugin package loading while preserving strict local instance routing.

## Selector Design (Detailed)

### Curated State View

Plugin should not receive root state. Instead:

```ts
interface PluginStateView {
  local: unknown;
  shared: Record<string, unknown>;
  system: {
    instanceId: string;
    packageId: string;
    runtimeVersion: string;
    now: number;
  };
}
```

### Selector Pipeline

1. `selectInstance(instanceId)`
2. `selectLocal(instanceId)`
3. `selectGrantedShared(instanceId)`
4. `selectPluginStateView(instanceId)`

Pseudo:

```ts
function selectPluginStateView(root: RootState, instanceId: string): PluginStateView {
  const inst = root.pluginRuntime.instances[instanceId];
  const grantedShared = Object.fromEntries(
    (inst.capabilities.readShared ?? []).map((d) => [d, root.pluginRuntime.shared[d]])
  );

  return {
    local: inst.localState,
    shared: grantedShared,
    system: {
      instanceId,
      packageId: inst.packageId,
      runtimeVersion: "1.0.0",
      now: Date.now(),
    },
  };
}
```

## Dispatch Gateway Design (Detailed)

### Gateway Responsibilities

- Attach source metadata.
- Validate shape.
- Enforce policy.
- Dispatch if allowed.
- Emit structured denial diagnostics if blocked.

### Denial Behavior

Denials should be explicit and observable.

- Plugin receives error from handler.
- Runtime records denial telemetry with reason.
- Optional dev console warning with instance/package/action summary.

Example denial record:

```json
{
  "kind": "dispatch_denied",
  "instanceId": "counter@a1b2c3d4",
  "packageId": "counter",
  "scope": "shared",
  "domain": "workspace",
  "type": "plugin.shared.workspace/setTitle",
  "reason": "capability-write-shared-missing",
  "timestamp": 1770595000000
}
```

## Plugin Author API Design

### Manifest Proposal

```ts
interface PluginManifestV1 {
  apiVersion: 1;
  packageId: string;
  title: string;
  description?: string;
  capabilities?: {
    readShared?: string[];
    writeShared?: string[];
    systemCommands?: string[];
  };
}
```

### Define Function Proposal

```ts
definePlugin((ctx) => ({
  manifest: {
    apiVersion: 1,
    packageId: "counter",
    title: "Counter",
    capabilities: {
      readShared: ["counter"],
      writeShared: ["counter"],
    },
  },
  initLocalState: () => ({ value: 0 }),
  reduceLocal(state, action) { ... },
  widgets: { ... },
}));
```

### Important Constraint

Host may downgrade granted capabilities below requested capabilities. Requested != granted.

## Backward Compatibility Adapter

A compatibility adapter allows old plugins to continue while migrating.

### Adapter Responsibilities

1. Translate legacy `createActions("plugin.x", ["y"])` output into envelope actions.
2. Infer scope from namespace mapping table.
3. Warn when unknown namespace encountered.

### Namespace Mapping Example

```ts
const LEGACY_NAMESPACE_MAP = {
  "plugin.counter": { scope: "shared", domain: "counter" },
  "plugin.greeter": { scope: "local" },
  "plugin.calculator": { scope: "local" },
};
```

### Adapter Warning Example

```text
[plugin-runtime][compat] Unknown legacy namespace "plugin.foo" from instance counter@a1b2c3d4.
Defaulting to local scope. Please migrate to ctx.actions.local/shared API.
```

## Runtime Path Unification Decision Tree

### Option 1: Trusted In-Process (Short-Term Pragmatic)

Use current in-process execution but enforce identity/scope architecture.

Pros:

- Lowest migration effort.
- Immediate fixes for your current problem.

Cons:

- No code isolation for untrusted plugins.

### Option 2: Worker-Isolated Runtime (Medium-Term)

Promote worker path after contract alignment.

Required fixes before promotion:

1. Node schema parity (`kind` vs `type`).
2. Event payload parity.
3. Dispatch policy gateway integration.
4. Contract tests between worker output and renderer.

### Option 3: QuickJS/VM Runtime (Long-Term)

Best isolation but highest complexity.

Recommendation: do Option 1 immediately for scope correctness, then incrementally move to Option 2/3 if trust model demands it.

## Performance Considerations

### Current

- Every widget render receives full root state object.
- Potential expensive rerenders as app grows.

### Proposed

- Selector-built minimal view reduces serialization/read overhead.
- Instance-local updates can avoid unrelated plugin rerenders.
- Envelope metadata improves tracing without major CPU cost.

### Performance Guardrails

1. Memoize `selectPluginStateView(instanceId)`.
2. Keep shared domains granular.
3. Avoid deep cloning large state in selectors.
4. Track per-instance render duration.

## Operational Diagnostics Runbook

### Symptoms and Checks

#### Symptom: Plugin actions appear ignored

Check:

1. Is action denied by policy? (`deniedActionCount` increments)
2. Is action routed to local/shared correct domain?
3. Does plugin instance exist in registry?

#### Symptom: Plugin reads stale data

Check:

1. Selector memoization invalidation.
2. local/shared domain mapping for instance.
3. plugin render trigger source.

#### Symptom: Two plugin instances affect each other unexpectedly

Check:

1. Does action meta include correct `instanceId`?
2. Is reducer using instance-local path?
3. Is shared domain intentionally used?

### Minimum Observability Fields

For every action (allowed/denied), log structured fields:

- `action.type`
- `meta.instanceId`
- `meta.scope`
- `meta.domain`
- `decision`
- `reason` (if denied)

## Governance Model for Shared Domains

Shared state is necessary but dangerous without ownership rules.

### Proposed Governance Rules

1. Every shared domain has an owner (team/module).
2. Domain has documented read/write semantics.
3. New write permissions require explicit review.
4. Domain reducers are host-owned, not plugin-owned.

### Shared Domain Contract Example

```ts
interface SharedDomainContract {
  domain: string;
  owner: string;
  schemaVersion: number;
  readPolicy: "open" | "capability";
  writePolicy: "none" | "capability" | "owner-only";
}
```

## Concrete Example: Solving Your Exact Requirement

Requirement: scope actions/state per plugin, while allowing shared state and shared actions.

### Example Setup

- Two instances of `counter` package:
  - `counter@A`
  - `counter@B`
- Shared domain: `leaderboard`

### Behavior

- `counter@A` local increment affects only `counter@A.local.value`.
- `counter@B` local increment affects only `counter@B.local.value`.
- Both instances can read `shared.leaderboard` if granted read.
- Only instances with `writeShared: ["leaderboard"]` can dispatch leaderboard updates.

### Why This Works

- Instance identity is host-owned and immutable.
- Local action routing always keyed by instance ID.
- Shared writes are explicit scope+domain operations with policy checks.

## Suggested Migration of Current Presets

### Counter

- Move counter value to local state.
- Optionally publish snapshot to shared `counter-summary` domain.

### Status Dashboard

- Remove direct root reads.
- Grant read-only access to selected shared domains.

### Greeter

- Keep name local by default.
- Optional shared greeting board as separate capability.

### Calculator

- Keep all arithmetic state local.
- Shared calculation history only if capability granted.

### VM Monitor

- Should be system plugin with elevated read/system command capabilities.

## Code Review Notes on Existing Claims vs Reality

Several UX strings and comments imply sandboxed QuickJS behavior. Active path currently does not execute in QuickJS worker. This mismatch can confuse future maintainers and users.

Recommendation:

- Either align implementation to message, or
- change messaging to “in-process trusted plugin runtime”.

## Data Contract Validation Recommendations

Add runtime validation for plugin definitions and UINode output.

### Plugin Definition Validator

Validate:

- manifest presence
- required fields
- widget map shape
- handler function presence for referenced event handlers

### UINode Validator

Validate `kind`, props shape, and recursive children.

Validation protects host against malformed plugin outputs and lowers debugging cost.

## CI and Quality Gates

Add tests to prevent future drift:

1. Identity invariants test suite.
2. Policy denial/allow tests.
3. Renderer contract tests for all runtime paths.
4. Preset plugin contract snapshot tests.

Minimum failing conditions for CI:

- any action without required envelope meta,
- any plugin render receiving root state directly,
- any runtime producing non-canonical UI node shape.

## Implementation Checklist (Detailed)

### Identity and Registry

- [ ] Add `instanceId` generator and parser helpers.
- [ ] Refactor plugin maps to `instanceId` keys.
- [ ] Persist `declaredId` as metadata only.
- [ ] Add collision tests for declared IDs.

### Dispatch and Policy

- [ ] Implement dispatch gateway.
- [ ] Add capability policy engine.
- [ ] Add deny telemetry.
- [ ] Replace raw plugin dispatch exposure.

### State Partition

- [ ] Introduce `pluginRuntime` slice.
- [ ] Migrate plugin local state to instance-local.
- [ ] Define shared domain registry.
- [ ] Remove hardcoded matcher branches.

### Plugin API

- [ ] Define manifest v1.
- [ ] Add `ctx.actions.local/shared/command` helpers.
- [ ] Add compatibility adapter.
- [ ] Migrate presets one by one.

### Runtime Unification

- [ ] Pick primary runtime path.
- [ ] Archive/remove secondary drift path.
- [ ] Add contract tests across runtime boundary.

## Decision Log (Suggested)

1. **D-001**: Make host-assigned `instanceId` authoritative runtime key.
2. **D-002**: Enforce scoped action envelopes with mandatory metadata.
3. **D-003**: Split plugin state into `local` per instance and governed `shared` domains.
4. **D-004**: Remove direct root state exposure from plugin render/event contexts.
5. **D-005**: Unify runtime path and node contract before expanding plugin surface.

## Long-Term Evolution Path

### Stage 1: Scope Correctness (Now)

- Fix IDs, actions, state model while staying in-process.

### Stage 2: Runtime Hardening (Next)

- Move to worker-isolated path with same contracts.

### Stage 3: Trust Boundary Expansion

- Add stricter sandboxing and signed plugin bundles if external plugins are required.

### Stage 4: Marketplace-Ready Governance

- Capability review workflows, schema-versioning, plugin quality scoring.

## Explicit Answer to Your Core Question

How should actions/state be scoped per plugin but still allow shared behaviors?

Answer:

1. **Identity**: bind everything to host-generated `instanceId`.
2. **Local isolation**: each instance has private `localState` and local reducers.
3. **Shared collaboration**: expose explicit shared domains (`shared.<domain>`) only through capabilities.
4. **Action enforcement**: all plugin actions pass through gateway that validates `scope + domain + capabilities`.
5. **State exposure**: plugin sees only `{ local, shared(granted), system }`, never root state.

This model gives strict boundaries by default and controlled sharing by design.


## Appendix D: End-to-End Sequence Examples

### Sequence 1: Local Action Dispatch (Allowed)

```text
User click
  -> WidgetRenderer emits UIEventRef
    -> Runtime invokes plugin handler(instanceId)
      -> plugin dispatches local action envelope
        -> dispatch gateway validates scope=local (always owned by source instance)
          -> reducer router routes to instance local reducer
            -> state.instances[instanceId].local updated
              -> selector builds plugin view
                -> widget re-render with updated local state
```

Critical invariant:

- No other instance local state is reachable in this flow.

### Sequence 2: Shared Action Dispatch (Denied)

```text
Plugin instance counter@A
  -> dispatches shared action domain=workspace
    -> gateway checks capabilities.writeShared includes workspace
      -> false
        -> action denied
          -> denial telemetry emitted
            -> plugin gets error
              -> no state mutation applied
```

Critical invariant:

- Shared domains are not writable by convention; they are writable by explicit policy.

### Sequence 3: Shared Action Dispatch (Allowed)

```text
Plugin instance monitor@X
  -> dispatches shared action domain=leaderboard
    -> gateway capability check passes
      -> shared reducer leaderboard handles action
        -> pluginRuntime.shared.leaderboard updated
          -> all plugins with readShared: [leaderboard] observe new value on next render
```

Critical invariant:

- Shared visibility is still capability-bounded per reader.

### Sequence 4: Plugin Reload (Same Package, New Instance)

```text
User reloads package counter
  -> runtime creates new instanceId counter@NEW
    -> previous instance counter@OLD remains unless explicitly removed
      -> local state starts fresh from initLocalState
        -> shared state remains unchanged
```

This sequence is how independent plugin sessions are achieved without collisions.

## Appendix E: Policy Engine Rules (Reference Draft)

### Rule Inputs

- `instanceId`
- `packageId`
- action envelope (`scope`, `domain`, `type`)
- granted capability set

### Rule Outputs

- `allow | deny`
- reason code

### Suggested Reason Codes

- `unknown-instance`
- `invalid-envelope`
- `capability-read-missing`
- `capability-write-shared-missing`
- `system-command-not-allowed`
- `scope-domain-mismatch`

### Rule Table Example

| Scope | Condition | Decision | Reason |
|---|---|---|---|
| local | source instance exists | allow | - |
| shared | domain in writeShared | allow | - |
| shared | domain not in writeShared | deny | capability-write-shared-missing |
| system | command in systemCommands | allow | - |
| system | command missing | deny | system-command-not-allowed |

## Appendix F: Anti-Patterns to Avoid During Refactor

1. Reintroducing raw `dispatch` into plugin context “temporarily”.
- This bypasses policy guarantees and usually becomes permanent debt.

2. Encoding scope only in action type string.
- Keep type strings descriptive, but authority should come from explicit metadata and policy.

3. Treating plugin-declared ID as map key.
- Keep declared ID as descriptive metadata only.

4. Using one shared catch-all object for collaboration.
- Shared domains should be narrow and owner-defined to avoid data swamp.

5. Carrying both old and new runtime paths indefinitely.
- Set a clear cutoff milestone to avoid permanent split-brain architecture.

6. Exposing full root state to “simplify migration”.
- Use compatibility adapter, not root state leakage.

## Appendix G: Example Capability Profiles

### Profile A: Fully Local Utility Plugin

```json
{
  "readShared": [],
  "writeShared": [],
  "systemCommands": []
}
```

Use cases:

- local calculators
- personal notes
- temporary scratch widgets

### Profile B: Read-Only Dashboard Plugin

```json
{
  "readShared": ["counter", "workspace"],
  "writeShared": [],
  "systemCommands": []
}
```

Use cases:

- analytics panels
- status monitors

### Profile C: Collaborative Editor Plugin

```json
{
  "readShared": ["workspace", "presence"],
  "writeShared": ["workspace"],
  "systemCommands": []
}
```

Use cases:

- shared document editing
- collaborative workflow boards

### Profile D: Runtime Operator Plugin

```json
{
  "readShared": ["*"],
  "writeShared": [],
  "systemCommands": ["restart-plugin", "toggle-plugin", "collect-diagnostics"]
}
```

Use cases:

- administrative tooling, internal-only plugins

## Appendix H: Rollout KPIs

Track these during migration:

1. Percentage of plugin actions using new envelope format.
2. Count of policy denials by reason code.
3. Number of plugins still relying on compatibility adapter.
4. Number of cross-instance state mutation incidents.
5. Time-to-debug for plugin scope issues before vs after migration.

Success target for phase completion:

- 100% envelope adoption for core presets,
- 0 unauthorized shared writes in test and staging,
- 0 runtime ID collisions,
- single runtime path in production mode.
