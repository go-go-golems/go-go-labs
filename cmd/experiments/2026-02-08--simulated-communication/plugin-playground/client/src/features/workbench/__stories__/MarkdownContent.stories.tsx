import type { Meta, StoryObj } from "@storybook/react-vite";
import { MarkdownContent } from "../components/MarkdownContent";

const RICH_MARKDOWN = `# Plugin Authoring Guide

Welcome to the **Plugin Playground**. This guide covers everything you need to write plugins.

## Quick Start

Every plugin is a single JavaScript file that calls \`definePlugin()\`:

\`\`\`js
definePlugin(({ ui }) => ({
  id: "hello",
  title: "Hello World",
  render: () => ui.text("Hello, world!"),
}));
\`\`\`

## State Management

Plugins manage state through \`reduceEvent\`:

\`\`\`typescript
definePlugin(({ ui }) => ({
  id: "counter",
  title: "Counter",
  initialState: { count: 0 },
  reduceEvent(state, event) {
    switch (event.type) {
      case "increment":
        return { count: state.count + 1 };
      case "decrement":
        return { count: state.count - 1 };
      default:
        return state;
    }
  },
  render(state) {
    return ui.column([
      ui.text(\`Count: \${state.count}\`),
      ui.row([
        ui.button("-", { event: { type: "decrement" } }),
        ui.button("+", { event: { type: "increment" } }),
      ]),
    ]);
  },
}));
\`\`\`

## Shared Domains

Plugins can read and write to **shared domains** for cross-instance communication:

| Domain | Purpose | Initial State |
|--------|---------|---------------|
| \`counter-summary\` | Aggregate counter values | \`{ totalValue: 0 }\` |
| \`greeter-profile\` | Shared greeter name | \`{ name: "" }\` |
| \`runtime-registry\` | Plugin registry metadata | \`[]\` |
| \`runtime-metrics\` | Runtime performance data | \`{}\` |

### Declaring Capabilities

\`\`\`js
capabilities: {
  readShared: ["counter-summary"],
  writeShared: ["counter-summary"],
}
\`\`\`

> **Note:** Plugins without write grants will have their shared dispatches *denied* by the runtime.

## UI Primitives

Available UI nodes:

- \`ui.text(content)\` — display text
- \`ui.button(label, opts)\` — clickable button
- \`ui.input(opts)\` — text input
- \`ui.row(children)\` — horizontal layout
- \`ui.column(children)\` — vertical layout
- \`ui.badge(text)\` — status badge

---

*Happy plugin authoring!*
`;

function Wrapper(props: { source: string }) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06] p-6" style={{ maxWidth: 700, maxHeight: 600, overflow: "auto" }}>
      <MarkdownContent {...props} />
    </div>
  );
}

const meta: Meta<typeof MarkdownContent> = {
  title: "Workbench/MarkdownContent",
  component: MarkdownContent,
  render: (args) => <Wrapper {...args} />,
};

export default meta;
type Story = StoryObj<typeof meta>;

export const RichDocument: Story = {
  args: { source: RICH_MARKDOWN },
};

export const CodeOnly: Story = {
  args: {
    source: "## Code Example\n\n```js\nconst x = 42;\nconsole.log(x);\n```\n\nInline code: `const y = 100`.",
  },
};

export const TableHeavy: Story = {
  args: {
    source: "## API Reference\n\n| Method | Args | Returns |\n|--------|------|--------|\n| `ui.text` | `string` | `UINode` |\n| `ui.button` | `label, opts` | `UINode` |\n| `ui.input` | `opts` | `UINode` |\n| `ui.row` | `UINode[]` | `UINode` |\n| `ui.column` | `UINode[]` | `UINode` |",
  },
};

export const Empty: Story = {
  args: { source: "" },
};
