import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

export interface Widget { id: number; name: string }

export const widgetsApi = createApi({
  reducerPath: 'widgetsApi',
  baseQuery: fetchBaseQuery({ baseUrl: '/api' }),
  tagTypes: ['Widgets'],
  endpoints: builder => ({
    getWidgets: builder.query<Widget[], void>({
      query: () => 'widgets',   // GET /api/widgets
      providesTags: ['Widgets'],
      keepUnusedDataFor: 30, // in seconds
    }),
  }),
});

export const { useGetWidgetsQuery } = widgetsApi; 