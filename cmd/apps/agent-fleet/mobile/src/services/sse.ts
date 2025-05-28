import { useEffect, useRef } from 'react';
import { useDispatch } from 'react-redux';
import { agentFleetApi } from './api';
import { setConnectionStatus } from '@/store/slices/uiSlice';
import { SSEEvent } from '@/types/api';

export function useSSEConnection() {
  const dispatch = useDispatch();
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    const connectSSE = () => {
      try {
        dispatch(setConnectionStatus('reconnecting'));
        
        eventSourceRef.current = new EventSource('https://api.agentfleet.dev/v1/stream', {
          // Note: EventSource doesn't support custom headers in browser
          // For auth, the API would need to use URL-based auth or cookies
        });

        eventSourceRef.current.onopen = () => {
          console.log('SSE connected');
          dispatch(setConnectionStatus('connected'));
        };

        eventSourceRef.current.onmessage = (event) => {
          try {
            const sseEvent: SSEEvent = JSON.parse(event.data);
            handleSSEEvent(sseEvent);
          } catch (error) {
            console.error('Failed to parse SSE event:', error);
          }
        };

        eventSourceRef.current.onerror = (error) => {
          console.error('SSE error:', error);
          dispatch(setConnectionStatus('disconnected'));
          
          // Reconnect after 5 seconds
          setTimeout(connectSSE, 5000);
        };

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
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      dispatch(setConnectionStatus('disconnected'));
    };
  }, [dispatch]);

  return {
    disconnect: () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
        dispatch(setConnectionStatus('disconnected'));
      }
    },
    reconnect: () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
      // The effect will automatically reconnect
    }
  };
}
