import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { EditorTabBar, type EditorTabInfo } from "../components/EditorTabBar";

const TABS: EditorTabInfo[] = [
  { id: "t1", label: "counter.js", dirty: false },
  { id: "t2", label: "greeter.js", dirty: true },
  { id: "t3", label: "status-dashboard.js", dirty: false },
];

function Wrapper(props: React.ComponentProps<typeof EditorTabBar>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]">
      <EditorTabBar {...props} />
    </div>
  );
}

const meta = {
  title: "Workbench/EditorTabBar",
  component: EditorTabBar,
  render: (args) => <Wrapper {...args} />,
  args: {
    onSelectTab: fn(),
    onCloseTab: fn(),
    onRun: fn(),
    onReload: fn(),
    running: false,
  },
} satisfies Meta<typeof EditorTabBar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    tabs: TABS,
    activeTabId: "t1",
  },
};

export const DirtyActive: Story = {
  args: {
    tabs: TABS,
    activeTabId: "t2",
  },
};

export const SingleTab: Story = {
  args: {
    tabs: [{ id: "t1", label: "counter.js", dirty: false }],
    activeTabId: "t1",
  },
};

export const Empty: Story = {
  args: {
    tabs: [],
    activeTabId: null,
  },
};

export const Running: Story = {
  args: {
    tabs: TABS,
    activeTabId: "t1",
    running: true,
  },
};

export const ManyTabs: Story = {
  args: {
    tabs: Array.from({ length: 10 }, (_, i) => ({
      id: `t${i}`,
      label: `plugin-${i}.js`,
      dirty: i % 3 === 0,
    })),
    activeTabId: "t3",
  },
};
