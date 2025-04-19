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
  const enterStart = i * 60; // every 2 s

  // slide from 60 px left
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