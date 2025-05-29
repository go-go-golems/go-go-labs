import { Agent } from './api';

export type RootStackParamList = {
  Main: undefined;
  AgentDetail: { agent: Agent };
};

export type MainTabParamList = {
  Fleet: undefined;
  Updates: undefined;
  Tasks: undefined;
};
