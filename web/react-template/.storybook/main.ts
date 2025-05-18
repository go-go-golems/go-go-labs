import { StorybookConfig } from '@storybook/react-vite';

const config: StorybookConfig = {
  stories: ['../src/**/*.stories.@(ts|tsx)'],
  addons: ['msw-storybook-addon'],
  framework: { name: '@storybook/react-vite', options: {} },
  staticDirs: ['../public'],
};
export default config; 