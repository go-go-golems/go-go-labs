# Dispatch Lifecycle

Every state change in the Plugin Playground flows through a single pipeline called the **dispatch lifecycle**. Understanding this pipeline is key to writing plugins that behave correctly and to debugging unexpected outcomes.

## The Big Picture

When a user clicks a button in a plugin's UI, the following sequence unfolds:

```
 User clicks button
       │
       ▼
 ┌─────────────┐
 │  UI Event   │  The WidgetRenderer captures the click and finds
 │  Captured   │  the UIEventRef attached to the button.
 └──────┬──────┘
        │
        ▼
 ┌─────────────┐
 │  Sandbox    │  The host sends the event to the QuickJS sandbox.
 │  event()    │  The plugin's handler function runs in isolation
 │             │  and returns a list of DispatchIntents.
 └──────┬──────┘
        │
        ▼
 ┌─────────────┐
 │  Intent     │  Each intent has a scope ("plugin" or "shared"),
 │  Routing    │  an action type, and an optional payload.
 │             │  The host routes each intent to the right reducer.
 └──────┬──────┘
        │
   ┌────┴────┐
   │         │
   ▼         ▼
 Plugin    Shared
 Scope     Scope
   │         │
   ▼         ▼
 ┌───────┐ ┌──────────┐
 │Reducer│ │Policy    │  Shared dispatches go through a capability
 │       │ │Check     │  check first. If the plugin doesn't have a
 │       │ │→ Reducer │  write grant, the dispatch is DENIED.
 └───┬───┘ └────┬─────┘
     │          │
     ▼          ▼
 ┌─────────────────┐
 │  New State      │  The reducer produces new state. The runtime
 │  + Timeline     │  records the dispatch in the timeline with
 │  Entry          │  its outcome (applied/denied/ignored).
 └────────┬────────┘
          │
          ▼
 ┌─────────────────┐
 │  Re-render      │  The host calls render() on all affected
 │  Widgets        │  plugin instances with their new state.
 └─────────────────┘
```

## Step by Step

### 1. UI Event Capture

Every interactive element in a plugin's UI carries a **UIEventRef** — a small object that names a handler function and optionally passes arguments:

```js
ui.button("Increment", {
  onClick: { handler: "increment" }
})

ui.button("Set Name", {
  onClick: { handler: "setName", args: { value: "Alice" } }
})
```

When the user clicks the button, the host's `WidgetRenderer` calls the runtime's `event()` function with the handler name and args.

### 2. Sandbox Execution

The event is sent to the QuickJS sandbox (running in a Web Worker). The sandbox looks up the named handler in the plugin's widget definition and calls it:

```js
handlers: {
  increment({ dispatchPluginAction, pluginState }) {
    const next = Number(pluginState?.value ?? 0) + 1;
    dispatchPluginAction("increment");
  }
}
```

The handler receives a **context object** with:

| Property | Description |
|----------|-------------|
| `pluginState` | This plugin instance's current local state |
| `globalState` | The projected shared/system state visible to this instance |
| `dispatchPluginAction(actionType, payload?)` | Emit a plugin-scoped dispatch intent |
| `dispatchSharedAction(domain, actionType, payload?)` | Emit a shared-scoped dispatch intent |

The handler doesn't modify state directly — it emits **intents**. These are collected and returned to the host as an array of `DispatchIntent` objects.

### 3. Intent Routing

Back on the host side, each intent is routed based on its `scope`:

```ts
interface DispatchIntent {
  scope: "plugin" | "shared";
  actionType: string;
  payload?: unknown;
  domain?: string;       // required when scope is "shared"
}
```

- **`scope: "plugin"`** → dispatched as `pluginActionDispatched` in the Redux store
- **`scope: "shared"`** → dispatched as `sharedActionDispatched` with the target domain

### 4. Plugin-Scoped Dispatch

For plugin-scoped intents, the runtime looks up a reducer based on the plugin's `packageId`:

| Package ID | Reducer | Supported Actions |
|-----------|---------|-------------------|
| `counter` | `reduceCounterPlugin` | `increment`, `decrement`, `reset` |
| `calculator` | `reduceCalculatorPlugin` | `digit`, `operation`, `equals`, `clear` |
| `greeter` | `reduceGreeterPlugin` | `nameChanged` |
| *(any other)* | `reduceGenericPlugin` | `state/replace`, `state/merge` |

The generic reducer is the fallback for custom plugins. It supports two universal action types:

- **`state/replace`** — completely replaces the plugin's local state
- **`state/merge`** — shallow-merges an object payload into existing state

### 5. Shared-Scoped Dispatch (with Policy Check)

Shared dispatches go through an additional **capability check** before reaching the reducer:

```
Intent: { scope: "shared", domain: "counter-summary", actionType: "set-instance" }
                    │
                    ▼
        ┌───────────────────┐
        │ Does this instance │
        │ have writeShared   │──── NO ──→ outcome: "denied"
        │ for this domain?   │            reason: "missing-write-grant:counter-summary"
        └────────┬──────────┘
                 │ YES
                 ▼
        ┌───────────────────┐
        │ Does a reducer     │
        │ exist for this     │──── NO ──→ outcome: "ignored"
        │ domain + action?   │            reason: "unsupported-action:..."
        └────────┬──────────┘
                 │ YES
                 ▼
        ┌───────────────────┐
        │ Apply reducer,     │──────────→ outcome: "applied"
        │ update shared state│
        └───────────────────┘
```

### 6. Timeline Recording

Every dispatch — regardless of outcome — is recorded in the **dispatch timeline**:

```ts
{
  dispatchId: "abc123",
  timestamp: 1707494400000,
  scope: "shared",
  actionType: "set-instance",
  instanceId: "counter-abc-1234",
  domain: "counter-summary",
  outcome: "applied",
  reason: null
}
```

The timeline is capped at 200 entries (oldest entries are evicted). You can inspect it in the DevTools → Timeline tab, filter by scope and outcome, and see exactly why a dispatch was denied or ignored.

### 7. Re-render

After all intents are processed, the host calls `render()` on every loaded plugin with its updated state. The render function runs in the sandbox, produces a new UI tree, and the `WidgetRenderer` diffs and updates the DOM.

## Putting It Together: A Counter Click

Here's what happens when you click "Increment" on the Counter plugin:

1. `WidgetRenderer` fires `event("counter-abc", "main", "increment", undefined)`
2. QuickJS runs the `increment` handler, which calls:
   - `dispatchPluginAction("increment")` → intent `{ scope: "plugin", actionType: "increment" }`
   - `dispatchSharedAction("counter-summary", "set-instance", { value: 6 })` → intent `{ scope: "shared", ... }`
3. Host processes intent #1: `reduceCounterPlugin` increments `value` to 6 → `outcome: "applied"`
4. Host processes intent #2: checks write grant for `counter-summary` → granted → `applyCounterSummarySetInstance` updates shared state → `outcome: "applied"`
5. Both dispatches are appended to the timeline
6. Host re-renders all plugins — Counter shows "6", Status Dashboard (if loaded) shows updated `totalValue`

## Key Design Decisions

**Why intents instead of direct mutation?** Intents make the dispatch pipeline auditable. Every state change goes through a policy check and is recorded. This is essential for debugging multi-plugin interactions.

**Why a bounded timeline?** 200 entries is enough for debugging without unbounded memory growth. In a production embedding you might increase this or stream entries to a backend.

**Why per-package reducers?** Preset plugins have specific state shapes and action semantics. The generic `state/merge` fallback keeps custom plugins simple while letting preset plugins use domain-specific actions like `increment` or `digit`.
