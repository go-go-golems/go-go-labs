import type { Meta, StoryObj } from "@storybook/react-vite";

function SmokeTest() {
  return (
    <div className="p-6 font-mono">
      <h1 className="text-2xl font-bold text-cyan-400 mb-4">PLUGIN WORKBENCH</h1>
      <p className="text-foreground text-sm mb-4">
        Storybook smoke test — if you see this with the dark brutalist theme, it works.
      </p>
      <div className="flex gap-2">
        <span className="px-2 py-1 text-xs border border-cyan-400/30 rounded-sm text-cyan-300">
          TAILWIND ✓
        </span>
        <span className="px-2 py-1 text-xs border border-cyan-400/30 rounded-sm text-cyan-300">
          THEME ✓
        </span>
        <span className="px-2 py-1 text-xs border border-cyan-400/30 rounded-sm text-cyan-300">
          FONTS ✓
        </span>
      </div>
    </div>
  );
}

const meta = {
  title: "Workbench/SmokeTest",
  component: SmokeTest,
} satisfies Meta<typeof SmokeTest>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
