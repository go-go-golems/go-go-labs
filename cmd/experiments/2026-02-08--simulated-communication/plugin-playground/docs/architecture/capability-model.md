# Capability Model

The Plugin Playground uses a **capability-based security model** to control how plugins interact with shared state. Rather than giving every plugin full access to everything, each plugin instance receives explicit **grants** that determine what it can read and write.

This model exists because plugins are untrusted code — they run in a QuickJS sandbox that prevents direct DOM access, but they still need a way to communicate with each other. Capabilities are the controlled channel for that communication.

## How It Works

When a plugin instance is registered with the runtime, it receives a set of **capability grants**:

```ts
interface CapabilityGrants {
  readShared: SharedDomainName[];   // which shared domains this instance can see
  writeShared: SharedDomainName[];  // which shared domains this instance can modify
  systemCommands: string[];         // reserved for future host-level commands
}
```

These grants are assigned by the host (the Playground UI) at load time and cannot be changed by the plugin itself.

### The Default: Deny Everything

If no grants are specified, a plugin gets empty arrays for everything:

```ts
const DEFAULT_GRANTS = {
  readShared: [],
  writeShared: [],
  systemCommands: [],
};
```

This means a plugin with no grants:
- Cannot see any shared domain data in its `globalState.shared`
- Will have all shared dispatch attempts **denied**
- Can still use its own local state freely

### Read Grants

A read grant for a domain means the plugin can see that domain's data in its `globalState.shared` object during rendering:

```
Plugin with readShared: ["counter-summary"]

globalState.shared = {
  "counter-summary": {      ← visible because of read grant
    totalValue: 15,
    instanceCount: 3,
    ...
  }
  // "greeter-profile" is NOT present — no read grant
}
```

Read grants are checked every time the host builds the projected `globalState` for a plugin instance. Domains without a read grant are simply omitted from the projection.

### Write Grants

A write grant for a domain means the plugin can dispatch shared actions targeting that domain:

```js
// In a handler:
dispatchSharedAction("counter-summary", "set-instance", { value: 5 });
```

If the plugin does NOT have a write grant for `counter-summary`, this dispatch will be:
- **Denied** with reason `missing-write-grant:counter-summary`
- Recorded in the timeline so you can see exactly what happened

## Dispatch Outcomes

Every dispatch (both plugin-scoped and shared-scoped) produces one of three outcomes:

| Outcome | Meaning | When it happens |
|---------|---------|-----------------|
| `applied` | State was changed | Reducer matched and processed the action |
| `denied` | Policy blocked the action | Missing write grant for the target domain |
| `ignored` | Action was not recognized | No reducer handles this domain/action combination |

These outcomes are visible in the DevTools Timeline tab. Filtering by `denied` is the fastest way to debug capability issues.

```
┌──────────┐     ┌─────────────┐     ┌──────────┐
│ Dispatch │────►│ Policy Gate │────►│ Reducer  │
│ Intent   │     │             │     │          │
└──────────┘     └──────┬──────┘     └─────┬────┘
                        │                   │
                   No grant?           No match?
                        │                   │
                        ▼                   ▼
                    "denied"           "ignored"
```

## Shared Domains

A **shared domain** is a named piece of state that lives outside any individual plugin. Multiple plugins can read from and write to the same domain, enabling cross-plugin communication.

### `counter-summary`

Aggregates counter values across all Counter plugin instances.

**State shape:**
```ts
{
  valuesByInstance: Record<string, number>,  // per-instance values
  totalValue: number,                        // sum of all values
  instanceCount: number,                     // number of contributing instances
  lastUpdatedInstanceId: string | null       // who last wrote
}
```

**Supported write actions:**

| Action | Payload | Effect |
|--------|---------|--------|
| `set-instance` | `{ value: number }` | Sets this instance's counter value and recalculates totals |

### `greeter-profile`

A shared name that all Greeter plugins can read.

**State shape:**
```ts
{
  name: string,                            // the shared name
  lastUpdatedInstanceId: string | null     // who last wrote
}
```

**Supported write actions:**

| Action | Payload | Effect |
|--------|---------|--------|
| `set-name` | `string` | Updates the shared name |

### `runtime-registry`

A **read-only** domain automatically maintained by the runtime. It provides a list of all loaded plugin instances.

**State shape:**
```ts
Array<{
  instanceId: string,
  packageId: string,
  title: string,
  status: "loaded" | "error",
  enabled: boolean,
  widgets: number
}>
```

No write actions are supported. Any attempt to write to this domain will be `ignored`.

### `runtime-metrics`

A **read-only** domain providing runtime telemetry.

**State shape:**
```ts
{
  pluginCount: number,
  dispatchCount: number,
  lastTimestamp: number | null,
  lastDispatchId: string | null,
  lastScope: "plugin" | "shared" | null,
  lastActionType: string | null,
  lastOutcome: "applied" | "denied" | "ignored" | null,
  lastReason: string | null
}
```

No write actions are supported.

## Capability Patterns

### Pattern: Read-Only Dashboard

A plugin that monitors the system but never modifies shared state:

```js
capabilities: {
  readShared: ["counter-summary", "runtime-registry", "runtime-metrics"],
  writeShared: [],
}
```

This is the safest pattern. The plugin can see everything but change nothing.

### Pattern: Read-Write Participant

A plugin that both reads and writes to a domain:

```js
capabilities: {
  readShared: ["counter-summary"],
  writeShared: ["counter-summary"],
}
```

This is the Counter plugin's pattern. It reads the aggregate totals to display them, and writes its own value when the user interacts.

### Pattern: Cross-Domain Reader

A plugin that reads from multiple domains to build a composite view:

```js
capabilities: {
  readShared: ["counter-summary", "greeter-profile"],
  writeShared: [],
}
```

### Pattern: Custom Plugin (Playground Default)

When you write custom code in the editor and click Run, the playground grants **all** shared domains:

```js
capabilities: {
  readShared: ["counter-summary", "greeter-profile", "runtime-registry", "runtime-metrics"],
  writeShared: ["counter-summary", "greeter-profile", "runtime-registry", "runtime-metrics"],
}
```

This is intentionally permissive because the playground is a sandbox for experimentation. In a production embedding, you would assign grants based on the plugin's declared needs.

## Debugging Capability Issues

If a plugin's shared dispatch is being denied:

1. Open DevTools → **Timeline** tab
2. Filter by **Outcome: denied**
3. Look at the **Reason** column — it will say `missing-write-grant:<domain>`
4. Check the plugin's capability grants in DevTools → **Capabilities** tab
5. If the plugin should have the grant, update its `capabilities` declaration

The Capabilities tab shows a grid of all instances × all domains with checkmarks for read/write access, making it easy to spot missing grants at a glance.
