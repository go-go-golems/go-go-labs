import { createSlice, type PayloadAction } from "@reduxjs/toolkit";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

export type DevToolsTab =
  | "timeline"
  | "state"
  | "capabilities"
  | "errors"
  | "shared"
  | "docs";

export interface EditorTab {
  id: string;
  /** Display label for the tab. */
  label: string;
  /** Package ID the tab originated from (preset id or "custom"). */
  packageId: string;
  /** Editor content. */
  code: string;
  /** Whether the code has unsaved changes vs. the last-run snapshot. */
  dirty: boolean;
  /** Instance ID of the currently-running instance for this tab (null if never run). */
  activeInstanceId: string | null;
}

export interface ErrorEntry {
  id: string;
  timestamp: number;
  kind: "load" | "render" | "event";
  instanceId: string | null;
  widgetId: string | null;
  message: string;
}

export interface WorkbenchState {
  // Layout
  sidebarCollapsed: boolean;
  devtoolsCollapsed: boolean;
  activeDevToolsTab: DevToolsTab;

  // Instance focus — the instance the user is inspecting
  focusedInstanceId: string | null;

  // Editor tabs
  editorTabs: EditorTab[];
  activeEditorTabId: string | null;

  // Error log
  errors: ErrorEntry[];
}

// ---------------------------------------------------------------------------
// Initial state
// ---------------------------------------------------------------------------

const initialState: WorkbenchState = {
  sidebarCollapsed: false,
  devtoolsCollapsed: false,
  activeDevToolsTab: "timeline",
  focusedInstanceId: null,
  editorTabs: [],
  activeEditorTabId: null,
  errors: [],
};

// ---------------------------------------------------------------------------
// Slice
// ---------------------------------------------------------------------------

const MAX_ERRORS = 200;
let nextTabCounter = 0;

const workbenchSlice = createSlice({
  name: "workbench",
  initialState,
  reducers: {
    // -- Layout ----------------------------------------------------------

    toggleSidebar(state) {
      state.sidebarCollapsed = !state.sidebarCollapsed;
    },

    setSidebarCollapsed(state, action: PayloadAction<boolean>) {
      state.sidebarCollapsed = action.payload;
    },

    toggleDevtools(state) {
      state.devtoolsCollapsed = !state.devtoolsCollapsed;
    },

    setDevtoolsCollapsed(state, action: PayloadAction<boolean>) {
      state.devtoolsCollapsed = action.payload;
    },

    setActiveDevToolsTab(state, action: PayloadAction<DevToolsTab>) {
      state.activeDevToolsTab = action.payload;
      // Opening a specific tab auto-expands the panel
      state.devtoolsCollapsed = false;
    },

    // -- Instance focus --------------------------------------------------

    focusInstance(state, action: PayloadAction<string | null>) {
      state.focusedInstanceId = action.payload;
    },

    // -- Editor tabs -----------------------------------------------------

    openEditorTab(
      state,
      action: PayloadAction<{
        packageId: string;
        label: string;
        code: string;
      }>
    ) {
      const id = `tab-${++nextTabCounter}-${Date.now()}`;
      const tab: EditorTab = {
        id,
        label: action.payload.label,
        packageId: action.payload.packageId,
        code: action.payload.code,
        dirty: false,
        activeInstanceId: null,
      };
      state.editorTabs.push(tab);
      state.activeEditorTabId = id;
    },

    closeEditorTab(state, action: PayloadAction<string>) {
      const tabId = action.payload;
      state.editorTabs = state.editorTabs.filter((t) => t.id !== tabId);
      if (state.activeEditorTabId === tabId) {
        state.activeEditorTabId =
          state.editorTabs.length > 0
            ? state.editorTabs[state.editorTabs.length - 1].id
            : null;
      }
    },

    setActiveEditorTab(state, action: PayloadAction<string>) {
      state.activeEditorTabId = action.payload;
    },

    updateEditorCode(
      state,
      action: PayloadAction<{ tabId: string; code: string }>
    ) {
      const tab = state.editorTabs.find((t) => t.id === action.payload.tabId);
      if (tab) {
        tab.code = action.payload.code;
        tab.dirty = true;
      }
    },

    markEditorTabClean(state, action: PayloadAction<string>) {
      const tab = state.editorTabs.find((t) => t.id === action.payload);
      if (tab) {
        tab.dirty = false;
      }
    },

    setTabActiveInstance(
      state,
      action: PayloadAction<{ tabId: string; instanceId: string | null }>
    ) {
      const tab = state.editorTabs.find(
        (t) => t.id === action.payload.tabId
      );
      if (tab) {
        tab.activeInstanceId = action.payload.instanceId;
      }
    },

    // -- Error log -------------------------------------------------------

    pushError(
      state,
      action: PayloadAction<Omit<ErrorEntry, "id" | "timestamp">>
    ) {
      state.errors.push({
        ...action.payload,
        id: `err-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
        timestamp: Date.now(),
      });
      if (state.errors.length > MAX_ERRORS) {
        state.errors.splice(0, state.errors.length - MAX_ERRORS);
      }
    },

    clearErrors(state) {
      state.errors = [];
    },
  },
});

export const {
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
} = workbenchSlice.actions;

export default workbenchSlice.reducer;
