import React from "react";
import { useDispatch, useSelector } from "react-redux";
import { CatalogShell } from "@/features/workbench/CatalogShell";
import { InspectorShell } from "@/features/workbench/InspectorShell";
import type { WidgetErrors, WidgetTrees } from "@/features/workbench/types";
import { WorkspaceShell } from "@/features/workbench/WorkspaceShell";
import { presetPlugins } from "@/lib/presetPlugins";
import { createInstanceId } from "@runtime/runtimeIdentity";
import {
  AppDispatch,
  CapabilityGrants,
  RootState,
  SharedDomainName,
  dispatchPluginAction,
  dispatchSharedAction,
  pluginRegistered,
  pluginRemoved,
  selectAllPluginState,
  selectGlobalStateForInstance,
  selectLoadedPluginIds,
} from "@runtime/redux-adapter/store";
import { quickjsSandboxClient } from "@runtime/worker/sandboxClient";
import type { LoadedPlugin } from "@runtime/contracts";
import type { UIEventRef } from "@runtime/uiTypes";

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
          <CatalogShell
            presets={presetPlugins}
            loadedCountsByPackage={loadedCountsByPackage}
            loadedPlugins={loadedPlugins}
            pluginMetaById={pluginMetaById}
            onLoadPreset={(presetId) => void loadPreset(presetId)}
            onUnloadPlugin={(instanceId) => void unloadPlugin(instanceId)}
          />
          <WorkspaceShell
            customCode={customCode}
            error={error}
            onCustomCodeChange={setCustomCode}
            onLoadCustom={() => void loadCustom()}
          />
          <InspectorShell
            loadedPlugins={loadedPlugins}
            pluginMetaById={pluginMetaById}
            widgetTrees={widgetTrees}
            widgetErrors={widgetErrors}
            onWidgetEvent={(instanceId, widgetId, eventRef, eventPayload) =>
              void handleEvent(instanceId, widgetId, eventRef, eventPayload)
            }
          />
        </div>
      </div>
    </div>
  );
}
