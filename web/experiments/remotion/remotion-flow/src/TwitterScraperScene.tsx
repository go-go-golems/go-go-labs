import { interpolate, useCurrentFrame, spring, useVideoConfig, random } from "remotion";
import { AbsoluteFill } from "remotion";
import React from "react";

// Twitter colors
const TWITTER_BLUE = "#1DA1F2";
const DARK_BG = "#121212";
const ACCENT_PURPLE = "#8E2DE2";

// Steps in our Twitter scraper + LLM workflow
const STEPS = [
  "Twitter Data",
  "Scraper",
  "Raw Data",
  "LLM Processing",
  "Analysis",
] as const;

// Layout constants
const BOX_W = 900;
const BOX_H = 400;
// Much faster animation timing
const STEP_DURATION = 30; // frames per step (was 60)

// Color gradients for visual appeal
const COLORS = [
  ["#1DA1F2", "#0C86C8"], // Twitter blue
  ["#8E2DE2", "#4A00E0"], // Purple for code
  ["#16A085", "#F4D03F"], // Green to Yellow for data
  ["#FF4B2B", "#FF416C"], // Red for processing
  ["#00c6ff", "#0072ff"], // Light blue for insights
];

// Helper for gradient backgrounds
const getBoxGradient = (i: number) => {
  const colors = COLORS[i % COLORS.length];
  return `linear-gradient(135deg, ${colors[0]}, ${colors[1]})`;
};

// Box component representing each step in workflow - now centers on screen with punch effect
const Box: React.FC<{ label: string; i: number; isActive: boolean }> = ({ label, i, isActive }) => {
  const frame = useCurrentFrame();
  const { fps, width, height } = useVideoConfig();
  const localFrame = frame % STEP_DURATION;
  
  // More aggressive spring physics for punchier animations
  const entryScale = spring({
    frame: localFrame,
    fps,
    config: { 
      damping: 12, // Less damping for more bounce
      mass: 0.4,   // Lighter mass
      stiffness: 200 // Higher stiffness for faster movement
    },
  });
  
  // Sharp punch effect on entry
  const punchEffect = interpolate(
    localFrame,
    [0, 5, 10],
    [0.7, 1.2, 1], // Quick overshoot and settle
    { extrapolateRight: 'clamp' }
  );
  
  // Dramatic rotation on entry
  const rotation = interpolate(
    localFrame,
    [0, 6, 8],
    [25, 0, 0],
    { extrapolateRight: 'clamp' }
  );
  
  // Sharp opacity change
  const opacity = interpolate(
    localFrame,
    [0, 6, STEP_DURATION - 5, STEP_DURATION],
    [0, 1, 1, 0],
    { extrapolateRight: 'clamp' }
  );
  
  // Vibration effect for punch
  const vibrationX = localFrame < 10 
    ? Math.sin(localFrame * 3) * (10 - localFrame) 
    : 0;
  
  // Center position
  const centerX = width / 2 - BOX_W / 2;
  const centerY = height / 2 - BOX_H / 2;
  
  // Box icons based on step type
  const getIcon = () => {
    switch(i) {
      case 0: return "🐦"; // Twitter
      case 1: return "🔍"; // Scraper
      case 2: return "📊"; // Data
      case 3: return "🧠"; // LLM
      case 4: return "📈"; // Analytics
      default: return "💻";
    }
  };

  // Only render the active box
  if (!isActive) return null;

  return (
    <>
      <div
        style={{
          position: "absolute",
          left: centerX + vibrationX,
          top: centerY,
          width: BOX_W,
          height: BOX_H,
          borderRadius: 30,
          background: getBoxGradient(i),
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "center",
          color: "white",
          fontSize: 60,
          fontWeight: "bold",
          boxShadow: `0 0 50px ${COLORS[i % COLORS.length][0]}`,
          transform: `scale(${entryScale * punchEffect}) rotate(${rotation}deg)`,
          opacity,
          transition: "all 0.1s ease-out", // Fast transitions
          zIndex: 10,
        }}
      >
        <div style={{ fontSize: 160, marginBottom: 0 }}>{getIcon()}</div>
        <div
          style={{
            textShadow: "0 4px 16px rgba(0,0,0,0.5)",
            textAlign: "center",
          }}
        >
          {label}
        </div>
      </div>
    </>
  );
};

// Code particles effect - faster movement
const CodeParticles: React.FC = () => {
  const particles = React.useMemo(() => {
    // Generate code snippets for particles
    const codeSnippets = [
      "async", "await", "const data", "fetch", 
      "import", "axios", "parse", "json", 
      "LLM", "API", "tweets", "analyze",
      "for(let i", "function", "return"
    ];
    
    return Array.from({ length: 40 }).map(() => ({
      x: random(null) * 100,
      y: random(null) * 100,
      text: codeSnippets[Math.floor(random(null) * codeSnippets.length)],
      size: random(null) * 4 + 10,
      speed: random(null) * 0.8 + 0.3, // Faster movement
      opacity: random(null) * 0.4 + 0.2,
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
            fontFamily: "monospace",
            fontSize: particle.size,
            color: i % 3 === 0 ? TWITTER_BLUE : i % 3 === 1 ? "#FFFFFF" : ACCENT_PURPLE,
            opacity: particle.opacity,
          }}
        >
          {particle.text}
        </div>
      ))}
    </div>
  );
};

// Star background - brighter and more dynamic
const StarryBackground: React.FC = () => {
  const stars = React.useMemo(() => {
    return Array.from({ length: 100 }).map(() => ({
      x: random(null) * 100,
      y: random(null) * 100,
      size: random(null) * 3 + 1,
      blinkSpeed: random(null) * 0.2 + 0.1, // Faster blinking
    }));
  }, []);

  const frame = useCurrentFrame();

  return (
    <div
      style={{
        position: "absolute",
        width: "100%",
        height: "100%",
        background: `linear-gradient(135deg, #15202B, ${DARK_BG})`,
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
            opacity: 0.3 + Math.sin(frame * star.blinkSpeed) * 0.6, // More dramatic blinking
          }}
        />
      ))}
    </div>
  );
};

// Title component - faster typing
const Title: React.FC = () => {
  const frame = useCurrentFrame();
  
  // Animation for title - faster fade-in
  const opacity = interpolate(frame, [0, 10], [0, 1], {
    extrapolateRight: "clamp",
  });
  
  // Typing effect - faster typing
  const titleText = "Building a Twitter Scraper with LLMs";
  const typedChars = Math.min(
    titleText.length,
    Math.floor(interpolate(frame, [10, 30], [0, titleText.length], {
      extrapolateRight: "clamp",
    }))
  );
  
  const displayedText = titleText.substring(0, typedChars);
  const cursor = frame % 20 < 10 || typedChars === titleText.length ? "|" : "";
  
  return (
    <div
      style={{
        position: "absolute",
        top: 100,
        width: "100%",
        textAlign: "center",
        color: "white",
        opacity,
      }}
    >
      <div
        style={{
          fontWeight: "bold",
          fontSize: 80,
          textShadow: `0 0 20px ${TWITTER_BLUE}`,
          fontFamily: "monospace",
        }}
      >
        {displayedText}{cursor}
      </div>
      <div
        style={{
          marginTop: 20,
          fontSize: 48,
          opacity: interpolate(
            frame,
            [30, 40],
            [0, 1],
            { extrapolateRight: "clamp" }
          ),
        }}
      >
        Coding with AI Assistants
      </div>
    </div>
  );
};

// Workflow container - now shows one item at a time in center
const WorkflowContainer: React.FC = () => {
  const frame = useCurrentFrame();
  
  // Calculate which step is currently active
  const currentStep = Math.min(
    Math.floor(frame / STEP_DURATION),
    STEPS.length - 1
  );
  
  return (
    <div style={{ position: "relative", marginTop: 100, width: "100%", height: "100%" }}>
      {STEPS.map((label, i) => (
        <Box 
          key={label} 
          label={label} 
          i={i} 
          isActive={i === currentStep} 
        />
      ))}
    </div>
  );
};

// Main scene composition
export const TwitterScraperScene: React.FC = () => (
  <AbsoluteFill style={{ backgroundColor: DARK_BG, overflow: "hidden" }}>
    <StarryBackground />
    <CodeParticles />
    <Title />
    <WorkflowContainer />
  </AbsoluteFill>
); 