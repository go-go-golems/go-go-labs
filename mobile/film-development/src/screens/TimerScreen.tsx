import React, { useState, useEffect, useRef } from "react";
import {
  View,
  Text,
  TouchableOpacity,
  ScrollView,
  StyleSheet,
  Alert,
  AppState,
} from "react-native";
import { SafeAreaView } from "react-native-safe-area-context";
import * as Haptics from "expo-haptics";
import { activateKeepAwakeAsync, deactivateKeepAwake } from "expo-keep-awake";
import { Audio } from "expo-av";

import { TimerDisplay } from "../components/TimerDisplay";
import { ProcessStepCard } from "../components/ProcessStepCard";
import { Roll, ProcessStep, TimerState } from "../types";
import {
  scheduleStepNotifications,
  cancelAllNotifications,
  requestNotificationPermissions,
} from "../services/notifications";
import { saveTimerState, getTimerState } from "../services/database";
import { TimerScreenProps } from "../navigation/types";

export function TimerScreen({ route }: TimerScreenProps) {
  const { rolls, processSteps } = route.params;
  const [currentRoll, setCurrentRoll] = useState(1);
  const [currentStep, setCurrentStep] = useState(0);
  const [timeLeft, setTimeLeft] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [completedSteps, setCompletedSteps] = useState<number[]>([]);
  const [notificationIds, setNotificationIds] = useState<string[]>([]);

  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const soundRef = useRef<Audio.Sound | null>(null);
  const appState = useRef(AppState.currentState);

  // Load notification sound
  useEffect(() => {
    const loadSound = async () => {
      const { sound } = await Audio.Sound.createAsync(
        require("../../assets/sounds/notification.wav"),
        { shouldPlay: false }
      );
      soundRef.current = sound;
    };

    loadSound();

    return () => {
      soundRef.current?.unloadAsync();
    };
  }, []);

  // Load saved state on component mount
  useEffect(() => {
    const loadSavedState = async () => {
      const savedState = await getTimerState();
      if (savedState) {
        setCurrentRoll(savedState.currentRoll);
        setCurrentStep(savedState.currentStep);
        setTimeLeft(savedState.timeLeft);
        setCompletedSteps(savedState.completedSteps);
        // Don't restore running state - user should manually restart
      }
    };

    loadSavedState();
    requestNotificationPermissions();
  }, []);

  // Save state whenever it changes
  useEffect(() => {
    const state: TimerState = {
      currentRoll,
      currentStep,
      timeLeft,
      isRunning,
      isPaused,
      completedSteps,
    };
    saveTimerState(state);
  }, [currentRoll, currentStep, timeLeft, isRunning, isPaused, completedSteps]);

  // Handle app state changes
  useEffect(() => {
    const handleAppStateChange = (nextAppState: any) => {
      appState.current = nextAppState;
      if (nextAppState === "background" && isRunning) {
        // App going to background while timer is running
        console.log("App backgrounded with timer running");
      }
    };

    const subscription = AppState.addEventListener(
      "change",
      handleAppStateChange
    );
    return () => subscription?.remove();
  }, [isRunning]);

  // Timer logic
  useEffect(() => {
    if (isRunning) {
      activateKeepAwakeAsync();
      intervalRef.current = setInterval(() => {
        setTimeLeft((prev) => {
          const newTime = prev - 1;
          // Play sound and haptic when timer finishes
          if (prev === 1) {
            soundRef.current?.replayAsync().catch(() => {});
            Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
          }
          return newTime;
        });
      }, 1000);
    } else {
      deactivateKeepAwake();
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    }

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [isRunning]);

  const timeToSeconds = (timeStr: string): number => {
    const [minutes, seconds] = timeStr.split(":").map(Number);
    return minutes * 60 + seconds;
  };

  const getCurrentStepTime = (): string => {
    if (currentStep === 1) {
      // Developer step
      return rolls[currentRoll - 1].devTime;
    }
    return processSteps[currentStep].time;
  };

  const startTimer = async () => {
    const stepTime = getCurrentStepTime();
    if (stepTime === "varies") {
      Alert.alert(
        "Manual Step",
        'This step requires manual timing. Press "Complete" when finished.'
      );
      return;
    }

    if (!isPaused) {
      setTimeLeft(timeToSeconds(stepTime));
      // Schedule notifications for remaining steps
      const remainingSteps = processSteps.slice(currentStep);
      const currentRollData = rolls[currentRoll - 1];
      const ids = await scheduleStepNotifications(
        remainingSteps,
        currentRollData,
        0
      );
      setNotificationIds(ids);
    }

    setIsRunning(true);
    setIsPaused(false);
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Light);
  };

  const pauseTimer = () => {
    setIsRunning(false);
    setIsPaused(true);
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Medium);
  };

  const stopTimer = async () => {
    setIsRunning(false);
    setIsPaused(false);
    setTimeLeft(0);
    await cancelAllNotifications();
    setNotificationIds([]);
    Haptics.impactAsync(Haptics.ImpactFeedbackStyle.Heavy);
  };

  const selectStep = (stepIndex: number) => {
    stopTimer();
    setCurrentStep(stepIndex);
    setTimeLeft(0);
  };

  const completeStep = () => {
    stopTimer();
    setCompletedSteps([...completedSteps, currentStep]);
    if (currentStep < processSteps.length - 1) {
      setCurrentStep(currentStep + 1);
    }
    setTimeLeft(0);
    Haptics.notificationAsync(Haptics.NotificationFeedbackType.Success);
  };

  const resetSteps = () => {
    stopTimer();
    setCurrentStep(0);
    setCompletedSteps([]);
    setTimeLeft(0);
    setIsPaused(false);
  };

  const nextRoll = () => {
    if (currentRoll < rolls.length) {
      setCurrentRoll(currentRoll + 1);
      resetSteps();
    }
  };

  const prevRoll = () => {
    if (currentRoll > 1) {
      setCurrentRoll(currentRoll - 1);
      resetSteps();
    }
  };

  const currentRollData = rolls[currentRoll - 1];

  return (
    <SafeAreaView style={styles.container}>
      <ScrollView
        style={styles.scrollView}
        showsVerticalScrollIndicator={false}
      >
        {/* Header */}
        <View style={styles.header}>
          <Text style={styles.title}>🎞️ Film Development Timer</Text>
          <Text style={styles.subtitle}>Cinestill Cs41 • 102°F/39°C</Text>
        </View>

        {/* Roll Selector */}
        <View style={styles.rollSelector}>
          <View style={styles.rollControls}>
            <TouchableOpacity
              style={[
                styles.rollButton,
                currentRoll === 1 && styles.disabledButton,
              ]}
              onPress={prevRoll}
              disabled={currentRoll === 1}
            >
              <Text style={styles.rollButtonText}>← Prev</Text>
            </TouchableOpacity>

            <View style={styles.rollInfo}>
              <Text style={styles.rollNumber}>
                Roll {currentRoll}/{rolls.length}
              </Text>
              <Text style={styles.rollDetails}>
                {currentRollData.film} • {currentRollData.pushPull}
              </Text>
            </View>

            <TouchableOpacity
              style={[
                styles.rollButton,
                currentRoll === rolls.length && styles.disabledButton,
              ]}
              onPress={nextRoll}
              disabled={currentRoll === rolls.length}
            >
              <Text style={styles.rollButtonText}>Next →</Text>
            </TouchableOpacity>
          </View>

          {currentRollData.notes ? (
            <View style={styles.notesContainer}>
              <Text style={styles.notesText}>💡 {currentRollData.notes}</Text>
            </View>
          ) : null}

          {/* Developer Mixing Instructions */}
          {currentRoll > 1 && (
            <View style={styles.developerInstructions}>
              <Text style={styles.developerTitle}>
                🧪 Developer Preparation:
              </Text>
              <Text style={styles.developerText}>
                • Use existing working solution from previous roll
              </Text>
              <Text style={styles.developerText}>
                • Development time: {currentRollData.devTime} (compensated for
                chemistry fatigue)
              </Text>
              {currentRoll === 2 && (
                <Text style={styles.developerText}>
                  • After this roll: replenish with 25mL fresh developer
                </Text>
              )}
            </View>
          )}

          {currentRoll === 1 && (
            <View style={styles.developerInstructions}>
              <Text style={styles.developerTitle}>
                🧪 Developer Preparation:
              </Text>
              <Text style={styles.developerText}>
                • Mix fresh Cinestill Cs41 developer solution
              </Text>
              <Text style={styles.developerText}>
                • Temperature: 102°F/39°C
              </Text>
              <Text style={styles.developerText}>
                • Development time: {currentRollData.devTime} (fresh chemistry)
              </Text>
            </View>
          )}
        </View>

        {/* Timer Display - Top Position */}
        <View style={styles.timerContainer}>
          <TimerDisplay timeLeft={timeLeft} isRunning={isRunning} />
        </View>

        {/* Timer Controls - Below Display */}
        <View style={styles.timerControlsContainer}>
          <View style={styles.timerControls}>
            {!isRunning && !isPaused && (
              <TouchableOpacity
                style={[
                  styles.controlButton,
                  styles.startButton,
                  getCurrentStepTime() === "varies" && styles.disabledButton,
                ]}
                onPress={startTimer}
                disabled={getCurrentStepTime() === "varies"}
              >
                <Text style={styles.controlButtonText}>▶️ Start</Text>
              </TouchableOpacity>
            )}

            {isPaused && (
              <TouchableOpacity
                style={[styles.controlButton, styles.startButton]}
                onPress={startTimer}
              >
                <Text style={styles.controlButtonText}>▶️ Resume</Text>
              </TouchableOpacity>
            )}

            {isRunning && (
              <TouchableOpacity
                style={[styles.controlButton, styles.pauseButton]}
                onPress={pauseTimer}
              >
                <Text style={styles.controlButtonText}>⏸️ Pause</Text>
              </TouchableOpacity>
            )}

            <TouchableOpacity
              style={[
                styles.controlButton,
                styles.stopButton,
                !isRunning && !isPaused && styles.disabledButton,
              ]}
              onPress={stopTimer}
              disabled={!isRunning && !isPaused}
            >
              <Text style={styles.controlButtonText}>⏹️ Stop</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={[styles.controlButton, styles.completeButton]}
              onPress={completeStep}
            >
              <Text style={styles.controlButtonText}>✅ Complete</Text>
            </TouchableOpacity>
          </View>
        </View>

        {/* Process Steps */}
        <View style={styles.stepsContainer}>
          {processSteps.map((step: ProcessStep, index: number) => (
            <ProcessStepCard
              key={step.id}
              step={step}
              index={index}
              isActive={index === currentStep}
              isCompleted={completedSteps.includes(index)}
              currentRoll={currentRollData}
              onPress={() => selectStep(index)}
            />
          ))}
        </View>

        {/* Reset Button */}
        <View style={styles.resetContainer}>
          <TouchableOpacity style={styles.resetButton} onPress={resetSteps}>
            <Text style={styles.resetButtonText}>🔄 Reset Roll</Text>
          </TouchableOpacity>
        </View>

        {/* Replenishment Reminder */}
        {currentStep === processSteps.length - 1 && (
          <View style={styles.replenishmentContainer}>
            <Text style={styles.replenishmentTitle}>🧪 After This Roll:</Text>
            <View style={styles.replenishmentSteps}>
              <Text style={styles.replenishmentText}>
                • Pour used 600mL back to WORKING bottle
              </Text>
              <Text style={styles.replenishmentText}>
                • Bleed off 25mL (≈4%) to waste
              </Text>
              <Text style={styles.replenishmentText}>
                • Replace with 25mL fresh from RESERVE
              </Text>
              <Text style={styles.replenishmentText}>
                • Cap both bottles, invert once to mix
              </Text>
            </View>
          </View>
        )}
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#111827",
  },
  scrollView: {
    flex: 1,
    padding: 16,
  },
  header: {
    alignItems: "center",
    marginBottom: 24,
  },
  title: {
    fontSize: 24,
    fontWeight: "bold",
    color: "#FFFFFF",
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: "#9CA3AF",
  },
  rollSelector: {
    backgroundColor: "#1F2937",
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
  },
  rollControls: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 16,
  },
  rollButton: {
    backgroundColor: "#3B82F6",
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
  },
  disabledButton: {
    backgroundColor: "#4B5563",
  },
  rollButtonText: {
    color: "#FFFFFF",
    fontWeight: "600",
  },
  rollInfo: {
    alignItems: "center",
  },
  rollNumber: {
    fontSize: 18,
    fontWeight: "bold",
    color: "#FFFFFF",
  },
  rollDetails: {
    fontSize: 14,
    color: "#9CA3AF",
  },
  notesContainer: {
    backgroundColor: "#FCD34D20",
    padding: 12,
    borderRadius: 8,
    borderLeftWidth: 4,
    borderLeftColor: "#FCD34D",
  },
  notesText: {
    color: "#FCD34D",
    fontSize: 14,
  },
  stepsContainer: {
    marginBottom: 24,
  },
  timerContainer: {
    backgroundColor: "#1F2937",
    borderRadius: 12,
    padding: 24,
    marginBottom: 24,
    alignItems: "center",
  },
  timerControls: {
    flexDirection: "row",
    flexWrap: "wrap",
    justifyContent: "center",
    gap: 8,
  },
  controlButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
    marginHorizontal: 4,
    marginVertical: 4,
  },
  startButton: {
    backgroundColor: "#10B981",
  },
  pauseButton: {
    backgroundColor: "#F59E0B",
  },
  stopButton: {
    backgroundColor: "#EF4444",
  },
  completeButton: {
    backgroundColor: "#3B82F6",
  },
  controlButtonText: {
    color: "#FFFFFF",
    fontWeight: "600",
    fontSize: 14,
  },
  resetContainer: {
    marginBottom: 24,
  },
  resetButton: {
    backgroundColor: "#F59E0B",
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: "center",
  },
  resetButtonText: {
    color: "#FFFFFF",
    fontWeight: "600",
    fontSize: 16,
  },
  replenishmentContainer: {
    backgroundColor: "#92400E20",
    borderWidth: 1,
    borderColor: "#F59E0B",
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
  },
  replenishmentTitle: {
    color: "#F59E0B",
    fontWeight: "bold",
    fontSize: 16,
    marginBottom: 12,
  },
  replenishmentSteps: {
    gap: 4,
  },
  replenishmentText: {
    color: "#FCD34D",
    fontSize: 14,
  },
  timerControlsContainer: {
    backgroundColor: "#1F2937",
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
  },
  checklistContainer: {
    backgroundColor: "#1F2937",
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
  },
  checklistTitle: {
    color: "#FFFFFF",
    fontWeight: "bold",
    fontSize: 16,
    marginBottom: 12,
  },
  checklistSteps: {
    gap: 6,
  },
  checklistText: {
    color: "#E5E7EB",
    fontSize: 14,
    lineHeight: 20,
  },
  developerInstructions: {
    backgroundColor: "#065F4620",
    borderWidth: 1,
    borderColor: "#10B981",
    borderRadius: 8,
    padding: 12,
    marginTop: 12,
  },
  developerTitle: {
    color: "#10B981",
    fontWeight: "bold",
    fontSize: 14,
    marginBottom: 8,
  },
  developerText: {
    color: "#6EE7B7",
    fontSize: 13,
    lineHeight: 18,
    marginBottom: 2,
  },
});
