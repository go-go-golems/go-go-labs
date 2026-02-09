import React from "react";
import { cn } from "@/lib/utils";
import { X } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface InstanceCardProps {
  instanceId: string;
  title: string;
  shortId: string;
  status: "loaded" | "error";
  focused?: boolean;
  /** The rendered UI tree (from WidgetRenderer). */
  children?: React.ReactNode;
  /** Error message when status is "error". */
  errorMessage?: string;
  onFocus?: () => void;
  onUnload?: () => void;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function InstanceCard({
  instanceId,
  title,
  shortId,
  status,
  focused = false,
  children,
  errorMessage,
  onFocus,
  onUnload,
  className,
}: InstanceCardProps) {
  return (
    <div
      data-part="instance-card"
      data-state={focused ? "focused" : undefined}
      data-instance-id={instanceId}
      onClick={onFocus}
      className={cn(
        "rounded-lg border transition-colors cursor-pointer",
        focused
          ? "border-blue-500/25 bg-slate-900/60"
          : "border-white/[0.08] bg-slate-900/40 hover:border-white/[0.12]",
        className,
      )}
    >
      {/* Header */}
      <div
        data-part="instance-card-header"
        className="flex items-center justify-between px-3 py-2 border-b border-white/[0.06]"
      >
        <div className="flex items-center gap-2 min-w-0">
          <span
            data-part="status-dot"
            className={cn(
              "w-1.5 h-1.5 rounded-full flex-shrink-0",
              status === "loaded" ? "bg-emerald-500" : "bg-red-500",
            )}
          />
          <span className="text-sm font-medium text-slate-200 truncate">{title}</span>
          <span className="text-xs font-mono text-slate-600 flex-shrink-0">{shortId}</span>
        </div>
        {onUnload && (
          <button
            data-part="instance-card-close"
            onClick={(e) => {
              e.stopPropagation();
              onUnload();
            }}
            className="p-0.5 rounded text-slate-600 hover:text-red-400 hover:bg-slate-800 transition-colors"
            title="Unload instance"
          >
            <X className="w-3.5 h-3.5" />
          </button>
        )}
      </div>

      {/* Body — widget rendering area or error */}
      <div data-part="instance-card-body" className="p-3 min-h-[3rem]">
        {status === "error" && errorMessage ? (
          <div data-part="instance-card-error" className="text-xs font-mono text-red-400 whitespace-pre-wrap">
            {errorMessage}
          </div>
        ) : children ? (
          children
        ) : (
          <div className="text-xs text-slate-600 italic">No UI rendered</div>
        )}
      </div>
    </div>
  );
}
