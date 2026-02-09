import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { StatePanel, type InstanceState } from "../components/StatePanel";

const INSTANCES: InstanceState[] = [
  { instanceId: "abc-1234", title: "Counter", shortId: "abc-1234", state: { count: 5 } },
  { instanceId: "def-5678", title: "Greeter", shortId: "def-5678", state: { name: "Alice", greeting: "Hello, Alice!" } },
  {
    instanceId: "ghi-9012",
    title: "Status Dashboard",
    shortId: "ghi-9012",
    state: {
      metrics: { totalPlugins: 3, totalDispatches: 47, uptime: "12m 34s" },
      registry: [
        { id: "abc-1234", title: "Counter", status: "loaded" },
        { id: "def-5678", title: "Greeter", status: "loaded" },
      ],
    },
  },
];

function Wrapper(props: React.ComponentProps<typeof StatePanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 280 }}>
      <StatePanel {...props} />
    </div>
  );
}

const meta = {
  title: "Workbench/Panels/StatePanel",
  component: StatePanel,
  render: (args) => <Wrapper {...args} />,
  args: { onFocusInstance: fn() },
} satisfies Meta<typeof StatePanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { instances: INSTANCES },
};

export const Focused: Story = {
  args: { instances: INSTANCES, focusedInstanceId: "abc-1234" },
};

export const Empty: Story = {
  args: { instances: [] },
};
