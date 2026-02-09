import React from "react";
import { cn } from "@/lib/utils";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface InstanceState {
  instanceId: string;
  title: string;
  shortId: string;
  state: unknown;
}

export interface StatePanelProps {
  instances: InstanceState[];
  focusedInstanceId?: string | null;
  onFocusInstance?: (instanceId: string) => void;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function StatePanel({
  instances,
  focusedInstanceId,
  onFocusInstance,
  className,
}: StatePanelProps) {
  // If an instance is focused, show only that one; otherwise show all
  const displayed = focusedInstanceId
    ? instances.filter((i) => i.instanceId === focusedInstanceId)
    : instances;

  return (
    <div data-part="state-panel" className={cn("flex h-full", className)}>
      {/* Instance list (left rail) */}
      <div
        data-part="state-instances"
        className="w-40 flex-shrink-0 border-r border-white/[0.06] overflow-y-auto py-1"
      >
        {instances.length === 0 ? (
          <div className="px-3 py-2 text-xs text-slate-600">No instances</div>
        ) : (
          instances.map((inst) => {
            const active = inst.instanceId === focusedInstanceId;
            return (
              <button
                key={inst.instanceId}
                data-part="state-instance-item"
                data-state={active ? "active" : undefined}
                onClick={() => onFocusInstance?.(inst.instanceId)}
                className={cn(
                  "w-full text-left px-3 py-1.5 text-xs transition-colors truncate",
                  active
                    ? "bg-slate-800 text-slate-200"
                    : "text-slate-500 hover:text-slate-300 hover:bg-slate-800/50",
                )}
              >
                <div className="truncate">{inst.title}</div>
                <div className="font-mono text-[10px] text-slate-600 truncate">{inst.shortId}</div>
              </button>
            );
          })
        )}
      </div>

      {/* JSON viewer (right) */}
      <div
        data-part="state-viewer"
        className="flex-1 min-w-0 overflow-auto p-3"
      >
        {displayed.length === 0 ? (
          <div className="flex items-center justify-center h-full text-xs text-slate-600">
            {instances.length === 0 ? "Load a plugin to view its state." : "Select an instance."}
          </div>
        ) : (
          displayed.map((inst) => (
            <div key={inst.instanceId} className="mb-4 last:mb-0">
              <div className="flex items-center gap-2 mb-1">
                <span className="text-xs font-medium text-slate-400">{inst.title}</span>
                <span className="text-[10px] font-mono text-slate-600">{inst.shortId}</span>
              </div>
              <pre
                data-part="state-json"
                className="text-xs font-mono text-slate-300 bg-slate-950/50 rounded-md p-3 overflow-auto whitespace-pre-wrap"
              >
                {JSON.stringify(inst.state, null, 2)}
              </pre>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
