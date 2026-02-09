# Capability Model and Shared Domain Reference

This document defines how shared access control works in `plugin-runtime`, and what each domain currently supports.

## Capability model

Capabilities are assigned per plugin instance at registration time:

```ts
type CapabilityGrants = {
  readShared: SharedDomainName[];
  writeShared: SharedDomainName[];
  systemCommands: string[];
};
```

Current behavior:

- Read access gates what appears in `globalState.shared` for the instance.
- Write access gates `dispatchSharedAction(domain, actionType, payload)` application.
- Missing write grant results in `outcome = denied` with reason `missing-write-grant:<domain>`.
- `systemCommands` is reserved for host-level command policy and is not actively consumed by reducers yet.
- Default is deny-by-default (`readShared = []`, `writeShared = []`, `systemCommands = []`).

## Dispatch outcomes

Every scoped dispatch is tracked with one of:

- `applied`: reducer accepted the action and state changed
- `denied`: policy blocked the action (for example, missing grant)
- `ignored`: action/domain pair is unsupported or no reducer matched

These outcomes are visible in:

- runtime metrics (`globalState.shared["runtime-metrics"]`)
- inspector timeline (`INSPECTOR -> TIMELINE`)

## Shared domain reference

### `counter-summary`

- Read grant exposes:
  - `totalValue: number`
  - `instanceCount: number`
  - `lastUpdatedInstanceId: string | null`
- Write grant allows:
  - `actionType = "set-instance"` with payload `{ value: number }`

### `greeter-profile`

- Read grant exposes:
  - `name: string`
  - `lastUpdatedInstanceId: string | null`
- Write grant allows:
  - `actionType = "set-name"` with payload as string-like value

### `runtime-registry`

- Read-only domain (derived by runtime)
- Exposes loaded plugin summaries:
  - `instanceId`
  - `packageId`
  - `title`
  - `status`
  - `enabled`
  - `widgets`
- No supported shared write actions

### `runtime-metrics`

- Read-only domain (derived by runtime)
- Exposes runtime counters and latest dispatch metadata:
  - `pluginCount`
  - `dispatchCount`
  - `lastTimestamp`
  - `lastDispatchId`
  - `lastScope`
  - `lastActionType`
  - `lastOutcome`
  - `lastReason`
- No supported shared write actions

## Practical policy guidance

- Grant only the domains a plugin must read/write.
- Prefer read-only dashboards over write-enabled plugins.
- Use timeline filters (`scope`, `outcome`, `instance`) during debugging to verify policy behavior quickly.
