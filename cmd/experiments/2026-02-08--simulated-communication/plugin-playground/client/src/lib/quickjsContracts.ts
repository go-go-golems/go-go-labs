import type { UINode } from "./uiTypes";

export type PackageId = string;
export type InstanceId = string;

export interface RuntimeErrorPayload {
  code: string;
  message: string;
  details?: unknown;
}

export interface DispatchIntent {
  scope: "plugin" | "shared";
  actionType: string;
  payload?: unknown;
  instanceId?: InstanceId;
  domain?: string;
}

export interface LoadedPlugin {
  packageId: PackageId;
  instanceId: InstanceId;
  declaredId?: string;
  title: string;
  description?: string;
  initialState?: unknown;
  widgets: string[];
}

export interface LoadPluginRequest {
  id: number;
  type: "loadPlugin";
  packageId: PackageId;
  instanceId: InstanceId;
  code: string;
}

export interface RenderRequest {
  id: number;
  type: "render";
  instanceId: InstanceId;
  widgetId: string;
  pluginState: unknown;
  globalState: unknown;
}

export interface EventRequest {
  id: number;
  type: "event";
  instanceId: InstanceId;
  widgetId: string;
  handler: string;
  args?: unknown;
  pluginState: unknown;
  globalState: unknown;
}

export interface DisposePluginRequest {
  id: number;
  type: "disposePlugin";
  instanceId: InstanceId;
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
