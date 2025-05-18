import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import type { RootState } from '../store/store';
import { selectItinerary } from '../store/itinerarySlice';
import { setView } from '../store/mapSlice';

const ItinerarySelector: React.FC = () => {
  const dispatch = useDispatch();
  const itineraries = useSelector((state: RootState) => state.itineraries.list);
  const selectedId = useSelector((state: RootState) => state.itineraries.selectedId);

  const handleSelectChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newId = e.target.value;
    dispatch(selectItinerary(newId));
    // Find the selected itinerary to recentre map
    const itinerary = itineraries.find(it => it.id === newId);
    if (itinerary && itinerary.points.length > 0) {
      const firstPoint = itinerary.points[0];
      dispatch(setView({ center: [firstPoint.lat, firstPoint.lng], zoom: 13 }));
    }
  };

  return (
    <div>
      <label htmlFor="itinerary-select"><strong>Select Itinerary: </strong></label>
      <select 
        id="itinerary-select" 
        value={selectedId || ''} 
        onChange={handleSelectChange}
      >
        {itineraries.map(it => (
          <option key={it.id} value={it.id}>{it.title}</option>
        ))}
      </select>
    </div>
  );
};

export default ItinerarySelector;