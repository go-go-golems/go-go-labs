// src/nodeConfig.ts

interface NodeStyle {
  color: string; // Primary color for border/accents
  backgroundColor: string; // Background for the node box
  icon: string; // Text/emoji representation
}

const config: Record<string, NodeStyle> = {
  goal: {
    color: "#3f51b5", // Indigo
    backgroundColor: "#e8eaf6",
    icon: "🎯",
  },
  subtask: {
    color: "#03a9f4", // Light Blue
    backgroundColor: "#e1f5fe",
    icon: "📚",
  },
  action: {
    color: "#ff9800", // Orange
    backgroundColor: "#fff3e0",
    icon: "⚡",
  },
  source: {
    color: "#4caf50", // Green
    backgroundColor: "#e8f5e9",
    icon: "🏁",
  },
  email: {
    color: "#009688", // Teal
    backgroundColor: "#e0f2f1",
    icon: "📧",
  },
  wait: {
    color: "#ffc107", // Amber
    backgroundColor: "#fff8e1",
    icon: "⏱️",
  },
  sms: {
    color: "#9c27b0", // Purple
    backgroundColor: "#f3e5f5",
    icon: "💬",
  },
  end: {
    color: "#f44336", // Red
    backgroundColor: "#ffebee",
    icon: "🛑",
  },
  default: {
    color: "#757575", // Grey
    backgroundColor: "#f5f5f5",
    icon: "❓",
  },
};

export const nodeConfig = (type: string = "default"): NodeStyle => {
  return config[type] || config.default;
};
