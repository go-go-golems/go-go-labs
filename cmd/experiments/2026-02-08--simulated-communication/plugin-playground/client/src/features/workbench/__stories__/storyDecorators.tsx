/**
 * Shared Storybook decorators for workbench stories.
 */
import React from "react";
import { Provider } from "react-redux";
import { configureStore } from "@reduxjs/toolkit";
import type { Decorator } from "@storybook/react-vite";
import workbenchReducer, { type WorkbenchState } from "@/store/workbenchSlice";

// ---------------------------------------------------------------------------
// Minimal runtime reducer stub for stories (no QuickJS dependency).
// Shape must match what selectors in components expect (state.runtime.*).
// ---------------------------------------------------------------------------

const runtimeInitialState = {
  plugins: {} as Record<string, any>,
  pluginStateById: {} as Record<string, unknown>,
  grantsByInstance: {} as Record<string, any>,
  shared: {
    "counter-summary": {
      valuesByInstance: {},
      totalValue: 0,
      instanceCount: 0,
      lastUpdatedInstanceId: null,
    },
    "greeter-profile": {
      name: "",
      lastUpdatedInstanceId: null,
    },
  },
  dispatchTrace: {
    count: 0,
    lastTimestamp: null,
    lastDispatchId: null,
    lastScope: null,
    lastActionType: null,
    lastOutcome: null,
    lastReason: null,
  },
  dispatchTimeline: [] as any[],
};

function runtimeStubReducer(state = runtimeInitialState, _action: any) {
  return state;
}

// ---------------------------------------------------------------------------
// Store factory
// ---------------------------------------------------------------------------

export function createStoryStore(workbenchOverrides?: Partial<WorkbenchState>) {
  const workbenchInitial: WorkbenchState = {
    sidebarCollapsed: false,
    devtoolsCollapsed: false,
    activeDevToolsTab: "timeline",
    focusedInstanceId: null,
    editorTabs: [],
    activeEditorTabId: null,
    errors: [],
    ...workbenchOverrides,
  };

  return configureStore({
    reducer: {
      runtime: runtimeStubReducer as any,
      workbench: workbenchReducer,
    },
    preloadedState: {
      runtime: runtimeInitialState as any,
      workbench: workbenchInitial,
    },
  });
}

// ---------------------------------------------------------------------------
// Decorator
// ---------------------------------------------------------------------------

/**
 * Returns a Storybook decorator that wraps stories in a Redux Provider.
 * Each story gets a fresh store so state doesn't leak between stories.
 *
 * Usage in meta:
 *   decorators: [withStore()]
 *   decorators: [withStore({ sidebarCollapsed: true })]
 */
export function withStore(overrides?: Partial<WorkbenchState>): Decorator {
  return (Story) => {
    // eslint-disable-next-line react-hooks/rules-of-hooks
    const store = React.useMemo(() => createStoryStore(overrides), []);
    return (
      <Provider store={store}>
        <Story />
      </Provider>
    );
  };
}
