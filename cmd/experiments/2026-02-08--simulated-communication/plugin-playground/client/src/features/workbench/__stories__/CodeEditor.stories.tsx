import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { CodeEditor } from "../components/CodeEditor";

// Sample plugin code
const COUNTER_CODE = `definePlugin(({ ui, shared }) => {
  let count = 0;

  return {
    id: "counter",
    title: "Counter",
    capabilities: {
      readShared: ["counter-summary"],
      writeShared: ["counter-summary"],
    },
    reduceEvent(state, event) {
      if (event.type === "increment") count++;
      if (event.type === "decrement") count--;
      if (event.type === "reset") count = 0;
      return { count };
    },
    render(state) {
      return ui.column([
        ui.text(\`Counter: \${state.count}\`),
        ui.row([
          ui.button("−", { event: { type: "decrement" } }),
          ui.button("Reset", { event: { type: "reset" } }),
          ui.button("+", { event: { type: "increment" } }),
        ]),
      ]);
    },
  };
});`;

function Wrapper(props: React.ComponentProps<typeof CodeEditor>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 400 }}>
      <CodeEditor {...props} />
    </div>
  );
}

// Stateful wrapper so typing works in Storybook
function StatefulEditor(props: React.ComponentProps<typeof CodeEditor>) {
  const [value, setValue] = React.useState(props.value);
  return <Wrapper {...props} value={value} onChange={setValue} />;
}

const meta = {
  title: "Workbench/CodeEditor",
  component: CodeEditor,
  render: (args) => <StatefulEditor {...args} />,
  args: {
    onChange: fn(),
    language: "javascript",
    readOnly: false,
  },
} satisfies Meta<typeof CodeEditor>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    value: COUNTER_CODE,
  },
};

export const EmptyWithPlaceholder: Story = {
  args: {
    value: "",
    placeholder: "// Start writing your plugin…",
  },
};

export const ReadOnly: Story = {
  args: {
    value: COUNTER_CODE,
    readOnly: true,
  },
};

export const ShortSnippet: Story = {
  args: {
    value: `definePlugin(({ ui }) => ({
  id: "hello",
  title: "Hello World",
  render: () => ui.text("Hello, world!"),
}));`,
  },
};
