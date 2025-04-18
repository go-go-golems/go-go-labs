import React, { useState } from 'react';
import { Canvas, Node, Edge, NodeChildProps } from 'reaflow';
import './App.css';

// Define an interface for our custom node data
interface MyNodeData {
  id: string;
  text: string;
  parent?: string;
  width?: number;
  height?: number;
}

const NODE_WIDTH = 260;
const NODE_HEIGHT = 100;

const App: React.FC = () => {
  // Let ELK determine parent sizes; define leaf node sizes
  const [nodes] = useState<MyNodeData[]>([
    { id: 'goal', text: 'Make Tea' }, // No explicit size
    { id: 'boil', text: 'Boil Water', parent: 'goal' }, // No explicit size
    { id: 'fill', text: 'Fill Kettle', parent: 'boil', width: NODE_WIDTH, height: NODE_HEIGHT },
    { id: 'heat', text: 'Heat Kettle', parent: 'boil', width: NODE_WIDTH, height: NODE_HEIGHT },
    { id: 'steep', text: 'Steep Tea', parent: 'goal' }, // No explicit size
    { id: 'place', text: 'Place Teabag', parent: 'steep', width: NODE_WIDTH, height: NODE_HEIGHT },
    { id: 'pour', text: 'Pour Water', parent: 'steep', width: NODE_WIDTH, height: NODE_HEIGHT }
  ]);

  // Define edges between the nodes
  const [edges] = useState([
    // sequence within Boil Water:
    { id: 'fill-heat', from: 'fill', to: 'heat' },
    // sequence within Steep Tea:
    { id: 'place-pour', from: 'place', to: 'pour' },
    // dependency between Boil Water and Steep Tea:
    { id: 'heat-pour', from: 'heat', to: 'pour' }
  ]);

  // Track the selected node
  const [selectedNode, setSelectedNode] = useState<string | null>(null);

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
          node={
            <Node
              style={(n: MyNodeData) => ({
                // Style nodes based on their level in the hierarchy
                fill: n.parent === undefined ? '#e3f2fd' :  // Top level
                      n.parent === 'goal' ? '#bbdefb' :  // Second level
                      '#fff8e1',  // Leaf nodes (actions)
                stroke: selectedNode === n.id ? '#f44336' : '#2196f3',
                strokeWidth: selectedNode === n.id ? 2 : 1,
                rx: 8,
                ry: 8
              })}
            >
              {(nodeProps: NodeChildProps) => { // Use NodeChildProps without generics for now
                const nodeData = nodeProps.node as MyNodeData; // Cast to our type
                // Use nodeData size if available (leaf), otherwise default (parent - will be overridden by ELK)
                const nodeWidth = nodeData.width || NODE_WIDTH;
                const nodeHeight = nodeData.height || NODE_HEIGHT;
                const isSelected = selectedNode === nodeData.id;
                
                // Determine node level for styling
                let nodeLevelClass = 'node-level-leaf';
                if (nodeData.parent === undefined) {
                  nodeLevelClass = 'node-level-root';
                } else if (nodeData.parent === 'goal') {
                  nodeLevelClass = 'node-level-subtask';
                }

                return (
                  <foreignObject x={0} y={0} width={nodeWidth} height={nodeHeight}>
                    <div 
                      className={`node-wrapper ${nodeLevelClass}`}
                      aria-selected={isSelected}
                      onClick={(e) => {
                        // Prevent click from propagating if it's inside the foreignObject
                        e.stopPropagation();
                        setSelectedNode(nodeData.id);
                        console.log(`Clicked on node: ${nodeData.text} (ID: ${nodeData.id})`);
                      }}
                    >
                      <div className="node-content">
                        <div className="node-details">
                          <h1>{nodeData.text}</h1>
                          <p>ID: {nodeData.id}{nodeData.parent ? ` (Parent: ${nodeData.parent})` : ''}</p>
                        </div>
                        {/* Icon placeholder can go here */}
                      </div>
                      {/* Tooltip using SVG title */}
                      <title>{`${nodeData.text} (ID: ${nodeData.id})`}</title>
                    </div>
                  </foreignObject>
                );
              }}
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
