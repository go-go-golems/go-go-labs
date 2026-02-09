import type { Meta, StoryObj } from "@storybook/react-vite";
import { WorkbenchLayout } from "../components/WorkbenchLayout";
import { Sidebar, type CatalogEntry, type RunningInstance } from "../components/Sidebar";
import { TopToolbar } from "../components/TopToolbar";
import { withStore } from "./storyDecorators";

// ---------------------------------------------------------------------------
// Shared fixture data
// ---------------------------------------------------------------------------

const CATALOG: CatalogEntry[] = [
  { id: "counter", title: "Counter", description: "Local counter + shared summary", capabilitySummary: "R/W" },
  { id: "calculator", title: "Calculator", description: "Basic arithmetic" },
  { id: "status-dashboard", title: "Status Dashboard", description: "Runtime metrics overview", capabilitySummary: "R" },
  { id: "greeter", title: "Greeter", description: "Input handling demo", capabilitySummary: "R/W" },
  { id: "greeter-shared-state", title: "Greeter Shared State", description: "Reads shared greeter profile", capabilitySummary: "R" },
  { id: "runtime-monitor", title: "Runtime Monitor", description: "Plugin registry table", capabilitySummary: "R" },
];

const RUNNING: RunningInstance[] = [
  {
    instanceId: "abc-1234-5678",
    title: "Counter",
    packageId: "counter",
    shortId: "abc-1234",
    status: "loaded",
    readGrants: ["counter-summary"],
    writeGrants: ["counter-summary"],
  },
  {
    instanceId: "def-5678-9012",
    title: "Greeter",
    packageId: "greeter",
    shortId: "def-5678",
    status: "loaded",
    readGrants: ["greeter-profile"],
    writeGrants: ["greeter-profile"],
  },
  {
    instanceId: "ghi-9012-3456",
    title: "Status Dashboard",
    packageId: "status-dashboard",
    shortId: "ghi-9012",
    status: "loaded",
    readGrants: ["counter-summary", "runtime-metrics", "runtime-registry"],
    writeGrants: [],
  },
];

// ---------------------------------------------------------------------------
// Placeholder slot fillers (vm-system-ui visual language)
// ---------------------------------------------------------------------------

function ToolbarPlaceholder() {
  return (
    <TopToolbar
      pluginCount={3}
      dispatchCount={47}
      health="healthy"
      errorCount={0}
    />
  );
}

function MainPlaceholder() {
  return (
    <div className="h-full flex text-sm">
      {/* Editor side */}
      <div className="flex-[6] border-r border-white/[0.06] flex flex-col">
        <div className="flex items-center h-9 px-2 border-b border-white/[0.06] gap-1">
          <div className="flex items-center gap-1 px-2.5 py-1 rounded-md bg-slate-800 text-xs text-slate-200">
            counter.js
          </div>
          <div className="flex items-center gap-1 px-2.5 py-1 rounded-md text-xs text-slate-500 hover:text-slate-300 hover:bg-slate-800/50 transition-colors cursor-pointer">
            greeter.js
          </div>
        </div>
        <div className="flex-1 p-4 font-mono text-xs text-slate-400 overflow-auto">
          <div><span className="text-blue-400">definePlugin</span>{"(({ ui }) => {"}</div>
          <div className="pl-4">{"return {"}</div>
          <div className="pl-8"><span className="text-slate-500">id</span>{': '}<span className="text-emerald-400">{'"counter"'}</span>,</div>
          <div className="pl-8"><span className="text-slate-500">title</span>{': '}<span className="text-emerald-400">{'"Counter"'}</span>,</div>
          <div className="pl-8">...</div>
          <div className="pl-4">{"};"}</div>
          <div>{"});"}</div>
        </div>
      </div>
      {/* Preview side */}
      <div className="flex-[4] p-4 overflow-auto">
        <div className="text-xs text-slate-500 mb-3 uppercase tracking-wider font-medium">Live Preview</div>
        <div className="rounded-lg border border-white/[0.08] bg-slate-900/50 p-4">
          <div className="flex items-center gap-2 mb-3">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-500" />
            <span className="text-sm font-medium text-slate-200">Counter</span>
            <span className="text-xs font-mono text-slate-600">abc-1234</span>
          </div>
          <div className="text-sm text-slate-300 mb-3">Counter: 5</div>
          <div className="flex gap-2">
            {["−", "Reset", "+"].map((label) => (
              <button key={label} className="px-3 py-1 text-xs rounded-md border border-white/[0.08] text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors">
                {label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function DevtoolsPlaceholder() {
  return (
    <div className="h-full flex flex-col text-xs">
      <div className="flex items-center h-9 px-2 border-b border-white/[0.06] flex-shrink-0 gap-1">
        {["Timeline", "State", "Capabilities", "Errors", "Shared", "Docs"].map((tab, i) => (
          <button
            key={tab}
            className={`px-2.5 py-1 rounded-md transition-colors ${
              i === 0 ? "bg-slate-800 text-slate-200" : "text-slate-500 hover:text-slate-300 hover:bg-slate-800/50"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>
      <div className="flex-1 min-h-0 overflow-auto p-3 font-mono">
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
    </div>
  );
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

const meta = {
  title: "Workbench/WorkbenchLayout",
  component: WorkbenchLayout,
  parameters: { layout: "fullscreen" },
  decorators: [withStore()],
  argTypes: {
    sidebarCollapsed: { control: "boolean" },
    devtoolsCollapsed: { control: "boolean" },
    unstyled: { control: "boolean" },
  },
} satisfies Meta<typeof WorkbenchLayout>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    toolbar: <ToolbarPlaceholder />,
    sidebar: <Sidebar catalog={CATALOG} running={RUNNING} focusedInstanceId="abc-1234-5678" />,
    children: <MainPlaceholder />,
    devtools: <DevtoolsPlaceholder />,
  },
};

export const SidebarCollapsed: Story = {
  decorators: [withStore({ sidebarCollapsed: true })],
  args: {
    ...Default.args,
    sidebar: <Sidebar catalog={CATALOG} running={RUNNING} collapsed />,
  },
};

export const DevtoolsCollapsed: Story = {
  decorators: [withStore({ devtoolsCollapsed: true })],
  args: Default.args,
};

export const NoSidebar: Story = {
  args: {
    toolbar: <ToolbarPlaceholder />,
    children: <MainPlaceholder />,
    devtools: <DevtoolsPlaceholder />,
  },
};

export const NoDevtools: Story = {
  args: {
    toolbar: <ToolbarPlaceholder />,
    sidebar: <Sidebar catalog={CATALOG} running={RUNNING} />,
    children: <MainPlaceholder />,
  },
};

export const Minimal: Story = {
  args: {
    children: (
      <div className="flex items-center justify-center h-full text-sm text-slate-500">
        Main content only — no sidebar, toolbar, or devtools.
      </div>
    ),
  },
};

export const Unstyled: Story = {
  args: {
    ...Default.args,
    unstyled: true,
  },
};
