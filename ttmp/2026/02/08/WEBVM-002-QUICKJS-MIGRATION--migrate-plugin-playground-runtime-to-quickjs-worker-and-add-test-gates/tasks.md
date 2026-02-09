# Tasks

## TODO

- [ ] Capture baseline behavior and command outputs for preset load/render/event before runtime changes
- [x] Add shared worker/client RPC contract types in client/src/lib/quickjsContracts.ts
- [x] Add runtime UI tree validation utilities in client/src/lib/uiSchema.ts with kind-based enforcement
- [x] Add dispatch intent schema and host-side validation helpers in client/src/lib/dispatchIntent.ts
- [x] Create QuickJS worker entrypoint in client/src/workers/quickjsRuntime.worker.ts
- [x] Implement plugin runtime/context registry and deterministic dispose helpers in worker runtime
- [x] Implement VM bootstrap source with definePlugin capture and plugin metadata extraction
- [x] Implement loadPlugin RPC path with structured success/error responses
- [x] Implement render RPC path that accepts plugin/global state snapshots and returns validated UINode trees
- [x] Implement event RPC path that invokes handlers and emits plugin/global dispatch intents
- [x] Enforce runtime memory and stack limits and add interrupt/deadline timeout handling
- [x] Create main-thread worker RPC wrapper in client/src/lib/quickjsSandboxClient.ts
- [x] Route Playground plugin load/unload flows through quickjsSandboxClient
- [x] Route Playground render flow through worker RPC and validated UINode payloads
- [x] Route Playground event handling through worker dispatch intents and existing store dispatch wrappers
- [x] Remove pluginManager runtime path and delete client/src/lib/pluginManager.ts after cutover
- [x] Update any remaining runtime imports/types/comments that assume in-process execution
- [x] Add Vitest unit tests for contract validation, intent validation, and runtime error mapping
- [x] Add integration tests for worker load/render/event/dispose lifecycle behavior
- [x] Add Playwright configuration and browser install bootstrap for e2e runtime checks
- [x] Add Playwright scenarios for preset flows, dispatch behavior, sandbox isolation, and infinite-loop timeout recovery
- [x] Add package scripts for test:unit, test:integration, test:e2e, and test:migration gates
- [x] Run full migration test gate locally and record results in ticket changelog and diary
