import React, { useState, useRef } from 'react';
import { Canvas, Node, Edge, NodeChildProps, ElkCanvasLayoutOptions } from 'reaflow';
import './App.css';
import { CustomNode } from './CustomNode'; // Import the new component

// Define an interface for our custom node data (Ensure it's exported)
export interface MyNodeData {
  id: string;
  text: string;
  parent?: string;
  width?: number;
  height?: number;
}

const NODE_WIDTH = 260;
const NODE_HEIGHT = 100;

// Define Layout Options
const layoutOptions: ElkCanvasLayoutOptions = {
  'elk.algorithm': 'layered',
  'elk.direction': 'DOWN',
  'elk.spacing.nodeNode': '80',
  'elk.portAlignment.default': 'CENTER',
  'elk.layered.spacing.nodeNodeBetweenLayers': '80',
};

const App: React.FC = () => {
  // Use state setters for nodes and edges
  const [nodes, setNodes] = useState<MyNodeData[]>([
    { id: 'goal', text: 'Make Tea' },
    { id: 'boil', text: 'Boil Water', parent: 'goal' },
    { id: 'fill', text: 'Fill Kettle', parent: 'boil', width: NODE_WIDTH, height: NODE_HEIGHT },
    { id: 'heat', text: 'Heat Kettle', parent: 'boil', width: NODE_WIDTH, height: NODE_HEIGHT },
    { id: 'steep', text: 'Steep Tea', parent: 'goal' },
    { id: 'place', text: 'Place Teabag', parent: 'steep', width: NODE_WIDTH, height: NODE_HEIGHT },
    { id: 'pour', text: 'Pour Water', parent: 'steep', width: NODE_WIDTH, height: NODE_HEIGHT }
  ]);
  const [edges, setEdges] = useState([
    { id: 'fill-heat', from: 'fill', to: 'heat', parent: 'boil' },
    { id: 'place-pour', from: 'place', to: 'pour', parent: 'steep' },
    { id: 'heat-pour', from: 'heat', to: 'pour' }
  ]);

  // Track selected node
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

  // Counter for generating unique IDs
  const nodeIdCounter = useRef(nodes.length + 1);

  // Function to add a new node
  const handleAddNode = (parentNode: MyNodeData) => {
    const newNodeId = `node-${nodeIdCounter.current++}`;
    const newEdgeId = `${parentNode.id}-${newNodeId}`;

    // Create a new leaf node
    const newNode: MyNodeData = {
      id: newNodeId,
      text: 'New Task',
      parent: parentNode.id,
      width: NODE_WIDTH, 
      height: NODE_HEIGHT
    };

    // Create the edge connecting to it
    const newEdge = {
      id: newEdgeId,
      from: parentNode.id,
      to: newNodeId,
      parent: newNode.parent
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
          maxWidth={1000}
          maxHeight={750}
          direction="DOWN"
          layoutOptions={layoutOptions}
          node={
            <Node
              // Removed the style prop here, as styling is handled inside CustomNode based on selection/level
            >
              {(nodeProps: NodeChildProps) => (
                // Use the CustomNode component
                <CustomNode
                  nodeProps={nodeProps}
                  selectedNode={selectedNode}
                  nodes={nodes} // Pass the full nodes array
                  onNodeClick={handleNodeClick} // Pass the click handler
                  onAddClick={handleAddNode}    // Pass the add handler
                />
              )}
            </Node>
          }
          edge={
            <Edge
              style={{
                stroke: '#78909c',
                strokeWidth: 1.5,
              }}
            />
          }
        />
      </div>

      <div className="info-panel">
        <h3>Selected Node</h3>
        {selectedNode ? (
          <div>
            <p>
              <strong>ID:</strong> {selectedNode}
            </p>
            <p>
              <strong>Text:</strong> {nodes.find(n => n.id === selectedNode)?.text || ''}
            </p>
            <p>
              <strong>Parent:</strong> {nodes.find(n => n.id === selectedNode)?.parent || 'None'}
            </p>
          </div>
        ) : (
          <p>Click on a node to see its details</p>
        )}
      </div>

      <div className="legend">
        <h3>Legend</h3>
        <div className="legend-item">
          <div className="legend-color node-level-root"></div>
          <span>Goal (Top Level)</span>
        </div>
        <div className="legend-item">
          <div className="legend-color node-level-subtask"></div>
          <span>Subtasks (Second Level)</span>
        </div>
        <div className="legend-item">
          <div className="legend-color node-level-leaf"></div>
          <span>Actions (Leaf Nodes)</span>
        </div>
      </div>
    </div>
  );
};

export default App;
