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
- [ ] Write runtime embedding docs with package usage examples and migration notes
