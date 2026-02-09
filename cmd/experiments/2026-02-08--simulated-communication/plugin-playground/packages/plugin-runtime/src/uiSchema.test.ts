import { describe, expect, it } from "vitest";
import { validateUINode } from "./uiSchema";

describe("validateUINode", () => {
  it("accepts a valid panel tree", () => {
    const node = validateUINode({
      kind: "panel",
      children: [
        { kind: "text", text: "hello" },
        {
          kind: "button",
          props: {
            label: "Click",
            onClick: { handler: "clicked", args: { id: 1 } },
          },
        },
      ],
    });

    expect(node.kind).toBe("panel");
  });

  it("rejects unsupported kinds", () => {
    expect(() => validateUINode({ kind: "unknown" })).toThrow(/not supported/i);
  });

  it("rejects input without string value", () => {
    expect(() =>
      validateUINode({
        kind: "input",
        props: { value: 42 },
      })
    ).toThrow(/value must be a string/i);
  });
});

