import type { Preview } from "@storybook/react-vite";
import "../client/src/index.css";

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
    (Story) => {
      // Ensure the dark class is on the root for Tailwind dark mode
      document.documentElement.classList.add("dark");
      return Story();
    },
  ],
};

export default preview;
