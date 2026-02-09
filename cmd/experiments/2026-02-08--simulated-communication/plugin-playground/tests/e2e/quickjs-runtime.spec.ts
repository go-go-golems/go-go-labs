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
