import { afterEach, describe, expect, it } from "vitest";
import { QuickJSRuntimeService } from "./quickjsRuntimeService";

const COUNTER_PLUGIN = `
definePlugin(({ ui }) => {
  return {
    id: "counter",
    title: "Counter",
    initialState: { value: 0 },
    widgets: {
      counter: {
        render({ pluginState }) {
          return ui.panel([
            ui.text("Counter: " + String(pluginState?.value ?? 0)),
            ui.button("Increment", { onClick: { handler: "increment" } }),
          ]);
        },
        handlers: {
          increment({ dispatchPluginAction }) {
            dispatchPluginAction("increment");
          },
        },
      },
    },
  };
});
`;

describe("QuickJSRuntimeService", () => {
  const services: QuickJSRuntimeService[] = [];

  afterEach(() => {
    for (const service of services) {
      for (const pluginId of service.health().plugins) {
        service.disposePlugin(pluginId);
      }
    }
    services.length = 0;
  });

  it("loads plugin and renders a valid tree", async () => {
    const service = new QuickJSRuntimeService();
    services.push(service);

    const plugin = await service.loadPlugin("counter", COUNTER_PLUGIN);
    expect(plugin.widgets).toEqual(["counter"]);

    const tree = service.render("counter", "counter", { value: 2 }, {});
    expect(tree.kind).toBe("panel");
  });

  it("returns dispatch intents from event handler calls", async () => {
    const service = new QuickJSRuntimeService();
    services.push(service);

    await service.loadPlugin("counter", COUNTER_PLUGIN);
    const intents = service.event("counter", "counter", "increment", undefined, { value: 0 }, {});

    expect(intents).toEqual([
      {
        scope: "plugin",
        pluginId: "counter",
        actionType: "increment",
        payload: undefined,
      },
    ]);
  });

  it("disposes plugin runtimes and rejects further renders", async () => {
    const service = new QuickJSRuntimeService();
    services.push(service);

    await service.loadPlugin("counter", COUNTER_PLUGIN);
    expect(service.disposePlugin("counter")).toBe(true);
    expect(service.disposePlugin("counter")).toBe(false);
    expect(() => service.render("counter", "counter", {}, {})).toThrow(/not found/i);
  });

  it("interrupts infinite render loops with timeout", async () => {
    const service = new QuickJSRuntimeService({ renderTimeoutMs: 10 });
    services.push(service);

    await service.loadPlugin(
      "loop",
      `
definePlugin(({ ui }) => {
  return {
    id: "loop",
    title: "Loop",
    widgets: {
      loop: {
        render() {
          while (true) {}
        },
        handlers: {},
      },
    },
  };
});
      `
    );

    expect(() => service.render("loop", "loop", {}, {})).toThrow(/interrupted/i);
  });
});

