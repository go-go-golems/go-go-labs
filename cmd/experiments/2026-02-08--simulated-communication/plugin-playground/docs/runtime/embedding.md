# Runtime Embedding Guide

The `plugin-runtime` package (`packages/plugin-runtime`) is designed to be embedded in any JavaScript/TypeScript application — not just the Plugin Playground UI. This guide shows how to use the runtime in your own projects.

## Package Architecture

```
packages/plugin-runtime/src/
├── contracts.ts          # TypeScript types: DispatchIntent, LoadedPlugin, etc.
├── uiTypes.ts            # UINode type definition
├── uiSchema.ts           # UINode validation
├── dispatchIntent.ts     # Intent validation
├── runtimeIdentity.ts    # Instance ID generation
├── runtimeService.ts     # QuickJS-based plugin execution engine
├── hostAdapter.ts        # Abstract adapter interface
├── redux-adapter/
│   └── store.ts          # Redux Toolkit store with state, policy, projections
└── worker/
    ├── sandboxClient.ts  # Web Worker transport client
    └── runtime.worker.ts # Worker entry point
```

In the playground app, all imports use the TypeScript path alias `@runtime/*` which resolves to `packages/plugin-runtime/src/*`.

## Two Execution Modes

The runtime supports two execution modes, depending on whether you want to run QuickJS on the main thread or in a Web Worker.

### Mode A: Direct Service (Main Thread)

Use `QuickJSRuntimeService` when you need synchronous access, or in environments without Web Worker support (Node.js tests, SSR, CLI tools).

```ts
import { QuickJSRuntimeService } from "@runtime/runtimeService";
import { createInstanceId } from "@runtime/runtimeIdentity";

// Create the runtime with timeout limits
const runtime = new QuickJSRuntimeService({
  memoryLimitBytes: 32 * 1024 * 1024,  // 32 MB per plugin
  stackLimitBytes: 1024 * 1024,         // 1 MB stack
  loadTimeoutMs: 1000,                  // max time to load a plugin
  renderTimeoutMs: 100,                 // max time to render a widget
  eventTimeoutMs: 100,                  // max time to handle an event
});

// Load a plugin
const packageId = "my-plugin";
const instanceId = createInstanceId(packageId);
const plugin = await runtime.loadPlugin(packageId, instanceId, pluginCode);
// plugin: { packageId, instanceId, title, widgets: ["main"], initialState: {...} }

// Render a widget
const pluginState = plugin.initialState ?? {};
const globalState = {
  self: { instanceId, packageId },
  shared: {},
  system: {},
};
const tree = runtime.render(instanceId, "main", pluginState, globalState);
// tree: UINode — a JSON tree of { kind, props, children, text }

// Handle an event
const intents = runtime.event(
  instanceId, "main", "increment", undefined,
  pluginState, globalState
);
// intents: DispatchIntent[] — actions to apply to your state store

// Clean up
runtime.disposePlugin(instanceId);
```

Each plugin gets its own QuickJS runtime + context, with independent memory and stack limits. The interrupt handler enforces timeouts — if a plugin takes longer than the configured limit, execution is aborted with a `RUNTIME_TIMEOUT` error.

### Mode B: Worker Transport (Web Worker)

Use `QuickJSSandboxClient` in browser applications to keep the main thread responsive. All QuickJS execution happens in a dedicated Web Worker.

```ts
import { QuickJSSandboxClient } from "@runtime/worker/sandboxClient";

const sandbox = new QuickJSSandboxClient();

// Same API, but everything is async
const plugin = await sandbox.loadPlugin(packageId, instanceId, pluginCode);
const tree = await sandbox.render(instanceId, "main", pluginState, globalState);
const intents = await sandbox.event(
  instanceId, "main", "increment", undefined, pluginState, globalState
);

// Health check — returns { ready: true, plugins: [...instanceIds] }
const health = await sandbox.health();

// Clean up
await sandbox.disposePlugin(instanceId);
sandbox.terminate(); // kills the worker
```

The sandbox client uses a request/response protocol over `postMessage`. Each call gets a unique ID, and responses are matched by ID. Errors from QuickJS are wrapped in `RuntimeErrorPayload` objects.

### Mode C: Host Adapter (Abstract Interface)

Use `RuntimeHostAdapter` when you want to write code that works with either mode:

```ts
import type { RuntimeHostAdapter } from "@runtime/hostAdapter";

function createMyApp(adapter: RuntimeHostAdapter) {
  // Works with both QuickJSRuntimeService and QuickJSSandboxClient
  const plugin = await adapter.loadPlugin({
    packageId: "my-plugin",
    instanceId: createInstanceId("my-plugin"),
    code: pluginSource,
  });
  // ...
}
```

## Redux Adapter: The State Layer

The `redux-adapter/store` module provides a complete state management layer for plugin instances, shared domains, capability grants, and dispatch telemetry.

### Store Shape

```ts
{
  runtime: {
    plugins: {
      [instanceId]: {
        instanceId, packageId, title, description,
        widgets, enabled, status
      }
    },
    pluginStateById: {
      [instanceId]: { /* plugin's local state */ }
    },
    grantsByInstance: {
      [instanceId]: {
        readShared: ["counter-summary"],
        writeShared: ["counter-summary"],
        systemCommands: [],
      }
    },
    shared: {
      "counter-summary": { totalValue, instanceCount, ... },
      "greeter-profile": { name, lastUpdatedInstanceId },
    },
    dispatchTrace: { count, lastTimestamp, lastScope, ... },
    dispatchTimeline: [ /* bounded array of timeline entries */ ],
  }
}
```

### Registration and Removal

```ts
import {
  store, pluginRegistered, pluginRemoved,
} from "@runtime/redux-adapter/store";

// Register after loading
store.dispatch(pluginRegistered({
  instanceId: "counter-abc-1234",
  packageId: "counter",
  title: "Counter",
  widgets: ["main"],
  initialState: { value: 0 },
  grants: {
    readShared: ["counter-summary"],
    writeShared: ["counter-summary"],
    systemCommands: [],
  },
}));

// Remove on teardown
store.dispatch(pluginRemoved("counter-abc-1234"));
```

### Dispatching Intents

After calling `event()` on the runtime, you get back an array of `DispatchIntent` objects. Feed them through the Redux adapter:

```ts
import {
  dispatchPluginAction,
  dispatchSharedAction,
} from "@runtime/redux-adapter/store";

function applyIntents(instanceId: string, intents: DispatchIntent[]) {
  for (const intent of intents) {
    if (intent.scope === "plugin") {
      dispatchPluginAction(store.dispatch, instanceId, intent.actionType, intent.payload);
    }
    if (intent.scope === "shared" && intent.domain) {
      dispatchSharedAction(
        store.dispatch, instanceId,
        intent.domain, intent.actionType, intent.payload
      );
    }
  }
}
```

Each dispatch is automatically:
- Assigned a unique `dispatchId` (nanoid)
- Timestamped
- Evaluated against capability grants (for shared dispatches)
- Recorded in the dispatch timeline

### Reading State for Rendering

```ts
import {
  selectPluginState,
  selectGlobalStateForInstance,
} from "@runtime/redux-adapter/store";

const state = store.getState();
const pluginState = selectPluginState(state, instanceId);
const globalState = selectGlobalStateForInstance(state, instanceId);
// globalState.shared only includes domains the instance has read grants for
```

### Timeline and Metrics

```ts
import {
  selectDispatchTimeline,
  selectGlobalState,
} from "@runtime/redux-adapter/store";

const timeline = selectDispatchTimeline(store.getState());
// Array of { dispatchId, timestamp, scope, actionType, outcome, reason, ... }

const global = selectGlobalState(store.getState());
// global.system.dispatchCount, global.system.pluginCount, etc.
```

## The Host Loop

Here's the complete sequence for hosting a plugin:

```
┌─────────────────────────────────────────────────────────────┐
│                      Host Application                        │
│                                                              │
│  1. loadPlugin(packageId, instanceId, code)                  │
│     └─ Returns: { title, widgets, initialState }             │
│                                                              │
│  2. pluginRegistered({ instanceId, ..., grants })            │
│     └─ Registers in Redux with capability grants             │
│                                                              │
│  3. render(instanceId, widgetId, pluginState, globalState)   │
│     └─ Returns: UINode tree                                  │
│     └─ Display the tree using your own renderer              │
│                                                              │
│  ┌── Event Loop ──────────────────────────────────────────┐  │
│  │                                                        │  │
│  │  4. User interacts with widget                         │  │
│  │     └─ event(instanceId, widgetId, handler, args, ...) │  │
│  │     └─ Returns: DispatchIntent[]                       │  │
│  │                                                        │  │
│  │  5. Apply intents through Redux adapter                │  │
│  │     └─ Policy check → Reducer → New state              │  │
│  │     └─ Timeline entry recorded                         │  │
│  │                                                        │  │
│  │  6. Re-render affected widgets with new state          │  │
│  │     └─ selectPluginState + selectGlobalState           │  │
│  │     └─ Back to step 4                                  │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  7. disposePlugin(instanceId) + pluginRemoved(instanceId)    │
│     └─ Frees QuickJS resources and Redux state               │
└─────────────────────────────────────────────────────────────┘
```

## Using Without Redux

You don't have to use the Redux adapter. If you have your own state management, you can use the runtime service directly and handle intents yourself:

```ts
const runtime = new QuickJSRuntimeService();
const plugin = await runtime.loadPlugin("my-plugin", "inst-1", code);

// Your own state
let myState = plugin.initialState ?? {};

// Render
const tree = runtime.render("inst-1", "main", myState, { self: null, shared: {}, system: {} });

// Handle event
const intents = runtime.event("inst-1", "main", "increment", undefined, myState, globalState);

// Apply intents with your own logic
for (const intent of intents) {
  if (intent.scope === "plugin" && intent.actionType === "state/merge") {
    myState = { ...myState, ...intent.payload };
  }
  if (intent.scope === "plugin" && intent.actionType === "state/replace") {
    myState = intent.payload;
  }
  // Handle shared intents as needed
}

// Re-render with new state
const newTree = runtime.render("inst-1", "main", myState, globalState);
```

## Error Handling

All runtime errors are wrapped in `RuntimeErrorPayload`:

```ts
interface RuntimeErrorPayload {
  code: string;    // "RUNTIME_ERROR", "RUNTIME_TIMEOUT", "UNKNOWN_ERROR"
  message: string; // Human-readable message
  details?: unknown;
}
```

- **`RUNTIME_TIMEOUT`** — Plugin execution exceeded the configured timeout
- **`RUNTIME_ERROR`** — JavaScript error inside the QuickJS sandbox
- **`UNKNOWN_ERROR`** — Unexpected error type

In worker mode, errors are serialized across the `postMessage` boundary and re-thrown on the client side.

## Security Notes

- Each plugin runs in its own QuickJS context with no access to the host's `window`, `document`, `fetch`, or any browser API
- Memory and stack limits prevent resource exhaustion
- Timeouts prevent infinite loops
- Capability grants prevent unauthorized cross-plugin state access
- All state crosses a JSON serialization boundary — no object references leak between plugins or between a plugin and the host
