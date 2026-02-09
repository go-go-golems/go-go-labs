import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    environment: "node",
    include: ["client/src/**/*.integration.test.ts", "packages/plugin-runtime/src/**/*.integration.test.ts"],
  },
});
