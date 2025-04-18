import { Canvas, NodeData as ReaflowNodeData } from 'reaflow';
import { NodeData } from '../types/htn';
import { CustomNode } from './CustomNode';
import { CustomEdge } from './CustomEdge';

interface HTNGraphProps {
  nodes: NodeData[];
  edges: any[];
  onNodeClick: (node: NodeData, event: React.MouseEvent<Element, MouseEvent>) => void;
}

export function HTNGraph({ nodes, edges, onNodeClick }: HTNGraphProps) {
  // Reaflow expects the node data under the `properties` key when passing a component instance
  const nodesWithProperties = nodes.map(node => ({ ...node, properties: node }));

  // Ensure the handler passed to CustomNode matches its expected signature
  const handleNodeClick = (event: React.MouseEvent<SVGGElement>, nodeProps: ReaflowNodeData<NodeData>) => {
    // Extract our original NodeData structure from the 'properties' field
    const originalNodeData = nodeProps.data as NodeData;
    if (originalNodeData) {
      onNodeClick(originalNodeData, event);
    }
  };

  return (
    <Canvas
      nodes={nodesWithProperties} // Pass nodes with data nested under 'properties'
      edges={edges}
      // node={<CustomNode onClick={handleNodeClick} properties={nodesWithProperties} />} // Pass instance of CustomNode with correctly typed handler
      // edge={<CustomEdge />} // Pass instance of CustomEdge - Reaflow injects props
      fit={true}
      direction="DOWN"
      zoomable={true}
      pannable={true}
      readonly // Make canvas readonly if not implementing node dragging/selection features
    />
  );
} 