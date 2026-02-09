import type { Meta, StoryObj } from "@storybook/react-vite";
import { TimelinePanel, type TimelineEntry } from "../components/TimelinePanel";

const now = Date.now();

const ENTRIES: TimelineEntry[] = [
  { id: "d1", timestamp: now - 1000, scope: "plugin", outcome: "applied", actionType: "increment", instanceId: "abc-1234-5678", shortInstanceId: "abc-1234" },
  { id: "d2", timestamp: now - 1000, scope: "shared", outcome: "applied", actionType: "set-instance", instanceId: "abc-1234-5678", shortInstanceId: "abc-1234", domain: "counter-summary" },
  { id: "d3", timestamp: now - 2200, scope: "shared", outcome: "denied", actionType: "set-name", instanceId: "def-5678-9012", shortInstanceId: "def-5678", domain: "greeter-profile", reason: "No write grant" },
  { id: "d4", timestamp: now - 3500, scope: "plugin", outcome: "applied", actionType: "reset", instanceId: "abc-1234-5678", shortInstanceId: "abc-1234" },
  { id: "d5", timestamp: now - 5000, scope: "shared", outcome: "applied", actionType: "set-name", instanceId: "ghi-9012-3456", shortInstanceId: "ghi-9012", domain: "greeter-profile" },
  { id: "d6", timestamp: now - 7000, scope: "plugin", outcome: "ignored", actionType: "unknown-action", instanceId: "def-5678-9012", shortInstanceId: "def-5678" },
];

function Wrapper(props: React.ComponentProps<typeof TimelinePanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 260 }}>
      <TimelinePanel {...props} />
    </div>
  );
}

const meta = {
  title: "Workbench/Panels/TimelinePanel",
  component: TimelinePanel,
  render: (args) => <Wrapper {...args} />,
} satisfies Meta<typeof TimelinePanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { entries: ENTRIES },
};

export const WithFocus: Story = {
  args: { entries: ENTRIES, focusedInstanceId: "abc-1234-5678" },
};

export const Empty: Story = {
  args: { entries: [] },
};

export const ManyEntries: Story = {
  args: {
    entries: Array.from({ length: 50 }, (_, i) => ({
      id: `d${i}`,
      timestamp: now - i * 200,
      scope: (i % 3 === 0 ? "shared" : "plugin") as "shared" | "plugin",
      outcome: (["applied", "applied", "denied", "ignored"] as const)[i % 4],
      actionType: ["increment", "set-instance", "set-name", "reset"][i % 4],
      instanceId: `inst-${i % 3}`,
      shortInstanceId: `inst-${i % 3}`,
      domain: i % 3 === 0 ? "counter-summary" : undefined,
    })),
  },
};
