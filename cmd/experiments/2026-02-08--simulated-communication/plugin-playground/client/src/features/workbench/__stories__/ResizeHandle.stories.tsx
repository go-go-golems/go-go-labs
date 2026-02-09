import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { ResizeHandle } from "../components/ResizeHandle";

function ResizeDemo() {
  const [height, setHeight] = React.useState(200);
  return (
    <div className="bg-slate-900 rounded-lg overflow-hidden border border-white/[0.06]" style={{ height: 500 }}>
      <div className="flex flex-col h-full">
        {/* Main area */}
        <div className="flex-1 min-h-0 flex items-center justify-center text-sm text-slate-500">
          Main content — drag the handle below to resize the panel
        </div>

        {/* Resize handle */}
        <ResizeHandle
          size={height}
          onResize={setHeight}
          minSize={80}
          maxSize={400}
        />

        {/* Resizable panel */}
        <div
          className="flex-shrink-0 bg-slate-800/50 flex items-center justify-center text-xs text-slate-400 font-mono"
          style={{ height }}
        >
          Panel height: {height}px
        </div>
      </div>
    </div>
  );
}

const meta: Meta = {
  title: "Workbench/ResizeHandle",
};

export default meta;
type Story = StoryObj;

export const Interactive: Story = {
  render: () => <ResizeDemo />,
};
