import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Task } from '@/types/api';

interface TaskItemProps {
  task: Task;
  agentName?: string;
}

export default function TaskItem({ task, agentName }: TaskItemProps) {
  const getStatusColor = (status: Task['status']) => {
    switch (status) {
      case 'completed':
        return '#10B981'; // green
      case 'failed':
        return '#EF4444'; // red
      case 'in_progress':
        return '#3B82F6'; // blue
      case 'assigned':
        return '#F59E0B'; // yellow
      case 'pending':
      default:
        return '#6B7280'; // gray
    }
  };

  const getPriorityColor = (priority: Task['priority']) => {
    switch (priority) {
      case 'urgent':
        return '#EF4444'; // red
      case 'high':
        return '#F59E0B'; // orange
      case 'medium':
        return '#3B82F6'; // blue
      case 'low':
      default:
        return '#6B7280'; // gray
    }
  };

  const getStatusIcon = (status: Task['status']) => {
    switch (status) {
      case 'completed':
        return '✅';
      case 'failed':
        return '❌';
      case 'in_progress':
        return '🔄';
      case 'assigned':
        return '👤';
      case 'pending':
      default:
        return '⏳';
    }
  };

  const formatTimeAgo = (timestamp: string) => {
    const now = new Date();
    const time = new Date(timestamp);
    const diffMs = now.getTime() - time.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    
    if (diffMins < 1) return 'now';
    if (diffMins < 60) return `${diffMins}m ago`;
    const diffHours = Math.floor(diffMins / 60);
    if (diffHours < 24) return `${diffHours}h ago`;
    const diffDays = Math.floor(diffHours / 24);
    return `${diffDays}d ago`;
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <View style={styles.titleRow}>
          <Text style={styles.icon}>
            {getStatusIcon(task.status)}
          </Text>
          <Text style={styles.title} numberOfLines={1}>
            {task.title}
          </Text>
          <View style={[styles.priorityBadge, { backgroundColor: getPriorityColor(task.priority) }]}>
            <Text style={styles.priorityText}>{task.priority}</Text>
          </View>
        </View>
        <View style={styles.statusRow}>
          <View style={[styles.statusBadge, { backgroundColor: getStatusColor(task.status) }]}>
            <Text style={styles.statusText}>{task.status}</Text>
          </View>
          <Text style={styles.timestamp}>
            {formatTimeAgo(task.created_at)}
          </Text>
        </View>
      </View>

      <Text style={styles.description} numberOfLines={3}>
        {task.description}
      </Text>

      {task.assigned_agent_id && (
        <View style={styles.assignmentRow}>
          <Text style={styles.assignmentText}>
            👤 Assigned to: {agentName || `Agent ${task.assigned_agent_id.slice(0, 8)}`}
          </Text>
          {task.assigned_at && (
            <Text style={styles.assignmentTime}>
              {formatTimeAgo(task.assigned_at)}
            </Text>
          )}
        </View>
      )}

      {task.completed_at && (
        <View style={styles.completionRow}>
          <Text style={styles.completionText}>
            ✅ Completed: {formatTimeAgo(task.completed_at)}
          </Text>
        </View>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#111827',
    padding: 16,
    marginHorizontal: 16,
    marginVertical: 4,
    borderRadius: 8,
    borderLeftWidth: 3,
    borderLeftColor: '#374151',
  },
  header: {
    marginBottom: 12,
  },
  titleRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  icon: {
    fontSize: 16,
    marginRight: 8,
  },
  title: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
    flex: 1,
  },
  priorityBadge: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 12,
    marginLeft: 8,
  },
  priorityText: {
    color: '#ffffff',
    fontSize: 10,
    fontWeight: '500',
    textTransform: 'uppercase',
  },
  statusRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 6,
  },
  statusText: {
    color: '#ffffff',
    fontSize: 12,
    fontWeight: '500',
  },
  timestamp: {
    color: '#9CA3AF',
    fontSize: 12,
  },
  description: {
    color: '#D1D5DB',
    fontSize: 14,
    lineHeight: 20,
    marginBottom: 12,
  },
  assignmentRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#374151',
  },
  assignmentText: {
    color: '#9CA3AF',
    fontSize: 12,
    flex: 1,
  },
  assignmentTime: {
    color: '#6B7280',
    fontSize: 12,
  },
  completionRow: {
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#374151',
  },
  completionText: {
    color: '#10B981',
    fontSize: 12,
    fontWeight: '500',
  },
});
