interface RpcRequest {
  id: number;
  type: "loadPlugin" | "render" | "event";
  pluginId: string;
  code?: string;
  widgetId?: string;
  handler?: string;
  state?: any;
  event?: any;
}

interface RpcResponse {
  id: number;
  ok: boolean;
  result?: any;
  error?: any;
}

interface DispatchMessage {
  type: "dispatch";
  pluginId: string;
  actionJson: string;
}

interface PluginContext {
  plugin: any;
}

const plugins = new Map<string, PluginContext>();

function createSandboxContext(pluginId: string) {
  const dispatchedActions: any[] = [];

  return {
    ui: {
      panel: (children: any) => ({ type: "panel", children }),
      row: (children: any) => ({ type: "row", children }),
      text: (text: string) => ({ type: "text", text }),
      button: (label: string, props: any) => ({ type: "button", props: { label, ...props } }),
      input: (props: any) => ({ type: "input", props }),
      badge: (text: string) => ({ type: "badge", text }),
      table: (rows: any, props: any) => ({ type: "table", rows, props })
    },
    createActions: (namespace: string, names: string[]) => {
      const actions: any = {};
      names.forEach(name => {
        actions[name] = (payload: any) => ({
          type: `${namespace}/${name}`,
          payload
        });
      });
      return actions;
    },
    __hostDispatch: (action: any) => {
      const msg: DispatchMessage = {
        type: "dispatch",
        pluginId,
        actionJson: JSON.stringify(action)
      };
      postMessage(msg);
    }
  };
}

async function handleLoadPlugin(msg: RpcRequest) {
  const { pluginId, code } = msg;

  try {
    const context = createSandboxContext(pluginId);
    
    // Create a function that runs the plugin code in a sandbox
    const pluginFn = new Function(
      "definePlugin",
      "ui",
      "createActions",
      `
      let __plugin = null;
      function definePlugin(fn) {
        __plugin = fn({
          ui: ui,
          createActions: createActions
        });
      }
      ${code}
      return __plugin;
      `
    );

    const plugin = pluginFn(
      (fn: any) => {
        const result = fn(context);
        plugins.set(pluginId, { plugin: result });
        return result;
      },
      context.ui,
      context.createActions
    );

    plugins.set(pluginId, { plugin });

    return {
      id: plugin?.id || "unknown",
      title: plugin?.title || "Untitled",
      description: plugin?.description || "",
      widgets: plugin?.widgets ? Object.keys(plugin.widgets) : []
    };
  } catch (err: any) {
    throw new Error(`Failed to load plugin: ${err.message}`);
  }
}

async function handleRender(msg: RpcRequest) {
  const { pluginId, widgetId, state } = msg;
  const ctx = plugins.get(pluginId);

  if (!ctx) {
    throw new Error(`Plugin not found: ${pluginId}`);
  }

  const plugin = ctx.plugin;
  const widget = plugin.widgets?.[widgetId || ""];

  if (!widget) {
    throw new Error(`Widget not found: ${widgetId}`);
  }

  try {
    const tree = widget.render({ state });
    return tree;
  } catch (err: any) {
    throw new Error(`Render failed: ${err.message}`);
  }
}

async function handleEvent(msg: RpcRequest) {
  const { pluginId, widgetId, handler, state, event } = msg;
  const ctx = plugins.get(pluginId);

  if (!ctx) {
    throw new Error(`Plugin not found: ${pluginId}`);
  }

  const plugin = ctx.plugin;
  const widget = plugin.widgets?.[widgetId || ""];

  if (!widget) {
    throw new Error(`Widget not found: ${widgetId}`);
  }

  const fn = widget.handlers?.[handler || ""];

  if (!fn) {
    throw new Error(`Handler not found: ${handler}`);
  }

  try {
    const dispatch = (action: any) => {
      const msg: DispatchMessage = {
        type: "dispatch",
        pluginId,
        actionJson: JSON.stringify(action)
      };
      postMessage(msg);
    };

    fn({ dispatch, state, event });
    return null;
  } catch (err: any) {
    throw new Error(`Event handler failed: ${err.message}`);
  }
}

self.onmessage = async (e: MessageEvent<RpcRequest>) => {
  const msg = e.data;
  let resp: RpcResponse;

  try {
    let result;

    if (msg.type === "loadPlugin") {
      result = await handleLoadPlugin(msg);
    } else if (msg.type === "render") {
      result = await handleRender(msg);
    } else if (msg.type === "event") {
      result = await handleEvent(msg);
    } else {
      throw new Error(`Unknown message type: ${(msg as any).type}`);
    }

    resp = { id: msg.id, ok: true, result };
  } catch (error: any) {
    resp = { id: msg.id, ok: false, error: error?.message || String(error) };
  }

  self.postMessage(resp);
};
