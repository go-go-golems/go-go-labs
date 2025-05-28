import React from "react";
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Animated,
} from "react-native";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "@/store";
import {
  dismissNotification,
  removeNotification,
  setFeedbackModalVisible,
  setFeedbackTargetAgent,
} from "@/store/slices/uiSlice";
import { useGetAgentQuery } from "@/services/api";

export default function NotificationBanner() {
  const dispatch = useDispatch();
  const notifications = useSelector((state: RootState) =>
    state.ui.notifications.filter((n) => !n.dismissed)
  );

  const activeNotification = notifications[0]; // Show the most recent notification

  const { data: agent } = useGetAgentQuery(activeNotification?.agentId || "", {
    skip: !activeNotification?.agentId,
  });

  if (!activeNotification) {
    return null;
  }

  const handlePress = () => {
    if (activeNotification.type === "question") {
      // Open feedback modal for questions
      dispatch(setFeedbackTargetAgent(agent || null));
      dispatch(setFeedbackModalVisible(true));
    }
    dispatch(dismissNotification(activeNotification.id));
  };

  const handleDismiss = () => {
    dispatch(dismissNotification(activeNotification.id));
  };

  const getNotificationStyle = () => {
    switch (activeNotification.type) {
      case "question":
        return { backgroundColor: "#FEF3C7", borderColor: "#F59E0B" };
      case "warning":
        return { backgroundColor: "#FEF1F2", borderColor: "#F97316" };
      case "error":
        return { backgroundColor: "#FEF2F2", borderColor: "#EF4444" };
      default:
        return { backgroundColor: "#EBF8FF", borderColor: "#3B82F6" };
    }
  };

  const getNotificationIcon = () => {
    switch (activeNotification.type) {
      case "question":
        return "❓";
      case "warning":
        return "⚠️";
      case "error":
        return "❌";
      default:
        return "ℹ️";
    }
  };

  const getTextColor = () => {
    switch (activeNotification.type) {
      case "question":
        return "#92400E";
      case "warning":
        return "#C2410C";
      case "error":
        return "#B91C1C";
      default:
        return "#1E40AF";
    }
  };

  return (
    <TouchableOpacity
      style={[styles.container, getNotificationStyle()]}
      onPress={handlePress}
      activeOpacity={0.8}
    >
      <View style={styles.content}>
        <Text style={styles.icon}>{getNotificationIcon()}</Text>
        <View style={styles.textContainer}>
          <Text style={[styles.title, { color: getTextColor() }]}>
            {activeNotification.title}
          </Text>
          <Text
            style={[styles.message, { color: getTextColor() }]}
            numberOfLines={2}
          >
            {activeNotification.message}
          </Text>
          <Text style={[styles.agentName, { color: getTextColor() }]}>
            {activeNotification.agentName}
          </Text>
        </View>
        <TouchableOpacity
          style={styles.dismissButton}
          onPress={handleDismiss}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <Text style={[styles.dismissText, { color: getTextColor() }]}>✕</Text>
        </TouchableOpacity>
      </View>

      {activeNotification.type === "question" && (
        <View style={styles.actionHint}>
          <Text style={[styles.actionHintText, { color: getTextColor() }]}>
            Tap to respond
          </Text>
        </View>
      )}
    </TouchableOpacity>
  );
}

const styles = StyleSheet.create({
  container: {
    margin: 16,
    borderRadius: 8,
    borderWidth: 2,
    overflow: "hidden",
  },
  content: {
    flexDirection: "row",
    alignItems: "center",
    padding: 12,
  },
  icon: {
    fontSize: 20,
    marginRight: 12,
  },
  textContainer: {
    flex: 1,
  },
  title: {
    fontSize: 16,
    fontWeight: "600",
    marginBottom: 2,
  },
  message: {
    fontSize: 14,
    marginBottom: 4,
  },
  agentName: {
    fontSize: 12,
    fontWeight: "500",
  },
  dismissButton: {
    padding: 4,
  },
  dismissText: {
    fontSize: 16,
    fontWeight: "bold",
  },
  actionHint: {
    paddingHorizontal: 12,
    paddingBottom: 8,
  },
  actionHintText: {
    fontSize: 12,
    fontStyle: "italic",
  },
});
