# Plugin Authoring Guide

This guide walks you through writing plugins for the Plugin Playground, from your first "Hello World" to plugins that manage state, handle events, and communicate through shared domains.

## Your First Plugin

Every plugin is a single JavaScript file that calls `definePlugin()`. The function receives a **host context** object and must return a **plugin definition**:

```js
definePlugin(({ ui }) => ({
  id: "hello",
  title: "Hello World",
  initialState: {},
  widgets: {
    main: {
      render() {
        return ui.text("Hello from my first plugin!");
      },
      handlers: {},
    },
  },
}));
```

To try it out:

1. Click a preset in the sidebar (or click **New Plugin**) to open an editor tab
2. Replace the code with the snippet above
3. Click **Run** to load it into the sandbox
4. See "Hello from my first plugin!" in the Live Preview pane

## Plugin Definition Reference

The object you return from `definePlugin()` has this shape:

```js
{
  id: "my-plugin",              // Unique identifier (string)
  title: "My Plugin",           // Display name shown in the UI
  description: "Optional",      // Description (optional)
  initialState: { count: 0 },   // Starting state (optional, defaults to {})
  widgets: {                     // Map of widget definitions
    main: {                      // Widget ID (you can have multiple)
      render({ pluginState, globalState }) { ... },
      handlers: { ... },
    },
  },
}
```

### `id` (required)

A string identifier for your plugin. This doesn't need to be globally unique — the runtime assigns a unique `instanceId` when loading.

### `title` (required)

The human-readable name shown in the sidebar and instance cards.

### `initialState` (optional)

The starting state for your plugin. Can be any JSON-serializable value. If omitted, defaults to `{}`.

### `widgets` (required)

An object mapping widget IDs to widget definitions. Most plugins have a single widget called `main`, but you can define multiple widgets if your plugin needs separate UI sections.

## Widgets: Render + Handlers

Each widget has two parts:

### `render(context)`

A function that receives the current state and returns a UI tree (see [UI DSL Reference](../architecture/ui-dsl.md)):

```js
render({ pluginState, globalState }) {
  return ui.column([
    ui.text("Count: " + pluginState.count),
    ui.button("+1", { onClick: { handler: "increment" } }),
  ]);
}
```

The `render` function is called every time state changes. It must be **pure** — no side effects, no async calls, just build and return a UI tree.

**Context properties:**

| Property | Description |
|----------|-------------|
| `pluginState` | Your plugin's local state (the value of `initialState` initially) |
| `globalState` | Projected state including shared domains you have read access to |
| `globalState.self` | `{ instanceId, packageId }` — your identity |
| `globalState.shared` | Shared domain values (only domains you have read grants for) |
| `globalState.system` | Runtime metrics and plugin registry |

### `handlers`

An object mapping handler names to functions. Handler names must match the `handler` strings used in your UI event refs:

```js
handlers: {
  increment({ dispatchPluginAction, pluginState }) {
    dispatchPluginAction("state/merge", { count: pluginState.count + 1 });
  },

  reset({ dispatchPluginAction }) {
    dispatchPluginAction("state/replace", { count: 0 });
  },
}
```

**Handler context properties:**

| Property | Description |
|----------|-------------|
| `pluginState` | Current local state |
| `globalState` | Projected shared/system state |
| `dispatchPluginAction(actionType, payload?)` | Emit a plugin-scoped state change |
| `dispatchSharedAction(domain, actionType, payload?)` | Emit a shared-domain state change |

The second argument to the handler is the `args` value from the `UIEventRef`:

```js
// In render:
ui.button("Set to 5", { onClick: { handler: "setValue", args: 5 } })

// In handlers:
setValue({ dispatchPluginAction }, value) {
  dispatchPluginAction("state/replace", { count: value });
}
```

## State Management

### Local State

Your plugin's local state is private — no other plugin can see or modify it. To change it, dispatch a plugin-scoped action from a handler.

For custom plugins, the runtime provides two built-in action types:

#### `state/merge`

Shallow-merges an object into your current state:

```js
// State before: { count: 3, name: "Alice" }
dispatchPluginAction("state/merge", { count: 4 });
// State after:  { count: 4, name: "Alice" }
```

#### `state/replace`

Completely replaces your state:

```js
dispatchPluginAction("state/replace", { count: 0 });
// State is now exactly { count: 0 }, regardless of what it was before
```

### Shared State

Plugins can also read and write **shared domains** — named pieces of state that persist across plugin instances. See the [Capability Model](../architecture/capability-model.md) for details on which domains exist and how grants work.

To write to a shared domain:

```js
dispatchSharedAction("counter-summary", "set-instance", { value: 5 });
```

To read from a shared domain (in render):

```js
render({ globalState }) {
  const summary = globalState.shared?.["counter-summary"];
  return ui.text("Total across all counters: " + (summary?.totalValue ?? 0));
}
```

> **Important:** You can only read domains listed in your `readShared` grants, and only write to domains listed in your `writeShared` grants. Unauthorized writes are denied and logged in the timeline.

## Capabilities

When your plugin is loaded from a preset, its capabilities are defined in the preset definition:

```js
capabilities: {
  readShared: ["counter-summary"],
  writeShared: ["counter-summary"],
}
```

When you write custom code in the editor and click Run, the playground grants access to **all** shared domains automatically — this makes the sandbox convenient for experimentation.

## Debugging Your Plugin

### DevTools → Timeline

Shows every dispatch (plugin and shared) with its outcome. Filter by:
- **Scope**: `plugin` or `shared`
- **Outcome**: `applied`, `denied`, or `ignored`

If your state isn't changing, check if your dispatches show `applied` or `ignored`.

### DevTools → State

Shows the JSON state of each plugin instance. Verify your state shape matches what you expect after dispatching.

### DevTools → Capabilities

Shows a grid of which instances have read/write access to which domains. If shared dispatches are being denied, check here first.

### DevTools → Errors

Shows any runtime errors (load failures, render errors, event handler exceptions).

## Common Patterns

### Toggle pattern

```js
initialState: { expanded: false },
widgets: {
  main: {
    render({ pluginState }) {
      const items = [
        ui.button(pluginState.expanded ? "Collapse" : "Expand", {
          onClick: { handler: "toggle" },
        }),
      ];
      if (pluginState.expanded) {
        items.push(ui.text("Hidden content is now visible!"));
      }
      return ui.column(items);
    },
    handlers: {
      toggle({ dispatchPluginAction, pluginState }) {
        dispatchPluginAction("state/merge", { expanded: !pluginState.expanded });
      },
    },
  },
},
```

### Form with input

```js
initialState: { name: "", submitted: false },
widgets: {
  main: {
    render({ pluginState }) {
      if (pluginState.submitted) {
        return ui.column([
          ui.text("Hello, " + pluginState.name + "!"),
          ui.button("Reset", { onClick: { handler: "reset" } }),
        ]);
      }
      return ui.column([
        ui.input({
          value: pluginState.name,
          placeholder: "Your name",
          onChange: { handler: "nameChanged" },
        }),
        ui.button("Submit", { onClick: { handler: "submit" } }),
      ]);
    },
    handlers: {
      nameChanged({ dispatchPluginAction }, args) {
        dispatchPluginAction("state/merge", { name: args.value });
      },
      submit({ dispatchPluginAction }) {
        dispatchPluginAction("state/merge", { submitted: true });
      },
      reset({ dispatchPluginAction }) {
        dispatchPluginAction("state/replace", { name: "", submitted: false });
      },
    },
  },
},
```

### Reading shared state

```js
capabilities: {
  readShared: ["counter-summary", "runtime-metrics"],
},
// ...
render({ globalState }) {
  const counter = globalState.shared?.["counter-summary"];
  const metrics = globalState.shared?.["runtime-metrics"];
  return ui.column([
    ui.text("Counter total: " + (counter?.totalValue ?? 0)),
    ui.text("Dispatch count: " + (metrics?.dispatchCount ?? 0)),
  ]);
}
```

## Next Steps

- See [Plugin Examples](examples.md) for complete worked examples
- Read the [UI DSL Reference](../architecture/ui-dsl.md) for all available UI nodes
- Read the [Dispatch Lifecycle](../architecture/dispatch-lifecycle.md) to understand how state changes flow
- Read the [Capability Model](../architecture/capability-model.md) for shared domain details
