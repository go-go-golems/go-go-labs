// Widget Renderer — interprets data-only UI trees into React components.
// Visual style: vm-system-ui slate palette (slate borders, blue accent,
// Inter + JetBrains Mono fonts, compact sizing).

import React from "react";
import type { UINode, UIEventRef } from "@runtime/uiTypes";

interface WidgetRendererProps {
  tree: UINode | null;
  onEvent: (ref: UIEventRef, eventPayload?: any) => void;
}

export function WidgetRenderer({ tree, onEvent }: WidgetRendererProps) {
  if (!tree) return <div className="text-xs text-slate-600 font-mono">No widget tree</div>;

  return <>{renderNode(tree, onEvent)}</>;
}

function renderNode(node: UINode, onEvent: (ref: UIEventRef, eventPayload?: any) => void): React.ReactNode {
  if (!node) return null;

  switch (node.kind) {
    case "panel":
      return (
        <div className="border border-white/[0.08] rounded-lg p-3 mb-2 bg-slate-900/50">
          {(node.children ?? []).map((c, i) => (
            <React.Fragment key={i}>{renderNode(c, onEvent)}</React.Fragment>
          ))}
        </div>
      );

    case "row":
      return (
        <div className="flex gap-2 items-center mb-2">
          {(node.children ?? []).map((c, i) => (
            <React.Fragment key={i}>{renderNode(c, onEvent)}</React.Fragment>
          ))}
        </div>
      );

    case "column":
      return (
        <div className="flex flex-col gap-1.5">
          {(node.children ?? []).map((c, i) => (
            <React.Fragment key={i}>{renderNode(c, onEvent)}</React.Fragment>
          ))}
        </div>
      );

    case "text":
      return <div className="text-sm text-slate-300 mb-1">{node.text}</div>;

    case "badge":
      return (
        <span className="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-medium font-mono uppercase tracking-wider border border-blue-500/30 text-blue-400 bg-blue-500/10">
          {node.text}
        </span>
      );

    case "button": {
      const { label, onClick, variant } = node.props;
      return (
        <button
          onClick={() => onClick && onEvent(onClick, onClick.args)}
          className={
            variant === "destructive"
              ? "px-3 py-1.5 text-xs font-medium rounded-md border border-red-500/30 text-red-400 bg-red-500/10 hover:bg-red-500/20 transition-colors"
              : "px-3 py-1.5 text-xs font-medium rounded-md border border-white/[0.1] text-slate-300 hover:text-slate-100 hover:bg-slate-800/60 transition-colors"
          }
        >
          {label}
        </button>
      );
    }

    case "input": {
      const { value, placeholder, onChange } = node.props;
      return (
        <input
          type="text"
          value={value}
          placeholder={placeholder}
          onChange={(e) => onChange && onEvent(onChange, { value: e.target.value })}
          className="w-full px-2.5 py-1.5 text-xs font-mono bg-slate-800 border border-white/[0.08] rounded-md text-slate-300 placeholder:text-slate-600 outline-none focus:border-blue-500/40 focus:ring-1 focus:ring-blue-500/20 transition-colors"
        />
      );
    }

    case "counter": {
      const { value, onIncrement, onDecrement } = node.props;
      return (
        <div className="flex items-center gap-2 border border-white/[0.08] rounded-lg p-2 bg-slate-900/30">
          <button
            onClick={() => onDecrement && onEvent(onDecrement)}
            className="w-7 h-7 flex items-center justify-center text-sm rounded-md border border-white/[0.1] text-slate-400 hover:text-slate-200 hover:bg-slate-800/60 transition-colors"
          >
            −
          </button>
          <span className="font-mono text-base font-semibold text-blue-400 min-w-[3ch] text-center tabular-nums">
            {value}
          </span>
          <button
            onClick={() => onIncrement && onEvent(onIncrement)}
            className="w-7 h-7 flex items-center justify-center text-sm rounded-md border border-white/[0.1] text-slate-400 hover:text-slate-200 hover:bg-slate-800/60 transition-colors"
          >
            +
          </button>
        </div>
      );
    }

    case "table": {
      const { headers, rows } = node.props;
      return (
        <div className="border border-white/[0.08] rounded-lg overflow-hidden">
          <table className="w-full text-xs font-mono">
            <thead>
              <tr className="border-b border-white/[0.06] bg-slate-800/50">
                {headers.map((h: string, i: number) => (
                  <th key={i} className="text-left px-3 py-2 font-medium text-slate-400 uppercase tracking-wider text-[10px]">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {rows.map((row: any[], i: number) => (
                <tr key={i} className="border-b border-white/[0.04] hover:bg-slate-800/30 transition-colors">
                  {row.map((cell, j) => (
                    <td key={j} className="px-3 py-1.5 text-slate-300">
                      {String(cell)}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
    }

    default:
      return null;
  }
}
