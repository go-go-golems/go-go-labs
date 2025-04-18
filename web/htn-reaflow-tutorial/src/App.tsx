import { useState, useEffect, useCallback } from 'react';
import { HTNGraph } from './components/HTNGraph';
import { generateGraphFromHTNVisible, examples } from './utils/htnUtils';
import { Task, NodeData } from './types/htn';
import './App.css';

function App() {
  const [htnData, setHtnData] = useState<Task>(examples.houseProject);
  const [expandedNodes, setExpandedNodes] = useState<Set<string>>(new Set());
  const [graphData, setGraphData] = useState<{ nodes: NodeData[]; edges: any[] }>({ nodes: [], edges: [] });

  useEffect(() => {
    const { nodes, edges } = generateGraphFromHTNVisible(htnData, expandedNodes);
    setGraphData({ nodes, edges });
  }, [htnData, expandedNodes]);

  const handleNodeClick = useCallback((node: NodeData, event: React.MouseEvent<Element, MouseEvent>) => {
    console.log('Node clicked:', node, event);
    if (node.data?.hasChildren) {
      setExpandedNodes(prev => {
        const newSet = new Set(prev);
        if (prev.has(node.id)) {
          newSet.delete(node.id);
        } else {
          newSet.add(node.id);
        }
        return newSet;
      });
    }
  }, []);

  const handleExampleChange = (example: keyof typeof examples) => {
    setHtnData(examples[example]);
    setExpandedNodes(new Set());
  };

  return (
    <div className="app-container">
      <div className="controls">
        <h1>HTN Visualizer</h1>
        <div className="example-buttons">
          <button onClick={() => handleExampleChange('houseProject')}>House Project</button>
          <button onClick={() => handleExampleChange('softwareProject')}>Software Project</button>
        </div>
      </div>
      <div className="graph-container" style={{ height: 'calc(100vh - 120px)' }}>
        <HTNGraph
          nodes={graphData.nodes}
          edges={graphData.edges}
          onNodeClick={handleNodeClick}
        />
      </div>
    </div>
  );
}

export default App; 