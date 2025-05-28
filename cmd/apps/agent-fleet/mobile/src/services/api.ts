import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
  Agent,
  Event,
  TodoItem,
  Task,
  Command,
  FleetStatus,
  AgentsResponse,
  EventsResponse,
  TasksResponse,
  TodosResponse,
  CommandsResponse,
  RecentUpdatesResponse,
} from '@/types/api';

const baseQuery = fetchBaseQuery({
  baseUrl: 'https://api.agentfleet.dev/v1/',
  prepareHeaders: (headers, { getState }) => {
    // TODO: Add auth token from state
    // const token = (getState() as RootState).auth.token;
    // if (token) {
    //   headers.set('authorization', `Bearer ${token}`);
    // }
    headers.set('authorization', 'Bearer demo-token');
    return headers;
  },
});

export const agentFleetApi = createApi({
  reducerPath: 'agentFleetApi',
  baseQuery,
  tagTypes: ['Agent', 'Event', 'Todo', 'Task', 'Command', 'FleetStatus'],
  endpoints: (builder) => ({
    // Agents
    getAgents: builder.query<AgentsResponse, { status?: string; limit?: number; offset?: number }>({
      query: (params) => ({
        url: 'agents',
        params,
      }),
      providesTags: ['Agent'],
    }),
    
    getAgent: builder.query<Agent, string>({
      query: (agentId) => `agents/${agentId}`,
      providesTags: (result, error, agentId) => [{ type: 'Agent', id: agentId }],
    }),

    // Agent Events
    getAgentEvents: builder.query<EventsResponse, {
      agentId: string;
      type?: string;
      since?: string;
      limit?: number;
      offset?: number;
    }>({
      query: ({ agentId, ...params }) => ({
        url: `agents/${agentId}/events`,
        params,
      }),
      providesTags: (result, error, { agentId }) => [
        { type: 'Event', id: agentId },
      ],
    }),

    // Agent Todos
    getAgentTodos: builder.query<TodosResponse, string>({
      query: (agentId) => `agents/${agentId}/todos`,
      providesTags: (result, error, agentId) => [
        { type: 'Todo', id: agentId },
      ],
    }),

    updateTodo: builder.mutation<TodoItem, {
      agentId: string;
      todoId: string;
      updates: Partial<Pick<TodoItem, 'completed' | 'current' | 'text'>>;
    }>({
      query: ({ agentId, todoId, updates }) => ({
        url: `agents/${agentId}/todos/${todoId}`,
        method: 'PATCH',
        body: updates,
      }),
      invalidatesTags: (result, error, { agentId }) => [
        { type: 'Todo', id: agentId },
      ],
    }),

    // Commands
    getAgentCommands: builder.query<CommandsResponse, {
      agentId: string;
      status?: string;
      limit?: number;
    }>({
      query: ({ agentId, ...params }) => ({
        url: `agents/${agentId}/commands`,
        params,
      }),
      providesTags: (result, error, { agentId }) => [
        { type: 'Command', id: agentId },
      ],
    }),

    sendCommand: builder.mutation<Command, {
      agentId: string;
      content: string;
      type: 'instruction' | 'feedback' | 'question';
    }>({
      query: ({ agentId, ...body }) => ({
        url: `agents/${agentId}/commands`,
        method: 'POST',
        body,
      }),
      invalidatesTags: (result, error, { agentId }) => [
        { type: 'Command', id: agentId },
        { type: 'Agent', id: agentId },
      ],
    }),

    // Tasks
    getTasks: builder.query<TasksResponse, {
      status?: string;
      assigned_agent_id?: string;
      priority?: string;
      limit?: number;
      offset?: number;
    }>({
      query: (params) => ({
        url: 'tasks',
        params,
      }),
      providesTags: ['Task'],
    }),

    createTask: builder.mutation<Task, {
      title: string;
      description: string;
      priority: 'low' | 'medium' | 'high' | 'urgent';
      assigned_agent_id?: string;
    }>({
      query: (body) => ({
        url: 'tasks',
        method: 'POST',
        body,
      }),
      invalidatesTags: ['Task'],
    }),

    // Fleet Operations
    getFleetStatus: builder.query<FleetStatus, void>({
      query: () => 'fleet/status',
      providesTags: ['FleetStatus'],
    }),

    getRecentUpdates: builder.query<RecentUpdatesResponse, {
      limit?: number;
      since?: string;
    }>({
      query: (params) => ({
        url: 'fleet/recent-updates',
        params,
      }),
      providesTags: ['Event'],
    }),
  }),
});

export const {
  useGetAgentsQuery,
  useGetAgentQuery,
  useGetAgentEventsQuery,
  useGetAgentTodosQuery,
  useUpdateTodoMutation,
  useGetAgentCommandsQuery,
  useSendCommandMutation,
  useGetTasksQuery,
  useCreateTaskMutation,
  useGetFleetStatusQuery,
  useGetRecentUpdatesQuery,
} = agentFleetApi;
