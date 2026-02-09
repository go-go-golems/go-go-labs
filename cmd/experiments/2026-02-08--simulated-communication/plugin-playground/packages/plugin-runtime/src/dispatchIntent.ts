import type { DispatchIntent } from "./contracts";

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function validateDispatchIntent(value: unknown, instanceId: string): DispatchIntent {
  if (!isRecord(value)) {
    throw new Error("Dispatch intent must be an object");
  }

  if (value.scope !== "plugin" && value.scope !== "shared") {
    throw new Error("Dispatch intent scope must be 'plugin' or 'shared'");
  }

  if (typeof value.actionType !== "string" || value.actionType.length === 0) {
    throw new Error("Dispatch intent actionType must be a non-empty string");
  }

  if (value.scope === "plugin") {
    return {
      scope: "plugin",
      instanceId,
      actionType: value.actionType,
      payload: value.payload,
    };
  }

  if (typeof value.domain !== "string" || value.domain.length === 0) {
    throw new Error("Shared dispatch intent domain must be a non-empty string");
  }

  return {
    scope: "shared",
    domain: value.domain,
    actionType: value.actionType,
    payload: value.payload,
  };
}

export function validateDispatchIntents(value: unknown, instanceId: string): DispatchIntent[] {
  if (!Array.isArray(value)) {
    throw new Error("Dispatch intents result must be an array");
  }

  return value.map((intent) => validateDispatchIntent(intent, instanceId));
}
