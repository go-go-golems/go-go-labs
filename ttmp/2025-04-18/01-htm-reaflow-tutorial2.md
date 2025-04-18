# Building an Interactive Hierarchical Task Planning Network with Reaflow (React + TypeScript)

This tutorial guides you through building an interactive, hierarchical task planning network visualization using the Reaflow library within a React and TypeScript project. We'll cover setup, core Reaflow concepts, creating nested nodes, advanced customization using `foreignObject` for HTML/CSS styling, and handling user interactions.

## Introduction to Reaflow

Reaflow is an open-source React library designed for creating node-based diagrams like flowcharts, organizational charts, and workflow editors. Its core strength lies in the **`<Canvas>`** component, which leverages the powerful **Eclipse Layout Kernel (ELK)** engine ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=%2A%20)) to automatically arrange nodes and edges. You define the structure (nodes, hierarchy, connections), and Reaflow handles the visual positioning.

Key features relevant to our task planning network include:

- **Automatic Layout:** ELK arranges nodes and routes edges, minimizing overlaps.
- **Hierarchical Rendering:** Built-in support for nesting nodes within parents (`parent` property), automatically adjusting parent sizes.
- **Interactivity:** Supports zooming, panning, and event handling (clicks, hovers) on nodes/edges.
- **Customization:** Allows extensive styling and rendering customization, including embedding custom React components or HTML within nodes.

These capabilities make Reaflow ideal for visualizing **Hierarchical Task Networks (HTNs)**, where complex goals decompose into sub-tasks and actions.

Under the hood, Reaflow renders diagrams as SVG. By default, nodes are simple rectangles with text, and edges are orthogonal lines. However, as we'll see, you can override this rendering significantly using features like SVG's `<foreignObject>` to embed rich HTML content within nodes, allowing for complex styling with standard CSS.

## Project Setup (React + TypeScript + Reaflow using Vite)

Let's set up a modern React project using Vite and install Reaflow.

1.  **Initialize React+TypeScript Project with Vite:** Vite provides a fast development experience.

    ```bash
    # Create the project directory
    mkdir my-planner
    cd my-planner

    # Initialize Vite project within the current directory
    npm init vite@latest . -- --template react-ts
    ```

    Follow the prompts if any. This creates a basic React + TypeScript project structure.

2.  **Install Dependencies:** Install React, Reaflow, and development dependencies.

    ```bash
    npm install
    npm install reaflow --save
    ```

    This installs React, Reaflow, and their dependencies. Reaflow includes its own TypeScript types ([reaflow - npm](https://www.npmjs.com/package/reaflow#:~:text=Image%3A%20TypeScript%20icon%2C%20indicating%20that,in%20type%20declarations)).

3.  **Basic Verification (Optional):** You can run `npm run dev` and check if the default Vite app loads correctly in your browser (usually at `http://localhost:5173` or the next available port).

## Creating Basic Nodes and Edges

Reaflow diagrams are driven by data. You provide arrays of node and edge objects to the `<Canvas>` component.

- **Node Data:** At minimum, requires a unique `id`. Typically includes `text` for labels, and can include `parent` for hierarchy, `width`/`height` for explicit sizing, and custom `data` fields.
- **Edge Data:** Requires a unique `id`, and `from`/`to` fields referencing the connected node IDs.

Let's modify `src/App.tsx` to render two nodes and an edge:

```tsx
// src/App.tsx
import React from "react";
import { Canvas } from "reaflow";
import "./App.css"; // We'll add styles later

// Define node data structure
interface MyNodeData {
  id: string;
  text: string;
  parent?: string;
  width?: number;
  height?: number;
}

const App: React.FC = () => {
  const nodes: MyNodeData[] = [
    { id: "1", text: "Task 1", width: 150, height: 50 },
    { id: "2", text: "Task 2", width: 150, height: 50 },
  ];
  const edges = [{ id: "1-2", from: "1", to: "2" }];

  return (
    <div className="app-container">
      <h1>Simple Reaflow Diagram</h1>
      <div className="canvas-container">
        <Canvas
          nodes={nodes}
          edges={edges}
          maxWidth={600}
          maxHeight={400}
          fit={true} // Zoom/pan to fit content initially
        />
      </div>
    </div>
  );
};

export default App;
```

We also need basic CSS for layout in `src/App.css` and `src/index.css`:

```css
/* src/index.css - Basic body/root styles */
:root {
  font-family: "Segoe UI", Tahoma, Geneva, Verdana, sans-serif;
  /* ... other base styles */
  color: #213547;
  background-color: #ffffff;
}
body {
  margin: 0;
  display: flex;
  min-height: 100vh;
}
#root {
  width: 100%;
}
/* ... (optional) dark mode styles */
@media (prefers-color-scheme: dark) {
  :root {
    color: rgba(255, 255, 255, 0.87);
    background-color: #242424;
  }
  /* Adjust component backgrounds for dark mode if needed later */
}
```

```css
/* src/App.css - App layout styles */
.app-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 2rem;
  text-align: center;
}
.canvas-container {
  width: 100%;
  height: 400px; /* Match maxHeight */
  margin: 2rem 0;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  overflow: hidden;
  background-color: #fafafa;
}
/* Dark mode adjustment for canvas */
@media (prefers-color-scheme: dark) {
  .canvas-container {
    background-color: #2d2d2d !important;
    border-color: #444 !important;
  }
}
```

Running `npm run dev` should now display two rectangular nodes ("Task 1", "Task 2") connected by an arrow, automatically positioned by Reaflow.

## Representing Hierarchical Relationships (Parent-Child Nodes)

HTNs involve tasks containing subtasks. Reaflow handles this via the `parent` property on a node, setting its value to the `id` of the parent node.

Let's modify `App.tsx` to make "Task 2" and a new "Task 3" children of "Task 1":

```tsx
// src/App.tsx (inside App component)
const nodes: MyNodeData[] = [
  // Parent node - Let ELK determine size
  { id: "1", text: "Parent Task" },
  // Child nodes - Give explicit size
  { id: "2", text: "Subtask A", parent: "1", width: 150, height: 50 },
  { id: "3", text: "Subtask B", parent: "1", width: 150, height: 50 },
];
const edges = [
  // Edge between children
  { id: "2-3", from: "2", to: "3" },
];

// Canvas height might need adjustment
return (
  <div className="app-container">
    <h1>Hierarchical Diagram</h1>
    <div className="canvas-container" style={{ height: "500px" }}>
      {" "}
      {/* Increased height */}
      <Canvas
        nodes={nodes}
        edges={edges}
        maxWidth={600}
        maxHeight={500} // Match container height
        fit={true}
        direction="DOWN" // Suggest layout direction
      />
    </div>
  </div>
);
```

**Important Layout Concept:** Notice we removed `width` and `height` from the parent node (`id: '1'`). When dealing with hierarchy, it's generally best practice to **let the layout engine (ELK) determine the size of parent nodes**. ELK will automatically calculate the required space to encompass all children plus internal padding. Providing explicit sizes for parents can conflict with this, leading to overlapping nodes as seen in earlier debugging steps. We _do_ provide sizes for the leaf nodes ('2', '3') as they don't contain others.

Running this will show "Parent Task" rendered as a larger container node (you might see its bounding box) with "Subtask A" and "Subtask B" positioned inside it, connected by an edge. ELK manages the internal layout and parent sizing.

## Advanced Customization: `foreignObject` and CSS Styling

While Reaflow allows basic styling via props on `<Node>` and `<Edge>`, complex UI often requires more control. Reaflow supports embedding arbitrary HTML content within nodes using SVG's `<foreignObject>`. This unlocks the full power of HTML and CSS for node appearance.

Let's refactor our node rendering to use this technique, aiming for a style similar to the reference `@Diagram` example.

**1. Refactor Node Rendering in `App.tsx`:**

```tsx
// src/App.tsx
import React, { useState } from "react";
// Import NodeChildProps for typing
import { Canvas, Node, Edge, NodeChildProps } from "reaflow";
import "./App.css";

// Interface for our node data (if not already defined)
interface MyNodeData {
  id: string;
  text: string;
  parent?: string;
  width?: number;
  height?: number;
}

const NODE_WIDTH = 260; // Standard width for leaf nodes
const NODE_HEIGHT = 100; // Standard height for leaf nodes

const App: React.FC = () => {
  // Use the "Make Tea" example data
  const [nodes] = useState<MyNodeData[]>([
    // Parent nodes (no size specified)
    { id: "goal", text: "Make Tea" },
    { id: "boil", text: "Boil Water", parent: "goal" },
    { id: "steep", text: "Steep Tea", parent: "goal" },
    // Leaf nodes (actions with explicit size)
    {
      id: "fill",
      text: "Fill Kettle",
      parent: "boil",
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    },
    {
      id: "heat",
      text: "Heat Kettle",
      parent: "boil",
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    },
    {
      id: "place",
      text: "Place Teabag",
      parent: "steep",
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    },
    {
      id: "pour",
      text: "Pour Water",
      parent: "steep",
      width: NODE_WIDTH,
      height: NODE_HEIGHT,
    },
  ]);

  const [edges] = useState([
    { id: "fill-heat", from: "fill", to: "heat" },
    { id: "place-pour", from: "place", to: "pour" },
    { id: "heat-pour", from: "heat", to: "pour" }, // Cross-hierarchy dependency
  ]);

  // State to track the selected node ID
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  return (
    <div className="app-container">
      <h1>Hierarchical Task Planning Network</h1>
      <p>A visualization of an AI planning scenario: Making a Cup of Tea</p>

      <div className="canvas-container" style={{ height: "750px" }}>
        {" "}
        {/* Adjusted size */}
        <Canvas
          nodes={nodes}
          edges={edges}
          maxWidth={1000} // Adjusted size
          maxHeight={750} // Adjusted size
          fit={true}
          direction="DOWN"
          node={
            // Provide a Node component with custom rendering via children function
            <Node>
              {(nodeProps: NodeChildProps) => {
                // Typed props provide node data and context
                // Cast the generic node data to our specific type
                const nodeData = nodeProps.node as MyNodeData;
                // Use explicit size if available (leaf), otherwise ELK determines parent size
                const nodeWidth = nodeData.width || nodeProps.width; // Fallback to width calculated by ELK/Reaflow
                const nodeHeight = nodeData.height || nodeProps.height; // Fallback to height calculated by ELK/Reaflow
                const isSelected = selectedNode === nodeData.id;

                // Determine CSS class based on hierarchy level
                let nodeLevelClass = "node-level-leaf"; // Default to leaf
                if (nodeData.parent === undefined) {
                  nodeLevelClass = "node-level-root";
                } else if (
                  nodes.find(
                    (n) => n.id === nodeData.parent && n.parent === undefined
                  )
                ) {
                  // If parent is a root node, this is a subtask
                  nodeLevelClass = "node-level-subtask";
                }

                return (
                  // foreignObject allows embedding HTML
                  <foreignObject
                    x={0}
                    y={0}
                    width={nodeWidth}
                    height={nodeHeight}
                  >
                    <div
                      // Apply CSS classes for styling and level indication
                      className={`node-wrapper ${nodeLevelClass}`}
                      // Indicate selection for CSS styling
                      aria-selected={isSelected}
                      // Handle clicks within the HTML content
                      onClick={(e) => {
                        e.stopPropagation(); // Prevent Reaflow's default drag/select
                        setSelectedNode(nodeData.id);
                        console.log(`Clicked: ${nodeData.text}`);
                      }}
                    >
                      {/* Define HTML structure for the node content */}
                      <div className="node-content">
                        <div className="node-details">
                          <h1>{nodeData.text}</h1>
                          <p>
                            ID: {nodeData.id}
                            {nodeData.parent
                              ? ` (Parent: ${nodeData.parent})`
                              : ""}
                          </p>
                        </div>
                        {/* Can add icons or other elements here */}
                      </div>
                      {/* Use standard HTML title attribute for simple tooltips */}
                      <title>{`${nodeData.text} (ID: ${nodeData.id})`}</title>
                    </div>
                  </foreignObject>
                );
              }}
            </Node>
          }
          edge={
            // Optional: Customize edge appearance
            <Edge style={{ stroke: "#78909c", strokeWidth: 1.5 }} />
          }
        />
      </div>

      {/* Info Panel to display selected node details */}
      <div className="info-panel">
        <h3>Selected Node</h3>
        {selectedNode ? (
          <div>
            <p>
              <strong>ID:</strong> {selectedNode}
            </p>
            <p>
              <strong>Text:</strong>{" "}
              {nodes.find((n) => n.id === selectedNode)?.text || ""}
            </p>
            <p>
              <strong>Parent:</strong>{" "}
              {nodes.find((n) => n.id === selectedNode)?.parent || "None"}
            </p>
          </div>
        ) : (
          <p>Click on a node to see its details</p>
        )}
      </div>

      {/* Legend to explain node styling */}
      <div className="legend">
        <h3>Legend</h3>
        <div className="legend-item">
          <div className="legend-color node-level-root"></div>
          <span>Goal (Top Level)</span>
        </div>
        <div className="legend-item">
          <div className="legend-color node-level-subtask"></div>
          <span>Subtasks (Group)</span>
        </div>
        <div className="legend-item">
          <div className="legend-color node-level-leaf"></div>
          <span>Actions (Leaf)</span>
        </div>
      </div>
    </div>
  );
};

export default App;
```

**Key Changes in `App.tsx`:**

- **`<Node>` Children Function:** We provide a function as the child of `<Node>`. This function receives `nodeProps` (typed as `NodeChildProps`) containing details about the node being rendered (`nodeProps.node` holds our data), its calculated size (`nodeProps.width`, `nodeProps.height`), etc.
- **`<foreignObject>`:** Inside the function, we return an SVG `<foreignObject>` element. This element acts as a bridge, allowing standard HTML (`<div>`, `<h1>`, `<p>`, etc.) to be rendered within the SVG canvas at the node's position.
- **HTML Structure:** Inside the `<foreignObject>`, we create a `div` with class `node-wrapper`. This wrapper gets classes based on hierarchy (`node-level-root`, `node-level-subtask`, `node-level-leaf`) and an `aria-selected` attribute for styling selected nodes.
- **Data Access:** We access our node data via `nodeProps.node` (after casting it to `MyNodeData`).
- **Click Handling:** An `onClick` handler is added to the `node-wrapper` div. `e.stopPropagation()` is important here to prevent Reaflow's default canvas-level interactions (like dragging the node) when clicking inside our custom HTML. We update the `selectedNode` state.
- **Info Panel & Legend:** Added simple React components below the canvas to show details of the `selectedNode` and explain the color coding defined in CSS.

**2. Add Corresponding CSS in `App.css`:**

```css
/* src/App.css */

/* ... (existing .app-container, .canvas-container styles) ... */

/* Styles for the Info Panel and Legend */
.info-panel,
.legend {
  margin: 2rem auto;
  padding: 1rem;
  max-width: 400px;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  background-color: #fff;
  text-align: left;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}
.info-panel h3,
.legend h3 {
  margin-top: 0;
  color: #1976d2; /* Match title color */
  border-bottom: 1px solid #e0e0e0;
  padding-bottom: 0.5rem;
  margin-bottom: 1rem;
}
.legend-item {
  display: flex;
  align-items: center;
  margin-bottom: 0.5rem;
}
.legend-color {
  width: 20px;
  height: 20px;
  margin-right: 10px;
  border: 1px solid #ccc; /* Base border */
  border-radius: 4px;
  /* Keep border-left separate for override */
}

/* Styles for HTML nodes rendered via foreignObject */
.node-wrapper {
  width: 100%;
  height: 100%;
  background-color: #fff; /* Default background */
  border: 1px solid #ccc; /* Default border */
  border-radius: 4px;
  border-left-width: 5px; /* Emphasize the colored border */
  transition: border-color 0.3s, box-shadow 0.3s;
  cursor: pointer;
  overflow: hidden; /* Prevent content overflow */
  display: flex; /* Ensure content aligns well */
  flex-direction: column; /* Stack content vertically */
  box-sizing: border-box; /* Include padding/border in size */
}

/* Selection Style */
.node-wrapper[aria-selected="true"] {
  border-color: #f44336; /* Red border for selection */
  box-shadow: 0 0 8px rgba(244, 67, 54, 0.5);
}

/* Hierarchy Level Styles (using border-left-color) */
.node-wrapper.node-level-root {
  border-left-color: #2196f3; /* Blue for root */
  background-color: #e3f2fd;
}
.node-wrapper.node-level-subtask {
  border-left-color: #4caf50; /* Green for subtasks */
  background-color: #e8f5e9;
}
.node-wrapper.node-level-leaf {
  border-left-color: #ffc107; /* Amber for leaves */
  background-color: #fff8e1;
}

/* Match legend colors to node styles */
.legend-color.node-level-root {
  background-color: #e3f2fd;
  border-left: 5px solid #2196f3;
}
.legend-color.node-level-subtask {
  background-color: #e8f5e9;
  border-left: 5px solid #4caf50;
}
.legend-color.node-level-leaf {
  background-color: #fff8e1;
  border-left: 5px solid #ffc107;
}

.node-content {
  flex-grow: 1; /* Allow content to fill space */
  padding: 10px;
  overflow: hidden; /* Prevent internal overflow */
}

.node-details h1 {
  margin: 0 0 5px 0;
  font-size: 14px;
  font-weight: bold;
  color: #333;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.node-details p {
  margin: 0;
  font-size: 12px;
  color: #666;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Critical: Make Reaflow's default node invisible when using foreignObject */
.reaflow-node > rect {
  fill: transparent !important;
  stroke: transparent !important;
}

/* Dark mode adjustments */
@media (prefers-color-scheme: dark) {
  .info-panel,
  .legend {
    background-color: #333 !important;
    border-color: #444 !important;
  }
  .legend-color {
    border-color: #555 !important;
  }
  .node-wrapper {
    background-color: #424242;
    border-color: #616161;
  }
  .node-wrapper.node-level-root {
    background-color: #1e2c3b;
  }
  .node-wrapper.node-level-subtask {
    background-color: #2e3b2e;
  }
  .node-wrapper.node-level-leaf {
    background-color: #4d402c;
  }
  .node-details h1 {
    color: #eee;
  }
  .node-details p {
    color: #ccc;
  }
}
```

**Key CSS Changes:**

- **Node Structure Styles:** Added rules for `.node-wrapper`, `.node-content`, `.node-details` to control layout (using flexbox), padding, and text overflow.
- **Level/Selection Styling:** Uses `border-left-color` and `background-color` for hierarchy levels, and `border-color`/`box-shadow` for selection (`[aria-selected='true']`).
- **Legend Styling:** Updated `.legend-color` styles to match the node appearance.
- **Hiding Default Node:** The `.reaflow-node > rect { ... }` rule is essential. When using `foreignObject`, you typically want to hide the default SVG rectangle that Reaflow might still render underneath.
- **Dark Mode:** Added adjustments for node backgrounds and text colors in dark mode.

Now, running `npm run dev` should display the "Make Tea" hierarchical diagram using styled HTML nodes. You can click nodes to select them (highlighted with a red border and details shown in the info panel), and hover to see the simple SVG title tooltip. The layout should be correct because we are letting ELK size the parent containers.

## Best Practices and Gotchas for Reaflow HTN Apps

- **Parent Node Sizing:** For hierarchical layouts, let ELK/Reaflow determine parent node sizes by omitting `width`/`height` on container nodes. Specify sizes only for leaf nodes or nodes where specific dimensions are required.
- **`foreignObject` for Rich Styling:** Use `<foreignObject>` when you need complex HTML structure, CSS styling, or embedded React components within nodes. Remember to hide the default SVG node (`.reaflow-node > rect`) via CSS.
- **Event Handling in `foreignObject`:** When adding interactive elements (like buttons or complex click handlers) inside the `foreignObject`'s HTML, use `event.stopPropagation()` to prevent conflicts with Reaflow's canvas-level interactions (drag, pan, selection).
- **Unique IDs:** Ensure all node and edge `id`s are unique across the entire graph.
- **Performance:** For very large graphs, consider disabling animations (`animated={false}` on Canvas) and potentially tuning ELK layout options (`layoutOptions` prop on Canvas) for speed or specific layout algorithms. See ELK documentation for available options.
- **State Management:** Keep your `nodes` and `edges` in React state (e.g., `useState`, `useReducer`, or a state management library) so the diagram updates automatically when the data changes.
- **Layout Stability:** Adding/removing nodes dynamically will cause the layout to reflow. If this is disruptive during user interaction (like dragging), you might temporarily disable updates or layout calculations.

## Conclusion & Next Steps

We've successfully built an interactive hierarchical task planning visualization using Reaflow, React, and TypeScript. We leveraged Reaflow's automatic layout and hierarchy features, and significantly customized the node appearance using `<foreignObject>` and CSS. We also added interactivity through node selection and informational panels.

This project demonstrates the power of Reaflow for visualizing structured, hierarchical data like HTNs. From here, you could extend this by:

- Adding icons to nodes.
- Implementing drag-and-drop to modify the plan.
- Showing task status (e.g., completed, blocked) with different styles.
- Integrating with a real AI planning backend.

Refer to the `README.md` in the project directory for running instructions and `docs/CONCEPTS.md` for more background on HTNs and Reaflow's applicability. The official Reaflow documentation and Storybook examples ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=Quick%20Links)) are excellent resources for exploring more advanced features.
