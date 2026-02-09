import React from "react";
import { cn } from "@/lib/utils";
import { Terminal, Layers, Activity, AlertTriangle, ChevronDown } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type HealthStatus = "healthy" | "degraded" | "error";

export interface TopToolbarProps {
  /** Number of currently loaded plugin instances. */
  pluginCount: number;
  /** Total dispatch events since session start. */
  dispatchCount: number;
  /** Overall runtime health. */
  health: HealthStatus;
  /** Number of errors in the error log. */
  errorCount: number;
  /** Fired when the user clicks the health badge (e.g. to open errors tab). */
  onHealthClick?: () => void;
  /** Fired when the user clicks the menu chevron. */
  onMenuClick?: () => void;
  /** Strip all styling; render only data-part attributes. */
  unstyled?: boolean;
  className?: string;
}

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

const HEALTH_STYLES: Record<HealthStatus, { dot: string; label: string }> = {
  healthy:  { dot: "bg-emerald-500", label: "healthy" },
  degraded: { dot: "bg-amber-500",   label: "degraded" },
  error:    { dot: "bg-red-500",     label: "error" },
};

function ToolbarDivider() {
  return <span data-part="toolbar-divider" className="text-slate-700 select-none">·</span>;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function TopToolbar({
  pluginCount,
  dispatchCount,
  health,
  errorCount,
  onHealthClick,
  onMenuClick,
  unstyled = false,
  className,
}: TopToolbarProps) {
  const hs = HEALTH_STYLES[health];

  if (unstyled) {
    return (
      <div data-part="toolbar-body" className={className}>
        <div data-part="toolbar-brand">
          <span data-part="toolbar-logo" />
          <span data-part="toolbar-title">Plugin Workbench</span>
        </div>
        <div data-part="toolbar-stats">
          <span data-part="toolbar-stat-plugins">{pluginCount} plugins</span>
          <span data-part="toolbar-stat-dispatches">{dispatchCount} dispatches</span>
          <span data-part="toolbar-stat-health" data-state={health}>{hs.label}</span>
          {errorCount > 0 && (
            <span data-part="toolbar-stat-errors">{errorCount} errors</span>
          )}
        </div>
      </div>
    );
  }

  return (
    <div
      data-part="toolbar-body"
      className={cn(
        "flex items-center justify-between px-4 h-10",
        className,
      )}
    >
      {/* Left: brand */}
      <div data-part="toolbar-brand" className="flex items-center gap-3">
        <div className="flex items-center gap-2">
          <div
            data-part="toolbar-logo"
            className="w-6 h-6 rounded-md bg-blue-600 flex items-center justify-center"
          >
            <Terminal className="w-3.5 h-3.5 text-white" />
          </div>
          <span
            data-part="toolbar-title"
            className="text-sm font-semibold text-slate-100 tracking-tight"
          >
            Plugin Workbench
          </span>
        </div>
      </div>

      {/* Right: stats */}
      <div data-part="toolbar-stats" className="flex items-center gap-3 text-xs text-slate-500">
        {/* Plugin count */}
        <span
          data-part="toolbar-stat-plugins"
          className="flex items-center gap-1.5"
        >
          <Layers className="w-3.5 h-3.5" />
          <span className="tabular-nums">{pluginCount}</span>
          <span className="hidden sm:inline">plugin{pluginCount !== 1 ? "s" : ""}</span>
        </span>

        <ToolbarDivider />

        {/* Dispatch count */}
        <span
          data-part="toolbar-stat-dispatches"
          className="flex items-center gap-1.5"
        >
          <Activity className="w-3.5 h-3.5" />
          <span className="tabular-nums">{dispatchCount}</span>
          <span className="hidden sm:inline">dispatch{dispatchCount !== 1 ? "es" : ""}</span>
        </span>

        <ToolbarDivider />

        {/* Health */}
        <button
          data-part="toolbar-stat-health"
          data-state={health}
          onClick={onHealthClick}
          className="flex items-center gap-1.5 hover:text-slate-300 transition-colors"
        >
          <span className={cn("w-1.5 h-1.5 rounded-full", hs.dot)} />
          <span>{hs.label}</span>
        </button>

        {/* Error count (only shown when > 0) */}
        {errorCount > 0 && (
          <>
            <ToolbarDivider />
            <button
              data-part="toolbar-stat-errors"
              onClick={onHealthClick}
              className="flex items-center gap-1.5 text-red-400 hover:text-red-300 transition-colors"
            >
              <AlertTriangle className="w-3.5 h-3.5" />
              <span className="tabular-nums">{errorCount}</span>
            </button>
          </>
        )}

        {/* Menu */}
        {onMenuClick && (
          <>
            <ToolbarDivider />
            <button
              data-part="toolbar-menu"
              onClick={onMenuClick}
              className="p-1 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors"
              title="Menu"
            >
              <ChevronDown className="w-3.5 h-3.5" />
            </button>
          </>
        )}
      </div>
    </div>
  );
}
