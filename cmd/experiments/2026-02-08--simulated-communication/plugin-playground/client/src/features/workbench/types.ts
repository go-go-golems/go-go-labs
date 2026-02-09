import type { LoadedPlugin } from "@runtime/contracts";
import type { UINode } from "@runtime/uiTypes";

export type LoadedPluginMap = Record<string, LoadedPlugin>;
export type WidgetTrees = Record<string, Record<string, UINode>>;
export type WidgetErrors = Record<string, Record<string, string>>;
