# Hierarchical Task Planning Network with Reaflow

This project demonstrates how to build a hierarchical task planning network visualization using [Reaflow](https://reaflow.dev/), a React library for node-based visualizations.

The example shows an AI planning scenario: "Making a Cup of Tea", with hierarchical tasks and dependencies.

## Features

- Hierarchical node structure (parent-child relationships)
- Automatic layout of nodes and edges
- Node styling based on hierarchy level
- Interactive node selection
- Information panel showing selected node details
- Color legend for node types
- Responsive design

## Getting Started

### Prerequisites

- Node.js (v14+)
- npm or yarn

### Installation

1. Clone the repository
2. Navigate to the project directory
3. Install dependencies:

```bash
npm install
```

### Running the Application

Start the development server:

```bash
npm run dev
```

Open your browser and visit: `http://localhost:5173`

## How It Works

This application demonstrates key Reaflow concepts:

1. **Hierarchical Structure**: Using the `parent` property to create nested node relationships
2. **Node Styling**: Applying different styles based on node hierarchy level
3. **Interactive Features**: Implementing click handlers to select nodes
4. **Edge Connections**: Defining edges between nodes, including cross-hierarchy connections
5. **Auto Layout**: Using Reaflow's automatic layout capabilities

## Application Structure

- **App.tsx**: Main component that defines the nodes, edges, and rendering logic
- **App.css**: Styles for the application UI
- **index.css**: Base styles and dark mode support

## Learn More

- [Reaflow Documentation](https://reaflow.dev/)
- [React Documentation](https://react.dev/)

## Customizing the Visualization

To modify the task network:

1. Edit the `nodes` array in `App.tsx` to add/remove tasks
2. Edit the `edges` array to modify task dependencies
3. Update the styling in the `Node` component to change node appearance

## Acknowledgements

Based on the tutorial from the hierarchical task planning network tutorial.
