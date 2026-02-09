import React from "react";
import { cn } from "@/lib/utils";
import { ChevronUp, ChevronDown } from "lucide-react";
import type { DevToolsTab } from "@/store/workbenchSlice";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface DevToolsTabDef {
  id: DevToolsTab;
  label: string;
}

export const DEVTOOLS_TABS: DevToolsTabDef[] = [
  { id: "timeline", label: "Timeline" },
  { id: "state", label: "State" },
  { id: "capabilities", label: "Capabilities" },
  { id: "errors", label: "Errors" },
  { id: "shared", label: "Shared" },
  { id: "docs", label: "Docs" },
];

export interface DevToolsPanelProps {
  activeTab: DevToolsTab;
  collapsed?: boolean;
  /** Error badge count shown on the Errors tab. */
  errorCount?: number;
  onSelectTab?: (tab: DevToolsTab) => void;
  onToggleCollapse?: () => void;
  /** Tab content — rendered below the tab bar. */
  children?: React.ReactNode;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function DevToolsPanel({
  activeTab,
  collapsed = false,
  errorCount = 0,
  onSelectTab,
  onToggleCollapse,
  children,
  className,
}: DevToolsPanelProps) {
  return (
    <div
      data-part="devtools-panel"
      data-state={collapsed ? "collapsed" : "expanded"}
      className={cn("flex flex-col h-full", className)}
    >
      {/* Tab bar */}
      <div
        data-part="devtools-tabbar"
        className="flex items-center h-9 px-2 border-b border-white/[0.06] flex-shrink-0 gap-0.5"
      >
        {/* Collapse toggle */}
        <button
          data-part="devtools-toggle"
          onClick={onToggleCollapse}
          className="p-1 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors mr-1"
          title={collapsed ? "Expand devtools" : "Collapse devtools"}
        >
          {collapsed ? (
            <ChevronUp className="w-3.5 h-3.5" />
          ) : (
            <ChevronDown className="w-3.5 h-3.5" />
          )}
        </button>

        {/* Tabs */}
        {DEVTOOLS_TABS.map((tab) => {
          const active = tab.id === activeTab;
          return (
            <button
              key={tab.id}
              data-part="devtools-tab"
              data-state={active ? "active" : undefined}
              data-tab={tab.id}
              onClick={() => onSelectTab?.(tab.id)}
              className={cn(
                "flex items-center gap-1 px-2.5 py-1 rounded-md text-xs transition-colors",
                active
                  ? "bg-slate-800 text-slate-200"
                  : "text-slate-500 hover:text-slate-300 hover:bg-slate-800/50",
              )}
            >
              {tab.label}
              {tab.id === "errors" && errorCount > 0 && (
                <span
                  data-part="devtools-tab-badge"
                  className="ml-0.5 px-1.5 py-0 rounded-full text-[10px] font-semibold tabular-nums bg-red-500/20 text-red-400"
                >
                  {errorCount}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Content area — hidden when collapsed */}
      {!collapsed && (
        <div data-part="devtools-content" className="flex-1 min-h-0 overflow-hidden">
          {children}
        </div>
      )}
    </div>
  );
}
