// Design Philosophy: Technical Brutalism - Redux as the contract boundary
// All plugin state lives here, plugins can only influence via actions

import { configureStore, createSlice, PayloadAction } from "@reduxjs/toolkit";
import type { PluginMeta } from "@/lib/pluginSandboxClient";

export type PluginStatus = "idle" | "loading" | "loaded" | "error";

export interface LoadedPlugin {
  id: string;
  code: string;
  meta: PluginMeta;
  status: PluginStatus;
  error?: string;
  enabled: boolean;
}

interface PluginsState {
  plugins: Record<string, LoadedPlugin>;
  counter: number;
  calculator: {
    display: string;
    accumulator: number;
    operation: string | null;
  };
  greeter: {
    name: string;
  };
}

const initialState: PluginsState = {
  plugins: {},
  counter: 0,
  calculator: {
    display: "0",
    accumulator: 0,
    operation: null,
  },
  greeter: {
    name: "",
  },
};

const pluginsSlice = createSlice({
  name: "plugins",
  initialState,
  reducers: {
    pluginLoadStarted(state, action: PayloadAction<{ id: string; code: string }>) {
      state.plugins[action.payload.id] = {
        id: action.payload.id,
        code: action.payload.code,
        meta: { id: action.payload.id, widgets: [] },
        status: "loading",
        enabled: true,
      };
    },
    pluginLoadSucceeded(state, action: PayloadAction<{ id: string; meta: PluginMeta }>) {
      const plugin = state.plugins[action.payload.id];
      if (plugin) {
        plugin.status = "loaded";
        plugin.meta = action.payload.meta;
        plugin.error = undefined;
      }
    },
    pluginLoadFailed(state, action: PayloadAction<{ id: string; error: string }>) {
      const plugin = state.plugins[action.payload.id];
      if (plugin) {
        plugin.status = "error";
        plugin.error = action.payload.error;
      }
    },
    pluginToggled(state, action: PayloadAction<string>) {
      const plugin = state.plugins[action.payload];
      if (plugin) {
        plugin.enabled = !plugin.enabled;
      }
    },
    pluginRemoved(state, action: PayloadAction<string>) {
      delete state.plugins[action.payload];
    },
    pluginCodeUpdated(state, action: PayloadAction<{ id: string; code: string }>) {
      const plugin = state.plugins[action.payload.id];
      if (plugin) {
        plugin.code = action.payload.code;
        plugin.status = "idle";
      }
    },
  },
  extraReducers: (builder) => {
    builder.addMatcher(
      (action) => action.type.startsWith("plugin."),
      (state, action: any) => {
        if (action.type === "plugin.counter/incremented") {
          state.counter += 1;
        } else if (action.type === "plugin.counter/decremented") {
          state.counter -= 1;
        } else if (action.type === "plugin.counter/reset") {
          state.counter = 0;
        } else if (action.type === "plugin.calculator/digit") {
          const digit = action.payload;
          if (state.calculator.display === "0") {
            state.calculator.display = String(digit);
          } else {
            state.calculator.display += String(digit);
          }
        } else if (action.type === "plugin.calculator/clear") {
          state.calculator.display = "0";
          state.calculator.accumulator = 0;
          state.calculator.operation = null;
        } else if (action.type === "plugin.calculator/operation") {
          const op = action.payload;
          state.calculator.accumulator = parseFloat(state.calculator.display);
          state.calculator.operation = op;
          state.calculator.display = "0";
        } else if (action.type === "plugin.calculator/equals") {
          const current = parseFloat(state.calculator.display);
          let result = current;
          if (state.calculator.operation === "+") {
            result = state.calculator.accumulator + current;
          } else if (state.calculator.operation === "-") {
            result = state.calculator.accumulator - current;
          } else if (state.calculator.operation === "*") {
            result = state.calculator.accumulator * current;
          } else if (state.calculator.operation === "/") {
            result = state.calculator.accumulator / current;
          }
          state.calculator.display = String(result);
          state.calculator.accumulator = 0;
          state.calculator.operation = null;
        } else if (action.type === "plugin.greeter/nameChanged") {
          state.greeter.name = action.payload || "";
        }
      }
    );
  },
});

export const {
  pluginLoadStarted,
  pluginLoadSucceeded,
  pluginLoadFailed,
  pluginToggled,
  pluginRemoved,
  pluginCodeUpdated,
} = pluginsSlice.actions;

export const store = configureStore({
  reducer: {
    plugins: pluginsSlice.reducer,
    counter: (state: number = 0, action: any) => {
      if (action.type === "plugin.counter/incremented") return state + 1;
      if (action.type === "plugin.counter/decremented") return state - 1;
      if (action.type === "plugin.counter/reset") return 0;
      return state;
    },
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
