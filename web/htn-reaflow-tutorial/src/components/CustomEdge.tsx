import { Edge as ReaflowEdge, EdgeProps } from 'reaflow';

// Minimal custom edge - we just want to pass it as a component instance
// And potentially customize style or behavior later
export function CustomEdge(props: Partial<EdgeProps>) {
  return (
    <ReaflowEdge
      {...props}
      style={{ stroke: "#555" }} // Style defined here
    />
  );
} 