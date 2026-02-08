import type { UINode } from "./uiTypes";

export interface PluginInstance {
  id: string;
  declaredId?: string;
  title: string;
  description?: string;
  initialState?: unknown;
  widgets: Record<string, WidgetInstance>;
}

interface WidgetInstance {
  render: (context: { pluginState: unknown; globalState: unknown }) => UINode;
  handlers: Record<
    string,
    (
      context: {
        pluginId: string;
        pluginState: unknown;
        globalState: unknown;
        dispatchPluginAction: (actionType: string, payload?: unknown) => void;
        dispatchGlobalAction: (actionType: string, payload?: unknown) => void;
      },
      args?: unknown
    ) => void
  >;
}

interface PluginDefinition {
  id?: string;
  title: string;
  description?: string;
  initialState?: unknown;
  widgets: Record<string, WidgetInstance>;
}

interface PluginBuildContext {
  ui: any;
}

export class PluginManager {
  private plugins: Map<string, PluginInstance> = new Map();

  async loadPlugin(pluginId: string, code: string, context: PluginBuildContext): Promise<PluginInstance> {
    return new Promise((resolve, reject) => {
      try {
        let pluginDef: PluginDefinition | null = null;

        const definePlugin = (fn: (ctx: PluginBuildContext) => PluginDefinition) => {
          pluginDef = fn(context);
        };

        const fn = new Function("definePlugin", code);
        fn(definePlugin);

        if (!pluginDef) {
          throw new Error("Plugin did not call definePlugin");
        }

        const def = pluginDef as PluginDefinition;

        const resolvedPlugin: PluginInstance = {
          id: pluginId,
          declaredId: def.id,
          title: def.title,
          description: def.description,
          initialState: def.initialState,
          widgets: def.widgets,
        };

        this.plugins.set(pluginId, resolvedPlugin);
        resolve(resolvedPlugin);
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

  renderWidget(pluginId: string, widgetId: string, pluginState: unknown, globalState: unknown): UINode {
    const plugin = this.plugins.get(pluginId);
    if (!plugin) {
      throw new Error(`Plugin not found: ${pluginId}`);
    }

    const widget = plugin.widgets[widgetId];
    if (!widget) {
      throw new Error(`Widget not found: ${widgetId}`);
    }

    return widget.render({ pluginState, globalState });
  }

  callHandler(
    pluginId: string,
    widgetId: string,
    handlerName: string,
    dispatchPluginAction: (actionType: string, payload?: unknown) => void,
    dispatchGlobalAction: (actionType: string, payload?: unknown) => void,
    args?: unknown,
    pluginState?: unknown,
    globalState?: unknown
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

    handler(
      {
        pluginId,
        pluginState,
        globalState,
        dispatchPluginAction,
        dispatchGlobalAction,
      },
      args
    );
  }

  removePlugin(id: string): void {
    this.plugins.delete(id);
  }

  clear(): void {
    this.plugins.clear();
  }
}

export const pluginManager = new PluginManager();
