/**
 * App-level Redux store.
 *
 * Composes the reusable runtime reducer (from packages/plugin-runtime)
 * with the app-specific workbench UI reducer.
 */
import { configureStore } from "@reduxjs/toolkit";
import { useDispatch, useSelector } from "react-redux";
import { runtimeReducer } from "@runtime/redux-adapter/store";
import workbenchReducer from "./workbenchSlice";

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

export const store = configureStore({
  reducer: {
    runtime: runtimeReducer,
    workbench: workbenchReducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

// Typed hooks
export const useAppDispatch = useDispatch.withTypes<AppDispatch>();
export const useAppSelector = useSelector.withTypes<RootState>();

// ---------------------------------------------------------------------------
// Re-export runtime selectors and actions so consumers import from one place.
// The runtime selectors expect { runtime: ... } which our RootState satisfies.
// ---------------------------------------------------------------------------

export {
  pluginRegistered,
  pluginRemoved,
  pluginActionDispatched,
  sharedActionDispatched,
  selectPluginState,
  selectAllPluginState,
  selectLoadedPluginIds,
  selectDispatchTimeline,
  selectGlobalState,
  selectGlobalStateForInstance,
  dispatchPluginAction,
  dispatchSharedAction,
} from "@runtime/redux-adapter/store";

export type {
  DispatchTimelineEntry,
  CapabilityGrants,
  SharedDomainName,
  RuntimePlugin,
  PluginStatus,
  DispatchOutcome,
} from "@runtime/redux-adapter/store";

// Re-export workbench actions and types
export {
  toggleSidebar,
  setSidebarCollapsed,
  toggleDevtools,
  setDevtoolsCollapsed,
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
} from "./workbenchSlice";

export type {
  WorkbenchState,
  DevToolsTab,
  EditorTab,
  ErrorEntry,
} from "./workbenchSlice";
