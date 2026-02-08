import React from "react";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "@/store/store";
import { WidgetRenderer } from "./WidgetRenderer";
import { minimalCounterPlugin } from "@/lib/minimalPlugin";
import type { UINode } from "@/lib/uiTypes";

/**
 * Minimal plugin widget that demonstrates:
 * 1. Reading Redux state
 * 2. Rendering UI from plugin
 * 3. Handling events and dispatching Redux actions
 */
export function MinimalPluginWidget() {
  const state = useSelector((s: RootState) => s);
  const dispatch = useDispatch();
  const [tree, setTree] = React.useState<UINode | null>(null);
  const [error, setError] = React.useState<string | null>(null);

  // Re-render when state changes
  React.useEffect(() => {
    try {
      const rendered = minimalCounterPlugin.render({
        state: {
          counter: {
            value: state.counter,
          },
        },
        dispatch,
      });
      setTree(rendered);
      setError(null);
    } catch (err) {
      setError(`Render error: ${String(err)}`);
      console.error("Render error:", err);
    }
  }, [state.counter, dispatch]);

  const handleEvent = (eventRef: any) => {
    try {
      const handler = (minimalCounterPlugin.handlers as any)[eventRef.handler];
      if (!handler) {
        throw new Error(`Handler not found: ${eventRef.handler}`);
      }
      handler(
        {
          state: {
            counter: {
              value: state.counter,
            },
          },
          dispatch,
        },
        eventRef.args
      );
    } catch (err) {
      setError(`Event error: ${String(err)}`);
      console.error("Event error:", err);
    }
  };

  if (error) {
    return (
      <div className="p-4 bg-red-900/20 border border-red-500 rounded text-red-400">
        {error}
      </div>
    );
  }

  if (!tree) {
    return <div className="text-cyan-400">Loading...</div>;
  }

  return <WidgetRenderer tree={tree} onEvent={handleEvent} />;
}
