# Plugin Examples

This document contains complete, ready-to-paste plugin examples, ordered from simple to advanced. Each example demonstrates specific concepts and patterns.

## Example 1: Minimal Counter

**Concepts:** Local state, `state/merge`, button handlers

The simplest interactive plugin — a counter with increment/decrement buttons.

```js
definePlugin(({ ui }) => ({
  id: "minimal-counter",
  title: "Minimal Counter",
  initialState: { value: 0 },
  widgets: {
    main: {
      render({ pluginState }) {
        const value = Number(pluginState?.value ?? 0);
        return ui.column([
          ui.text("Count: " + value),
          ui.row([
            ui.button("-", { onClick: { handler: "decrement" } }),
            ui.button("+", { onClick: { handler: "increment" } }),
          ]),
        ]);
      },
      handlers: {
        increment({ dispatchPluginAction, pluginState }) {
          dispatchPluginAction("state/merge", {
            value: Number(pluginState?.value ?? 0) + 1,
          });
        },
        decrement({ dispatchPluginAction, pluginState }) {
          dispatchPluginAction("state/merge", {
            value: Number(pluginState?.value ?? 0) - 1,
          });
        },
      },
    },
  },
}));
```

**Key points:**
- `state/merge` shallow-merges the payload into existing state
- Handler context gives you `pluginState` so you can compute the next value
- Always use `Number()` to coerce values from state — state comes back as plain JSON

---

## Example 2: Name Form with Input

**Concepts:** Text input, `onChange` handler, conditional rendering

A form that greets the user by name, demonstrating input handling and conditional UI.

```js
definePlugin(({ ui }) => ({
  id: "name-form",
  title: "Name Form",
  initialState: { name: "", submitted: false },
  widgets: {
    main: {
      render({ pluginState }) {
        const name = String(pluginState?.name ?? "");
        const submitted = Boolean(pluginState?.submitted);

        if (submitted) {
          return ui.panel([
            ui.text("Welcome, " + name + "! 🎉"),
            ui.button("Start Over", {
              onClick: { handler: "reset" },
              variant: "destructive",
            }),
          ]);
        }

        return ui.panel([
          ui.text("What's your name?"),
          ui.input({
            value: name,
            placeholder: "Type your name...",
            onChange: { handler: "nameChanged" },
          }),
          ui.button("Submit", { onClick: { handler: "submit" } }),
        ]);
      },
      handlers: {
        nameChanged({ dispatchPluginAction }, args) {
          dispatchPluginAction("state/merge", { name: args?.value ?? "" });
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
}));
```

**Key points:**
- `ui.input` fires `onChange` on every keystroke with `{ value: "..." }` as args
- Use `state/replace` when you want to reset state completely
- Conditional rendering is just `if/else` — return different UI trees based on state

---

## Example 3: Todo List

**Concepts:** Dynamic lists, multiple handlers, computed UI from arrays

A todo list that demonstrates working with array state and rendering dynamic lists.

```js
definePlugin(({ ui }) => ({
  id: "todo-list",
  title: "Todo List",
  initialState: {
    items: [],
    nextId: 1,
    draft: "",
  },
  widgets: {
    main: {
      render({ pluginState }) {
        const items = Array.isArray(pluginState?.items) ? pluginState.items : [];
        const draft = String(pluginState?.draft ?? "");
        const done = items.filter((i) => i.done).length;

        const rows = items.map((item) => [
          item.done ? "✅" : "⬜",
          item.text,
          String(item.id),
        ]);

        return ui.column([
          ui.text("Todo List (" + done + "/" + items.length + " done)"),
          ui.row([
            ui.input({
              value: draft,
              placeholder: "New todo...",
              onChange: { handler: "updateDraft" },
            }),
            ui.button("Add", { onClick: { handler: "addItem" } }),
          ]),
          items.length > 0
            ? ui.table({ headers: ["Status", "Task", "ID"], rows })
            : ui.text("No todos yet — add one above!"),
          ui.row([
            ui.button("Toggle First", { onClick: { handler: "toggleFirst" } }),
            ui.button("Clear Done", {
              onClick: { handler: "clearDone" },
              variant: "destructive",
            }),
          ]),
        ]);
      },
      handlers: {
        updateDraft({ dispatchPluginAction }, args) {
          dispatchPluginAction("state/merge", { draft: args?.value ?? "" });
        },
        addItem({ dispatchPluginAction, pluginState }) {
          const draft = String(pluginState?.draft ?? "").trim();
          if (!draft) return;

          const items = Array.isArray(pluginState?.items) ? [...pluginState.items] : [];
          const nextId = Number(pluginState?.nextId ?? 1);

          items.push({ id: nextId, text: draft, done: false });
          dispatchPluginAction("state/replace", {
            items,
            nextId: nextId + 1,
            draft: "",
          });
        },
        toggleFirst({ dispatchPluginAction, pluginState }) {
          const items = Array.isArray(pluginState?.items) ? [...pluginState.items] : [];
          if (items.length === 0) return;
          items[0] = { ...items[0], done: !items[0].done };
          dispatchPluginAction("state/merge", { items });
        },
        clearDone({ dispatchPluginAction, pluginState }) {
          const items = Array.isArray(pluginState?.items) ? pluginState.items : [];
          dispatchPluginAction("state/merge", {
            items: items.filter((i) => !i.done),
          });
        },
      },
    },
  },
}));
```

**Key points:**
- Use `state/replace` when you need to update multiple fields atomically
- Use `state/merge` when updating a single field (like `items`)
- Always copy arrays before mutation (`[...pluginState.items]`)
- The `ui.table` node accepts `{ headers, rows }` — rows is an array of arrays

---

## Example 4: Calculator

**Concepts:** Complex state machine, many handlers, operation chaining

A four-function calculator demonstrating how to model a multi-step interaction.

```js
definePlugin(({ ui }) => ({
  id: "calculator",
  title: "Calculator",
  initialState: { display: "0", accumulator: 0, operation: null },
  widgets: {
    main: {
      render({ pluginState }) {
        const display = String(pluginState?.display ?? "0");
        return ui.panel([
          ui.badge("Display: " + display),
          ui.row([
            ui.button("7", { onClick: { handler: "digit", args: "7" } }),
            ui.button("8", { onClick: { handler: "digit", args: "8" } }),
            ui.button("9", { onClick: { handler: "digit", args: "9" } }),
            ui.button("÷", { onClick: { handler: "op", args: "/" } }),
          ]),
          ui.row([
            ui.button("4", { onClick: { handler: "digit", args: "4" } }),
            ui.button("5", { onClick: { handler: "digit", args: "5" } }),
            ui.button("6", { onClick: { handler: "digit", args: "6" } }),
            ui.button("×", { onClick: { handler: "op", args: "*" } }),
          ]),
          ui.row([
            ui.button("1", { onClick: { handler: "digit", args: "1" } }),
            ui.button("2", { onClick: { handler: "digit", args: "2" } }),
            ui.button("3", { onClick: { handler: "digit", args: "3" } }),
            ui.button("−", { onClick: { handler: "op", args: "-" } }),
          ]),
          ui.row([
            ui.button("0", { onClick: { handler: "digit", args: "0" } }),
            ui.button("=", { onClick: { handler: "equals" } }),
            ui.button("C", { onClick: { handler: "clear" }, variant: "destructive" }),
            ui.button("+", { onClick: { handler: "op", args: "+" } }),
          ]),
        ]);
      },
      handlers: {
        digit({ dispatchPluginAction, pluginState }, digit) {
          const display = String(pluginState?.display ?? "0");
          const next = display === "0" ? String(digit) : display + String(digit);
          dispatchPluginAction("state/merge", { display: next });
        },
        op({ dispatchPluginAction, pluginState }, operation) {
          dispatchPluginAction("state/replace", {
            display: "0",
            accumulator: parseFloat(String(pluginState?.display ?? "0")),
            operation: String(operation),
          });
        },
        equals({ dispatchPluginAction, pluginState }) {
          const current = parseFloat(String(pluginState?.display ?? "0"));
          const acc = Number(pluginState?.accumulator ?? 0);
          const op = String(pluginState?.operation ?? "");

          let result = current;
          if (op === "+") result = acc + current;
          else if (op === "-") result = acc - current;
          else if (op === "*") result = acc * current;
          else if (op === "/") result = current !== 0 ? acc / current : 0;

          dispatchPluginAction("state/replace", {
            display: String(result),
            accumulator: 0,
            operation: null,
          });
        },
        clear({ dispatchPluginAction }) {
          dispatchPluginAction("state/replace", {
            display: "0",
            accumulator: 0,
            operation: null,
          });
        },
      },
    },
  },
}));
```

**Key points:**
- Pass data through `args` to reuse a single handler for multiple buttons
- Use `state/replace` when the state transition depends on many fields together
- Handler args are the second parameter — they come from `UIEventRef.args`

---

## Example 5: Shared Counter (Cross-Plugin Communication)

**Concepts:** Shared domains, `dispatchSharedAction`, reading shared state

A counter that publishes its value to the `counter-summary` shared domain so other plugins can see it.

```js
definePlugin(({ ui }) => ({
  id: "shared-counter",
  title: "Shared Counter",
  description: "Publishes count to counter-summary domain",
  initialState: { count: 0 },
  widgets: {
    main: {
      render({ pluginState, globalState }) {
        const count = Number(pluginState?.count ?? 0);
        const summary = globalState?.shared?.["counter-summary"];
        const total = Number(summary?.totalValue ?? 0);
        const instances = Number(summary?.instanceCount ?? 0);

        return ui.panel([
          ui.text("My count: " + count),
          ui.row([
            ui.badge("Total across " + instances + " instances: " + total),
          ]),
          ui.row([
            ui.button("-", { onClick: { handler: "decrement" } }),
            ui.button("Reset", {
              onClick: { handler: "reset" },
              variant: "destructive",
            }),
            ui.button("+", { onClick: { handler: "increment" } }),
          ]),
        ]);
      },
      handlers: {
        increment({ dispatchPluginAction, dispatchSharedAction, pluginState }) {
          const next = Number(pluginState?.count ?? 0) + 1;
          dispatchPluginAction("state/merge", { count: next });
          dispatchSharedAction("counter-summary", "set-instance", { value: next });
        },
        decrement({ dispatchPluginAction, dispatchSharedAction, pluginState }) {
          const next = Number(pluginState?.count ?? 0) - 1;
          dispatchPluginAction("state/merge", { count: next });
          dispatchSharedAction("counter-summary", "set-instance", { value: next });
        },
        reset({ dispatchPluginAction, dispatchSharedAction }) {
          dispatchPluginAction("state/replace", { count: 0 });
          dispatchSharedAction("counter-summary", "set-instance", { value: 0 });
        },
      },
    },
  },
}));
```

**Key points:**
- `dispatchSharedAction(domain, actionType, payload)` writes to a shared domain
- `globalState.shared["counter-summary"]` reads the current shared state
- Each handler dispatches **both** a local action and a shared action — the local action updates the plugin's own state, and the shared action updates the cross-plugin aggregate
- This works because custom plugins in the playground get all domain grants automatically

---

## Example 6: Dashboard (Read-Only Shared State)

**Concepts:** Read-only shared access, runtime-metrics, runtime-registry, table rendering

A monitoring dashboard that reads from multiple shared domains to display system status.

```js
definePlugin(({ ui }) => ({
  id: "dashboard",
  title: "System Dashboard",
  description: "Read-only view of runtime state",
  widgets: {
    main: {
      render({ globalState }) {
        const metrics = globalState?.shared?.["runtime-metrics"] ?? {};
        const registry = globalState?.shared?.["runtime-registry"] ?? [];
        const counter = globalState?.shared?.["counter-summary"] ?? {};
        const greeter = globalState?.shared?.["greeter-profile"] ?? {};

        const plugins = Array.isArray(registry) ? registry : [];

        return ui.column([
          ui.panel([
            ui.text("📊 Runtime Metrics"),
            ui.row([
              ui.badge("Plugins: " + (metrics.pluginCount ?? 0)),
              ui.badge("Dispatches: " + (metrics.dispatchCount ?? 0)),
              ui.badge("Last: " + (metrics.lastActionType ?? "none")),
              ui.badge(
                "Outcome: " + (metrics.lastOutcome ?? "—")
              ),
            ]),
          ]),

          ui.panel([
            ui.text("📦 Shared Domains"),
            ui.row([
              ui.badge("Counter total: " + (counter.totalValue ?? 0)),
              ui.badge("Greeter name: " + (greeter.name || "(empty)")),
            ]),
          ]),

          ui.panel([
            ui.text("🔌 Plugin Registry (" + plugins.length + ")"),
            plugins.length > 0
              ? ui.table({
                  headers: ["Instance", "Package", "Status", "Widgets"],
                  rows: plugins.map((p) => [
                    String(p.instanceId ?? ""),
                    String(p.packageId ?? ""),
                    String(p.status ?? ""),
                    String(p.widgets ?? 0),
                  ]),
                })
              : ui.text("No plugins loaded"),
          ]),
        ]);
      },
      handlers: {},
    },
  },
}));
```

**Key points:**
- A plugin with no handlers is a pure display widget — it re-renders when any state changes
- `runtime-registry` and `runtime-metrics` are read-only system domains
- Always check for `null`/`undefined` in shared state — domains might not be populated yet

---

## Example 7: Multi-Widget Plugin

**Concepts:** Multiple widgets, separate concerns

A plugin that exposes two separate widgets — one for controls, one for display.

```js
definePlugin(({ ui }) => ({
  id: "multi-widget",
  title: "Multi-Widget Demo",
  initialState: { items: ["Alpha", "Bravo", "Charlie"], selected: null },
  widgets: {
    controls: {
      render({ pluginState }) {
        const items = Array.isArray(pluginState?.items) ? pluginState.items : [];
        return ui.panel([
          ui.text("Controls"),
          ...items.map((item, i) =>
            ui.button("Select: " + item, {
              onClick: { handler: "select", args: i },
            })
          ),
          ui.button("Clear", {
            onClick: { handler: "clear" },
            variant: "destructive",
          }),
        ]);
      },
      handlers: {
        select({ dispatchPluginAction }, index) {
          dispatchPluginAction("state/merge", { selected: index });
        },
        clear({ dispatchPluginAction }) {
          dispatchPluginAction("state/merge", { selected: null });
        },
      },
    },
    display: {
      render({ pluginState }) {
        const items = Array.isArray(pluginState?.items) ? pluginState.items : [];
        const selected = pluginState?.selected;
        const label =
          selected !== null && selected !== undefined
            ? "Selected: " + items[selected]
            : "Nothing selected";

        return ui.panel([
          ui.text("Display"),
          ui.badge(label),
        ]);
      },
      handlers: {},
    },
  },
}));
```

**Key points:**
- Each widget has its own `render` and `handlers`, but they share the same `pluginState`
- Handlers in one widget can update state that another widget reads
- The host renders each widget independently — both get the same state snapshot

---

## Tips and Gotchas

### State is JSON-serialized
Your state crosses a serialization boundary (QuickJS → JSON → host). This means:
- No `Date` objects, `Map`, `Set`, or class instances — use plain objects and arrays
- Numbers come back as numbers, strings as strings, `null` as `null`
- `undefined` values are stripped during serialization

### Handlers must be synchronous
Handlers run inside QuickJS, which is a synchronous runtime. No `async/await`, no `setTimeout`, no `fetch`. The only way to cause effects is to dispatch intents.

### Always coerce values from state
State values may not be the type you expect after serialization:
```js
// Do this:
const count = Number(pluginState?.count ?? 0);
// Not this:
const count = pluginState.count; // might be undefined or a string
```

### Keep render functions pure
`render()` is called frequently. Don't put side effects in it. Build your UI tree and return it.

### The `args` parameter is flexible
You can pass any JSON-serializable value as handler args:
```js
// Single value
ui.button("Five", { onClick: { handler: "set", args: 5 } })
// Object
ui.button("Save", { onClick: { handler: "save", args: { draft: true } } })
// The handler receives it as the second parameter
save(ctx, args) { /* args = { draft: true } */ }
```
