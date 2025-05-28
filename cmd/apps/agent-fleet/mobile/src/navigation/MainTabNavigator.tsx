import React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { MainTabParamList } from '@/types/navigation';
import FleetScreen from '@/screens/FleetScreen';
import UpdatesScreen from '@/screens/UpdatesScreen';
import TasksScreen from '@/screens/TasksScreen';

const Tab = createBottomTabNavigator<MainTabParamList>();

export default function MainTabNavigator() {
  return (
    <Tab.Navigator
      screenOptions={{
        tabBarStyle: {
          backgroundColor: '#000000',
          borderTopColor: '#374151',
        },
        tabBarActiveTintColor: '#3B82F6',
        tabBarInactiveTintColor: '#9CA3AF',
        headerStyle: {
          backgroundColor: '#000000',
        },
        headerTintColor: '#ffffff',
        headerTitleStyle: {
          fontWeight: 'bold',
        },
      }}
    >
      <Tab.Screen
        name="Fleet"
        component={FleetScreen}
        options={{
          title: '🤖 Fleet',
          headerTitle: 'Agent Fleet',
        }}
      />
      <Tab.Screen
        name="Updates"
        component={UpdatesScreen}
        options={{
          title: '📢 Updates',
          headerTitle: 'Recent Updates',
        }}
      />
      <Tab.Screen
        name="Tasks"
        component={TasksScreen}
        options={{
          title: '📋 Tasks',
          headerTitle: 'Task Management',
        }}
      />
    </Tab.Navigator>
  );
}
