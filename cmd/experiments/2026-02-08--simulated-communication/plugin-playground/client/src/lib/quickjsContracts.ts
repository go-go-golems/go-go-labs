import type { UINode } from "./uiTypes";

export interface RuntimeErrorPayload {
  code: string;
  message: string;
  details?: unknown;
}

export interface DispatchIntent {
  scope: "plugin" | "global";
  actionType: string;
  payload?: unknown;
  pluginId?: string;
}

export interface LoadedPlugin {
  id: string;
  declaredId?: string;
  title: string;
  description?: string;
  initialState?: unknown;
  widgets: string[];
}

export interface LoadPluginRequest {
  id: number;
  type: "loadPlugin";
  pluginId: string;
  code: string;
}

export interface RenderRequest {
  id: number;
  type: "render";
  pluginId: string;
  widgetId: string;
  pluginState: unknown;
  globalState: unknown;
}

export interface EventRequest {
  id: number;
  type: "event";
  pluginId: string;
  widgetId: string;
  handler: string;
  args?: unknown;
  pluginState: unknown;
  globalState: unknown;
}

export interface DisposePluginRequest {
  id: number;
  type: "disposePlugin";
  pluginId: string;
}

export interface HealthRequest {
  id: number;
  type: "health";
}

export type WorkerRequest =
  | LoadPluginRequest
  | RenderRequest
  | EventRequest
  | DisposePluginRequest
  | HealthRequest;

export interface LoadPluginResult {
  plugin: LoadedPlugin;
}

export interface RenderResult {
  tree: UINode;
}

export interface EventResult {
  intents: DispatchIntent[];
}

export interface DisposePluginResult {
  disposed: boolean;
}

export interface HealthResult {
  ready: true;
  plugins: string[];
}

export type WorkerResult = LoadPluginResult | RenderResult | EventResult | DisposePluginResult | HealthResult;

export interface WorkerSuccessResponse {
  id: number;
  ok: true;
  result: WorkerResult;
}

export interface WorkerErrorResponse {
  id: number;
  ok: false;
  error: RuntimeErrorPayload;
}

export type WorkerResponse = WorkerSuccessResponse | WorkerErrorResponse;

