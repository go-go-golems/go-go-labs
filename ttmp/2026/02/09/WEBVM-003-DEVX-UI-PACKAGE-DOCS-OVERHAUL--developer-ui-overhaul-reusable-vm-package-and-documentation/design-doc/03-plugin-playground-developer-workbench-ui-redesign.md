---
Title: Plugin Playground Developer Workbench - UI Redesign
Ticket: WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/WidgetRenderer.tsx
      Note: Widget renderer - moved from inspector to live preview pane
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/CatalogShell.tsx
      Note: Current left panel - redesigned as sidebar with tree navigation
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/InspectorShell.tsx
      Note: Current right panel - redesigned as bottom devtools panel
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/WorkspaceShell.tsx
      Note: Current center panel - redesigned as dominant editor+preview pane
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: Preset plugin catalog - feeds the sidebar tree
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: |-
        Current orchestration component - to be replaced by Workbench layout
        Current orchestration component - replaced by WorkbenchLayout in redesign
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/README.md
      Note: Docs index - bundled into DocsPanel via ?raw import
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/architecture/capability-model.md
      Note: Capability model doc - bundled into DocsPanel
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/migration/changelog-vm-api.md
      Note: Migration changelog - bundled into DocsPanel
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/plugin-authoring/quickstart.md
      Note: Plugin authoring quickstart - primary doc for DocsPanel
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/docs/runtime/embedding.md
      Note: Embedding guide - bundled into DocsPanel
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/redux-adapter/store.ts
      Note: Redux store with state/policy/selectors - data source for all inspector panels
    - Path: cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/uiTypes.ts
      Note: UINode types defining the widget vocabulary
ExternalSources: []
Summary: Complete UI redesign specification for the plugin-playground developer workbench, with ASCII wireframes for every screen state, component decomposition, and interaction flows.
LastUpdated: 2026-02-09T09:02:00Z
WhatFor: Implementation blueprint for the plugin-playground UI overhaul - detailed enough to code directly from
WhenToUse: Reference during implementation of the new workbench layout, component creation, and interaction wiring
---



# Plugin Playground Developer Workbench — UI Redesign

## Executive Summary

The current plugin-playground UI is a 3-equal-column layout that treats plugin loading, code editing, and inspection as equal-weight activities. In practice, the developer's primary activity is **writing plugin code and seeing the result** — everything else is supporting context. The redesign shifts to an **IDE-like workbench layout** with:

- A narrow **sidebar** for navigation (plugin catalog + loaded instances)
- A dominant **center pane** split between code editor and live preview
- A collapsible **bottom panel** for developer tools (timeline, state, capabilities, errors)
- A **top toolbar** for runtime controls and global status

The goal: make this a **developer-oriented UI** on top of the reusable `plugin-runtime` package, optimized for the edit→load→interact→inspect loop.

---

## 1. Problems with the Current UI

### 1.1 Current Layout (As-Is)

```
┌─────────────────────────────────────────────────────────────────────────┐
│  PLUGIN PLAYGROUND                                                       │
│  Unified Runtime v1 - Plugin/Global State and Action Scoping             │
├──────────────────────┬──────────────────────┬────────────────────────────┤
│  CATALOG             │  WORKSPACE           │  INSPECTOR                 │
│                      │                      │                            │
│  [Counter        ]   │  ┌────────────────┐  │  [WIDGETS] [TIMELINE]      │
│  [Calculator     ]   │  │ textarea       │  │  [SHARED]                  │
│  [Status Dash    ]   │  │                │  │                            │
│  [Greeter        ]   │  │ (no syntax     │  │  Counter [abc-123]         │
│  [Greeter Shared ]   │  │  highlighting) │  │  ┌─────────────────────┐   │
│  [Runtime Monitor]   │  │                │  │  │ Counter: 0          │   │
│                      │  │                │  │  │ [Shared total: 0]   │   │
│  ─── LOADED ───      │  │                │  │  │ [-] [Reset] [+]     │   │
│  Counter [abc-123] X │  └────────────────┘  │  └─────────────────────┘   │
│                      │  [  LOAD PLUGIN  ]   │                            │
│                      │                      │  (timeline/shared tabs     │
│                      │  (error display)     │   hidden behind tabs)      │
│                      │                      │                            │
├──────────────────────┴──────────────────────┴────────────────────────────┤
│  width: 33%            width: 33%            width: 33%                  │
└─────────────────────────────────────────────────────────────────────────┘
```

### 1.2 Specific Problems

| # | Problem | Impact |
|---|---------|--------|
| 1 | **Equal-width 3-column layout** | Code editor gets only 33% — too narrow to write real plugins |
| 2 | **Widget preview is in the Inspector** (right panel) | Edit code in center, see result in right panel — eyes bounce left↔right constantly |
| 3 | **No syntax highlighting** | Plain textarea; hard to write/read JavaScript |
| 4 | **No visible plugin state** | Can only see state through widget rendering or raw JSON dump |
| 5 | **No capability visualization** | No way to see what grants a plugin has without reading the code |
| 6 | **Timeline is behind a tab** | The most useful debugging tool is hidden; have to click to switch |
| 7 | **Shared state is raw JSON** | `JSON.stringify(sharedState, null, 2)` — no structure, no domain separation |
| 8 | **No per-instance focus** | Can't zoom into one plugin's state, timeline, and capabilities |
| 9 | **Catalog is a flat button list** | No description, no capability badges, no visual distinction between preset types |
| 10 | **No keyboard shortcuts** | No Ctrl+Enter to load, no quick navigation |
| 11 | **Preset code not loadable into editor** | Can't inspect/modify preset code — only load as-is |
| 12 | **No error context** | Errors show as red text string with no stack trace or action context |

---

## 2. Proposed Layout: Developer Workbench

### 2.1 Master Layout (Proposed)

The layout follows an IDE convention: narrow sidebar, dominant editor, bottom devtools.

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ ▣ PLUGIN WORKBENCH           [■ 3 plugins] [↻ 47 dispatches] [⚡ healthy]  [☰] │
├────────┬────────────────────────────────────────────────────────────────────────┤
│SIDEBAR │  EDITOR + PREVIEW PANE                                                │
│        │                                                                        │
│ 📦 CAT │  ┌─── editor tabs ───────────────────────────────────────────────────┐ │
│ ├ Coun │  │ [custom.js ×] [counter.js] [greeter.js]                     [▶RUN]│ │
│ ├ Calc │  ├───────────────────────────────┬───────────────────────────────────┤ │
│ ├ Dash │  │ CODE EDITOR (60%)             │ LIVE PREVIEW (40%)               │ │
│ ├ Gree │  │                               │                                  │ │
│ ├ GrSt │  │  1│ definePlugin(({ ui }) =>  │  ┌──────────────────────────┐    │ │
│ └ RtMo │  │  2│   return {                │  │ Counter: 5               │    │ │
│        │  │  3│     id: "counter",        │  │ ┌Shared total: 5┐       │    │ │
│ 🔌 RUN │  │  4│     title: "Counter",     │  │ ┌Instances: 1───┐       │    │ │
│ ├ ● co │  │  5│     initialState: {       │  │ [−]  [Reset]  [+]      │    │ │
│ │  abc1 │  │  6│       value: 0           │  └──────────────────────────┘    │ │
│ │  R/W: │  │  7│     },                   │                                  │ │
│ │  ctr- │  │  8│     widgets: {           │  ┌──────────────────────────┐    │ │
│ ├ ● gr │  │  9│       counter: {          │  │ Hello, World!            │    │ │
│ │  def4 │  │ 10│         render({ ...     │  │ [___________________]    │    │ │
│ │  R/W: │  │ 11│           ...            │  └──────────────────────────┘    │ │
│ │  gre- │  │   │                          │                                  │ │
│        │  │   │ (syntax highlighted,      │  (live-updates on state change)  │ │
│        │  │   │  line numbers,            │                                  │ │
│        │  │   │  monospace font)          │                                  │ │
│        │  └───┴───────────────────────────┴──────────────────────────────────┘ │
│        ├────────────────────────────────────────────────────────────────────────┤
│        │  DEVTOOLS PANEL (collapsible, drag-resizable)                         │
│        │  [Timeline▾] [State] [Capabilities] [Errors] [Shared] [📖Docs] [▲▼] │
│        │                                                                        │
│        │  ┌─ Timeline ──────────────────────────────────────────────────────┐  │
│        │  │ scope:[all▾] outcome:[all▾] instance:[________] domain:[all▾]  │  │
│        │  │                                                                 │  │
│        │  │  09:01:23.456  plugin   applied  counter/increment   abc-123   │  │
│        │  │  09:01:23.457  shared   applied  counter-summary/set  abc-123  │  │
│        │  │  09:01:22.100  shared   denied   greeter/set-name     def-456  │  │
│        │  │                         ^^^^^^   missing-write-grant            │  │
│        │  │  09:01:21.050  plugin   applied  greeter/nameChanged  def-456  │  │
│        │  └─────────────────────────────────────────────────────────────────┘  │
│        │                                                                        │
└────────┴────────────────────────────────────────────────────────────────────────┘
```

### 2.2 Layout Dimensions and Breakpoints

```
┌──────────────────────────────────────────────────────────────────┐
│                        LAYOUT GRID                               │
│                                                                  │
│  Sidebar:  240px fixed (collapsible to 48px icon-only)          │
│  Editor:   60% of remaining width                                │
│  Preview:  40% of remaining width                                │
│  Devtools: 280px default height (drag-resizable, min 120px)     │
│                                                                  │
│  ┌──────┬─────────────────────┬──────────────────┐              │
│  │240px │     60% remain      │   40% remain     │              │
│  │      │                     │                   │              │
│  │      │                     │                   │  flex-1      │
│  │      │                     │                   │              │
│  │      ├─────────────────────┴──────────────────┤              │
│  │      │          devtools: 280px               │              │
│  └──────┴────────────────────────────────────────┘              │
│                                                                  │
│  Mobile (<768px):  sidebar hidden, full-width stacked           │
│  Tablet (768-1024): sidebar collapsed, editor full, preview tab │
│  Desktop (>1024):  full layout as shown                         │
└──────────────────────────────────────────────────────────────────┘
```

---

## 3. Component Breakdown: Sidebar

### 3.1 Sidebar — Expanded State

```
┌──────────────────────┐
│ ▣ WORKBENCH     [«]  │  ← collapse toggle
├──────────────────────┤
│                      │
│ 📦 CATALOG           │  ← collapsible section
│ ┌──────────────────┐ │
│ │ ► Counter        │ │  ← click to load into editor + load
│ │   counter ∙ R/W  │ │  ← package id ∙ capability summary
│ ├──────────────────┤ │
│ │ ► Calculator     │ │
│ │   calculator     │ │  ← no shared grants → no badge
│ ├──────────────────┤ │
│ │ ► Status Dash    │ │
│ │   status ∙ R    │ │  ← read-only grants
│ ├──────────────────┤ │
│ │ ► Greeter        │ │
│ │   greeter ∙ R/W  │ │
│ ├──────────────────┤ │
│ │ ► Greeter Shared │ │
│ │   greet-sh ∙ R  │ │
│ ├──────────────────┤ │
│ │ ► Runtime Mon.   │ │
│ │   rt-mon ∙ R    │ │
│ └──────────────────┘ │
│                      │
│ 🔌 RUNNING (3)       │  ← collapsible section
│ ┌──────────────────┐ │
│ │ ● Counter        │ │  ← ● = green dot (loaded/healthy)
│ │   abc-1234       │ │  ← truncated instance ID
│ │   ├ R: ctr-sum   │ │  ← read grants
│ │   └ W: ctr-sum   │ │  ← write grants
│ │            [✕]   │ │  ← unload button
│ ├──────────────────┤ │
│ │ ● Greeter        │ │
│ │   def-5678       │ │
│ │   ├ R: grt-prof  │ │
│ │   └ W: grt-prof  │ │
│ │            [✕]   │ │
│ ├──────────────────┤ │
│ │ ● Status Dash    │ │
│ │   ghi-9012       │ │
│ │   └ R: ctr,rt,rg │ │
│ │            [✕]   │ │
│ └──────────────────┘ │
│                      │
│ ─────────────────    │
│ [+ New Plugin]       │  ← opens blank editor tab
│                      │
└──────────────────────┘
```

### 3.2 Sidebar — Collapsed State

```
┌────┐
│ [»]│  ← expand toggle
├────┤
│ 📦 │  ← catalog icon (tooltip: "Catalog")
│    │
│ 🔌 │  ← running icon with count badge
│ (3)│
│    │
│ [+]│  ← new plugin
└────┘
```

### 3.3 Sidebar Interactions

| Action | Behavior |
|--------|----------|
| Click preset in CATALOG | Load preset code into new editor tab + auto-run `loadPlugin` |
| Click preset again | Create another instance (same packageId, new instanceId) |
| Click running instance | Focus that instance: select its editor tab, highlight its widgets in preview, filter timeline to that instance |
| Click ✕ on instance | Confirm dialog → `disposePlugin` + `pluginRemoved` |
| Click [+ New Plugin] | Open blank editor tab with starter template |
| Click [«]/[»] | Toggle sidebar width between 240px and 48px |

---

## 4. Component Breakdown: Editor + Preview Pane

### 4.1 Editor Pane — With Tabs and Controls

```
┌─── Editor Tab Bar ─────────────────────────────────────────────────────┐
│ [custom.js ×] [counter.js ●] [greeter.js ●]              [▶ RUN] [⟳] │
│               ^^^^^ dot = has unsaved changes                          │
└────────────────────────────────────────────────────────────────────────┘
┌─── Code Editor ───────────────────────────────────────────────────────┐
│  1 │ definePlugin(({ ui }) => {                                       │
│  2 │   return {                                                       │
│  3 │     id: "my-plugin",                                             │
│  4 │     title: "My Plugin",                                          │
│  5 │     description: "Custom plugin",                                │
│  6 │     initialState: { count: 0 },                                  │
│  7 │     widgets: {                                                   │
│  8 │       main: {                                                    │
│  9 │         render({ pluginState, globalState }) {                   │
│ 10 │           const count = Number(pluginState?.count ?? 0);         │
│ 11 │           return ui.panel([                                      │
│ 12 │             ui.text("Count: " + count),                          │
│ 13 │             ui.row([                                             │
│ 14 │               ui.button("-", { onClick: { handler: "dec" } }),   │
│ 15 │               ui.button("+", { onClick: { handler: "inc" } }),   │
│ 16 │             ]),                                                  │
│ 17 │           ]);                                                    │
│ 18 │         },                                                       │
│ 19 │         handlers: {                                              │
│ 20 │           inc({ dispatchPluginAction, pluginState }) {           │
│ 21 │             const n = Number(pluginState?.count ?? 0) + 1;       │
│ 22 │             dispatchPluginAction("state/merge", { count: n });   │
│ 23 │           },                                                     │
│ 24 │           dec({ dispatchPluginAction, pluginState }) {           │
│ 25 │             const n = Number(pluginState?.count ?? 0) - 1;       │
│ 26 │             dispatchPluginAction("state/merge", { count: n });   │
│ 27 │           },                                                     │
│ 28 │         },                                                       │
│ 29 │       },                                                         │
│ 30 │     },                                                           │
│ 31 │   };                                                             │
│ 32 │ });                                                              │
│    │                                                                  │
│    │  ← monospace font (JetBrains Mono), line numbers, keyword        │
│    │    highlighting via a lightweight highlighter or CodeMirror       │
└───────────────────────────────────────────────────────────────────────┘
```

### 4.2 Live Preview Pane — Rendering Widget Output

```
┌─── Live Preview ──────────────────────────────────────────┐
│                                                            │
│  ╔════════════════════════════════════════════════════╗    │
│  ║ COUNTER [abc-1234]                         LOADED ║    │
│  ╟────────────────────────────────────────────────────╢    │
│  ║                                                    ║    │
│  ║  ┌───────────────────────────────────────────┐    ║    │
│  ║  │ Counter: 5                                │    ║    │
│  ║  │ ┌Shared total: 5┐ ┌Instances: 1──┐       │    ║    │
│  ║  │ [Decrement]  [Reset]  [Increment]         │    ║    │
│  ║  └───────────────────────────────────────────┘    ║    │
│  ║                                                    ║    │
│  ╚════════════════════════════════════════════════════╝    │
│                                                            │
│  ╔════════════════════════════════════════════════════╗    │
│  ║ GREETER [def-5678]                         LOADED ║    │
│  ╟────────────────────────────────────────────────────╢    │
│  ║                                                    ║    │
│  ║  ┌───────────────────────────────────────────┐    ║    │
│  ║  │ Hello, World!                             │    ║    │
│  ║  │ [World________________________]           │    ║    │
│  ║  └───────────────────────────────────────────┘    ║    │
│  ║                                                    ║    │
│  ╚════════════════════════════════════════════════════╝    │
│                                                            │
│  (scrollable if widgets overflow)                          │
│                                                            │
│  ┌─ empty state ────────────────────────────────────┐     │
│  │                                                   │     │
│  │   No plugins loaded.                             │     │
│  │                                                   │     │
│  │   Load a preset from the sidebar or write         │     │
│  │   custom plugin code and press ▶ RUN.            │     │
│  │                                                   │     │
│  │   Ctrl+Enter to run from editor.                  │     │
│  │                                                   │     │
│  └───────────────────────────────────────────────────┘     │
└────────────────────────────────────────────────────────────┘
```

### 4.3 Editor Tab Bar Interactions

| Action | Behavior |
|--------|----------|
| Click ▶ RUN | Load/reload the current tab's code as a plugin (new instance if first run, or dispose+reload if existing) |
| Click ⟳ | Re-render all widgets without reloading plugin code |
| Ctrl+Enter | Same as ▶ RUN (keyboard shortcut) |
| Click tab | Switch to that editor tab |
| Click × on tab | Close tab (confirm if unsaved changes) |
| Preset click in sidebar | Opens new tab with preset code, auto-runs |
| [+ New Plugin] | Opens new tab with template code |

### 4.4 Preview Pane — Instance Card

Each loaded plugin instance renders as a card in the preview pane. The card shows:

```
╔══════════════════════════════════════════════════════╗
║  ● PLUGIN TITLE [instance-id-short]         STATUS  ║
║    packageId: counter                                ║
╟──────────────────────────────────────────────────────╢
║                                                      ║
║  ┌── widget: counter ────────────────────────────┐  ║
║  │  (rendered UINode tree)                       │  ║
║  └───────────────────────────────────────────────┘  ║
║                                                      ║
║  ┌── widget: settings ───────────────────────────┐  ║
║  │  (rendered UINode tree)                       │  ║
║  └───────────────────────────────────────────────┘  ║
║                                                      ║
║  grants: R[counter-summary] W[counter-summary]       ║
╚══════════════════════════════════════════════════════╝
```

The status indicator colors:
- `●` green = LOADED, healthy
- `●` amber = RENDERING (in-flight render)
- `●` red = ERROR (render or load failure)

---

## 5. Component Breakdown: DevTools Panel

### 5.1 DevTools — Tab Overview

The bottom panel is a tabbed devtools area inspired by browser DevTools.

```
┌─── DevTools ──────────────────────────────────────────────────────────────────┐
│ [Timeline▾] [State] [Capabilities] [Errors] [Shared Domains] [📖Docs] [▲ ▼] │
├───────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   (content of selected tab)                                                  │
│                                                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

The [▲] button collapses devtools to just the tab bar (saves vertical space).
The [▼] button expands it. Drag-resizable handle on the top edge.

### 5.2 DevTools — Timeline Tab

The timeline shows every dispatch intent with full context, color-coded by outcome.

```
┌─── Timeline ────────────────────────────────────────────────────────────┐
│ Filters: scope:[all ▾] outcome:[all ▾] instance:[________] [🔍] [CLR] │
│          domain:[all ▾]  action:[________]                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│ TIME         SCOPE    OUTCOME   ACTION                INSTANCE  DOMAIN │
│ ─────────── ──────── ───────── ─────────────────────  ──────── ─────── │
│ 09:01:23.4  plugin   ✅applied  increment              abc-123  -      │
│ 09:01:23.4  shared   ✅applied  set-instance           abc-123  ctr-su │
│ 09:01:22.1  shared   🚫denied   set-name               def-456  grt-pr │
│                                  └─ reason: missing-write-grant:greeter │
│ 09:01:21.0  plugin   ✅applied  nameChanged            def-456  -      │
│ 09:01:20.5  plugin   ⚪ignored  unknown-action          ghi-789  -      │
│                                  └─ reason: no-local-reducer-match      │
│                                                                         │
│ (color key: green=applied, red=denied, gray=ignored)                   │
│ (click row to expand: shows full payload JSON)                          │
│                                                                         │
│ ─── expanded row ───                                                    │
│ 09:01:23.4  shared   ✅applied  set-instance           abc-123  ctr-su │
│   ┌─ payload ────────────────────────────────────────────────┐         │
│   │ { "value": 6 }                                           │         │
│   └──────────────────────────────────────────────────────────┘         │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.3 DevTools — State Tab

Shows per-instance plugin state as structured, syntax-highlighted JSON.

```
┌─── State ───────────────────────────────────────────────────────────────┐
│ Instance: [abc-1234 (Counter) ▾]                            [⟳ refresh]│
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  Plugin State (abc-1234):                                              │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │ {                                                                 │ │
│  │   "value": 5                                                      │ │
│  │ }                                                                 │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
│  Global State (projected for abc-1234):                                │
│  ┌───────────────────────────────────────────────────────────────────┐ │
│  │ {                                                                 │ │
│  │   "self": { "instanceId": "abc-1234", "packageId": "counter" },  │ │
│  │   "shared": {                                                     │ │
│  │     "counter-summary": {                                          │ │
│  │       "totalValue": 5,                                            │ │
│  │       "instanceCount": 1,                                         │ │
│  │       "lastUpdatedInstanceId": "abc-1234"                         │ │
│  │     }                                                             │ │
│  │   },                                                              │ │
│  │   "system": { "pluginCount": 3, "dispatchCount": 47 }            │ │
│  │ }                                                                 │ │
│  └───────────────────────────────────────────────────────────────────┘ │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.4 DevTools — Capabilities Tab

Shows capability grants per instance with visual grant/deny indicators.

```
┌─── Capabilities ────────────────────────────────────────────────────────┐
│                                                                         │
│  ┌─ abc-1234 (Counter) ────────────────────────────────────────────┐   │
│  │                                                                  │   │
│  │  Read Shared:                                                    │   │
│  │    ✅ counter-summary    ❌ greeter-profile                      │   │
│  │    ❌ runtime-registry   ❌ runtime-metrics                      │   │
│  │                                                                  │   │
│  │  Write Shared:                                                   │   │
│  │    ✅ counter-summary    ❌ greeter-profile                      │   │
│  │    ❌ runtime-registry   ❌ runtime-metrics                      │   │
│  │                                                                  │   │
│  │  System Commands: (none)                                         │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─ def-5678 (Greeter) ────────────────────────────────────────────┐   │
│  │  Read:  ✅ greeter-profile                                       │   │
│  │  Write: ✅ greeter-profile                                       │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
│  ┌─ ghi-9012 (Status Dashboard) ──────────────────────────────────┐   │
│  │  Read:  ✅ counter-summary ✅ runtime-metrics ✅ runtime-registry │   │
│  │  Write: (none — read-only dashboard)                             │   │
│  └──────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.5 DevTools — Errors Tab

Collects all errors (load failures, render errors, event handler errors) in a log stream.

```
┌─── Errors ──────────────────────────────────────────────────────────────┐
│ [🗑 Clear]                                                              │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  09:02:15.123  RENDER_ERROR   abc-1234 / counter                       │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ TypeError: Cannot read property 'value' of undefined             │  │
│  │   at render (abc-1234.plugin.js:12)                              │  │
│  │   at globalThis.__pluginHost.render (plugin-bootstrap.js:42)     │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  09:02:10.456  LOAD_ERROR     custom                                   │
│  ┌──────────────────────────────────────────────────────────────────┐  │
│  │ SyntaxError: Unexpected token '}'                                │  │
│  │   at custom.plugin.js:7                                          │  │
│  └──────────────────────────────────────────────────────────────────┘  │
│                                                                         │
│  (empty state: "No errors — all systems operational ✅")                │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.6 DevTools — Shared Domains Tab

Shows each shared domain's current state with per-domain cards and last-update info.

```
┌─── Shared Domains ──────────────────────────────────────────────────────┐
│                                                                         │
│  ┌─ counter-summary ──────────────────────────────────────┐            │
│  │  totalValue:            5                              │            │
│  │  instanceCount:         1                              │            │
│  │  lastUpdatedInstanceId: abc-1234                       │            │
│  │  valuesByInstance:                                     │            │
│  │    abc-1234: 5                                         │            │
│  │                                                        │            │
│  │  Writers: abc-1234 (Counter)                           │            │
│  │  Readers: abc-1234 (Counter), ghi-9012 (Status Dash)  │            │
│  └────────────────────────────────────────────────────────┘            │
│                                                                         │
│  ┌─ greeter-profile ──────────────────────────────────────┐            │
│  │  name:                  "World"                         │            │
│  │  lastUpdatedInstanceId: def-5678                       │            │
│  │                                                        │            │
│  │  Writers: def-5678 (Greeter)                           │            │
│  │  Readers: def-5678 (Greeter), jkl-3456 (Greeter Sh.)  │            │
│  └────────────────────────────────────────────────────────┘            │
│                                                                         │
│  ┌─ runtime-registry ─────────────────────────────────────┐            │
│  │  (read-only projection — 3 plugins registered)         │            │
│  │  Readers: ghi-9012 (Status Dash), mno-7890 (Rt Mon)   │            │
│  └────────────────────────────────────────────────────────┘            │
│                                                                         │
│  ┌─ runtime-metrics ──────────────────────────────────────┐            │
│  │  pluginCount:    3                                     │            │
│  │  dispatchCount:  47                                    │            │
│  │  lastScope:      shared                                │            │
│  │  lastOutcome:    applied                               │            │
│  │  Readers: ghi-9012 (Status Dash)                       │            │
│  └────────────────────────────────────────────────────────┘            │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### 5.7 DevTools — Docs Tab

The Docs tab surfaces the project's own markdown documentation inside the workbench,
rendered inline with syntax-highlighted code blocks. Every doc (and every code block
within a doc) has a one-click **copy-to-clipboard** button so content can be pasted
straight into an LLM chat window.

#### 5.7.1 Layout — Tree + Rendered Doc

```
┌─── Docs ────────────────────────────────────────────────────────────────────┐
│                                                                              │
│ ┌─ nav (220px) ──────┬─ rendered doc ──────────────────────────────────────┐│
│ │                     │                                                    ││
│ │ 📖 DOCS             │  # Plugin Authoring Quickstart           [📋 Copy] ││
│ │                     │                                                    ││
│ │ ▸ Overview          │  This quickstart shows how to write a              ││
│ │ ▾ Plugin Authoring  │  plugin that runs in the WebVM playground          ││
│ │   ● Quickstart  ◀  │  runtime.                                          ││
│ │ ▸ Architecture      │                                                    ││
│ │ ▸ Runtime           │  ## 1) Write a plugin with `definePlugin`         ││
│ │ ▸ Migration         │                                                    ││
│ │                     │  The runtime expects plugin code to call           ││
│ │                     │  `definePlugin((host) => ({ ... }))`.             ││
│ │                     │                                                    ││
│ │ ─────────────────   │  ```js                                   [📋]     ││
│ │ [📋 Copy All Docs]  │  definePlugin(({ ui }) => {                       ││
│ │                     │    return {                                        ││
│ │                     │      id: "hello-counter",                         ││
│ │                     │      title: "Hello Counter",                      ││
│ │                     │      ...                                          ││
│ │                     │    };                                              ││
│ │                     │  });                                               ││
│ │                     │  ```                                               ││
│ │                     │                                                    ││
│ │                     │  ## 2) Understand handler context                  ││
│ │                     │                                                    ││
│ │                     │  Each widget handler receives:                     ││
│ │                     │  - `pluginState`: local state for this ...         ││
│ │                     │  - `globalState`: projected shared/system ...      ││
│ │                     │  - `dispatchPluginAction(actionType, ...)`        ││
│ │                     │  - `dispatchSharedAction(domain, ...)`            ││
│ │                     │                                                    ││
│ └─────────────────────┴────────────────────────────────────────────────────┘│
└──────────────────────────────────────────────────────────────────────────────┘
```

#### 5.7.2 Navigation Tree

The left nav mirrors the `docs/` directory structure:

```
📖 DOCS
├── Overview               ← docs/README.md
├── Plugin Authoring
│   └── Quickstart         ← docs/plugin-authoring/quickstart.md
├── Architecture
│   └── Capability Model   ← docs/architecture/capability-model.md
├── Runtime
│   └── Embedding Guide    ← docs/runtime/embedding.md
└── Migration
    └── VM API Changelog   ← docs/migration/changelog-vm-api.md
```

The tree is generated at build time from the bundled doc manifest (see §5.7.5).
Active doc is highlighted with `◀` marker and accent color.

#### 5.7.3 Copy-to-Clipboard Behaviors

Three distinct copy targets, each with its own button:

| Button | Location | What it copies |
|--------|----------|----------------|
| **[📋 Copy]** on doc heading | Top-right of rendered doc pane | Full raw markdown of the current doc (not HTML — the original `.md` source) |
| **[📋]** on code fence | Top-right of each ` ```code``` ` block | Just that code block's content (no fences, no language tag) |
| **[📋 Copy All Docs]** | Bottom of nav tree | Concatenation of ALL docs as raw markdown, separated by `---` and `# filename` headers |

**Why raw markdown?** LLMs consume markdown far better than rendered HTML.
Copying raw source means the user can paste directly into Claude/ChatGPT/etc.
and the model sees the original formatting, code blocks, and headings intact.

Visual feedback after copy: the button text briefly changes to `✅ Copied!`
(1.5s, then reverts). Uses the `navigator.clipboard.writeText()` API.

#### 5.7.4 "Copy All Docs" Output Format

When the user clicks **[📋 Copy All Docs]**, the clipboard receives a single
string assembled from every bundled doc, formatted for LLM consumption:

```markdown
# Plugin Playground Documentation

---

# docs/README.md

(full raw markdown of README.md)

---

# docs/plugin-authoring/quickstart.md

(full raw markdown of quickstart.md)

---

# docs/architecture/capability-model.md

(full raw markdown of capability-model.md)

---

# docs/runtime/embedding.md

(full raw markdown of embedding.md)

---

# docs/migration/changelog-vm-api.md

(full raw markdown of changelog-vm-api.md)
```

This format gives the LLM clear file provenance per section and clean separators.

#### 5.7.5 Build-Time Bundling Strategy

The docs live in `docs/` as plain markdown files. They need to be available to the
browser at runtime as both **raw source** (for copy-to-clipboard) and **rendered
HTML** (for display). The bundling approach uses Vite's built-in raw-import
capability with zero extra dependencies beyond a lightweight markdown renderer:

**Step 1 — Vite `?raw` imports at build time**

Create a doc manifest module that imports each markdown file as a raw string:

```ts
// client/src/lib/docsManifest.ts

// Vite ?raw suffix imports the file content as a string at build time.
// No runtime file-system access needed — docs are embedded in the JS bundle.

import readmeRaw from "../../../docs/README.md?raw";
import quickstartRaw from "../../../docs/plugin-authoring/quickstart.md?raw";
import capabilityModelRaw from "../../../docs/architecture/capability-model.md?raw";
import embeddingRaw from "../../../docs/runtime/embedding.md?raw";
import changelogRaw from "../../../docs/migration/changelog-vm-api.md?raw";

export interface DocEntry {
  /** Display title in nav tree */
  title: string;
  /** Category / parent folder for nav grouping */
  category: string;
  /** Relative path from docs/ root (for display and "Copy All" headers) */
  path: string;
  /** Raw markdown source (for copy-to-clipboard) */
  raw: string;
}

export const docs: DocEntry[] = [
  {
    title: "Overview",
    category: "Overview",
    path: "docs/README.md",
    raw: readmeRaw,
  },
  {
    title: "Quickstart",
    category: "Plugin Authoring",
    path: "docs/plugin-authoring/quickstart.md",
    raw: quickstartRaw,
  },
  {
    title: "Capability Model",
    category: "Architecture",
    path: "docs/architecture/capability-model.md",
    raw: capabilityModelRaw,
  },
  {
    title: "Embedding Guide",
    category: "Runtime",
    path: "docs/runtime/embedding.md",
    raw: embeddingRaw,
  },
  {
    title: "VM API Changelog",
    category: "Migration",
    path: "docs/migration/changelog-vm-api.md",
    raw: changelogRaw,
  },
];

/**
 * Build the concatenated "all docs" string for the Copy All button.
 */
export function buildAllDocsMarkdown(): string {
  const parts = docs.map((d) => `# ${d.path}\n\n${d.raw}`);
  return `# Plugin Playground Documentation\n\n---\n\n${parts.join("\n\n---\n\n")}`;
}
```

**Step 2 — Markdown rendering in the browser**

Use a lightweight markdown-to-HTML library to render the raw strings.
Recommended: **`marked`** (~40KB gzipped, zero config) or **`markdown-it`** (~35KB).
Both support code-fence extraction which we need for per-block copy buttons.

```ts
// client/src/lib/renderMarkdown.ts

import { marked } from "marked";   // or markdown-it

// Configure for code highlighting + extracting code blocks
const renderer = new marked.Renderer();

// Override code block rendering to wrap each in a container with a copy button target
renderer.code = function ({ text, lang }) {
  const escaped = text.replace(/</g, "&lt;").replace(/>/g, "&gt;");
  // data-raw attribute holds the raw code for clipboard copy
  return `<div class="doc-code-block" data-raw="${encodeURIComponent(text)}">
    <div class="doc-code-header">
      <span class="doc-code-lang">${lang ?? ""}</span>
      <button class="doc-copy-code" title="Copy code block">📋</button>
    </div>
    <pre><code class="language-${lang ?? "text"}">${escaped}</code></pre>
  </div>`;
};

export function renderDoc(raw: string): string {
  return marked(raw, { renderer });
}
```

**Step 3 — Vite config: allow raw imports from docs/**

The existing `vite.config.ts` already has `fs.strict: true`. Since `docs/` is a
sibling of `client/` (both under the plugin-playground root), the `?raw` imports
resolve within the project. No config changes needed — Vite's `?raw` works with
any file path the bundler can resolve at build time.

If a new alias is desired for readability:

```ts
// addition to vite.config.ts resolve.alias
"@docs": path.resolve(import.meta.dirname, "docs"),
```

Then imports become:
```ts
import readmeRaw from "@docs/README.md?raw";
```

**Step 4 — Bundle size impact**

Current docs total: ~12.5KB raw markdown (5 files, 371 lines).
After gzip: ~4KB addition to the JS bundle. Negligible.
Even if docs grow 10×, raw-import bundling stays practical up to ~100KB.

#### 5.7.6 DocsPanel React Component

```
DocsPanel
├── DocsNav (220px sidebar)
│   ├── DocsCategoryGroup[]
│   │   └── DocsNavItem[]
│   └── CopyAllDocsButton
└── DocsContent (flex-1)
    ├── DocsContentHeader (title + copy-doc button)
    └── DocsRenderedMarkdown (dangerouslySetInnerHTML with sanitized marked output)
        └── DocCodeBlock[] (event delegation for per-block copy buttons)
```

State:
- `selectedDocPath: string` — which doc is active (defaults to first)
- No server calls — everything is pre-bundled

#### 5.7.7 Interaction Table

| Action | Behavior |
|--------|----------|
| Click nav item | Load that doc's raw markdown, render to HTML, display in content pane |
| Click [📋 Copy] on doc header | `navigator.clipboard.writeText(currentDoc.raw)` → toast "Copied quickstart.md" |
| Click [📋] on code fence | Extract raw code from `data-raw` attribute → clipboard → toast "Code copied" |
| Click [📋 Copy All Docs] | `navigator.clipboard.writeText(buildAllDocsMarkdown())` → toast "All docs copied (12.5KB)" |
| Keyboard: Ctrl+Shift+D | Switch to Docs tab (new shortcut) |

#### 5.7.8 Styling

The rendered markdown uses the existing theme tokens and monospace typography:

```
Rendered doc styling:
  - h1:  Space Mono Bold 18px, accent color, uppercase
  - h2:  Space Mono Bold 15px, foreground
  - h3:  Space Mono Bold 13px, muted-foreground
  - p:   Space Mono 13px, foreground, line-height 1.6
  - code (inline): JetBrains Mono 12px, background surface, accent border
  - code (block):  JetBrains Mono 12px, background oklch(0.12), border accent/20
  - links: accent color, underline on hover
  - lists: accent bullet markers
  - tables: border accent/20, header row accent/10 background
  - hr:  border accent/30
  - blockquote: left border 3px accent/50, padding-left 1rem, muted-foreground
```

Copy buttons use the existing Button component (variant="ghost", size="sm") with
the clipboard icon from lucide-react (already a dependency).

---

## 6. Top Toolbar

```
┌─────────────────────────────────────────────────────────────────────────┐
│ ▣ PLUGIN WORKBENCH      [■ 3 plugins] [↻ 47 dispatches] [⚡ OK]  [☰] │
│                          ^^^^^^^^^^    ^^^^^^^^^^^^^^^^   ^^^^^^^  ^^^  │
│                          badge/count   running total     health   menu  │
└─────────────────────────────────────────────────────────────────────────┘
```

| Element | Behavior |
|---------|----------|
| ▣ PLUGIN WORKBENCH | App title, links to / |
| [■ N plugins] | Click → focus sidebar RUNNING section |
| [↻ N dispatches] | Click → open devtools Timeline tab |
| [⚡ OK/ERR] | Runtime health indicator (calls `sandbox.health()`) |
| [☰] | Menu: Reset All, Export State, Import State, Toggle Theme, About |

---

## 7. Interaction Flows

### 7.1 Flow: Load Preset → Edit → Re-run

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User clicks "Counter" in sidebar CATALOG                 │
│    → New editor tab "counter.js" opens with preset code     │
│    → Auto-runs: loadPlugin() → pluginRegistered()           │
│    → Instance appears in RUNNING section with ● green       │
│    → Widget renders in Live Preview pane                     │
│                                                              │
│ 2. User modifies code in editor (e.g., changes label text)  │
│    → Tab shows ● unsaved indicator                          │
│    → Preview still shows old version                         │
│                                                              │
│ 3. User presses Ctrl+Enter or clicks ▶ RUN                  │
│    → Old instance disposed (disposePlugin + pluginRemoved)  │
│    → New instance loaded with modified code                  │
│    → Preview updates immediately                             │
│    → Timeline shows dispose + register events                │
│                                                              │
│ 4. User interacts with widget (clicks [+] on counter)        │
│    → Event dispatched through runtime                        │
│    → Intent → reducer → state update → re-render            │
│    → Timeline entry appears in devtools (if open)           │
│    → State tab updates (if focused on this instance)         │
└─────────────────────────────────────────────────────────────┘
```

### 7.2 Flow: Debug Denied Shared Write

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User writes custom plugin with dispatchSharedAction()     │
│    → Custom plugins get empty grants (deny-by-default)      │
│                                                              │
│ 2. User runs plugin and triggers shared write                │
│    → Widget renders normally (local state may update)        │
│    → Shared state does NOT change                            │
│                                                              │
│ 3. User notices shared state didn't change                   │
│    → Opens devtools → Timeline tab                           │
│    → Sees entry: shared  🚫denied  set-instance   xyz-...   │
│    → Expands row → reason: "missing-write-grant:counter-su…" │
│                                                              │
│ 4. User opens Capabilities tab                               │
│    → Sees instance xyz-... has Write: (none)                 │
│    → Understands the issue: custom plugins need explicit     │
│      capability grants                                       │
│                                                              │
│ 5. Future: Capabilities tab could show [Grant] buttons       │
│    to dynamically add grants for debugging                   │
└─────────────────────────────────────────────────────────────┘
```

### 7.3 Flow: Multi-Instance Inspection

```
┌─────────────────────────────────────────────────────────────┐
│ 1. User loads Counter preset 3 times                         │
│    → 3 instances in RUNNING: abc-1, abc-2, abc-3            │
│    → 3 widget cards in Preview, each with own counter value │
│                                                              │
│ 2. User clicks instance "abc-2" in sidebar                   │
│    → Editor shows counter code                              │
│    → Preview scrolls to / highlights abc-2's card           │
│    → Devtools State tab auto-selects abc-2                  │
│    → Devtools Timeline auto-filters to abc-2                │
│                                                              │
│ 3. User increments abc-2's counter                           │
│    → abc-2's card updates: Counter: 1                       │
│    → Shared Domains tab: counter-summary.totalValue         │
│      now includes abc-2's contribution                       │
│    → Status Dashboard widget (if loaded) updates too        │
└─────────────────────────────────────────────────────────────┘
```

---

## 8. Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Enter` | Run current editor tab (load/reload plugin) |
| `Ctrl+Shift+N` | New plugin tab |
| `Ctrl+W` | Close current tab |
| `Ctrl+1..9` | Switch to editor tab N |
| `Ctrl+\`` | Toggle devtools panel |
| `Ctrl+Shift+T` | Focus devtools Timeline tab |
| `Ctrl+Shift+S` | Focus devtools State tab |
| `Ctrl+B` | Toggle sidebar |
| `Ctrl+Shift+D` | Focus devtools Docs tab |
| `Escape` | Dismiss any modal/dropdown |

---

## 9. Empty / Zero States

### 9.1 No Plugins Loaded

```
┌─── Full App — Empty State ──────────────────────────────────────────────┐
│ SIDEBAR       │   EDITOR + PREVIEW                                      │
│               │                                                         │
│ 📦 CATALOG    │   ┌──────────────────────────────────────────────────┐  │
│ (6 presets)   │   │                                                  │  │
│               │   │       Welcome to the Plugin Workbench            │  │
│ 🔌 RUNNING    │   │                                                  │  │
│ (empty)       │   │  Get started:                                    │  │
│               │   │                                                  │  │
│               │   │  1. Click a preset in the sidebar to load it     │  │
│               │   │  2. Or click [+ New Plugin] to write your own    │  │
│               │   │                                                  │  │
│               │   │  Keyboard shortcuts:                             │  │
│               │   │    Ctrl+Enter   Run plugin                       │  │
│               │   │    Ctrl+B       Toggle sidebar                   │  │
│               │   │    Ctrl+`       Toggle devtools                  │  │
│               │   │                                                  │  │
│               │   │  Plugin API:                                     │  │
│               │   │    definePlugin(({ ui }) => ({                   │  │
│               │   │      id, title, initialState,                    │  │
│               │   │      widgets: { name: { render, handlers } }     │  │
│               │   │    }))                                           │  │
│               │   │                                                  │  │
│               │   └──────────────────────────────────────────────────┘  │
│               │                                                         │
│               │   DEVTOOLS: (collapsed — no data yet)                  │
└───────────────┴─────────────────────────────────────────────────────────┘
```

### 9.2 Plugin Load Error

```
┌─── Editor Tab Bar ─────────────────────────────────────────────┐
│ [broken.js ⚠] [counter.js ●]                         [▶ RUN]  │
├────────────────────────────────────────────────────────────────┘
│  CODE EDITOR                │  LIVE PREVIEW                     │
│                              │                                   │
│  1│ definePlugin(({ ui })   │  ╔════════════════════════════╗   │
│  2│   return {              │  ║  ⚠ LOAD ERROR              ║   │
│  3│     id: "broken"        │  ║                             ║   │
│  4│   }                     │  ║  SyntaxError: Unexpected    ║   │
│  5│ })                      │  ║  token '}' at line 4        ║   │
│  6│ // missing comma!       │  ║                             ║   │
│                              │  ║  [Dismiss] [Copy Error]    ║   │
│                              │  ╚════════════════════════════╝   │
│                              │                                   │
│                              │  (other loaded plugins still      │
│                              │   render below the error card)    │
└──────────────────────────────┴───────────────────────────────────┘
```

---

## 10. Component Architecture (React)

```
App
├── ThemeProvider
├── Redux Provider (store)
├── TooltipProvider
├── Toaster
└── WorkbenchLayout                    ← new top-level layout component
    ├── TopToolbar                     ← status badges, menu
    ├── Sidebar                        ← replaces CatalogShell
    │   ├── SidebarSection: Catalog
    │   │   └── CatalogItem[]         ← preset entries with capability badges
    │   ├── SidebarSection: Running
    │   │   └── InstanceItem[]         ← loaded instances with grants + unload
    │   └── NewPluginButton
    ├── MainPane                       ← replaces WorkspaceShell
    │   ├── EditorTabBar
    │   │   ├── EditorTab[]
    │   │   └── RunButton + ReloadButton
    │   ├── SplitView (horizontal)
    │   │   ├── CodeEditor             ← syntax-highlighted editor
    │   │   └── LivePreview            ← widget rendering
    │   │       └── InstanceCard[]
    │   │           └── WidgetRenderer ← moved from InspectorShell
    │   └── DevToolsPanel              ← replaces InspectorShell
    │       ├── DevToolsTabs
    │       ├── TimelinePanel
    │       ├── StatePanel
    │       ├── CapabilitiesPanel
    │       ├── ErrorsPanel
    │       ├── SharedDomainsPanel
    │       └── DocsPanel              ← NEW: embedded docs viewer
    │           ├── DocsNav            ← tree nav from docsManifest
    │           ├── DocsContent        ← rendered markdown
    │           └── CopyAllDocsButton  ← concat all docs → clipboard
    └── (modals/overlays)
        ├── ConfirmUnloadDialog
        └── MenuDropdown
```

### 10.1 New vs Existing Components

| Component | Status | Notes |
|-----------|--------|-------|
| `WorkbenchLayout` | **NEW** | Replaces `Playground.tsx` as top-level orchestrator |
| `TopToolbar` | **NEW** | Runtime badges + menu |
| `Sidebar` | **REWRITE** of `CatalogShell` | Tree nav + capability badges + instance focus |
| `CodeEditor` | **NEW** | Replace textarea; CodeMirror 6 or similar |
| `EditorTabBar` | **NEW** | Multi-tab management |
| `LivePreview` | **NEW** | Dedicated preview pane with instance cards |
| `InstanceCard` | **NEW** | Per-instance header + widget rendering |
| `WidgetRenderer` | **KEEP** | Existing component, moved to LivePreview |
| `DevToolsPanel` | **REWRITE** of `InspectorShell` | Tabbed bottom panel, more tabs |
| `TimelinePanel` | **EXTRACT** | From InspectorShell timeline tab |
| `StatePanel` | **NEW** | Per-instance state viewer |
| `CapabilitiesPanel` | **NEW** | Grant visualization |
| `ErrorsPanel` | **NEW** | Error log stream |
| `SharedDomainsPanel` | **EXTRACT+ENHANCE** | From InspectorShell shared tab |
| `DocsPanel` | **NEW** | Embedded docs viewer with markdown rendering + copy-to-clipboard |
| `DocsNav` | **NEW** | Tree navigation built from docsManifest |
| `DocsContent` | **NEW** | Rendered markdown with per-block copy buttons |

---

## 11. Data Flow Changes

### 11.1 Current Flow (Problem)

```
Playground.tsx
├── owns ALL state coordination
├── owns custom code string
├── owns widget trees + errors
├── owns plugin meta map
├── calls sandbox client directly
└── re-renders ALL widgets on ANY state change
```

### 11.2 Proposed Flow (Solution)

```
WorkbenchLayout.tsx
├── provides layout structure only
├── delegates to feature-specific hooks/controllers
│
├── usePluginOrchestrator()          ← hook: load/unload/reload
│   ├── calls sandboxClient
│   ├── dispatches to Redux
│   └── manages editor tab → instance mapping
│
├── useWidgetRenderer(instanceId)    ← hook: per-instance rendering
│   ├── selects instance state from Redux
│   ├── computes globalState projection
│   └── calls sandboxClient.render() only when deps change
│
├── useEditorTabs()                  ← hook: tab state management
│   ├── tracks open tabs, active tab, dirty state
│   └── maps tab → packageId + code content
│
└── useDevTools()                    ← hook: devtools state
    ├── timeline filtering
    ├── selected instance for state viewer
    └── error accumulation
```

### 11.3 Key Improvement: Selective Re-rendering

Current: ANY Redux state change → re-render ALL widgets (O(n*w) sandbox calls).

Proposed: Track which instances are affected by each dispatch outcome:
- Plugin-scoped dispatch → only re-render that instance's widgets
- Shared-scoped dispatch → re-render instances that READ that domain
- Use React.memo + stable selector references to prevent cascade

---

## 12. Design Decisions

### D1: IDE Layout over Equal Columns

**Decision:** Use sidebar + dominant editor + bottom devtools instead of 3 equal columns.

**Rationale:** The primary developer activity is writing and testing plugin code. The editor deserves 60%+ of screen width. Inspection is a secondary activity that benefits from wide horizontal space (timeline table columns).

**Alternative considered:** Keep 3 columns but make center column wider. Rejected because it still separates preview from editor (right panel) and pushes timeline below the fold.

### D2: Code Editor over Textarea

**Decision:** Use a real code editor component (CodeMirror 6 or Monaco) instead of `<textarea>`.

**Rationale:** Syntax highlighting, line numbers, bracket matching, and auto-indent are essential for writing JavaScript plugins. Without them, the edit experience is painful.

**Trade-off:** Adds ~100-200KB bundle weight. Worth it for a developer tool.

**Recommendation:** CodeMirror 6 — lighter than Monaco, good JS support, theming via CSS.

### D3: Live Preview Next to Editor

**Decision:** Widget preview lives in the right half of the editor pane, not in a separate panel.

**Rationale:** The edit→preview loop is the tightest feedback cycle in the app. Having code on the left and preview on the right (same horizontal eye-line) minimizes context-switching.

### D4: DevTools as Bottom Panel

**Decision:** Inspection tools (timeline, state, capabilities, errors, shared domains) live in a collapsible bottom panel.

**Rationale:** Follows established IDE convention (VS Code terminal panel, browser DevTools). Developers already have muscle memory for this layout. Wide horizontal space benefits table-like data (timeline, capabilities).

### D5: Capability Badges in Sidebar

**Decision:** Show capability summary (R/W) on each catalog preset and each running instance in the sidebar.

**Rationale:** Capabilities are the most common source of confusion (why didn't my shared write work?). Making them visible at a glance prevents the need to dig into code or the Capabilities devtools tab.

### D6: Instance Focus

**Decision:** Clicking a running instance in the sidebar focuses the entire UI on that instance (editor tab, preview highlight, state viewer, timeline filter).

**Rationale:** When debugging a specific plugin instance among many, you want all panels to align on the same instance. This eliminates manual filter-setting in each panel.

### D7: Embedded Docs with Copy-to-Clipboard (LLM Workflow)

**Decision:** Bundle all project markdown docs into the app at build time via Vite `?raw` imports, render them in a DevTools tab, and provide per-doc / per-code-block / all-docs copy-to-clipboard buttons that copy **raw markdown** (not HTML).

**Rationale:** Developers using this tool frequently need to share context with LLMs (Claude, ChatGPT). The most common workflow is: read docs → copy relevant section → paste into LLM prompt. Raw markdown is the ideal format because LLMs parse it natively. Having docs embedded in the app (vs. linking to external files) means no context-switching to a file browser or docs site. The "Copy All Docs" button enables a single-click "give the LLM all the context" workflow.

**Alternative considered:** Link to docs on GitHub or serve them from a separate docs site. Rejected because it breaks the single-pane-of-glass developer experience and adds a network dependency. Also considered serving docs from the Express server via an API endpoint — rejected because `?raw` imports are simpler, produce zero runtime overhead, and the total docs size (~12.5KB) is negligible in the bundle.

**Trade-off:** Docs must be rebuilt when markdown files change (a `vite dev` HMR restart). Acceptable for docs that change infrequently.

---

## 13. Theme and Visual Design

The existing "Technical Brutalism" theme (dark terminal aesthetic with electric cyan accents) is **kept and enhanced**:

```
┌─── Color Tokens ──────────────────────────────────────────┐
│                                                            │
│  Background:     oklch(0.15 0.01 240)  ← deep charcoal   │
│  Foreground:     oklch(0.95 0 0)       ← stark white      │
│  Accent:         oklch(0.75 0.15 195)  ← electric cyan    │
│  Warning:        oklch(0.70 0.18 75)   ← amber            │
│  Error:          oklch(0.65 0.22 25)   ← danger red       │
│  Success:        oklch(0.70 0.18 145)  ← muted green      │
│  Muted:          oklch(0.55 0.02 240)  ← dim gray         │
│  Surface:        oklch(0.18 0.01 240)  ← card background  │
│  Border:         oklch(0.30 0.02 240)  ← subtle edge      │
│  Glow:           rgba(0,255,255,0.15)  ← active element   │
│                                                            │
│  Typography:                                               │
│    UI labels:    Space Mono 14px                           │
│    Code:         JetBrains Mono 13px                       │
│    Headings:     Space Mono Bold 16px uppercase            │
│    Badges:       Space Mono 11px uppercase tracking-wider  │
│                                                            │
│  Corners:        2px (sharp, brutalist)                    │
│  Borders:        1px solid with accent/30 opacity          │
│  Active glow:    box-shadow 0 0 12px accent/40             │
│  Transitions:    150ms ease-out (no bounce/elastic)        │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

---

## 14. Package Boundary: What's Reusable vs App-Specific

```
┌─── packages/plugin-runtime/ (REUSABLE) ────────────────────┐
│                                                              │
│  Core:                                                       │
│  ├── contracts.ts          types, request/response shapes   │
│  ├── uiTypes.ts            UINode kind vocabulary           │
│  ├── uiSchema.ts           UINode validation                │
│  ├── dispatchIntent.ts     intent validation                │
│  ├── runtimeService.ts     QuickJS engine wrapper           │
│  └── runtimeIdentity.ts    instance ID generation           │
│                                                              │
│  Worker transport:                                           │
│  ├── worker/runtime.worker.ts    web worker entry point     │
│  └── worker/sandboxClient.ts     RPC client for worker      │
│                                                              │
│  Redux adapter:                                              │
│  └── redux-adapter/store.ts      state + reducers + policy  │
│                                                              │
│  Host adapter interface:                                     │
│  └── hostAdapter.ts              backend-agnostic interface │
│                                                              │
│  ⬆ NO React. NO UI components. NO theme. NO layout.        │
│  ⬆ Consumable by any host: React app, Node.js test, CLI.   │
└──────────────────────────────────────────────────────────────┘

┌─── client/ (APP-SPECIFIC — the playground UI) ─────────────┐
│                                                              │
│  Layout:                                                     │
│  ├── WorkbenchLayout.tsx                                    │
│  ├── TopToolbar.tsx                                         │
│  └── theme / CSS                                            │
│                                                              │
│  Features:                                                   │
│  ├── sidebar/  (Catalog, Running instances)                 │
│  ├── editor/   (Tabs, CodeEditor, RunControls)              │
│  ├── preview/  (LivePreview, InstanceCard, WidgetRenderer)  │
│  └── devtools/ (Timeline, State, Capabilities, Errors,      │
│                 SharedDomains, Docs)                         │
│                                                              │
│  Hooks:                                                      │
│  ├── usePluginOrchestrator.ts                               │
│  ├── useWidgetRenderer.ts                                   │
│  ├── useEditorTabs.ts                                       │
│  └── useDevTools.ts                                         │
│                                                              │
│  Lib:                                                        │
│  ├── presetPlugins.ts                                       │
│  ├── docsManifest.ts  ← ?raw imports of docs/*.md           │
│  └── renderMarkdown.ts ← marked renderer + code-block copy │
│                                                              │
│  Assets (bundled at build time via Vite ?raw):               │
│  └── docs/*.md  → embedded as strings in docsManifest.ts    │
│                                                              │
│  ⬆ React + Tailwind + theme. Imports @runtime/* only.       │
│  ⬆ Could be replaced by a different UI without touching     │
│    the runtime package.                                      │
└──────────────────────────────────────────────────────────────┘
```

---

## 15. Implementation Plan

### Phase 1: Layout Shell (no new functionality)

1. Create `WorkbenchLayout.tsx` with sidebar + main pane + devtools panel structure
2. Create `TopToolbar.tsx` with static badges
3. Port existing `CatalogShell` content into `Sidebar` component
4. Port existing `WorkspaceShell` textarea into editor area (still textarea for now)
5. Port existing `InspectorShell` widget tab into `LivePreview` pane
6. Port existing `InspectorShell` timeline/shared tabs into `DevToolsPanel`
7. Wire routing: Home → WorkbenchLayout instead of Playground

### Phase 2: Editor Enhancement

1. Replace textarea with CodeMirror 6 (JS mode, dark theme)
2. Add editor tab management (useEditorTabs hook)
3. Add Ctrl+Enter shortcut for run
4. Add preset code loading into editor tabs
5. Add dirty-state indicator on tabs

### Phase 3: Sidebar Enhancement

1. Add capability badges to catalog entries
2. Add capability display to running instances
3. Add instance-focus behavior (click instance → focus all panels)
4. Add sidebar collapse/expand toggle

### Phase 4: DevTools Enhancement

1. Add State tab with per-instance state viewer
2. Add Capabilities tab with grant/deny grid
3. Add Errors tab with error log stream
4. Enhance Shared Domains tab with reader/writer attribution
5. Enhance Timeline with row expansion (payload JSON)
6. Add domain filter to Timeline

### Phase 4b: Docs Panel

1. Add `marked` (or `markdown-it`) dependency for markdown rendering
2. Create `client/src/lib/docsManifest.ts` with Vite `?raw` imports of all `docs/*.md` files
3. Create `client/src/lib/renderMarkdown.ts` with custom renderer for code-block copy buttons
4. Build `DocsNav` component (tree navigation from manifest categories)
5. Build `DocsContent` component (rendered markdown with per-block copy buttons)
6. Build `DocsPanel` container wiring nav + content + "Copy All Docs" button
7. Wire into DevToolsPanel as new tab
8. Add Ctrl+Shift+D keyboard shortcut
9. Style rendered markdown with existing theme tokens (see §5.7.8)

### Phase 5: Orchestration Refactor

1. Extract `usePluginOrchestrator` hook from Playground logic
2. Extract `useWidgetRenderer` with per-instance selective rendering
3. Wire devtools data sources through hooks
4. Remove monolithic Playground.tsx

---

## 16. Open Questions

1. **Code editor choice:** CodeMirror 6 vs Monaco Editor? CM6 is lighter (~150KB vs ~2MB) but Monaco has richer IntelliSense. For a plugin playground, CM6 is likely sufficient.

2. **Capability grant UI:** Should custom plugins have a UI to request/grant capabilities dynamically? This would be useful for debugging but adds complexity. Recommendation: Phase 5+ feature.

3. **Plugin persistence:** Should editor tab content persist across page reloads (localStorage)? Recommendation: Yes, simple localStorage serialization of tab state.

4. **Preset code editability:** When loading a preset, should the code be editable? Recommendation: Yes — load into editor tab as a copy. Original preset code is never modified.

5. **Mobile layout:** How much effort for responsive design? Recommendation: Minimal — this is a developer tool, primarily used on desktop. Stack sidebar + editor + devtools vertically on mobile, but don't optimize heavily.

6. **Docs rendering library:** `marked` vs `markdown-it` vs `remark/rehype`? `marked` is simplest (~40KB) and sufficient for rendering + code-block extraction. `markdown-it` offers more plugin extensibility. `remark/rehype` is the React ecosystem standard but heavier. Recommendation: `marked` for simplicity, upgrade if plugin needs emerge.

7. **Docs auto-discovery vs explicit manifest:** Should the doc manifest be auto-generated (e.g., via a Vite plugin that globs `docs/**/*.md`) or hand-maintained? Auto-discovery is more maintainable but adds build complexity. Recommendation: Start with explicit manifest (5 files), add auto-discovery if docs grow beyond ~15 files.

---

## References

- Prior design doc: `design-doc/01-deep-pass-ui-overhaul-runtime-packaging-and-docs-plan.md`
- Codebase audit: `design-doc/02-deep-pass-refresh-current-codebase-audit-and-ui-runtime-docs-roadmap.md`
- Capability model: `docs/architecture/capability-model.md`
- Plugin authoring quickstart: `docs/plugin-authoring/quickstart.md`
- Runtime embedding guide: `docs/runtime/embedding.md`
