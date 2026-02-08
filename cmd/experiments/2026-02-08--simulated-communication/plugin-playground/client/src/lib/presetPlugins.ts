// Preset plugins for the plugin playground
import type { UINode, UIEventRef } from "./uiTypes";

export interface PluginDefinition {
  id: string;
  title: string;
  description: string;
  code: string; // The plugin code as a string
}

// Calculator Plugin
export const calculatorPlugin: PluginDefinition = {
  id: "calculator",
  title: "Simple Calculator",
  description: "A basic calculator with +, -, *, / operations",
  code: `
definePlugin(({ ui, createActions }) => {
  const actions = createActions("plugin.calculator", [
    "digit", "operation", "equals", "clear"
  ]);

  return {
    id: "calculator",
    title: "Calculator",
    widgets: {
      display: {
        render({ state }) {
          const display = state.calculator?.display || "0";
          return ui.panel([
            ui.text(\`Display: \${display}\`),
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
              ui.button("C", { onClick: { handler: "clear" } }),
              ui.button("+", { onClick: { handler: "operation", args: "+" } }),
            ]),
          ]);
        },
        handlers: {
          digit({ dispatch }, digit) {
            dispatch({ type: "plugin.calculator/digit", payload: digit });
          },
          operation({ dispatch }, op) {
            dispatch({ type: "plugin.calculator/operation", payload: op });
          },
          equals({ dispatch }) {
            dispatch({ type: "plugin.calculator/equals" });
          },
          clear({ dispatch }) {
            dispatch({ type: "plugin.calculator/clear" });
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
  description: "Shows system and plugin status information",
  code: `
definePlugin(({ ui, createActions }) => {
  return {
    id: "status-dashboard",
    title: "Status Dashboard",
    widgets: {
      status: {
        render({ state }) {
          const pluginCount = Object.keys(state.plugins?.plugins || {}).length;
          const counterValue = state.counter || 0;
          
          return ui.panel([
            ui.text("System Status"),
            ui.row([
              ui.badge(\`Plugins Loaded: \${pluginCount}\`),
              ui.badge(\`Counter Value: \${counterValue}\`),
            ]),
            ui.text("All systems operational"),
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
  description: "A simple greeter that responds to your name",
  code: `
definePlugin(({ ui, createActions }) => {
  const actions = createActions("plugin.greeter", ["nameChanged"]);

  return {
    id: "greeter",
    title: "Greeter",
    widgets: {
      greeter: {
        render({ state }) {
          const name = state.greeter?.name || "";
          const greeting = name ? \`Hello, \${name}! 👋\` : "Enter your name...";
          
          return ui.panel([
            ui.text(greeting),
            ui.input(name, {
              placeholder: "Your name",
              onChange: { handler: "updateName" },
            }),
          ]);
        },
        handlers: {
          updateName({ dispatch }, args) {
            const name = args?.value || "";
            dispatch({ type: "plugin.greeter/nameChanged", payload: name });
          },
        },
      },
    },
  };
});
  `,
};

// VM Monitor Plugin
export const vmMonitorPlugin: PluginDefinition = {
  id: "vm-monitor",
  title: "VM Monitor",
  description: "Monitors plugin VM status and metrics",
  code: `
definePlugin(({ ui, createActions }) => {
  return {
    id: "vm-monitor",
    title: "VM Monitor",
    widgets: {
      monitor: {
        render({ state }) {
          const plugins = state.plugins?.plugins || {};
          const pluginList = Object.values(plugins).map(p => ({
            name: p.id,
            status: p.status,
            widgets: p.meta?.widgets?.length || 0,
          }));
          
          return ui.panel([
            ui.text("Plugin Registry"),
            ui.text(\`Total: \${pluginList.length} plugins\`),
            ...pluginList.map(p => 
              ui.row([
                ui.badge(p.name),
                ui.badge(\`Status: \${p.status}\`),
                ui.badge(\`Widgets: \${p.widgets}\`),
              ])
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

export const presetPlugins = [
  calculatorPlugin,
  statusDashboardPlugin,
  greeterPlugin,
  vmMonitorPlugin,
];
