import React from "react";
import { cn } from "@/lib/utils";
import { Trash2 } from "lucide-react";
import type { ErrorEntry } from "@/store/workbenchSlice";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface ErrorsPanelProps {
  errors: ErrorEntry[];
  onClear?: () => void;
  className?: string;
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function formatTime(ts: number): string {
  return new Date(ts).toLocaleTimeString("en-US", { hour12: false, fractionalSecondDigits: 1 });
}

const KIND_STYLES: Record<ErrorEntry["kind"], string> = {
  load: "text-red-400",
  render: "text-amber-400",
  event: "text-orange-400",
};

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function ErrorsPanel({
  errors,
  onClear,
  className,
}: ErrorsPanelProps) {
  return (
    <div data-part="errors-panel" className={cn("flex flex-col h-full", className)}>
      {/* Header bar */}
      <div
        data-part="errors-header"
        className="flex items-center justify-between px-3 py-1.5 border-b border-white/[0.06] flex-shrink-0"
      >
        <span className="text-[10px] text-slate-500 tabular-nums">
          {errors.length} error{errors.length !== 1 ? "s" : ""}
        </span>
        {onClear && errors.length > 0 && (
          <button
            data-part="errors-clear"
            onClick={onClear}
            className="flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] text-slate-500 hover:text-red-400 hover:bg-slate-800/50 transition-colors"
          >
            <Trash2 className="w-3 h-3" />
            Clear
          </button>
        )}
      </div>

      {/* Error list */}
      <div data-part="errors-body" className="flex-1 min-h-0 overflow-y-auto px-3 py-1">
        {errors.length === 0 ? (
          <div className="flex items-center justify-center h-full text-xs text-slate-600">
            No errors — all clear.
          </div>
        ) : (
          <div className="space-y-1">
            {errors.map((err) => (
              <div
                key={err.id}
                data-part="error-entry"
                className="py-1 font-mono text-xs group"
              >
                <div className="flex items-baseline gap-2">
                  <span className="text-slate-600 flex-shrink-0">[{formatTime(err.timestamp)}]</span>
                  <span className={cn("flex-shrink-0 uppercase text-[10px] font-semibold", KIND_STYLES[err.kind])}>
                    {err.kind}
                  </span>
                  {err.instanceId && (
                    <span className="text-slate-600 flex-shrink-0">{err.instanceId.slice(0, 10)}</span>
                  )}
                </div>
                <div className="text-red-300 whitespace-pre-wrap pl-4 mt-0.5">{err.message}</div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
