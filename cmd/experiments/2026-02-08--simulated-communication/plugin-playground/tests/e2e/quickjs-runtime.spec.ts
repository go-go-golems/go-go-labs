import { expect, test } from "@playwright/test";

test("greeter input updates local and shared greeter state", async ({ page }) => {
  await page.goto("/");

  await page.getByRole("button", { name: /Interactive Greeter/i }).click();
  await page.getByRole("button", { name: /Greeter Shared State/i }).click();

  const input = page.getByPlaceholder("Your name");
  await input.fill("Ada");

  await expect(page.getByText("Hello, Ada!", { exact: true })).toBeVisible();
  await expect(page.getByText("Shared greeting: Hello, Ada!")).toBeVisible();
});

test("runaway render code is interrupted and does not crash the app", async ({ page }) => {
  await page.goto("/");

  const infinitePluginCode = `
definePlugin(({ ui }) => {
  return {
    id: "infinite-loop",
    title: "Infinite Loop",
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
  `;

  await page.getByPlaceholder("definePlugin(({ ui }) => { ... })").fill(infinitePluginCode);
  await page.getByRole("button", { name: "LOAD PLUGIN" }).click();

  await expect(page.getByText(/Render error:/i)).toBeVisible();
  await expect(page.getByRole("heading", { name: "PLUGIN PLAYGROUND" })).toBeVisible();
});

test("counter can run as multiple independent instances", async ({ page }) => {
  await page.goto("/");

  const counterButton = page.getByRole("button", { name: /^Counter/ });
  await counterButton.click();
  await counterButton.click();

  await expect(page.getByText("Counter: 0")).toHaveCount(2);

  await page.getByRole("button", { name: "Increment" }).first().click();

  await expect(page.getByText("Counter: 1")).toHaveCount(1);
  await expect(page.getByText("Counter: 0")).toHaveCount(1);
});

test("custom plugins without write grants cannot mutate shared domains", async ({ page }) => {
  await page.goto("/");

  await page.getByRole("button", { name: /Status Dashboard/ }).click();

  const deniedWriterPlugin = `
definePlugin(({ ui }) => {
  return {
    id: "denied-writer",
    title: "Denied Writer",
    widgets: {
      denied: {
        render() {
          return ui.panel([
            ui.text("Denied writer plugin"),
            ui.button("Attempt Shared Write", { onClick: { handler: "attemptWrite" } }),
          ]);
        },
        handlers: {
          attemptWrite({ dispatchSharedAction }) {
            dispatchSharedAction("counter-summary", "set-instance", { value: 999 });
          },
        },
      },
    },
  };
});
  `;

  await page.getByPlaceholder("definePlugin(({ ui }) => { ... })").fill(deniedWriterPlugin);
  await page.getByRole("button", { name: "LOAD PLUGIN" }).click();
  await page.getByRole("button", { name: "Attempt Shared Write" }).click();

  await expect(page.getByText("Shared Counter: 0")).toBeVisible();
});
