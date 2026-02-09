# Plugin Playground Documentation

The Plugin Playground is a browser-based sandbox for developing, testing, and inspecting **plugins** that run inside an isolated JavaScript runtime. Each plugin is a small, self-contained program that declares its own state, renders a UI, and communicates with other plugins through **shared domains** — all while running safely in a QuickJS sandbox that prevents plugins from accessing the host page directly.

Think of it as a miniature operating system for UI widgets: each plugin gets its own process (a sandboxed JS context), its own memory (local state), and a controlled set of system calls (dispatch intents) that the host runtime evaluates according to a capability-based security policy.

## Who is this for?

- **Plugin authors** who want to build interactive widgets that run inside the playground
- **Runtime embedders** who want to use `plugin-runtime` in their own applications
- **Contributors** who want to understand the architecture before modifying the runtime

## Core Concepts

Before diving into the docs, here's a quick map of the key ideas:

```
┌─────────────────────────────────────────────────────────────┐
│                    Plugin Playground                         │
│                                                             │
│  ┌──────────┐   ┌──────────┐   ┌──────────┐               │
│  │ Plugin A  │   │ Plugin B  │   │ Plugin C  │    Sandboxed │
│  │ (Counter) │   │ (Greeter) │   │ (Monitor) │    plugins   │
│  └─────┬─────┘   └─────┬─────┘   └─────┬─────┘             │
│        │               │               │                    │
│        ▼               ▼               ▼                    │
│  ┌─────────────────────────────────────────────────┐        │
│  │              Dispatch Pipeline                   │        │
│  │  Intent → Policy Check → Reducer → New State    │        │
│  └─────────────────────┬───────────────────────────┘        │
│                        │                                     │
│  ┌─────────────────────▼───────────────────────────┐        │
│  │              Shared Domains                      │        │
│  │  counter-summary │ greeter-profile │ runtime-*   │        │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

- **Plugin** — A JavaScript file that calls `definePlugin()` to declare its behavior
- **UI DSL** — A set of builder functions (`ui.text()`, `ui.button()`, etc.) that produce JSON UI trees
- **Dispatch Intent** — A message a plugin emits saying "I want to change state"
- **Shared Domain** — A named piece of shared state that multiple plugins can read from and write to
- **Capability Grant** — Permission assigned to a plugin instance to read or write a shared domain

## Documentation Map

### Getting Started
- **[Plugin Authoring Quickstart](plugin-authoring/quickstart.md)** — Write your first plugin, understand the full API, and learn the patterns
- **[Plugin Examples](plugin-authoring/examples.md)** — Worked examples from simple to advanced

### Architecture
- **[Dispatch Lifecycle](architecture/dispatch-lifecycle.md)** — How actions flow from user click to state change
- **[UI DSL Reference](architecture/ui-dsl.md)** — Every UI node type, its props, and how rendering works
- **[Capability Model](architecture/capability-model.md)** — Shared domain access control, grant policies, and the security model

### Embedding & Migration
- **[Runtime Embedding Guide](runtime/embedding.md)** — Use `plugin-runtime` in your own app
- **[VM API Changelog](migration/changelog-vm-api.md)** — Breaking changes and migration notes
