import { useEffect, useRef } from 'react';
import { useDispatch } from 'react-redux';
import { agentFleetApi } from './api';
import { setConnectionStatus } from '@/store/slices/uiSlice';
import { SSEEvent } from '@/types/api';
import EventSource from 'react-native-sse';
import { AGENT_FLEET_API_BASE_URL } from './api';

export function useSSEConnection() {
  const dispatch = useDispatch();
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    let es: EventSource | null = null;
    const connectSSE = () => {
      try {
        dispatch(setConnectionStatus('reconnecting'));

        es = new EventSource(`${AGENT_FLEET_API_BASE_URL}stream`);

        es.addEventListener('open', () => {
          console.log('SSE connected');
          dispatch(setConnectionStatus('connected'));
        });

        es.addEventListener('message', (event: any) => {
          try {
            const sseEvent: SSEEvent = JSON.parse(event.data);
            handleSSEEvent(sseEvent);
          } catch (error) {
            console.error('Failed to parse SSE event:', error);
          }
        });

        es.addEventListener('error', (event: any) => {
          console.error('SSE error:', event);
          dispatch(setConnectionStatus('disconnected'));
          // Reconnect after 5 seconds
          setTimeout(connectSSE, 5000);
        });
      } catch (error) {
        console.error('Failed to connect SSE:', error);
        dispatch(setConnectionStatus('disconnected'));
      }
    };

    const handleSSEEvent = (sseEvent: SSEEvent) => {
      switch (sseEvent.event) {
        case 'agent_status_changed':
        case 'agent_progress_updated':
          // Invalidate agents cache to refetch updated data
          dispatch(agentFleetApi.util.invalidateTags(['Agent']));
          break;
          
        case 'agent_event_created':
          // Invalidate events cache for the specific agent
          dispatch(agentFleetApi.util.invalidateTags([
            { type: 'Event', id: sseEvent.data.agent_id }
          ]));
          break;
          
        case 'todo_updated':
          // Invalidate todos cache for the specific agent
          dispatch(agentFleetApi.util.invalidateTags([
            { type: 'Todo', id: sseEvent.data.agent_id }
          ]));
          break;
          
        case 'task_assigned':
          // Invalidate tasks cache
          dispatch(agentFleetApi.util.invalidateTags(['Task']));
          break;
          
        case 'command_received':
          // Invalidate commands cache for the specific agent
          dispatch(agentFleetApi.util.invalidateTags([
            { type: 'Command', id: sseEvent.data.agent_id }
          ]));
          break;
          
        case 'agent_question_posted':
          // Invalidate agents cache and show notification
          dispatch(agentFleetApi.util.invalidateTags(['Agent']));
          // TODO: Show push notification or in-app notification
          break;
          
        default:
          console.log('Unknown SSE event:', sseEvent.event);
      }
    };

    // Connect on mount
    connectSSE();

    // Cleanup on unmount
    return () => {
      if (es) {
        es.removeAllEventListeners();
        es.close();
        es = null;
      }
      dispatch(setConnectionStatus('disconnected'));
    };
  }, [dispatch]);

  return {
    disconnect: () => {
      // Not strictly needed, as cleanup handles it
      dispatch(setConnectionStatus('disconnected'));
    },
    reconnect: () => {
      // The effect will automatically reconnect
    }
  };
}
