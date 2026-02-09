import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { WidgetRenderer } from "@/components/WidgetRenderer";
import { Button } from "@/components/ui/button";
import { presetPlugins } from "@/lib/presetPlugins";
import { quickjsSandboxClient } from "@/lib/quickjsSandboxClient";
import type { LoadedPlugin } from "@/lib/quickjsContracts";
import type { UINode, UIEventRef } from "@/lib/uiTypes";
import {
  AppDispatch,
  RootState,
  dispatchGlobalAction,
  dispatchPluginAction,
  pluginRegistered,
  pluginRemoved,
  selectAllPluginState,
  selectGlobalState,
  selectLoadedPluginIds,
} from "@/store/store";

type WidgetTrees = Record<string, Record<string, UINode>>;
type WidgetErrors = Record<string, Record<string, string>>;

export default function Playground() {
  const dispatch = useDispatch<AppDispatch>();
  const loadedPlugins = useSelector((s: RootState) => selectLoadedPluginIds(s));
  const pluginStateById = useSelector((s: RootState) => selectAllPluginState(s));
  const globalState = useSelector((s: RootState) => selectGlobalState(s));

  const [pluginMetaById, setPluginMetaById] = React.useState<Record<string, LoadedPlugin>>({});
  const [widgetTrees, setWidgetTrees] = React.useState<WidgetTrees>({});
  const [widgetErrors, setWidgetErrors] = React.useState<WidgetErrors>({});
  const [customCode, setCustomCode] = React.useState<string>("");
  const [error, setError] = React.useState<string>("");

  const registerLoadedPlugin = React.useCallback(
    (plugin: LoadedPlugin) => {
      dispatch(
        pluginRegistered({
          id: plugin.id,
          title: plugin.title,
          description: plugin.description,
          widgets: plugin.widgets,
          initialState: plugin.initialState,
        })
      );

      setPluginMetaById((current) => ({
        ...current,
        [plugin.id]: plugin,
      }));
    },
    [dispatch]
  );

  React.useEffect(() => {
    setPluginMetaById((current) => {
      const next: Record<string, LoadedPlugin> = {};
      for (const pluginId of loadedPlugins) {
        if (current[pluginId]) {
          next[pluginId] = current[pluginId];
        }
      }
      return next;
    });
  }, [loadedPlugins]);

  React.useEffect(() => {
    let cancelled = false;

    const renderAllWidgets = async () => {
      const nextTrees: WidgetTrees = {};
      const nextErrors: WidgetErrors = {};

      for (const pluginId of loadedPlugins) {
        const pluginMeta = pluginMetaById[pluginId];
        if (!pluginMeta) {
          continue;
        }

        const pluginState = pluginStateById[pluginId] ?? {};
        nextTrees[pluginId] = {};

        for (const widgetId of pluginMeta.widgets) {
          try {
            const tree = await quickjsSandboxClient.render(pluginId, widgetId, pluginState, globalState);
            nextTrees[pluginId][widgetId] = tree;
          } catch (err) {
            if (!nextErrors[pluginId]) {
              nextErrors[pluginId] = {};
            }
            nextErrors[pluginId][widgetId] = String(err);
          }
        }
      }

      if (!cancelled) {
        setWidgetTrees(nextTrees);
        setWidgetErrors(nextErrors);
      }
    };

    void renderAllWidgets();

    return () => {
      cancelled = true;
    };
  }, [globalState, loadedPlugins, pluginMetaById, pluginStateById]);

  const loadPreset = async (presetId: string) => {
    try {
      setError("");
      const preset = presetPlugins.find((p) => p.id === presetId);
      if (!preset) {
        throw new Error(`Preset not found: ${presetId}`);
      }

      const plugin = await quickjsSandboxClient.loadPlugin(preset.id, preset.code);
      registerLoadedPlugin(plugin);
    } catch (err) {
      setError(`Failed to load preset: ${String(err)}`);
    }
  };

  const loadCustom = async () => {
    try {
      setError("");
      if (!customCode.trim()) {
        throw new Error("Plugin code cannot be empty");
      }

      const pluginId = `custom-${Date.now()}`;
      const plugin = await quickjsSandboxClient.loadPlugin(pluginId, customCode);
      registerLoadedPlugin(plugin);
    } catch (err) {
      setError(`Failed to load custom plugin: ${String(err)}`);
    }
  };

  const unloadPlugin = async (pluginId: string) => {
    try {
      await quickjsSandboxClient.disposePlugin(pluginId);
    } catch (err) {
      console.warn(`Failed to dispose plugin runtime for ${pluginId}:`, err);
    }

    dispatch(pluginRemoved(pluginId));
    setWidgetTrees((current) => {
      const next = { ...current };
      delete next[pluginId];
      return next;
    });
    setWidgetErrors((current) => {
      const next = { ...current };
      delete next[pluginId];
      return next;
    });
  };

  const handleEvent = async (
    pluginId: string,
    widgetId: string,
    eventRef: UIEventRef,
    eventPayload?: unknown
  ) => {
    try {
      const pluginState = pluginStateById[pluginId] ?? {};
      const handlerArgs = eventPayload ?? eventRef.args;
      const intents = await quickjsSandboxClient.event(
        pluginId,
        widgetId,
        eventRef.handler,
        handlerArgs,
        pluginState,
        globalState
      );

      for (const intent of intents) {
        if (intent.scope === "plugin") {
          dispatchPluginAction(dispatch, pluginId, intent.actionType, intent.payload);
          continue;
        }

        dispatchGlobalAction(dispatch, intent.actionType, intent.payload);
      }
    } catch (err) {
      console.error("Event error:", err);
      setError(`Event failed: ${String(err)}`);
    }
  };

  return (
    <div className="min-h-dvh bg-background text-foreground p-4">
      <div className="max-w-7xl mx-auto h-full min-h-0 flex flex-col">
        <h1 className="text-3xl font-bold text-cyan-400 mb-2 font-mono">PLUGIN PLAYGROUND</h1>
        <p className="text-muted-foreground mb-6 font-mono text-sm">
          Unified Runtime v1 - Plugin/Global State and Action Scoping
        </p>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4 flex-1 min-h-0">
          <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
            <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">PRESETS</h2>
            <div className="flex-1 min-h-0 overflow-y-auto pr-1">
              <div className="space-y-2">
                {presetPlugins.map((preset) => (
                  <Button
                    key={preset.id}
                    onClick={() => loadPreset(preset.id)}
                    variant={loadedPlugins.includes(preset.id) ? "default" : "outline"}
                    className="w-full justify-start font-mono text-xs"
                  >
                    {preset.title}
                    {loadedPlugins.includes(preset.id) && " ✓"}
                  </Button>
                ))}
              </div>

              <div className="mt-6 border-t border-cyan-400/20 pt-4">
                <h3 className="text-sm font-bold text-cyan-400 mb-2 font-mono">LOADED</h3>
                <div className="space-y-1">
                  {loadedPlugins.map((id) => (
                    <div key={id} className="flex items-center justify-between text-xs font-mono">
                      <span>{id}</span>
                      <button onClick={() => void unloadPlugin(id)} className="text-red-400 hover:text-red-300">
                        X
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
            <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">CUSTOM PLUGIN</h2>
            <textarea
              value={customCode}
              onChange={(e) => setCustomCode(e.target.value)}
              placeholder="definePlugin(({ ui }) => { ... })"
              className="w-full flex-1 min-h-[12rem] bg-background/50 border border-cyan-400/20 rounded p-2 font-mono text-xs text-foreground resize-none focus:outline-none focus:border-cyan-400"
            />
            <Button onClick={() => void loadCustom()} className="w-full mt-2 font-mono text-xs">
              LOAD PLUGIN
            </Button>
            {error && <div className="mt-2 text-red-400 text-xs font-mono">{error}</div>}
          </div>

          <div className="border border-cyan-400/30 rounded-sm p-4 bg-card/50 h-full min-h-0 flex flex-col">
            <h2 className="text-lg font-bold text-cyan-400 mb-4 font-mono">LIVE WIDGETS</h2>
            {loadedPlugins.length === 0 ? (
              <div className="text-muted-foreground text-xs font-mono">No plugins loaded</div>
            ) : (
              <div className="space-y-4 flex-1 min-h-0 overflow-y-auto pr-1">
                {loadedPlugins.map((pluginId) => {
                  const plugin = pluginMetaById[pluginId];
                  if (!plugin) {
                    return (
                      <div key={pluginId} className="text-muted-foreground text-xs font-mono">
                        Loading plugin metadata: {pluginId}
                      </div>
                    );
                  }

                  return (
                    <div key={pluginId} className="border border-cyan-400/20 rounded p-2 bg-background/30">
                      <div className="text-xs font-bold text-cyan-400 mb-2 font-mono">{plugin.title}</div>
                      <div className="space-y-2">
                        {plugin.widgets.map((widgetId) => {
                          const widgetError = widgetErrors[pluginId]?.[widgetId];
                          if (widgetError) {
                            return (
                              <div key={widgetId} className="text-red-400 text-xs font-mono">
                                Render error: {widgetError}
                              </div>
                            );
                          }

                          const tree = widgetTrees[pluginId]?.[widgetId];
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
                                  void handleEvent(pluginId, widgetId, eventRef, eventPayload)
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
        </div>
      </div>
    </div>
  );
}
