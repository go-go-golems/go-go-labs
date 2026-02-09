import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { ErrorsPanel } from "../components/ErrorsPanel";
import type { ErrorEntry } from "@/store/workbenchSlice";

const now = Date.now();

const ERRORS: ErrorEntry[] = [
  {
    id: "e1", timestamp: now - 1000, kind: "load",
    instanceId: "err-0000-0000", widgetId: null,
    message: "SyntaxError: Unexpected token '}' at line 12",
  },
  {
    id: "e2", timestamp: now - 3000, kind: "render",
    instanceId: "abc-1234-5678", widgetId: "w-btn-1",
    message: "TypeError: Cannot read properties of undefined (reading 'render')\n    at WidgetRenderer.tsx:42:17",
  },
  {
    id: "e3", timestamp: now - 5000, kind: "event",
    instanceId: "def-5678-9012", widgetId: null,
    message: "ReferenceError: x is not defined\n    at plugin.js:8:3",
  },
];

function Wrapper(props: React.ComponentProps<typeof ErrorsPanel>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 260 }}>
      <ErrorsPanel {...props} />
    </div>
  );
}

const meta = {
  title: "Workbench/Panels/ErrorsPanel",
  component: ErrorsPanel,
  render: (args) => <Wrapper {...args} />,
  args: { onClear: fn() },
} satisfies Meta<typeof ErrorsPanel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: { errors: ERRORS },
};

export const Empty: Story = {
  args: { errors: [] },
};

export const SingleError: Story = {
  args: { errors: [ERRORS[0]] },
};

export const ManyErrors: Story = {
  args: {
    errors: Array.from({ length: 25 }, (_, i) => ({
      id: `e${i}`,
      timestamp: now - i * 500,
      kind: (["load", "render", "event"] as const)[i % 3],
      instanceId: `inst-${i % 4}`,
      widgetId: null,
      message: `Error #${i}: Something went wrong in ${["load", "render", "event"][i % 3]} phase`,
    })),
  },
};
