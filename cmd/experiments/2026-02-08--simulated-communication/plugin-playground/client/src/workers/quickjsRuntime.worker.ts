/// <reference lib="webworker" />

import type { WorkerRequest, WorkerResponse } from "../lib/quickjsContracts";
import { QuickJSRuntimeService, toRuntimeError } from "../lib/quickjsRuntimeService";

const runtimeService = new QuickJSRuntimeService();

async function handleRequest(request: WorkerRequest) {
  switch (request.type) {
    case "loadPlugin":
      return { plugin: await runtimeService.loadPlugin(request.pluginId, request.code) };
    case "render":
      return {
        tree: runtimeService.render(
          request.pluginId,
          request.widgetId,
          request.pluginState,
          request.globalState
        ),
      };
    case "event":
      return {
        intents: runtimeService.event(
          request.pluginId,
          request.widgetId,
          request.handler,
          request.args,
          request.pluginState,
          request.globalState
        ),
      };
    case "disposePlugin":
      return { disposed: runtimeService.disposePlugin(request.pluginId) };
    case "health":
      return runtimeService.health();
    default:
      throw new Error(`Unknown request type: ${(request as { type?: string }).type ?? "unknown"}`);
  }
}

const workerScope = self as unknown as DedicatedWorkerGlobalScope;

workerScope.onmessage = async (event: MessageEvent<WorkerRequest>) => {
  const request = event.data;

  try {
    const result = await handleRequest(request);
    const response: WorkerResponse = {
      id: request.id,
      ok: true,
      result,
    };
    workerScope.postMessage(response);
  } catch (error) {
    const response: WorkerResponse = {
      id: request.id,
      ok: false,
      error: toRuntimeError(error),
    };
    workerScope.postMessage(response);
  }
};

export {};
