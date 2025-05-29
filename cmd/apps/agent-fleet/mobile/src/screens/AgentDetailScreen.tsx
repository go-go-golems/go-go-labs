import React, { useCallback, useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Alert,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { RouteProp, useRoute } from '@react-navigation/native';
import { NativeStackNavigationProp, NativeStackScreenProps } from '@react-navigation/native-stack';
import {
  useGetAgentEventsQuery,
  useGetAgentTodosQuery,
  useGetAgentCommandsQuery,
  useSendCommandMutation,
  useUpdateTodoMutation,
} from '@/services/api';
import { RootStackParamList } from '@/types/navigation';
import { TodoItem, Event, Command } from '@/types/api';

type AgentDetailScreenProps = NativeStackScreenProps<RootStackParamList, 'AgentDetail'>;

export default function AgentDetailScreen({ route }: AgentDetailScreenProps) {
  const { agent } = route.params;
  const [commandText, setCommandText] = useState('');
  const [showLogs, setShowLogs] = useState(false);

  const {
    data: eventsData,
    isLoading: eventsLoading,
    refetch: refetchEvents,
  } = useGetAgentEventsQuery({ agentId: agent.id, limit: 50 });

  const {
    data: todosData,
    isLoading: todosLoading,
    refetch: refetchTodos,
  } = useGetAgentTodosQuery(agent.id);

  const {
    data: commandsData,
    isLoading: commandsLoading,
    refetch: refetchCommands,
  } = useGetAgentCommandsQuery({ agentId: agent.id, limit: 20 });

  const [sendCommand, { isLoading: sendingCommand }] = useSendCommandMutation();
  const [updateTodo] = useUpdateTodoMutation();

  const handleSendCommand = useCallback(async () => {
    if (!commandText.trim()) {
      Alert.alert('Error', 'Please enter a command');
      return;
    }

    try {
      const commandType = agent.pending_question ? 'feedback' : 'instruction';
      
      await sendCommand({
        agentId: agent.id,
        content: commandText.trim(),
        type: commandType,
      }).unwrap();

      setCommandText('');
      Alert.alert('Success', 'Command sent successfully');
      refetchCommands();
    } catch (error) {
      Alert.alert('Error', 'Failed to send command');
    }
  }, [commandText, agent.id, agent.pending_question, sendCommand, refetchCommands]);

  const handleToggleTodo = useCallback(async (todo: TodoItem) => {
    try {
      await updateTodo({
        agentId: agent.id,
        todoId: todo.id,
        updates: { completed: !todo.completed },
      }).unwrap();
      refetchTodos();
    } catch (error) {
      Alert.alert('Error', 'Failed to update todo');
    }
  }, [agent.id, updateTodo, refetchTodos]);

  const getStatusColor = (status: typeof agent.status) => {
    switch (status) {
      case 'active':
        return '#10B981';
      case 'idle':
        return '#6B7280';
      case 'waiting_feedback':
        return '#F59E0B';
      case 'error':
        return '#EF4444';
      default:
        return '#6B7280';
    }
  };

  const getCommandInterface = () => {
    const isFeedbackMode = agent.pending_question;
    const isDisabled = agent.status === 'idle';
    
    return (
      <View style={styles.commandContainer}>
        <Text style={styles.sectionTitle}>
          {isFeedbackMode ? '💬 Respond to Question' : '⚡ Send Command'}
        </Text>
        
        {isFeedbackMode && (
          <View style={styles.questionContainer}>
            <Text style={styles.questionText}>❓ {agent.pending_question}</Text>
          </View>
        )}
        
        <TextInput
          style={[
            styles.commandInput,
            {
              borderColor: isFeedbackMode ? '#F59E0B' : isDisabled ? '#6B7280' : '#3B82F6',
              backgroundColor: isDisabled ? '#1F2937' : '#111827',
            }
          ]}
          placeholder={
            isDisabled 
              ? 'Agent is idle' 
              : isFeedbackMode 
                ? 'Provide feedback...' 
                : 'Enter command...'
          }
          placeholderTextColor="#6B7280"
          value={commandText}
          onChangeText={setCommandText}
          multiline
          numberOfLines={3}
          textAlignVertical="top"
          editable={!isDisabled}
        />
        
        <TouchableOpacity
          style={[
            styles.sendButton,
            {
              backgroundColor: isFeedbackMode ? '#F59E0B' : '#3B82F6',
              opacity: (sendingCommand || isDisabled) ? 0.6 : 1,
            }
          ]}
          onPress={handleSendCommand}
          disabled={sendingCommand || isDisabled}
        >
          {sendingCommand ? (
            <ActivityIndicator size="small" color="#ffffff" />
          ) : (
            <Text style={styles.sendButtonText}>
              {isFeedbackMode ? '💬 Send Feedback' : '⚡ Send Command'}
            </Text>
          )}
        </TouchableOpacity>
      </View>
    );
  };

  const renderStatsGrid = () => (
    <View style={styles.statsGrid}>
      <View style={styles.statItem}>
        <Text style={styles.statValue}>{agent.files_changed}</Text>
        <Text style={styles.statLabel}>Files Changed</Text>
      </View>
      <View style={styles.statItem}>
        <Text style={[styles.statValue, { color: '#10B981' }]}>+{agent.lines_added}</Text>
        <Text style={styles.statLabel}>Lines Added</Text>
      </View>
      <View style={styles.statItem}>
        <Text style={[styles.statValue, { color: '#EF4444' }]}>-{agent.lines_removed}</Text>
        <Text style={styles.statLabel}>Lines Removed</Text>
      </View>
      <View style={styles.statItem}>
        <Text style={styles.statValue}>{agent.progress}%</Text>
        <Text style={styles.statLabel}>Progress</Text>
      </View>
    </View>
  );

  const renderTodoList = () => {
    if (todosLoading) {
      return <ActivityIndicator size="small" color="#3B82F6" />;
    }

    const todos = todosData?.todos || [];

    return (
      <View style={styles.todoContainer}>
        <Text style={styles.sectionTitle}>📋 Todo List</Text>
        {todos.length === 0 ? (
          <Text style={styles.emptyText}>No todos available</Text>
        ) : (
          todos.map((todo) => (
            <TouchableOpacity
              key={todo.id}
              style={styles.todoItem}
              onPress={() => handleToggleTodo(todo)}
            >
              <Text style={styles.todoIcon}>
                {todo.completed ? '✓' : todo.current ? '●' : '○'}
              </Text>
              <Text
                style={[
                  styles.todoText,
                  {
                    textDecorationLine: todo.completed ? 'line-through' : 'none',
                    color: todo.completed ? '#6B7280' : todo.current ? '#ffffff' : '#D1D5DB',
                    backgroundColor: todo.current ? '#1F2937' : 'transparent',
                  }
                ]}
              >
                {todo.text}
              </Text>
            </TouchableOpacity>
          ))
        )}
      </View>
    );
  };

  const renderEventHistory = () => {
    if (eventsLoading) {
      return <ActivityIndicator size="small" color="#3B82F6" />;
    }

    const events = eventsData?.events || [];

    return (
      <View style={styles.eventsContainer}>
        <Text style={styles.sectionTitle}>📝 Event History</Text>
        {events.length === 0 ? (
          <Text style={styles.emptyText}>No events available</Text>
        ) : (
          events.slice(0, 10).map((event) => (
            <View key={event.id} style={styles.eventItem}>
              <Text style={styles.eventTime}>
                {new Date(event.timestamp).toLocaleTimeString()}
              </Text>
              <Text style={styles.eventType}>{event.type}</Text>
              <Text style={styles.eventMessage} numberOfLines={2}>
                {event.message}
              </Text>
            </View>
          ))
        )}
      </View>
    );
  };

  const renderDebugLogs = () => {
    if (!showLogs) {
      return (
        <TouchableOpacity
          style={styles.toggleLogsButton}
          onPress={() => setShowLogs(true)}
        >
          <Text style={styles.toggleLogsText}>🔍 Show Debug Logs</Text>
        </TouchableOpacity>
      );
    }

    if (commandsLoading) {
      return <ActivityIndicator size="small" color="#3B82F6" />;
    }

    const commands = commandsData?.commands || [];

    return (
      <View style={styles.logsContainer}>
        <View style={styles.logsHeader}>
          <Text style={styles.sectionTitle}>🔍 Debug Logs</Text>
          <TouchableOpacity onPress={() => setShowLogs(false)}>
            <Text style={styles.hideLogsText}>Hide</Text>
          </TouchableOpacity>
        </View>
        {commands.length === 0 ? (
          <Text style={styles.emptyText}>No commands available</Text>
        ) : (
          commands.map((command) => (
            <View key={command.id} style={styles.logItem}>
              <Text style={styles.logTime}>
                {new Date(command.sent_at).toLocaleTimeString()}
              </Text>
              <Text style={styles.logType}>{command.type}</Text>
              <Text style={styles.logContent}>{command.content}</Text>
              {command.response && (
                <Text style={styles.logResponse}>→ {command.response}</Text>
              )}
            </View>
          ))
        )}
      </View>
    );
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
        {/* Header */}
        <View style={styles.header}>
          <View style={styles.statusRow}>
            <Text style={[styles.statusIndicator, { color: getStatusColor(agent.status) }]}>
              ●
            </Text>
            <Text style={styles.agentName}>{agent.name}</Text>
            <View style={[styles.statusBadge, { backgroundColor: getStatusColor(agent.status) }]}>
              <Text style={styles.statusText}>{agent.status}</Text>
            </View>
          </View>
          <Text style={styles.worktree}>🌿 {agent.worktree}</Text>
        </View>

        {/* Command Interface */}
        {getCommandInterface()}

        {/* Current Task */}
        <View style={styles.taskContainer}>
          <Text style={styles.sectionTitle}>🎯 Current Task</Text>
          <Text style={styles.taskText}>
            {agent.current_task || 'No current task assigned'}
          </Text>
        </View>

        {/* Stats Grid */}
        {renderStatsGrid()}

        {/* Todo List */}
        {renderTodoList()}

        {/* Event History */}
        {renderEventHistory()}

        {/* Debug Logs */}
        {renderDebugLogs()}
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000000',
  },
  scrollView: {
    flex: 1,
  },
  header: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  statusRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  statusIndicator: {
    fontSize: 20,
    marginRight: 8,
  },
  agentName: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: 'bold',
    flex: 1,
  },
  statusBadge: {
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 16,
  },
  statusText: {
    color: '#ffffff',
    fontSize: 12,
    fontWeight: '600',
  },
  worktree: {
    color: '#9CA3AF',
    fontSize: 14,
  },
  commandContainer: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  questionContainer: {
    backgroundColor: '#FEF3C7',
    padding: 12,
    borderRadius: 8,
    marginBottom: 12,
  },
  questionText: {
    color: '#92400E',
    fontSize: 14,
  },
  commandInput: {
    borderWidth: 1,
    borderRadius: 8,
    padding: 12,
    color: '#ffffff',
    fontSize: 14,
    marginBottom: 12,
    minHeight: 80,
  },
  sendButton: {
    paddingVertical: 12,
    borderRadius: 8,
    alignItems: 'center',
  },
  sendButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
  },
  sectionTitle: {
    color: '#ffffff',
    fontSize: 18,
    fontWeight: 'bold',
    marginBottom: 12,
  },
  taskContainer: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  taskText: {
    color: '#D1D5DB',
    fontSize: 14,
    lineHeight: 20,
  },
  statsGrid: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  statItem: {
    width: '50%',
    alignItems: 'center',
    marginBottom: 16,
  },
  statValue: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: 'bold',
  },
  statLabel: {
    color: '#9CA3AF',
    fontSize: 12,
    marginTop: 4,
  },
  todoContainer: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  todoItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
    paddingHorizontal: 4,
    borderRadius: 4,
    marginBottom: 4,
  },
  todoIcon: {
    fontSize: 16,
    marginRight: 12,
    color: '#9CA3AF',
  },
  todoText: {
    flex: 1,
    fontSize: 14,
    paddingVertical: 4,
    paddingHorizontal: 8,
    borderRadius: 4,
  },
  eventsContainer: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  eventItem: {
    backgroundColor: '#111827',
    padding: 12,
    borderRadius: 8,
    marginBottom: 8,
  },
  eventTime: {
    color: '#9CA3AF',
    fontSize: 12,
    marginBottom: 4,
  },
  eventType: {
    color: '#3B82F6',
    fontSize: 12,
    fontWeight: '600',
    marginBottom: 4,
  },
  eventMessage: {
    color: '#D1D5DB',
    fontSize: 14,
  },
  toggleLogsButton: {
    margin: 16,
    padding: 12,
    backgroundColor: '#111827',
    borderRadius: 8,
    alignItems: 'center',
  },
  toggleLogsText: {
    color: '#3B82F6',
    fontSize: 16,
    fontWeight: '600',
  },
  logsContainer: {
    padding: 16,
  },
  logsHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  hideLogsText: {
    color: '#3B82F6',
    fontSize: 14,
  },
  logItem: {
    backgroundColor: '#0F1419',
    padding: 12,
    borderRadius: 8,
    marginBottom: 8,
    borderLeftWidth: 2,
    borderLeftColor: '#374151',
  },
  logTime: {
    color: '#6B7280',
    fontSize: 11,
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
    marginBottom: 4,
  },
  logType: {
    color: '#F59E0B',
    fontSize: 11,
    fontWeight: '600',
    marginBottom: 4,
  },
  logContent: {
    color: '#D1D5DB',
    fontSize: 12,
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
    marginBottom: 4,
  },
  logResponse: {
    color: '#10B981',
    fontSize: 12,
    fontFamily: Platform.OS === 'ios' ? 'Menlo' : 'monospace',
  },
  emptyText: {
    color: '#6B7280',
    fontSize: 14,
    fontStyle: 'italic',
    textAlign: 'center',
    paddingVertical: 16,
  },
});
