Perfect, I'll create a tutorial that shows how to render a hierarchical task network (HTN) using Reaflow in a React + Vite app. It will cover parsing JSON HTN input, mapping it to nodes and edges, handling collapsible sub-networks, and optionally styling nodes. I'll also include some ready-made JSON presets for testing.

I'll get started and let you know when the tutorial is ready for review.

# Rendering a Hierarchical Task Network (HTN) with Reaflow in React + Vite

In this tutorial, we will build a **hierarchical task planning network (HTN) visualizer** using **Reaflow** in a React application powered by **Vite**. We'll walk through accepting an HTN defined in JSON, converting it into Reaflow's node/edge format, and rendering an interactive, collapsible task hierarchy diagram. Along the way, we'll cover setting up the React+Vite project, parsing the JSON, implementing expandable sub-networks, and customizing the appearance of nodes and edges.

## Overview of Reaflow and Project Setup

Reaflow is a **React library for building workflow diagrams and flow charts**. It provides a high-performance, highly customizable graph canvas component for rendering nodes and edges ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=REAFLOW%20is%20a%20modular%20diagram,complex%20visualizations%20with%20total%20customizability)). Notably, Reaflow supports **nesting of nodes/edges** (hierarchical graphs) out-of-the-box ([reaflow-extended | Yarn](https://classic.yarnpkg.com/en/package/reaflow-extended#:~:text=,Undo%2FRedo%20helper)), which makes it ideal for visualizing an HTN structure. We will leverage these capabilities to create an interactive HTN diagram.

**Project Setup:** We'll start by creating a new React project using Vite and installing Reaflow:

```bash
# Create a new Vite React project
npm create vite@latest htn-visualizer -- --template react

cd htn-visualizer

# Install Reaflow library
npm install reaflow

# Start the development server
npm run dev
```

This sets up a basic React+Vite app and adds Reaflow as a dependency. Now we can use Reaflow's components in our React code. According to the official docs, Reaflow's primary component is the `Canvas`, which we can import and use to render nodes and edges ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=Import%20the%20component%20into%20your,add%20some%20nodes%20and%20edges)).

## Defining the HTN JSON Input

Our HTN will be provided as **structured JSON**. Let's assume an HTN is represented as a nested object where each task can have a list of `subtasks`. Each task has an `id` (unique identifier) and a `name` (label). For example, consider an HTN for _"Build a House"_:

```json
{
  "id": "task1",
  "name": "Build a House",
  "subtasks": [
    {
      "id": "task1.1",
      "name": "Lay Foundation"
    },
    {
      "id": "task1.2",
      "name": "Build Walls",
      "subtasks": [
        { "id": "task1.2.1", "name": "Install Doors" },
        { "id": "task1.2.2", "name": "Install Windows" }
      ]
    },
    {
      "id": "task1.3",
      "name": "Install Roof"
    }
  ]
}
```

In this JSON, **"Build a House"** is the top-level task with three subtasks. One of those subtasks (_"Build Walls"_) itself has two subtasks. This hierarchical structure could continue to further depths as needed. In a real application, you might allow users to input their own JSON or select from preset examples. For our tutorial, we'll define a couple of sample JSON presets in the code for demonstration.

## Parsing HTN JSON into Reaflow Nodes and Edges

Reaflow's `Canvas` requires data as an array of **nodes** and an array of **edges**. Each node should have a unique `id` and display text/label, and each edge defines a connection from a source node to a target node (by their ids). In Reaflow v5, for example, a node object might look like `{ id: '1', text: 'Task Name' }` and an edge like `{ id: '1-2', from: '1', to: '2', text: '...' }` ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=nodes%3D,)). We need to convert our nested HTN JSON into this flat list of nodes and edges.

We can achieve this by **traversing the JSON recursively**. For each task encountered, create a node entry. For each subtask, create an edge from the parent task to the subtask, then recurse into the subtask. This approach is similar to how some JSON visualization tools transform JSON into graph data for Reaflow ([JSONCrack Codebase Analysis — Part 4.2.1.1 — JsonEditor — debouncedUpdateJson | by TThroo | Medium](https://medium.com/@tthroo/jsoncrack-codebase-analysis-part-4-2-1-1-jsoneditor-debouncedupdatejson-e51f66166e93#:~:text=So%2C%20in%20the%20end%2C%20parser,Reaflow%20to%20render%20a%20map)). Let's write a helper function to do this conversion:

```jsx
// Define types for clarity (optional)
type Task = {
  id: string,
  name: string,
  subtasks?: Task[],
};

// Helper to generate nodes and edges from an HTN task object
function generateGraphFromHTN(
  task: Task,
  parentId: string | null = null,
  nodes: any[] = [],
  edges: any[] = []
) {
  // Create a node for the current task
  const nodeId = task.id;
  nodes.push({ id: nodeId, text: task.name });

  // If there's a parent task, create an edge from parent to this task
  if (parentId) {
    edges.push({ id: `${parentId}->${nodeId}`, from: parentId, to: nodeId });
  }

  // Recurse into subtasks if any
  if (task.subtasks) {
    for (const subtask of task.subtasks) {
      generateGraphFromHTN(subtask, nodeId, nodes, edges);
    }
  }
  return { nodes, edges };
}
```

In the code above, we accumulate nodes and edges in the arrays. Each task becomes a node, and each parent-child relationship becomes an edge. For example, given the "Build a House" JSON, this function will output nodes like `[{id: "task1", text: "Build a House"}, {id: "task1.1", text: "Lay Foundation"}, ...]` and edges like `[{id: "task1->task1.1", from: "task1", to: "task1.1"}, ...]`. We can test it quickly:

```jsx
// Example usage:
const htnData = { ... JSON from above ... };
const graphData = generateGraphFromHTN(htnData);
console.log(graphData.nodes, graphData.edges);
```

At this point, we have a flat list of nodes and edges ready to render with Reaflow.

## Rendering the HTN Graph with Reaflow's Canvas

Now, let's integrate Reaflow into our React app to display the graph. We will create a React component (e.g., `<HTNGraph />`) that takes the generated nodes and edges and renders a `Canvas`.

First, import the `Canvas` (and optionally other components) from Reaflow:

```jsx
import { Canvas } from "reaflow";
```

We can then use the `Canvas` component in our JSX. A minimal usage passes in the `nodes` and `edges` arrays as props ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=Import%20the%20component%20into%20your,add%20some%20nodes%20and%20edges)) ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=nodes%3D,)):

```jsx
function HTNGraph({ nodes, edges }) {
  return (
    <Canvas
      nodes={nodes}
      edges={edges}
      fit={true} // auto-fit the view to the graph
      direction="DOWN" // lay out top-down (hierarchical)
      zoomable={true} // allow zooming/panning
      pannable={true}
    />
  );
}
```

Here we set a few useful props:

- `fit={true}`: makes the graph auto-fit into the canvas on load (so you see the whole network).
- `direction="DOWN"`: tells Reaflow to lay out the graph vertically top-to-bottom (which makes sense for a task hierarchy). Reaflow uses the ELK layout engine under the hood for automatic layouts ([reaflow-extended | Yarn](https://classic.yarnpkg.com/en/package/reaflow-extended#:~:text=,Undo%2FRedo%20helper)), so our tasks should appear in a top-down tree structure by default.
- `zoomable` and `pannable`: enable interactive zooming and panning of the diagram.

At this stage, if we pass in the nodes/edges from our earlier conversion, the HTN would be visualized as a static directed acyclic graph (tree). For example, the "Build a House" HTN would appear with "Build a House" at the top, arrows pointing to "Lay Foundation", "Build Walls", and "Install Roof" beneath it, and further arrows from "Build Walls" to its subtasks.

([image]()) _Figure: An example of a fully expanded HTN graph for the "Build a House" scenario. Each task is a node, and arrows point from a task to its subtasks. In our Reaflow app, we will make these subtask groups collapsible._

## Implementing Collapsible Sub-networks (Expandable Tasks)

The real power of an interactive HTN viewer is the ability to **collapse or expand** sub-networks of tasks. In our context, that means we want to hide or show the subtasks of a given task on demand (for example, clicking on a task node toggles its subtasks).

**Approach:** We will control which nodes/edges are rendered based on an _expanded/collapsed state_. One way to do this is to maintain a set of expanded node IDs in React state. Initially, you might start with no tasks expanded (or perhaps just the top-level tasks expanded), then when a user clicks a node, update the state to add or remove that node from the expanded set. We will then regenerate the `nodes` and `edges` lists to reflect the new state.

Let's enhance our graph generation to consider expansion state:

```jsx
// Maintain a set of expanded node IDs
const [expandedNodes, setExpandedNodes] = useState < Set < string >> new Set();

// Modified generator that only includes children if their parent is expanded
function generateGraphFromHTNVisible(
  task: Task,
  parentId: string | null = null,
  nodes: any[] = [],
  edges: any[] = []
) {
  const nodeId = task.id;
  nodes.push({
    id: nodeId,
    text:
      task.name +
      (task.subtasks // indicate collapsible
        ? expandedNodes.has(nodeId)
          ? " (-)"
          : " (+)"
        : ""),
    data: { hasChildren: !!task.subtasks }, // store extra info if needed
  });
  if (parentId) {
    edges.push({ id: `${parentId}->${nodeId}`, from: parentId, to: nodeId });
  }
  // Only traverse into subtasks if this node is expanded
  if (task.subtasks) {
    if (expandedNodes.has(nodeId)) {
      for (const subtask of task.subtasks) {
        generateGraphFromHTNVisible(subtask, nodeId, nodes, edges);
      }
    }
  }
  return { nodes, edges };
}
```

In this version, if a task has subtasks and is **not** in the `expandedNodes` set, we do not add its subtasks to the graph. We also append a `" (+)"` or `" (-)"` to the node label to indicate if it can be expanded or collapsed. (A more elegant approach could be to draw an expand/collapse icon on the node, but using a text indicator keeps things simple here.) We also put a flag `hasChildren` in the node's `data` for potential styling use.

Now we need to handle toggling the expansion state when a user interacts. Reaflow's `Canvas` allows us to capture node click events via the `onNodeClick` prop. We can use this to toggle the expansion:

```jsx
// Click handler to expand/collapse nodes
function handleNodeClick(node: NodeData) {
  if (node.data.hasChildren) {
    setExpandedNodes(prevSet => {
      const newSet = new Set(prevSet);
      if (prevSet.has(node.id)) {
        newSet.delete(node.id);    // collapse if already expanded
      } else {
        newSet.add(node.id);       // expand if collapsed
      }
      return newSet;
    });
  }
}
...
// In the Canvas component:
<Canvas
  ...
  onNodeClick={handleNodeClick}
/>
```

Whenever a node with children is clicked, we update the `expandedNodes` state. We use a `Set` for quick lookup and toggling. After updating, we need to recalculate the graph data (nodes and edges). In React, we can do this by calling our `generateGraphFromHTNVisible` function inside a `useEffect` that depends on `expandedNodes` (and the original `htnData`):

```jsx
useEffect(() => {
  const { nodes, edges } = generateGraphFromHTNVisible(htnData);
  setGraphData({ nodes, edges });
}, [htnData, expandedNodes]);
```

Here, `setGraphData` would set a state holding the current visible graph nodes/edges (the `graphData` can be stored in a state similar to `expandedNodes`). Each time the expanded set changes (or if a new HTN is loaded), we compute the new graph structure to render.

Now our `<HTNGraph>` component can use `graphData.nodes` and `graphData.edges` from state, and it will reflect the expanded/collapsed view. Initially, `expandedNodes` might be an empty set, meaning all tasks start collapsed (only top-level tasks visible). If the top-level of your HTN is a single root task, you might choose to initialize that as expanded so that its first level of subtasks is shown by default. This is up to the desired UX.

**At runtime:** The user will see the top-level tasks. A task that has hidden subtasks will show a "(-)" in its label. Clicking it triggers `handleNodeClick`, which adds that task's ID to `expandedNodes`. The effect runs, generating new nodes/edges including that task's children, and the graph re-renders showing the expanded sub-network. If they click the task again (now "-" in the label), it collapses, removing its children from the rendered graph.

This mechanism is essentially what the interactive Reaflow demo on CodeSandbox implements for expandable nodes (in that demo, clicking parent nodes toggles their child nodes in the diagram). Our approach mirrors that functionality.

## Customizing Node and Edge Appearance

Reaflow allows for extensive customization of how nodes and edges are drawn. By default, nodes might appear as dark rectangles with white text, and edges as straight lines with arrows. We can tailor the style to better represent our HTN.

**Approach: Custom Components**

A clean way to customize rendering in Reaflow is to create dedicated React components for nodes and edges and pass instances of these components to the `Canvas` props `node` and `edge`. This approach is used in some Reaflow examples and keeps the rendering logic separate from the main graph component.

**1. Create a Custom Edge Component (`CustomEdge.tsx`)**

For our HTN, the default edge appearance is mostly fine, but we might want to ensure consistent styling or hide labels. We create a simple wrapper around Reaflow's `Edge`:

```tsx
// src/components/CustomEdge.tsx
import { Edge as ReaflowEdge, EdgeProps } from "reaflow";

export function CustomEdge(props: EdgeProps) {
  return (
    <ReaflowEdge
      {...props}
      style={{ stroke: "#555" }} // Basic style
      label={null} // No label needed for hierarchy
    />
  );
}
```

**2. Create a Custom Node Component (`CustomNode.tsx`)**

This component will handle the visual representation of our tasks. We want to:

- Color-code nodes based on whether they have children (compound vs. primitive tasks).
- Display the task name and its unique ID.
- Use `foreignObject` to allow standard HTML/CSS for layout within the SVG node.

```tsx
// src/components/CustomNode.tsx
import { Node as ReaflowNode, NodeProps, NodeChildProps } from "reaflow";
import { NodeData } from "../types/htn"; // Assuming NodeData type is defined

// Props received by the wrapper component from Reaflow
interface CustomNodeWrapperProps extends NodeProps {
  properties: NodeData; // Reaflow nests the node data here when passing component instances
}

// Props for the inner content renderer
interface NodeContentProps {
  nodeData: NodeData;
}

// Renders the actual HTML content inside the node
const NodeContent: React.FC<NodeContentProps> = ({ nodeData }) => {
  const { id, text, data } = nodeData;
  const hasChildren = data?.hasChildren ?? false;

  return (
    // Style the div using standard CSS
    <div
      style={{
        padding: "10px 15px",
        borderRadius: "4px",
        background: hasChildren ? "#D5E8D4" : "#FFF2CC", // Greenish for parents, yellowish for leaves
        border: "1px solid #888",
        textAlign: "center",
        color: "#333",
        height: "100%", // Ensure div fills foreignObject
        display: "flex",
        flexDirection: "column",
        justifyContent: "center",
      }}
    >
      <div style={{ fontWeight: "bold", marginBottom: "5px" }}>{text}</div>
      <div style={{ fontSize: "0.8em", color: "#666" }}>ID: {id}</div>
    </div>
  );
};

// The main component passed to Reaflow's 'node' prop
export function CustomNode({
  properties,
  ...restProps
}: CustomNodeWrapperProps) {
  return (
    <ReaflowNode
      {...restProps} // Pass down calculated props like x, y
      rx={5}
      ry={5} // Add rounded corners to the node shape
    >
      {(event: NodeChildProps) => (
        // Use foreignObject to render HTML inside SVG
        <foreignObject x={0} y={0} width={event.width} height={event.height}>
          <NodeContent nodeData={properties} />
        </foreignObject>
      )}
    </ReaflowNode>
  );
}
```

Key points here:

- The outer `CustomNode` receives the node data via a `properties` prop (this is how Reaflow passes data when using component instances for the `node` prop).
- It uses `foreignObject` to embed HTML within the SVG node, allowing for easier styling and layout with standard CSS.
- The `NodeContent` component handles the actual display logic (styling, showing text and ID).

**3. Update the Canvas Usage (`HTNGraph.tsx`)**

Now, we import these custom components and pass instances to the `Canvas`.

```tsx
// src/components/HTNGraph.tsx
import { Canvas } from "reaflow";
import { NodeData, EdgeData } from "../types/htn";
import { CustomNode } from "./CustomNode"; // Import custom node
import { CustomEdge } from "./CustomEdge"; // Import custom edge

interface HTNGraphProps {
  /* ... */
}

export function HTNGraph({ nodes, edges, onNodeClick }: HTNGraphProps) {
  // IMPORTANT: When passing a component instance to 'node',
  // Reaflow expects the data payload in a 'properties' field.
  const nodesWithProperties = nodes.map((node) => ({
    ...node,
    properties: node,
  }));

  // Handle click events passed up from CustomNode
  const handleNodeClick = (
    event: React.MouseEvent<Element, MouseEvent>,
    node: NodeData
  ) => {
    onNodeClick(node, event); // Forward to App component's handler
  };

  return (
    <Canvas
      nodes={nodesWithProperties} // Use nodes formatted for component instances
      edges={edges}
      node={<CustomNode onClick={handleNodeClick} />} // Pass INSTANCE of CustomNode
      edge={<CustomEdge />} // Pass INSTANCE of CustomEdge
      fit={true}
      direction="DOWN"
      zoomable={true}
      pannable={true}
      readonly // Often useful if not implementing drag/edit
    />
  );
}
```

Notice that we now map the original `nodes` array to add the `properties` field containing the node data, as required by Reaflow when using this custom component pattern. We also pass the `handleNodeClick` callback to our `CustomNode` instance.

With these changes, Reaflow will use our `CustomNode` and `CustomEdge` components to render the graph elements, giving us the desired appearance (color-coded nodes showing ID) and utilizing a component-based customization approach that aligns well with Reaflow's API design as seen in various examples.

## Integrating Sample Presets and Final App Structure

To tie it all together, you might structure your React component as follows:

```jsx
// Sample presets
const examples: { [key: string]: Task } = {
  houseProject: { ... },   // the Build a House JSON
  // other examples...
};

function App() {
  const [htnData, setHtnData] = useState(examples.houseProject);
  const [expandedNodes, setExpandedNodes] = useState(new Set<string>());
  const [graphData, setGraphData] = useState<{nodes: NodeData[], edges: EdgeData[]}>({ nodes: [], edges: [] });

  // Regenerate graph when htnData or expandedNodes changes
  useEffect(() => {
    const { nodes, edges } = generateGraphFromHTNVisible(htnData);
    setGraphData({ nodes, edges });
  }, [htnData, expandedNodes]);

  const handleNodeClick = useCallback((node: NodeData) => {
    if (node.data.hasChildren) {
      setExpandedNodes(prev => {
        const newSet = new Set(prev);
        if (prev.has(node.id)) newSet.delete(node.id);
        else newSet.add(node.id);
        return newSet;
      });
    }
  }, []);

  return (
    <div>
      {/* Optionally, UI to select different HTN presets */}
      <HTNGraph nodes={graphData.nodes} edges={graphData.edges} onNodeClick={handleNodeClick} />
    </div>
  );
}
```

In the above pseudocode, `HTNGraph` would be a wrapper around the `Canvas` as shown previously (including the `node` and `edge` customization). We use React state to handle data and expansion state, and `useEffect` to recompute the visible graph whenever those change. The preset selection UI could be as simple as a dropdown or buttons that call `setHtnData(exampleX)` with different JSON structures.

With this setup running (start the dev server with `npm run dev`), you should be able to interact with the HTN diagram in the browser. Clicking on tasks will expand or collapse their subtasks, and you can pan or zoom the view as needed to explore large task networks.

**Additional Tips:**

- _Layout adjustments:_ Reaflow's layout engine should handle most cases well. If your HTN is very large or complex, you might tweak layout options or use a different layout algorithm (Reaflow supports tree, layered, radial layouts, etc.). By default, a downward hierarchical layout is used which fits HTNs logically.
- _Performance:_ For extremely deep or large HTNs, consider virtualizing parts of the graph or limiting how many levels can be expanded at once, to avoid overwhelming the browser.
- _Further customization:_ You can listen to other events (like edge clicks, etc.) or add buttons to nodes (e.g., an explicit expand/collapse icon) using custom node rendering. The interactive Reaflow demo on CodeSandbox demonstrates many of these capabilities, and you can reference it for more advanced patterns.

## Conclusion

We have created a React + Vite application that takes an HTN described in JSON and visualizes it using Reaflow. By parsing the JSON into Reaflow's node and edge format and managing an expanded/collapsed state, we achieved an interactive hierarchical view where task groups can be expanded or collapsed. We also applied custom styling to make the diagram more informative and visually appealing.

Reaflow's powerful graph rendering engine made it straightforward to implement this HTN viewer – from handling automatic layouts to enabling rich customization for nodes and edges. With this foundation, you can extend the functionality further, such as editing the HTN via the graph, adding tooltips or descriptions on nodes, or integrating with other planning tools. Happy coding, and happy planning!

**Sources:**

- Reaflow Documentation and README (reaviz/reaflow) ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=REAFLOW%20is%20a%20modular%20diagram,complex%20visualizations%20with%20total%20customizability)) ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=Import%20the%20component%20into%20your,add%20some%20nodes%20and%20edges)) ([reaflow - npm](https://www.npmjs.com/package/reaflow/v/3.1.0?activeTab=code#:~:text=nodes%3D,))
- Reaflow Features (hierarchical graphs, etc.) ([reaflow-extended | Yarn](https://classic.yarnpkg.com/en/package/reaflow-extended#:~:text=,Undo%2FRedo%20helper))
- _JSONCrack_ open-source project analysis – converting JSON to Reaflow graph
