import type { Meta, StoryObj } from "@storybook/react-vite";
import { fn } from "storybook/test";
import { LivePreview } from "../components/LivePreview";
import { InstanceCard } from "../components/InstanceCard";

// ---------------------------------------------------------------------------
// Fake widget UIs for stories
// ---------------------------------------------------------------------------

function CounterWidget() {
  return (
    <div className="text-sm text-slate-300 space-y-3">
      <div>Counter: 5</div>
      <div className="flex gap-2">
        {["−", "Reset", "+"].map((label) => (
          <button key={label} className="px-3 py-1 text-xs rounded-md border border-white/[0.08] text-slate-400 hover:text-slate-200 hover:bg-slate-800/50 transition-colors">
            {label}
          </button>
        ))}
      </div>
    </div>
  );
}

function GreeterWidget() {
  return (
    <div className="text-sm text-slate-300 space-y-2">
      <div className="text-xs text-slate-500">Enter your name:</div>
      <input
        type="text"
        value="Alice"
        readOnly
        className="w-full px-2 py-1 text-xs bg-slate-800 border border-white/[0.08] rounded text-slate-300"
      />
      <div className="text-slate-400">Hello, <span className="text-slate-200">Alice</span>!</div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Wrapper
// ---------------------------------------------------------------------------

function Wrapper(props: React.ComponentProps<typeof LivePreview>) {
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 500, width: 400 }}>
      <LivePreview {...props} />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Stories
// ---------------------------------------------------------------------------

const meta = {
  title: "Workbench/LivePreview",
  component: LivePreview,
  render: (args) => <Wrapper {...args} />,
} satisfies Meta<typeof LivePreview>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    children: (
      <>
        <InstanceCard
          instanceId="abc-1234"
          title="Counter"
          shortId="abc-1234"
          status="loaded"
          focused
          onFocus={fn()}
          onUnload={fn()}
        >
          <CounterWidget />
        </InstanceCard>
        <InstanceCard
          instanceId="def-5678"
          title="Greeter"
          shortId="def-5678"
          status="loaded"
          onFocus={fn()}
          onUnload={fn()}
        >
          <GreeterWidget />
        </InstanceCard>
      </>
    ),
  },
};

export const SingleInstance: Story = {
  args: {
    children: (
      <InstanceCard
        instanceId="abc-1234"
        title="Counter"
        shortId="abc-1234"
        status="loaded"
        focused
        onFocus={fn()}
        onUnload={fn()}
      >
        <CounterWidget />
      </InstanceCard>
    ),
  },
};

export const WithError: Story = {
  args: {
    children: (
      <InstanceCard
        instanceId="err-0000"
        title="Broken Plugin"
        shortId="err-0000"
        status="error"
        errorMessage={"TypeError: Cannot read properties of undefined (reading 'render')\n    at eval (plugin.js:12:5)\n    at sandbox.ts:42:17"}
        onFocus={fn()}
        onUnload={fn()}
      />
    ),
  },
};

export const Empty: Story = {
  args: {},
};
