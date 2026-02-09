import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { DevToolsPanel } from "../components/DevToolsPanel";

// ---------------------------------------------------------------------------
// Placeholder tab content
// ---------------------------------------------------------------------------

function TimelinePlaceholder() {
  return (
    <div className="p-3 font-mono text-xs overflow-auto h-full">
      <table className="w-full text-left">
        <thead>
          <tr className="text-slate-600">
            <th className="pb-2 pr-4 font-medium">Time</th>
            <th className="pb-2 pr-4 font-medium">Scope</th>
            <th className="pb-2 pr-4 font-medium">Outcome</th>
            <th className="pb-2 pr-4 font-medium">Action</th>
            <th className="pb-2 font-medium">Instance</th>
          </tr>
        </thead>
        <tbody className="text-slate-400">
          <tr><td className="py-0.5 pr-4 text-slate-600">09:01:23.4</td><td className="py-0.5 pr-4">plugin</td><td className="py-0.5 pr-4 text-emerald-500">applied</td><td className="py-0.5 pr-4 text-slate-300">increment</td><td className="py-0.5 text-slate-600">abc-1234</td></tr>
          <tr><td className="py-0.5 pr-4 text-slate-600">09:01:23.4</td><td className="py-0.5 pr-4">shared</td><td className="py-0.5 pr-4 text-emerald-500">applied</td><td className="py-0.5 pr-4 text-slate-300">set-instance</td><td className="py-0.5 text-slate-600">abc-1234</td></tr>
          <tr><td className="py-0.5 pr-4 text-slate-600">09:01:22.1</td><td className="py-0.5 pr-4">shared</td><td className="py-0.5 pr-4 text-red-400">denied</td><td className="py-0.5 pr-4 text-slate-300">set-name</td><td className="py-0.5 text-slate-600">def-5678</td></tr>
        </tbody>
      </table>
    </div>
  );
}

function StatePlaceholder() {
  return (
    <div className="p-3 font-mono text-xs text-slate-400">
      <div className="text-slate-500 mb-2">Instance: abc-1234 (Counter)</div>
      <pre className="text-slate-300">{JSON.stringify({ count: 5 }, null, 2)}</pre>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Wrapper
// ---------------------------------------------------------------------------

function Wrapper(props: React.ComponentProps<typeof DevToolsPanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 280 }}>
      <DevToolsPanel {...props} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

const meta = {
  title: "Workbench/DevToolsPanel",
  component: DevToolsPanel,
  render: (args) => <Wrapper {...args} />,
  args: {
    onSelectTab: fn(),
    onToggleCollapse: fn(),
  },
} satisfies Meta<typeof DevToolsPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const TimelineTab: Story = {
  args: {
    activeTab: "timeline",
    children: <TimelinePlaceholder />,
  },
};

export const StateTab: Story = {
  args: {
    activeTab: "state",
    children: <StatePlaceholder />,
  },
};

export const WithErrors: Story = {
  args: {
    activeTab: "errors",
    errorCount: 3,
    children: (
      <div className="p-3 text-xs font-mono text-red-400 space-y-1">
        <div>[09:01:22.1] TypeError: Cannot read properties of undefined</div>
        <div>[09:01:18.7] ReferenceError: x is not defined</div>
        <div>[09:01:15.2] Plugin load failed: syntax error at line 12</div>
      </div>
    ),
  },
};

export const Collapsed: Story = {
  args: {
    activeTab: "timeline",
    collapsed: true,
    children: <TimelinePlaceholder />,
  },
};

export const DocsTab: Story = {
  args: {
    activeTab: "docs",
    children: (
      <div className="p-3 text-xs text-slate-400">
        <h3 className="text-sm font-medium text-slate-200 mb-2">Plugin Authoring Quickstart</h3>
        <p>Every plugin is a single JavaScript file that calls <code className="text-blue-400">definePlugin()</code>...</p>
      </div>
    ),
  },
};
