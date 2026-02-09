# Changelog

## 2026-02-09

- Initial workspace created


## 2026-02-09

Added deep-pass design document with concrete removal candidates, UI overhaul direction, package extraction plan, and documentation backlog.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/design-doc/01-deep-pass-ui-overhaul-runtime-packaging-and-docs-plan.md — Primary WEBVM-003 planning artifact


## 2026-02-09

Adjusted architecture plan to a two-package model: one reusable plugin-runtime package (with internal core/worker/redux-adapter modules) and one plugin-playground app package.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/design-doc/01-deep-pass-ui-overhaul-runtime-packaging-and-docs-plan.md — Package strategy revised from multi-package split to 1 runtime + 1 app


## 2026-02-09

Published deep-pass refresh grounded in current codebase, updated task breakdown to reflect completed cleanup work, and set the refreshed doc as the active implementation source of truth.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/design-doc/02-deep-pass-refresh-current-codebase-audit-and-ui-runtime-docs-roadmap.md — Current-state audit with concrete findings and implementation roadmap
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/index.md — Key link switched to refreshed deep-pass document
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Updated execution tasks with completed cleanup and phased follow-ups


## 2026-02-09

Uploaded refreshed deep-pass design document to reMarkable for review: /ai/2026/02/09/WEBVM-003-DEVX/WEBVM-003 Deep Pass Refresh

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/design-doc/02-deep-pass-refresh-current-codebase-audit-and-ui-runtime-docs-roadmap.md — Source markdown used for uploaded review PDF


## 2026-02-09

Completed cleanup task 2: removed remaining debug/template leftovers by deleting WidgetRenderer debug globals/logging, removing index.html analytics placeholders/comment block, deleting unused useMobile hook, and dropping unused fs import from Vite config.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/index.html — Removed stale template comment block and unresolved analytics placeholders
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/WidgetRenderer.tsx — Removed debug console/global writes from button handler
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/hooks/useMobile.tsx — Deleted unused hook
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vite.config.ts — Removed unused node:fs import


## 2026-02-09

Completed theme-stack unification task 3 by removing next-themes and wiring Sonner to the existing app ThemeContext, leaving a single theme-provider model in the playground app.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/ui/sonner.tsx — Toaster now reads theme from local ThemeContext
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json — Removed next-themes dependency
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/pnpm-lock.yaml — Lockfile updated after dependency removal
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 3 complete


## 2026-02-09

Completed task 4 by scaffolding packages/plugin-runtime and migrating contracts, runtime identity, schema/intent validators, and QuickJS runtime service into the new package; rewired app imports and test config to consume package sources.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/contracts.ts — Migrated runtime request/response and intent contracts
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/runtimeIdentity.ts — Migrated instance ID generation helper
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/runtimeService.ts — Migrated QuickJS runtime core
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vite.config.ts — Added @runtime alias for package source consumption
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vitest.config.ts — Unit tests now include package runtime tests
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/vitest.integration.config.ts — Integration tests now include package runtime tests
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 4 complete


## 2026-02-09

Completed task 5 by moving worker wrapper/client transport into plugin-runtime package and adding host adapter interfaces for non-UI embedding.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx — App now consumes sandbox client from runtime package
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/hostAdapter.ts — Added runtime host adapter interfaces for embedding
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/worker/runtime.worker.ts — Moved runtime worker into package worker module
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/worker/sandboxClient.ts — Moved sandbox client into runtime package worker module
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 5 complete


## 2026-02-09

Completed task 6 by moving Redux runtime reducer/policy/selector logic into plugin-runtime's internal redux-adapter module and switching app consumers to package imports.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/App.tsx — Provider store import now comes from runtime package
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts — SharedDomainName now sourced from runtime redux adapter
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx — Runtime dispatch/selectors now imported from runtime redux adapter
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/package.json — Added redux-adapter export path
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/redux-adapter/store.ts — Moved runtime redux adapter logic into runtime package
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 6 complete


## 2026-02-09

Completed task 7 by refactoring Playground into modular workbench UI shells (Catalog, Workspace, Inspector) while preserving runtime behavior.

### Related Files

- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/CatalogShell.tsx — Developer catalog shell for presets and loaded instances
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/InspectorShell.tsx — Inspector shell for rendered widgets and runtime feedback
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/WorkspaceShell.tsx — Workspace shell for custom plugin authoring
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx — Orchestration page now composes modular shell components
- /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 7 complete


## 2026-02-09 - Completed task 8

Implemented runtime timeline and shared-domain inspector panels with filtering and bounded timeline retention in runtime state.

### Related Files

- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/InspectorShell.tsx — Added widgets/timeline/shared tabs with filters and shared-state panel
- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx — Wired timeline/shared selectors into Inspector shell props
- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/redux-adapter/store.ts — Added timeline entry model
- ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 8 complete


## 2026-02-09 - Completed task 9

Published plugin authoring quickstart and capability/shared-domain reference docs aligned with current runtime policy and intent outcomes.

### Related Files

- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/architecture/capability-model.md — Capability policy and shared domain/action reference
- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/plugin-authoring/quickstart.md — New authoring quickstart with working custom-plugin patterns and handler context
- ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 9 complete


## 2026-02-09 - Completed task 10

Published runtime embedding guide with host-loop examples and migration notes documenting breaking import-path changes to the plugin-runtime package.

### Related Files

- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/README.md — Top-level docs navigation index
- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/migration/changelog-vm-api.md — WEBVM-003 migration map and compatibility statement
- cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/runtime/embedding.md — Runtime embedding patterns for direct service and worker-backed hosts
- ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/tasks.md — Marked task 10 complete

