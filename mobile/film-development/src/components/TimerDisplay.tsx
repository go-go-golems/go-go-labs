import React from 'react';
import { View, Text, StyleSheet } from 'react-native';

interface TimerDisplayProps {
  timeLeft: number;
  isRunning: boolean;
}

export const TimerDisplay: React.FC<TimerDisplayProps> = ({ timeLeft, isRunning }) => {
  const secondsToTime = (seconds: number): string => {
    const isNegative = seconds < 0;
    const absSeconds = Math.abs(seconds);
    const mins = Math.floor(absSeconds / 60);
    const secs = absSeconds % 60;
    const timeStr = `${mins}:${secs.toString().padStart(2, '0')}`;
    return isNegative ? `-${timeStr}` : timeStr;
  };

  return (
    <View style={styles.container}>
      <Text style={[styles.timeText, timeLeft < 0 && styles.overtimeText]}>
        {secondsToTime(timeLeft)}
      </Text>
      {timeLeft < 0 && (
        <Text style={styles.overtimeLabel}>⚠️ OVERTIME</Text>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    alignItems: 'center',
    justifyContent: 'center',
    padding: 20,
  },
  timeText: {
    fontSize: 48,
    fontFamily: 'monospace',
    fontWeight: 'bold',
    color: '#FFFFFF',
    marginBottom: 8,
  },
  overtimeText: {
    color: '#EF4444',
  },
  overtimeLabel: {
    color: '#EF4444',
    fontSize: 14,
    fontWeight: '600',
  },
});