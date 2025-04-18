import { NodeData as ReaflowNodeData } from 'reaflow';

export type Task = {
  id: string;
  name: string;
  subtasks?: Task[];
};

// Use Reaflow's NodeData as the base, making 'text' optional as it is in the source.
// Add our custom 'data' property structure.
export type NodeData = ReaflowNodeData<{
  hasChildren?: boolean;
}> & {
  // We ensure text is always provided by our generator, but make it optional
  // here to align with the base ReaflowNodeData type used internally.
  text?: string;
};

// Remove the unused custom EdgeData type
/*
export type EdgeData = {
  id: string;
  from: string;
  to: string;
  text?: string;
};
*/
