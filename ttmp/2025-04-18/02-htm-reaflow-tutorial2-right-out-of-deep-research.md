Perfect. I’ll put together a detailed, step-by-step tutorial on building a hierarchical task planning network for AI planning using Reaflow in a React + TypeScript project. I’ll walk through everything from setting up the environment and understanding Reaflow’s core concepts, to rendering customized, interactive nodes and edges, and progressively building a task planning network.

I’ll let you know when it’s ready!

# Building a Hierarchical Task Planning Network with Reaflow (React + TypeScript)

## Introduction to Reaflow

Reaflow is an open-source React library for building node-based visualizations such as workflow editors, flowcharts, and hierarchical diagrams. It provides a **Canvas** component that automatically arranges nodes and edges using the ELK graph layout engine ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=%2A%20)). This means you can define nodes and their connections, and Reaflow will handle positioning them in a readable layout. Reaflow supports features like zooming, panning, drag-and-drop edge creation, and even nested (hierarchical) nodes ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=,Node%2FEdge%20selection%20helper)). These capabilities make Reaflow an excellent choice for representing **Hierarchical Task Networks (HTNs)** or any scenario where high-level tasks break down into sub-tasks.

Under the hood, Reaflow represents your graph as an SVG. Each node is an SVG group (`<g>`) element that can contain shapes (like rectangles) and text, and edges are SVG path/line elements. By default, nodes are rendered as rectangular boxes with a text label, and edges are orthogonal connectors (with arrowheads for directed flow). Because Reaflow uses SVG, you have full control to style nodes/edges via props or custom components, and you can attach event handlers (onClick, onHover, etc.) for interactivity. In this tutorial, we will set up a React + TypeScript project with Reaflow, then walk through creating a hierarchical task planning diagram step-by-step – from basic nodes and edges to multi-level nested tasks with interactive behaviors.

## Project Setup (React + TypeScript + Reaflow)

Let’s start by setting up a new React project with TypeScript and installing Reaflow:

1. **Initialize React+TypeScript Project:** Use your preferred method to bootstrap a React TypeScript app. For example, using Create React App:

   ```bash
   npx create-react-app my-planner --template typescript
   cd my-planner
   ```

   Or with Vite:

   ```bash
   npm init vite@latest my-planner -- --template react-ts
   cd my-planner
   ```

   This will create a React project with TypeScript support.

2. **Install Reaflow:** Inside the project directory, install Reaflow from npm (it includes its own TypeScript declarations ([reaflow - npm](https://www.npmjs.com/package/reaflow#:~:text=Image%3A%20TypeScript%20icon%2C%20indicating%20that,in%20type%20declarations)) ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=Install%20the%20package%20via%20NPM%3A))):

   ```bash
   npm install reaflow --save
   ```

   This makes the `Canvas` and other Reaflow components available to import.

3. **Verify Setup:** Open the project in your code editor. In `src/App.tsx` (or wherever you will use the diagram), import the Reaflow Canvas and try a simple render to verify everything is working:

   ```tsx
   import React from "react";
   import { Canvas } from "reaflow";

   const App: React.FC = () => {
     return (
       <Canvas
         nodes={[]}
         edges={[]}
         fit={true}
         zoomable={false}
         maxHeight={400}
         maxWidth={600}
       />
     );
   };

   export default App;
   ```

   Here we render an empty Canvas with a fixed size. The props `maxHeight`/`maxWidth` ensure the diagram canvas is bounded (you can also control size via CSS or container style). We set `fit={true}` so the view auto-zooms to fit content, and `zoomable={false}` temporarily since we have no content yet. Run `npm start` to ensure the app loads without errors (you should see an empty white area where the canvas is). Now we’re ready to build our task network!

## Creating Basic Nodes and Edges

To understand Reaflow basics, let’s start by adding a couple of simple nodes and an edge between them. In Reaflow, you provide arrays of **node objects** and **edge objects** to the Canvas:

- **Node data:** Each node needs at least an `id` (unique string) and typically a `text` label to display. You can also specify visual or hierarchical properties (more on that later).
- **Edge data:** Each edge needs an `id` and references the ids of the nodes it connects (`from` and `to`). By default edges are directed (from -> to) and will render with an arrowhead.

Let’s create two nodes and one edge connecting them:

```tsx
// Within App.tsx
const nodes = [
  { id: "1", text: "Task 1" },
  { id: "2", text: "Task 2" },
];
const edges = [{ id: "1-2", from: "1", to: "2" }];

return <Canvas nodes={nodes} edges={edges} maxHeight={400} maxWidth={600} />;
```

In this code, we define two nodes with ids "1" and "2", and a single edge from node "1" to node "2". Reaflow will automatically lay out node 1 and node 2 and draw an arrow from 1 to 2 ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=Import%20the%20component%20into%20your,add%20some%20nodes%20and%20edges)) ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=text%3A%20%271%27%20%7D%2C%20,)). You don’t need to provide x/y coordinates for the nodes – the ELKJS layout will position them (by default in a top-down or left-right flow based on connectivity).

([image]()) _Figure: A simple Reaflow diagram with two nodes (“Task 1” and “Task 2”) connected by an edge. Reaflow’s automatic layout arranges the nodes and draws the directed edge._

If you run the app now, you should see two rectangular nodes labeled "Task 1" and "Task 2" with an arrow between them. Try resizing the browser window – by default Reaflow will keep the diagram centered and you can pan/zoom (unless `zoomable` is false). This basic example confirms our setup is working and illustrates the core concept: **define nodes and edges as data, and let Reaflow render the graph.**

## Representing Hierarchical Relationships (Parent-Child Nodes)

Now for the key feature: building a hierarchical task network where high-level tasks contain sub-tasks. Reaflow supports **nested nodes** (also called _subflows_ or parent-child nodes) out-of-the-box ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=,Node%2FEdge%20selection%20helper)). There are two ways to establish hierarchy in Reaflow:

- **Using `parent` property:** The simplest method is to give a node a `parent` field equal to the `id` of its parent node ([Draw edge to nested node from other hierarchy level · Issue #26 · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/issues/26#:~:text=id%3A%20%272,)). This automatically nests that node inside the parent node’s bounding box.
- **Using nested data structures:** Reaflow can also accept a nested array of child nodes within a node object (internally, it will treat it similarly). For clarity, we will use the `parent` property approach.

Let’s extend our example by making "Task 2" a sub-task inside "Task 1". We will treat "Task 1" as a parent container and "Task 2" as a child node:

```tsx
const nodes = [
  { id: "1", text: "Parent Task" },
  { id: "2", text: "Subtask A", parent: "1" },
  { id: "3", text: "Subtask B", parent: "1" },
];
const edges = [{ id: "2-3", from: "2", to: "3" }];
```

In this data, nodes "2" and "3" both have `parent: '1'`, meaning they will be rendered _inside_ node "1". We also added an edge from 2 to 3 to indicate an ordering or dependency between the subtasks (e.g. Subtask A must occur before Subtask B). The Canvas usage remains the same (`<Canvas nodes={nodes} edges={edges} ... />`).

**What happens visually?** Reaflow will draw node "1" as a **group node** or container, and draw nodes "2" and "3" within it. By default, the parent node "1" may be rendered as a larger rectangle outlining the boundary around its children (it might not have a filled color so that the children are visible inside). The children "Subtask A" and "Subtask B" will appear as normal nodes inside that container, with the edge from A to B drawn inside the parent. This creates a clear visual hierarchy: one big task containing two smaller tasks.

([image]()) _Figure: A parent node (“Parent Task”) containing two child nodes (“Subtask 1” and “Subtask 2”). The child nodes are connected by an arrow (showing their sequence). The parent task is rendered as a container (outer rounded rectangle) that groups its subtasks._

Notice in the figure how **Parent Task** is shown as an enclosing box around **Subtask 1** and **Subtask 2**. Reaflow’s layout ensures that the parent node’s size adjusts to fit its children. The `text` label of the parent (if any) is typically shown at the top-center of the container. In our code we labeled it "Parent Task", so you should see that title on the group. The children are laid out inside (here we got a vertical stacking with an arrow from Subtask 1 to Subtask 2). You can add edges from or to nested nodes just like any others – Reaflow will route them properly across hierarchy boundaries. For example, if we had another node outside the parent, we could connect it to a child node (Reaflow will draw the connecting lines into/out of the container) ([Draw edge to nested node from other hierarchy level · Issue #26 · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/issues/26#:~:text=%7B%20id%3A%20%271,from%3A%20%272%27%2C%20to%3A%20%273%27)) ([Draw edge to nested node from other hierarchy level · Issue #26 · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/issues/26#:~:text=id%3A%20%272,)).

### Multi-Level Hierarchy (Sub-subtasks)

You can nest nodes to multiple levels. For instance, a subtask can itself have children by using the same `parent` property. There is essentially no hard limit to nesting depth (aside from keeping IDs unique and managing complexity). Each parent node becomes a container for its children, and the Canvas will handle drawing all nested relationships.

**Important:** When using nested nodes, ensure each node’s `id` is unique across the entire graph (even across levels). This avoids any ambiguity when referencing `parent` or edges. Also, keep in mind that a parent node’s textual label might overlap with child nodes if the parent is small – by default, Reaflow expands parent nodes to fit children, so this usually isn’t an issue, but test different label lengths or manually adjust styling if needed.

Now that we can represent hierarchy, let’s look at how to customize the appearance and interactivity of these nodes.

## Customizing Nodes and Edges (Styling & Interactive Behavior)

Reaflow provides several ways to customize how nodes and edges are rendered and how they behave:

- **Custom Node Components:** You can override the default node rendering by supplying a custom component or using the provided `<Node>` element with custom props. This allows you to change node shapes, add icons, or even embed custom JSX inside a node.
- **Styling via props:** The built-in components have props for basic styling (like `style`, `className`, `label`, `icon`, `width/height`, etc.). For example, a Node can be given a `style` object to adjust its SVG rectangle’s appearance (fill color, stroke, border radius).
- **Event Handlers:** Nodes and edges support events such as `onClick`, `onMouseEnter`/`onMouseLeave` (hover), `onDragStart`/`onDragEnd`, etc., which you can use to make the diagram interactive ([reaflow/src/symbols/Node/Node.tsx at master · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/blob/master/src/symbols/Node/Node.tsx#L362#:~:text=onClick%3F%3A%20%28event%3A%20React.MouseEvent,void)) ([reaflow/src/symbols/Node/Node.tsx at master · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/blob/master/src/symbols/Node/Node.tsx#L362#:~:text=match%20at%20L1318%20onClick%3F)). This is great for reacting to user actions (e.g., clicking a task node to show details or mark it complete).
- **Ports and Edge Interaction:** By default, Reaflow nodes have connection points (ports) that allow creating or rearranging edges via drag-and-drop. There are props to customize this (like enabling an “add edge” handle on a node or customizing the appearance of the connecting port), but for a planning network you might keep the default behavior or even disable interactive edge creation if not needed.

Let’s illustrate some of these customizations:

### 1. Changing Node Appearance

Say we want to distinguish different types of tasks by color or shape. The simplest way is using the **style prop** on nodes. We can provide a `node` renderer to Canvas that applies a style. For example, to make all nodes light green with rounder corners:

```tsx
import { Canvas, Node } from "reaflow";

<Canvas
  nodes={nodes}
  edges={edges}
  node={<Node style={{ fill: "#c8e6c9", stroke: "#388e3c" }} rx={6} ry={6} />}
/>;
```

Here we import the `<Node>` component from Reaflow and use it inside the Canvas via the `node` prop. We pass a `style` object to set the SVG rectangle fill color and stroke (border) color. The props `rx` and `ry` set the corner radius for the rectangle (making it more rounded). This will apply to all nodes by default. You could also target specific nodes: each node data object can include a `style` or custom property which your custom Node component reads (via the `properties` prop inside Node). For instance, if our node data had a `type` field, we could conditionally set style based on that in a Node render function.

For more complex rendering, Reaflow allows the `Node`’s children to be a function that returns arbitrary JSX ([reaflow/src/symbols/Node/Node.tsx at master · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/blob/master/src/symbols/Node/Node.tsx#L362#:~:text=match%20at%20L1552%20,Fragment)). This means you can draw custom SVG elements or even use a foreignObject to include HTML content inside a node. For example, you might create a custom node that displays an icon and some text:

```tsx
<Canvas node={
  <Node width={100} height={60}>
    {(nodeProps) => (
      <>
        <rect width="100" height="60" fill="#fff8e1" stroke="#ffb300" rx={4} ry={4}/>
        <text x={50} y={30} textAnchor="middle" fill="#000">
          {nodeProps.properties.text}
        </text>
      </>
    )}
  </Node>
} ... />
```

In this snippet, we override the default node rendering completely. We draw a rectangle and a text element manually (centering the text). We use `nodeProps.properties.text` which contains the label from our node data. You could add more JSX here – even images or shapes – to suit your needs. (Note: When doing custom children like this, ensure you size the Node via `width`/`height` props and match your drawing to those dimensions).

### 2. Adding Tooltips

To add a tooltip that appears when you hover over a node, you have a couple of approaches:

- **SVG `<title>` element:** The easiest method is to include a `<title>` tag inside the Node’s SVG group. SVG titles will display as native tooltips on hover. For example, if our node data has a `description` field, our custom Node could include `<title>{nodeProps.properties.description}</title>` inside.
- **Custom Tooltip Component:** For richer tooltips (styled boxes, etc.), you can use state to track the “hovered” node and render a React component positioned near the node. Reaflow’s `onMouseEnter` and `onMouseLeave` events can help here. For instance, we could maintain a state like `const [hovered, setHovered] = useState<string|null>(null)` for the hovered node id. Then:
  ```tsx
  <Canvas node={
    <Node
      onEnter={(_, node) => setHovered(node.id)}
      onLeave={(_, node) => setHovered(prev => prev === node.id ? null : prev)}
      // ...other props
    />
  } ... />
  {hovered && <Tooltip nodeId={hovered} />}
  ```
  In this pseudocode, when a node’s onEnter fires, we store its id, and onLeave we clear it. The `<Tooltip>` component (which you create) would take the nodeId and perhaps lookup more info to display. You’d also need to position the tooltip; you can get node coordinates from Reaflow if needed (via Canvas callbacks or by querying the DOM). For simplicity, using the SVG `<title>` might be sufficient for most cases.

### 3. Click Handlers and Selection

Handling clicks on nodes or edges is straightforward. You can use `onClick` on the Node component. For example, to alert the node’s name when clicked:

```tsx
<Canvas node={
  <Node onClick={(_event, node) => alert(`Clicked ${node.text}`)} />
} ... />
```

The click handler receives the node data object (here named `node` which includes `id`, `text`, etc.). In a real app, instead of an `alert`, you might toggle some state to mark the task as selected and then, say, highlight it or show details in a side panel. Reaflow has a built-in notion of selection as well (each node has a `selectable` prop, true by default, and Canvas can indicate which node is selected). You could leverage that or manage selection in your own state.

For example, to highlight a node on click, you could store a `selectedId` in state and then pass a different style for the Node if `node.id === selectedId`. This can be done by providing a custom Node child function that checks the `nodeProps.properties.id`. Alternatively, Reaflow might provide a selection style out of the box (check the documentation for any `selection` props on Canvas/Node).

### 4. Edge Customization

Edges can also be customized. You can import the `<Edge>` component from 'reaflow' and pass an `edge` prop to Canvas similar to how we did for Node. For instance, you could change edge colors, or add an **“add button”** on edges to insert a new node (Reaflow has an `add` prop on Edge to facilitate this) ([vitest for adding nodes from edges · reaviz reaflow · Discussion #218 · GitHub](https://github.com/reaviz/reaflow/discussions/218#:~:text=Now%2C%20I%20would%20like%20to,onAdd%20prop%20for%20Edge%20component)). In a planning network, you might not need fancy edge types, but it’s good to know you can adjust things like making edges curved or straight. By default, Reaflow edges use orthogonal routing with elbow points. If you prefer straight lines, you could try setting an ELK option or use a different edge type if provided.

As an example, to simply make edges thicker and red, you could do:

```tsx
<Canvas edge={<Edge style={{ stroke: 'red', strokeWidth: 2 }} />} ... />
```

This will clone the default edge but apply the given SVG stroke style.

## Example: Hierarchical AI Planning Diagram

Now let’s put it all together with a concrete example. Suppose we have an AI planning scenario: **Making a Cup of Tea**. This is a goal that can be broken into sub-tasks and actions:

- **Goal:** Make Tea (top-level goal)
  - **Task 1:** Boil Water (first subtask of making tea)
    - **Action 1.1:** Fill Kettle (child of Boil Water)
    - **Action 1.2:** Heat Kettle (child of Boil Water)
  - **Task 2:** Steep Tea (second subtask of making tea)
    - **Action 2.1:** Place Teabag (child of Steep Tea)
    - **Action 2.2:** Pour Water (child of Steep Tea)

And we know that _Task 2 (Steep Tea)_ cannot start until _Task 1 (Boil Water)_ is done (specifically, you must have hot water to pour). We will model this by an edge from the last action of Boil Water to the first action of Steep Tea.

Here’s how we can define this in Reaflow data:

```tsx
const nodes = [
  { id: "goal", text: "Make Tea" },
  // Task 1: Boil Water (parent = goal)
  { id: "boil", text: "Boil Water", parent: "goal" },
  { id: "fill", text: "Fill Kettle", parent: "boil" },
  { id: "heat", text: "Heat Kettle", parent: "boil" },
  // Task 2: Steep Tea (parent = goal)
  { id: "steep", text: "Steep Tea", parent: "goal" },
  { id: "place", text: "Place Teabag", parent: "steep" },
  { id: "pour", text: "Pour Water", parent: "steep" },
];
const edges = [
  // sequence within Boil Water:
  { id: "fill-heat", from: "fill", to: "heat" },
  // sequence within Steep Tea:
  { id: "place-pour", from: "place", to: "pour" },
  // dependency between Boil Water and Steep Tea:
  { id: "heat-pour", from: "heat", to: "pour" },
];
```

A few things to note in this data:

- We list **Make Tea** (`id: "goal"`) with no parent (it’s top-level). This will be the outer container for everything.
- **Boil Water** and **Steep Tea** both have `parent: "goal"`, making them siblings inside the Make Tea container.
- The low-level actions have parents `"boil"` or `"steep"` accordingly, nesting them one level deeper.
- Edges: We added two intra-task edges (`fill -> heat` and `place -> pour`) to denote order within each task. We also added an inter-task edge (`heat -> pour`) to denote that “Pour Water” (in Steep Tea) depends on “Heat Kettle” (in Boil Water). Reaflow supports edges that connect nodes across different parents; it will draw them entering or leaving the parent containers as needed ([Draw edge to nested node from other hierarchy level · Issue #26 · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/issues/26#:~:text=%7B%20id%3A%20%271,from%3A%20%272%27%2C%20to%3A%20%273%27)).

Now, rendering this with `<Canvas nodes={nodes} edges={edges} />` will produce a multi-level diagram. Make sure to enable scroll/zoom if the diagram is larger than the viewport. You might also want to set `fit={true}` on Canvas so it auto-zooms to fit all nodes initially.

([image]()) _Figure: Hierarchical task planning graph for the "Make Tea" example. The top-level goal **Make Tea** contains two subtask groups (**Boil Water** and **Steep Tea**). Each subtask group contains actions (yellow nodes), and edges indicate the sequence. Notice the arrow from **Heat Kettle** to **Pour Water** crossing between the two groups, representing a dependency between the subplans._

In the figure, **Make Tea** is the outer container (with a bold label at top). Inside it, **Task: Boil Water** and **Task: Steep Tea** are shown as sub-containers (with slightly shaded backgrounds). Within **Boil Water**, the actions **Fill Kettle** and **Heat Kettle** are ordered top-down. Within **Steep Tea**, **Place Teabag** and **Pour Water** are ordered. The edge from _Heat Kettle_ to _Pour Water_ goes from the Boil Water container to the Steep Tea container, indicating that the Steep Tea subtask waits for Boil Water to finish heating. This kind of visualization is very useful in AI planning to see the hierarchy of tasks and the dependencies between them.

You can style this further (for example, we used different fill colors for demonstration: perhaps yellow for primitive actions, gray for subtask containers). All these can be achieved with custom Node rendering as discussed earlier. If you implemented this in code, you now have an interactive diagram: you can click on nodes, maybe highlight them or show tooltips as configured. The structure is defined entirely by the data we passed in.

## Best Practices and Gotchas for Reaflow Node-Based Apps

Building node-based applications with Reaflow is powerful, and here are some best practices and tips to keep things running smoothly:

- **Organize Your Data:** Maintain your `nodes` and `edges` in state or context so that you can easily add, remove, or update nodes. If you modify the nodes/edges arrays, the Canvas will update the layout automatically. Use stable IDs for nodes so that React/Reaflow can track elements – avoid reusing an old ID for a new node without unmounting the previous one.
- **Unique IDs:** As mentioned, ensure every node and edge ID is unique. A common pitfall is forgetting to change the `id` when copying objects. Non-unique ids can lead to rendering bugs or lost events.
- **Performance with Large Graphs:** Reaflow’s layout (ELKJS) is sophisticated but can be heavy for very large graphs. For big planning networks, consider using the `animate={false}` prop on Canvas (to disable layout animations) and possibly throttle how often you update the layout. Reaflow will re-layout whenever the nodes/edges props change; if you are doing frequent updates (e.g., streaming in changes), try batching them.
- **Using Layout Options:** Reaflow allows passing ELK layout options via the Canvas `layoutOptions` prop. For example, you can specify the algorithm (like layered vs tree), spacing, hierarchy constraints, etc. The defaults are usually fine, but if you need a specific ordering (e.g., all parent tasks left-to-right), you could experiment with those options as documented in ELK. (Refer to Reaflow docs for how to format these options).
- **Interactive Edges and Ports:** By default, users can drag from a node’s port to create a new edge. If your use case doesn’t need users adding connections, you might disable this. You can set `node={<Node dragType="default" />}` to remove the special multi-port dragging behavior, or simply ignore new edge events. Conversely, if you _do_ allow users to build the plan graph, Reaflow’s built-in handlers (like an `onAdd` event on Edge for when a new edge is drawn ([vitest for adding nodes from edges · reaviz reaflow · Discussion #218 · GitHub](https://github.com/reaviz/reaflow/discussions/218#:~:text=Now%2C%20I%20would%20like%20to,onAdd%20prop%20for%20Edge%20component))) can be used to insert a node or connect nodes in your state.
- **Styling Considerations:** Reaflow’s components come with default styles (e.g., class names for selected nodes, or for nodes that contain children). If you want to use CSS to override styles, inspect the rendered DOM to find class names or use the provided `className` prop on Node/Edge to apply your own. Often, using the `style` prop as we did is sufficient for colors and basic appearance. For consistency, define a color scheme for different levels or types of tasks (perhaps use one color for goal nodes, another for subtask containers, another for primitive actions).
- **Refer to Storybook Examples:** The official Reaflow documentation site (Storybook) has many demos covering features like **undo/redo**, **selection**, **drag & drop**, etc. If you plan to build a full-fledged editor for task networks, those examples are invaluable ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=Quick%20Links)). For instance, they demonstrate adding a small “+” button on edges to insert nodes, keyboard shortcuts, and more.

By following these practices, you can avoid common issues and create a smooth user experience. For example, always test how your diagram behaves when tasks are added or removed dynamically – the layout will change to accommodate new nodes, which can sometimes shift things around. You might want to lock certain positions or disable layout during an active drag. Reaflow gives you control to do that (via properties like `animated` or using `onLayoutChange` events to intervene).

## Conclusion

In this tutorial, we covered how to build a hierarchical task planning network using Reaflow in a React+TypeScript application. We started with a simple two-node graph and progressed to nested, multi-level task structures. Along the way, we learned how Reaflow’s architecture (powered by ELKJS) automatically handles layout and how we can customize node appearance and interactivity. We created an illustrative AI planning example (“Make Tea” HTN) to see a complex hierarchical graph in action.

With Reaflow, you can focus on the logic of your planning application – defining tasks, subtasks, and their relationships – and let the library take care of rendering a clear, interactive visualization. The combination of hierarchical grouping and interactive features (click, hover, drag) means you can not only display plans but also allow users to explore and edit them intuitively.

Feel free to explore further with Reaflow’s extensive features: for example, adding **icons** to nodes to represent task types, using **conditional styling** to highlight critical paths, or integrating with state management so that updates to your planning logic immediately reflect in the diagram. Reaflow is a versatile library, and with the step-by-step approach from this tutorial, you should be well on your way to building robust node-based planning UIs. Happy coding!

**References:** The Reaflow documentation and community examples were referenced for features and usage details ([reaflow | Yarn](https://classic.yarnpkg.com/en/package/reaflow#:~:text=%2A%20)) ([Draw edge to nested node from other hierarchy level · Issue #26 · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/issues/26#:~:text=id%3A%20%272,)) ([reaflow/src/symbols/Node/Node.tsx at master · reaviz/reaflow · GitHub](https://github.com/reaviz/reaflow/blob/master/src/symbols/Node/Node.tsx#L362#:~:text=onClick%3F%3A%20%28event%3A%20React.MouseEvent,void)).
