import React from 'react';
import { NodeChildProps } from 'reaflow';
import { MyNodeData } from './App'; // Assuming MyNodeData is exported from App.tsx or moved elsewhere

// Define constants or import them
const NODE_WIDTH = 260;
const NODE_HEIGHT = 100;

interface CustomNodeProps {
  nodeProps: NodeChildProps<MyNodeData>; // Pass the nodeProps from Reaflow
  selectedNode: string | null;
  nodes: MyNodeData[]; // Needed to check if a node is a parent
  onNodeClick: (id: string) => void; // Callback for selection
  onAddClick: (node: MyNodeData) => void; // Callback for adding node
}

export const CustomNode: React.FC<CustomNodeProps> = ({ 
  nodeProps,
  selectedNode,
  nodes,
  onNodeClick,
  onAddClick
}) => {
  // --- Start of moved logic from App.tsx --- 
  const nodeData = nodeProps.node;
  // Size calculations remain the same
  const nodeWidth = nodeData.width || NODE_WIDTH;
  const nodeHeight = nodeData.height || NODE_HEIGHT;
  const isSelected = selectedNode === nodeData.id;
  // Check if it has children or is root
  const isParent = nodeData.parent === undefined || nodes.some(n => n.parent === nodeData.id); 

  // Style calculations remain the same
  const fill = nodeData.parent === undefined ? '#e3f2fd' :
               nodeData.parent === 'goal' ? '#bbdefb' :
               '#fff8e1';
  const stroke = isSelected ? '#f44336' : '#2196f3';
  const strokeWidth = isSelected ? 2 : 1;
  const textY = isParent ? 20 : nodeHeight / 2;
  const textAnchor = isParent ? "start" : "middle";
  const textX = isParent ? 15 : nodeWidth / 2;

  // Determine if add button should be shown (e.g., not on the root)
  const showAddButton = nodeData.id !== 'goal'; 

  return (
    <g 
      onClick={(e) => {
        e.stopPropagation();
        onNodeClick(nodeData.id); // Use callback
        console.log(`Clicked on node: ${nodeData.text} (ID: ${nodeData.id})`);
      }}
      style={{ cursor: 'pointer' }}
    >
      {/* Node rectangle */}
      <rect 
        width={nodeWidth}
        height={nodeHeight}
        fill={fill}
        stroke={stroke}
        strokeWidth={strokeWidth}
        rx={8} 
        ry={8}
      />
      {/* Node text */}
      <text 
        x={textX} 
        y={textY} 
        textAnchor={textAnchor}
        dominantBaseline="middle"
        fill="#333"
        fontSize="14px"
        fontWeight={isParent ? 'bold' : 'normal'}
      >
        {nodeData.text}
      </text>
      
      {/* Add Node Button (+) - positioned bottom-right */}
      {showAddButton && (
        <g 
          transform={`translate(${nodeWidth - 25}, ${nodeHeight - 25})`} 
          onClick={(e) => {
            e.stopPropagation(); 
            onAddClick(nodeData); // Use callback
          }}
          style={{ cursor: 'pointer' }}
        >
          <circle cx="0" cy="0" r="10" fill="#4caf50" stroke="#fff" strokeWidth="1" />
          <text 
            x="0"
            y="0"
            textAnchor="middle"
            dominantBaseline="central" 
            fill="white"
            fontSize="14px"
            fontWeight="bold"
            style={{ pointerEvents: 'none' }} 
          >
            +
          </text>
          <title>Add child node</title>
        </g>
      )}

      {/* Existing Tooltip */}
      <title>{`${nodeData.text} (ID: ${nodeData.id})`}</title>
    </g>
  );
  // --- End of moved logic --- 
}; 