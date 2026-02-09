import type { SharedDomainName } from "@runtime/redux-adapter/store";

// Preset plugins for the unified v1 runtime contract

export interface PluginDefinition {
  id: string;
  title: string;
  description: string;
  capabilities?: {
    readShared?: SharedDomainName[];
    writeShared?: SharedDomainName[];
    systemCommands?: string[];
  };
  code: string;
}

// Counter Plugin
export const counterPlugin: PluginDefinition = {
  id: "counter",
  title: "Counter",
  description: "Simple local counter with shared counter summary updates",
  capabilities: {
    readShared: ["counter-summary"],
    writeShared: ["counter-summary"],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "counter",
    title: "Counter",
    description: "Simple counter",
    initialState: { value: 0 },
    widgets: {
      counter: {
        render({ pluginState, globalState }) {
          const value = Number(pluginState?.value ?? 0);
          const sharedCounter = globalState?.shared?.["counter-summary"];
          const totalValue = Number(sharedCounter?.totalValue ?? 0);
          const instanceCount = Number(sharedCounter?.instanceCount ?? 0);

          return ui.panel([
            ui.text("Counter: " + value),
            ui.row([
              ui.badge("Shared total: " + totalValue),
              ui.badge("Instances: " + instanceCount),
            ]),
            ui.row([
              ui.button("Decrement", { onClick: { handler: "decrement" } }),
              ui.button("Reset", { onClick: { handler: "reset" }, variant: "destructive" }),
              ui.button("Increment", { onClick: { handler: "increment" } }),
            ]),
          ]);
        },
        handlers: {
          increment({ dispatchPluginAction, dispatchSharedAction, pluginState }) {
            const next = Number(pluginState?.value ?? 0) + 1;
            dispatchPluginAction("increment");
            dispatchSharedAction("counter-summary", "set-instance", { value: next });
          },
          decrement({ dispatchPluginAction, dispatchSharedAction, pluginState }) {
            const next = Number(pluginState?.value ?? 0) - 1;
            dispatchPluginAction("decrement");
            dispatchSharedAction("counter-summary", "set-instance", { value: next });
          },
          reset({ dispatchPluginAction, dispatchSharedAction }) {
            dispatchPluginAction("reset");
            dispatchSharedAction("counter-summary", "set-instance", { value: 0 });
          },
        },
      },
    },
  };
});
  `,
};

// Calculator Plugin
export const calculatorPlugin: PluginDefinition = {
  id: "calculator",
  title: "Simple Calculator",
  description: "A basic calculator with +, -, *, / operations",
  capabilities: {
    readShared: [],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "calculator",
    title: "Calculator",
    description: "Basic arithmetic calculator",
    initialState: {
      display: "0",
      accumulator: 0,
      operation: null,
    },
    widgets: {
      display: {
        render({ pluginState }) {
          const display = String(pluginState?.display ?? "0");
          return ui.panel([
            ui.text("Display: " + display),
            ui.row([
              ui.button("7", { onClick: { handler: "digit", args: 7 } }),
              ui.button("8", { onClick: { handler: "digit", args: 8 } }),
              ui.button("9", { onClick: { handler: "digit", args: 9 } }),
              ui.button("/", { onClick: { handler: "operation", args: "/" } }),
            ]),
            ui.row([
              ui.button("4", { onClick: { handler: "digit", args: 4 } }),
              ui.button("5", { onClick: { handler: "digit", args: 5 } }),
              ui.button("6", { onClick: { handler: "digit", args: 6 } }),
              ui.button("*", { onClick: { handler: "operation", args: "*" } }),
            ]),
            ui.row([
              ui.button("1", { onClick: { handler: "digit", args: 1 } }),
              ui.button("2", { onClick: { handler: "digit", args: 2 } }),
              ui.button("3", { onClick: { handler: "digit", args: 3 } }),
              ui.button("-", { onClick: { handler: "operation", args: "-" } }),
            ]),
            ui.row([
              ui.button("0", { onClick: { handler: "digit", args: 0 } }),
              ui.button("=", { onClick: { handler: "equals" } }),
              ui.button("C", { onClick: { handler: "clear" }, variant: "destructive" }),
              ui.button("+", { onClick: { handler: "operation", args: "+" } }),
            ]),
          ]);
        },
        handlers: {
          digit({ dispatchPluginAction }, digit) {
            dispatchPluginAction("digit", digit);
          },
          operation({ dispatchPluginAction }, op) {
            dispatchPluginAction("operation", op);
          },
          equals({ dispatchPluginAction }) {
            dispatchPluginAction("equals");
          },
          clear({ dispatchPluginAction }) {
            dispatchPluginAction("clear");
          },
        },
      },
    },
  };
});
  `,
};

// Status Dashboard Plugin
export const statusDashboardPlugin: PluginDefinition = {
  id: "status-dashboard",
  title: "Status Dashboard",
  description: "Shows unified runtime status and shared domain metrics",
  capabilities: {
    readShared: ["counter-summary", "runtime-metrics", "runtime-registry"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "status-dashboard",
    title: "Status Dashboard",
    description: "Runtime status dashboard",
    widgets: {
      status: {
        render({ globalState }) {
          const counterSummary = globalState?.shared?.["counter-summary"] ?? {};
          const runtimeMetrics = globalState?.shared?.["runtime-metrics"] ?? {};
          const pluginCount = Number(runtimeMetrics?.pluginCount ?? 0);
          const dispatchCount = Number(runtimeMetrics?.dispatchCount ?? 0);
          const counterValue = Number(counterSummary?.totalValue ?? 0);

          return ui.panel([
            ui.text("System Status"),
            ui.row([
              ui.badge("Plugins: " + pluginCount),
              ui.badge("Shared Counter: " + counterValue),
              ui.badge("Dispatches: " + dispatchCount),
            ]),
            ui.table(
              [
                ["Plugin Count", String(pluginCount)],
                ["Shared Counter Total", String(counterValue)],
                ["Dispatch Count", String(dispatchCount)],
              ],
              { headers: ["Metric", "Value"] }
            ),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

// Greeter Plugin
export const greeterPlugin: PluginDefinition = {
  id: "greeter",
  title: "Interactive Greeter",
  description: "Simple local state demo with input handling",
  capabilities: {
    readShared: ["greeter-profile"],
    writeShared: ["greeter-profile"],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "greeter",
    title: "Greeter",
    description: "Simple greeter",
    initialState: { name: "" },
    widgets: {
      greeter: {
        render({ pluginState }) {
          const name = String(pluginState?.name ?? "");
          const greeting = name ? "Hello, " + name + "!" : "Enter your name...";

          return ui.panel([
            ui.text(greeting),
            ui.input(name, {
              placeholder: "Your name",
              onChange: { handler: "updateName" },
            }),
          ]);
        },
        handlers: {
          updateName({ dispatchPluginAction, dispatchSharedAction }, args) {
            const name = args?.value ?? "";
            dispatchPluginAction("nameChanged", name);
            dispatchSharedAction("greeter-profile", "set-name", name);
          },
        },
      },
    },
  };
});
  `,
};

// Runtime Monitor Plugin
export const runtimeMonitorPlugin: PluginDefinition = {
  id: "runtime-monitor",
  title: "Runtime Monitor",
  description: "Shows loaded plugin registry from shared runtime state",
  capabilities: {
    readShared: ["runtime-registry"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "runtime-monitor",
    title: "Runtime Monitor",
    description: "Runtime monitor",
    widgets: {
      monitor: {
        render({ globalState }) {
          const plugins = Array.isArray(globalState?.shared?.["runtime-registry"])
            ? globalState.shared["runtime-registry"]
            : [];

          return ui.panel([
            ui.text("Plugin Registry"),
            ui.text("Total: " + plugins.length + " plugins"),
            ui.table(
              plugins.map((p) => [
                String(p.instanceId ?? p.id ?? ""),
                String(p.packageId ?? ""),
                String(p.status),
                p.enabled ? "YES" : "NO",
                String(p.widgets),
              ]),
              { headers: ["Instance", "Package", "Status", "Enabled", "Widgets"] }
            ),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

// Shared Greeter State Viewer Plugin
export const sharedGreeterStatePlugin: PluginDefinition = {
  id: "greeter-shared-state",
  title: "Greeter Shared State",
  description: "Shows greeter state from shared domain",
  capabilities: {
    readShared: ["greeter-profile"],
    writeShared: [],
  },
  code: `
definePlugin(({ ui }) => {
  return {
    id: "greeter-shared-state",
    title: "Greeter Shared State",
    description: "Shared greeter state viewer",
    widgets: {
      sharedGreeter: {
        render({ globalState }) {
          const greeterShared = globalState?.shared?.["greeter-profile"] ?? {};
          const name = String(greeterShared?.name ?? "");
          const greeting = name ? "Shared greeting: Hello, " + name + "!" : "Shared greeting: (empty)";

          return ui.panel([
            ui.text("Reads from globalState.shared['greeter-profile']"),
            ui.badge(name ? "SYNCED" : "NO NAME"),
            ui.text(greeting),
          ]);
        },
        handlers: {},
      },
    },
  };
});
  `,
};

export const presetPlugins = [
  counterPlugin,
  calculatorPlugin,
  statusDashboardPlugin,
  greeterPlugin,
  sharedGreeterStatePlugin,
  runtimeMonitorPlugin,
];
