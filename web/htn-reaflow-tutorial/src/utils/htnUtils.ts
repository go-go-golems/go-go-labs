import { Task, NodeData, EdgeData } from "../types/htn";

export function generateGraphFromHTNVisible(
  task: Task,
  expandedNodes: Set<string>,
  parentId: string | null = null,
  nodes: NodeData[] = [],
  edges: EdgeData[] = []
): { nodes: NodeData[]; edges: EdgeData[] } {
  const nodeId = task.id;
  nodes.push({
    id: nodeId,
    text:
      task.name +
      (task.subtasks ? (expandedNodes.has(nodeId) ? " (-)" : " (+)") : ""),
    data: { hasChildren: !!task.subtasks },
  });

  if (parentId) {
    edges.push({
      id: `${parentId}->${nodeId}`,
      from: parentId,
      to: nodeId,
    });
  }

  if (task.subtasks && expandedNodes.has(nodeId)) {
    for (const subtask of task.subtasks) {
      generateGraphFromHTNVisible(subtask, expandedNodes, nodeId, nodes, edges);
    }
  }

  return { nodes, edges };
}

// Sample HTN data
export const examples: { [key: string]: Task } = {
  houseProject: {
    id: "task1",
    name: "Build a House",
    subtasks: [
      {
        id: "task1.1",
        name: "Lay Foundation",
      },
      {
        id: "task1.2",
        name: "Build Walls",
        subtasks: [
          { id: "task1.2.1", name: "Install Doors" },
          { id: "task1.2.2", name: "Install Windows" },
        ],
      },
      {
        id: "task1.3",
        name: "Install Roof",
      },
    ],
  },
  softwareProject: {
    id: "sw1",
    name: "Develop Software",
    subtasks: [
      {
        id: "sw1.1",
        name: "Requirements Analysis",
        subtasks: [
          { id: "sw1.1.1", name: "Gather Requirements" },
          { id: "sw1.1.2", name: "Document Specifications" },
        ],
      },
      {
        id: "sw1.2",
        name: "Implementation",
        subtasks: [
          { id: "sw1.2.1", name: "Write Code" },
          { id: "sw1.2.2", name: "Write Tests" },
        ],
      },
      {
        id: "sw1.3",
        name: "Deployment",
        subtasks: [
          { id: "sw1.3.1", name: "Build Release" },
          { id: "sw1.3.2", name: "Deploy to Production" },
        ],
      },
    ],
  },
};
