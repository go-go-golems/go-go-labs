# Tasks

## TODO

- [x] Remove dead/template components and unused UI wrappers/dependencies from plugin-playground
- [x] Remove remaining debug/template leftovers (WidgetRenderer globals/logs, index analytics placeholders/comment block, unused helper imports/hooks)
- [x] Unify theme stack (remove custom-vs-next-themes split and keep one provider model)
- [x] Create one `plugin-runtime` package scaffold and migrate contracts + QuickJS core into it
- [x] Move worker wrapper/client transport and host adapter interfaces into `plugin-runtime` for non-UI embedding
- [x] Move Redux runtime policy/reducer/selector logic into `plugin-runtime` internal `redux-adapter` module
- [x] Refactor Playground into modular developer workbench UI shells (`catalog`, `workspace`, `inspector`)
- [x] Implement runtime timeline and shared-domain inspector panels with filters
- [x] Write plugin authoring + capability model docs (`quickstart`, domain reference)
- [x] Write runtime embedding docs with package usage examples and migration notes
- [x] T11: Install and configure Storybook 8 with React/Vite/Tailwind, dark theme decorator matching brutalist theme, and verify it runs
- [x] T12: Create WorkbenchLayout shell component + story (sidebar/main/devtools skeleton with data-part attributes)
- [x] T13: Create Sidebar component + stories (catalog tree, running instances with capability badges, collapse toggle)
- [x] T14: Create TopToolbar component + story (runtime status badges, plugin count, dispatch count, health indicator)
- [x] T15: Create EditorTabBar + CodeEditor components + stories (tab management, syntax highlighting placeholder, run button)
- [x] T16: Create LivePreview + InstanceCard components + stories (widget rendering area, instance header with status)
- [x] T17: Create DevToolsPanel container + stories (tab bar with 6 tabs, collapse/expand, drag-resize handle)
- [x] T18: Create TimelinePanel component + story (dispatch table with scope/outcome/domain filters, expandable rows)
- [x] T19: Create StatePanel component + story (per-instance plugin state and globalState JSON viewer)
- [x] T20: Create CapabilitiesPanel component + story (per-instance grant/deny grid for all shared domains)
- [x] T21: Create ErrorsPanel component + story (error log stream with timestamps, clear button, empty state)
- [x] T22: Create SharedDomainsPanel component + story (per-domain cards with state, reader/writer attribution)
- [x] T23: Create DocsPanel + docsManifest + renderMarkdown + stories (embedded docs viewer, tree nav, 3-tier copy-to-clipboard)
- [x] T24: Wire WorkbenchLayout into app router replacing Playground, connect to runtime hooks
- [ ] T25: Fix loadPreset stale tab ID — return generated tab ID from openEditorTab
- [ ] T26: Fix loadCustom empty grants — infer capabilities from preset or allow user specification
- [ ] T27: Restyle WidgetRenderer to match vm-system-ui slate palette
- [ ] T28: Remove dead Playground.tsx and old shell components
- [ ] T29: Add vertical drag-resize handle to DevToolsPanel
