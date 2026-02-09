import type { Meta, StoryObj } from "@storybook/react-vite";
import { SharedDomainsPanel, type SharedDomainInfo } from "../components/SharedDomainsPanel";

const DOMAINS: SharedDomainInfo[] = [
  {
    name: "counter-summary",
    state: { valuesByInstance: { "abc-1234": 5 }, totalValue: 5, instanceCount: 1, lastUpdatedInstanceId: "abc-1234" },
    readers: [
      { instanceId: "abc-1234", title: "Counter", shortId: "abc-1234" },
      { instanceId: "ghi-9012", title: "Status Dashboard", shortId: "ghi-9012" },
    ],
    writers: [
      { instanceId: "abc-1234", title: "Counter", shortId: "abc-1234" },
    ],
  },
  {
    name: "greeter-profile",
    state: { name: "Alice", lastUpdatedInstanceId: "def-5678" },
    readers: [
      { instanceId: "def-5678", title: "Greeter", shortId: "def-5678" },
      { instanceId: "jkl-3456", title: "Greeter Shared State", shortId: "jkl-3456" },
    ],
    writers: [
      { instanceId: "def-5678", title: "Greeter", shortId: "def-5678" },
    ],
  },
  {
    name: "runtime-registry",
    state: [],
    readers: [
      { instanceId: "ghi-9012", title: "Status Dashboard", shortId: "ghi-9012" },
    ],
    writers: [],
  },
  {
    name: "runtime-metrics",
    state: { totalPlugins: 3, totalDispatches: 47 },
    readers: [
      { instanceId: "ghi-9012", title: "Status Dashboard", shortId: "ghi-9012" },
    ],
    writers: [],
  },
];

function Wrapper(props: React.ComponentProps<typeof SharedDomainsPanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 400 }}>
      <SharedDomainsPanel {...props} />
    </div>
  );
}

const meta = {
  title: "Workbench/Panels/SharedDomainsPanel",
  component: SharedDomainsPanel,
  render: (args) => <Wrapper {...args} />,
} satisfies Meta<typeof SharedDomainsPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { domains: DOMAINS },
};

export const SingleDomain: Story = {
  args: { domains: [DOMAINS[0]] },
};

export const Empty: Story = {
  args: { domains: [] },
};
