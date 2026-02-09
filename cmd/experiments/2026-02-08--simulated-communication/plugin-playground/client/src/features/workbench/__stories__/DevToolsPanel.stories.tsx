import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { DevToolsPanel } from "../components/DevToolsPanel";
import { TimelinePanel, type TimelineEntry } from "../components/TimelinePanel";
import { StatePanel } from "../components/StatePanel";
import { ErrorsPanel } from "../components/ErrorsPanel";
import { DocsPanel } from "../components/DocsPanel";
import type { DevToolsTab } from "@/store/workbenchSlice";
import { MOCK_TIMELINE, MOCK_PLUGIN_STATE, MOCK_PLUGINS, MOCK_ERRORS } from "./storyDecorators";

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

// ---------------------------------------------------------------------------
// Interactive story — click tabs, toggle collapse, see real panel content
// ---------------------------------------------------------------------------

const TIMELINE_ENTRIES: TimelineEntry[] = MOCK_TIMELINE.map((d, i) => ({
  id: `tl-${i}`,
  timestamp: d.timestamp,
  scope: d.scope as "plugin" | "shared",
  outcome: d.outcome as "applied" | "denied" | "ignored",
  actionType: d.actionType,
  instanceId: d.instanceId ?? "",
  shortInstanceId: (d.instanceId ?? "").slice(0, 10),
  domain: d.domain ?? undefined,
  reason: d.reason ?? undefined,
}));

const INSTANCE_STATES = Object.entries(MOCK_PLUGINS).map(([id, p]) => ({
  instanceId: id,
  title: (p as any).title,
  shortId: id.slice(0, 10),
  state: MOCK_PLUGIN_STATE[id] ?? {},
}));

const FIXTURE_DOCS = [
  { title: "Overview", category: "Overview", path: "docs/README.md", raw: "# Plugin Playground\n\nA sandbox for plugin development." },
  { title: "Quickstart", category: "Plugin Authoring", path: "docs/quickstart.md", raw: "# Quickstart\n\nCall `definePlugin()` to get started." },
];

function InteractiveDevTools() {
  const [activeTab, setActiveTab] = React.useState<DevToolsTab>("timeline");
  const [collapsed, setCollapsed] = React.useState(false);
  const [errors, setErrors] = React.useState(MOCK_ERRORS);
  const [focusedId, setFocusedId] = React.useState<string | null>(null);

  const tabContent: Record<DevToolsTab, React.ReactNode> = {
    timeline: <TimelinePanel entries={TIMELINE_ENTRIES} focusedInstanceId={focusedId} />,
    state: <StatePanel instances={INSTANCE_STATES} focusedInstanceId={focusedId} onFocusInstance={setFocusedId} />,
    capabilities: <div className="p-3 text-xs text-slate-500">Capabilities panel (see dedicated story)</div>,
    errors: <ErrorsPanel errors={errors} onClear={() => setErrors([])} />,
    shared: <div className="p-3 text-xs text-slate-500">Shared domains panel (see dedicated story)</div>,
    docs: <DocsPanel docs={FIXTURE_DOCS} />,
  };

  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: collapsed ? 36 : 320 }}>
      <DevToolsPanel
        activeTab={activeTab}
        collapsed={collapsed}
        errorCount={errors.length}
        onSelectTab={setActiveTab}
        onToggleCollapse={() => setCollapsed((c) => !c)}
      >
        {tabContent[activeTab]}
      </DevToolsPanel>
    </div>
  );
}

export const Interactive: Story = {
  args: { activeTab: "timeline" },
  render: () => <InteractiveDevTools />,
};
