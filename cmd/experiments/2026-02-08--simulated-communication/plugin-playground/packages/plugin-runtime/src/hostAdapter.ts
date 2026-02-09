import type {
  DispatchIntent,
  HealthResult,
  InstanceId,
  LoadedPlugin,
  PackageId,
} from "./contracts";
import type { UINode } from "./uiTypes";

export interface RuntimeLoadInput {
  packageId: PackageId;
  instanceId: InstanceId;
  code: string;
}

export interface RuntimeRenderInput {
  instanceId: InstanceId;
  widgetId: string;
  pluginState: unknown;
  globalState: unknown;
}

export interface RuntimeEventInput extends RuntimeRenderInput {
  handler: string;
  args: unknown;
}

export interface RuntimeHostAdapter {
  loadPlugin(input: RuntimeLoadInput): Promise<LoadedPlugin>;
  render(input: RuntimeRenderInput): Promise<UINode>;
  event(input: RuntimeEventInput): Promise<DispatchIntent[]>;
  disposePlugin(instanceId: InstanceId): Promise<boolean>;
  health(): Promise<HealthResult>;
  terminate?(): void;
}
