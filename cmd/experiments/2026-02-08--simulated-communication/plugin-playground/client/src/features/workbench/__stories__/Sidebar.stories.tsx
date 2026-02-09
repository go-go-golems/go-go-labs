import type { Meta, StoryObj } from "@storybook/react-vite";
import { Sidebar, type CatalogEntry, type RunningInstance } from "../components/Sidebar";
import { fn } from "storybook/test";

// ---------------------------------------------------------------------------
// Fixture data
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

const RUNNING_WITH_ERROR: RunningInstance[] = [
  ...RUNNING,
  {
    instanceId: "err-0000-0000",
    title: "Broken Plugin",
    packageId: "custom",
    shortId: "err-0000",
    status: "error",
    readGrants: [],
    writeGrants: [],
  },
];

// ---------------------------------------------------------------------------
// Wrapper to provide height context (Sidebar fills parent height)
// ---------------------------------------------------------------------------

function SidebarWrapper(props: React.ComponentProps<typeof Sidebar>) {
  return (
    <div
      style={{ width: props.collapsed ? 48 : 240, height: 600 }}
      className="border border-white/[0.06] rounded-lg overflow-hidden bg-slate-900"
    >
      <Sidebar {...props} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

const meta = {
  title: "Workbench/Sidebar",
  component: Sidebar,
  render: (args) => <SidebarWrapper {...args} />,
  args: {
    onToggleCollapse: fn(),
    onLoadPreset: fn(),
    onFocusInstance: fn(),
    onUnloadInstance: fn(),
    onNewPlugin: fn(),
  },
} satisfies Meta<typeof Sidebar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    catalog: CATALOG,
    running: RUNNING,
  },
};

export const WithFocusedInstance: Story = {
  args: {
    catalog: CATALOG,
    running: RUNNING,
    focusedInstanceId: "def-5678-9012",
  },
};

export const WithError: Story = {
  args: {
    catalog: CATALOG,
    running: RUNNING_WITH_ERROR,
  },
};

export const Empty: Story = {
  args: {
    catalog: CATALOG,
    running: [],
  },
};

export const Collapsed: Story = {
  args: {
    catalog: CATALOG,
    running: RUNNING,
    collapsed: true,
  },
};

export const ManyInstances: Story = {
  args: {
    catalog: CATALOG,
    running: Array.from({ length: 8 }, (_, i) => ({
      instanceId: `inst-${i}-${i}${i}${i}${i}`,
      title: CATALOG[i % CATALOG.length].title,
      packageId: CATALOG[i % CATALOG.length].id,
      shortId: `inst-${i}`,
      status: "loaded" as const,
      readGrants: i % 2 === 0 ? ["counter-summary"] : [],
      writeGrants: i % 3 === 0 ? ["counter-summary"] : [],
    })),
  },
};
