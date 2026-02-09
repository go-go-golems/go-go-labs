---
Title: "WebVM Plugin Playground - Vision and Architecture Explainer"
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
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: Main playground page orchestrating plugin load/render/event lifecycle
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts
      Note: Unified runtime slice with plugin/global state, scoped dispatch, selectors, and dispatch tracing
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/pluginManager.ts
      Note: Plugin execution engine with host-authoritative identity and scoped handler context
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/WidgetRenderer.tsx
      Note: React bridge that renders data-only UINode trees into real components
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiTypes.ts
      Note: Canonical UINode type contract for plugin-to-host UI communication
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: Built-in sample plugins demonstrating the v1 plugin authoring API
ExternalSources: []
Summary: "A comprehensive explainer of the WebVM Plugin Playground: what it is, what it wants to become, and how all the pieces fit together. Updated to reflect the v1 unified runtime cleanup."
LastUpdated: 2026-02-08T23:25:00Z
WhatFor: "Explain the end-to-end vision of the plugin playground system to newcomers, collaborators, and future-self."
WhenToUse: "Read this first when coming to the project. Use as the canonical 'what is this and where is it going' document."
---

# WebVM Plugin Playground — Vision and Architecture Explainer

## What Is This?

The WebVM Plugin Playground is a browser-based environment where small JavaScript programs — **plugins** — run inside a managed runtime, render interactive UI widgets, manage their own private state, and optionally observe or influence shared global state. Think of it as a tiny operating system for widgets that runs entirely in your browser.

The key idea: **plugins never touch the DOM directly**. Instead, they describe their UI as plain data trees, the host renders those trees into real React components, and all state flows through a unified Redux runtime slice that the host controls. Plugins are guests; the host is the landlord.

## The Big Picture

```
┌──────────────────────────────────────────────────────────────────┐
│                     Browser (Host Application)                   │
│                                                                  │
│  ┌──────────────┐   ┌──────────────┐   ┌───────────────────┐   │
│  │  Plugin List  │   │ Code Editor  │   │   Live Widgets    │   │
│  │  (sidebar)    │   │ (textarea)   │   │   (rendered UI)   │   │
│  └──────────────┘   └──────────────┘   └───────────────────┘   │
│         │                   │                     ▲              │
│         │                   │     ┌───────────────┘              │
│         │                   ▼     │                              │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                     Plugin Manager                        │   │
│  │                                                           │   │
│  │   ┌─────────┐  ┌─────────┐  ┌─────────┐                │   │
│  │   │ Counter  │  │ Greeter │  │ Calc    │  ... more       │   │
│  │   │ Plugin   │  │ Plugin  │  │ Plugin  │  plugins        │   │
│  │   └────┬─────┘  └────┬────┘  └────┬────┘                │   │
│  │        │              │            │                      │   │
│  │   Each plugin produces:                                   │   │
│  │   • UINode data tree       (what to show)                │   │
│  │   • Scoped action calls    (what to do on interaction)   │   │
│  └──────────────────┬───────────────────────────────────────┘   │
│                     │                                            │
│                     ▼                                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │              Unified Redux Runtime Slice                  │   │
│  │                                                           │   │
│  │   plugins:          { counter, greeter, calculator, ... } │   │
│  │   pluginStateById:                                        │   │
│  │     counter:      { value: 5 }                           │   │
│  │     greeter:      { name: "Alice" }                      │   │
│  │     calculator:   { display: "42", ... }                 │   │
│  │   globals:                                                │   │
│  │     counterValue: 5                                       │   │
│  │   dispatchTrace:                                          │   │
│  │     count: 27, lastDispatchId: "abc123", ...             │   │
│  └──────────────────────────────────────────────────────────┘   │
│                     │                                            │
│                     ▼                                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                  Widget Renderer                          │   │
│  │                                                           │   │
│  │   UINode { kind: "panel", children: [...] }              │   │
│  │       → real <div>, <Button>, <Input> React components   │   │
│  └──────────────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────────────┘
```

## Why Build This?

### The Problem

Modern web applications often need extensibility — the ability for third parties (or different teams) to add functionality without modifying the core application. But extensibility in the browser is hard:

- **Security**: You can't just `eval()` arbitrary JavaScript and trust it.
- **Isolation**: A rogue plugin shouldn't be able to read another plugin's data or freeze the whole page.
- **Consistency**: Every plugin should produce UI that looks and feels like it belongs in the host app.
- **Composability**: Plugins should be able to share certain state (a shared counter value, a workspace document) without becoming tightly coupled.

### The Solution

The Plugin Playground solves this by establishing a **strict contract** between the host and plugins:

1. **Plugins describe UI as data**, not React components. A plugin returns a tree of `UINode` objects like `{ kind: "button", props: { label: "Click me", onClick: { handler: "doThing" } } }`. The host decides how to render that.

2. **State is partitioned** into plugin-local state and shared global state, managed centrally by the host. A plugin receives only its own local state (`pluginState`) and a curated global view (`globalState`) — never the full Redux root.

3. **Actions flow through scoped dispatch functions** that automatically stamp every action with a unique `dispatchId`, timestamp, scope, and source plugin ID. A plugin can't secretly dispatch actions pretending to be another plugin.

4. **Global actions are allowlisted.** Only action types explicitly listed in `ALLOWED_GLOBAL_ACTION_TYPES` can be dispatched globally. Everything else is rejected.

5. **Future: execution will be sandboxed** via QuickJS (a lightweight JavaScript engine compiled to WebAssembly), running in a Web Worker. Plugin code will literally not be able to access `window`, `document`, `fetch`, or any browser API.

## Core Concepts

### 1. Plugins Are Pure Functions

A plugin is a JavaScript module that calls `definePlugin()` with a factory function. That factory receives a build context (just the `ui` builder) and returns a plugin definition:

```javascript
definePlugin(({ ui }) => ({
  id: "counter",
  title: "Counter",
  initialState: { value: 0 },

  widgets: {
    counter: {
      render({ pluginState }) {
        const value = Number(pluginState?.value ?? 0);
        return ui.panel([
          ui.text("Counter: " + value),
          ui.button("Increment", {
            onClick: { handler: "increment" }
          })
        ]);
      },
      handlers: {
        increment({ dispatchPluginAction }) {
          dispatchPluginAction("increment");
        }
      }
    }
  }
}));
```

Key properties:
- **`ui` builder** creates JSON-serializable node trees, not React elements.
- **`render()` receives `pluginState` and `globalState`** — never the full Redux tree.
- **Handlers receive scoped dispatch functions**, not raw Redux `dispatch`.
- **`initialState`** declares the plugin's starting local state.

### 2. The UINode Data Contract

Plugins never produce React components. They produce a plain data tree using a fixed set of node kinds:

| Kind       | Purpose                           | Key Props                        |
|------------|-----------------------------------|----------------------------------|
| `panel`    | Vertical container with border    | `children`                       |
| `row`      | Horizontal flex container         | `children`                       |
| `column`   | Vertical flex container           | `children`                       |
| `text`     | Text display                      | `text`                           |
| `badge`    | Small status label                | `text`                           |
| `button`   | Clickable button                  | `label`, `onClick` (event ref)   |
| `input`    | Text input field                  | `value`, `placeholder`, `onChange`|
| `counter`  | Increment/decrement widget        | `value`, `onIncrement`, `onDecrement` |
| `table`    | Data table                        | `headers`, `rows`                |

Event references are data too: `{ handler: "increment", args: 42 }`. When the user clicks a button, the host looks up the handler by name and calls it with a controlled context.

This design means:
- The host can **theme all plugins uniformly** (cyberpunk terminal aesthetic, or anything else).
- The host can **validate and sanitize** every UI tree before rendering.
- Plugin UI trees are **serializable** — they can cross a Web Worker boundary without issue.

### 3. State Partitioning: Plugin-Local vs Global

The unified runtime slice maintains two parallel state stores:

**`pluginStateById`** — a map of plugin ID → local state. Each plugin gets its own private compartment. The counter plugin sees `{ value: 5 }`; the greeter sees `{ name: "Alice" }`. Neither can see the other's data.

**`globals`** — shared state visible to every plugin via `selectGlobalState`. Currently includes the counter value, plugin count, dispatch count, last dispatch trace, and a summary of all loaded plugins. This is a curated view, not raw Redux state.

```
Runtime Slice
├── plugins:          { counter: {...}, greeter: {...}, ... }
├── pluginStateById:
│   ├── counter:      { value: 5 }
│   ├── calculator:   { display: "42", accumulator: 0, operation: null }
│   └── greeter:      { name: "Alice" }
├── globals:
│   └── counterValue: 5
└── dispatchTrace:
    ├── count: 27
    ├── lastDispatchId: "abc123"
    ├── lastScope: "plugin"
    └── lastActionType: "increment"
```

The **selectors** enforce the boundary:
- `selectPluginState(state, pluginId)` — returns only that plugin's local state.
- `selectGlobalState(state)` — returns the curated global view.
- `selectAllPluginState(state)` — used by Playground to render all widgets efficiently.

A plugin's `render()` function receives `{ pluginState, globalState }` — exactly what it needs, nothing more.

### 4. The Action System: Scoped Dispatch with Tracing

Plugins don't get raw Redux `dispatch`. Handlers receive two scoped functions:

- **`dispatchPluginAction(actionType, payload?)`** — dispatches an action targeting this plugin's local state. The host stamps it with a `dispatchId`, timestamp, `scope: "plugin"`, and the source `pluginId`.

- **`dispatchGlobalAction(actionType, payload?)`** — dispatches a global action. The host validates the action type against `ALLOWED_GLOBAL_ACTION_TYPES` and rejects unknown types with a thrown error.

Every dispatched action becomes a `ScopedDispatchPayload`:

```javascript
{
  dispatchId: "V1StGXR8_Z5jdHi6B-myT",  // unique nanoid
  timestamp: 1707430000000,
  scope: "plugin",                         // or "global"
  pluginId: "counter",                     // who dispatched
  actionType: "increment",                 // what they did
  payload: undefined                       // optional data
}
```

The dispatch trace is recorded in the runtime state (`dispatchTrace`), giving you built-in observability: how many actions have been dispatched, what the last one was, and who sent it. The Status Dashboard preset plugin reads this trace to show live dispatch metrics.

### 5. Host-Authoritative Plugin Identity

The host controls plugin identity. When a plugin is loaded, the host assigns the `pluginId`:

- **Preset plugins** use the preset's `id` (e.g., `"counter"`, `"calculator"`).
- **Custom plugins** get a generated ID (e.g., `"custom-1707430000000"`).

The plugin can declare its own `id` inside `definePlugin()`, but that's stored as `declaredId` — metadata only. All internal maps, state partitions, action routing, and the Redux registry are keyed by the host-assigned `pluginId`.

This means:
- A plugin can't fake its identity.
- The host can load the same plugin code multiple times with different IDs (future: independent instances with separate local state).
- Collisions are impossible because the host controls the key space.

### 6. The Plugin Lifecycle

```
┌──────────┐   loadPlugin()   ┌──────────┐   pluginRegistered()  ┌──────────┐
│  Preset  │ ──────────────→ │  Plugin   │ ───────────────────→ │  Redux   │
│  or Code │                  │  Manager  │                       │  Slice   │
└──────────┘                  └──────────┘                       └──────────┘
                                   │                                  │
                              stores plugin                    stores plugin
                              instance in Map                  metadata + 
                              (widgets, title,                 initial state
                               handlers)                       in runtime
```

1. **Load**: `pluginManager.loadPlugin(pluginId, code, { ui })` executes the plugin code via `new Function`, captures the definition, and stores the plugin instance keyed by `pluginId`.

2. **Register**: `dispatch(pluginRegistered({ id, title, widgets, initialState }))` adds the plugin to the Redux runtime slice, initializing its local state compartment.

3. **Render**: On every Redux state change, `Playground` reads `pluginStateById[pluginId]` and `globalState`, calls `pluginManager.renderWidget(pluginId, widgetId, pluginState, globalState)`, and passes the resulting `UINode` tree to `WidgetRenderer`.

4. **Event**: When the user interacts with a widget, `WidgetRenderer` calls `onEvent(eventRef)`. Playground routes this to `pluginManager.callHandler(...)`, which invokes the handler with scoped dispatch functions bound to the correct `pluginId`.

5. **Unload**: `pluginManager.removePlugin(pluginId)` + `dispatch(pluginRemoved(pluginId))` cleans up both the in-memory instance and the Redux state.

### 7. The Widget Renderer: Data Trees → Real UI

The `WidgetRenderer` is a React component that takes a `UINode` tree (plain data) and recursively renders it into real React components:

- `{ kind: "button", props: { label: "Click" } }` → `<Button>Click</Button>`
- `{ kind: "panel", children: [...] }` → `<div className="...">...</div>`
- `{ kind: "input", props: { value, onChange } }` → `<Input value={value} onChange={...} />`

The renderer controls all styling. The current aesthetic is "Technical Brutalism meets Cyberpunk Terminal" — dark backgrounds, electric cyan accents, monospace typography, glowing borders on interactive elements. But because plugins only produce data, the host could switch to any visual theme without changing a single plugin.

Event references (`onClick: { handler: "increment", args: 42 }`) are translated into callbacks that route through the host's scoped dispatch functions. The plugin never gets a direct reference to a DOM element or a React event.

### 8. The Global Action Allowlist

Not every action can be dispatched globally. The store maintains a set:

```typescript
const ALLOWED_GLOBAL_ACTION_TYPES = new Set(["counter/set"]);
```

If a plugin calls `dispatchGlobalAction("anything/else", ...)`, the host throws an error before it ever reaches the reducer. This prevents plugins from accidentally or maliciously mutating global state in unintended ways.

New global action types must be explicitly added to the allowlist — a deliberate design decision that keeps the global surface area small and auditable.

## How It All Fits Together: A Complete Interaction

Let's trace a button click from the user's finger all the way through the system:

**1. User clicks "Increment" on the counter widget.**

The React `<Button>` in `WidgetRenderer` fires its `onClick`. The renderer finds the event reference: `{ handler: "increment" }`.

**2. Playground routes the event to the plugin manager.**

```typescript
pluginManager.callHandler(
  "counter",       // pluginId
  "counter",       // widgetId
  "increment",     // handler name
  (actionType, payload) => dispatchPluginAction(dispatch, "counter", actionType, payload),
  (actionType, payload) => dispatchGlobalAction(dispatch, actionType, payload),
  undefined,       // args
  pluginState,     // { value: 5 }
  globalState      // curated global view
);
```

Note: the dispatch functions are **pre-bound to the plugin's ID**. The handler cannot change which plugin the action is attributed to.

**3. The handler executes.**

```javascript
increment({ dispatchPluginAction, dispatchGlobalAction, pluginState }) {
  const next = Number(pluginState?.value ?? 0) + 1;
  dispatchPluginAction("increment");
  dispatchGlobalAction("counter/set", next);
}
```

Two dispatch calls: one plugin-scoped (to update local state), one global (to mirror the value into shared globals).

**4. Plugin action is stamped and dispatched.**

`pluginActionDispatched.prepare("counter", "increment")` creates:

```javascript
{
  dispatchId: "V1StGXR8_Z5jdHi6B-myT",
  timestamp: 1707430000000,
  scope: "plugin",
  pluginId: "counter",
  actionType: "increment"
}
```

The reducer routes to `reduceCounterPlugin`, which increments the value and mirrors it to `globals.counterValue`.

**5. Global action is validated and dispatched.**

`"counter/set"` is in `ALLOWED_GLOBAL_ACTION_TYPES` ✓. The action is stamped with its own `dispatchId` and the global reducer sets `globals.counterValue = 6`.

**6. React re-renders.**

Redux state change triggers selector updates. `pluginStateById["counter"]` is now `{ value: 6 }`. `globalState.counterValue` is now `6`. Playground re-renders the counter widget with the new `pluginState`, and also re-renders the Status Dashboard (which reads `globalState.counterValue`).

**7. User sees "Counter: 6".**

Total time: well under a frame at 60fps.

## What Makes This Different from Just Using iframes?

| Concern           | iframe Approach               | Plugin Playground Approach          |
|-------------------|-------------------------------|-------------------------------------|
| **UI consistency** | Each iframe is a separate app | All UI rendered by host, uniform theme |
| **State sharing** | Requires postMessage plumbing | Built-in global state + selectors    |
| **Performance**   | Heavy (separate document per plugin) | Lightweight (data trees, shared renderer) |
| **Security**      | Good (same-origin separation) | Good today (scoped dispatch), better with QuickJS |
| **Size**          | Large (full framework per plugin) | Tiny (plugins are ~30 lines of JS)   |
| **Composability** | Hard to coordinate            | Native: globals + dispatch tracing   |

## The Preset Plugins: What Ships Out of the Box

The system includes five built-in plugins that demonstrate different patterns:

| Plugin               | What It Shows                                                        |
|----------------------|----------------------------------------------------------------------|
| **Counter**          | Local state + global mirroring via dual dispatch                     |
| **Calculator**       | Complex local state, multi-step operations, no global interaction    |
| **Greeter**          | Text input, onChange events, plugin-local state only                 |
| **Status Dashboard** | Read-only global state observation — badges, tables, dispatch trace  |
| **Runtime Monitor**  | System introspection — reads `globalState.plugins` array             |

Each plugin is a self-contained JavaScript string. You can paste it into the code editor, modify it, and reload it live.

Notice the spectrum of patterns:
- **Calculator** and **Greeter** are purely local — they never touch `dispatchGlobalAction` or read `globalState`.
- **Counter** bridges both worlds — it updates local state *and* mirrors a value to globals.
- **Status Dashboard** and **Runtime Monitor** are read-only observers — they have no handlers at all, just render functions that display global metrics.

## Design Philosophy: Technical Brutalism

The visual language is intentional: **raw, honest, functional**.

- **Dark terminal aesthetic**: Deep charcoal backgrounds, high contrast white text, electric cyan accents.
- **Monospace everywhere**: JetBrains Mono for code, Space Mono for UI labels. No serif, no sans-serif pretending to be friendly.
- **Glowing borders**: Interactive elements pulse with subtle cyan box-shadows. You always know what's clickable.
- **No decoration**: No gradients, no rounded corners, no drop shadows (except the functional glow). Every pixel serves a purpose.
- **Transparency**: Plugin states, dispatch traces, runtime metrics — all visible. The system doesn't hide its machinery.

This isn't just aesthetics; it reflects the system's values: **explicit over implicit, visible over hidden, functional over decorative**.

## Architecture: What Was Cleaned Up

The codebase recently went through a significant cleanup that removed dead code paths and consolidated into a single unified architecture:

**Removed (dead/drifted code):**
- `pluginSandboxClient.ts` — alternate in-process sandbox using `window.definePlugin`; accumulated contract drift.
- `pluginSandbox.worker.ts` — Web Worker path that used `type` instead of `kind` in UI nodes; never wired to active route.
- `presets.ts` — duplicate preset catalog with different handler signatures than the active path.
- `minimalPlugin.ts` + `MinimalPluginWidget.tsx` — hardcoded demo that bypassed the plugin system entirely.
- `PluginWidget.tsx` + `PluginList.tsx` — components for the dormant sandbox path.

**What remains is one clean path:**
- `pluginManager.ts` — single execution engine, host-authoritative identity, scoped handler context.
- `store.ts` — unified `runtime` slice with `pluginStateById`, `globals`, `dispatchTrace`, scoped action creators with `dispatchId` stamping, and allowlisted global actions.
- `presetPlugins.ts` — single canonical preset catalog using the v1 API (`pluginState`/`globalState`, `dispatchPluginAction`/`dispatchGlobalAction`).
- `Playground.tsx` — single page driving all load/render/event orchestration through selectors.

There is now **one way to load a plugin, one way to render, one way to dispatch, and one place where state lives**.

## The Road Ahead

The system is being built in phases:

### Phase 1: Contract Correctness ✅ (Implemented)
- Host-authoritative plugin identity (`pluginId` assigned by host).
- `selectPluginState` / `selectGlobalState` selectors — plugins never see full Redux root.
- `dispatchPluginAction` / `dispatchGlobalAction` with automatic `dispatchId` stamping.
- Global action allowlist (`ALLOWED_GLOBAL_ACTION_TYPES`).
- Dispatch tracing in runtime state.
- Single unified runtime slice, single preset catalog, single execution path.
- Dead code paths removed.

### Phase 2: Real QuickJS Isolation (Next)
- Move plugin execution from in-process `new Function` to QuickJS WASM in a Web Worker.
- Implement a minimal bridge API: `hostDispatchPlugin`, `hostDispatchGlobal`, `hostNow`, `hostLog`.
- Add resource limits (memory cap per plugin, stack limit, execution timeout).
- Plugin code will have zero access to browser globals — isolation by construction, not convention.

### Phase 3: Multiple Instances
- Load the same plugin package multiple times with independent instance IDs and separate local state.
- Instance-aware dispatch routing.
- `packageId` + `instanceId` distinction.

### Phase 4: Capability Model
- Plugin manifests declare requested capabilities (`readShared`, `writeShared`, `systemCommands`).
- Host grants/denies capabilities at load time.
- Dispatch gateway enforces capability grants at runtime.
- Shared state organized into governed domains instead of a single `globals` object.

### Phase 5: Advanced Features
- Plugin hot-reload preserving state.
- Inter-plugin communication via shared domains.
- Plugin persistence (save/restore local state across sessions).
- Plugin versioning and API compatibility layers.
- Plugin marketplace or registry concept.

## Summary

The WebVM Plugin Playground is a **managed plugin runtime for the browser**. Plugins are small JavaScript programs that produce data-only UI trees, manage state through a unified Redux runtime slice with enforced scoping, and interact with the system exclusively through host-controlled dispatch functions that stamp every action with a unique trace ID.

The architecture enforces three key boundaries:

1. **State boundary**: plugins see only `pluginState` (their own) and `globalState` (curated shared view). Never the full store.
2. **Action boundary**: plugins dispatch through scoped functions that are pre-bound to their ID. Global actions are allowlisted.
3. **Render boundary**: plugins produce data trees (`UINode`), the host owns all DOM rendering and styling.

The current runtime uses in-process `new Function` execution with trust-based isolation. The next phase replaces this with QuickJS WASM in a Web Worker — real sandboxing where isolation is a property of the execution environment, not a convention that plugins are expected to follow.

It's a tiny, opinionated operating system for widgets — and it runs in your browser tab.
