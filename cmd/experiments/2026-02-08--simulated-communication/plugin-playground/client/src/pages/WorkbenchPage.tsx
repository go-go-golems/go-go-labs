/**
 * WorkbenchPage — top-level page that wires the new workbench components
 * to the plugin runtime sandbox.
 *
 * This replaces Playground.tsx as the main page of the app.
 */
import React from "react";
import { WorkbenchLayout } from "@/features/workbench/components/WorkbenchLayout";
import { Sidebar, type CatalogEntry, type RunningInstance } from "@/features/workbench/components/Sidebar";
import { TopToolbar, type HealthStatus } from "@/features/workbench/components/TopToolbar";
import { EditorTabBar, type EditorTabInfo } from "@/features/workbench/components/EditorTabBar";
import { CodeEditor } from "@/features/workbench/components/CodeEditor";
import { LivePreview } from "@/features/workbench/components/LivePreview";
import { InstanceCard } from "@/features/workbench/components/InstanceCard";
import { DevToolsPanel } from "@/features/workbench/components/DevToolsPanel";
import { TimelinePanel, type TimelineEntry } from "@/features/workbench/components/TimelinePanel";
import { StatePanel, type InstanceState } from "@/features/workbench/components/StatePanel";
import { CapabilitiesPanel, type InstanceCapabilities } from "@/features/workbench/components/CapabilitiesPanel";
import { ErrorsPanel } from "@/features/workbench/components/ErrorsPanel";
import { SharedDomainsPanel, type SharedDomainInfo } from "@/features/workbench/components/SharedDomainsPanel";
import { DocsPanel } from "@/features/workbench/components/DocsPanel";
import { docs as docEntries, buildAllDocsMarkdown } from "@/lib/docsManifest";
import { WidgetRenderer } from "@/components/WidgetRenderer";

import {
  useAppDispatch,
  useAppSelector,
  toggleSidebar,
  toggleDevtools,
  setActiveDevToolsTab,
  focusInstance,
  openEditorTab,
  closeEditorTab,
  setActiveEditorTab,
  updateEditorCode,
  markEditorTabClean,
  setTabActiveInstance,
  pushError,
  clearErrors,
  selectLoadedPluginIds,
  selectAllPluginState,
  selectDispatchTimeline,
  selectGlobalState,
  selectGlobalStateForInstance,
  pluginRegistered,
  pluginRemoved,
  dispatchPluginAction,
  dispatchSharedAction,
  type DevToolsTab,
  type SharedDomainName,
  type DispatchTimelineEntry,
} from "@/store";

import { presetPlugins } from "@/lib/presetPlugins";
import { createInstanceId } from "@runtime/runtimeIdentity";
import { quickjsSandboxClient } from "@runtime/worker/sandboxClient";
import type { LoadedPlugin } from "@runtime/contracts";
import type { CapabilityGrants } from "@runtime/redux-adapter/store";
import type { UIEventRef, UINode } from "@runtime/uiTypes";

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function capSummary(preset: (typeof presetPlugins)[number]): string {
  const r = (preset.capabilities?.readShared?.length ?? 0) > 0;
  const w = (preset.capabilities?.writeShared?.length ?? 0) > 0;
  if (r && w) return "R/W";
  if (r) return "R";
  return "";
}

const CATALOG: CatalogEntry[] = presetPlugins.map((p) => ({
  id: p.id,
  title: p.title,
  description: p.description,
  capabilitySummary: capSummary(p),
}));

const ALL_SHARED_DOMAINS: SharedDomainName[] = ["counter-summary", "greeter-profile", "runtime-registry", "runtime-metrics"];
const allDocsMarkdown = buildAllDocsMarkdown();

// Widget trees and errors are local state, not in the store
// (they come from async sandbox rendering, not user actions)
type WidgetTrees = Record<string, Record<string, UINode>>;
type WidgetErrors = Record<string, Record<string, string>>;

// ---------------------------------------------------------------------------
// Page component
// ---------------------------------------------------------------------------

export default function WorkbenchPage() {
  const dispatch = useAppDispatch();
  const rootState = useAppSelector((s) => s);

  // RTK workbench state
  const sidebarCollapsed = useAppSelector((s) => s.workbench.sidebarCollapsed);
  const devtoolsCollapsed = useAppSelector((s) => s.workbench.devtoolsCollapsed);
  const activeDevTab = useAppSelector((s) => s.workbench.activeDevToolsTab);
  const focusedInstanceId = useAppSelector((s) => s.workbench.focusedInstanceId);
  const editorTabs = useAppSelector((s) => s.workbench.editorTabs);
  const activeEditorTabId = useAppSelector((s) => s.workbench.activeEditorTabId);
  const errors = useAppSelector((s) => s.workbench.errors);

  // RTK runtime state
  const loadedPluginIds = useAppSelector(selectLoadedPluginIds);
  const pluginStateById = useAppSelector(selectAllPluginState);
  const dispatchTimeline = useAppSelector(selectDispatchTimeline);
  const sharedStateSnapshot = useAppSelector((s) => selectGlobalState(s).shared);
  const plugins = useAppSelector((s) => s.runtime.plugins);
  const grantsByInstance = useAppSelector((s) => s.runtime.grantsByInstance);

  // Local state for plugin meta + widget rendering (async, not in store)
  const [pluginMetaById, setPluginMetaById] = React.useState<Record<string, LoadedPlugin>>({});
  const [widgetTrees, setWidgetTrees] = React.useState<WidgetTrees>({});
  const [widgetErrors, setWidgetErrors] = React.useState<WidgetErrors>({});

  // Clean up pluginMetaById when plugins are removed
  React.useEffect(() => {
    setPluginMetaById((current) => {
      const next: Record<string, LoadedPlugin> = {};
      for (const id of loadedPluginIds) {
        if (current[id]) next[id] = current[id];
      }
      return next;
    });
  }, [loadedPluginIds]);

  // Re-render all widgets when state changes
  React.useEffect(() => {
    let cancelled = false;

    const renderAll = async () => {
      const nextTrees: WidgetTrees = {};
      const nextErrors: WidgetErrors = {};

      for (const instanceId of loadedPluginIds) {
        const meta = pluginMetaById[instanceId];
        if (!meta) continue;
        const pluginState = pluginStateById[instanceId] ?? {};
        const globalState = selectGlobalStateForInstance(rootState, instanceId);
        nextTrees[instanceId] = {};

        for (const widgetId of meta.widgets) {
          try {
            const tree = await quickjsSandboxClient.render(instanceId, widgetId, pluginState, globalState);
            nextTrees[instanceId][widgetId] = tree;
          } catch (err) {
            if (!nextErrors[instanceId]) nextErrors[instanceId] = {};
            nextErrors[instanceId][widgetId] = String(err);
            dispatch(pushError({ kind: "render", instanceId, widgetId, message: String(err) }));
          }
        }
      }

      if (!cancelled) {
        setWidgetTrees(nextTrees);
        setWidgetErrors(nextErrors);
      }
    };

    void renderAll();
    return () => { cancelled = true; };
  }, [loadedPluginIds, pluginMetaById, pluginStateById, rootState, dispatch]);

  // -----------------------------------------------------------------------
  // Actions
  // -----------------------------------------------------------------------

  const registerPlugin = React.useCallback(
    (plugin: LoadedPlugin, grants?: CapabilityGrants) => {
      dispatch(pluginRegistered({
        instanceId: plugin.instanceId,
        packageId: plugin.packageId,
        title: plugin.title,
        description: plugin.description,
        widgets: plugin.widgets,
        initialState: plugin.initialState,
        grants,
      }));
      setPluginMetaById((c) => ({ ...c, [plugin.instanceId]: plugin }));
    },
    [dispatch],
  );

  /** Open a preset in the editor (no sandbox execution). */
  const openPreset = React.useCallback((presetId: string) => {
    const preset = presetPlugins.find((p) => p.id === presetId);
    if (!preset) return;

    // Reuse existing editor tab for this package, or open a new one
    const existingTab = editorTabs.find((t) => t.packageId === preset.id);
    if (existingTab) {
      dispatch(setActiveEditorTab(existingTab.id));
    } else {
      dispatch(openEditorTab({ packageId: preset.id, label: `${preset.title}.js`, code: preset.code }));
    }
  }, [dispatch, editorTabs]);

  const runEditorTab = React.useCallback(async (code: string, packageId: string) => {
    try {
      if (!code.trim()) throw new Error("Plugin code cannot be empty");

      // If packageId matches a preset, use its capabilities.
      // Otherwise grant all shared domains (sandbox is for experimentation).
      const preset = presetPlugins.find((p) => p.id === packageId);
      const grants: CapabilityGrants = preset
        ? {
            readShared: preset.capabilities?.readShared ?? [],
            writeShared: preset.capabilities?.writeShared ?? [],
            systemCommands: preset.capabilities?.systemCommands ?? [],
          }
        : {
            readShared: ALL_SHARED_DOMAINS,
            writeShared: ALL_SHARED_DOMAINS,
            systemCommands: [],
          };

      const instanceId = createInstanceId(packageId);
      const plugin = await quickjsSandboxClient.loadPlugin(packageId, instanceId, code);
      registerPlugin(plugin, grants);
      dispatch(focusInstance(plugin.instanceId));
    } catch (err) {
      dispatch(pushError({ kind: "load", instanceId: null, widgetId: null, message: String(err) }));
    }
  }, [dispatch, registerPlugin]);

  const unloadPlugin = React.useCallback(async (instanceId: string) => {
    try { await quickjsSandboxClient.disposePlugin(instanceId); } catch {}
    dispatch(pluginRemoved(instanceId));
    setWidgetTrees((c) => { const n = { ...c }; delete n[instanceId]; return n; });
    setWidgetErrors((c) => { const n = { ...c }; delete n[instanceId]; return n; });
    if (focusedInstanceId === instanceId) dispatch(focusInstance(null));
  }, [dispatch, focusedInstanceId]);

  const handleEvent = React.useCallback(
    async (instanceId: string, widgetId: string, eventRef: UIEventRef, eventPayload?: unknown) => {
      try {
        const pluginState = pluginStateById[instanceId] ?? {};
        const globalState = selectGlobalStateForInstance(rootState, instanceId);
        const intents = await quickjsSandboxClient.event(
          instanceId, widgetId, eventRef.handler, eventPayload ?? eventRef.args, pluginState, globalState,
        );
        for (const intent of intents) {
          if (intent.scope === "plugin") {
            dispatchPluginAction(dispatch, instanceId, intent.actionType, intent.payload);
          } else if (intent.domain) {
            dispatchSharedAction(dispatch, instanceId, intent.domain as SharedDomainName, intent.actionType, intent.payload);
          }
        }
      } catch (err) {
        dispatch(pushError({ kind: "event", instanceId, widgetId, message: String(err) }));
      }
    },
    [dispatch, pluginStateById, rootState],
  );

  // -----------------------------------------------------------------------
  // Derived data for components
  // -----------------------------------------------------------------------

  const running: RunningInstance[] = React.useMemo(() =>
    loadedPluginIds.map((id) => ({
      instanceId: id,
      title: plugins[id]?.title ?? "Plugin",
      packageId: plugins[id]?.packageId ?? "unknown",
      shortId: id.length > 10 ? id.slice(0, 10) : id,
      status: (plugins[id]?.status ?? "loaded") as "loaded" | "error",
      readGrants: grantsByInstance[id]?.readShared ?? [],
      writeGrants: grantsByInstance[id]?.writeShared ?? [],
    })),
    [loadedPluginIds, plugins, grantsByInstance],
  );

  const health: HealthStatus = errors.length > 5 ? "error" : errors.length > 0 ? "degraded" : "healthy";

  const editorTabInfos: EditorTabInfo[] = editorTabs.map((t) => ({
    id: t.id,
    label: t.label,
    dirty: t.dirty,
  }));

  const activeTab = editorTabs.find((t) => t.id === activeEditorTabId);

  const timelineEntries: TimelineEntry[] = React.useMemo(() =>
    dispatchTimeline.map((d: DispatchTimelineEntry, i: number) => ({
      id: `tl-${i}`,
      timestamp: d.timestamp,
      scope: d.scope as "plugin" | "shared",
      outcome: d.outcome as "applied" | "denied" | "ignored",
      actionType: d.actionType,
      instanceId: d.instanceId ?? "",
      shortInstanceId: (d.instanceId ?? "").length > 10 ? (d.instanceId ?? "").slice(0, 10) : (d.instanceId ?? ""),
      domain: d.domain ?? undefined,
      reason: d.reason ?? undefined,
    })),
    [dispatchTimeline],
  );

  const instanceStates: InstanceState[] = React.useMemo(() =>
    loadedPluginIds.map((id) => ({
      instanceId: id,
      title: plugins[id]?.title ?? "Plugin",
      shortId: id.length > 10 ? id.slice(0, 10) : id,
      state: pluginStateById[id] ?? {},
    })),
    [loadedPluginIds, plugins, pluginStateById],
  );

  const instanceCapabilities: InstanceCapabilities[] = React.useMemo(() =>
    loadedPluginIds.map((id) => {
      const g = grantsByInstance[id] ?? { readShared: [], writeShared: [] };
      const grants: Record<string, { read: boolean; write: boolean }> = {};
      for (const d of ALL_SHARED_DOMAINS) {
        grants[d] = {
          read: g.readShared?.includes(d) ?? false,
          write: g.writeShared?.includes(d) ?? false,
        };
      }
      return {
        instanceId: id,
        title: plugins[id]?.title ?? "Plugin",
        shortId: id.length > 10 ? id.slice(0, 10) : id,
        grants,
      };
    }),
    [loadedPluginIds, plugins, grantsByInstance],
  );

  const sharedDomains: SharedDomainInfo[] = React.useMemo(() =>
    ALL_SHARED_DOMAINS.map((name) => ({
      name,
      state: (sharedStateSnapshot as any)?.[name] ?? {},
      readers: loadedPluginIds
        .filter((id) => grantsByInstance[id]?.readShared?.includes(name))
        .map((id) => ({
          instanceId: id,
          title: plugins[id]?.title ?? "Plugin",
          shortId: id.length > 10 ? id.slice(0, 10) : id,
        })),
      writers: loadedPluginIds
        .filter((id) => grantsByInstance[id]?.writeShared?.includes(name))
        .map((id) => ({
          instanceId: id,
          title: plugins[id]?.title ?? "Plugin",
          shortId: id.length > 10 ? id.slice(0, 10) : id,
        })),
    })),
    [sharedStateSnapshot, loadedPluginIds, plugins, grantsByInstance],
  );

  // -----------------------------------------------------------------------
  // DevTools tab content
  // -----------------------------------------------------------------------

  const devtoolsContent: Record<DevToolsTab, React.ReactNode> = {
    timeline: <TimelinePanel entries={timelineEntries} focusedInstanceId={focusedInstanceId} />,
    state: <StatePanel instances={instanceStates} focusedInstanceId={focusedInstanceId} onFocusInstance={(id) => dispatch(focusInstance(id))} />,
    capabilities: <CapabilitiesPanel domains={ALL_SHARED_DOMAINS} instances={instanceCapabilities} focusedInstanceId={focusedInstanceId} />,
    errors: <ErrorsPanel errors={errors} onClear={() => dispatch(clearErrors())} />,
    shared: <SharedDomainsPanel domains={sharedDomains} />,
    docs: <DocsPanel docs={docEntries} allDocsMarkdown={allDocsMarkdown} />,
  };

  // -----------------------------------------------------------------------
  // Render
  // -----------------------------------------------------------------------

  return (
    <WorkbenchLayout
      toolbar={
        <TopToolbar
          pluginCount={loadedPluginIds.length}
          dispatchCount={dispatchTimeline.length}
          health={health}
          errorCount={errors.length}
          onHealthClick={() => dispatch(setActiveDevToolsTab("errors"))}
        />
      }
      sidebar={
        <Sidebar
          catalog={CATALOG}
          running={running}
          collapsed={sidebarCollapsed}
          focusedInstanceId={focusedInstanceId ?? undefined}
          onToggleCollapse={() => dispatch(toggleSidebar())}
          onFocusInstance={(id) => dispatch(focusInstance(id))}
          onLoadPreset={openPreset}
          onUnloadInstance={(id) => void unloadPlugin(id)}
          onNewPlugin={() => dispatch(openEditorTab({ packageId: "custom", label: "untitled.js", code: "" }))}
        />
      }
      devtools={
        <DevToolsPanel
          activeTab={activeDevTab}
          collapsed={devtoolsCollapsed}
          errorCount={errors.length}
          onSelectTab={(tab) => dispatch(setActiveDevToolsTab(tab))}
          onToggleCollapse={() => dispatch(toggleDevtools())}
        >
          {devtoolsContent[activeDevTab]}
        </DevToolsPanel>
      }
    >
      {/* Main content: editor + preview split */}
      <div className="h-full flex">
        {/* Editor side */}
        <div className="flex-[6] border-r border-white/[0.06] flex flex-col min-w-0">
          <EditorTabBar
            tabs={editorTabInfos}
            activeTabId={activeEditorTabId}
            onSelectTab={(id) => dispatch(setActiveEditorTab(id))}
            onCloseTab={(id) => dispatch(closeEditorTab(id))}
            onRun={activeTab ? () => void runEditorTab(activeTab.code, activeTab.packageId) : undefined}
            onReload={activeTab ? () => void runEditorTab(activeTab.code, activeTab.packageId) : undefined}
          />
          {activeTab ? (
            <CodeEditor
              value={activeTab.code}
              onChange={(code) => dispatch(updateEditorCode({ tabId: activeTab.id, code }))}
            />
          ) : (
            <div className="flex-1 flex items-center justify-center text-sm text-slate-600">
              Open a plugin from the sidebar to start editing.
            </div>
          )}
        </div>

        {/* Preview side */}
        <div className="flex-[4] min-w-0">
          <LivePreview>
            {loadedPluginIds.map((instanceId) => {
              const meta = pluginMetaById[instanceId];
              const trees = widgetTrees[instanceId] ?? {};
              const errs = widgetErrors[instanceId] ?? {};
              const firstError = Object.values(errs)[0];

              return (
                <InstanceCard
                  key={instanceId}
                  instanceId={instanceId}
                  title={plugins[instanceId]?.title ?? "Plugin"}
                  shortId={instanceId.length > 10 ? instanceId.slice(0, 10) : instanceId}
                  status={plugins[instanceId]?.status === "error" ? "error" : "loaded"}
                  focused={focusedInstanceId === instanceId}
                  errorMessage={firstError}
                  onFocus={() => dispatch(focusInstance(instanceId))}
                  onUnload={() => void unloadPlugin(instanceId)}
                >
                  {meta?.widgets.map((widgetId) => {
                    const widgetTree = trees[widgetId];
                    if (!widgetTree) return null;
                    return (
                      <WidgetRenderer
                        key={widgetId}
                        tree={widgetTree}
                        onEvent={(eventRef, payload) => void handleEvent(instanceId, widgetId, eventRef, payload)}
                      />
                    );
                  })}
                </InstanceCard>
              );
            })}
          </LivePreview>
        </div>
      </div>
    </WorkbenchLayout>
  );
}
