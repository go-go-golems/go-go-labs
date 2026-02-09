import { configureStore, createSelector, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { nanoid } from "nanoid";
import type { InstanceId, PackageId } from "../contracts";

export type PluginStatus = "loaded" | "error";
export type SharedDomainName =
  | "counter-summary"
  | "greeter-profile"
  | "runtime-registry"
  | "runtime-metrics";

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

export interface CapabilityGrants {
  readShared: SharedDomainName[];
  writeShared: SharedDomainName[];
  systemCommands: string[];
}

interface CounterSummaryDomainState {
  valuesByInstance: Record<InstanceId, number>;
  totalValue: number;
  instanceCount: number;
  lastUpdatedInstanceId: InstanceId | null;
}

interface GreeterProfileDomainState {
  name: string;
  lastUpdatedInstanceId: InstanceId | null;
}

interface RuntimeState {
  plugins: Record<InstanceId, RuntimePlugin>;
  pluginStateById: Record<InstanceId, unknown>;
  grantsByInstance: Record<InstanceId, CapabilityGrants>;
  shared: {
    "counter-summary": CounterSummaryDomainState;
    "greeter-profile": GreeterProfileDomainState;
  };
  dispatchTrace: {
    count: number;
    lastDispatchId: string | null;
    lastScope: "plugin" | "shared" | null;
    lastActionType: string | null;
    lastOutcome: "applied" | "denied" | "ignored" | null;
    lastReason: string | null;
  };
}

interface ScopedDispatchPayload {
  dispatchId: string;
  timestamp: number;
  scope: "plugin" | "shared";
  actionType: string;
  instanceId?: InstanceId;
  domain?: SharedDomainName;
  payload?: unknown;
}

const DEFAULT_GRANTS: CapabilityGrants = {
  readShared: [],
  writeShared: [],
  systemCommands: [],
};

const initialState: RuntimeState = {
  plugins: {},
  pluginStateById: {},
  grantsByInstance: {},
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
    lastDispatchId: null,
    lastScope: null,
    lastActionType: null,
    lastOutcome: null,
    lastReason: null,
  },
};

function markDispatch(state: RuntimeState, dispatch: ScopedDispatchPayload) {
  state.dispatchTrace.count += 1;
  state.dispatchTrace.lastDispatchId = dispatch.dispatchId;
  state.dispatchTrace.lastScope = dispatch.scope;
  state.dispatchTrace.lastActionType = dispatch.actionType;
  state.dispatchTrace.lastOutcome = null;
  state.dispatchTrace.lastReason = null;
}

function markDispatchOutcome(
  state: RuntimeState,
  outcome: "applied" | "denied" | "ignored",
  reason: string | null = null
) {
  state.dispatchTrace.lastOutcome = outcome;
  state.dispatchTrace.lastReason = reason;
}

function hasWriteGrant(state: RuntimeState, instanceId: InstanceId, domain: SharedDomainName): boolean {
  const grants = state.grantsByInstance[instanceId] ?? DEFAULT_GRANTS;
  return grants.writeShared.includes(domain);
}

function hasReadGrant(state: RuntimeState, instanceId: InstanceId, domain: SharedDomainName): boolean {
  const grants = state.grantsByInstance[instanceId] ?? DEFAULT_GRANTS;
  return grants.readShared.includes(domain);
}

function reduceCounterPlugin(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
): boolean {
  const current = (state.pluginStateById[instanceId] as { value?: number } | undefined) ?? {};
  let value = Number(current.value ?? 0);

  if (actionType === "increment") {
    value += 1;
  } else if (actionType === "decrement") {
    value -= 1;
  } else if (actionType === "reset") {
    value = 0;
  } else {
    return false;
  }

  state.pluginStateById[instanceId] = { value };
  return true;
}

function reduceCalculatorPlugin(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
): boolean {
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
    return false;
  }

  state.pluginStateById[instanceId] = calculator;
  return true;
}

function reduceGreeterPlugin(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
): boolean {
  if (actionType !== "nameChanged") {
    return false;
  }

  const name = String(payload ?? "");
  state.pluginStateById[instanceId] = {
    name,
  };
  return true;
}

function reduceGenericPlugin(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
): boolean {
  if (actionType === "state/replace") {
    state.pluginStateById[instanceId] = payload ?? {};
    return true;
  }

  if (
    actionType === "state/merge" &&
    typeof payload === "object" &&
    payload !== null &&
    !Array.isArray(payload)
  ) {
    const current =
      typeof state.pluginStateById[instanceId] === "object" &&
      state.pluginStateById[instanceId] !== null &&
      !Array.isArray(state.pluginStateById[instanceId])
        ? (state.pluginStateById[instanceId] as Record<string, unknown>)
        : {};

    state.pluginStateById[instanceId] = {
      ...current,
      ...(payload as Record<string, unknown>),
    };
    return true;
  }

  return false;
}

function reducePluginScopedAction(
  state: RuntimeState,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
): boolean {
  const plugin = state.plugins[instanceId];
  if (!plugin) {
    return false;
  }

  if (plugin.packageId === "counter") {
    return reduceCounterPlugin(state, instanceId, actionType, payload);
  }

  if (plugin.packageId === "calculator") {
    return reduceCalculatorPlugin(state, instanceId, actionType, payload);
  }

  if (plugin.packageId === "greeter") {
    return reduceGreeterPlugin(state, instanceId, actionType, payload);
  }

  return reduceGenericPlugin(state, instanceId, actionType, payload);
}

function applyCounterSummarySetInstance(state: RuntimeState, instanceId: InstanceId, payload?: unknown): boolean {
  if (typeof payload !== "object" || payload === null || Array.isArray(payload)) {
    return false;
  }

  const value = Number((payload as { value?: unknown }).value ?? 0);
  const domain = state.shared["counter-summary"];
  domain.valuesByInstance[instanceId] = value;
  domain.instanceCount = Object.keys(domain.valuesByInstance).length;
  domain.totalValue = Object.values(domain.valuesByInstance).reduce((sum, next) => sum + next, 0);
  domain.lastUpdatedInstanceId = instanceId;
  return true;
}

function applyGreeterProfileSetName(state: RuntimeState, instanceId: InstanceId, payload?: unknown): boolean {
  const name = String(payload ?? "");
  state.shared["greeter-profile"].name = name;
  state.shared["greeter-profile"].lastUpdatedInstanceId = instanceId;
  return true;
}

function reduceSharedScopedAction(
  state: RuntimeState,
  instanceId: InstanceId,
  domain: SharedDomainName,
  actionType: string,
  payload?: unknown
): { outcome: "applied" | "ignored" | "denied"; reason: string | null } {
  if (!hasWriteGrant(state, instanceId, domain)) {
    return { outcome: "denied", reason: `missing-write-grant:${domain}` };
  }

  if (domain === "counter-summary") {
    const applied = actionType === "set-instance" && applyCounterSummarySetInstance(state, instanceId, payload);
    return applied
      ? { outcome: "applied", reason: null }
      : { outcome: "ignored", reason: `unsupported-action:${domain}/${actionType}` };
  }

  if (domain === "greeter-profile") {
    const applied = actionType === "set-name" && applyGreeterProfileSetName(state, instanceId, payload);
    return applied
      ? { outcome: "applied", reason: null }
      : { outcome: "ignored", reason: `unsupported-action:${domain}/${actionType}` };
  }

  return { outcome: "ignored", reason: `unsupported-domain:${domain}` };
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
        grants?: CapabilityGrants;
      }>
    ) {
      const { instanceId, packageId, title, description, widgets, initialState, grants } = action.payload;

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

      state.grantsByInstance[instanceId] = grants ?? DEFAULT_GRANTS;
    },

    pluginRemoved(state, action: PayloadAction<InstanceId>) {
      const instanceId = action.payload;
      delete state.plugins[instanceId];
      delete state.pluginStateById[instanceId];
      delete state.grantsByInstance[instanceId];

      delete state.shared["counter-summary"].valuesByInstance[instanceId];
      state.shared["counter-summary"].instanceCount = Object.keys(
        state.shared["counter-summary"].valuesByInstance
      ).length;
      state.shared["counter-summary"].totalValue = Object.values(
        state.shared["counter-summary"].valuesByInstance
      ).reduce((sum, next) => sum + next, 0);
      if (state.shared["counter-summary"].lastUpdatedInstanceId === instanceId) {
        state.shared["counter-summary"].lastUpdatedInstanceId = null;
      }

      if (state.shared["greeter-profile"].lastUpdatedInstanceId === instanceId) {
        state.shared["greeter-profile"].lastUpdatedInstanceId = null;
      }
    },

    pluginActionDispatched: {
      reducer(state, action: PayloadAction<ScopedDispatchPayload>) {
        const { instanceId, actionType, payload } = action.payload;
        if (!instanceId) {
          return;
        }

        markDispatch(state, action.payload);
        const applied = reducePluginScopedAction(state, instanceId, actionType, payload);
        markDispatchOutcome(state, applied ? "applied" : "ignored", applied ? null : "no-local-reducer-match");
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

    sharedActionDispatched: {
      reducer(state, action: PayloadAction<ScopedDispatchPayload>) {
        const { instanceId, domain, actionType, payload } = action.payload;
        if (!instanceId || !domain) {
          return;
        }

        markDispatch(state, action.payload);
        const result = reduceSharedScopedAction(state, instanceId, domain, actionType, payload);
        markDispatchOutcome(state, result.outcome, result.reason);
      },
      prepare(
        instanceId: InstanceId,
        domain: SharedDomainName,
        actionType: string,
        payload?: unknown
      ) {
        return {
          payload: {
            dispatchId: nanoid(),
            timestamp: Date.now(),
            scope: "shared" as const,
            instanceId,
            domain,
            actionType,
            payload,
          },
        };
      },
    },
  },
});

export const { pluginRegistered, pluginRemoved, pluginActionDispatched, sharedActionDispatched } =
  runtimeSlice.actions;

export const store = configureStore({
  reducer: {
    runtime: runtimeSlice.reducer,
  },
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

const selectRuntimeState = (state: RootState) => state.runtime;

function buildRuntimeRegistry(runtime: RuntimeState) {
  return Object.values(runtime.plugins).map((p) => ({
    id: p.instanceId,
    instanceId: p.instanceId,
    packageId: p.packageId,
    title: p.title,
    status: p.status,
    enabled: p.enabled,
    widgets: p.widgets.length,
  }));
}

function buildRuntimeMetrics(runtime: RuntimeState) {
  return {
    pluginCount: Object.keys(runtime.plugins).length,
    dispatchCount: runtime.dispatchTrace.count,
    lastDispatchId: runtime.dispatchTrace.lastDispatchId,
    lastScope: runtime.dispatchTrace.lastScope,
    lastActionType: runtime.dispatchTrace.lastActionType,
    lastOutcome: runtime.dispatchTrace.lastOutcome,
    lastReason: runtime.dispatchTrace.lastReason,
  };
}

function buildSharedForInstance(runtime: RuntimeState, instanceId: InstanceId) {
  const shared: Record<string, unknown> = {};

  if (hasReadGrant(runtime, instanceId, "counter-summary")) {
    const domain = runtime.shared["counter-summary"];
    shared["counter-summary"] = {
      totalValue: domain.totalValue,
      instanceCount: domain.instanceCount,
      lastUpdatedInstanceId: domain.lastUpdatedInstanceId,
    };
  }

  if (hasReadGrant(runtime, instanceId, "greeter-profile")) {
    shared["greeter-profile"] = runtime.shared["greeter-profile"];
  }

  if (hasReadGrant(runtime, instanceId, "runtime-registry")) {
    shared["runtime-registry"] = buildRuntimeRegistry(runtime);
  }

  if (hasReadGrant(runtime, instanceId, "runtime-metrics")) {
    shared["runtime-metrics"] = buildRuntimeMetrics(runtime);
  }

  return shared;
}

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
  self: null,
  shared: {
    "counter-summary": {
      totalValue: runtime.shared["counter-summary"].totalValue,
      instanceCount: runtime.shared["counter-summary"].instanceCount,
      lastUpdatedInstanceId: runtime.shared["counter-summary"].lastUpdatedInstanceId,
    },
    "greeter-profile": runtime.shared["greeter-profile"],
    "runtime-registry": buildRuntimeRegistry(runtime),
    "runtime-metrics": buildRuntimeMetrics(runtime),
  },
  system: {
    ...buildRuntimeMetrics(runtime),
    plugins: buildRuntimeRegistry(runtime),
  },
}));

export function selectGlobalStateForInstance(state: RootState, instanceId: InstanceId) {
  const runtime = state.runtime;

  return {
    self: {
      instanceId,
      packageId: runtime.plugins[instanceId]?.packageId ?? "unknown",
    },
    shared: buildSharedForInstance(runtime, instanceId),
    system: {
      ...buildRuntimeMetrics(runtime),
      plugins: buildRuntimeRegistry(runtime),
    },
  };
}

export function dispatchPluginAction(
  dispatch: AppDispatch,
  instanceId: InstanceId,
  actionType: string,
  payload?: unknown
) {
  dispatch(pluginActionDispatched(instanceId, actionType, payload));
}

export function dispatchSharedAction(
  dispatch: AppDispatch,
  instanceId: InstanceId,
  domain: SharedDomainName,
  actionType: string,
  payload?: unknown
) {
  dispatch(sharedActionDispatched(instanceId, domain, actionType, payload));
}
