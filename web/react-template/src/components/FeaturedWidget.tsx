import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import type { RootState, AppDispatch } from '../store';
import { toggleHighlight, clearFeaturedWidget } from '../features/featuredWidget/featuredWidgetSlice';

export const FeaturedWidget: React.FC = () => {
  const dispatch = useDispatch<AppDispatch>();
  const { widget, isHighlighted } = useSelector((state: RootState) => state.featuredWidget);

  if (!widget) {
    return <div className="featured-widget-empty">No widget is currently featured</div>;
  }

  return (
    <div className={`featured-widget ${isHighlighted ? 'highlighted' : ''}`}>
      <h3>Featured Widget</h3>
      <div className="widget-detail">
        <strong>ID:</strong> {widget.id}
      </div>
      <div className="widget-detail">
        <strong>Name:</strong> {widget.name}
      </div>
      <div className="widget-actions">
        <button onClick={() => dispatch(toggleHighlight())}>
          {isHighlighted ? 'Remove Highlight' : 'Highlight'}
        </button>
        <button onClick={() => dispatch(clearFeaturedWidget())}>
          Clear
        </button>
      </div>
    </div>
  );
}; 