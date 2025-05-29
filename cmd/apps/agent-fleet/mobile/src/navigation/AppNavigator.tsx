import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { RootStackParamList } from '@/types/navigation';
import MainTabNavigator from './MainTabNavigator';
import AgentDetailScreen from '@/screens/AgentDetailScreen';

const Stack = createNativeStackNavigator<RootStackParamList>();

export default function AppNavigator() {
  return (
    <NavigationContainer>
      <Stack.Navigator
        screenOptions={{
          headerStyle: {
            backgroundColor: '#000000',
          },
          headerTintColor: '#ffffff',
          headerTitleStyle: {
            fontWeight: 'bold',
          },
        }}
      >
        <Stack.Screen
          name="Main"
          component={MainTabNavigator}
          options={{ headerShown: false }}
        />
        <Stack.Screen
          name="AgentDetail"
          component={AgentDetailScreen}
          options={({ route }) => ({
            title: route.params.agent.name,
            presentation: 'modal',
          })}
        />
      </Stack.Navigator>
    </NavigationContainer>
  );
}
