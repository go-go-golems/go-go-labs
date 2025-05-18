import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import './App.css';
import MapView from './components/MapView';
import ItinerarySelector from './components/ItinerarySelector';
import { setItineraries } from './store/itinerarySlice';
import { setView } from './store/mapSlice';

const App: React.FC = () => {
  const dispatch = useDispatch();

  useEffect(() => {
    // Fetch the itineraries JSON on app load
    fetch('/itineraries.json')
      .then(response => response.json())
      .then((data) => {
        dispatch(setItineraries(data));
        // Optionally, set the map view to the first itinerary's first point:
        if (data.length > 0 && data[0].points.length > 0) {
          const firstPt = data[0].points[0];
          dispatch(setView({ center: [firstPt.lat, firstPt.lng], zoom: 13 }));
        }
      })
      .catch(err => {
        console.error("Failed to load itineraries:", err);
      });
  }, [dispatch]);

  return (
    <div style={{ padding: '1rem' }}>
      <h1>🗺️ San Francisco Tour Itineraries</h1>
      <p>Explore points of interest on the map by selecting an itinerary below.</p>
      <ItinerarySelector />
      <div style={{ marginTop: '1rem' }}>
        <MapView />
      </div>
    </div>
  );
};

export default App;