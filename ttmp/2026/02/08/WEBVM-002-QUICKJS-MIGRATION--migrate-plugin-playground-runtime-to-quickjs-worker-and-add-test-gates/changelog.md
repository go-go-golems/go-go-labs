# Changelog

## 2026-02-08

- Initial workspace created


## 2026-02-08

Created execution ticket artifacts for QuickJS migration: authored detailed implementation/test guide, added step-by-step migration tasks, and recorded a detailed research diary.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/design-doc/01-quickjs-migration-implementation-guide-and-test-strategy.md — Detailed migration implementation and Playwright test strategy
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/index.md — Updated landing page with migration scope and links
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/reference/01-diary.md — Detailed planning and troubleshooting diary
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/tasks.md — Added granular execution checklist


## 2026-02-08

Uploaded bundled PDF to reMarkable containing WEBVM-001 doc 02, WEBVM-002 implementation guide, WEBVM-002 tasks, and WEBVM-002 diary; verified remote artifact path.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/index.md — Added reMarkable upload link
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/reference/01-diary.md — Diary updated with upload and verification details


## 2026-02-08

Implemented QuickJS runtime cutover: added worker/runtime contracts and validators, switched Playground to async worker RPC for load/render/event, removed legacy pluginManager/new Function path, and validated with pnpm check + pnpm build.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/dispatchIntent.ts — Dispatch intent validation helpers
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/pluginManager.ts — Deleted legacy in-process runtime
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts — Shared worker request/response and runtime types
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsSandboxClient.ts — Main-thread RPC client for worker runtime
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiSchema.ts — UINode validation for runtime responses
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx — Cutover from pluginManager to quickjsSandboxClient
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/workers/quickjsRuntime.worker.ts — QuickJS runtime service with load/render/event/dispose and limits
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vite.config.ts — Set worker format=es to support worker code-splitting build


## 2026-02-08

Completed migration test gates: added unit/integration/e2e suites, Playwright config, migration scripts, fixed selector stability render loop, and validated full pipeline with pnpm test:migration.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/dispatchIntent.test.ts — Dispatch intent validation unit tests
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.integration.test.ts — Runtime load/render/event/dispose/timeout integration tests
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts — Extracted testable runtime service
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/uiSchema.test.ts — UINode validation unit tests
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts — Memoized selectors to stop render-loop regression
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json — Added migration test scripts
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/playwright.config.ts — Playwright webServer/test project configuration
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/tests/e2e/quickjs-runtime.spec.ts — Playwright runtime behavior tests

