# Plugin Authoring Quickstart

This quickstart shows how to write a plugin that runs in the WebVM playground runtime.

## 1) Write a plugin with `definePlugin`

The runtime expects plugin code to call `definePlugin((host) => ({ ... }))`.
`host.ui` provides a small UI builder API (`panel`, `row`, `text`, `button`, `input`, `badge`, `table`).

```js
definePlugin(({ ui }) => {
  return {
    id: "hello-counter",
    title: "Hello Counter",
    description: "Minimal custom plugin",
    initialState: { count: 0, name: "world" },
    widgets: {
      main: {
        render({ pluginState }) {
          const count = Number(pluginState?.count ?? 0);
          const name = String(pluginState?.name ?? "world");
          return ui.panel([
            ui.text("Hello, " + name),
            ui.badge("Count: " + count),
            ui.row([
              ui.button("Increment", { onClick: { handler: "increment" } }),
              ui.button("Rename", { onClick: { handler: "rename", args: { value: "friend" } } }),
            ]),
          ]);
        },
        handlers: {
          increment({ dispatchPluginAction, pluginState }) {
            const next = Number(pluginState?.count ?? 0) + 1;
            dispatchPluginAction("state/merge", { count: next });
          },
          rename({ dispatchPluginAction }, args) {
            dispatchPluginAction("state/merge", { name: String(args?.value ?? "world") });
          },
        },
      },
    },
  };
});
```

## 2) Understand handler context

Each widget handler receives:

- `pluginState`: local state for this plugin instance
- `globalState`: projected shared/system state visible to this instance
- `dispatchPluginAction(actionType, payload?)`: emits a local plugin action intent
- `dispatchSharedAction(domain, actionType, payload?)`: emits a shared-domain action intent

## 3) Local state rules in current runtime

For custom plugins loaded from the editor (`packageId = "custom"`), built-in reducers only handle:

- `state/replace` with any payload
- `state/merge` with an object payload

Use those two action types for predictable local state updates in custom plugins.

## 4) Shared-domain writes require capability grants

Shared writes are deny-by-default. In playground today:

- Preset plugins can declare `capabilities.writeShared` and `capabilities.readShared`.
- Custom editor plugins are loaded with empty grants, so shared writes are denied.

When denied, runtime still records the dispatch in timeline with `outcome = denied` and a reason like:
`missing-write-grant:<domain>`.

## 5) Validate quickly in UI

1. Open the playground and paste plugin code in the workspace editor.
2. Click `LOAD CUSTOM PLUGIN`.
3. Open `INSPECTOR -> WIDGETS` to verify rendering.
4. Open `INSPECTOR -> TIMELINE` to inspect dispatched actions and outcomes.
