import React from 'react';
import { NodeChildProps } from 'reaflow';
import { MyNodeData } from './App';
import { nodeConfig } from './nodeConfig';

// Define node dimensions directly here or import if needed elsewhere
const NODE_WIDTH = 260;
const NODE_HEIGHT = 164; // Matches reachat-codesandbox including button space
const NODE_BUTTON_SIZE = 32; // From reachat style

// --- Helper Components (can remain as they are type/data driven) ---

interface NodeStatsProps {
  showStats?: boolean;
  stats?: Record<string, number>;
}

const NodeStats: React.FC<NodeStatsProps> = ({ showStats, stats }) =>
  showStats && stats ? (
    <ul className="node-stats">
      {Object.entries(stats).map(([label, count]) => (
        <li key={label}>
          <span>{label}</span>
          <strong>{count}</strong>
        </li>
      ))}
    </ul>
  ) : null;

interface NodeContentProps {
  node: MyNodeData;
  selected?: boolean;
  onClick?: () => void;
}

const NodeContent: React.FC<NodeContentProps> = ({ node, selected, onClick }) => {
  const nodeData = node.data || { type: 'default', title: 'Default Title' }; // Fallback
  const type = nodeData.type || 'default';
  const title = nodeData.title || 'Node Title';
  const description = nodeData.description || 'Node Description';
  const showError = nodeData.showError;

  const { color, icon, backgroundColor } = nodeConfig(type);

  // Use inline styles derived from nodeConfig for simplicity
  const nodeContentStyle: React.CSSProperties = {
    borderLeft: `4px solid ${color}`,
    backgroundColor: backgroundColor,
    color: '#333' // Default text color
  };

  const nodeIconStyle: React.CSSProperties = {
    color: color // Use the node's primary color for the icon
  };

  return (
    <div
      className="node-content"
      style={nodeContentStyle}
      onClick={onClick}
      aria-selected={selected}
    >
      {showError && <div className="node-error-badge"></div>}
      <div className="node-icon" style={nodeIconStyle}>{icon}</div>
      <div className="node-details">
        <h1>{title} (ID: {node.id})</h1>
        <p>{description}</p>
      </div>
      <NodeStats stats={nodeData.stats} showStats={nodeData.showStats} />
    </div>
  );
};

// --- Main CustomNode Component ---

export interface CustomNodeProps {
  nodeProps: NodeChildProps; // Use generic type
  selectedNode: string | null;
  // nodes: MyNodeData[]; // No longer needed for parent check
  onNodeClick: (id: string) => void;
  onAddClick: (node: MyNodeData) => void;
}

export const CustomNode: React.FC<CustomNodeProps> = ({
  nodeProps,
  selectedNode,
  onNodeClick,
  onAddClick
}) => {
  const { node, x, y } = nodeProps;
  const width = node.width || NODE_WIDTH;
  const height = node.height || NODE_HEIGHT;
  const isSelected = selectedNode === node.id;
  const isDisabled = false; // Add logic if needed

  // Determine if add button should be shown (e.g., hide for 'end' type)
  const showAddButton = node.data?.type !== 'end';

  return (
    <foreignObject x={0} y={0}
     width={width} height={height}
     >
      {/* Apply CSS reset/base styles via className */}
      <div className="node-style-reset node-wrapper">
        <NodeContent
          node={node}
          selected={isSelected}
          onClick={onNodeClick ? () => onNodeClick(node.id) : undefined}
        />

        {/* Add Button - similar structure to reachat */}
        {showAddButton && (
          <div className="add-button">
            {/* Basic button for now, can enhance with icons later */}
            <button
              disabled={isDisabled}
              // size="middle"
              // shape="circle"
              // icon={<PlusOutlined />}
              onClick={(e) => {
                e.stopPropagation(); // Prevent node click
                if (onAddClick) onAddClick(node);
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