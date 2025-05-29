import React, { useCallback, useMemo } from 'react';
import {
  View,
  Text,
  FlatList,
  RefreshControl,
  StyleSheet,
  ActivityIndicator,
  TouchableOpacity,
} from 'react-native';
import { useGetRecentUpdatesQuery, useGetAgentsQuery } from '@/services/api';
import { Event } from '@/types/api';
import UpdateItem from '@/components/UpdateItem';

export default function UpdatesScreen() {
  const {
    data: updatesData,
    isLoading: updatesLoading,
    error: updatesError,
    refetch: refetchUpdates,
  } = useGetRecentUpdatesQuery({ limit: 50 });

  const {
    data: agentsData,
  } = useGetAgentsQuery({ limit: 100 });

  const agentNamesMap = useMemo(() => {
    if (!agentsData?.agents) return {};
    return agentsData.agents.reduce((acc, agent) => {
      acc[agent.id] = agent.name;
      return acc;
    }, {} as Record<string, string>);
  }, [agentsData]);

  const handleRefresh = useCallback(() => {
    refetchUpdates();
  }, [refetchUpdates]);

  const renderUpdateItem = useCallback(({ item }: { item: Event }) => (
    <UpdateItem
      event={item}
      agentName={agentNamesMap[item.agent_id]}
    />
  ), [agentNamesMap]);

  const renderEmptyState = () => (
    <View style={styles.emptyState}>
      <Text style={styles.emptyStateText}>📢</Text>
      <Text style={styles.emptyStateTitle}>No Recent Updates</Text>
      <Text style={styles.emptyStateSubtitle}>
        Updates from your agents will appear here
      </Text>
    </View>
  );

  const renderError = () => (
    <View style={styles.errorState}>
      <Text style={styles.errorText}>❌ Failed to load updates</Text>
      <TouchableOpacity style={styles.retryButton} onPress={handleRefresh}>
        <Text style={styles.retryButtonText}>Retry</Text>
      </TouchableOpacity>
    </View>
  );

  const renderHeader = () => (
    <View style={styles.header}>
      <Text style={styles.headerTitle}>Recent Updates</Text>
      <Text style={styles.headerSubtitle}>
        Latest activity from your agent fleet
      </Text>
    </View>
  );

  if (updatesLoading && !updatesData) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#3B82F6" />
        <Text style={styles.loadingText}>Loading updates...</Text>
      </View>
    );
  }

  if (updatesError) {
    return renderError();
  }

  const updates = updatesData?.updates || [];

  return (
    <View style={styles.container}>
      <FlatList
        data={updates}
        renderItem={renderUpdateItem}
        keyExtractor={(item) => item.id}
        ListHeaderComponent={renderHeader}
        ListEmptyComponent={renderEmptyState}
        refreshControl={
          <RefreshControl
            refreshing={updatesLoading}
            onRefresh={handleRefresh}
            tintColor="#3B82F6"
            colors={['#3B82F6']}
          />
        }
        contentContainerStyle={
          updates.length === 0 ? styles.emptyContainer : styles.contentContainer
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
  headerTitle: {
    color: '#ffffff',
    fontSize: 24,
    fontWeight: 'bold',
    marginBottom: 4,
  },
  headerSubtitle: {
    color: '#9CA3AF',
    fontSize: 14,
  },
  contentContainer: {
    paddingBottom: 16,
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
