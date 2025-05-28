import React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  Animated,
} from 'react-native';
import { Agent } from '@/types/api';

interface AgentCardProps {
  agent: Agent;
  onPress: () => void;
}

export default function AgentCard({ agent, onPress }: AgentCardProps) {
  const getStatusColor = (status: Agent['status']) => {
    switch (status) {
      case 'active':
        return '#10B981'; // green
      case 'idle':
        return '#6B7280'; // gray
      case 'waiting_feedback':
        return '#F59E0B'; // orange
      case 'error':
        return '#EF4444'; // red
      default:
        return '#6B7280';
    }
  };

  const getStatusIndicator = (status: Agent['status']) => {
    switch (status) {
      case 'active':
        return '●';
      case 'idle':
        return '○';
      case 'waiting_feedback':
        return '⚠️';
      case 'error':
        return '❌';
      default:
        return '○';
    }
  };

  const getBorderColor = () => {
    if (agent.status === 'waiting_feedback') {
      return '#F59E0B'; // orange for feedback needed
    }
    return '#374151'; // default gray
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
    <TouchableOpacity
      style={[
        styles.card,
        {
          borderColor: getBorderColor(),
          borderWidth: agent.status === 'waiting_feedback' ? 2 : 1,
        },
      ]}
      onPress={onPress}
      activeOpacity={0.7}
    >
      {/* Header */}
      <View style={styles.header}>
        <View style={styles.statusRow}>
          <Text style={[styles.statusIndicator, { color: getStatusColor(agent.status) }]}>
            {getStatusIndicator(agent.status)}
          </Text>
          <Text style={styles.agentName}>{agent.name}</Text>
          <View style={[styles.statusBadge, { backgroundColor: getStatusColor(agent.status) }]}>
            <Text style={styles.statusText}>{agent.status}</Text>
          </View>
        </View>
      </View>

      {/* Current Task */}
      <View style={styles.taskSection}>
        <Text style={styles.taskText} numberOfLines={2}>
          {agent.current_task || 'No current task'}
        </Text>
      </View>

      {/* Question Box if pending */}
      {agent.pending_question && (
        <View style={styles.questionBox}>
          <Text style={styles.questionText} numberOfLines={2}>
            ❓ {agent.pending_question}
          </Text>
        </View>
      )}

      {/* Metadata Row */}
      <View style={styles.metadataRow}>
        <Text style={styles.metadataText}>🌿 {agent.worktree}</Text>
        <Text style={styles.metadataText}>📝 {formatTimeAgo(agent.last_commit)}</Text>
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
    backgroundColor: '#111827',
    borderRadius: 8,
    padding: 16,
    marginHorizontal: 16,
    marginVertical: 8,
  },
  header: {
    height: 48,
    justifyContent: 'center',
  },
  statusRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  statusIndicator: {
    fontSize: 16,
    marginRight: 8,
  },
  agentName: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
    flex: 1,
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 12,
  },
  statusText: {
    color: '#ffffff',
    fontSize: 12,
    fontWeight: '500',
  },
  taskSection: {
    height: 32,
    justifyContent: 'center',
    borderTopWidth: 1,
    borderTopColor: '#374151',
    paddingTop: 8,
  },
  taskText: {
    color: '#D1D5DB',
    fontSize: 14,
  },
  questionBox: {
    backgroundColor: '#FEF3C7',
    borderRadius: 6,
    padding: 8,
    marginTop: 8,
  },
  questionText: {
    color: '#92400E',
    fontSize: 14,
  },
  metadataRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    height: 24,
    alignItems: 'center',
    borderTopWidth: 1,
    borderTopColor: '#374151',
    paddingTop: 8,
  },
  statsRow: {
    flexDirection: 'row',
    alignItems: 'center',
    height: 24,
    gap: 16,
  },
  metadataText: {
    color: '#9CA3AF',
    fontSize: 12,
  },
  additionsText: {
    color: '#10B981',
    fontSize: 12,
    fontWeight: '500',
  },
  deletionsText: {
    color: '#EF4444',
    fontSize: 12,
    fontWeight: '500',
  },
  progressContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    height: 16,
    marginTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#374151',
    paddingTop: 8,
  },
  progressBar: {
    flex: 1,
    height: 4,
    backgroundColor: '#374151',
    borderRadius: 2,
    marginRight: 8,
  },
  progressFill: {
    height: '100%',
    borderRadius: 2,
  },
  progressText: {
    color: '#9CA3AF',
    fontSize: 12,
    fontWeight: '500',
    minWidth: 32,
  },
});
