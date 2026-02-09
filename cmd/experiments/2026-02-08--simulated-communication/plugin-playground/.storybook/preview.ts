import type { Preview } from "@storybook/react-vite";
import "../client/src/index.css";
import { withStore } from "../client/src/features/workbench/__stories__/storyDecorators";

const preview: Preview = {
  parameters: {
    backgrounds: {
      default: "brutalist-dark",
      values: [
        { name: "brutalist-dark", value: "oklch(0.15 0.01 240)" },
        { name: "white", value: "#ffffff" },
      ],
    },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
  decorators: [
    // Global dark-class decorator
    (Story) => {
      document.documentElement.classList.add("dark");
      return Story();
    },
    // Global Redux store — every story gets a fresh default store.
    // Individual stories can override with withStore({ ... }) in their own decorators.
    withStore(),
  ],
};

export default preview;
