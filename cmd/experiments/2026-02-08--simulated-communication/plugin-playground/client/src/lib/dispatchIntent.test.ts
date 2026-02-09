import { describe, expect, it } from "vitest";
import { validateDispatchIntent, validateDispatchIntents } from "./dispatchIntent";

describe("validateDispatchIntent", () => {
  it("normalizes plugin scoped intents with instanceId", () => {
    const intent = validateDispatchIntent(
      {
        scope: "plugin",
        actionType: "increment",
        payload: { by: 1 },
      },
      "counter"
    );

    expect(intent).toEqual({
      scope: "plugin",
      instanceId: "counter",
      actionType: "increment",
      payload: { by: 1 },
    });
  });

  it("accepts global scoped intents", () => {
    const intent = validateDispatchIntent(
      {
        scope: "global",
        actionType: "counter/set",
        payload: 4,
      },
      "counter"
    );

    expect(intent.scope).toBe("global");
    expect(intent.actionType).toBe("counter/set");
  });

  it("rejects malformed intent arrays", () => {
    expect(() => validateDispatchIntents({} as never, "x")).toThrow(/must be an array/i);
  });
});
