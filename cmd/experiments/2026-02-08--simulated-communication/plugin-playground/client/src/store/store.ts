import { configureStore, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { nanoid } from "nanoid";

export type PluginStatus = "loaded" | "error";

export interface RuntimePlugin {
  id: string;
  title: string;
  description?: string;
  widgets: string[];
  enabled: boolean;
  status: PluginStatus;
  error?: string;
}

interface RuntimeState {
  plugins: Record<string, RuntimePlugin>;
  pluginStateById: Record<string, unknown>;
  globals: {
    counterValue: number;
    greeterName: string;
  };
  dispatchTrace: {
    count: number;
    lastDispatchId: string | null;
    lastScope: "plugin" | "global" | null;
    lastActionType: string | null;
  };
}

interface ScopedDispatchPayload {
  dispatchId: string;
  timestamp: number;
  scope: "plugin" | "global";
  actionType: string;
  pluginId?: string;
  payload?: unknown;
}

const ALLOWED_GLOBAL_ACTION_TYPES = new Set(["counter/set"]);

const initialState: RuntimeState = {
  plugins: {},
  pluginStateById: {},
  globals: {
    counterValue: 0,
    greeterName: "",
  },
  dispatchTrace: {
    count: 0,
    lastDispatchId: null,
    lastScope: null,
    lastActionType: null,
  },
};

function markDispatch(state: RuntimeState, dispatch: ScopedDispatchPayload) {
  state.dispatchTrace.count += 1;
  state.dispatchTrace.lastDispatchId = dispatch.dispatchId;
  state.dispatchTrace.lastScope = dispatch.scope;
  state.dispatchTrace.lastActionType = dispatch.actionType;
}

function reduceCounterPlugin(
  state: RuntimeState,
  pluginId: string,
  actionType: string,
  payload?: unknown
) {
  const current = (state.pluginStateById[pluginId] as { value?: number } | undefined) ?? {};
  let value = Number(current.value ?? 0);

  if (actionType === "increment") {
    value += 1;
  } else if (actionType === "decrement") {
    value -= 1;
  } else if (actionType === "reset") {
    value = 0;
  } else {
    return;
  }

  state.pluginStateById[pluginId] = { value };
  state.globals.counterValue = value;
}

function reduceCalculatorPlugin(
  state: RuntimeState,
  pluginId: string,
  actionType: string,
  payload?: unknown
) {
  const current =
    (state.pluginStateById[pluginId] as
      | { display?: string; accumulator?: number; operation?: string | null }
      | undefined) ?? {};

  const calculator = {
    display: String(current.display ?? "0"),
    accumulator: Number(current.accumulator ?? 0),
    operation: (current.operation ?? null) as string | null,
  };

  if (actionType === "digit") {
    const digit = String(payload ?? "");
    calculator.display = calculator.display === "0" ? digit : calculator.display + digit;
  } else if (actionType === "clear") {
    calculator.display = "0";
    calculator.accumulator = 0;
    calculator.operation = null;
  } else if (actionType === "operation") {
    calculator.accumulator = parseFloat(calculator.display);
    calculator.operation = String(payload ?? "");
    calculator.display = "0";
  } else if (actionType === "equals") {
    const currentValue = parseFloat(calculator.display);
    let result = currentValue;

    if (calculator.operation === "+") {
      result = calculator.accumulator + currentValue;
    } else if (calculator.operation === "-") {
      result = calculator.accumulator - currentValue;
    } else if (calculator.operation === "*") {
      result = calculator.accumulator * currentValue;
    } else if (calculator.operation === "/") {
      result = calculator.accumulator / currentValue;
    }

    calculator.display = String(result);
    calculator.accumulator = 0;
    calculator.operation = null;
  } else {
    return;
  }

  state.pluginStateById[pluginId] = calculator;
}

function reduceGreeterPlugin(
  state: RuntimeState,
  pluginId: string,
  actionType: string,
  payload?: unknown
) {
  if (actionType !== "nameChanged") {
    return;
  }

  const name = String(payload ?? "");
  state.pluginStateById[pluginId] = {
    name,
  };
  state.globals.greeterName = name;
}

function reducePluginScopedAction(
  state: RuntimeState,
  pluginId: string,
  actionType: string,
  payload?: unknown
) {
  if (!state.plugins[pluginId]) {
    return;
  }

  if (pluginId === "counter") {
    reduceCounterPlugin(state, pluginId, actionType, payload);
    return;
  }

  if (pluginId === "calculator") {
    reduceCalculatorPlugin(state, pluginId, actionType, payload);
    return;
  }

  if (pluginId === "greeter") {
    reduceGreeterPlugin(state, pluginId, actionType, payload);
  }
}

function reduceGlobalScopedAction(state: RuntimeState, actionType: string, payload?: unknown) {
  if (actionType === "counter/set") {
    state.globals.counterValue = Number(payload ?? 0);
  }
}

const runtimeSlice = createSlice({
  name: "runtime",
  initialState,
  reducers: {
    pluginRegistered(
      state,
      action: PayloadAction<{
        id: string;
        title: string;
        description?: string;
        widgets: string[];
        initialState?: unknown;
      }>
    ) {
      const { id, title, description, widgets, initialState } = action.payload;

      state.plugins[id] = {
        id,
        title,
        description,
        widgets,
        enabled: true,
        status: "loaded",
      };

      if (initialState !== undefined) {
        state.pluginStateById[id] = initialState;
      }

      if (id === "counter") {
        const counter = (state.pluginStateById[id] as { value?: number } | undefined) ?? { value: 0 };
        state.globals.counterValue = Number(counter.value ?? 0);
      }

      if (id === "greeter") {
        const greeter = (state.pluginStateById[id] as { name?: string } | undefined) ?? { name: "" };
        state.globals.greeterName = String(greeter.name ?? "");
      }
    },

    pluginRemoved(state, action: PayloadAction<string>) {
      delete state.plugins[action.payload];
      delete state.pluginStateById[action.payload];

      if (action.payload === "counter") {
        state.globals.counterValue = 0;
      }

      if (action.payload === "greeter") {
        state.globals.greeterName = "";
      }
    },

    pluginActionDispatched: {
      reducer(state, action: PayloadAction<ScopedDispatchPayload>) {
        const { pluginId, actionType, payload } = action.payload;
        if (!pluginId) {
          return;
        }

        markDispatch(state, action.payload);
        reducePluginScopedAction(state, pluginId, actionType, payload);
      },
      prepare(pluginId: string, actionType: string, payload?: unknown) {
        return {
          payload: {
            dispatchId: nanoid(),
            timestamp: Date.now(),
            scope: "plugin" as const,
            pluginId,
            actionType,
            payload,
          },
        };
      },
    },

    globalActionDispatched: {
      reducer(state, action: PayloadAction<ScopedDispatchPayload>) {
        if (!ALLOWED_GLOBAL_ACTION_TYPES.has(action.payload.actionType)) {
          return;
        }

        markDispatch(state, action.payload);
        reduceGlobalScopedAction(state, action.payload.actionType, action.payload.payload);
      },
      prepare(actionType: string, payload?: unknown) {
        return {
          payload: {
            dispatchId: nanoid(),
            timestamp: Date.now(),
            scope: "global" as const,
            actionType,
            payload,
          },
        };
      },
    },
  },
});

export const {
  pluginRegistered,
  pluginRemoved,
  pluginActionDispatched,
  globalActionDispatched,
} = runtimeSlice.actions;

export const store = configureStore({
  reducer: {
    runtime: runtimeSlice.reducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export function selectLoadedPluginIds(state: RootState): string[] {
  return Object.keys(state.runtime.plugins);
}

export function selectPluginState(state: RootState, pluginId: string): unknown {
  return state.runtime.pluginStateById[pluginId] ?? {};
}

export function selectAllPluginState(state: RootState): Record<string, unknown> {
  return state.runtime.pluginStateById;
}

export function selectGlobalState(state: RootState) {
  return {
    counterValue: state.runtime.globals.counterValue,
    greeterName: state.runtime.globals.greeterName,
    pluginCount: Object.keys(state.runtime.plugins).length,
    dispatchCount: state.runtime.dispatchTrace.count,
    lastDispatchId: state.runtime.dispatchTrace.lastDispatchId,
    lastScope: state.runtime.dispatchTrace.lastScope,
    lastActionType: state.runtime.dispatchTrace.lastActionType,
    plugins: Object.values(state.runtime.plugins).map((p) => ({
      id: p.id,
      title: p.title,
      status: p.status,
      enabled: p.enabled,
      widgets: p.widgets.length,
    })),
  };
}

export function dispatchPluginAction(
  dispatch: AppDispatch,
  pluginId: string,
  actionType: string,
  payload?: unknown
) {
  dispatch(pluginActionDispatched(pluginId, actionType, payload));
}

export function dispatchGlobalAction(dispatch: AppDispatch, actionType: string, payload?: unknown) {
  if (!ALLOWED_GLOBAL_ACTION_TYPES.has(actionType)) {
    throw new Error(`Global action not allowed: ${actionType}`);
  }

  dispatch(globalActionDispatched(actionType, payload));
}
