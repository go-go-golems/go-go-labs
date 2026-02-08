// Minimal Counter Plugin - QuickJS + Redux Integration Demo
// This is a hardcoded plugin to demonstrate the full flow working end-to-end

import type { UINode, UIEventRef } from "./uiTypes";

export interface MinimalPluginContext {
  state: {
    counter: {
      value: number;
    };
  };
  dispatch: (action: any) => void;
}

/**
 * Minimal counter plugin that:
 * 1. Reads counter value from Redux state
 * 2. Renders a UI with buttons
 * 3. Dispatches Redux actions on button clicks
 */
export const minimalCounterPlugin = {
  id: "minimal-counter",
  title: "Minimal Counter",
  description: "A simple counter plugin demonstrating QuickJS + Redux integration",

  render(context: MinimalPluginContext): UINode {
    const count = context.state.counter.value;

    return {
      kind: "panel",
      children: [
        {
          kind: "text",
          text: `Counter: ${count}`,
          props: {
            className: "text-lg font-bold text-cyan-400",
          },
        },
        {
          kind: "row",
          children: [
            {
              kind: "button",
              props: {
                label: "Decrement",
                onClick: { handler: "decrement" } as UIEventRef,
              },
            },
            {
              kind: "button",
              props: {
                label: "Increment",
                onClick: { handler: "increment" } as UIEventRef,
              },
            },
            {
              kind: "button",
              props: {
                label: "Reset",
                onClick: { handler: "reset" } as UIEventRef,
              },
            },
          ],
        },
      ],
    };
  },

  handlers: {
    increment(context: MinimalPluginContext) {
      context.dispatch({
        type: "plugin.counter/incremented",
        payload: 1,
      });
    },

    decrement(context: MinimalPluginContext) {
      context.dispatch({
        type: "plugin.counter/decremented",
        payload: 1,
      });
    },

    reset(context: MinimalPluginContext) {
      context.dispatch({
        type: "plugin.counter/reset",
        payload: 0,
      });
    },
  },
};
