import { Node as ReaflowNode, NodeProps, NodeChildProps, NodeData as ReaflowNodeData } from 'reaflow';
import { NodeData } from '../types/htn';

// Define the props our custom node wrapper will receive
// We expect Reaflow to pass the node data in the 'properties' prop
// Use Reaflow's internal NodeData<any> for the event data type
interface CustomNodeWrapperProps extends Partial<NodeProps> {
  properties: NodeData;
  onClick?: (event: React.MouseEvent<SVGGElement>, data: ReaflowNodeData<NodeData>) => void;
}

// Define props for the inner content renderer
// This receives the structured data from NodeData
interface NodeContentProps {
  nodeData: NodeData;
}

// Inner component to render the node's content
const NodeContent: React.FC<NodeContentProps> = ({ nodeData }) => {
  const { id, text, data } = nodeData;
  const hasChildren = data?.hasChildren ?? false;

  return (
    <div style={{
      padding: '10px 15px',
      borderRadius: '4px',
      background: hasChildren ? "#D5E8D4" : "#FFF2CC",
      border: '1px solid #888',
      textAlign: 'center',
      color: '#333',
      height: '100%', // Fill the foreignObject
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'center'
    }}>
      {/* Handle optional text gracefully */}
      <div style={{ fontWeight: 'bold', marginBottom: '5px' }}>{text ?? ''}</div>
      <div style={{ fontSize: '0.8em', color: '#666' }}>ID: {id}</div>
    </div>
  );
};

// The main custom node component passed to Reaflow
export function CustomNode({ properties, onClick, ...restProps }: CustomNodeWrapperProps) {
  return (
    <ReaflowNode
      {...restProps} // Pass down other props like x, y, etc.
      onClick={onClick} // Pass the onClick handler to the underlying ReaflowNode
      rx={5} // Slightly rounded corners for the node container
      ry={5}
    >
      {(event: NodeChildProps) => (
        // Use foreignObject to render complex HTML/React content inside SVG
        <foreignObject x={0} y={0} width={event.width} height={event.height}>
          {/* Pass the actual node data to the content renderer */}
          <NodeContent nodeData={properties} />
        </foreignObject>
      )}
    </ReaflowNode>
  );
} 