Perfect. I'll put together a detailed, step-by-step tutorial on building a hierarchical task planning network for AI planning using Reaflow in a React + TypeScript project. I'll walk through everything from setting up the environment and understanding Reaflow's core concepts, to rendering customized, interactive nodes and edges, and progressively building a task planning network.

I'll let you know when it's ready!

# Building a Hierarchical Task Planning Network with Reaflow (React + TypeScript)

## Introduction to Reaflow

Reaflow is an open-source React library for building node-based visualizations such as workflow editors, flowcharts, and diagrams. It provides a **Canvas** component that automatically arranges nodes and edges using the ELK graph layout engine. This means you define nodes and their connections, and Reaflow handles positioning them. For complex node structures with rich styling, Reaflow allows using `<foreignObject>` within the SVG node, enabling standard HTML rendering styled via CSS. Edges are SVG path elements.

In this tutorial, we will set up a React + TypeScript project with Reaflow to build a **Hierarchical Task Network (HTN)** visualization. Unlike Reaflow's built-in nested node capability, we will represent the hierarchy using a **flat graph structure**, where parent-child relationships are explicitly defined by **directed edges**. This approach can be simpler for some layout algorithms and data structures.

## Project Setup (React + TypeScript + Reaflow)

1.  **Initialize React+TypeScript Project (Vite recommended):**
    ```bash
    npm init vite@latest my-planner -- --template react-ts
    cd my-planner
    npm install
    ```
2.  **Install Reaflow:**
    ```bash
    npm install reaflow --save
    ```
3.  **Basic Setup (`src/App.tsx` and `src/App.css`):**
    Import `Canvas`, create basic app structure, and add minimal CSS for containers.

    ```tsx
    // src/App.tsx
    import React from "react";
    import { Canvas } from "reaflow";
    import "./App.css";

    const App: React.FC = () => (
      <div className="app-container">
        <h1>Flat HTN Planner</h1>
        <div className="canvas-container">
          <Canvas
            nodes={[]}
            edges={[]}
            fit={true}
            maxWidth={800}
            maxHeight={600}
          />
        </div>
      </div>
    );
    export default App;
    ```

    ```css
    /* src/App.css */
    .app-container {
      display: flex;
      flex-direction: column;
      align-items: center;
      padding: 20px;
      background-color: #f0f0f0;
    }
    .canvas-container {
      border: 1px solid #ccc;
      margin-bottom: 20px;
      background-color: white; /* Add dotted background later */
    }
    /* Info panel/Legend styles */
    .info-panel,
    .legend {
      border: 1px solid #ccc;
      padding: 15px;
      margin-top: 15px;
      background-color: white;
      width: 80%;
      max-width: 600px;
    }
    ```

    Run `npm run dev`. You should see the title and an empty canvas area.

## Defining Node Data Structure (Flat Graph)

Our nodes need to store type, title, description, icon, stats, etc. We define `MyNodeData`. Note the absence of the `parent` property, as hierarchy is handled by edges.

```typescript
// src/App.tsx
export interface MyNodeData {
  id: string;
  // No parent property needed
  width?: number; // Optional: Set fixed size for all nodes
  height?: number;
  data?: {
    type: string; // e.g., 'goal', 'subtask', 'action'
    title: string;
    description?: string;
    icon?: string; // Emoji or key
    stats?: Record<string, number>;
    showStats?: boolean;
    showError?: boolean;
  };
}

// Standard node dimensions
const NODE_WIDTH = 260;
const NODE_HEIGHT = 164;
```

## Creating Nodes and Edges for a Flat Hierarchy

In `App.tsx`, define `initialNodes` and `initialEdges`. The key difference is how hierarchy is represented:

- **Nodes:** All nodes are defined at the top level (no `parent` property).
- **Edges:** Create two types of edges:
  - **Hierarchy Edges:** Explicitly draw edges from parent nodes to child nodes (e.g., `goal` -> `boil`, `boil` -> `fill`). We'll give these a specific `className` (`edge-hierarchy`) for styling.
  - **Dependency/Sequence Edges:** Edges representing task order or constraints (e.g., `fill` -> `heat`, `heat` -> `pour`). These do _not_ get the hierarchy class name.

```typescript
// src/App.tsx (inside App component)
import React, { useState, useRef } from 'react';
// ... other imports

const initialNodes: MyNodeData[] = [
  // All nodes defined flat, assign width/height to all for consistency
  { id: 'goal', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'goal', title: 'Make Tea', ... } },
  { id: 'boil', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'subtask', title: 'Boil Water', ... } },
  { id: 'fill', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'action', title: 'Fill Kettle', ... } },
  { id: 'heat', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'action', title: 'Heat Kettle', ... } },
  { id: 'steep', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'subtask', title: 'Steep Tea', ... } },
  { id: 'place', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'action', title: 'Place Teabag', ... } },
  { id: 'pour', width: NODE_WIDTH, height: NODE_HEIGHT, data: { type: 'action', title: 'Pour Water', ... } },
];

const initialEdges = [
  // Hierarchy Edges (Parent -> Child)
  { id: 'goal-boil', from: 'goal', to: 'boil', className: 'edge-hierarchy' },
  { id: 'goal-steep', from: 'goal', to: 'steep', className: 'edge-hierarchy' },
  { id: 'boil-fill', from: 'boil', to: 'fill', className: 'edge-hierarchy' },
  { id: 'boil-heat', from: 'boil', to: 'heat', className: 'edge-hierarchy' },
  { id: 'steep-place', from: 'steep', to: 'place', className: 'edge-hierarchy' },
  { id: 'steep-pour', from: 'steep', to: 'pour', className: 'edge-hierarchy' },

  // Dependency/Sequence Edges
  { id: 'fill-heat', from: 'fill', to: 'heat' },
  { id: 'place-pour', from: 'place', to: 'pour' },
  { id: 'heat-pour', from: 'heat', to: 'pour' }, // Cross-hierarchy dependency
];

const App: React.FC = () => {
  const [nodes, setNodes] = useState<MyNodeData[]>(initialNodes);
  const [edges, setEdges] = useState(initialEdges);
  // ... state for selection, nodeIdCounter, etc. ...

  // ... handleNodeClick ...

  const handleAddNode = (parentNode: MyNodeData) => {
    const newNodeId = `node-${nodeIdCounter.current++}`;
    const newEdgeId = `${parentNode.id}-${newNodeId}`; // Edge from parent to new node

    const newNode: MyNodeData = {
      id: newNodeId,
      // No parent property on the node itself
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
      data: { type: 'action', title: 'New Task', ... },
    };

    // Create the new HIERARCHY edge
    const newEdge = {
      id: newEdgeId,
      from: parentNode.id,
      to: newNodeId,
      className: 'edge-hierarchy', // Style as a hierarchy edge
      // No parent property on the edge
    };

    setNodes(prev => [...prev, newNode]);
    setEdges(prev => [...prev, newEdge]);
  };

  // ... return statement rendering Canvas ...
};
```

## Customizing Nodes with `<foreignObject>`

We use the `<foreignObject>` approach as described previously to render HTML nodes styled with CSS. The `CustomNode.tsx` component implementation remains largely the same as in the previous update, using helper components `NodeContent` and `NodeStats`. Crucially, `CustomNode` **no longer needs the `nodes` prop** because it doesn't need to check for parent/child status – all nodes are rendered identically based on their `data`.

```typescript
// src/CustomNode.tsx (Key parts)
import React from "react";
import { NodeChildProps } from "reaflow";
import { MyNodeData } from "./App";
import { nodeConfig } from "./nodeConfig";

// ... NodeStats and NodeContent helper components (unchanged) ...

export interface CustomNodeProps {
  nodeProps: NodeChildProps<MyNodeData>;
  selectedNode: string | null;
  // No 'nodes' prop needed
  onNodeClick: (id: string) => void;
  onAddClick: (node: MyNodeData) => void;
}

export const CustomNode: React.FC<CustomNodeProps> = (
  {
    /* Destructure props */
  }
) => {
  const { node } = nodeProps;
  const width = node.width || NODE_WIDTH;
  const height = node.height || NODE_HEIGHT;
  const isSelected = selectedNode === node.id;
  const showAddButton = node.data?.type !== "end"; // Logic based on type

  return (
    <foreignObject x={0} y={0} width={width} height={height}>
      <div className="node-style-reset node-wrapper">
        <NodeContent
          node={node}
          selected={isSelected}
          onClick={() => onNodeClick(node.id)}
        />
        {showAddButton && (
          <div className="add-button">
            <button
              onClick={(e) => {
                /* ... */
              }}
            >
              +
            </button>
          </div>
        )}
      </div>
    </foreignObject>
  );
};
```

The `nodeConfig.ts` file remains unchanged.

## Styling the Flat Graph (`App.css`)

The CSS for the nodes (`.node-wrapper`, `.node-content`, etc.) remains the same as in the previous `<foreignObject>` example. The key addition is styling for the different edge types:

```css
/* src/App.css */

/* ... existing .app-container, .canvas-container, .info-panel, .legend styles ... */
/* ... existing node styles (.node-wrapper, .node-content, etc.) ... */

/* Edge Styles */
.reaflow-edge path {
  /* Default style for DEPENDENCY/SEQUENCE edges */
  stroke: #546e7a; /* Darker grey/blue */
  stroke-width: 1.5px;
  fill: none;
}

/* Style for HIERARCHY edges */
.reaflow-edge.edge-hierarchy path {
  stroke: #adb5bd; /* Lighter grey */
  stroke-dasharray: 5 2; /* Dashed line */
  stroke-width: 1px;
}

/* Arrowhead styles need to match */
/* Default arrowhead (solid, darker) - assumes default markerEnd */
.reaflow-edge .reaflow-marker path {
  fill: #546e7a;
}

/* Hierarchy arrowhead (lighter, matches dashed line) */
/* Requires using a specific marker referenced by hierarchy edges */
.reaflow-edge.edge-hierarchy .reaflow-marker path {
  fill: #adb5bd;
}
```

To apply the different arrowheads correctly, you'll typically define separate SVG `<marker>` elements in your `App.tsx` within a `<defs>` tag inside the `Canvas`, and reference the appropriate marker `id` from the `markerEnd` prop of the respective `<Edge>` components (or directly on the edge data object).

```typescript
// Simplified example in App.tsx return statement
<Canvas ... >
  {/* ... node prop ... */}
  edge={(
     <Edge
       // Default edge style (can use inline or rely on CSS)
       markerEnd="url(#arrow-default)"
     />
  )}
  // Pass edges array here
>
 { /* Define markers */}
 <defs>
   <marker id="arrow-default" ... >
     <path d="M 0 0 L 10 5 L 0 10 z" fill="#546e7a" />
   </marker>
   <marker id="arrow-hierarchy" ... >
      <path d="M 0 0 L 10 5 L 0 10 z" fill="#adb5bd" />
   </marker>
 </defs>
</Canvas>
```

And ensure your hierarchy edge data includes `markerEnd: 'url(#arrow-hierarchy)'` or similar.

## Tuning the Layout Engine (ELK Options)

For a flat graph representing a hierarchy, the `layered` algorithm is still very suitable. Ensure the direction and spacing are configured for clarity:

```typescript
// src/App.tsx
const layoutOptions: ElkCanvasLayoutOptions = {
  "elk.algorithm": "layered",
  "elk.direction": "DOWN",
  "elk.spacing.nodeNode": "80", // Spacing between siblings
  "elk.layered.spacing.nodeNodeBetweenLayers": "100", // Vertical spacing between levels
  "elk.portAlignment.default": "CENTER",
};
```

## Example: Flat Hierarchical AI Planning Diagram

Rendering the flat node/edge structure with the `layered` layout results in:

([image]()) _Figure: Flat hierarchical task planning graph. Hierarchy is shown by dashed grey edges flowing downwards. Solid darker edges show sequence/dependencies. All nodes are rendered at the same visual level but positioned by the layout algorithm based on connections._

The visual grouping is gone, but the parent-child relationships are clearly indicated by the styled hierarchy edges. The layout algorithm arranges nodes into layers based on these connections.

## Best Practices and Gotchas (Flat Graph Approach)

- **Clear Edge Styling:** Differentiating hierarchy edges from dependency edges visually (e.g., color, dash pattern, thickness) is crucial for readability.
- **Layout Algorithm Choice:** While `layered` works well, explore other ELK algorithms (`mrtree`, `stress`) if your graph structure becomes more complex or less tree-like.
- **Node Sizing:** Using consistent node sizes (`width`/`height` on all nodes) often leads to more predictable layouts in flat graphs compared to letting ELK size parent nodes.
- **Edge Routing:** ELK handles edge routing. Ensure spacing options (`nodeNode`, `nodeNodeBetweenLayers`) are sufficient to prevent excessive edge crossings or overlaps.
- **Data Complexity:** For very deep or wide hierarchies, a flat graph can become visually overwhelming. Consider alternative visualizations or interaction patterns (like expand/collapse) if needed.

## Conclusion (Updated)

This tutorial demonstrated an alternative approach to visualizing hierarchical data in Reaflow: using a **flat graph structure** where hierarchy is explicitly defined by **styled parent-to-child edges**, rather than visual nesting. We adapted the node data structure, modified the edge definitions, updated the `CustomNode` component to remove parent-specific logic, and applied distinct CSS styling to hierarchy versus dependency edges. This method, combined with ELK's `layered` layout algorithm, provides a clear way to represent task breakdowns and dependencies without relying on nested visual containers.

**References:** The Reaflow documentation and community examples were referenced for features and usage details ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=%2A%20)) ([Draw edge to nested node from other hierarchy level · Issue #26 · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/issues/26#:~:text=id%3A%20%272,)) ([reaflow/src/symbols/Node/Node.tsx at master · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/blob/master/src/symbols/Node/Node.tsx#L362#:~:text=onClick%3F%3A%20%28event%3A%20React.MouseEvent,void)).
