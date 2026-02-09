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

  it("accepts shared scoped intents", () => {
    const intent = validateDispatchIntent(
      {
        scope: "shared",
        domain: "counter-summary",
        actionType: "set-instance",
        payload: 4,
      },
      "counter"
    );

    expect(intent.scope).toBe("shared");
    expect(intent.domain).toBe("counter-summary");
    expect(intent.actionType).toBe("set-instance");
  });

  it("rejects shared intents without domain", () => {
    expect(() =>
      validateDispatchIntent(
        {
          scope: "shared",
          actionType: "set-instance",
        },
        "counter"
      )
    ).toThrow(/domain/i);
  });

  it("rejects malformed intent arrays", () => {
    expect(() => validateDispatchIntents({} as never, "x")).toThrow(/must be an array/i);
  });
});
