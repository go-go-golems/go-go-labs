import type { DispatchIntent } from "./quickjsContracts";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function validateDispatchIntent(value: unknown, pluginId: string): DispatchIntent {
  if (!isRecord(value)) {
    throw new Error("Dispatch intent must be an object");
  }

  if (value.scope !== "plugin" && value.scope !== "global") {
    throw new Error("Dispatch intent scope must be 'plugin' or 'global'");
  }

  if (typeof value.actionType !== "string" || value.actionType.length === 0) {
    throw new Error("Dispatch intent actionType must be a non-empty string");
  }

  if (value.scope === "plugin") {
    return {
      scope: "plugin",
      pluginId,
      actionType: value.actionType,
      payload: value.payload,
    };
  }

  return {
    scope: "global",
    actionType: value.actionType,
    payload: value.payload,
  };
}

export function validateDispatchIntents(value: unknown, pluginId: string): DispatchIntent[] {
  if (!Array.isArray(value)) {
    throw new Error("Dispatch intents result must be an array");
  }

  return value.map((intent) => validateDispatchIntent(intent, pluginId));
}

