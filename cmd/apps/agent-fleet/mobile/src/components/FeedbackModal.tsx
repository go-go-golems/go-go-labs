import React, { useState } from "react";
import {
  Modal,
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from "react-native";
import { useSelector, useDispatch } from "react-redux";
import { RootState } from "@/store";
import {
  setFeedbackModalVisible,
  setFeedbackTargetAgent,
} from "@/store/slices/uiSlice";
import { useSendCommandMutation } from "@/services/api";

export default function FeedbackModal() {
  const dispatch = useDispatch();
  const { feedbackModalVisible, feedbackTargetAgent } = useSelector(
    (state: RootState) => state.ui
  );
  const [feedbackText, setFeedbackText] = useState("");
  const [feedbackType, setFeedbackType] = useState<"feedback" | "instruction">(
    "feedback"
  );

  const [sendCommand, { isLoading: isSending }] = useSendCommandMutation();

  const handleClose = () => {
    dispatch(setFeedbackModalVisible(false));
    dispatch(setFeedbackTargetAgent(null));
    setFeedbackText("");
    setFeedbackType("feedback");
  };

  const handleSend = async () => {
    if (!feedbackTargetAgent || !feedbackText.trim()) {
      Alert.alert("Error", "Please enter a response");
      return;
    }

    try {
      await sendCommand({
        agentId: feedbackTargetAgent.id,
        content: feedbackText.trim(),
        type: feedbackType,
      }).unwrap();

      Alert.alert("Success", "Response sent to agent");
      handleClose();
    } catch (error) {
      console.error("Failed to send feedback:", error);
      Alert.alert("Error", "Failed to send response. Please try again.");
    }
  };

  if (!feedbackTargetAgent) {
    return null;
  }

  return (
    <Modal
      visible={feedbackModalVisible}
      animationType="slide"
      presentationStyle="pageSheet"
      onRequestClose={handleClose}
    >
      <KeyboardAvoidingView
        style={styles.container}
        behavior={Platform.OS === "ios" ? "padding" : "height"}
      >
        <View style={styles.header}>
          <TouchableOpacity onPress={handleClose} style={styles.closeButton}>
            <Text style={styles.closeButtonText}>✕</Text>
          </TouchableOpacity>
          <Text style={styles.title}>Respond to Agent</Text>
          <View style={styles.placeholder} />
        </View>

        <ScrollView
          style={styles.content}
          contentContainerStyle={styles.contentContainer}
        >
          <View style={styles.agentInfo}>
            <Text style={styles.agentName}>{feedbackTargetAgent.name}</Text>
            <View style={styles.statusContainer}>
              <Text style={[styles.status, { color: "#F59E0B" }]}>
                {feedbackTargetAgent.status}
              </Text>
            </View>
          </View>

          {feedbackTargetAgent.pending_question && (
            <View style={styles.questionContainer}>
              <Text style={styles.questionLabel}>Agent's Question:</Text>
              <Text style={styles.questionText}>
                {feedbackTargetAgent.pending_question}
              </Text>
            </View>
          )}

          {feedbackTargetAgent.current_task && (
            <View style={styles.taskContainer}>
              <Text style={styles.taskLabel}>Current Task:</Text>
              <Text style={styles.taskText}>
                {feedbackTargetAgent.current_task}
              </Text>
            </View>
          )}

          <View style={styles.responseTypeContainer}>
            <Text style={styles.responseTypeLabel}>Response Type:</Text>
            <View style={styles.responseTypeButtons}>
              <TouchableOpacity
                style={[
                  styles.responseTypeButton,
                  feedbackType === "feedback" &&
                    styles.responseTypeButtonActive,
                ]}
                onPress={() => setFeedbackType("feedback")}
              >
                <Text
                  style={[
                    styles.responseTypeButtonText,
                    feedbackType === "feedback" &&
                      styles.responseTypeButtonTextActive,
                  ]}
                >
                  💬 Feedback
                </Text>
              </TouchableOpacity>
              <TouchableOpacity
                style={[
                  styles.responseTypeButton,
                  feedbackType === "instruction" &&
                    styles.responseTypeButtonActive,
                ]}
                onPress={() => setFeedbackType("instruction")}
              >
                <Text
                  style={[
                    styles.responseTypeButtonText,
                    feedbackType === "instruction" &&
                      styles.responseTypeButtonTextActive,
                  ]}
                >
                  📋 Instruction
                </Text>
              </TouchableOpacity>
            </View>
          </View>

          <View style={styles.inputContainer}>
            <Text style={styles.inputLabel}>Your Response:</Text>
            <TextInput
              style={styles.textInput}
              value={feedbackText}
              onChangeText={setFeedbackText}
              placeholder={
                feedbackType === "feedback"
                  ? "Provide feedback or answer the agent's question..."
                  : "Give specific instructions to the agent..."
              }
              placeholderTextColor="#9CA3AF"
              multiline
              numberOfLines={6}
              textAlignVertical="top"
            />
          </View>
        </ScrollView>

        <View style={styles.actions}>
          <TouchableOpacity style={styles.cancelButton} onPress={handleClose}>
            <Text style={styles.cancelButtonText}>Cancel</Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[
              styles.sendButton,
              (!feedbackText.trim() || isSending) && styles.sendButtonDisabled,
            ]}
            onPress={handleSend}
            disabled={!feedbackText.trim() || isSending}
          >
            <Text
              style={[
                styles.sendButtonText,
                (!feedbackText.trim() || isSending) &&
                  styles.sendButtonTextDisabled,
              ]}
            >
              {isSending ? "Sending..." : "Send Response"}
            </Text>
          </TouchableOpacity>
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#000000",
  },
  header: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: "#374151",
  },
  closeButton: {
    padding: 8,
  },
  closeButtonText: {
    color: "#9CA3AF",
    fontSize: 18,
    fontWeight: "bold",
  },
  title: {
    color: "#FFFFFF",
    fontSize: 18,
    fontWeight: "600",
  },
  placeholder: {
    width: 34,
  },
  content: {
    flex: 1,
  },
  contentContainer: {
    padding: 16,
  },
  agentInfo: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    marginBottom: 16,
    padding: 12,
    backgroundColor: "#111827",
    borderRadius: 8,
  },
  agentName: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  statusContainer: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 12,
    backgroundColor: "#374151",
  },
  status: {
    fontSize: 12,
    fontWeight: "500",
    textTransform: "capitalize",
  },
  questionContainer: {
    marginBottom: 16,
    padding: 12,
    backgroundColor: "#FEF3C7",
    borderRadius: 8,
  },
  questionLabel: {
    color: "#92400E",
    fontSize: 14,
    fontWeight: "600",
    marginBottom: 4,
  },
  questionText: {
    color: "#92400E",
    fontSize: 14,
    lineHeight: 20,
  },
  taskContainer: {
    marginBottom: 16,
    padding: 12,
    backgroundColor: "#111827",
    borderRadius: 8,
  },
  taskLabel: {
    color: "#9CA3AF",
    fontSize: 12,
    fontWeight: "600",
    marginBottom: 4,
  },
  taskText: {
    color: "#D1D5DB",
    fontSize: 14,
    lineHeight: 20,
  },
  responseTypeContainer: {
    marginBottom: 16,
  },
  responseTypeLabel: {
    color: "#FFFFFF",
    fontSize: 14,
    fontWeight: "600",
    marginBottom: 8,
  },
  responseTypeButtons: {
    flexDirection: "row",
    gap: 12,
  },
  responseTypeButton: {
    flex: 1,
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: "#374151",
    backgroundColor: "#111827",
    alignItems: "center",
  },
  responseTypeButtonActive: {
    borderColor: "#3B82F6",
    backgroundColor: "#1E3A8A",
  },
  responseTypeButtonText: {
    color: "#D1D5DB",
    fontSize: 14,
    fontWeight: "500",
  },
  responseTypeButtonTextActive: {
    color: "#FFFFFF",
  },
  inputContainer: {
    marginBottom: 16,
  },
  inputLabel: {
    color: "#FFFFFF",
    fontSize: 14,
    fontWeight: "600",
    marginBottom: 8,
  },
  textInput: {
    backgroundColor: "#111827",
    color: "#FFFFFF",
    fontSize: 16,
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: "#374151",
    minHeight: 120,
  },
  actions: {
    flexDirection: "row",
    gap: 12,
    padding: 16,
    borderTopWidth: 1,
    borderTopColor: "#374151",
  },
  cancelButton: {
    flex: 1,
    paddingVertical: 12,
    borderRadius: 8,
    backgroundColor: "#374151",
    alignItems: "center",
  },
  cancelButtonText: {
    color: "#D1D5DB",
    fontSize: 16,
    fontWeight: "600",
  },
  sendButton: {
    flex: 2,
    paddingVertical: 12,
    borderRadius: 8,
    backgroundColor: "#3B82F6",
    alignItems: "center",
  },
  sendButtonDisabled: {
    backgroundColor: "#6B7280",
  },
  sendButtonText: {
    color: "#FFFFFF",
    fontSize: 16,
    fontWeight: "600",
  },
  sendButtonTextDisabled: {
    color: "#9CA3AF",
  },
});
