import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { Event } from '@/types/api';

interface UpdateItemProps {
  event: Event;
  agentName?: string;
}

export default function UpdateItem({ event, agentName }: UpdateItemProps) {
  const getEventColor = (type: Event['type']) => {
    switch (type) {
      case 'success':
      case 'commit':
        return '#10B981'; // green
      case 'error':
        return '#EF4444'; // red
      case 'question':
        return '#F59E0B'; // yellow
      case 'info':
      case 'start':
      case 'command':
      default:
        return '#3B82F6'; // blue
    }
  };

  const getEventIcon = (type: Event['type']) => {
    switch (type) {
      case 'success':
        return '✅';
      case 'error':
        return '❌';
      case 'question':
        return '❓';
      case 'commit':
        return '📝';
      case 'start':
        return '🚀';
      case 'command':
        return '💬';
      case 'info':
      default:
        return 'ℹ️';
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
        <Text style={styles.agentName}>
          {agentName || `Agent ${event.agent_id.slice(0, 8)}`}
        </Text>
        <Text style={styles.timestamp}>
          {formatTimeAgo(event.timestamp)}
        </Text>
      </View>
      <View style={styles.content}>
        <Text style={styles.icon}>
          {getEventIcon(event.type)}
        </Text>
        <Text
          style={[
            styles.message,
            { color: getEventColor(event.type) }
          ]}
          numberOfLines={3}
        >
          {event.message}
        </Text>
      </View>
      {event.metadata && Object.keys(event.metadata).length > 0 && (
        <View style={styles.metadata}>
          {Object.entries(event.metadata).map(([key, value]) => (
            <Text key={key} style={styles.metadataText}>
              {key}: {String(value)}
            </Text>
          ))}
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
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  agentName: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
  },
  timestamp: {
    color: '#9CA3AF',
    fontSize: 12,
  },
  content: {
    flexDirection: 'row',
    alignItems: 'flex-start',
  },
  icon: {
    fontSize: 16,
    marginRight: 8,
    marginTop: 2,
  },
  message: {
    flex: 1,
    fontSize: 14,
    lineHeight: 20,
  },
  metadata: {
    marginTop: 8,
    paddingTop: 8,
    borderTopWidth: 1,
    borderTopColor: '#374151',
  },
  metadataText: {
    color: '#6B7280',
    fontSize: 12,
    marginBottom: 2,
  },
});
