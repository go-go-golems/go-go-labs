# UI DSL Reference

Plugins don't render React components directly. Instead, they build **UI trees** — plain JSON objects that describe what the UI should look like. The host's `WidgetRenderer` interprets these trees and renders them as real DOM elements.

This separation is fundamental to the plugin sandbox model: plugin code runs inside QuickJS (which has no DOM access), so UI must be expressed as serializable data that crosses the sandbox boundary.

## How Rendering Works

```
Plugin (QuickJS sandbox)              Host (browser)
┌─────────────────────┐              ┌─────────────────────┐
│                     │              │                     │
│  render(state) {    │   JSON       │  WidgetRenderer     │
│    return ui.column(├─────────────►│  interprets tree    │
│      ui.text("Hi"), │   UINode     │  into React/DOM     │
│      ui.button("+") │              │                     │
│    );               │              │                     │
│  }                  │              │                     │
└─────────────────────┘              └─────────────────────┘
```

Every call to a `ui.*` function returns a **UINode** — a plain object with a `kind` field and type-specific properties. The `render` function in your widget must return a single UINode (typically a layout container like `column` or `panel`).

## The `ui` Object

The `ui` object is provided by the `definePlugin` host context. It has the following builder functions:

### Layout Nodes

These nodes contain other nodes and control how they're arranged.

#### `ui.panel(children)`

A bordered container with padding. Use it to group related content.

```js
ui.panel([
  ui.text("Section Title"),
  ui.text("Some content here"),
])
```

**Renders as:** A rounded card with a subtle border and background.

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"panel"` | Always `"panel"` |
| `children` | `UINode[]` | Child nodes |

#### `ui.row(children)`

A horizontal flex container. Children are laid out left to right.

```js
ui.row([
  ui.button("Cancel"),
  ui.button("Save"),
])
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"row"` | Always `"row"` |
| `children` | `UINode[]` | Child nodes |

#### `ui.column(children)`

A vertical flex container. Children are stacked top to bottom. This is the most common top-level container.

```js
ui.column([
  ui.text("Line 1"),
  ui.text("Line 2"),
])
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"column"` | Always `"column"` |
| `children` | `UINode[]` | Child nodes |

### Content Nodes

#### `ui.text(content)`

A text paragraph. The simplest UI element.

```js
ui.text("Hello, world!")
ui.text("Counter: " + state.count)
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"text"` | Always `"text"` |
| `text` | `string` | The text content |

#### `ui.badge(content)`

An inline badge — useful for status labels, tags, or small metadata.

```js
ui.badge("Status: Active")
ui.badge("R/W")
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"badge"` | Always `"badge"` |
| `text` | `string` | The badge content |

### Interactive Nodes

These nodes produce events that flow through the dispatch lifecycle.

#### `ui.button(label, options)`

A clickable button. The `onClick` property is a **UIEventRef** that names a handler function.

```js
ui.button("Click me", {
  onClick: { handler: "handleClick" }
})

ui.button("Set to 5", {
  onClick: { handler: "setValue", args: { value: 5 } }
})

ui.button("Delete", {
  onClick: { handler: "delete" },
  variant: "destructive"
})
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"button"` | Always `"button"` |
| `props.label` | `string` | Button text |
| `props.onClick` | `UIEventRef?` | Handler reference (see below) |
| `props.variant` | `string?` | `"destructive"` for danger-styled buttons |

#### `ui.input(options)`

A text input field. The `onChange` handler fires on every keystroke.

```js
ui.input({
  value: state.name,
  placeholder: "Enter your name",
  onChange: { handler: "nameChanged" }
})
```

When the user types, the handler receives `{ value: "current text" }` as its args.

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"input"` | Always `"input"` |
| `props.value` | `string` | Current input value |
| `props.placeholder` | `string?` | Placeholder text |
| `props.onChange` | `UIEventRef?` | Handler for value changes |

#### `ui.counter(options)`

A pre-built counter widget with increment/decrement buttons and a numeric display.

```js
ui.counter({
  value: state.count,
  onIncrement: { handler: "increment" },
  onDecrement: { handler: "decrement" }
})
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"counter"` | Always `"counter"` |
| `props.value` | `number` | Current value |
| `props.onIncrement` | `UIEventRef?` | Handler for + button |
| `props.onDecrement` | `UIEventRef?` | Handler for − button |

### Data Nodes

#### `ui.table(options)`

A data table with headers and rows.

```js
ui.table({
  headers: ["Name", "Status", "Plugins"],
  rows: [
    ["Counter", "loaded", 2],
    ["Greeter", "loaded", 1],
  ]
})
```

| Property | Type | Description |
|----------|------|-------------|
| `kind` | `"table"` | Always `"table"` |
| `props.headers` | `string[]` | Column headers |
| `props.rows` | `any[][]` | Row data (each row is an array of cells) |

## UIEventRef

Every interactive node uses `UIEventRef` objects to link UI interactions to handler functions:

```ts
type UIEventRef = {
  handler: string;   // Name of the handler function in the widget
  args?: any;        // Optional arguments passed to the handler
};
```

The `handler` string must match a key in the widget's `handlers` object. The `args` value is passed to the handler as its second argument.

## Composition Patterns

### Conditional rendering

```js
render(state) {
  const items = [ui.text("Always shown")];

  if (state.isLoggedIn) {
    items.push(ui.text("Welcome back, " + state.name));
  } else {
    items.push(ui.button("Log in", { onClick: { handler: "login" } }));
  }

  return ui.column(items);
}
```

### Dynamic lists

```js
render(state) {
  const rows = state.items.map(item => [item.name, item.status]);
  return ui.column([
    ui.text("Items: " + state.items.length),
    ui.table({ headers: ["Name", "Status"], rows }),
  ]);
}
```

### Nested layouts

```js
render(state) {
  return ui.panel([
    ui.text("Header"),
    ui.row([
      ui.column([
        ui.text("Left"),
        ui.button("A", { onClick: { handler: "a" } }),
      ]),
      ui.column([
        ui.text("Right"),
        ui.button("B", { onClick: { handler: "b" } }),
      ]),
    ]),
  ]);
}
```

## Type Definition

For reference, here's the complete TypeScript type:

```ts
type UIEventRef = { handler: string; args?: any };

type UINode =
  | { kind: "panel" | "row" | "column"; props?: any; children?: UINode[] }
  | { kind: "text" | "badge"; props?: any; text: string }
  | { kind: "button"; props: { label: string; onClick?: UIEventRef; variant?: string } }
  | { kind: "input"; props: { value: string; placeholder?: string; onChange?: UIEventRef } }
  | { kind: "counter"; props: { value: number; onIncrement?: UIEventRef; onDecrement?: UIEventRef } }
  | { kind: "table"; props: { headers: string[]; rows: any[][] } };
```
