import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import { useGetWidgetsQuery } from '../services/widgetsApi';
import { setFeaturedWidget } from '../features/featuredWidget/featuredWidgetSlice';
import type { AppDispatch } from '../store';

export const WidgetList = () => {
  const { 
    data, 
    isLoading, 
    error,
    refetch,
    isFetching
  } = useGetWidgetsQuery(undefined, {
    // Disable caching for this example to see changes immediately
    refetchOnMountOrArgChange: true,
    // Refetch on window focus
    refetchOnFocus: true
  });
  
  const dispatch = useDispatch<AppDispatch>();

  // Force refetch on component mount
  useEffect(() => {
    refetch();
  }, [refetch]);

  if (isLoading || isFetching) return <>Loading…</>;
  if (error) return <>Error</>;
  if (!data || data.length === 0) return <>No widgets found</>;

  return (
    <div>
      <ul className="widget-list">
        {data.map(widget => (
          <li key={widget.id} className="widget-list-item">
            {widget.name}
            <button 
              onClick={() => dispatch(setFeaturedWidget(widget))}
              className="feature-button"
            >
              Feature
            </button>
          </li>
        ))}
      </ul>
      <button 
        onClick={() => refetch()}
        className="refresh-button"
      >
        Refresh Widgets
      </button>
    </div>
  );
}; 