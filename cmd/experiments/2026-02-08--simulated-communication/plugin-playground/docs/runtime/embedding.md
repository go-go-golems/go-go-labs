# Runtime Embedding Guide

This guide shows how to embed `plugin-runtime` outside the current playground page.

## Package shape

`packages/plugin-runtime` is the reusable runtime package. It currently exports:

- Runtime contracts and types (`contracts`, `uiTypes`, `uiSchema`)
- Runtime engine (`runtimeService`)
- Worker transport (`worker/sandboxClient`, `worker/runtime.worker`)
- Host adapter contracts (`hostAdapter`)
- Runtime identity helper (`runtimeIdentity`)
- Runtime state policy adapter (`redux-adapter/store`)

In this repository, consumers import it through the TypeScript alias `@runtime/*`.

## Embedding pattern A: direct runtime service (no web worker)

Use this in Node/test hosts or controlled server-side execution.

```ts
import { QuickJSRuntimeService } from "@runtime/runtimeService";
import { createInstanceId } from "@runtime/runtimeIdentity";

const runtime = new QuickJSRuntimeService({
  loadTimeoutMs: 1000,
  renderTimeoutMs: 100,
  eventTimeoutMs: 100,
});

const packageId = "my-plugin";
const instanceId = createInstanceId(packageId);

const plugin = await runtime.loadPlugin(packageId, instanceId, pluginCode);
const initialPluginState = plugin.initialState ?? {};
const initialGlobalState = { self: { instanceId, packageId }, shared: {}, system: {} };

const tree = runtime.render(instanceId, plugin.widgets[0], initialPluginState, initialGlobalState);
const intents = runtime.event(
  instanceId,
  plugin.widgets[0],
  "onClick",
  undefined,
  initialPluginState,
  initialGlobalState
);
```

## Embedding pattern B: browser worker transport

Use `QuickJSSandboxClient` when the runtime should execute in a dedicated worker.

```ts
import { QuickJSSandboxClient } from "@runtime/worker/sandboxClient";

const sandbox = new QuickJSSandboxClient();
const plugin = await sandbox.loadPlugin(packageId, instanceId, pluginCode);
const tree = await sandbox.render(instanceId, widgetId, pluginState, globalState);
const intents = await sandbox.event(instanceId, widgetId, handler, args, pluginState, globalState);
await sandbox.disposePlugin(instanceId);
sandbox.terminate();
```

## Host loop with Redux adapter

`redux-adapter/store` provides state, policy, projections, and helper dispatchers.

```ts
import {
  store,
  pluginRegistered,
  pluginRemoved,
  dispatchPluginAction,
  dispatchSharedAction,
  selectPluginState,
  selectGlobalStateForInstance,
} from "@runtime/redux-adapter/store";

function applyIntents(instanceId: string, intents: Array<any>) {
  for (const intent of intents) {
    if (intent.scope === "plugin") {
      dispatchPluginAction(store.dispatch, instanceId, intent.actionType, intent.payload);
      continue;
    }
    if (intent.scope === "shared" && intent.domain) {
      dispatchSharedAction(store.dispatch, instanceId, intent.domain, intent.actionType, intent.payload);
    }
  }
}

function buildRenderInputs(instanceId: string) {
  const state = store.getState();
  return {
    pluginState: selectPluginState(state, instanceId),
    globalState: selectGlobalStateForInstance(state, instanceId),
  };
}
```

Recommended host sequence:

1. `loadPlugin(...)`
2. `pluginRegistered(...)` with explicit capability grants
3. render widget(s)
4. on event: call runtime `event(...)`, apply intents through Redux adapter, re-render
5. on teardown: `disposePlugin(...)`, then `pluginRemoved(...)`

## Capability and projection notes

- Shared reads are projected per instance based on `readShared` grants.
- Shared writes require `writeShared` grants or dispatch will be denied.
- Runtime health/registry domains are read-only projections.
- Use `selectDispatchTimeline` to expose debugging telemetry in host UIs.

## When to use `RuntimeHostAdapter`

Use `RuntimeHostAdapter` when building abstractions that should support multiple execution backends (for example, worker-backed runtime in browser and direct service runtime in tests). Keep UI/application orchestration code depending on the adapter interface, not concrete transport classes.
