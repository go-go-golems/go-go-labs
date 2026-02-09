import type { Meta, StoryObj } from "@storybook/react-vite";
import { CapabilitiesPanel, type InstanceCapabilities } from "../components/CapabilitiesPanel";

const DOMAINS = ["counter-summary", "greeter-profile", "runtime-registry", "runtime-metrics"];

const INSTANCES: InstanceCapabilities[] = [
  {
    instanceId: "abc-1234", title: "Counter", shortId: "abc-1234",
    grants: {
      "counter-summary": { read: true, write: true },
      "greeter-profile": { read: false, write: false },
      "runtime-registry": { read: false, write: false },
      "runtime-metrics": { read: false, write: false },
    },
  },
  {
    instanceId: "def-5678", title: "Greeter", shortId: "def-5678",
    grants: {
      "counter-summary": { read: false, write: false },
      "greeter-profile": { read: true, write: true },
      "runtime-registry": { read: false, write: false },
      "runtime-metrics": { read: false, write: false },
    },
  },
  {
    instanceId: "ghi-9012", title: "Status Dashboard", shortId: "ghi-9012",
    grants: {
      "counter-summary": { read: true, write: false },
      "greeter-profile": { read: false, write: false },
      "runtime-registry": { read: true, write: false },
      "runtime-metrics": { read: true, write: false },
    },
  },
];

function Wrapper(props: React.ComponentProps<typeof CapabilitiesPanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 240 }}>
      <CapabilitiesPanel {...props} />
    </div>
  );
}

const meta = {
  title: "Workbench/Panels/CapabilitiesPanel",
  component: CapabilitiesPanel,
  render: (args) => <Wrapper {...args} />,
} satisfies Meta<typeof CapabilitiesPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { domains: DOMAINS, instances: INSTANCES },
};

export const WithFocus: Story = {
  args: { domains: DOMAINS, instances: INSTANCES, focusedInstanceId: "ghi-9012" },
};

export const Empty: Story = {
  args: { domains: DOMAINS, instances: [] },
};
