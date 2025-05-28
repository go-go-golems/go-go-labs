import React, { useCallback, useState, useMemo } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  FlatList,
  RefreshControl,
  StyleSheet,
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useGetTasksQuery, useCreateTaskMutation, useGetAgentsQuery } from '@/services/api';
import { Task, TaskPriority } from '@/types/api';
import TaskItem from '@/components/TaskItem';

export default function TasksScreen() {
  const [taskTitle, setTaskTitle] = useState('');
  const [taskDescription, setTaskDescription] = useState('');
  const [selectedPriority, setSelectedPriority] = useState<TaskPriority>('medium');

  const {
    data: tasksData,
    isLoading: tasksLoading,
    error: tasksError,
    refetch: refetchTasks,
  } = useGetTasksQuery({ limit: 50 });

  const {
    data: agentsData,
  } = useGetAgentsQuery({ limit: 100 });

  const [createTask, { isLoading: createLoading }] = useCreateTaskMutation();

  const agentNamesMap = useMemo(() => {
    if (!agentsData?.agents) return {};
    return agentsData.agents.reduce((acc, agent) => {
      acc[agent.id] = agent.name;
      return acc;
    }, {} as Record<string, string>);
  }, [agentsData]);

  const handleRefresh = useCallback(() => {
    refetchTasks();
  }, [refetchTasks]);

  const handleSubmitTask = useCallback(async () => {
    if (!taskTitle.trim() || !taskDescription.trim()) {
      Alert.alert('Error', 'Please provide both title and description');
      return;
    }

    try {
      await createTask({
        title: taskTitle.trim(),
        description: taskDescription.trim(),
        priority: selectedPriority,
      }).unwrap();

      setTaskTitle('');
      setTaskDescription('');
      setSelectedPriority('medium');
      
      Alert.alert('Success', 'Task created successfully');
      refetchTasks();
    } catch (error) {
      Alert.alert('Error', 'Failed to create task');
    }
  }, [taskTitle, taskDescription, selectedPriority, createTask, refetchTasks]);

  const renderTaskItem = useCallback(({ item }: { item: Task }) => (
    <TaskItem
      task={item}
      agentName={item.assigned_agent_id ? agentNamesMap[item.assigned_agent_id] : undefined}
    />
  ), [agentNamesMap]);

  const renderPriorityButton = (priority: TaskPriority, label: string, color: string) => (
    <TouchableOpacity
      key={priority}
      style={[
        styles.priorityButton,
        {
          backgroundColor: selectedPriority === priority ? color : 'transparent',
          borderColor: color,
        }
      ]}
      onPress={() => setSelectedPriority(priority)}
    >
      <Text
        style={[
          styles.priorityButtonText,
          {
            color: selectedPriority === priority ? '#ffffff' : color,
          }
        ]}
      >
        {label}
      </Text>
    </TouchableOpacity>
  );

  const renderTaskInput = () => (
    <View style={styles.inputContainer}>
      <Text style={styles.inputTitle}>Submit Task</Text>
      
      <TextInput
        style={styles.titleInput}
        placeholder="Task title..."
        placeholderTextColor="#6B7280"
        value={taskTitle}
        onChangeText={setTaskTitle}
        multiline={false}
        maxLength={100}
      />

      <TextInput
        style={styles.descriptionInput}
        placeholder="Describe the task..."
        placeholderTextColor="#6B7280"
        value={taskDescription}
        onChangeText={setTaskDescription}
        multiline
        numberOfLines={4}
        textAlignVertical="top"
        maxLength={500}
      />

      <View style={styles.priorityContainer}>
        <Text style={styles.priorityLabel}>Priority:</Text>
        <View style={styles.priorityButtons}>
          {renderPriorityButton('low', 'Low', '#6B7280')}
          {renderPriorityButton('medium', 'Medium', '#3B82F6')}
          {renderPriorityButton('high', 'High', '#F59E0B')}
          {renderPriorityButton('urgent', 'Urgent', '#EF4444')}
        </View>
      </View>

      <View style={styles.submitRow}>
        <Text style={styles.submitHint}>
          Task will be distributed to available agents
        </Text>
        <TouchableOpacity
          style={[
            styles.submitButton,
            { opacity: createLoading ? 0.6 : 1 }
          ]}
          onPress={handleSubmitTask}
          disabled={createLoading}
        >
          {createLoading ? (
            <ActivityIndicator size="small" color="#ffffff" />
          ) : (
            <Text style={styles.submitButtonText}>🚀 Submit</Text>
          )}
        </TouchableOpacity>
      </View>
    </View>
  );

  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Text style={styles.emptyStateText}>📋</Text>
      <Text style={styles.emptyStateTitle}>No Tasks</Text>
      <Text style={styles.emptyStateSubtitle}>
        Create your first task using the form above
      </Text>
    </View>
  );

  const renderError = () => (
    <View style={styles.errorState}>
      <Text style={styles.errorText}>❌ Failed to load tasks</Text>
      <TouchableOpacity style={styles.retryButton} onPress={handleRefresh}>
        <Text style={styles.retryButtonText}>Retry</Text>
      </TouchableOpacity>
    </View>
  );

  const renderTaskQueue = () => {
    if (tasksLoading && !tasksData) {
      return (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#3B82F6" />
          <Text style={styles.loadingText}>Loading tasks...</Text>
        </View>
      );
    }

    if (tasksError) {
      return renderError();
    }

    const tasks = tasksData?.tasks || [];

    return (
      <View style={styles.queueContainer}>
        <Text style={styles.queueTitle}>Task Queue ({tasks.length})</Text>
        {tasks.length === 0 ? (
          renderEmptyState()
        ) : (
          <FlatList
            data={tasks}
            renderItem={renderTaskItem}
            keyExtractor={(item) => item.id}
            refreshControl={
              <RefreshControl
                refreshing={tasksLoading}
                onRefresh={handleRefresh}
                tintColor="#3B82F6"
                colors={['#3B82F6']}
              />
            }
            showsVerticalScrollIndicator={false}
            style={styles.taskList}
          />
        )}
      </View>
    );
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      keyboardVerticalOffset={Platform.OS === 'ios' ? 90 : 0}
    >
      <View style={styles.content}>
        {renderTaskInput()}
        {renderTaskQueue()}
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000000',
  },
  content: {
    flex: 1,
  },
  inputContainer: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  inputTitle: {
    color: '#ffffff',
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  titleInput: {
    backgroundColor: '#111827',
    color: '#ffffff',
    borderRadius: 8,
    padding: 12,
    fontSize: 16,
    borderWidth: 1,
    borderColor: '#374151',
    marginBottom: 12,
  },
  descriptionInput: {
    backgroundColor: '#111827',
    color: '#ffffff',
    borderRadius: 8,
    padding: 12,
    fontSize: 14,
    borderWidth: 1,
    borderColor: '#374151',
    height: 100,
    marginBottom: 16,
  },
  priorityContainer: {
    marginBottom: 16,
  },
  priorityLabel: {
    color: '#D1D5DB',
    fontSize: 14,
    fontWeight: '500',
    marginBottom: 8,
  },
  priorityButtons: {
    flexDirection: 'row',
    gap: 8,
  },
  priorityButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 16,
    borderWidth: 1,
  },
  priorityButtonText: {
    fontSize: 12,
    fontWeight: '500',
  },
  submitRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  submitHint: {
    color: '#9CA3AF',
    fontSize: 12,
    flex: 1,
  },
  submitButton: {
    backgroundColor: '#3B82F6',
    paddingVertical: 10,
    paddingHorizontal: 20,
    borderRadius: 8,
    minWidth: 100,
    alignItems: 'center',
  },
  submitButtonText: {
    color: '#ffffff',
    fontSize: 14,
    fontWeight: '600',
  },
  queueContainer: {
    flex: 1,
    padding: 16,
  },
  queueTitle: {
    color: '#ffffff',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 16,
  },
  taskList: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    color: '#9CA3AF',
    fontSize: 16,
    marginTop: 16,
  },
  emptyState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
  },
  emptyStateText: {
    fontSize: 64,
    marginBottom: 16,
  },
  emptyStateTitle: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 8,
  },
  emptyStateSubtitle: {
    color: '#9CA3AF',
    fontSize: 16,
    textAlign: 'center',
  },
  errorState: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 32,
  },
  errorText: {
    color: '#EF4444',
    fontSize: 18,
    textAlign: 'center',
    marginBottom: 16,
  },
  retryButton: {
    backgroundColor: '#3B82F6',
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
  },
  retryButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
  },
});
