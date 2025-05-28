import React, { useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  RefreshControl,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { NativeStackNavigationProp } from '@react-navigation/native-stack';
import { useGetAgentsQuery, useGetFleetStatusQuery } from '@/services/api';
import { RootStackParamList } from '@/types/navigation';
import { Agent } from '@/types/api';
import AgentCard from '@/components/AgentCard';

type FleetScreenNavigationProp = NativeStackNavigationProp<RootStackParamList>;

export default function FleetScreen() {
  const navigation = useNavigation<FleetScreenNavigationProp>();
  
  const {
    data: agentsData,
    isLoading: agentsLoading,
    error: agentsError,
    refetch: refetchAgents,
  } = useGetAgentsQuery({ limit: 50 });

  const {
    data: fleetStatus,
    isLoading: statusLoading,
    error: statusError,
    refetch: refetchStatus,
  } = useGetFleetStatusQuery();

  const handleRefresh = useCallback(() => {
    refetchAgents();
    refetchStatus();
  }, [refetchAgents, refetchStatus]);

  const handleAgentPress = useCallback((agent: Agent) => {
    navigation.navigate('AgentDetail', { agent });
  }, [navigation]);

  const handleSyncAll = useCallback(() => {
    // TODO: Implement sync all functionality
    handleRefresh();
  }, [handleRefresh]);

  const renderAgentCard = useCallback(({ item }: { item: Agent }) => (
    <AgentCard
      agent={item}
      onPress={() => handleAgentPress(item)}
    />
  ), [handleAgentPress]);

  const renderHeader = () => (
    <View style={styles.header}>
      {fleetStatus && (
        <View style={styles.statsContainer}>
          <View style={styles.statRow}>
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{fleetStatus.total_agents}</Text>
              <Text style={styles.statLabel}>Total Agents</Text>
            </View>
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{fleetStatus.active_agents}</Text>
              <Text style={styles.statLabel}>Active</Text>
            </View>
            <View style={styles.statItem}>
              <Text style={[styles.statNumber, { color: '#F59E0B' }]}>
                {fleetStatus.agents_needing_feedback}
              </Text>
              <Text style={styles.statLabel}>Need Feedback</Text>
            </View>
          </View>
          <View style={styles.statRow}>
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{fleetStatus.pending_tasks}</Text>
              <Text style={styles.statLabel}>Pending Tasks</Text>
            </View>
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{fleetStatus.total_files_changed}</Text>
              <Text style={styles.statLabel}>Files Changed</Text>
            </View>
            <View style={styles.statItem}>
              <Text style={styles.statNumber}>{fleetStatus.total_commits_today}</Text>
              <Text style={styles.statLabel}>Commits Today</Text>
            </View>
          </View>
        </View>
      )}
      
      <TouchableOpacity style={styles.syncButton} onPress={handleSyncAll}>
        <Text style={styles.syncButtonText}>🔄 Sync All</Text>
      </TouchableOpacity>
    </View>
  );

  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Text style={styles.emptyStateText}>🤖</Text>
      <Text style={styles.emptyStateTitle}>No Agents Found</Text>
      <Text style={styles.emptyStateSubtitle}>
        Pull to refresh or check your connection
      </Text>
    </View>
  );

  const renderError = () => (
    <View style={styles.errorState}>
      <Text style={styles.errorText}>❌ Failed to load agents</Text>
      <TouchableOpacity style={styles.retryButton} onPress={handleRefresh}>
        <Text style={styles.retryButtonText}>Retry</Text>
      </TouchableOpacity>
    </View>
  );

  if (agentsLoading && !agentsData) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#3B82F6" />
        <Text style={styles.loadingText}>Loading agents...</Text>
      </View>
    );
  }

  if (agentsError) {
    return renderError();
  }

  const agents = agentsData?.agents || [];

  return (
    <View style={styles.container}>
      <FlatList
        data={agents}
        renderItem={renderAgentCard}
        keyExtractor={(item) => item.id}
        ListHeaderComponent={renderHeader}
        ListEmptyComponent={renderEmptyState}
        refreshControl={
          <RefreshControl
            refreshing={agentsLoading}
            onRefresh={handleRefresh}
            tintColor="#3B82F6"
            colors={['#3B82F6']}
          />
        }
        contentContainerStyle={
          agents.length === 0 ? styles.emptyContainer : undefined
        }
        showsVerticalScrollIndicator={false}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#000000',
  },
  header: {
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#374151',
  },
  statsContainer: {
    marginBottom: 16,
  },
  statRow: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginBottom: 8,
  },
  statItem: {
    alignItems: 'center',
  },
  statNumber: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: 'bold',
  },
  statLabel: {
    color: '#9CA3AF',
    fontSize: 12,
    marginTop: 4,
  },
  syncButton: {
    backgroundColor: '#3B82F6',
    paddingVertical: 12,
    paddingHorizontal: 24,
    borderRadius: 8,
    alignItems: 'center',
  },
  syncButtonText: {
    color: '#ffffff',
    fontSize: 16,
    fontWeight: '600',
  },
  loadingContainer: {
    flex: 1,
    backgroundColor: '#000000',
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    color: '#9CA3AF',
    fontSize: 16,
    marginTop: 16,
  },
  emptyContainer: {
    flex: 1,
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
    backgroundColor: '#000000',
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
