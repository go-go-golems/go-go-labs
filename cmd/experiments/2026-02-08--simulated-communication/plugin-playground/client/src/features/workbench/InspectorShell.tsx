import React from "react";
import { WidgetRenderer } from "@/components/WidgetRenderer";
import type { DispatchTimelineEntry } from "@runtime/redux-adapter/store";
import type { UIEventRef } from "@runtime/uiTypes";
import type { LoadedPluginMap, WidgetErrors, WidgetTrees } from "./types";

interface InspectorShellProps {
  loadedPlugins: string[];
  pluginMetaById: LoadedPluginMap;
  widgetTrees: WidgetTrees;
  widgetErrors: WidgetErrors;
  dispatchTimeline: DispatchTimelineEntry[];
  sharedState: Record<string, unknown>;
  onWidgetEvent: (
    instanceId: string,
    widgetId: string,
    eventRef: UIEventRef,
    eventPayload?: unknown
  ) => void;
}

export function InspectorShell({
  loadedPlugins,
  pluginMetaById,
  widgetTrees,
  widgetErrors,
  dispatchTimeline,
  sharedState,
  onWidgetEvent,
}: InspectorShellProps) {
  const [activeTab, setActiveTab] = React.useState<"widgets" | "timeline" | "shared">("widgets");
  const [scopeFilter, setScopeFilter] = React.useState<"all" | "plugin" | "shared">("all");
  const [outcomeFilter, setOutcomeFilter] = React.useState<"all" | "applied" | "denied" | "ignored">(
    "all"
  );
  const [instanceFilter, setInstanceFilter] = React.useState("");

  const filteredTimeline = React.useMemo(() => {
    return dispatchTimeline
      .filter((entry) => {
        if (scopeFilter !== "all" && entry.scope !== scopeFilter) {
          return false;
        }
        if (outcomeFilter !== "all" && entry.outcome !== outcomeFilter) {
          return false;
        }
        if (instanceFilter && !(entry.instanceId ?? "").includes(instanceFilter.trim())) {
          return false;
        }
        return true;
      })
      .slice()
      .reverse();
  }, [dispatchTimeline, scopeFilter, outcomeFilter, instanceFilter]);

  return (
    <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
      <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">INSPECTOR</h2>
      <div className="flex gap-2 mb-3">
        <button
          onClick={() => setActiveTab("widgets")}
          className={`px-2 py-1 text-xs font-mono border rounded ${
            activeTab === "widgets" ? "border-cyan-400 text-cyan-300" : "border-cyan-400/20 text-muted-foreground"
          }`}
        >
          WIDGETS
        </button>
        <button
          onClick={() => setActiveTab("timeline")}
          className={`px-2 py-1 text-xs font-mono border rounded ${
            activeTab === "timeline" ? "border-cyan-400 text-cyan-300" : "border-cyan-400/20 text-muted-foreground"
          }`}
        >
          TIMELINE
        </button>
        <button
          onClick={() => setActiveTab("shared")}
          className={`px-2 py-1 text-xs font-mono border rounded ${
            activeTab === "shared" ? "border-cyan-400 text-cyan-300" : "border-cyan-400/20 text-muted-foreground"
          }`}
        >
          SHARED
        </button>
      </div>

      {activeTab === "widgets" && (
        <>
          {loadedPlugins.length === 0 ? (
            <div className="text-muted-foreground text-xs font-mono">No plugins loaded</div>
          ) : (
            <div className="space-y-4 flex-1 min-h-0 overflow-y-auto pr-1">
              {loadedPlugins.map((instanceId) => {
                const plugin = pluginMetaById[instanceId];
                if (!plugin) {
                  return (
                    <div key={instanceId} className="text-muted-foreground text-xs font-mono">
                      Loading plugin metadata: {instanceId}
                    </div>
                  );
                }

                return (
                  <div key={instanceId} className="border border-cyan-400/20 rounded p-2 bg-background/30">
                    <div className="text-xs font-bold text-cyan-400 mb-2 font-mono">
                      {plugin.title} [{instanceId}]
                    </div>
                    <div className="space-y-2">
                      {plugin.widgets.map((widgetId) => {
                        const widgetError = widgetErrors[instanceId]?.[widgetId];
                        if (widgetError) {
                          return (
                            <div key={widgetId} className="text-red-400 text-xs font-mono">
                              Render error: {widgetError}
                            </div>
                          );
                        }

                        const tree = widgetTrees[instanceId]?.[widgetId];
                        if (!tree) {
                          return (
                            <div key={widgetId} className="text-muted-foreground text-xs font-mono">
                              Rendering {widgetId}...
                            </div>
                          );
                        }

                        return (
                          <div key={widgetId}>
                            <WidgetRenderer
                              tree={tree}
                              onEvent={(eventRef, eventPayload) =>
                                onWidgetEvent(instanceId, widgetId, eventRef, eventPayload)
                              }
                            />
                          </div>
                        );
                      })}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </>
      )}

      {activeTab === "timeline" && (
        <div className="flex-1 min-h-0 flex flex-col">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-2 mb-3">
            <select
              value={scopeFilter}
              onChange={(e) => setScopeFilter(e.target.value as typeof scopeFilter)}
              className="bg-background border border-cyan-400/20 rounded px-2 py-1 text-xs font-mono"
            >
              <option value="all">scope: all</option>
              <option value="plugin">scope: plugin</option>
              <option value="shared">scope: shared</option>
            </select>
            <select
              value={outcomeFilter}
              onChange={(e) => setOutcomeFilter(e.target.value as typeof outcomeFilter)}
              className="bg-background border border-cyan-400/20 rounded px-2 py-1 text-xs font-mono"
            >
              <option value="all">outcome: all</option>
              <option value="applied">outcome: applied</option>
              <option value="denied">outcome: denied</option>
              <option value="ignored">outcome: ignored</option>
            </select>
            <input
              value={instanceFilter}
              onChange={(e) => setInstanceFilter(e.target.value)}
              placeholder="instance filter"
              className="bg-background border border-cyan-400/20 rounded px-2 py-1 text-xs font-mono"
            />
          </div>
          <div className="space-y-2 flex-1 min-h-0 overflow-y-auto pr-1">
            {filteredTimeline.map((entry) => (
              <div key={entry.dispatchId} className="border border-cyan-400/20 rounded p-2 bg-background/30">
                <div className="text-xs font-mono text-cyan-300">
                  {new Date(entry.timestamp).toLocaleTimeString()} · {entry.scope} · {entry.outcome}
                </div>
                <div className="text-xs font-mono text-foreground mt-1">{entry.actionType}</div>
                <div className="text-[11px] font-mono text-muted-foreground mt-1">
                  instance={entry.instanceId ?? "-"} domain={entry.domain ?? "-"}
                  {entry.reason ? ` reason=${entry.reason}` : ""}
                </div>
              </div>
            ))}
            {filteredTimeline.length === 0 && (
              <div className="text-muted-foreground text-xs font-mono">No timeline entries for current filters</div>
            )}
          </div>
        </div>
      )}

      {activeTab === "shared" && (
        <div className="flex-1 min-h-0 overflow-y-auto pr-1">
          <pre className="text-[11px] font-mono whitespace-pre-wrap break-words border border-cyan-400/20 rounded p-2 bg-background/30">
            {JSON.stringify(sharedState, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
