import { configureStore, createSelector, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { nanoid } from "nanoid";
import type { InstanceId, PackageId } from "@/lib/quickjsContracts";

export type PluginStatus = "loaded" | "error";

export interface RuntimePlugin {
  instanceId: InstanceId;
  packageId: PackageId;
  title: string;
  description?: string;
  widgets: string[];
  enabled: boolean;
  status: PluginStatus;
  error?: string;
}

interface RuntimeState {
  plugins: Record<InstanceId, RuntimePlugin>;
  pluginStateById: Record<InstanceId, unknown>;
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
  instanceId?: InstanceId;
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
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
) {
  const current = (state.pluginStateById[instanceId] as { value?: number } | undefined) ?? {};
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

  state.pluginStateById[instanceId] = { value };
  state.globals.counterValue = value;
}

function reduceCalculatorPlugin(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
) {
  const current =
    (state.pluginStateById[instanceId] as
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

  state.pluginStateById[instanceId] = calculator;
}

function reduceGreeterPlugin(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
) {
  if (actionType !== "nameChanged") {
    return;
  }

  const name = String(payload ?? "");
  state.pluginStateById[instanceId] = {
    name,
  };
  state.globals.greeterName = name;
}

function recomputeCounterMirror(state: RuntimeState) {
  const counterInstances = Object.values(state.plugins).filter((plugin) => plugin.packageId === "counter");
  if (counterInstances.length === 0) {
    state.globals.counterValue = 0;
    return;
  }

  const last = counterInstances[counterInstances.length - 1];
  const local = state.pluginStateById[last.instanceId] as { value?: number } | undefined;
  state.globals.counterValue = Number(local?.value ?? 0);
}

function recomputeGreeterMirror(state: RuntimeState) {
  const greeterInstances = Object.values(state.plugins).filter((plugin) => plugin.packageId === "greeter");
  if (greeterInstances.length === 0) {
    state.globals.greeterName = "";
    return;
  }

  const last = greeterInstances[greeterInstances.length - 1];
  const local = state.pluginStateById[last.instanceId] as { name?: string } | undefined;
  state.globals.greeterName = String(local?.name ?? "");
}

function reducePluginScopedAction(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
) {
  const plugin = state.plugins[instanceId];
  if (!plugin) {
    return;
  }

  if (plugin.packageId === "counter") {
    reduceCounterPlugin(state, instanceId, actionType, payload);
  } else if (plugin.packageId === "calculator") {
    reduceCalculatorPlugin(state, instanceId, actionType, payload);
  } else if (plugin.packageId === "greeter") {
    reduceGreeterPlugin(state, instanceId, actionType, payload);
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
        instanceId: InstanceId;
        packageId: PackageId;
        title: string;
        description?: string;
        widgets: string[];
        initialState?: unknown;
      }>
    ) {
      const { instanceId, packageId, title, description, widgets, initialState } = action.payload;

      state.plugins[instanceId] = {
        instanceId,
        packageId,
        title,
        description,
        widgets,
        enabled: true,
        status: "loaded",
      };

      if (initialState !== undefined) {
        state.pluginStateById[instanceId] = initialState;
      }

      if (packageId === "counter") {
        const counter =
          (state.pluginStateById[instanceId] as { value?: number } | undefined) ?? { value: 0 };
        state.globals.counterValue = Number(counter.value ?? 0);
      }

      if (packageId === "greeter") {
        const greeter =
          (state.pluginStateById[instanceId] as { name?: string } | undefined) ?? { name: "" };
        state.globals.greeterName = String(greeter.name ?? "");
      }
    },

    pluginRemoved(state, action: PayloadAction<InstanceId>) {
      const instanceId = action.payload;
      const removed = state.plugins[instanceId];
      if (!removed) {
        return;
      }

      delete state.plugins[instanceId];
      delete state.pluginStateById[instanceId];

      if (removed.packageId === "counter") {
        recomputeCounterMirror(state);
      } else if (removed.packageId === "greeter") {
        recomputeGreeterMirror(state);
      }
    },

    pluginActionDispatched: {
      reducer(state, action: PayloadAction<ScopedDispatchPayload>) {
        const { instanceId, actionType, payload } = action.payload;
        if (!instanceId) {
          return;
        }

        markDispatch(state, action.payload);
        reducePluginScopedAction(state, instanceId, actionType, payload);
      },
      prepare(instanceId: InstanceId, actionType: string, payload?: unknown) {
        return {
          payload: {
            dispatchId: nanoid(),
            timestamp: Date.now(),
            scope: "plugin" as const,
            instanceId,
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

const selectRuntimeState = (state: RootState) => state.runtime;

export function selectPluginState(state: RootState, instanceId: InstanceId): unknown {
  return state.runtime.pluginStateById[instanceId] ?? {};
}

export function selectAllPluginState(state: RootState): Record<InstanceId, unknown> {
  return state.runtime.pluginStateById;
}

export const selectLoadedPluginIds = createSelector([selectRuntimeState], (runtime) =>
  Object.keys(runtime.plugins)
);

export const selectGlobalState = createSelector([selectRuntimeState], (runtime) => ({
  counterValue: runtime.globals.counterValue,
  greeterName: runtime.globals.greeterName,
  pluginCount: Object.keys(runtime.plugins).length,
  dispatchCount: runtime.dispatchTrace.count,
  lastDispatchId: runtime.dispatchTrace.lastDispatchId,
  lastScope: runtime.dispatchTrace.lastScope,
  lastActionType: runtime.dispatchTrace.lastActionType,
  plugins: Object.values(runtime.plugins).map((p) => ({
    id: p.instanceId,
    instanceId: p.instanceId,
    packageId: p.packageId,
    title: p.title,
    status: p.status,
    enabled: p.enabled,
    widgets: p.widgets.length,
  })),
}));

export function dispatchPluginAction(
  dispatch: AppDispatch,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
) {
  dispatch(pluginActionDispatched(instanceId, actionType, payload));
}

export function dispatchGlobalAction(dispatch: AppDispatch, actionType: string, payload?: unknown) {
  if (!ALLOWED_GLOBAL_ACTION_TYPES.has(actionType)) {
    throw new Error(`Global action not allowed: ${actionType}`);
  }

  dispatch(globalActionDispatched(actionType, payload));
}
