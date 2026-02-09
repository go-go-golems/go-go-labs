import { getQuickJS } from "quickjs-emscripten";
import type { QuickJSContext, QuickJSRuntime } from "quickjs-emscripten";
import { validateDispatchIntents } from "./dispatchIntent";
import type {
  DispatchIntent,
  InstanceId,
  LoadedPlugin,
  PackageId,
  RuntimeErrorPayload,
} from "./quickjsContracts";
import { validateUINode } from "./uiSchema";

const BOOTSTRAP_SOURCE = `
const __ui = {
  text(content) {
    return { kind: "text", text: String(content) };
  },
  button(label, props = {}) {
    return { kind: "button", props: { label: String(label), ...props } };
  },
  input(value, props = {}) {
    return { kind: "input", props: { value: String(value ?? ""), ...props } };
  },
  row(children = []) {
    return { kind: "row", children: Array.isArray(children) ? children : [] };
  },
  panel(children = []) {
    return { kind: "panel", children: Array.isArray(children) ? children : [] };
  },
  badge(text) {
    return { kind: "badge", text: String(text) };
  },
  table(rows = [], props = {}) {
    return {
      kind: "table",
      props: {
        headers: Array.isArray(props?.headers) ? props.headers : [],
        rows: Array.isArray(rows) ? rows : [],
      },
    };
  },
};

let __plugin = null;
let __dispatchIntents = [];

function definePlugin(factory) {
  if (typeof factory !== "function") {
    throw new Error("definePlugin requires a factory function");
  }
  __plugin = factory({ ui: __ui });
}

globalThis.__pluginHost = {
  getMeta() {
    if (!__plugin || typeof __plugin !== "object") {
      throw new Error("Plugin did not register via definePlugin");
    }
    if (!__plugin.widgets || typeof __plugin.widgets !== "object") {
      throw new Error("Plugin widgets must be an object");
    }

    return {
      declaredId: typeof __plugin.id === "string" ? __plugin.id : undefined,
      title: String(__plugin.title ?? "Untitled Plugin"),
      description: typeof __plugin.description === "string" ? __plugin.description : undefined,
      initialState: __plugin.initialState,
      widgets: Object.keys(__plugin.widgets),
    };
  },

  render(widgetId, pluginState, globalState) {
    const widget = __plugin?.widgets?.[widgetId];
    if (!widget || typeof widget.render !== "function") {
      throw new Error("Widget not found or render() is missing: " + String(widgetId));
    }

    return widget.render({ pluginState, globalState });
  },

  event(widgetId, handlerName, args, pluginState, globalState) {
    const widget = __plugin?.widgets?.[widgetId];
    if (!widget) {
      throw new Error("Widget not found: " + String(widgetId));
    }

    const handler = widget.handlers?.[handlerName];
    if (typeof handler !== "function") {
      throw new Error("Handler not found: " + String(handlerName));
    }

    __dispatchIntents = [];

    const dispatchPluginAction = (actionType, payload) => {
      __dispatchIntents.push({
        scope: "plugin",
        actionType: String(actionType),
        payload,
      });
    };

    const dispatchGlobalAction = (actionType, payload) => {
      __dispatchIntents.push({
        scope: "global",
        actionType: String(actionType),
        payload,
      });
    };

    handler(
      {
        pluginState,
        globalState,
        dispatchPluginAction,
        dispatchGlobalAction,
      },
      args
    );

    return __dispatchIntents.slice();
  },
};
`;

interface PluginVm {
  packageId: PackageId;
  instanceId: InstanceId;
  runtime: QuickJSRuntime;
  context: QuickJSContext;
  deadlineMs: number;
}

export interface QuickJSRuntimeServiceOptions {
  memoryLimitBytes?: number;
  stackLimitBytes?: number;
  loadTimeoutMs?: number;
  renderTimeoutMs?: number;
  eventTimeoutMs?: number;
}

const DEFAULT_OPTIONS: Required<QuickJSRuntimeServiceOptions> = {
  memoryLimitBytes: 32 * 1024 * 1024,
  stackLimitBytes: 1024 * 1024,
  loadTimeoutMs: 1000,
  renderTimeoutMs: 100,
  eventTimeoutMs: 100,
};

function toJsLiteral(value: unknown): string {
  const encoded = JSON.stringify(value);
  return encoded === undefined ? "undefined" : encoded;
}

function formatQuickJSError(errorDump: unknown): string {
  if (typeof errorDump === "string") {
    return errorDump;
  }
  if (errorDump && typeof errorDump === "object") {
    const details = errorDump as { name?: string; message?: string };
    if (details.name && details.message) {
      return `${details.name}: ${details.message}`;
    }
    if (details.message) {
      return details.message;
    }
  }
  return "Unknown QuickJS runtime error";
}

function withDeadline<T>(vm: PluginVm, timeoutMs: number, fn: () => T): T {
  vm.deadlineMs = Date.now() + timeoutMs;
  try {
    return fn();
  } finally {
    vm.deadlineMs = Number.POSITIVE_INFINITY;
  }
}

function evalToNative<T>(vm: PluginVm, code: string, filename: string, timeoutMs: number): T {
  const context = vm.context;
  const result = withDeadline(vm, timeoutMs, () => context.evalCode(code, filename));
  if (result.error) {
    const dumped = context.dump(result.error);
    result.error.dispose();
    throw new Error(formatQuickJSError(dumped));
  }

  try {
    return context.dump(result.value) as T;
  } finally {
    result.value.dispose();
  }
}

function evalCodeOrThrow(vm: PluginVm, code: string, filename: string, timeoutMs: number): void {
  const context = vm.context;
  const result = withDeadline(vm, timeoutMs, () => context.evalCode(code, filename));
  if (result.error) {
    const dumped = context.dump(result.error);
    result.error.dispose();
    throw new Error(formatQuickJSError(dumped));
  }
  result.value.dispose();
}

function validateLoadedPluginMeta(
  packageId: PackageId,
  instanceId: InstanceId,
  value: unknown
): LoadedPlugin {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new Error("Plugin metadata must be an object");
  }

  const meta = value as {
    declaredId?: unknown;
    title?: unknown;
    description?: unknown;
    initialState?: unknown;
    widgets?: unknown;
  };

  if (!Array.isArray(meta.widgets) || meta.widgets.some((widgetId) => typeof widgetId !== "string")) {
    throw new Error("Plugin metadata widgets must be string[]");
  }

  return {
    packageId,
    instanceId,
    declaredId: typeof meta.declaredId === "string" ? meta.declaredId : undefined,
    title: typeof meta.title === "string" ? meta.title : "Untitled Plugin",
    description: typeof meta.description === "string" ? meta.description : undefined,
    initialState: meta.initialState,
    widgets: meta.widgets,
  };
}

export function toRuntimeError(error: unknown): RuntimeErrorPayload {
  if (error instanceof Error) {
    const interrupted = error.message.includes("interrupted");
    return {
      code: interrupted ? "RUNTIME_TIMEOUT" : "RUNTIME_ERROR",
      message: error.message,
    };
  }

  return {
    code: "UNKNOWN_ERROR",
    message: String(error),
  };
}

export class QuickJSRuntimeService {
  private readonly options: Required<QuickJSRuntimeServiceOptions>;

  private readonly vms = new Map<InstanceId, PluginVm>();

  constructor(options: QuickJSRuntimeServiceOptions = {}) {
    this.options = {
      ...DEFAULT_OPTIONS,
      ...options,
    };
  }

  private async createPluginVm(packageId: PackageId, instanceId: InstanceId): Promise<PluginVm> {
    const QuickJS = await getQuickJS();
    const runtime = QuickJS.newRuntime();
    const context = runtime.newContext();

    const vm: PluginVm = {
      packageId,
      instanceId,
      runtime,
      context,
      deadlineMs: Number.POSITIVE_INFINITY,
    };

    runtime.setMemoryLimit(this.options.memoryLimitBytes);
    runtime.setMaxStackSize(this.options.stackLimitBytes);
    runtime.setInterruptHandler(() => Date.now() > vm.deadlineMs);

    evalCodeOrThrow(vm, BOOTSTRAP_SOURCE, "plugin-bootstrap.js", this.options.loadTimeoutMs);
    return vm;
  }

  private getVmOrThrow(instanceId: InstanceId): PluginVm {
    const vm = this.vms.get(instanceId);
    if (!vm) {
      throw new Error(`Plugin runtime not found: ${instanceId}`);
    }
    return vm;
  }

  async loadPlugin(packageId: PackageId, instanceId: InstanceId, code: string): Promise<LoadedPlugin> {
    if (this.vms.has(instanceId)) {
      throw new Error(`Plugin runtime already exists: ${instanceId}`);
    }

    const vm = await this.createPluginVm(packageId, instanceId);

    try {
      evalCodeOrThrow(vm, code, `${instanceId}.plugin.js`, this.options.loadTimeoutMs);
      const meta = evalToNative<unknown>(
        vm,
        "globalThis.__pluginHost.getMeta()",
        "plugin-meta.js",
        this.options.loadTimeoutMs
      );
      const plugin = validateLoadedPluginMeta(packageId, instanceId, meta);
      this.vms.set(instanceId, vm);
      return plugin;
    } catch (error) {
      vm.context.dispose();
      vm.runtime.dispose();
      throw error;
    }
  }

  render(instanceId: InstanceId, widgetId: string, pluginState: unknown, globalState: unknown) {
    const vm = this.getVmOrThrow(instanceId);
    const tree = evalToNative<unknown>(
      vm,
      `globalThis.__pluginHost.render(${toJsLiteral(widgetId)}, ${toJsLiteral(
        pluginState
      )}, ${toJsLiteral(globalState)})`,
      `${instanceId}.render.js`,
      this.options.renderTimeoutMs
    );

    return validateUINode(tree);
  }

  event(
    instanceId: InstanceId,
    widgetId: string,
    handler: string,
    args: unknown,
    pluginState: unknown,
    globalState: unknown
  ): DispatchIntent[] {
    const vm = this.getVmOrThrow(instanceId);
    const intents = evalToNative<unknown>(
      vm,
      `globalThis.__pluginHost.event(${toJsLiteral(widgetId)}, ${toJsLiteral(
        handler
      )}, ${toJsLiteral(args)}, ${toJsLiteral(pluginState)}, ${toJsLiteral(globalState)})`,
      `${instanceId}.event.js`,
      this.options.eventTimeoutMs
    );

    return validateDispatchIntents(intents, instanceId);
  }

  disposePlugin(instanceId: InstanceId): boolean {
    const vm = this.vms.get(instanceId);
    if (!vm) {
      return false;
    }

    this.vms.delete(instanceId);
    vm.context.dispose();
    vm.runtime.dispose();
    return true;
  }

  health() {
    return {
      ready: true as const,
      plugins: Array.from(this.vms.keys()),
    };
  }
}
