// Design Philosophy: Technical Brutalism - Plugins as pure functions
// Each preset demonstrates different widget types and Redux interactions

export interface PresetPlugin {
  id: string;
  name: string;
  description: string;
  code: string;
}

export const PRESET_PLUGINS: PresetPlugin[] = [
  {
    id: "counter",
    name: "Counter Plugin",
    description: "Simple counter with increment/decrement buttons. Demonstrates Redux state management.",
    code: `definePlugin(({ ui, createActions }) => {
  const actions = createActions("plugin.counter", [
    "incremented",
    "decremented",
    "reset"
  ]);

  return {
    id: "counter",
    title: "Counter Control",
    description: "A simple counter widget",
    
    widgets: {
      CounterWidget: {
        title: "Counter",
        render({ state }) {
          const count = (state && state.plugins && state.plugins.counter) || 0;
          return ui.panel([
            ui.text("Current Count: " + count),
            ui.row([
              ui.button("Decrement", { onClick: { handler: "decrement" } }),
              ui.button("Reset", { onClick: { handler: "reset" }, variant: "destructive" }),
              ui.button("Increment", { onClick: { handler: "increment" } })
            ])
          ]);
        },
        
        handlers: {
          increment({ dispatch }) {
            dispatch(actions.incremented());
          },
          decrement({ dispatch }) {
            dispatch(actions.decremented());
          },
          reset({ dispatch }) {
            dispatch(actions.reset());
          }
        }
      }
    }
  };
});`,
  },
  {
    id: "status",
    name: "Status Dashboard",
    description: "Shows plugin status with badges and tables. Demonstrates reading Redux state.",
    code: `definePlugin(({ ui, createActions }) => {
  return {
    id: "status",
    title: "Status Dashboard",
    description: "Shows system status",
    
    widgets: {
      StatusWidget: {
        title: "System Status",
        render({ state }) {
          const counter = (state && state.plugins && state.plugins.counter) || 0;
          
          return ui.panel([
            ui.text("System Overview"),
            ui.row([
              ui.badge("ONLINE"),
              ui.badge("VM: ACTIVE"),
              ui.badge("PLUGINS: ACTIVE")
            ]),
            ui.text(""),
            ui.text("State Snapshot:"),
            ui.table(
              [
                ["Counter Value", String(counter)],
                ["VM Status", "Running"]
              ],
              { headers: ["Metric", "Value"] }
            )
          ]);
        },
        
        handlers: {}
      }
    }
  };
});`,
  },
  {
    id: "greeter",
    name: "Interactive Greeter",
    description: "Input field with dynamic greeting. Demonstrates local widget state and input handling.",
    code: `definePlugin(({ ui, createActions }) => {
  const actions = createActions("plugin.greeter", ["nameChanged"]);

  return {
    id: "greeter",
    title: "Greeter Widget",
    description: "Interactive greeting generator",
    
    widgets: {
      GreeterWidget: {
        title: "Greeter",
        render({ state }) {
          const name = (state && state.plugins && state.plugins.greeter && state.plugins.greeter.name) || "";
          const greeting = name ? "Hello, " + name + "! Welcome to the Plugin Playground." : "Enter your name to get started.";
          
          return ui.panel([
            ui.text("Enter your name:"),
            ui.input(name, { 
              placeholder: "Your name...",
              onChange: { handler: "nameChanged" }
            }),
            ui.text(""),
            ui.text(greeting)
          ]);
        },
        
        handlers: {
          nameChanged({ dispatch, event }) {
            dispatch(actions.nameChanged(event.value));
          }
        }
      }
    }
  };
});`,
  },
  {
    id: "calculator",
    name: "Simple Calculator",
    description: "Basic calculator with multiple operations. Shows complex widget composition.",
    code: `definePlugin(({ ui, createActions }) => {
  const actions = createActions("plugin.calculator", [
    "digit",
    "clear",
    "operation",
    "equals"
  ]);

  return {
    id: "calculator",
    title: "Calculator",
    description: "Basic arithmetic calculator",
    
    widgets: {
      CalcWidget: {
        title: "Calculator",
        render({ state }) {
          const calc = (state && state.plugins && state.plugins.calculator) || { display: "0", accumulator: 0, operation: null };
          const display = calc.display || "0";
          
          return ui.panel([
            ui.text("CALC DISPLAY"),
            ui.text("=" + display),
            ui.text(""),
            ui.row([
              ui.button("7", { onClick: { handler: "digit", args: 7 } }),
              ui.button("8", { onClick: { handler: "digit", args: 8 } }),
              ui.button("9", { onClick: { handler: "digit", args: 9 } })
            ]),
            ui.row([
              ui.button("4", { onClick: { handler: "digit", args: 4 } }),
              ui.button("5", { onClick: { handler: "digit", args: 5 } }),
              ui.button("6", { onClick: { handler: "digit", args: 6 } })
            ]),
            ui.row([
              ui.button("1", { onClick: { handler: "digit", args: 1 } }),
              ui.button("2", { onClick: { handler: "digit", args: 2 } }),
              ui.button("3", { onClick: { handler: "digit", args: 3 } })
            ]),
            ui.row([
              ui.button("0", { onClick: { handler: "digit", args: 0 } }),
              ui.button("+", { onClick: { handler: "operation", args: "+" } }),
              ui.button("-", { onClick: { handler: "operation", args: "-" } })
            ]),
            ui.row([
              ui.button("*", { onClick: { handler: "operation", args: "*" } }),
              ui.button("/", { onClick: { handler: "operation", args: "/" } }),
              ui.button("=", { onClick: { handler: "equals" } })
            ]),
            ui.button("CLEAR", { onClick: { handler: "clear" }, variant: "destructive" })
          ]);
        },
        
        handlers: {
          digit({ dispatch, event }) {
            dispatch(actions.digit(event.args));
          },
          clear({ dispatch }) {
            dispatch(actions.clear());
          },
          operation({ dispatch, event }) {
            dispatch(actions.operation(event.args));
          },
          equals({ dispatch }) {
            dispatch(actions.equals());
          }
        }
      }
    }
  };
});`,
  },
  {
    id: "monitor",
    name: "VM Monitor",
    description: "Shows VM execution stats and plugin metadata. Advanced state inspection.",
    code: `definePlugin(({ ui, createActions }) => {
  return {
    id: "monitor",
    title: "VM Monitor",
    description: "QuickJS VM monitoring",
    
    widgets: {
      MonitorWidget: {
        title: "VM Monitor",
        render({ state }) {
          return ui.panel([
            ui.row([
              ui.badge("QUICKJS"),
              ui.badge("WASM"),
              ui.badge("ISOLATED")
            ]),
            ui.text(""),
            ui.text("Loaded Plugins:"),
            ui.table(
              [
                ["Counter", "ACTIVE", "YES", "1"],
                ["Greeter", "ACTIVE", "YES", "1"],
                ["Calculator", "ACTIVE", "YES", "1"]
              ],
              { headers: ["Plugin", "Status", "Enabled", "Widgets"] }
            )
          ]);
        },
        
        handlers: {}
      }
    }
  };
});`,
  },
];
