/**
 * Shared Storybook decorators and mock data for workbench stories.
 *
 * Three layers:
 *  1. createStoryStore()  — configurable RTK store with runtime stub
 *  2. withStore()         — Storybook decorator wrapping stories in <Provider>
 *  3. MOCK_*              — pre-populated fixture data for interactive stories
 */
import React from "react";
import { Provider } from "react-redux";
import { configureStore } from "@reduxjs/toolkit";
import type { Decorator } from "@storybook/react-vite";
import workbenchReducer, { type WorkbenchState } from "@/store/workbenchSlice";

// ---------------------------------------------------------------------------
// Mock runtime state — rich enough that selectors return useful data
// ---------------------------------------------------------------------------

export const MOCK_PLUGINS = {
  "abc-1234-5678": {
    instanceId: "abc-1234-5678",
    packageId: "counter",
    title: "Counter",
    description: "Local counter + shared summary",
    widgets: ["main"],
    status: "loaded",
  },
  "def-5678-9012": {
    instanceId: "def-5678-9012",
    packageId: "greeter",
    title: "Greeter",
    description: "Input handling demo",
    widgets: ["main"],
    status: "loaded",
  },
  "ghi-9012-3456": {
    instanceId: "ghi-9012-3456",
    packageId: "status-dashboard",
    title: "Status Dashboard",
    description: "Runtime metrics overview",
    widgets: ["main"],
    status: "loaded",
  },
} as Record<string, any>;

export const MOCK_GRANTS = {
  "abc-1234-5678": {
    readShared: ["counter-summary"],
    writeShared: ["counter-summary"],
    systemCommands: [],
  },
  "def-5678-9012": {
    readShared: ["greeter-profile"],
    writeShared: ["greeter-profile"],
    systemCommands: [],
  },
  "ghi-9012-3456": {
    readShared: ["counter-summary", "runtime-metrics", "runtime-registry"],
    writeShared: [],
    systemCommands: [],
  },
} as Record<string, any>;

export const MOCK_PLUGIN_STATE = {
  "abc-1234-5678": { count: 5 },
  "def-5678-9012": { name: "Alice", greeting: "Hello, Alice!" },
  "ghi-9012-3456": { metrics: { totalPlugins: 3, totalDispatches: 47 } },
} as Record<string, unknown>;

const now = Date.now();

export const MOCK_TIMELINE = [
  { id: "d1", timestamp: now - 1000, scope: "plugin", outcome: "applied", actionType: "increment", instanceId: "abc-1234-5678", domain: null, reason: null },
  { id: "d2", timestamp: now - 1000, scope: "shared", outcome: "applied", actionType: "set-instance", instanceId: "abc-1234-5678", domain: "counter-summary", reason: null },
  { id: "d3", timestamp: now - 2200, scope: "shared", outcome: "denied", actionType: "set-name", instanceId: "def-5678-9012", domain: "greeter-profile", reason: "No write grant" },
  { id: "d4", timestamp: now - 3500, scope: "plugin", outcome: "applied", actionType: "reset", instanceId: "abc-1234-5678", domain: null, reason: null },
  { id: "d5", timestamp: now - 5000, scope: "shared", outcome: "applied", actionType: "set-name", instanceId: "ghi-9012-3456", domain: "greeter-profile", reason: null },
  { id: "d6", timestamp: now - 7000, scope: "plugin", outcome: "ignored", actionType: "unknown-action", instanceId: "def-5678-9012", domain: null, reason: null },
] as any[];

export const MOCK_ERRORS = [
  { id: "e1", timestamp: now - 1200, kind: "render" as const, instanceId: "abc-1234-5678", widgetId: "main", message: "TypeError: Cannot read properties of undefined (reading 'render')" },
  { id: "e2", timestamp: now - 4500, kind: "event" as const, instanceId: "def-5678-9012", widgetId: null, message: "ReferenceError: x is not defined\n    at plugin.js:8:3" },
];

function createRuntimeState(overrides?: Partial<{
  plugins: Record<string, any>;
  pluginStateById: Record<string, unknown>;
  grantsByInstance: Record<string, any>;
  dispatchTimeline: any[];
}>) {
  return {
    plugins: overrides?.plugins ?? {},
    pluginStateById: overrides?.pluginStateById ?? {},
    grantsByInstance: overrides?.grantsByInstance ?? {},
    shared: {
      "counter-summary": { valuesByInstance: { "abc-1234-5678": 5 }, totalValue: 5, instanceCount: 1, lastUpdatedInstanceId: "abc-1234-5678" },
      "greeter-profile": { name: "Alice", lastUpdatedInstanceId: "def-5678-9012" },
      "runtime-registry": [],
      "runtime-metrics": { totalPlugins: 3, totalDispatches: 47 },
    },
    dispatchTrace: {
      count: overrides?.dispatchTimeline?.length ?? 0,
      lastTimestamp: null,
      lastDispatchId: null,
      lastScope: null,
      lastActionType: null,
      lastOutcome: null,
      lastReason: null,
    },
    dispatchTimeline: overrides?.dispatchTimeline ?? [],
  };
}

function runtimeStubReducer(state: any, _action: any) {
  return state ?? createRuntimeState();
}

// ---------------------------------------------------------------------------
// Store factory
// ---------------------------------------------------------------------------

export interface StoryStoreOptions {
  workbench?: Partial<WorkbenchState>;
  runtime?: Partial<Parameters<typeof createRuntimeState>[0]>;
}

export function createStoryStore(opts?: StoryStoreOptions) {
  const workbenchInitial: WorkbenchState = {
    sidebarCollapsed: false,
    devtoolsCollapsed: false,
    activeDevToolsTab: "timeline",
    focusedInstanceId: null,
    editorTabs: [],
    activeEditorTabId: null,
    errors: [],
    ...opts?.workbench,
  };

  return configureStore({
    reducer: {
      runtime: runtimeStubReducer as any,
      workbench: workbenchReducer,
    },
    preloadedState: {
      runtime: createRuntimeState(opts?.runtime) as any,
      workbench: workbenchInitial,
    },
  });
}

/**
 * Pre-populated store with 3 plugins, grants, timeline, 2 errors, editor tabs.
 * Good base for interactive stories.
 */
export function createPopulatedStore(overrides?: Partial<WorkbenchState>) {
  return createStoryStore({
    runtime: {
      plugins: MOCK_PLUGINS,
      pluginStateById: MOCK_PLUGIN_STATE,
      grantsByInstance: MOCK_GRANTS,
      dispatchTimeline: MOCK_TIMELINE,
    },
    workbench: {
      focusedInstanceId: "abc-1234-5678",
      errors: MOCK_ERRORS,
      editorTabs: [
        { id: "tab-1", label: "counter.js", packageId: "counter", code: "definePlugin(({ ui }) => { ... })", dirty: false, activeInstanceId: "abc-1234-5678" },
        { id: "tab-2", label: "greeter.js", packageId: "greeter", code: "definePlugin(({ ui }) => { ... })", dirty: true, activeInstanceId: "def-5678-9012" },
      ],
      activeEditorTabId: "tab-1",
      ...overrides,
    },
  });
}

// ---------------------------------------------------------------------------
// Decorators
// ---------------------------------------------------------------------------

/**
 * Basic decorator — each story gets a fresh default store.
 * Optionally pre-seed workbench state.
 */
export function withStore(overrides?: Partial<WorkbenchState>): Decorator {
  return (Story) => {
    const store = React.useMemo(() => createStoryStore({ workbench: overrides }), []);
    return (
      <Provider store={store}>
        <Story />
      </Provider>
    );
  };
}

/**
 * Decorator with a fully populated mock store (3 plugins, timeline, errors).
 * Good for interactive stories where you want data already in the store.
 */
export function withPopulatedStore(overrides?: Partial<WorkbenchState>): Decorator {
  return (Story) => {
    const store = React.useMemo(() => createPopulatedStore(overrides), []);
    return (
      <Provider store={store}>
        <Story />
      </Provider>
    );
  };
}
