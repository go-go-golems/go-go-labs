# Hierarchical Task Planning Networks & Reaflow

## What are Hierarchical Task Networks (HTNs)?

Hierarchical Task Networks (HTNs) are a planning methodology used in artificial intelligence to represent and solve complex planning problems. The key features of HTNs include:

1. **Hierarchical Decomposition**: Complex tasks are broken down into simpler subtasks, forming a hierarchy.
2. **Task Ordering**: Specifies the order in which subtasks must be executed.
3. **Preconditions and Effects**: Define when tasks can be executed and what changes they make to the world state.

HTNs are widely used in:

- Game AI systems
- Robotics
- Automated planning and scheduling
- Natural language processing (for task-based dialogues)

## Why Reaflow for Visualizing HTNs?

[Reaflow](https://reaflow.dev/) is an ideal library for representing HTNs visually for several reasons:

1. **Native Hierarchy Support**: Reaflow's `parent` property makes it easy to represent task/subtask relationships.
2. **Automatic Layout**: The ELK graph layout engine positions nodes optimally without manual coordination.
3. **Interactive Capabilities**: Click, drag, selection, and other interactions help explore complex plans.
4. **Customizable Appearance**: Nodes and edges can be styled to represent different task types or states.
5. **Edge Routing**: Cross-hierarchy connections allow representation of dependencies between branches.

## Implementation Patterns

### 1. Representing Task Hierarchy

In our implementation, we represent hierarchy using the `parent` property:

```tsx
const nodes = [
  { id: "goal", text: "Make Tea" },
  { id: "boil", text: "Boil Water", parent: "goal" },
  { id: "fill", text: "Fill Kettle", parent: "boil" },
];
```

### 2. Task Dependencies

Dependencies are represented as edges between nodes:

```tsx
const edges = [{ id: "fill-heat", from: "fill", to: "heat" }];
```

### 3. Cross-Hierarchy Dependencies

Reaflow handles edges between nodes in different parts of the hierarchy:

```tsx
// Edge from a node in one branch to a node in another branch
{ id: 'heat-pour', from: 'heat', to: 'pour' }
```

### 4. Visual Differentiation

Different node levels are styled differently:

```tsx
style={(n) => ({
  fill: n.parent === undefined ? '#e3f2fd' :  // Top level
        n.parent === 'goal' ? '#bbdefb' :  // Second level
        '#fff8e1',  // Leaf nodes (actions)
})}
```

## Extending the Implementation

The current implementation could be extended to represent:

1. **Task States**: Using different colors to indicate completed, in-progress, or blocked tasks
2. **Preconditions**: Adding visual indicators for task requirements
3. **Plan Execution**: Animating the current state of execution through the task network
4. **Interactive Planning**: Allowing users to modify the plan by adding/removing tasks
5. **Multiple Abstraction Levels**: Showing/hiding levels of detail based on zoom level

## References

- [Reaflow Documentation](https://reaflow.dev/)
- Erol, K., Hendler, J., & Nau, D. S. (1994). HTN planning: Complexity and expressivity. AAAI, 94, 1123-1128.
- Ghallab, M., Nau, D., & Traverso, P. (2004). Automated planning: theory and practice. Elsevier.
