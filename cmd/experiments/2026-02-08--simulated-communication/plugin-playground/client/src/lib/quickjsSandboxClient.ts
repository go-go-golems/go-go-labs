import type { UINode } from "./uiTypes";
import type {
  DispatchIntent,
  DisposePluginRequest,
  EventRequest,
  HealthRequest,
  HealthResult,
  InstanceId,
  LoadPluginRequest,
  LoadedPlugin,
  PackageId,
  RenderRequest,
  RuntimeErrorPayload,
  WorkerResponse,
} from "./quickjsContracts";

type PendingRequest = {
  resolve: (value: any) => void;
  reject: (reason?: unknown) => void;
};

type RequestWithoutId =
  | Omit<LoadPluginRequest, "id">
  | Omit<RenderRequest, "id">
  | Omit<EventRequest, "id">
  | Omit<DisposePluginRequest, "id">
  | Omit<HealthRequest, "id">;

function toError(error: RuntimeErrorPayload): Error {
  const prefixedCode = error.code ? `[${error.code}] ` : "";
  return new Error(`${prefixedCode}${error.message}`);
}

export class QuickJSSandboxClient {
  private worker: Worker;

  private nextId = 1;

  private pending = new Map<number, PendingRequest>();

  constructor() {
    this.worker = new Worker(new URL("../workers/quickjsRuntime.worker.ts", import.meta.url), {
      type: "module",
    });

    this.worker.onmessage = this.handleWorkerMessage;
    this.worker.onerror = (event) => {
      const error = new Error(`QuickJS worker error: ${event.message}`);
      this.pending.forEach((pending) => pending.reject(error));
      this.pending.clear();
    };
  }

  private handleWorkerMessage = (event: MessageEvent<WorkerResponse>) => {
    const response = event.data;
    const pending = this.pending.get(response.id);
    if (!pending) {
      return;
    }

    this.pending.delete(response.id);

    if (response.ok) {
      pending.resolve(response.result);
      return;
    }

    pending.reject(toError(response.error));
  };

  private postRequest<TResponse>(request: RequestWithoutId): Promise<TResponse> {
    const id = this.nextId++;
    const requestWithId = { id, ...request };

    return new Promise<TResponse>((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
      this.worker.postMessage(requestWithId);
    });
  }

  async loadPlugin(packageId: PackageId, instanceId: InstanceId, code: string): Promise<LoadedPlugin> {
    const result = await this.postRequest<{ plugin: LoadedPlugin }>(
      {
        type: "loadPlugin",
        packageId,
        instanceId,
        code,
      } satisfies Omit<LoadPluginRequest, "id">
    );

    return result.plugin;
  }

  async render(
    instanceId: InstanceId,
    widgetId: string,
    pluginState: unknown,
    globalState: unknown
  ): Promise<UINode> {
    const result = await this.postRequest<{ tree: UINode }>(
      {
        type: "render",
        instanceId,
        widgetId,
        pluginState,
        globalState,
      } satisfies Omit<RenderRequest, "id">
    );

    return result.tree;
  }

  async event(
    instanceId: InstanceId,
    widgetId: string,
    handler: string,
    args: unknown,
    pluginState: unknown,
    globalState: unknown
  ): Promise<DispatchIntent[]> {
    const result = await this.postRequest<{ intents: DispatchIntent[] }>(
      {
        type: "event",
        instanceId,
        widgetId,
        handler,
        args,
        pluginState,
        globalState,
      } satisfies Omit<EventRequest, "id">
    );

    return result.intents;
  }

  async disposePlugin(instanceId: InstanceId): Promise<boolean> {
    const result = await this.postRequest<{ disposed: boolean }>(
      {
        type: "disposePlugin",
        instanceId,
      } satisfies Omit<DisposePluginRequest, "id">
    );

    return result.disposed;
  }

  async health(): Promise<HealthResult> {
    return this.postRequest<HealthResult>({ type: "health" } satisfies Omit<HealthRequest, "id">);
  }

  terminate() {
    const error = new Error("QuickJS worker terminated");
    this.pending.forEach((pending) => pending.reject(error));
    this.pending.clear();
    this.worker.terminate();
  }
}

export const quickjsSandboxClient = new QuickJSSandboxClient();
