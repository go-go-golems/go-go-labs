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
const ARROW_LENGTH = 80;
const TOTAL_WIDTH = STEPS.length * (BOX_W + SPACING) - SPACING;

// Colors with gradients for more visual appeal
const COLORS = [
  ["#FF512F", "#F09819"], // Orangish
  ["#1A2980", "#26D0CE"], // Blue to Cyan
  ["#8E2DE2", "#4A00E0"], // Purple
  ["#00c6ff", "#0072ff"], // Lighter Blue
  ["#FF416C", "#FF4B2B"], // Pinkish Red
  ["#4776E6", "#8E54E9"], // Blue to Purple
  ["#16A085", "#F4D03F"], // Green to Yellow
  ["#667eea", "#764ba2"], // Periwinkle
  ["#11998e", "#38ef7d"], // Green
];

const getBoxGradient = (i: number) => {
  const colors = COLORS[i % COLORS.length];
  return `linear-gradient(135deg, ${colors[0]}, ${colors[1]})`;
};

const Arrow: React.FC<{ fromX: number; i: number; scrollPosition: number }> = ({ fromX, i, scrollPosition }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enterStart = i * 60 + 30; // appear slightly after the box

  // Draw line animation
  const lineProgress = interpolate(
    frame,
    [enterStart, enterStart + 25],
    [0, 1],
    { extrapolateRight: "clamp" }
  );

  // Fade in
  const opacity = interpolate(
    frame,
    [enterStart, enterStart + 20],
    [0, 1],
    { extrapolateRight: "clamp" }
  );

  // Slight bounce with spring
  const arrowHeadScale = spring({
    frame: frame - enterStart - 15,
    fps,
    config: { damping: 12, mass: 0.3 },
  });

  const arrowLength = ARROW_LENGTH * lineProgress;
  
  // Adjust position for scrolling
  const adjustedX = fromX - scrollPosition;

  return (
    <div
      style={{
        position: "absolute",
        left: adjustedX + BOX_W + 10,
        top: 420 + BOX_H / 2 - 1,
        width: arrowLength,
        height: 2,
        opacity,
        background: "white",
        display: "flex",
        alignItems: "center",
        justifyContent: "flex-end",
      }}
    >
      {/* Arrow head - triangle */}
      <div
        style={{
          width: 0,
          height: 0,
          borderTop: "6px solid transparent",
          borderBottom: "6px solid transparent",
          borderLeft: "10px solid white",
          transform: `scale(${arrowHeadScale})`,
          transformOrigin: "left center",
          position: "absolute",
          right: -10,
          top: -5,
        }}
      />
    </div>
  );
};

const Box: React.FC<{ label: string; i: number; scrollPosition: number }> = ({ label, i, scrollPosition }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enterStart = i * 60; // every 2 s
  const isLastBox = i === STEPS.length - 1;

  // slide from 60 px left
  const translateX = interpolate(
    frame,
    [enterStart, enterStart + 30],
    [-60, 0],
    { extrapolateRight: "clamp" }
  );
  
  // Scale animation with spring physics
  const scale = spring({
    frame: frame - enterStart,
    fps,
    config: { damping: 200, mass: 0.9 },
  });

  // Subtle floating animation after appearance
  const floatY = i % 2 === 0 
    ? Math.sin((frame - enterStart - 30) / 20) * 4
    : Math.sin((frame - enterStart - 30) / 25) * 3;
  
  const yOffset = interpolate(
    frame,
    [enterStart, enterStart + 30, Number.MAX_SAFE_INTEGER],
    [0, 0, floatY],
    { extrapolateRight: "clamp" }
  );

  // Text opacity effect
  const textOpacity = interpolate(
    frame,
    [enterStart, enterStart + 40],
    [0, 1],
    { extrapolateRight: "clamp" }
  );

  // Glow effect intensity
  const glowIntensity = interpolate(
    frame, 
    [enterStart, enterStart + 40, enterStart + 70, enterStart + 100],
    [0, 8, 4, 3],
    { extrapolateRight: "clamp" }
  );

  // Special pulsing effect for the final box
  const finalBoxPulse = isLastBox 
    ? 3 + Math.sin(frame * 0.1) * 2 
    : 0;
  
  const finalBoxScale = isLastBox && frame > enterStart + 60
    ? 1 + Math.sin(frame * 0.05) * 0.03
    : 1;

  const posX = i * (BOX_W + SPACING);
  // Adjust position for scrolling
  const adjustedX = posX - scrollPosition;

  return (
    <>
      <div
        style={{
          position: "absolute",
          left: adjustedX,
          top: 420,
          width: BOX_W,
          height: BOX_H,
          borderRadius: 12,
          background: getBoxGradient(i),
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          color: "white",
          fontSize: 28,
          fontWeight: "bold",
          boxShadow: isLastBox 
            ? `0 0 ${glowIntensity + finalBoxPulse}px ${COLORS[i % COLORS.length][0]}, 0 0 ${glowIntensity + finalBoxPulse + 5}px rgba(255,255,255,0.5)` 
            : `0 0 ${glowIntensity}px ${COLORS[i % COLORS.length][0]}`,
          transform: `translateX(${translateX}px) translateY(${yOffset}px) scale(${scale * finalBoxScale})`,
          transition: "box-shadow 0.3s",
        }}
      >
        <div
          style={{
            opacity: textOpacity,
            textShadow: "0 2px 4px rgba(0,0,0,0.3)",
          }}
        >
          {label}
        </div>
      </div>
      
      {/* Add arrows between boxes, but not after the last box */}
      {i < STEPS.length - 1 && (
        <Arrow fromX={posX} i={i} scrollPosition={scrollPosition} />
      )}
    </>
  );
};

// Add a starry background
const StarryBackground: React.FC = () => {
  const stars = React.useMemo(() => {
    return Array.from({ length: 100 }).map((_, i) => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 3 + 1,
      blinkSpeed: Math.random() * 2 + 0.5,
    }));
  }, []);

  const frame = useCurrentFrame();

  return (
    <div
      style={{
        position: "absolute",
        width: "100%",
        height: "100%",
        background: "linear-gradient(to bottom, #0d0d2b, #0d1117)",
        overflow: "hidden",
      }}
    >
      {stars.map((star, i) => (
        <div
          key={i}
          style={{
            position: "absolute",
            left: `${star.x}%`,
            top: `${star.y}%`,
            width: star.size,
            height: star.size,
            borderRadius: "50%",
            backgroundColor: "white",
            opacity: 0.3 + Math.sin(frame * 0.1 * star.blinkSpeed) * 0.7,
          }}
        />
      ))}
    </div>
  );
};

// Particles system for added visual flair
const Particles: React.FC = () => {
  const particles = React.useMemo(() => {
    return Array.from({ length: 30 }).map(() => ({
      x: Math.random() * 100,
      y: Math.random() * 100,
      size: Math.random() * 6 + 2,
      speed: Math.random() * 0.3 + 0.1,
      color: COLORS[Math.floor(Math.random() * COLORS.length)][0],
    }));
  }, []);

  const frame = useCurrentFrame();

  return (
    <div style={{ position: "absolute", width: "100%", height: "100%" }}>
      {particles.map((particle, i) => (
        <div
          key={i}
          style={{
            position: "absolute",
            left: `${particle.x}%`,
            top: `${(particle.y + frame * particle.speed) % 100}%`,
            width: particle.size,
            height: particle.size,
            borderRadius: "50%",
            background: particle.color,
            opacity: 0.3,
            filter: "blur(2px)",
          }}
        />
      ))}
    </div>
  );
};

// Title component
const Title: React.FC = () => {
  const frame = useCurrentFrame();
  const opacity = interpolate(frame, [0, 20], [0, 1], {
    extrapolateRight: "clamp",
  });
  
  // Subtle movement for the title
  const titleOffset = Math.sin(frame * 0.02) * 5;
  
  return (
    <div
      style={{
        position: "absolute",
        top: 150 + titleOffset,
        width: "100%",
        textAlign: "center",
        color: "white",
        fontWeight: "bold",
        fontSize: 48,
        opacity,
        textShadow: "0 0 10px rgba(255,255,255,0.5)",
      }}
    >
      Content Creation Workflow
    </div>
  );
};

// ScrollableContainer to keep items visible
const ScrollableContainer: React.FC = () => {
  const frame = useCurrentFrame();
  const { width } = useVideoConfig();
  
  // Calculate viewport width
  const viewportWidth = width;
  
  // Calculate scroll position to ensure that the active box is always visible
  // We'll start scrolling once we reach the 4th box to keep everything in frame
  const lastVisibleIndex = Math.floor(frame / 60);
  const scrollingStartsAt = 4; // Start scrolling at the 5th box (index 4)
  
  // Only scroll if we're past the scrolling threshold
  let scrollPosition = 0;
  if (lastVisibleIndex >= scrollingStartsAt) {
    // Calculate position of the current active box
    const activeBoxPosition = lastVisibleIndex * (BOX_W + SPACING);
    // Center it in the viewport (accounting for the first few boxes we want to keep visible)
    scrollPosition = Math.max(0, activeBoxPosition - viewportWidth * 0.5 + BOX_W);
  }
  
  return (
    <div style={{ position: "relative", width: "100%", height: "100%" }}>
      {STEPS.map((label, i) => (
        <Sequence key={label} from={i * 60} durationInFrames={540 - i * 60}>
          <Box label={label} i={i} scrollPosition={scrollPosition} />
        </Sequence>
      ))}
    </div>
  );
};

export const FlowScene: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: "#0d1117", overflow: "hidden" }}>
    <StarryBackground />
    <Particles />
    <Title />
    <ScrollableContainer />
  </AbsoluteFill>
); 