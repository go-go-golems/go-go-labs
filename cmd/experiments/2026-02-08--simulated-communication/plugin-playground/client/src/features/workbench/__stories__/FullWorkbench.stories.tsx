/**
 * FullWorkbench — interactive composed story.
 *
 * Wires all workbench components together with a live RTK store.
 * Click catalog items to "load" plugins, click instances to focus them,
 * toggle sidebar/devtools, switch devtools tabs — everything works.
 */
import React from "react";
import type { Meta, StoryObj } from "@storybook/react-vite";
import { useDispatch, useSelector } from "react-redux";

import { WorkbenchLayout } from "../components/WorkbenchLayout";
import { Sidebar, type CatalogEntry, type RunningInstance } from "../components/Sidebar";
import { TopToolbar, type HealthStatus } from "../components/TopToolbar";
import { EditorTabBar, type EditorTabInfo } from "../components/EditorTabBar";
import { CodeEditor } from "../components/CodeEditor";
import { LivePreview } from "../components/LivePreview";
import { InstanceCard } from "../components/InstanceCard";
import { DevToolsPanel } from "../components/DevToolsPanel";
import { TimelinePanel, type TimelineEntry } from "../components/TimelinePanel";
import { StatePanel, type InstanceState } from "../components/StatePanel";
import { CapabilitiesPanel, type InstanceCapabilities } from "../components/CapabilitiesPanel";
import { ErrorsPanel } from "../components/ErrorsPanel";
import { SharedDomainsPanel, type SharedDomainInfo } from "../components/SharedDomainsPanel";
import { DocsPanel } from "../components/DocsPanel";

import {
  toggleSidebar,
  toggleDevtools,
  setActiveDevToolsTab,
  focusInstance,
  openEditorTab,
  closeEditorTab,
  setActiveEditorTab,
  updateEditorCode,
  pushError,
  clearErrors,
  type DevToolsTab,
  type ErrorEntry,
} from "@/store/workbenchSlice";

import { withPopulatedStore } from "./storyDecorators";

// ---------------------------------------------------------------------------
// Catalog
// ---------------------------------------------------------------------------

const CATALOG: CatalogEntry[] = [
  { id: "counter", title: "Counter", description: "Local counter + shared summary", capabilitySummary: "R/W" },
  { id: "calculator", title: "Calculator", description: "Basic arithmetic" },
  { id: "status-dashboard", title: "Status Dashboard", description: "Runtime metrics", capabilitySummary: "R" },
  { id: "greeter", title: "Greeter", description: "Input handling demo", capabilitySummary: "R/W" },
  { id: "greeter-shared-state", title: "Greeter Shared", description: "Reads shared profile", capabilitySummary: "R" },
  { id: "runtime-monitor", title: "Runtime Monitor", description: "Plugin registry table", capabilitySummary: "R" },
];

const SAMPLE_CODE: Record<string, string> = {
  counter: `definePlugin(({ ui, shared }) => {
  let count = 0;
  return {
    id: "counter",
    title: "Counter",
    capabilities: {
      readShared: ["counter-summary"],
      writeShared: ["counter-summary"],
    },
    reduceEvent(state, event) {
      if (event.type === "increment") count++;
      if (event.type === "decrement") count--;
      if (event.type === "reset") count = 0;
      return { count };
    },
    render(state) {
      return ui.column([
        ui.text(\`Counter: \${state.count}\`),
        ui.row([
          ui.button("−", { event: { type: "decrement" } }),
          ui.button("Reset", { event: { type: "reset" } }),
          ui.button("+", { event: { type: "increment" } }),
        ]),
      ]);
    },
  };
});`,
  greeter: `definePlugin(({ ui }) => ({
  id: "greeter",
  title: "Greeter",
  render(state) {
    return ui.column([
      ui.input({ value: state?.name ?? "", placeholder: "Your name" }),
      ui.text(\`Hello, \${state?.name ?? "world"}!\`),
    ]);
  },
}));`,
};

const ALL_SHARED_DOMAINS = ["counter-summary", "greeter-profile", "runtime-registry", "runtime-metrics"];

const FIXTURE_DOCS = [
  { title: "Overview", category: "Overview", path: "docs/README.md", raw: "# Plugin Playground\n\nA browser sandbox for developing and testing plugins.\n\n## Features\n- Hot-reload plugin code\n- Shared state across instances\n- Capability-based security" },
  { title: "Quickstart", category: "Plugin Authoring", path: "docs/quickstart.md", raw: "# Quickstart\n\nCall `definePlugin()` to create a plugin.\n\n```js\ndefinePlugin(({ ui }) => ({\n  id: 'hello',\n  render: () => ui.text('Hello!'),\n}));\n```" },
  { title: "Capabilities", category: "Architecture", path: "docs/capabilities.md", raw: "# Capability Model\n\nPlugins declare read/write access to shared domains.\n\n| Domain | Purpose |\n|--------|----------|\n| counter-summary | Aggregate counter values |" },
];

// ---------------------------------------------------------------------------
// Connected workbench (reads/writes RTK store)
// ---------------------------------------------------------------------------

function ConnectedWorkbench() {
  const dispatch = useDispatch();

  // Workbench state from store
  const wb = useSelector((s: any) => s.workbench) as import("@/store/workbenchSlice").WorkbenchState;
  const runtime = useSelector((s: any) => s.runtime);

  // Local state for mock "loaded plugins" (since we don't have a real sandbox)
  const [mockPlugins, setMockPlugins] = React.useState<Record<string, {
    instanceId: string; title: string; packageId: string; status: "loaded" | "error";
    readGrants: string[]; writeGrants: string[];
    widgetContent: string;
  }>>({
    "abc-1234-5678": {
      instanceId: "abc-1234-5678", title: "Counter", packageId: "counter", status: "loaded",
      readGrants: ["counter-summary"], writeGrants: ["counter-summary"],
      widgetContent: "Counter: 5",
    },
    "def-5678-9012": {
      instanceId: "def-5678-9012", title: "Greeter", packageId: "greeter", status: "loaded",
      readGrants: ["greeter-profile"], writeGrants: ["greeter-profile"],
      widgetContent: "Hello, Alice!",
    },
    "ghi-9012-3456": {
      instanceId: "ghi-9012-3456", title: "Status Dashboard", packageId: "status-dashboard", status: "loaded",
      readGrants: ["counter-summary", "runtime-metrics", "runtime-registry"], writeGrants: [],
      widgetContent: "3 plugins loaded • 47 dispatches",
    },
  });

  let nextMockId = React.useRef(100);

  // Handlers
  const handleLoadPreset = (presetId: string) => {
    const preset = CATALOG.find((c) => c.id === presetId);
    if (!preset) return;
    const id = `mock-${++nextMockId.current}`;
    setMockPlugins((prev) => ({
      ...prev,
      [id]: {
        instanceId: id,
        title: preset.title,
        packageId: presetId,
        status: "loaded",
        readGrants: preset.capabilitySummary?.includes("R") ? ["counter-summary"] : [],
        writeGrants: preset.capabilitySummary?.includes("W") ? ["counter-summary"] : [],
        widgetContent: `${preset.title} output (mock)`,
      },
    }));
    dispatch(focusInstance(id));
    const code = SAMPLE_CODE[presetId] ?? `// ${preset.title} plugin code`;
    dispatch(openEditorTab({ packageId: presetId, label: `${preset.title.toLowerCase()}.js`, code }));
  };

  const handleUnload = (instanceId: string) => {
    setMockPlugins((prev) => {
      const next = { ...prev };
      delete next[instanceId];
      return next;
    });
    if (wb.focusedInstanceId === instanceId) dispatch(focusInstance(null));
  };

  // Derived data
  const running: RunningInstance[] = Object.values(mockPlugins).map((p) => ({
    instanceId: p.instanceId,
    title: p.title,
    packageId: p.packageId,
    shortId: p.instanceId.slice(0, 10),
    status: p.status,
    readGrants: p.readGrants,
    writeGrants: p.writeGrants,
  }));

  const health: HealthStatus = wb.errors.length > 5 ? "error" : wb.errors.length > 0 ? "degraded" : "healthy";

  const editorTabInfos: EditorTabInfo[] = wb.editorTabs.map((t: any) => ({
    id: t.id, label: t.label, dirty: t.dirty,
  }));
  const activeTab = wb.editorTabs.find((t: any) => t.id === wb.activeEditorTabId);

  // Mock timeline entries
  const now = Date.now();
  const timelineEntries: TimelineEntry[] = [
    { id: "t1", timestamp: now - 1000, scope: "plugin", outcome: "applied", actionType: "increment", instanceId: "abc-1234-5678", shortInstanceId: "abc-1234-5" },
    { id: "t2", timestamp: now - 1000, scope: "shared", outcome: "applied", actionType: "set-instance", instanceId: "abc-1234-5678", shortInstanceId: "abc-1234-5", domain: "counter-summary" },
    { id: "t3", timestamp: now - 2200, scope: "shared", outcome: "denied", actionType: "set-name", instanceId: "def-5678-9012", shortInstanceId: "def-5678-9", domain: "greeter-profile", reason: "No write grant" },
    { id: "t4", timestamp: now - 3500, scope: "plugin", outcome: "applied", actionType: "reset", instanceId: "abc-1234-5678", shortInstanceId: "abc-1234-5" },
  ];

  const instanceStates: InstanceState[] = Object.entries(mockPlugins).map(([id, p]) => ({
    instanceId: id, title: p.title, shortId: id.slice(0, 10),
    state: id === "abc-1234-5678" ? { count: 5 } : id === "def-5678-9012" ? { name: "Alice" } : { status: "ok" },
  }));

  const instanceCapabilities: InstanceCapabilities[] = Object.entries(mockPlugins).map(([id, p]) => {
    const grants: Record<string, { read: boolean; write: boolean }> = {};
    for (const d of ALL_SHARED_DOMAINS) {
      grants[d] = { read: p.readGrants.includes(d), write: p.writeGrants.includes(d) };
    }
    return { instanceId: id, title: p.title, shortId: id.slice(0, 10), grants };
  });

  const sharedDomains: SharedDomainInfo[] = ALL_SHARED_DOMAINS.map((name) => ({
    name,
    state: name === "counter-summary" ? { totalValue: 5, instanceCount: 1 } : name === "greeter-profile" ? { name: "Alice" } : {},
    readers: Object.values(mockPlugins).filter((p) => p.readGrants.includes(name)).map((p) => ({ instanceId: p.instanceId, title: p.title, shortId: p.instanceId.slice(0, 10) })),
    writers: Object.values(mockPlugins).filter((p) => p.writeGrants.includes(name)).map((p) => ({ instanceId: p.instanceId, title: p.title, shortId: p.instanceId.slice(0, 10) })),
  }));

  // DevTools content
  const devtoolsContent: Record<DevToolsTab, React.ReactNode> = {
    timeline: <TimelinePanel entries={timelineEntries} focusedInstanceId={wb.focusedInstanceId} />,
    state: <StatePanel instances={instanceStates} focusedInstanceId={wb.focusedInstanceId} onFocusInstance={(id) => dispatch(focusInstance(id))} />,
    capabilities: <CapabilitiesPanel domains={ALL_SHARED_DOMAINS} instances={instanceCapabilities} focusedInstanceId={wb.focusedInstanceId} />,
    errors: <ErrorsPanel errors={wb.errors} onClear={() => dispatch(clearErrors())} />,
    shared: <SharedDomainsPanel domains={sharedDomains} />,
    docs: <DocsPanel docs={FIXTURE_DOCS} allDocsMarkdown={FIXTURE_DOCS.map((d) => `# ${d.path}\n\n${d.raw}`).join("\n---\n")} />,
  };

  return (
    <WorkbenchLayout
      toolbar={
        <TopToolbar
          pluginCount={Object.keys(mockPlugins).length}
          dispatchCount={timelineEntries.length}
          health={health}
          errorCount={wb.errors.length}
          onHealthClick={() => dispatch(setActiveDevToolsTab("errors"))}
          onMenuClick={() => dispatch(pushError({ kind: "event", instanceId: null, widgetId: null, message: "Menu clicked (mock error to test error flow)" }))}
        />
      }
      sidebar={
        <Sidebar
          catalog={CATALOG}
          running={running}
          collapsed={wb.sidebarCollapsed}
          focusedInstanceId={wb.focusedInstanceId ?? undefined}
          onToggleCollapse={() => dispatch(toggleSidebar())}
          onFocusInstance={(id) => dispatch(focusInstance(id))}
          onLoadPreset={handleLoadPreset}
          onUnloadInstance={handleUnload}
          onNewPlugin={() => dispatch(openEditorTab({ packageId: "custom", label: "untitled.js", code: "" }))}
        />
      }
      devtools={
        <DevToolsPanel
          activeTab={wb.activeDevToolsTab}
          collapsed={wb.devtoolsCollapsed}
          errorCount={wb.errors.length}
          onSelectTab={(tab) => dispatch(setActiveDevToolsTab(tab))}
          onToggleCollapse={() => dispatch(toggleDevtools())}
        >
          {devtoolsContent[wb.activeDevToolsTab]}
        </DevToolsPanel>
      }
    >
      <div className="h-full flex">
        {/* Editor */}
        <div className="flex-[6] border-r border-white/[0.06] flex flex-col min-w-0">
          <EditorTabBar
            tabs={editorTabInfos}
            activeTabId={wb.activeEditorTabId}
            onSelectTab={(id) => dispatch(setActiveEditorTab(id))}
            onCloseTab={(id) => dispatch(closeEditorTab(id))}
            onRun={() => {}}
            onReload={() => {}}
          />
          {activeTab ? (
            <CodeEditor
              value={activeTab.code}
              onChange={(code) => dispatch(updateEditorCode({ tabId: activeTab.id, code }))}
            />
          ) : (
            <div className="flex-1 flex items-center justify-center text-sm text-slate-600">
              Click a catalog item to load a plugin and start editing.
            </div>
          )}
        </div>

        {/* Preview */}
        <div className="flex-[4] min-w-0">
          <LivePreview>
            {Object.values(mockPlugins).map((p) => (
              <InstanceCard
                key={p.instanceId}
                instanceId={p.instanceId}
                title={p.title}
                shortId={p.instanceId.slice(0, 10)}
                status={p.status}
                focused={wb.focusedInstanceId === p.instanceId}
                onFocus={() => dispatch(focusInstance(p.instanceId))}
                onUnload={() => handleUnload(p.instanceId)}
              >
                <div className="text-sm text-slate-300">{p.widgetContent}</div>
              </InstanceCard>
            ))}
          </LivePreview>
        </div>
      </div>
    </WorkbenchLayout>
  );
}

// ---------------------------------------------------------------------------
// Meta + story
// ---------------------------------------------------------------------------

const meta: Meta = {
  title: "Workbench/FullWorkbench",
  parameters: { layout: "fullscreen" },
};

export default meta;
type Story = StoryObj;

export const Interactive: Story = {
  decorators: [withPopulatedStore()],
  render: () => <ConnectedWorkbench />,
};

export const EmptyStart: Story = {
  decorators: [withPopulatedStore({
    editorTabs: [],
    activeEditorTabId: null,
    focusedInstanceId: null,
    errors: [],
  })],
  render: () => <ConnectedWorkbench />,
};
