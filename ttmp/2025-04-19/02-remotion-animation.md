# Remotion Tutorial – Animating the Workflow `idea → o3 → refine → tutorial.md → cursor → implement → refine → update tutorial → final product`

---

## 1  What You’ll Build

A 1080‑p, 30 fps video that draws nine nodes (with arrows) in sequence to illustrate a typical programming content‑creation loop. Each node slides in, fades, and gently scales for emphasis. The resulting **MP4** is ready to drop into your YouTube timeline.

```
idea ─▶ o3 ─▶ refine ─▶ tutorial.md ─▶ cursor ─▶ implement ─▶ refine ─▶ update tutorial ─▶ final product
```

---

## 2  Prerequisites

| Tool              | Version tested                            |
| ----------------- | ----------------------------------------- |
| Node.js           | ≥ 18 LTS                                  |
| pnpm / npm / yarn | any                                       |
| FFmpeg            | optional (Remotion can render without it) |

```bash
# check versions
node -v
npm -v  # or pnpm -v / yarn -v
```

---

## 3  Project Scaffold

```bash
npx create-video remotion-flow
cd remotion-flow
npm install   # or pnpm install / yarn
```

This creates a TypeScript‑ready Remotion workspace:

```
remotion-flow/
 ├─ package.json
 ├─ remotion.config.ts
 └─ src/
     ├─ Video.tsx      # default composition registry
     ├─ HelloWorld.tsx # starter scene – we’ll delete this
     └─ index.tsx
```

Delete `HelloWorld.tsx` and any references in `Video.tsx` so we start clean.

---

## 4  Design Parameters

- **FPS:** `30`
- **Resolution:** `1920 × 1080`
- **Total duration:** `9 steps × 60 frames (2 s) = 540 frames` (18 s)
- **Easing:** spring‑based slide‑in (`spring` util)

---

## 5  Create `FlowScene.tsx`

```tsx
import { interpolate, useCurrentFrame, spring, useVideoConfig } from "remotion";
import { AbsoluteFill, Sequence } from "remotion";
import React from "react";

const STEPS = [
  "idea",
  "o3",
  "refine",
  "tutorial.md",
  "cursor",
  "implement",
  "refine",
  "update tutorial",
  "final product",
] as const;

const BOX_W = 260;
const BOX_H = 90;
const SPACING = 120;

const Box: React.FC<{ label: string; i: number }> = ({ label, i }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enterStart = i * 60; // every 2 s

  // slide from 60 px left
  const translateX = interpolate(
    frame,
    [enterStart, enterStart + 30],
    [-60, 0],
    { extrapolateRight: "clamp" }
  );
  const scale = spring({
    frame: frame - enterStart,
    fps,
    config: { damping: 200, mass: 0.9 },
  });

  return (
    <div
      style={{
        position: "absolute",
        left: i * (BOX_W + SPACING),
        top: 420,
        width: BOX_W,
        height: BOX_H,
        borderRadius: 12,
        background: "#1f2d3d",
        display: "flex",
        justifyContent: "center",
        alignItems: "center",
        color: "white",
        fontSize: 28,
        transform: `translateX(${translateX}px) scale(${scale})`,
      }}
    >
      {label}
    </div>
  );
};

export const FlowScene: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: "#0d1117" }}>
    {STEPS.map((label, i) => (
      <Sequence key={label} from={i * 60} durationInFrames={540 - i * 60}>
        <Box label={label} i={i} />
      </Sequence>
    ))}
  </AbsoluteFill>
);
```

### Notes

- **`Sequence`** gates the component so each box mounts exactly when its portion of the timeline starts.
- **Arrows:** For a simple first version, we let the viewer infer connections; add `<svg>` arrows later if desired.

---

## 6  Register the Composition

Edit `src/Video.tsx`:

```tsx
import { Composition } from "remotion";
import { FlowScene } from "./FlowScene";

export const Video: React.FC = () => (
  <>
    <Composition
      id="flow"
      component={FlowScene}
      durationInFrames={540}
      fps={30}
      width={1920}
      height={1080}
    />
  </>
);
```

---

## 7  Interactive Preview

```bash
npm run start     # opens localhost:3000 with Remotion Studio
```

Scrub the timeline to check timings; tweak `enterStart` or add arrows.

---

## 8  Render to MP4

```bash
npm run build
# or explicitly:
remotion render flow out/flow.mp4 \
  --codec h264 \
  --quality 90
```

Expect ~10–20 s render on a modern machine.

---

## 9  Optional Polish

- **SVG Arrows:** Add an `<svg>` overlay inside each `Sequence` whose stroke length animates with `interpolate` for a draw‑on effect.
- **Brand Fonts:** Wrap content in a `loadFont()` call inside `remotion.config.ts`.
- **Color Scheme:** Use Tailwind or theme tokens if you already have a design‑system file.
- **Texture:** `<LinearGradient>` from `remotion-gradient` for subtle depth on boxes.

---

## 10  Drop into Your Editor

The resulting `out/flow.mp4` can be dragged directly into DaVinci Resolve, Premiere, or OBS scenes. Because the background is solid, chroma‑keying isn’t needed. If you want an **alpha channel**, re‑render with `--codec=prores --proresProfile=4444` and set the background to `transparent`.

---

## 11  Recap

1. Scaffold with `create-video`.
2. Encode each workflow step in an array.
3. Map→Sequence→Box for modular timing.
4. Render once; reuse anywhere.

You now have a reusable Remotion scene you can version‑control alongside your tutorial source code. Happy shipping!
