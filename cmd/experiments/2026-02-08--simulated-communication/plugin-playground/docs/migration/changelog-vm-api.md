# VM API Migration Notes (WEBVM-003)

This document captures the runtime/package migration applied in WEBVM-003.

## Scope

The codebase was simplified to:

- one reusable runtime package: `packages/plugin-runtime`
- one playground application package: `client` and app wiring

Backward-compat path aliases for old runtime locations were removed.

## Import migration map

Update imports from app-local runtime files to `@runtime/*`:

- `client/src/lib/quickjsContracts.ts` -> `@runtime/contracts`
- `client/src/lib/uiTypes.ts` -> `@runtime/uiTypes`
- `client/src/lib/uiSchema.ts` -> `@runtime/uiSchema`
- `client/src/lib/dispatchIntent.ts` -> `@runtime/dispatchIntent`
- `client/src/lib/runtimeIdentity.ts` -> `@runtime/runtimeIdentity`
- `client/src/lib/quickjsRuntimeService.ts` -> `@runtime/runtimeService`
- `client/src/lib/quickjsSandboxClient.ts` -> `@runtime/worker/sandboxClient`
- `client/src/workers/quickjsRuntime.worker.ts` -> `@runtime/worker/runtime.worker`
- `client/src/store/store.ts` -> `@runtime/redux-adapter/store`

## Runtime behavior changes to account for

### 1) Capability policy is explicit and deny-by-default

Each plugin instance receives grants at registration:

- `readShared`
- `writeShared`
- `systemCommands` (reserved)

Missing write grants now produce denied shared dispatches with reason metadata.

### 2) Dispatch timeline is now first-class state

The runtime reducer now tracks bounded dispatch history (`MAX_TIMELINE_ENTRIES = 200`) with:

- scope (`plugin` or `shared`)
- action type
- instance/domain
- outcome (`applied`, `denied`, `ignored`)
- reason

Hosts can read it through `selectDispatchTimeline`.

### 3) Runtime projections include dispatch timestamps

`runtime-metrics` now includes `lastTimestamp`, useful for timeline-aligned diagnostics.

## Host migration checklist

1. Replace old imports with `@runtime/*` exports.
2. Keep registration and capability assignment in host code (`pluginRegistered(..., grants)`).
3. Route event intents through adapter helpers:
   - `dispatchPluginAction(...)`
   - `dispatchSharedAction(...)`
4. Rebuild render inputs with:
   - `selectPluginState(...)`
   - `selectGlobalStateForInstance(...)`
5. Surface runtime diagnostics:
   - `selectDispatchTimeline(...)`
   - `globalState.shared["runtime-metrics"]`
6. Remove reliance on any deleted app-local runtime file paths.

## Compatibility statement

This migration is intentionally breaking for old app-local runtime import paths. The supported API surface is now the `@runtime/*` exports from `packages/plugin-runtime`.
