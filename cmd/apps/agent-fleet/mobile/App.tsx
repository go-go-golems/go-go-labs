import React from 'react';
import { StatusBar } from 'expo-status-bar';
import { Provider } from 'react-redux';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { store } from './src/store';
import AppNavigator from './src/navigation/AppNavigator';
import { useSSEConnection } from './src/services/sse';

function AppContent() {
  // Set up SSE connection for real-time updates
  useSSEConnection();
  
  return (
    <>
      <AppNavigator />
      <StatusBar style="light" backgroundColor="#000000" />
    </>
  );
}

export default function App() {
  return (
    <Provider store={store}>
      <SafeAreaProvider>
        <AppContent />
      </SafeAreaProvider>
    </Provider>
  );
}
