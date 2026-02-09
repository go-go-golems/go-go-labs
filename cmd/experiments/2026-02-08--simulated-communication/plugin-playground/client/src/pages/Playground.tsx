import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { WidgetRenderer } from "@/components/WidgetRenderer";
import { Button } from "@/components/ui/button";
import { presetPlugins } from "@/lib/presetPlugins";
import { createInstanceId } from "@runtime/runtimeIdentity";
import { quickjsSandboxClient } from "@runtime/worker/sandboxClient";
import type { LoadedPlugin } from "@runtime/contracts";
import type { UINode, UIEventRef } from "@runtime/uiTypes";
import {
  AppDispatch,
  CapabilityGrants,
  RootState,
  dispatchPluginAction,
  dispatchSharedAction,
  pluginRegistered,
  pluginRemoved,
  selectAllPluginState,
  selectGlobalStateForInstance,
  selectLoadedPluginIds,
  SharedDomainName,
} from "@/store/store";

type WidgetTrees = Record<string, Record<string, UINode>>;
type WidgetErrors = Record<string, Record<string, string>>;

export default function Playground() {
  const dispatch = useDispatch<AppDispatch>();
  const rootState = useSelector((s: RootState) => s);
  const loadedPlugins = useSelector((s: RootState) => selectLoadedPluginIds(s));
  const pluginStateById = useSelector((s: RootState) => selectAllPluginState(s));

  const [pluginMetaById, setPluginMetaById] = React.useState<Record<string, LoadedPlugin>>({});
  const [widgetTrees, setWidgetTrees] = React.useState<WidgetTrees>({});
  const [widgetErrors, setWidgetErrors] = React.useState<WidgetErrors>({});
  const [customCode, setCustomCode] = React.useState<string>("");
  const [error, setError] = React.useState<string>("");

  const registerLoadedPlugin = React.useCallback(
    (plugin: LoadedPlugin, grants?: CapabilityGrants) => {
      dispatch(
        pluginRegistered({
          instanceId: plugin.instanceId,
          packageId: plugin.packageId,
          title: plugin.title,
          description: plugin.description,
          widgets: plugin.widgets,
          initialState: plugin.initialState,
          grants,
        })
      );

      setPluginMetaById((current) => ({
        ...current,
        [plugin.instanceId]: plugin,
      }));
    },
    [dispatch]
  );

  React.useEffect(() => {
    setPluginMetaById((current) => {
      const next: Record<string, LoadedPlugin> = {};
      for (const instanceId of loadedPlugins) {
        if (current[instanceId]) {
          next[instanceId] = current[instanceId];
        }
      }
      return next;
    });
  }, [loadedPlugins]);

  const loadedCountsByPackage = React.useMemo(() => {
    const counts: Record<string, number> = {};
    for (const instanceId of loadedPlugins) {
      const meta = pluginMetaById[instanceId];
      if (!meta) {
        continue;
      }
      counts[meta.packageId] = (counts[meta.packageId] ?? 0) + 1;
    }
    return counts;
  }, [loadedPlugins, pluginMetaById]);

  React.useEffect(() => {
    let cancelled = false;

    const renderAllWidgets = async () => {
      const nextTrees: WidgetTrees = {};
      const nextErrors: WidgetErrors = {};

      for (const instanceId of loadedPlugins) {
        const pluginMeta = pluginMetaById[instanceId];
        if (!pluginMeta) {
          continue;
        }

        const pluginState = pluginStateById[instanceId] ?? {};
        const globalState = selectGlobalStateForInstance(rootState, instanceId);
        nextTrees[instanceId] = {};

        for (const widgetId of pluginMeta.widgets) {
          try {
            const tree = await quickjsSandboxClient.render(instanceId, widgetId, pluginState, globalState);
            nextTrees[instanceId][widgetId] = tree;
          } catch (err) {
            if (!nextErrors[instanceId]) {
              nextErrors[instanceId] = {};
            }
            nextErrors[instanceId][widgetId] = String(err);
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
  }, [loadedPlugins, pluginMetaById, pluginStateById, rootState]);

  const loadPreset = async (presetId: string) => {
    try {
      setError("");
      const preset = presetPlugins.find((p) => p.id === presetId);
      if (!preset) {
        throw new Error(`Preset not found: ${presetId}`);
      }

      const instanceId = createInstanceId(preset.id);
      const plugin = await quickjsSandboxClient.loadPlugin(preset.id, instanceId, preset.code);
      registerLoadedPlugin(plugin, {
        readShared: preset.capabilities?.readShared ?? [],
        writeShared: preset.capabilities?.writeShared ?? [],
        systemCommands: preset.capabilities?.systemCommands ?? [],
      });
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

      const packageId = "custom";
      const instanceId = createInstanceId(packageId);
      const plugin = await quickjsSandboxClient.loadPlugin(packageId, instanceId, customCode);
      registerLoadedPlugin(plugin, {
        readShared: [],
        writeShared: [],
        systemCommands: [],
      });
    } catch (err) {
      setError(`Failed to load custom plugin: ${String(err)}`);
    }
  };

  const unloadPlugin = async (instanceId: string) => {
    try {
      await quickjsSandboxClient.disposePlugin(instanceId);
    } catch (err) {
      console.warn(`Failed to dispose plugin runtime for ${instanceId}:`, err);
    }

    dispatch(pluginRemoved(instanceId));
    setWidgetTrees((current) => {
      const next = { ...current };
      delete next[instanceId];
      return next;
    });
    setWidgetErrors((current) => {
      const next = { ...current };
      delete next[instanceId];
      return next;
    });
  };

  const handleEvent = React.useCallback(
    async (instanceId: string, widgetId: string, eventRef: UIEventRef, eventPayload?: unknown) => {
      try {
        const pluginState = pluginStateById[instanceId] ?? {};
        const globalState = selectGlobalStateForInstance(rootState, instanceId);
        const handlerArgs = eventPayload ?? eventRef.args;
        const intents = await quickjsSandboxClient.event(
          instanceId,
          widgetId,
          eventRef.handler,
          handlerArgs,
          pluginState,
          globalState
        );

        for (const intent of intents) {
          if (intent.scope === "plugin") {
            dispatchPluginAction(dispatch, instanceId, intent.actionType, intent.payload);
            continue;
          }

          if (!intent.domain) {
            continue;
          }
          dispatchSharedAction(
            dispatch,
            instanceId,
            intent.domain as SharedDomainName,
            intent.actionType,
            intent.payload
          );
        }
      } catch (err) {
        console.error("Event error:", err);
        setError(`Event failed: ${String(err)}`);
      }
    },
    [dispatch, pluginStateById, rootState]
  );

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
                    variant={(loadedCountsByPackage[preset.id] ?? 0) > 0 ? "default" : "outline"}
                    className="w-full justify-start font-mono text-xs"
                  >
                    {preset.title}
                    {(loadedCountsByPackage[preset.id] ?? 0) > 0 &&
                      ` (${loadedCountsByPackage[preset.id]})`}
                  </Button>
                ))}
              </div>

              <div className="mt-6 border-t border-cyan-400/20 pt-4">
                <h3 className="text-sm font-bold text-cyan-400 mb-2 font-mono">LOADED</h3>
                <div className="space-y-1">
                  {loadedPlugins.map((instanceId) => (
                    <div key={instanceId} className="flex items-center justify-between text-xs font-mono">
                      <span>
                        {pluginMetaById[instanceId]?.title ?? "Plugin"} [{instanceId}]
                      </span>
                      <button
                        onClick={() => void unloadPlugin(instanceId)}
                        className="text-red-400 hover:text-red-300"
                      >
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
                                  void handleEvent(instanceId, widgetId, eventRef, eventPayload)
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
