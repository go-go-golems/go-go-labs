// In-process plugin sandbox - simpler and easier to debug than web workers
import type { Store } from "@reduxjs/toolkit";
import type { UINode, UIEventRef } from "./uiTypes";

export interface PluginMeta {
  id: string;
  title?: string;
  description?: string;
  widgets: string[];
}

interface UIBuilder {
  text: (content: string) => UINode;
  button: (label: string, props?: any) => UINode;
  input: (value: string, props?: any) => UINode;
  row: (children: UINode[]) => UINode;
  panel: (children: UINode[]) => UINode;
  badge: (text: string) => UINode;
  table: (rows: any[][], props?: any) => UINode;
}

interface ActionCreators {
  [key: string]: (...args: any[]) => any;
}

interface PluginContext {
  ui: UIBuilder;
  createActions: (namespace: string, actionNames: string[]) => ActionCreators;
}

interface PluginDefinition {
  id: string;
  title?: string;
  description?: string;
  widgets: Record<string, WidgetDefinition>;
}

interface WidgetDefinition {
  title: string;
  render: (context: { state: any }) => UINode;
  handlers: Record<string, (context: any, args?: any) => void>;
}

export class PluginSandboxClient {
  private store: Store;
  private plugins: Map<string, PluginDefinition> = new Map();

  constructor({
    store,
    workerUrl,
    allowDispatch,
  }: {
    store: Store;
    workerUrl?: URL;
    allowDispatch?: (pluginId: string, action: any) => boolean;
  }) {
    this.store = store;
  }

  async loadPlugin(pluginId: string, code: string): Promise<PluginMeta> {
    try {
      // Create the UI builder
      const ui: UIBuilder = {
        text: (content: string) => ({
          kind: "text",
          text: content,
        } as UINode),
        button: (label: string, props?: any) => ({
          kind: "button",
          props: {
            label,
            ...props,
          },
        } as UINode),
        input: (value: string, props?: any) => ({
          kind: "input",
          props: {
            value,
            ...props,
          },
        } as UINode),
        row: (children: UINode[]) => ({
          kind: "row",
          children,
        } as UINode),
        panel: (children: UINode[]) => ({
          kind: "panel",
          children,
        } as UINode),
        badge: (text: string) => ({
          kind: "badge",
          text,
        } as UINode),
        table: (rows: any[][], props?: any) => ({
          kind: "table",
          props: {
            headers: props?.headers || [],
            rows,
          },
        } as UINode),
      };

      // Create action creators
      const createActions = (namespace: string, actionNames: string[]) => {
        const actions: ActionCreators = {};
        for (const name of actionNames) {
          actions[name] = (payload?: any) => ({
            type: `${namespace}/${name}`,
            payload,
          });
        }
        return actions;
      };

      // Create the plugin context
      const context: PluginContext = {
        ui,
        createActions,
      };

      // Define the global definePlugin function
      let pluginDef: PluginDefinition | null = null;
      (window as any).definePlugin = (fn: (context: PluginContext) => PluginDefinition) => {
        pluginDef = fn(context);
      };

      // Execute the plugin code
      const fn = new Function(code);
      fn();

      if (!pluginDef) {
        throw new Error("Plugin did not call definePlugin");
      }

      // Store the plugin
      this.plugins.set(pluginId, pluginDef as PluginDefinition);

      // Return metadata
      const def = pluginDef as PluginDefinition;
      return {
        id: def.id,
        title: def.title,
        description: def.description,
        widgets: Object.keys(def.widgets),
      };
    } catch (error) {
      throw new Error(`Failed to load plugin: ${String(error)}`);
    }
  }

  async render(pluginId: string, widgetId: string, state: any): Promise<UINode> {
    try {
      const plugin = this.plugins.get(pluginId);
      if (!plugin) {
        throw new Error(`Plugin not found: ${pluginId}`);
      }

      const widget = plugin.widgets[widgetId];
      if (!widget) {
        throw new Error(`Widget not found: ${widgetId}`);
      }

      // Render the widget
      const tree = widget.render({ state });
      return tree;
    } catch (error) {
      throw new Error(`Failed to render widget: ${String(error)}`);
    }
  }

  async event(
    pluginId: string,
    widgetId: string,
    handlerName: string,
    eventPayload?: any,
    state?: any
  ): Promise<void> {
    try {
      const plugin = this.plugins.get(pluginId);
      if (!plugin) {
        throw new Error(`Plugin not found: ${pluginId}`);
      }

      const widget = plugin.widgets[widgetId];
      if (!widget) {
        throw new Error(`Widget not found: ${widgetId}`);
      }

      const handler = widget.handlers[handlerName];
      if (!handler) {
        throw new Error(`Handler not found: ${handlerName}`);
      }

      // Create dispatch function that dispatches to Redux
      const dispatch = (action: any) => {
        this.store.dispatch(action);
      };

      // Call the handler
      handler(
        {
          dispatch,
          event: eventPayload,
          state,
        },
        eventPayload?.args
      );
    } catch (error) {
      console.error(`Failed to handle event: ${String(error)}`);
      throw error;
    }
  }

  terminate(): void {
    // No-op for in-process sandbox
  }
}
