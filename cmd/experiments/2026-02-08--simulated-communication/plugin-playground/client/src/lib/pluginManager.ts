// Plugin Manager - handles loading, rendering, and event handling for plugins
import type { UINode, UIEventRef } from "./uiTypes";

export interface PluginInstance {
  id: string;
  title: string;
  widgets: Record<string, WidgetInstance>;
}

interface WidgetInstance {
  render: (context: any) => UINode;
  handlers: Record<string, (context: any, args?: any) => void>;
}

export class PluginManager {
  private plugins: Map<string, PluginInstance> = new Map();

  async loadPlugin(
    code: string,
    context: {
      ui: any;
      createActions: any;
    }
  ): Promise<PluginInstance> {
    return new Promise((resolve, reject) => {
      try {
        let plugin: PluginInstance | null = null;

        // Create the definePlugin function
        const definePlugin = (fn: (ctx: any) => PluginInstance) => {
          plugin = fn(context);
        };

        // Execute the plugin code
        const fn = new Function("definePlugin", code);
        fn(definePlugin);

        if (!plugin) {
          throw new Error("Plugin did not call definePlugin");
        }

        this.plugins.set((plugin as PluginInstance).id, plugin as PluginInstance);
        resolve(plugin);
      } catch (error) {
        reject(new Error(`Failed to load plugin: ${String(error)}`));
      }
    });
  }

  getPlugin(id: string): PluginInstance | undefined {
    return this.plugins.get(id);
  }

  getAllPlugins(): PluginInstance[] {
    return Array.from(this.plugins.values());
  }

  renderWidget(
    pluginId: string,
    widgetId: string,
    state: any
  ): UINode {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new Error(`Plugin not found: ${pluginId}`);
    }

    const widget = plugin.widgets[widgetId];
    if (!widget) {
      throw new Error(`Widget not found: ${widgetId}`);
    }

    return widget.render({ state });
  }

  callHandler(
    pluginId: string,
    widgetId: string,
    handlerName: string,
    dispatch: any,
    args?: any,
    state?: any
  ): void {
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

    handler({ dispatch, state }, args);
  }

  removePlugin(id: string): void {
    this.plugins.delete(id);
  }

  clear(): void {
    this.plugins.clear();
  }
}

export const pluginManager = new PluginManager();
