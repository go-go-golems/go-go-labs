import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

// Define our single API slice object
export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: fetchBaseQuery({ baseUrl: 'http://localhost:8080/api' }),
  tagTypes: ['StreamInfo'],
  endpoints: (builder) => ({
    getStreamInfo: builder.query({
      query: () => '/stream-info',
      transformResponse: (response) => {
        // Extract the data from the response
        if (response.success && response.data) {
          // The API response has completedSteps (lowercase), not CompletedSteps (uppercase)
          return {
            id: response.data.id,
            title: response.data.title,
            description: response.data.description,
            startTime: response.data.startTime,
            language: response.data.language,
            githubRepo: response.data.githubRepo,
            viewerCount: response.data.viewerCount,
            completedSteps: response.data.completedSteps.map(step => step.content),
            activeStep: response.data.activeStep ? response.data.activeStep.content : "",
            upcomingSteps: response.data.upcomingSteps.map(step => step.content),
          };
        }
        return null;
      },
      providesTags: ['StreamInfo'],
    }),
    updateStreamInfo: builder.mutation({
      query: (streamInfo) => ({
        url: '/stream-info',
        method: 'POST',
        body: streamInfo,
      }),
      invalidatesTags: ['StreamInfo'],
    }),
    addStep: builder.mutation({
      query: (step) => ({
        url: '/steps',
        method: 'POST',
        body: step,
      }),
      invalidatesTags: ['StreamInfo'],
    }),
    updateStepStatus: builder.mutation({
      query: (data) => ({
        url: '/steps/status',
        method: 'POST',
        body: data,
      }),
      invalidatesTags: ['StreamInfo'],
    }),
  }),
});

// Export the auto-generated hooks for the endpoints
export const { 
  useGetStreamInfoQuery, 
  useUpdateStreamInfoMutation,
  useAddStepMutation,
  useUpdateStepStatusMutation
} = apiSlice;
