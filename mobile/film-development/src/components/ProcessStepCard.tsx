import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { ProcessStep, Roll } from '../types';

interface ProcessStepCardProps {
  step: ProcessStep;
  index: number;
  isActive: boolean;
  isCompleted: boolean;
  currentRoll: Roll;
  onPress: () => void;
}

export const ProcessStepCard: React.FC<ProcessStepCardProps> = ({
  step,
  index,
  isActive,
  isCompleted,
  currentRoll,
  onPress,
}) => {
  const getStepTime = (): string => {
    if (step.id === 'developer') {
      return currentRoll.devTime;
    }
    return step.time;
  };

  const getStepIcon = (): string => {
    if (isCompleted) return '✅';
    if (isActive) return '⏱️';
    return '⏸️';
  };

  return (
    <TouchableOpacity 
      style={[
        styles.container,
        isActive && styles.activeContainer,
        isCompleted && styles.completedContainer,
      ]} 
      onPress={onPress}
    >
      <View style={styles.header}>
        <View style={styles.titleRow}>
          <Text style={styles.icon}>{getStepIcon()}</Text>
          <Text style={styles.stepName}>{step.name}</Text>
        </View>
        <Text style={styles.timeText}>{getStepTime()}</Text>
      </View>
      <Text style={styles.description}>{step.description}</Text>
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#1F2937',
    borderRadius: 12,
    borderWidth: 2,
    borderColor: '#4B5563',
    padding: 16,
    marginVertical: 6,
  },
  activeContainer: {
    borderColor: '#3B82F6',
    backgroundColor: '#1E3A8A20',
  },
  completedContainer: {
    borderColor: '#10B981',
    backgroundColor: '#064E3B20',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  icon: {
    fontSize: 18,
    marginRight: 8,
  },
  stepName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFFFFF',
    flex: 1,
  },
  timeText: {
    fontSize: 18,
    fontFamily: 'monospace',
    color: '#FFFFFF',
    fontWeight: 'bold',
  },
  description: {
    fontSize: 14,
    color: '#9CA3AF',
    lineHeight: 20,
  },
});