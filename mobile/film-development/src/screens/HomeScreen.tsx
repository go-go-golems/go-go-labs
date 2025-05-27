import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';

import { Roll, ProcessStep } from '../types';
import { getRolls, getProcessSteps, initializeDatabase } from '../services/database';
import { HomeScreenProps } from '../navigation/types';

export function HomeScreen({ navigation }: HomeScreenProps) {
  const [rolls, setRolls] = useState<Roll[]>([]);
  const [processSteps, setProcessSteps] = useState<ProcessStep[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        await initializeDatabase();
        const [rollsData, stepsData] = await Promise.all([
          getRolls(),
          getProcessSteps(),
        ]);
        setRolls(rollsData);
        setProcessSteps(stepsData);
      } catch (error) {
        console.error('Error loading data:', error);
      } finally {
        setIsLoading(false);
      }
    };

    loadData();
  }, []);

  const navigateToTimer = () => {
    navigation.navigate('Timer', { rolls, processSteps });
  };

  if (isLoading) {
    return (
      <SafeAreaView style={styles.container}>
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#3B82F6" />
          <Text style={styles.loadingText}>Loading...</Text>
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={styles.container}>
      <View style={styles.content}>
        <View style={styles.header}>
          <Text style={styles.title}>🎞️ Film Development</Text>
          <Text style={styles.subtitle}>Cinestill Cs41 Timer</Text>
        </View>

        <View style={styles.infoContainer}>
          <Text style={styles.infoText}>
            This app helps you time your film development process with precise timing for each step.
          </Text>
          <Text style={styles.infoText}>
            • {rolls.length} rolls configured
          </Text>
          <Text style={styles.infoText}>
            • {processSteps.length} development steps
          </Text>
          <Text style={styles.infoText}>
            • Background notifications when steps complete
          </Text>
        </View>

        <View style={styles.processOverview}>
          <Text style={styles.processTitle}>Development Process:</Text>
          {processSteps.map((step, index) => (
            <View key={step.id} style={styles.processStep}>
              <Text style={styles.processStepNumber}>{index + 1}.</Text>
              <View style={styles.processStepContent}>
                <Text style={styles.processStepName}>{step.name}</Text>
                <Text style={styles.processStepTime}>{step.time}</Text>
              </View>
            </View>
          ))}
        </View>

        <TouchableOpacity style={styles.startButton} onPress={navigateToTimer}>
          <Text style={styles.startButtonText}>Start Timer</Text>
        </TouchableOpacity>

        <View style={styles.footer}>
          <Text style={styles.footerText}>
            Keep your phone nearby during development for timer notifications
          </Text>
        </View>
      </View>
    </SafeAreaView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#111827',
  },
  content: {
    flex: 1,
    padding: 24,
    justifyContent: 'space-between',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    color: '#FFFFFF',
    marginTop: 16,
    fontSize: 16,
  },
  header: {
    alignItems: 'center',
    marginBottom: 32,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#FFFFFF',
    marginBottom: 8,
    textAlign: 'center',
  },
  subtitle: {
    fontSize: 18,
    color: '#9CA3AF',
    textAlign: 'center',
  },
  infoContainer: {
    backgroundColor: '#1F2937',
    borderRadius: 12,
    padding: 20,
    marginBottom: 24,
  },
  infoText: {
    color: '#E5E7EB',
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 8,
  },
  processOverview: {
    backgroundColor: '#1F2937',
    borderRadius: 12,
    padding: 20,
    marginBottom: 32,
  },
  processTitle: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: '600',
    marginBottom: 16,
  },
  processStep: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  processStepNumber: {
    color: '#3B82F6',
    fontSize: 16,
    fontWeight: 'bold',
    width: 24,
  },
  processStepContent: {
    flex: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  processStepName: {
    color: '#E5E7EB',
    fontSize: 16,
    flex: 1,
  },
  processStepTime: {
    color: '#9CA3AF',
    fontSize: 14,
    fontFamily: 'monospace',
  },
  startButton: {
    backgroundColor: '#3B82F6',
    borderRadius: 12,
    paddingVertical: 16,
    paddingHorizontal: 32,
    alignItems: 'center',
    marginBottom: 24,
  },
  startButtonText: {
    color: '#FFFFFF',
    fontSize: 18,
    fontWeight: 'bold',
  },
  footer: {
    alignItems: 'center',
  },
  footerText: {
    color: '#6B7280',
    fontSize: 14,
    textAlign: 'center',
    lineHeight: 20,
  },
});