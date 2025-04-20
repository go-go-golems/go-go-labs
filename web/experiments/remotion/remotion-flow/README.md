# Content Creation Workflow Animation

A 1080p, 30fps Remotion animation that illustrates a typical programming content creation workflow. Each step in the workflow animates sequentially with stylish transitions, effects, and a flowing design.

## Workflow Steps

```
idea ─▶ o3 ─▶ refine ─▶ tutorial.md ─▶ cursor ─▶ implement ─▶ refine ─▶ update tutorial ─▶ final product
```

## Features

- ✨ Fluid animations with spring physics
- 🎨 Colorful gradient boxes for each workflow step
- ➡️ Animated arrows connecting each step
- 🌌 Starry background with parallax effect
- ✨ Particle system for ambient motion
- 📜 Auto-scrolling container to keep active items in view
- 💫 Special pulsing effect for the final product box

## Getting Started

### Prerequisites

- Node.js ≥ 18 LTS
- npm, yarn, or pnpm

### Installation

```bash
# Clone the repository
# Navigate to the project directory
cd web/experiments/remotion/remotion-flow

# Install dependencies
npm install
```

### Development

```bash
# Start the development server
npm run dev
```

This will open the Remotion Studio at [http://localhost:3000](http://localhost:3000) where you can preview and scrub through the animation.

### Rendering

The project includes several rendering options:

```bash
# Standard quality render
npm run render

# High quality render
npm run render-hq

# ProRes render with alpha channel (for compositing)
npm run render-prores
```

Rendered videos will be saved to the `out/` directory.

## Customization

You can customize various aspects of the animation by modifying the following in `FlowScene.tsx`:

- `STEPS` array: Change the workflow steps
- `COLORS` array: Modify the gradient colors
- Animation timings and durations
- Box sizes, spacing, and positioning
- Background effects and particle systems

## Implementation Details

The animation is built with Remotion and React, using:

- Interpolation for smooth transitions
- Spring physics for natural motion
- Dynamic scrolling to follow the active step
- CSS transitions and transforms for visual effects
- React hooks for animation state management

## References

Based on the tutorial document: `@02-remotion-animation.md`
