import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { TopToolbar } from "../components/TopToolbar";

// ---------------------------------------------------------------------------
// Wrapper — provide dark background matching the toolbar surface
// ---------------------------------------------------------------------------

function ToolbarWrapper(props: React.ComponentProps<typeof TopToolbar>) {
  return (
    <div className="bg-slate-900 border-b border-white/[0.06] rounded-lg overflow-hidden">
      <TopToolbar {...props} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

const meta = {
  title: "Workbench/TopToolbar",
  component: TopToolbar,
  render: (args) => <ToolbarWrapper {...args} />,
  args: {
    onHealthClick: fn(),
    onMenuClick: fn(),
  },
  argTypes: {
    health: { control: "select", options: ["healthy", "degraded", "error"] },
    pluginCount: { control: { type: "number", min: 0, max: 20 } },
    dispatchCount: { control: { type: "number", min: 0, max: 9999 } },
    errorCount: { control: { type: "number", min: 0, max: 100 } },
  },
} satisfies Meta<typeof TopToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    pluginCount: 3,
    dispatchCount: 47,
    health: "healthy",
    errorCount: 0,
  },
};

export const WithErrors: Story = {
  args: {
    pluginCount: 4,
    dispatchCount: 123,
    health: "degraded",
    errorCount: 3,
  },
};

export const ErrorState: Story = {
  args: {
    pluginCount: 2,
    dispatchCount: 89,
    health: "error",
    errorCount: 12,
  },
};

export const Empty: Story = {
  args: {
    pluginCount: 0,
    dispatchCount: 0,
    health: "healthy",
    errorCount: 0,
  },
};

export const HighActivity: Story = {
  args: {
    pluginCount: 8,
    dispatchCount: 4217,
    health: "healthy",
    errorCount: 0,
  },
};

export const NoMenu: Story = {
  args: {
    pluginCount: 3,
    dispatchCount: 47,
    health: "healthy",
    errorCount: 0,
    onMenuClick: undefined,
  },
};

export const Unstyled: Story = {
  args: {
    pluginCount: 3,
    dispatchCount: 47,
    health: "healthy",
    errorCount: 1,
    unstyled: true,
  },
};
