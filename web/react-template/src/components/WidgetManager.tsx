import React from 'react';
import { WidgetList } from './WidgetList';
import { FeaturedWidget } from './FeaturedWidget';

export const WidgetManager: React.FC = () => {
  return (
    <div className="widget-manager">
      <div className="widget-manager-left">
        <h2>Available Widgets</h2>
        <WidgetList />
      </div>
      <div className="widget-manager-right">
        <FeaturedWidget />
      </div>
    </div>
  );
}; 