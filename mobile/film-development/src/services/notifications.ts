import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';
import { ProcessStep, Roll } from '../types';

// Configure notification handler
Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowBanner: true,
    shouldShowList: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
  }),
});

export const requestNotificationPermissions = async (): Promise<boolean> => {
  const { status } = await Notifications.requestPermissionsAsync();
  if (status !== 'granted') {
    console.warn('Notification permission not granted');
    return false;
  }
  return true;
};

export const scheduleStepNotifications = async (
  steps: ProcessStep[], 
  currentRoll: Roll,
  startingStepIndex: number = 0
): Promise<string[]> => {
  // Cancel any existing notifications
  await Notifications.cancelAllScheduledNotificationsAsync();
  
  const notificationIds: string[] = [];
  let accumulatedTime = 0;
  
  for (let i = startingStepIndex; i < steps.length; i++) {
    const step = steps[i];
    let stepDuration = 0;
    
    // Calculate step duration
    if (step.id === 'developer') {
      // Developer step uses roll-specific time
      const [minutes, seconds] = currentRoll.devTime.split(':').map(Number);
      stepDuration = minutes * 60 + seconds;
    } else if (step.time !== 'varies') {
      const [minutes, seconds] = step.time.split(':').map(Number);
      stepDuration = minutes * 60 + seconds;
    } else {
      continue; // Skip steps with 'varies' time
    }
    
    accumulatedTime += stepDuration;
    
    const notificationId = await Notifications.scheduleNotificationAsync({
      content: {
        title: 'Step Complete!',
        body: `"${step.name}" step is done! ${i < steps.length - 1 ? 'Start the next step.' : 'All steps completed!'}`,
        sound: true,
        data: {
          stepIndex: i,
          stepName: step.name,
          rollNumber: currentRoll.roll,
        },
      },
      trigger: { 
        type: Notifications.SchedulableTriggerInputTypes.TIME_INTERVAL,
        seconds: accumulatedTime 
      },
    });
    
    notificationIds.push(notificationId);
  }
  
  return notificationIds;
};

export const cancelAllNotifications = async (): Promise<void> => {
  await Notifications.cancelAllScheduledNotificationsAsync();
};

export const scheduleStepNotification = async (
  stepName: string,
  delaySeconds: number,
  rollNumber: number,
  stepIndex: number
): Promise<string> => {
  return await Notifications.scheduleNotificationAsync({
    content: {
      title: 'Step Complete!',
      body: `"${stepName}" step is done!`,
      sound: true,
      data: {
        stepIndex,
        stepName,
        rollNumber,
      },
    },
    trigger: { 
      type: Notifications.SchedulableTriggerInputTypes.TIME_INTERVAL,
      seconds: delaySeconds 
    },
  });
};

// Listen for notification responses (when user taps notification)
export const addNotificationResponseListener = (
  listener: (response: Notifications.NotificationResponse) => void
) => {
  return Notifications.addNotificationResponseReceivedListener(listener);
};

// Listen for notifications received while app is in foreground
export const addNotificationReceivedListener = (
  listener: (notification: Notifications.Notification) => void
) => {
  return Notifications.addNotificationReceivedListener(listener);
};