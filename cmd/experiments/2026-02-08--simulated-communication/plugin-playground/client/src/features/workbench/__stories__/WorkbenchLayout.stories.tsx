import type { Meta, StoryObj } from "@storybook/react-vite";
import { WorkbenchLayout } from "../components/WorkbenchLayout";

// ---------------------------------------------------------------------------
// Placeholder slot contents (realistic sizes for layout validation)
// ---------------------------------------------------------------------------

function ToolbarPlaceholder() {
  return (
    <div className="flex items-center gap-4 px-4 h-10 font-mono text-sm">
      <span className="font-bold text-cyan-400">▣ PLUGIN WORKBENCH</span>
      <span className="text-xs border border-cyan-400/30 rounded-sm px-2 py-0.5 text-cyan-300">
        ■ 3 plugins
      </span>
      <span className="text-xs border border-cyan-400/30 rounded-sm px-2 py-0.5 text-cyan-300">
        ↻ 47 dispatches
      </span>
      <span className="text-xs border border-cyan-400/30 rounded-sm px-2 py-0.5 text-emerald-400">
        ⚡ healthy
      </span>
    </div>
  );
}

function SidebarPlaceholder() {
  return (
    <div className="p-3 font-mono text-xs space-y-4">
      <div>
        <div className="text-cyan-400 font-bold mb-2">📦 CATALOG</div>
        {["Counter", "Calculator", "Status Dashboard", "Greeter", "Greeter Shared", "Runtime Monitor"].map(
          (name) => (
            <div key={name} className="py-1 px-2 rounded-sm hover:bg-card/50 text-foreground cursor-pointer">
              {name}
            </div>
          )
        )}
      </div>
      <div className="border-t border-border pt-3">
        <div className="text-cyan-400 font-bold mb-2">🔌 RUNNING (2)</div>
        <div className="space-y-2">
          <div className="py-1 px-2 rounded-sm bg-card/30">
            <div>● Counter <span className="text-muted-foreground">abc-12</span></div>
            <div className="text-[10px] text-muted-foreground mt-0.5">R/W: ctr-sum</div>
          </div>
          <div className="py-1 px-2 rounded-sm bg-card/30">
            <div>● Greeter <span className="text-muted-foreground">def-56</span></div>
            <div className="text-[10px] text-muted-foreground mt-0.5">R/W: grt-prof</div>
          </div>
        </div>
      </div>
    </div>
  );
}

function MainPlaceholder() {
  return (
    <div className="h-full flex font-mono text-xs">
      {/* Editor area */}
      <div className="flex-[6] border-r border-border p-4 overflow-auto">
        <div className="text-muted-foreground mb-2">CODE EDITOR</div>
        <pre className="text-foreground">
          {`definePlugin(({ ui }) => {\n  return {\n    id: "counter",\n    title: "Counter",\n    ...\n  };\n});`}
        </pre>
      </div>
      {/* Preview area */}
      <div className="flex-[4] p-4 overflow-auto">
        <div className="text-muted-foreground mb-2">LIVE PREVIEW</div>
        <div className="border border-cyan-400/20 rounded-sm p-3 bg-card/30">
          <div className="text-cyan-400 font-bold text-sm mb-2">COUNTER [abc-12]</div>
          <div className="mb-2">Counter: 5</div>
          <div className="flex gap-2">
            <button className="px-2 py-1 border border-cyan-400/30 rounded-sm text-cyan-300">−</button>
            <button className="px-2 py-1 border border-cyan-400/30 rounded-sm text-cyan-300">Reset</button>
            <button className="px-2 py-1 border border-cyan-400/30 rounded-sm text-cyan-300">+</button>
          </div>
        </div>
      </div>
    </div>
  );
}

function DevtoolsPlaceholder() {
  return (
    <div className="h-full flex flex-col font-mono text-xs">
      {/* Tab bar */}
      <div className="flex items-center gap-1 px-2 h-9 border-b border-border flex-shrink-0">
        {["Timeline", "State", "Capabilities", "Errors", "Shared", "📖 Docs"].map((tab, i) => (
          <button
            key={tab}
            className={`px-2 py-1 rounded-sm ${
              i === 0 ? "border border-cyan-400 text-cyan-300" : "text-muted-foreground border border-transparent"
            }`}
          >
            {tab}
          </button>
        ))}
      </div>
      {/* Content */}
      <div className="flex-1 min-h-0 overflow-auto p-3">
        <div className="space-y-1">
          <div className="flex gap-4">
            <span className="text-muted-foreground w-20">09:01:23.4</span>
            <span className="text-cyan-300 w-14">plugin</span>
            <span className="text-emerald-400 w-16">✅ applied</span>
            <span className="text-foreground">increment</span>
            <span className="text-muted-foreground">abc-123</span>
          </div>
          <div className="flex gap-4">
            <span className="text-muted-foreground w-20">09:01:23.4</span>
            <span className="text-cyan-300 w-14">shared</span>
            <span className="text-emerald-400 w-16">✅ applied</span>
            <span className="text-foreground">set-instance</span>
            <span className="text-muted-foreground">abc-123</span>
          </div>
          <div className="flex gap-4">
            <span className="text-muted-foreground w-20">09:01:22.1</span>
            <span className="text-cyan-300 w-14">shared</span>
            <span className="text-red-400 w-16">🚫 denied</span>
            <span className="text-foreground">set-name</span>
            <span className="text-muted-foreground">def-456</span>
          </div>
        </div>
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
  parameters: {
    layout: "fullscreen",
  },
  argTypes: {
    sidebarCollapsed: { control: "boolean" },
    devtoolsCollapsed: { control: "boolean" },
  },
} satisfies Meta<typeof WorkbenchLayout>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    toolbar: <ToolbarPlaceholder />,
    sidebar: <SidebarPlaceholder />,
    children: <MainPlaceholder />,
    devtools: <DevtoolsPlaceholder />,
  },
};

export const SidebarCollapsed: Story = {
  args: {
    ...Default.args,
    sidebarCollapsed: true,
  },
};

export const DevtoolsCollapsed: Story = {
  args: {
    ...Default.args,
    devtoolsCollapsed: true,
  },
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
    sidebar: <SidebarPlaceholder />,
    children: <MainPlaceholder />,
  },
};

export const Minimal: Story = {
  args: {
    children: (
      <div className="flex items-center justify-center h-full text-muted-foreground font-mono">
        Main content only — no sidebar, toolbar, or devtools
      </div>
    ),
  },
};
