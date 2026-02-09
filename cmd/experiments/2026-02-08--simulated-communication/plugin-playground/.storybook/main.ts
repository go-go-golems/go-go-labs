import type { StorybookConfig } from "@storybook/react-vite";
import path from "path";
import { fileURLToPath } from "url";
import { mergeConfig } from "vite";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const config: StorybookConfig = {
  stories: ["../client/src/**/*.stories.@(ts|tsx|mdx)"],
  addons: ["@storybook/addon-a11y", "@storybook/addon-docs"],
  framework: "@storybook/react-vite",
  viteFinal(config) {
    return mergeConfig(config, {
      resolve: {
        alias: {
          "@": path.resolve(__dirname, "..", "client", "src"),
          "@runtime": path.resolve(
            __dirname,
            "..",
            "packages",
            "plugin-runtime",
            "src"
          ),
          "@shared": path.resolve(__dirname, "..", "shared"),
          "@docs": path.resolve(__dirname, "..", "docs"),
        },
      },
    });
  },
};
export default config;
