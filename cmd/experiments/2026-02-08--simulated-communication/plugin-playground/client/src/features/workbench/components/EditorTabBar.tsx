import React from "react";
import { cn } from "@/lib/utils";
import { X, Play, RotateCw } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface EditorTabInfo {
  id: string;
  label: string;
  dirty: boolean;
}

export interface EditorTabBarProps {
  tabs: EditorTabInfo[];
  activeTabId: string | null;
  onSelectTab?: (tabId: string) => void;
  onCloseTab?: (tabId: string) => void;
  onRun?: () => void;
  onReload?: () => void;
  /** Whether a plugin is currently loading/running. */
  running?: boolean;
  className?: string;
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function EditorTabBar({
  tabs,
  activeTabId,
  onSelectTab,
  onCloseTab,
  onRun,
  onReload,
  running = false,
  className,
}: EditorTabBarProps) {
  return (
    <div
      data-part="editor-tabbar"
      className={cn(
        "flex items-center h-9 px-1 border-b border-white/[0.06] flex-shrink-0 gap-0.5",
        className,
      )}
    >
      {/* Tabs */}
      <div data-part="editor-tabs" className="flex items-center gap-0.5 min-w-0 overflow-x-auto flex-1">
        {tabs.length === 0 && (
          <span className="px-2 text-xs text-slate-600 italic">No open tabs</span>
        )}
        {tabs.map((tab) => {
          const active = tab.id === activeTabId;
          return (
            <button
              key={tab.id}
              data-part="editor-tab"
              data-state={active ? "active" : undefined}
              onClick={() => onSelectTab?.(tab.id)}
              className={cn(
                "flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs transition-colors group max-w-[10rem]",
                active
                  ? "bg-slate-800 text-slate-200"
                  : "text-slate-500 hover:text-slate-300 hover:bg-slate-800/50",
              )}
            >
              <span className="truncate">{tab.label}</span>
              {tab.dirty && (
                <span
                  data-part="editor-tab-dirty"
                  className="w-1.5 h-1.5 rounded-full bg-blue-400 flex-shrink-0"
                  title="Unsaved changes"
                />
              )}
              {onCloseTab && (
                <span
                  role="button"
                  data-part="editor-tab-close"
                  onClick={(e) => {
                    e.stopPropagation();
                    onCloseTab(tab.id);
                  }}
                  className={cn(
                    "p-0.5 rounded flex-shrink-0 transition-colors",
                    "text-slate-600 hover:text-red-400 hover:bg-slate-700",
                    active ? "opacity-100" : "opacity-0 group-hover:opacity-100",
                  )}
                >
                  <X className="w-3 h-3" />
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Actions */}
      <div data-part="editor-actions" className="flex items-center gap-1 flex-shrink-0 ml-1">
        {onReload && (
          <button
            data-part="editor-reload"
            onClick={onReload}
            disabled={running}
            className="p-1.5 rounded-md text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
            title="Reload plugin"
          >
            <RotateCw className="w-3.5 h-3.5" />
          </button>
        )}
        {onRun && (
          <button
            data-part="editor-run"
            onClick={onRun}
            disabled={running}
            className={cn(
              "flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium transition-colors",
              "bg-blue-600/20 text-blue-400 hover:bg-blue-600/30 hover:text-blue-300",
              "disabled:opacity-30 disabled:cursor-not-allowed",
            )}
            title="Run plugin"
          >
            <Play className="w-3 h-3" />
            <span>Run</span>
          </button>
        )}
      </div>
    </div>
  );
}
