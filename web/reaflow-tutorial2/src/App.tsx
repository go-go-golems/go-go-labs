import React, { useState, useRef } from 'react';
import { Canvas, Node, Edge, NodeChildProps, ElkCanvasLayoutOptions } from 'reaflow';
import './App.css';
import { CustomNode } from './CustomNode'; // Import the new component
import { nodeConfig } from './nodeConfig'; // Import node config

// Define an interface for our custom node data
export interface MyNodeData {
  id: string;
  // parent?: string; // No longer needed for flat structure
  text?: string; // Keep text for compatibility or specific uses if needed
  width?: number;
  height?: number;
  // Add fields inspired by reachat-codesandbox
  data?: {
    type: string; // e.g., 'goal', 'subtask', 'action', 'source', 'email', 'wait', 'sms', 'end'
    title: string;
    description?: string;
    icon?: string; // Placeholder for icon representation (e.g., emoji, text)
    stats?: Record<string, number>;
    showStats?: boolean;
    showError?: boolean; // For badge
  };
}

// Define node dimensions based on reachat-codesandbox/style.ts
const NODE_WIDTH = 260; // From example
const NODE_HEIGHT = 164; // From example (includes button space)

// Define Layout Options
const layoutOptions: ElkCanvasLayoutOptions = {
  'elk.algorithm': 'layered',
  'elk.direction': 'DOWN',
  'elk.spacing.nodeNode': '80',
  'elk.portAlignment.default': 'CENTER',
  'elk.layered.spacing.nodeNodeBetweenLayers': '80',
};

// Initial Data - Flat structure
const initialNodes: MyNodeData[] = [
  {
    id: 'goal',
    width: NODE_WIDTH, height: NODE_HEIGHT, // Give all nodes a size now
    data: { type: 'goal', title: 'Make Tea', description: "Top-level objective" }
  },
  {
    id: 'boil',
    width: NODE_WIDTH, height: NODE_HEIGHT,
    data: { type: 'subtask', title: 'Boil Water', description: "Heat water to boiling point" }
  },
  {
    id: 'fill',
    width: NODE_WIDTH, height: NODE_HEIGHT,
    data: { type: 'action', title: 'Fill Kettle', description: "Add water to the kettle", icon: '💧', stats: { 'Time': 5, 'Units': 1 }, showStats: true }
  },
  {
    id: 'heat',
    width: NODE_WIDTH, height: NODE_HEIGHT,
    data: { type: 'action', title: 'Heat Kettle', description: "Activate heating element", icon: '🔥', showError: true } // Example error
  },
  {
    id: 'steep',
    width: NODE_WIDTH, height: NODE_HEIGHT,
    data: { type: 'subtask', title: 'Steep Tea', description: "Infuse tea leaves in hot water" }
  },
  {
    id: 'place',
    width: NODE_WIDTH, height: NODE_HEIGHT,
    data: { type: 'action', title: 'Place Teabag', description: "Put teabag in cup", icon: '🍵' }
  },
  {
    id: 'pour',
    width: NODE_WIDTH, height: NODE_HEIGHT,
    data: { type: 'action', title: 'Pour Water', description: "Add hot water to cup", icon: '➡️', stats: { 'Volume': 200, 'Temp': 95 }, showStats: true }
  }
];

// Edges: Represent hierarchy and dependencies
const initialEdges = [
  // Hierarchy Edges
  { id: 'goal-boil', from: 'goal', to: 'boil', className: 'edge-hierarchy' },
  { id: 'goal-steep', from: 'goal', to: 'steep', className: 'edge-hierarchy' },
  { id: 'boil-fill', from: 'boil', to: 'fill', className: 'edge-hierarchy' },
  { id: 'boil-heat', from: 'boil', to: 'heat', className: 'edge-hierarchy' },
  { id: 'steep-place', from: 'steep', to: 'place', className: 'edge-hierarchy' },
  { id: 'steep-pour', from: 'steep', to: 'pour', className: 'edge-hierarchy' },

  // Dependency/Sequence Edges (within original hierarchy levels)
  { id: 'fill-heat', from: 'fill', to: 'heat' }, // Removed parent: 'boil'
  { id: 'place-pour', from: 'place', to: 'pour' },// Removed parent: 'steep'

  // Dependency/Sequence Edges (across original hierarchy levels)
  { id: 'heat-pour', from: 'heat', to: 'pour' } // Cross-subtask dependency
];

const App: React.FC = () => {
  // Use state setters for nodes and edges with initial data
  const [nodes, setNodes] = useState<MyNodeData[]>(initialNodes);
  const [edges, setEdges] = useState(initialEdges);

  // Track selected node
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  // Counter for generating unique IDs
  const nodeIdCounter = useRef(nodes.length + 1);

  // Function to add a new node
  const handleAddNode = (parentNode: MyNodeData) => {
    const newNodeId = `node-${nodeIdCounter.current++}`;
    const newEdgeId = `${parentNode.id}-${newNodeId}`;

    // Create a new leaf node with default data structure
    const newNode: MyNodeData = {
      id: newNodeId,
      width: NODE_WIDTH, // Use standard dimensions
      height: NODE_HEIGHT,
      data: {
        type: 'action', // Default type
        title: 'New Task',
        description: 'Specify details',
        icon: '🆕'
      }
    };

    // Create the edge connecting to it
    const newEdge = {
      id: newEdgeId,
      from: parentNode.id,
      to: newNodeId,
      className: 'edge-hierarchy' 
    };

    // Update state
    setNodes(prevNodes => [...prevNodes, newNode]);
    setEdges(prevEdges => [...prevEdges, newEdge]);

    console.log(`Added node ${newNodeId} branching from ${parentNode.id}`);
  };

  // Function to handle node selection (extracted for clarity)
  const handleNodeClick = (id: string) => {
    setSelectedNode(id);
  };

  return (
    <div className="app-container">
      <h1>Hierarchical Task Planning Network with Reaflow</h1>
      <p>A visualization of an AI planning scenario: Making a Cup of Tea</p>
      
      <div className="canvas-container">
        <Canvas
          nodes={nodes}
          edges={edges}
          fit={true}
          // maxWidth={1200} // Increase size maybe
          // maxHeight={800}
          direction="DOWN"
          layoutOptions={layoutOptions}
          zoomable={true} // Enable zoom
          pannable={true} // Enable pan
          node={
            <Node
              // Removed the style prop here, as styling is handled inside CustomNode
              // Ensure the node size is passed if not using layout defaults
            >
              {(nodeProps: NodeChildProps) => (
                // Use the CustomNode component
                <CustomNode
                  nodeProps={nodeProps}
                  selectedNode={selectedNode}
                  // nodes={nodes}
                  onNodeClick={handleNodeClick}
                  onAddClick={handleAddNode}
                />
              )}
            </Node>
          }
          edge={
            <Edge
              style={{
                stroke: '#78909c', // Keep edge style simple for now
                strokeWidth: 1.5,
              }}
            />
          }
        />
      </div>

      {/* Info Panel - Update to show new data fields */}
      <div className="info-panel">
        <h3>Selected Node</h3>
        {selectedNode ? (
          (() => {
            const node = nodes.find(n => n.id === selectedNode);
            const nodeData = node?.data;
            return node ? (
              <div>
                <p><strong>ID:</strong> {node.id}</p>
                <p><strong>Type:</strong> {nodeData?.type || 'N/A'}</p>
                <p><strong>Title:</strong> {nodeData?.title || node.text || 'N/A'}</p>
                <p><strong>Description:</strong> {nodeData?.description || 'N/A'}</p>
                {nodeData?.stats && (
                  <p><strong>Stats:</strong> {JSON.stringify(nodeData.stats)}</p>
                )}
              </div>
            ) : <p>Node not found</p>;
          })()
        ) : (
          <p>Click on a node to see its details</p>
        )}
      </div>

      {/* Legend - Update to reflect new node types/styles */}
      <div className="legend">
        <h3>Legend</h3>
        {/* Example legend items - map nodeConfig or define manually */}
        <div className="legend-item">
           <div className="legend-color" style={{ backgroundColor: nodeConfig('goal').backgroundColor, borderLeft: `4px solid ${nodeConfig('goal').color}` }}></div>
          <span>Goal</span>
        </div>
         <div className="legend-item">
           <div className="legend-color" style={{ backgroundColor: nodeConfig('subtask').backgroundColor, borderLeft: `4px solid ${nodeConfig('subtask').color}` }}></div>
          <span>Subtask</span>
        </div>
        <div className="legend-item">
           <div className="legend-color" style={{ backgroundColor: nodeConfig('action').backgroundColor, borderLeft: `4px solid ${nodeConfig('action').color}` }}></div>
          <span>Action</span>
        </div>
         {/* Add more based on nodeConfig */}
      </div>

      {/* Add specific styles for hierarchy edges */}
      <Edge
        id="edge-hierarchy-style" // Placeholder ID for potential global style reference
        style={{
          stroke: '#adb5bd', // Lighter grey for hierarchy
          strokeDasharray: '5 2', // Dashed line
          strokeWidth: 1,
        }}
        // markerEnd='url(#arrow-hierarchy)'
      />

      <defs>
        <marker
          id="arrow-hierarchy"
          viewBox="0 0 10 10"
          refX="8"
          refY="5"
          markerWidth="6"
          markerHeight="6"
          orient="auto-start-reverse"
        >
          <path d="M 0 0 L 10 5 L 0 10 z" fill="#adb5bd" />
        </marker>
      </defs>
    </div>
  );
};

export default App;
