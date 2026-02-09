# Changelog

## 2026-02-08

- Initial workspace created


## 2026-02-08

Created a deep architecture review covering plugin identity authority, action/state scoping gaps, and a phased implementation strategy for host-assigned instance IDs with capability-gated shared domains.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/01-plugin-action-and-state-scoping-architecture-review.md — Primary 15+ page analysis document


## 2026-02-08

Updated ticket index with summary, key deliverable links, and reMarkable upload location.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/index.md — Ticket landing page now points to analysis doc and uploaded PDF


## 2026-02-08

Updated design-doc 01 to adopt simplified v1 selector/action model (selectPluginState/selectGlobalState + dispatchPluginAction/dispatchGlobalAction with global dispatchId) and added design-doc 02 for real QuickJS isolation plus mock runtime removal plan.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/01-plugin-action-and-state-scoping-architecture-review.md — Simplified v1 model update with explicit pros/cons
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/02-quickjs-isolation-architecture-and-mock-runtime-removal-plan.md — New QuickJS isolation and mock path removal architecture
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/index.md — Updated landing page links and summary


## 2026-02-08

Uploaded combined reMarkable bundle containing design-doc 01 (simplified v1 scoping update) and design-doc 02 (QuickJS isolation/removal plan).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/index.md — Added direct link to combined reMarkable bundle


## 2026-02-08

Added design-doc 03: comprehensive vision and architecture explainer document covering the full end-state of the plugin playground system, written for newcomers


## 2026-02-08

Updated design-doc 03 (vision explainer) to reflect v1 unified runtime cleanup: removed references to dead code paths, updated all code examples to v1 API (pluginState/globalState, dispatchPluginAction/dispatchGlobalAction), added section on what was cleaned up, updated architecture diagram to show unified runtime slice, updated phase roadmap to mark Phase 1 as complete


## 2026-02-08

Updated design-doc 02 with repository reality-check notes, corrected related file paths, and linked execution handoff to WEBVM-002 implementation ticket.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/02-quickjs-isolation-architecture-and-mock-runtime-removal-plan.md — Added reality check and handoff section
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-002-QUICKJS-MIGRATION--migrate-plugin-playground-runtime-to-quickjs-worker-and-add-test-gates/design-doc/01-quickjs-migration-implementation-guide-and-test-strategy.md — Execution-level follow-on guide


## 2026-02-08

Added cross-ticket reMarkable bundle link for WEBVM-002 execution package that includes updated QuickJS architecture doc 02.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/index.md — Added WEBVM-002 bundle link


## 2026-02-08

Added design-doc 04: Phase 3-4 design brief with concrete instructions for architect to produce the multi-instance identity (packageId/instanceId) and capability model design, grounded in actual post-QuickJS codebase

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/design-doc/04-phase-3-4-design-brief-multi-instance-identity-and-capability-model.md — New design brief for Phase 3-4


## 2026-02-08

Step 2: migrated QuickJS contracts/runtime/worker/client to packageId+instanceId identity boundary (commit 414b68a)

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsContracts.ts — New runtime contract identity fields
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts — Instance-keyed VM service implementation
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/reference/01-diary.md — Recorded implementation step details and validation evidence


## 2026-02-09

Step 3: implemented instance-based store routing and Playground multi-instance lifecycle (commit 96c6225)

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx — Unique instance ID generation and per-instance rendering
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts — Package-based local reducer dispatch by instance
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/reference/01-diary.md — Recorded Step 3 details and validations


## 2026-02-09

Step 4: implemented shared-domain capability model, migrated presets to dispatchSharedAction, and expanded integration/e2e coverage (commit 709df40)

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/quickjsRuntimeService.ts — VM bootstrap shared dispatch API
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/store/store.ts — Phase 4 shared domains + grant enforcement
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/tests/e2e/quickjs-runtime.spec.ts — Regression tests for new behavior
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS--scope-plugin-actions-and-state-for-webvm/reference/01-diary.md — Recorded Step 4 implementation and validation details

