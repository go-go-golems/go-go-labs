import React from "react";
import { cn } from "@/lib/utils";
import { ChevronLeft, ChevronRight, Plus, X } from "lucide-react";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export interface CatalogEntry {
  id: string;
  title: string;
  description?: string;
  /** Abbreviated capability summary, e.g. "R/W", "R", or empty. */
  capabilitySummary?: string;
}

export interface RunningInstance {
  instanceId: string;
  title: string;
  packageId: string;
  shortId: string;
  status: "loaded" | "error";
  readGrants: string[];
  writeGrants: string[];
}

export interface SidebarProps {
  catalog: CatalogEntry[];
  running: RunningInstance[];
  collapsed?: boolean;
  focusedInstanceId?: string;
  onToggleCollapse?: () => void;
  onLoadPreset?: (presetId: string) => void;
  onFocusInstance?: (instanceId: string) => void;
  onUnloadInstance?: (instanceId: string) => void;
  onNewPlugin?: () => void;
  className?: string;
}

// ---------------------------------------------------------------------------
// Collapsed sidebar
// ---------------------------------------------------------------------------

function CollapsedSidebar({
  running,
  onToggleCollapse,
  onNewPlugin,
}: Pick<SidebarProps, "running" | "onToggleCollapse" | "onNewPlugin">) {
  return (
    <div data-part="sidebar-body" data-state="collapsed" className="flex flex-col items-center py-3 gap-3">
      <button
        onClick={onToggleCollapse}
        className="p-1 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors"
        title="Expand sidebar"
      >
        <ChevronRight className="w-4 h-4" />
      </button>
      <span className="text-sm select-none" title="Catalog">📦</span>
      <div className="relative" title={`${running.length} running`}>
        <span className="text-sm select-none">🔌</span>
        {running.length > 0 && (
          <span className="absolute -top-1.5 -right-2.5 text-[10px] font-semibold text-blue-400 tabular-nums">
            {running.length}
          </span>
        )}
      </div>
      <button
        onClick={onNewPlugin}
        className="p-1 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors"
        title="New plugin"
      >
        <Plus className="w-4 h-4" />
      </button>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main sidebar component
// ---------------------------------------------------------------------------

export function Sidebar({
  catalog,
  running,
  collapsed = false,
  focusedInstanceId,
  onToggleCollapse,
  onLoadPreset,
  onFocusInstance,
  onUnloadInstance,
  onNewPlugin,
  className,
}: SidebarProps) {
  if (collapsed) {
    return (
      <CollapsedSidebar
        running={running}
        onToggleCollapse={onToggleCollapse}
        onNewPlugin={onNewPlugin}
      />
    );
  }

  return (
    <div
      data-part="sidebar-body"
      data-state="expanded"
      className={cn("flex flex-col h-full text-sm", className)}
    >
      {/* Header */}
      <div
        data-part="sidebar-header"
        className="flex items-center justify-between px-3 h-10 border-b border-white/[0.06] flex-shrink-0"
      >
        <span className="font-semibold text-slate-100 text-sm tracking-tight">Workbench</span>
        <button
          onClick={onToggleCollapse}
          className="p-1 rounded text-slate-500 hover:text-slate-200 hover:bg-slate-800/50 transition-colors"
          title="Collapse sidebar"
        >
          <ChevronLeft className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Scrollable content */}
      <div className="flex-1 min-h-0 overflow-y-auto">
        {/* Catalog section */}
        <section data-part="sidebar-catalog" className="px-2 pt-3 pb-2">
          <h3 className="px-2 mb-1.5 text-xs font-medium text-slate-500 uppercase tracking-wider">
            Catalog
          </h3>
          <div className="space-y-px">
            {catalog.map((entry) => (
              <button
                key={entry.id}
                data-part="catalog-item"
                onClick={() => onLoadPreset?.(entry.id)}
                className="w-full text-left px-2 py-1.5 rounded-md text-slate-300 hover:text-slate-100 hover:bg-slate-800/60 transition-colors group"
              >
                <div className="flex items-center justify-between gap-1">
                  <span className="truncate text-[13px]">{entry.title}</span>
                  {entry.capabilitySummary && (
                    <span className="flex-shrink-0 text-[10px] font-mono text-slate-600 group-hover:text-slate-500">
                      {entry.capabilitySummary}
                    </span>
                  )}
                </div>
                {entry.description && (
                  <div className="text-xs text-slate-600 group-hover:text-slate-500 truncate mt-0.5">
                    {entry.description}
                  </div>
                )}
              </button>
            ))}
          </div>
        </section>

        {/* Running instances section */}
        <section data-part="sidebar-running" className="px-2 pt-2 pb-3">
          <h3 className="px-2 mb-1.5 text-xs font-medium text-slate-500 uppercase tracking-wider">
            Running
            {running.length > 0 && (
              <span className="ml-1.5 text-slate-600">({running.length})</span>
            )}
          </h3>

          {running.length === 0 ? (
            <p className="px-2 text-xs text-slate-600">No plugins loaded</p>
          ) : (
            <div className="space-y-1">
              {running.map((instance) => {
                const focused = focusedInstanceId === instance.instanceId;
                return (
                  <div
                    key={instance.instanceId}
                    data-part="instance-item"
                    data-state={focused ? "focused" : undefined}
                    onClick={() => onFocusInstance?.(instance.instanceId)}
                    className={cn(
                      "px-2 py-1.5 rounded-md transition-colors cursor-pointer",
                      focused
                        ? "bg-blue-500/10 ring-1 ring-blue-500/25"
                        : "hover:bg-slate-800/60"
                    )}
                  >
                    <div className="flex items-center justify-between gap-1">
                      <span className="flex items-center gap-1.5 min-w-0">
                        <span
                          data-part="status-dot"
                          className={cn(
                            "inline-block w-1.5 h-1.5 rounded-full flex-shrink-0",
                            instance.status === "loaded" ? "bg-emerald-500" : "bg-red-500"
                          )}
                        />
                        <span className="text-[13px] text-slate-200 truncate">{instance.title}</span>
                      </span>
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          onUnloadInstance?.(instance.instanceId);
                        }}
                        className="p-0.5 rounded text-slate-600 hover:text-red-400 hover:bg-slate-800 transition-colors flex-shrink-0 opacity-0 group-hover:opacity-100 [div:hover>&]:opacity-100"
                        title="Unload"
                      >
                        <X className="w-3 h-3" />
                      </button>
                    </div>
                    <div className="text-[11px] font-mono text-slate-600 mt-0.5 truncate pl-[18px]">
                      {instance.shortId}
                    </div>
                    {(instance.readGrants.length > 0 || instance.writeGrants.length > 0) && (
                      <div className="mt-0.5 pl-[18px] space-y-px">
                        {instance.readGrants.length > 0 && (
                          <div className="text-[10px] font-mono text-slate-600 truncate">
                            <span className="text-slate-500">R</span> {instance.readGrants.join(", ")}
                          </div>
                        )}
                        {instance.writeGrants.length > 0 && (
                          <div className="text-[10px] font-mono text-slate-600 truncate">
                            <span className="text-slate-500">W</span> {instance.writeGrants.join(", ")}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </section>
      </div>

      {/* New plugin button */}
      <div data-part="sidebar-footer" className="flex-shrink-0 border-t border-white/[0.06] p-2">
        <button
          onClick={onNewPlugin}
          className="w-full flex items-center justify-center gap-1.5 h-8 rounded-md text-xs font-medium text-slate-400 hover:text-slate-200 border border-white/[0.08] hover:border-white/[0.15] hover:bg-slate-800/50 transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          New Plugin
        </button>
      </div>
    </div>
  );
}
