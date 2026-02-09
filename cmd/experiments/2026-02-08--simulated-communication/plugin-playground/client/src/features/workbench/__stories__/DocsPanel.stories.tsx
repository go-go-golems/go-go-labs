import type { Meta, StoryObj } from "@storybook/react-vite";
import { DocsPanel, type DocEntry } from "../components/DocsPanel";

// ---------------------------------------------------------------------------
// Fixture docs (inline markdown so stories don't depend on ?raw imports)
// ---------------------------------------------------------------------------

const FIXTURE_DOCS: DocEntry[] = [
  {
    title: "Overview",
    category: "Overview",
    path: "docs/README.md",
    raw: `# Plugin Playground

A browser-based sandbox for developing, testing, and inspecting plugins
that run inside the VM runtime.

## Quick Start

1. Write a plugin in the code editor
2. Click **Run** to load it into the sandbox
3. See the live output in the preview pane
4. Inspect dispatches in the Timeline tab

## Features

- Hot-reload plugin code
- Shared state across plugin instances
- Capability-based security model
- Full dispatch timeline with filtering`,
  },
  {
    title: "Quickstart",
    category: "Plugin Authoring",
    path: "docs/plugin-authoring/quickstart.md",
    raw: `# Plugin Authoring Quickstart

Every plugin is a single JavaScript file that calls \`definePlugin()\`.

## Minimal Example

\`\`\`js
definePlugin(({ ui }) => ({
  id: "hello",
  title: "Hello World",
  render: () => ui.text("Hello, world!"),
}));
\`\`\`

## With State

\`\`\`js
definePlugin(({ ui }) => ({
  id: "counter",
  title: "Counter",
  reduceEvent(state = { count: 0 }, event) {
    if (event.type === "increment") return { count: state.count + 1 };
    return state;
  },
  render(state) {
    return ui.column([
      ui.text(\`Count: \${state.count}\`),
      ui.button("+1", { event: { type: "increment" } }),
    ]);
  },
}));
\`\`\`

## Capabilities

Plugins declare which shared domains they can read/write:

\`\`\`js
capabilities: {
  readShared: ["counter-summary"],
  writeShared: ["counter-summary"],
}
\`\`\``,
  },
  {
    title: "Capability Model",
    category: "Architecture",
    path: "docs/architecture/capability-model.md",
    raw: `# Capability Model

The runtime enforces a capability-based security model for shared state.

## Shared Domains

| Domain | Purpose |
|--------|---------|
| counter-summary | Aggregate counter values |
| greeter-profile | Shared greeter name |
| runtime-registry | Plugin registry metadata |
| runtime-metrics | Runtime performance data |

## Grant Enforcement

- **Read grants**: plugin receives current domain state in every render cycle
- **Write grants**: plugin can dispatch shared actions targeting the domain
- **Denied dispatches**: logged in timeline with "denied" outcome`,
  },
  {
    title: "Embedding Guide",
    category: "Runtime",
    path: "docs/runtime/embedding.md",
    raw: `# Runtime Embedding Guide

The plugin-runtime package can be embedded in any web application.

## Installation

\`\`\`bash
npm install @go-go-labs/plugin-runtime
\`\`\`

## Basic Usage

\`\`\`ts
import { createRuntime } from "@go-go-labs/plugin-runtime";

const runtime = createRuntime();
await runtime.loadPlugin(code, { capabilities: { readShared: [], writeShared: [] } });
\`\`\``,
  },
  {
    title: "VM API Changelog",
    category: "Migration",
    path: "docs/migration/changelog-vm-api.md",
    raw: `# VM API Changelog

## v0.3.0 (current)

- Added shared domain capability grants
- Added dispatch timeline logging
- Added \`reduceSharedEvent\` hook

## v0.2.0

- Introduced \`definePlugin()\` API
- Added \`ui.row()\`, \`ui.column()\`, \`ui.button()\`, \`ui.text()\`
- Added event-driven state management via \`reduceEvent\`

## v0.1.0

- Initial prototype
- Simple eval-based plugin loading`,
  },
];

const ALL_DOCS_MD = FIXTURE_DOCS.map((d) => `# ${d.path}\n\n${d.raw}`).join("\n\n---\n\n");

// ---------------------------------------------------------------------------
// Wrapper
// ---------------------------------------------------------------------------

function Wrapper(props: React.ComponentProps<typeof DocsPanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 400 }}>
      <DocsPanel {...props} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

const meta = {
  title: "Workbench/Panels/DocsPanel",
  component: DocsPanel,
  render: (args) => <Wrapper {...args} />,
} satisfies Meta<typeof DocsPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    docs: FIXTURE_DOCS,
    allDocsMarkdown: ALL_DOCS_MD,
  },
};

export const SingleDoc: Story = {
  args: {
    docs: [FIXTURE_DOCS[0]],
  },
};

export const Empty: Story = {
  args: {
    docs: [],
  },
};
