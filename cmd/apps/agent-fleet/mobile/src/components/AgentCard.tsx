import React from "react";
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Animated,
} from "react-native";
import { useDispatch } from "react-redux";
import { Agent } from "@/types/api";
import { setFeedbackModalVisible, setFeedbackTargetAgent } from "@/store/slices/uiSlice";

interface AgentCardProps {
  agent: Agent;
  onPress: () => void;
}

export default function AgentCard({ agent, onPress }: AgentCardProps) {
  const dispatch = useDispatch();
  const getStatusColor = (status: Agent["status"]) => {
    switch (status) {
      case "active":
        return "#10B981"; // green
      case "idle":
        return "#6B7280"; // gray
      case "waiting_feedback":
        return "#F59E0B"; // orange
      case "error":
        return "#EF4444"; // red
      case "finished":
        return "#8B5CF6"; // purple
      case "warning":
        return "#F97316"; // orange-red
      default:
        return "#6B7280";
    }
  };

  const getStatusIndicator = (status: Agent["status"]) => {
    switch (status) {
      case "active":
        return "●";
      case "idle":
        return "○";
      case "waiting_feedback":
        return "⚠️";
      case "error":
        return "❌";
      case "finished":
        return "✅";
      case "warning":
        return "⚠️";
      default:
        return "○";
    }
  };

  const getBorderColor = () => {
    if (agent.status === "waiting_feedback" || agent.status === "warning") {
      return "#F59E0B"; // orange for feedback needed or warning
    }
    if (agent.status === "error") {
      return "#EF4444"; // red for error
    }
    return "#374151"; // default gray
  };

  const formatTimeAgo = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diffMs = now.getTime() - time.getTime();
    const diffMins = Math.floor(diffMs / 60000);

    if (diffMins < 1) return "now";
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  const handleQuickFeedback = (e: any) => {
    e.stopPropagation(); // Prevent card press
    dispatch(setFeedbackTargetAgent(agent));
    dispatch(setFeedbackModalVisible(true));
  };

  return (
    <TouchableOpacity
      style={[
        styles.card,
        {
          borderColor: getBorderColor(),
          borderWidth:
            agent.status === "waiting_feedback" ||
            agent.status === "warning" ||
            agent.status === "error"
              ? 2
              : 1,
        },
      ]}
      onPress={onPress}
      activeOpacity={0.7}
    >
      {/* Header */}
      <View style={styles.header}>
        <View style={styles.statusRow}>
          <Text
            style={[
              styles.statusIndicator,
              { color: getStatusColor(agent.status) },
            ]}
          >
            {getStatusIndicator(agent.status)}
          </Text>
          <Text style={styles.agentName}>{agent.name}</Text>
          <View
            style={[
              styles.statusBadge,
              { backgroundColor: getStatusColor(agent.status) },
            ]}
          >
            <Text style={styles.statusText}>{agent.status}</Text>
          </View>
        </View>
      </View>

      {/* Current Task */}
      <View style={styles.taskSection}>
        <Text style={styles.taskText} numberOfLines={2}>
          {agent.current_task || "No current task"}
        </Text>
      </View>

      {/* Current Step */}
      {agent.current_step && (
        <View style={styles.stepSection}>
          <Text style={styles.stepLabel}>Current Step:</Text>
          <Text style={styles.stepText} numberOfLines={2}>
            {agent.current_step}
          </Text>
        </View>
      )}

      {/* Recent Steps */}
      {agent.recent_steps && agent.recent_steps.length > 0 && (
        <View style={styles.recentStepsSection}>
          <Text style={styles.stepLabel}>Recent Steps:</Text>
          {agent.recent_steps.slice(0, 3).map((step, index) => (
            <View key={step.id} style={styles.stepItem}>
              <Text
                style={[
                  styles.stepItemText,
                  {
                    color:
                      step.status === "completed"
                        ? "#10B981"
                        : step.status === "failed"
                        ? "#EF4444"
                        : "#D1D5DB",
                  },
                ]}
                numberOfLines={1}
              >
                {step.status === "completed"
                  ? "✓"
                  : step.status === "failed"
                  ? "✗"
                  : "○"}{" "}
                {step.text}
              </Text>
            </View>
          ))}
        </View>
      )}

      {/* Question Box if pending */}
      {agent.pending_question && (
        <View style={styles.questionBox}>
          <View style={styles.questionContent}>
            <Text style={styles.questionText} numberOfLines={2}>
              ❓ {agent.pending_question}
            </Text>
            <TouchableOpacity
              style={styles.quickResponseButton}
              onPress={handleQuickFeedback}
            >
              <Text style={styles.quickResponseButtonText}>Respond</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}

      {/* Warning Box if present */}
      {agent.warning_message && (
        <View style={styles.warningBox}>
          <Text style={styles.warningText} numberOfLines={2}>
            ⚠️ {agent.warning_message}
          </Text>
        </View>
      )}

      {/* Error Box if present */}
      {agent.error_message && (
        <View style={styles.errorBox}>
          <Text style={styles.errorText} numberOfLines={2}>
            ❌ {agent.error_message}
          </Text>
        </View>
      )}

      {/* Metadata Row */}
      <View style={styles.metadataRow}>
        <Text style={styles.metadataText}>🌿 {agent.worktree}</Text>
        <Text style={styles.metadataText}>
          📝 {formatTimeAgo(agent.last_commit)}
        </Text>
      </View>

      {/* Stats Row */}
      <View style={styles.statsRow}>
        <Text style={styles.metadataText}>📁 {agent.files_changed} files</Text>
        <Text style={styles.additionsText}>+{agent.lines_added}</Text>
        <Text style={styles.deletionsText}>-{agent.lines_removed}</Text>
      </View>

      {/* Progress Bar */}
      <View style={styles.progressContainer}>
        <View style={styles.progressBar}>
          <View
            style={[
              styles.progressFill,
              {
                width: `${agent.progress}%`,
                backgroundColor: getStatusColor(agent.status),
              },
            ]}
          />
        </View>
        <Text style={styles.progressText}>{agent.progress}%</Text>
      </View>
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  card: {
    backgroundColor: "#111827",
    borderRadius: 8,
    padding: 16,
    marginHorizontal: 16,
    marginVertical: 8,
  },
  header: {
    height: 48,
    justifyContent: "center",
  },
  statusRow: {
    flexDirection: "row",
    alignItems: "center",
  },
  statusIndicator: {
    fontSize: 16,
    marginRight: 8,
  },
  agentName: {
    color: "#ffffff",
    fontSize: 16,
    fontWeight: "600",
    flex: 1,
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 12,
  },
  statusText: {
    color: "#ffffff",
    fontSize: 12,
    fontWeight: "500",
  },
  taskSection: {
    height: 32,
    justifyContent: "center",
    borderTopWidth: 1,
    borderTopColor: "#374151",
    paddingTop: 8,
  },
  taskText: {
    color: "#D1D5DB",
    fontSize: 14,
  },
  questionBox: {
    backgroundColor: "#FEF3C7",
    borderRadius: 6,
    padding: 8,
    marginTop: 8,
  },
  questionContent: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  questionText: {
    color: "#92400E",
    fontSize: 14,
    flex: 1,
    marginRight: 12,
  },
  quickResponseButton: {
    backgroundColor: "#F59E0B",
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 4,
  },
  quickResponseButtonText: {
    color: "#FFFFFF",
    fontSize: 12,
    fontWeight: "600",
  },
  warningBox: {
    backgroundColor: "#FEF1F2",
    borderRadius: 6,
    padding: 8,
    marginTop: 8,
  },
  warningText: {
    color: "#B91C1C",
    fontSize: 14,
  },
  errorBox: {
    backgroundColor: "#FEF2F2",
    borderRadius: 6,
    padding: 8,
    marginTop: 8,
    borderWidth: 1,
    borderColor: "#EF4444",
  },
  errorText: {
    color: "#B91C1C",
    fontSize: 14,
    fontWeight: "600",
  },
  stepSection: {
    marginTop: 8,
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: "#374151",
  },
  stepLabel: {
    color: "#9CA3AF",
    fontSize: 12,
    fontWeight: "500",
    marginBottom: 4,
  },
  stepText: {
    color: "#D1D5DB",
    fontSize: 14,
  },
  recentStepsSection: {
    marginTop: 8,
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: "#374151",
  },
  stepItem: {
    marginTop: 2,
  },
  stepItemText: {
    fontSize: 12,
  },
  metadataRow: {
    flexDirection: "row",
    justifyContent: "space-between",
    height: 24,
    alignItems: "center",
    borderTopWidth: 1,
    borderTopColor: "#374151",
    paddingTop: 8,
  },
  statsRow: {
    flexDirection: "row",
    alignItems: "center",
    height: 24,
    gap: 16,
  },
  metadataText: {
    color: "#9CA3AF",
    fontSize: 12,
  },
  additionsText: {
    color: "#10B981",
    fontSize: 12,
    fontWeight: "500",
  },
  deletionsText: {
    color: "#EF4444",
    fontSize: 12,
    fontWeight: "500",
  },
  progressContainer: {
    flexDirection: "row",
    alignItems: "center",
    height: 16,
    marginTop: 8,
    borderTopWidth: 1,
    borderTopColor: "#374151",
    paddingTop: 8,
  },
  progressBar: {
    flex: 1,
    height: 4,
    backgroundColor: "#374151",
    borderRadius: 2,
    marginRight: 8,
  },
  progressFill: {
    height: "100%",
    borderRadius: 2,
  },
  progressText: {
    color: "#9CA3AF",
    fontSize: 12,
    fontWeight: "500",
    minWidth: 32,
  },
});
