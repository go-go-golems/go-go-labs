// Preset plugins for the unified v1 runtime contract

export interface PluginDefinition {
  id: string;
  title: string;
  description: string;
  code: string;
}

// Counter Plugin
export const counterPlugin: PluginDefinition = {
  id: "counter",
  title: "Counter",
  description: "Simple local counter with mirrored global counter value",
  code: `
definePlugin(({ ui }) => {
  return {
    id: "counter",
    title: "Counter",
    description: "Simple counter",
    initialState: { value: 0 },
    widgets: {
      counter: {
        render({ pluginState }) {
          const value = Number(pluginState?.value ?? 0);
          return ui.panel([
            ui.text("Counter: " + value),
            ui.row([
              ui.button("Decrement", { onClick: { handler: "decrement" } }),
              ui.button("Reset", { onClick: { handler: "reset" }, variant: "destructive" }),
              ui.button("Increment", { onClick: { handler: "increment" } }),
            ]),
          ]);
        },
        handlers: {
          increment({ dispatchPluginAction, dispatchGlobalAction, pluginState }) {
            const next = Number(pluginState?.value ?? 0) + 1;
            dispatchPluginAction("increment");
            dispatchGlobalAction("counter/set", next);
          },
          decrement({ dispatchPluginAction, dispatchGlobalAction, pluginState }) {
            const next = Number(pluginState?.value ?? 0) - 1;
            dispatchPluginAction("decrement");
            dispatchGlobalAction("counter/set", next);
          },
          reset({ dispatchPluginAction, dispatchGlobalAction }) {
            dispatchPluginAction("reset");
            dispatchGlobalAction("counter/set", 0);
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
  description: "Shows unified runtime status and global metrics",
  code: `
definePlugin(({ ui }) => {
  return {
    id: "status-dashboard",
    title: "Status Dashboard",
    description: "Runtime status dashboard",
    widgets: {
      status: {
        render({ globalState }) {
          const pluginCount = Number(globalState?.pluginCount ?? 0);
          const counterValue = Number(globalState?.counterValue ?? 0);
          const dispatchCount = Number(globalState?.dispatchCount ?? 0);

          return ui.panel([
            ui.text("System Status"),
            ui.row([
              ui.badge("Plugins: " + pluginCount),
              ui.badge("Counter: " + counterValue),
              ui.badge("Dispatches: " + dispatchCount),
            ]),
            ui.table(
              [
                ["Plugin Count", String(pluginCount)],
                ["Counter Value", String(counterValue)],
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
          updateName({ dispatchPluginAction }, args) {
            dispatchPluginAction("nameChanged", args?.value ?? "");
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
  description: "Shows loaded plugin registry from global runtime state",
  code: `
definePlugin(({ ui }) => {
  return {
    id: "runtime-monitor",
    title: "Runtime Monitor",
    description: "Runtime monitor",
    widgets: {
      monitor: {
        render({ globalState }) {
          const plugins = Array.isArray(globalState?.plugins) ? globalState.plugins : [];

          return ui.panel([
            ui.text("Plugin Registry"),
            ui.text("Total: " + plugins.length + " plugins"),
            ui.table(
              plugins.map((p) => [
                String(p.id),
                String(p.status),
                p.enabled ? "YES" : "NO",
                String(p.widgets),
              ]),
              { headers: ["Plugin", "Status", "Enabled", "Widgets"] }
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
  description: "Shows greeter state mirrored into shared global runtime state",
  code: `
definePlugin(({ ui }) => {
  return {
    id: "greeter-shared-state",
    title: "Greeter Shared State",
    description: "Shared greeter state viewer",
    widgets: {
      sharedGreeter: {
        render({ globalState }) {
          const name = String(globalState?.greeterName ?? "");
          const greeting = name ? "Shared greeting: Hello, " + name + "!" : "Shared greeting: (empty)";

          return ui.panel([
            ui.text("Reads from globalState.greeterName"),
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
