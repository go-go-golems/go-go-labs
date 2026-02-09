import { WidgetRenderer } from "@/components/WidgetRenderer";
import type { UIEventRef } from "@runtime/uiTypes";
import type { LoadedPluginMap, WidgetErrors, WidgetTrees } from "./types";

interface InspectorShellProps {
  loadedPlugins: string[];
  pluginMetaById: LoadedPluginMap;
  widgetTrees: WidgetTrees;
  widgetErrors: WidgetErrors;
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
  onWidgetEvent,
}: InspectorShellProps) {
  return (
    <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
      <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">INSPECTOR</h2>
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
    </div>
  );
}
