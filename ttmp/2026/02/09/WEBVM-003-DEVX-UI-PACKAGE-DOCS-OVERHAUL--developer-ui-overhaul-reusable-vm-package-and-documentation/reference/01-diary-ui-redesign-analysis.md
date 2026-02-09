---
Title: Diary - UI Redesign Analysis
Ticket: WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/Playground.tsx
      Note: "Primary UI orchestration component - target of redesign analysis"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/CatalogShell.tsx
      Note: "Left panel - plugin catalog and loaded instances list"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/WorkspaceShell.tsx
      Note: "Center panel - code editor textarea and load button"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/features/workbench/InspectorShell.tsx
      Note: "Right panel - widgets/timeline/shared tabs"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/components/WidgetRenderer.tsx
      Note: "Widget tree renderer - maps UINode to React components"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/contracts.ts
      Note: "Core runtime contracts - LoadedPlugin, DispatchIntent, WorkerRequest/Response"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/uiTypes.ts
      Note: "UINode type system - the widget kind vocabulary"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/packages/plugin-runtime/src/redux-adapter/store.ts
      Note: "Redux store - state, reducers, selectors, policy, capability grants"
    - Path: /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/lib/presetPlugins.ts
      Note: "6 preset plugins with embedded code strings and capability declarations"
ExternalSources: []
Summary: "Implementation diary for the UI redesign analysis and design document creation for the plugin-playground developer workbench."
LastUpdated: 2026-02-09T09:01:00Z
WhatFor: "Track the analysis and design work for the plugin-playground UI overhaul"
WhenToUse: "Reference during UI redesign implementation; review checkpoint for design decisions"
---

# Diary - UI Redesign Analysis

## Goal

Capture the step-by-step analysis of the current plugin-playground UI, the design reasoning for a better developer-oriented workbench, and produce a comprehensive design document with ASCII wireframes that can serve as the implementation blueprint.

## Step 1: Deep Codebase Analysis

I read every source file in the plugin-playground to build a complete mental model of the current architecture, UI layout, data flow, and runtime contracts. The system is a QuickJS-sandboxed plugin runtime with a React frontend that lets developers load, run, and inspect plugins through a web UI.

### Prompt Context

**User prompt (verbatim):** "Work on ticket WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL, keep a frequent diary, and analyze the plugin-playground UI in /home/manuel/workspaces/2026-02-08/plugin-playground/go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground and design a better one, using ascii screenshots and laying out the functionality, in a new design document that you then store in the ticket and upload to remarkable. The goal being to make this a dev oriented UI on top of a preusable package."

**Assistant interpretation:** Analyze the current plugin-playground UI exhaustively, then produce a detailed design document with ASCII wireframes for a redesigned developer-oriented workbench UI. Store the design doc in the existing ticket and upload to reMarkable.

**Inferred user intent:** Get a thorough, tangible design blueprint (not vague suggestions) for how the playground should look and work as a developer tool, with enough detail to implement directly.

### What I did

- Read all 15+ source files across client/, packages/plugin-runtime/, server/, docs/
- Read the existing WEBVM-003 design docs (01 and 02) for prior analysis
- Read package.json for dependency inventory
- Read ideas.md for design philosophy context
- Mapped the component tree: App → Router → Home → Playground → {CatalogShell, WorkspaceShell, InspectorShell}
- Mapped the data flow: Playground.tsx → sandboxClient (worker RPC) → runtimeService (QuickJS) → intents → Redux store → re-render
- Inventoried the UINode kinds: panel, row, column, text, badge, button, input, counter, table
- Identified 6 preset plugins: counter, calculator, status-dashboard, greeter, greeter-shared-state, runtime-monitor
- Cataloged capability model: readShared/writeShared/systemCommands with 4 shared domains

### Why

Can't design a better UI without understanding exactly what the current one does, what data flows through it, what runtime contracts constrain it, and where the pain points are.

### What worked

- The codebase is well-organized after the WEBVM-003 Phase B extraction (plugin-runtime is already a separate package)
- Phase C (UI overhaul) was planned but not yet implemented - the 3 shell components exist but are minimal
- The existing design doc 02 already identifies the right problems (monolithic Playground, missing timeline filtering, missing capability inspection)

### What I learned

Key findings about the current UI:
1. **3-column equal-width layout** (CatalogShell | WorkspaceShell | InspectorShell) - no column is emphasized
2. **CatalogShell** is just a flat button list of 6 presets + a loaded-instances list with X buttons
3. **WorkspaceShell** is a bare textarea + one "LOAD PLUGIN" button + error display
4. **InspectorShell** has 3 tabs (WIDGETS/TIMELINE/SHARED) - the most functional panel
5. **No syntax highlighting** in the code editor
6. **No way to see plugin state** directly - only through widget rendering
7. **No capability visualization** - you can't see what grants a plugin has
8. **Timeline filtering** exists but is basic (scope/outcome/instance text filter)
9. **Widget rendering** is in the Inspector panel, separated from the code editor - awkward edit/preview loop
10. **No responsive design** - 3 columns at lg breakpoint, stacked on mobile

### What was tricky to build

N/A - this was analysis only, no code changes.

### What warrants a second pair of eyes

N/A - analysis step.

### What should be done in the future

Proceed to Step 2: design the new UI layout with ASCII wireframes.

### Code review instructions

N/A - no code changes in this step.

## Step 2: Design Document Creation

I wrote a comprehensive design document (~56KB) with 16 sections covering the complete UI redesign. The document includes ASCII wireframes for every major screen state, component architecture, interaction flows, data flow changes, design decisions with rationale, the package boundary between reusable runtime and app-specific UI, and a phased implementation plan.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Produce the actual design document with ASCII wireframes, store it in the ticket.

**Inferred user intent:** Have a detailed, implementable blueprint for the new UI.

### What I did

- Created design doc `03-plugin-playground-developer-workbench-ui-redesign.md` via docmgr
- Wrote 16 sections with detailed ASCII wireframes:
  1. Problems with current UI (12 specific issues with a comparison table)
  2. Master layout wireframe (IDE-like: sidebar + editor/preview + devtools)
  3. Layout dimensions and breakpoints
  4. Sidebar expanded/collapsed states with interaction table
  5. Editor pane with tab bar, line numbers, syntax highlighting mockup
  6. Live Preview pane with instance cards and empty states
  7. DevTools panel with 5 tabs: Timeline, State, Capabilities, Errors, Shared Domains
  8. Timeline tab with filters, color-coded outcomes, expandable rows
  9. State tab with per-instance plugin state + global state projection
  10. Capabilities tab with visual grant/deny grid per instance
  11. Errors tab with error log stream
  12. Shared Domains tab with reader/writer attribution
  13. Top toolbar with runtime status badges
  14. 3 detailed interaction flow narratives
  15. Keyboard shortcuts table
  16. Empty/zero state wireframes

### Why

The current UI has 12 specific problems that make it awkward for the developer edit→run→inspect loop. The biggest issues are: editor only gets 33% width, preview is separated from editor in a different panel, no syntax highlighting, no capability visualization, and the most useful debugging tool (timeline) is hidden behind a tab.

### What worked

- The IDE-like layout (sidebar + editor/preview + bottom devtools) directly addresses every identified problem
- ASCII wireframes are precise enough to implement from (specific pixel sizes, component names, interaction behaviors)
- The component architecture maps cleanly to React feature modules
- The package boundary is well-defined: runtime package has zero React/UI dependencies

### What didn't work

N/A - design work, not implementation.

### What I learned

- The existing InspectorShell already has the timeline filtering that the new design needs — it just needs to be promoted from a hidden tab to a first-class devtools panel
- The 3-shell architecture (CatalogShell, WorkspaceShell, InspectorShell) was a good intermediate step but the wrong final layout for a developer tool
- The key insight: preview belongs next to the editor (same eye-line), not in a separate column

### What was tricky to build

Designing the sidebar to show capability grants compactly was tricky. Full domain names like "counter-summary" and "greeter-profile" are too long for a 240px sidebar. Solution: abbreviate in the sidebar (ctr-sum, grt-prof) with full names in the Capabilities devtools tab.

### What warrants a second pair of eyes

- The 5-tab devtools panel might be too many tabs. Could Capabilities be folded into State tab? Could Shared Domains be folded into State tab? Worth discussing.
- The "instance focus" behavior (click instance → all panels focus) is powerful but might be disorienting if the user doesn't expect their timeline filter to change.

### What should be done in the future

1. Implement Phase 1 (layout shell) to validate the design in browser
2. Evaluate CodeMirror 6 vs Monaco for the code editor
3. Consider adding localStorage persistence for editor tabs

### Code review instructions

Read the design doc: `ttmp/.../design-doc/03-plugin-playground-developer-workbench-ui-redesign.md`
- Start at Section 2.1 (master layout) for the big picture
- Section 10 for component architecture
- Section 12 for design decisions with rationale

## Step 3: Ticket Bookkeeping and reMarkable Upload

Stored the design doc and diary in the WEBVM-003 ticket via docmgr, updated the changelog, related files, and uploaded a bundled PDF (design doc + diary) to the reMarkable cloud at `/ai/2026/02/09/WEBVM-003/`.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete the workflow: store in ticket, upload to reMarkable.

**Inferred user intent:** Have the design available for offline review on the reMarkable tablet.

### What I did

- Created diary doc via `docmgr doc add --doc-type reference`
- Created design doc via `docmgr doc add --doc-type design-doc`
- Related 8+ files to the design doc via frontmatter
- Updated ticket changelog with the design doc creation entry
- Uploaded bundled PDF (design doc + diary) to reMarkable at `/ai/2026/02/09/WEBVM-003/`
- Bundle includes table of contents with depth 2

### Why

The user explicitly asked for ticket storage + reMarkable upload. Having the design doc as a PDF on the tablet enables offline review and annotation.

### What worked

- `remarquee upload bundle` combined both docs into a single PDF with ToC
- Dry run confirmed paths and remote directory before actual upload
- Upload succeeded on first attempt

### What didn't work

N/A.

### What I learned

- `remarquee upload bundle` with `--toc-depth 2` is the cleanest way to get multiple markdown docs into a single navigable reMarkable PDF.

### What was tricky to build

N/A - standard workflow.

### What warrants a second pair of eyes

N/A.

### What should be done in the future

N/A.

### Code review instructions

- Check `docmgr doc list --ticket WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL` to verify docs are registered
- Check `remarquee cloud ls /ai/2026/02/09/WEBVM-003 --long --non-interactive` to verify upload

## Step 4: Docs Panel Addition

Added a comprehensive Docs panel design to the existing design document. The Docs panel is a new DevTools tab that embeds all project markdown documentation inside the workbench, with a tree navigation sidebar, rendered markdown display, and three tiers of copy-to-clipboard (per code block, per doc, all docs). The design includes the full Vite `?raw` bundling strategy, React component tree, `docsManifest.ts` module with concrete code, `renderMarkdown.ts` with custom marked renderer, styling spec, and interaction table.

### Prompt Context

**User prompt (verbatim):** "add a docs pane where one can read the docs, and also copy to clipboard so that things can be pasted to an LLM. It should display the documentation we have in the repo, maybe through some bundling of assets so it can be displayed in the UI"

**Assistant interpretation:** Design a new panel in the workbench that displays the 5 existing documentation markdown files from the `docs/` directory, with copy-to-clipboard functionality optimized for pasting into LLM conversations.

**Inferred user intent:** Developers should be able to read project docs and copy them (especially for LLM prompts) without leaving the workbench UI. The docs should be bundled at build time so no runtime file access is needed.

### What I did

- Added section 5.7 "DevTools — Docs Tab" to the design doc with 8 subsections:
  - 5.7.1: Layout wireframe (220px nav + rendered doc pane, ASCII art)
  - 5.7.2: Navigation tree structure mirroring `docs/` directory
  - 5.7.3: Three copy-to-clipboard targets (per code block, per doc, all docs) with behavior table
  - 5.7.4: "Copy All Docs" output format specification (markdown with file provenance headers)
  - 5.7.5: Build-time bundling strategy using Vite `?raw` imports — includes full code for `docsManifest.ts` and `renderMarkdown.ts`
  - 5.7.6: React component architecture (DocsPanel → DocsNav + DocsContent)
  - 5.7.7: Interaction table
  - 5.7.8: Styling spec using existing theme tokens
- Added design decision D7 explaining the "embedded docs with raw markdown copy" rationale
- Updated master layout wireframe to show [📖Docs] tab
- Updated DevTools tab overview bar
- Updated component architecture tree to include DocsPanel subtree
- Updated new-vs-existing components table (+3 new components)
- Updated package boundary section to include docsManifest.ts and renderMarkdown.ts
- Updated implementation plan with Phase 4b (9 steps)
- Updated keyboard shortcuts table (+Ctrl+Shift+D)
- Added 2 new open questions (#6: rendering library, #7: auto-discovery vs manifest)
- Related 5 docs/*.md files to the design doc via docmgr

### Why

The current workflow for sharing docs with an LLM is: find the file in a terminal, cat it, copy from terminal, paste into LLM. This is friction-heavy. Embedding docs in the UI with one-click copy makes the LLM workflow seamless.

### What worked

- Vite `?raw` imports are the simplest bundling approach — zero build plugins needed, the docs become inline strings in the JS bundle
- The total docs payload is only ~12.5KB raw (~4KB gzipped), so bundle size impact is negligible
- The three-tier copy model (code block / single doc / all docs) covers the three main LLM paste workflows: "here's one example", "here's the full doc", "here's everything you need to know"

### What didn't work

N/A.

### What I learned

- Vite `?raw` imports work with any file path the bundler can resolve — no config changes needed since `docs/` is within the project root
- `marked` is only ~40KB and sufficient for rendering + custom code block handling via renderer override
- The custom renderer approach (overriding `renderer.code`) is the cleanest way to inject copy buttons per code fence without a full AST walk

### What was tricky to build

The "Copy All Docs" format needed thought. Simply concatenating the raw markdown would lose file provenance — the LLM wouldn't know which doc each section came from. Solution: prefix each doc with a `# docs/path/to/file.md` header and separate with `---` rules. This gives the LLM clear file boundaries.

Also: the code block copy button needed a way to reference the raw code content from a click handler on the rendered HTML. Solution: encode the raw code into a `data-raw` attribute on the wrapper div, then use event delegation to read it back on click. This avoids maintaining a parallel data structure.

### What warrants a second pair of eyes

- The `dangerouslySetInnerHTML` for rendered markdown needs sanitization review. `marked` doesn't sanitize by default. Should add DOMPurify or marked's built-in sanitizer option.
- Whether 6 DevTools tabs is too many and Docs should be a separate top-level view rather than a devtools tab. Counter-argument: devtools is where developers look for reference material.

### What should be done in the future

1. If docs grow beyond ~15 files, replace the explicit `docsManifest.ts` with a Vite plugin that auto-discovers `docs/**/*.md` via `import.meta.glob`
2. Consider adding search within docs (simple text filter on raw content)
3. Consider adding "Copy as context prompt" that wraps docs in an LLM-friendly system prompt template

### Code review instructions

- Read §5.7 in the design doc for the full Docs panel spec
- Read §D7 for the design decision rationale
- Check §5.7.5 for the concrete `docsManifest.ts` and `renderMarkdown.ts` code

## Step 5: Storybook Setup, RTK State, Themable Components (T11+T12+T13)

Installed Storybook 10, created the RTK workbench slice for UI state, built the foundational WorkbenchLayout and Sidebar components using vm-system-ui visual style with the react-modular-themable-storybook data-part/CSS-variable pattern, and wired everything through stories with a Redux Provider decorator.

### Prompt Context

**User prompt (verbatim):** "add detailed implementation tasks... use storybook stories as the main way to scaffold... use rtk for state... match the style of vm-system-ui, but use the themable pattern."

**Assistant interpretation:** Create RTK-managed UI state, build components matching vm-system-ui's slate/Inter/JetBrains Mono visual language while using data-widget/data-part/CSS variable theming hooks, and scaffold everything story-first.

**Inferred user intent:** Get a working component system that is visually consistent with the existing vm-system-ui project, architecturally sound with proper state management and theming extension points, and incrementally verifiable through Storybook.

**Commit (code):** 5169169 — "T12+T13: RTK workbench slice, vm-system-ui style, themable layout + sidebar"

### What I did

- Installed Storybook 10 (`@storybook/react-vite`), configured `.storybook/main.ts` with Vite path aliases and `.storybook/preview.ts` with dark theme decorator + global Redux Provider
- Exported `runtimeReducer` from `packages/plugin-runtime/src/redux-adapter/store.ts` to enable clean store composition
- Created `client/src/store/workbenchSlice.ts`: layout state (sidebar/devtools collapse), editor tabs, instance focus, devtools active tab, error log — all as RTK `createSlice` with typed actions
- Created `client/src/store/index.ts`: composed store with `{ runtime: runtimeReducer, workbench: workbenchReducer }`, typed hooks (`useAppDispatch`/`useAppSelector`), re-exports of all runtime + workbench actions/selectors
- Created `workbench.css`: `:where([data-widget="workbench"])` scoped CSS variables (--wb-color-*, --wb-font-*, --wb-space-*, --wb-sidebar-width etc.) with low-specificity selectors for layout structure
- Created `WorkbenchLayout.tsx`: reads sidebar/devtools state from RTK store, supports prop overrides for stories, `unstyled` prop for consumer CSS
- Created `Sidebar.tsx`: pure presentational component using vm-system-ui slate palette (slate-800, slate-500 borders, Inter font, blue accent, emerald status dots)
- Created `ConnectedSidebar.tsx`: thin RTK wrapper that selects running instances from runtime state and dispatches workbench actions
- Created `storyDecorators.tsx`: `withStore()` factory returning Storybook `Decorator` type, with runtime stub reducer and configurable workbench preload
- Built 14 stories across 3 story files, all rendering cleanly

### Why

The user asked for vm-system-ui visual consistency + themable pattern + RTK state + Storybook scaffolding. This combines all four: the components look like vm-system-ui (same color palette, typography, interaction patterns), use data-part attributes and CSS variables for theming extensibility, read/write UI state through RTK, and are fully testable in Storybook.

### What worked

- The global `withStore()` decorator in `preview.ts` ensures every story gets a Redux Provider automatically, eliminating the "could not find react-redux context value" error
- Per-story `withStore({ sidebarCollapsed: true })` works via nested Provider shadowing — inner Provider wins
- Exporting `runtimeReducer` from the runtime package was a one-line change that enabled clean store composition without hacks
- The presentational/connected split (Sidebar vs ConnectedSidebar) keeps stories simple while the real app gets full RTK wiring

### What didn't work

- First storybook startup failed: `__dirname is not defined` in ESM context. Fixed by using `import.meta.url` + `fileURLToPath`.
- First `withStore` decorator had wrong signature (`(Story: React.ComponentType)` instead of Storybook's `Decorator` type). Fixed by using `@storybook/react-vite`'s `Decorator` type.
- `@storybook/test` module not found for `fn()` — Storybook 10 moved it to `storybook/test`. Fixed import.

### What I learned

- Storybook 10 uses `storybook/test` not `@storybook/test` for `fn()`
- The `:where()` pseudo-class for CSS selectors is perfect for themable component systems — zero specificity means any consumer CSS wins
- vm-system-ui's visual language is: slate-950 background, slate-900 cards, white/[0.06] borders (not oklch border variables), Inter for UI text, JetBrains Mono for code, blue-600 accent, emerald-500 success, compact text-xs/text-sm sizing

### What was tricky to build

Getting the RTK store composition right without depending on the runtime package's singleton `store`. The runtime package exports `configureStore({ reducer: { runtime: runtimeSlice.reducer } })` as a singleton, but we need to compose it with our workbench reducer. Solution: export the raw slice reducer (`runtimeReducer = runtimeSlice.reducer`) from the runtime package and compose in the app.

The storybook decorator typing was also tricky — Storybook 10's `Decorator` type expects `(Story, context) => ReactNode` where `Story` is a component, and you render it as `<Story />`.

### What warrants a second pair of eyes

- The `storyDecorators.tsx` runtime stub reducer shape must stay in sync with what runtime selectors expect. If the runtime state shape changes, the stub needs updating or stories will break.
- The `withStore()` global + per-story nested Provider pattern works but is unconventional. If Storybook changes Provider resolution order, it could break.

### What should be done in the future

- Continue with T14-T24: TopToolbar, EditorTabBar, CodeEditor, LivePreview, DevTools panels, DocsPanel
- Consider extracting the `storyDecorators` runtime stub into a `@runtime/test-utils` export

### Code review instructions

- Start at `client/src/store/workbenchSlice.ts` for the RTK state shape
- Check `client/src/features/workbench/styles/workbench.css` for the CSS variable theming contract
- Check `client/src/features/workbench/components/Sidebar.tsx` for the vm-system-ui visual style
- Run `pnpm storybook` and browse all 14 stories at http://localhost:6006

---

## Appendix: Technical Reference

### Technical details

Current component tree:
```
App
├── ErrorBoundary
├── Provider (Redux store)
├── ThemeProvider (dark-only)
├── TooltipProvider
├── Toaster (sonner)
└── Router
    └── Home → Playground
        ├── CatalogShell (left 1/3)
        │   ├── Preset buttons (6)
        │   └── Loaded instances list
        ├── WorkspaceShell (center 1/3)
        │   ├── textarea (custom code)
        │   ├── LOAD PLUGIN button
        │   └── Error display
        └── InspectorShell (right 1/3)
            ├── Tab: WIDGETS → WidgetRenderer per widget
            ├── Tab: TIMELINE → filtered dispatch entries
            └── Tab: SHARED → JSON.stringify(sharedState)
```

Current data flow:
```
User clicks preset/loads custom code
  → quickjsSandboxClient.loadPlugin() [worker RPC]
  → QuickJSRuntimeService.loadPlugin() [QuickJS eval]
  → returns LoadedPlugin metadata
  → dispatch(pluginRegistered(..., grants))
  → Redux state updated

On state change:
  → Playground useEffect re-renders ALL widgets
  → quickjsSandboxClient.render() for each widget
  → WidgetRenderer maps UINode tree to React

On user interaction with widget:
  → WidgetRenderer calls onEvent(eventRef, payload)
  → quickjsSandboxClient.event() → returns DispatchIntent[]
  → Each intent dispatched: pluginActionDispatched or sharedActionDispatched
  → Redux reducer applies/denies/ignores → timeline entry
  → State change triggers re-render cycle
```
