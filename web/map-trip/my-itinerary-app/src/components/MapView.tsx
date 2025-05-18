import { useEffect, useRef } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import L from 'leaflet';
import type { Map as LeafletMap, Marker as LeafletMarker } from 'leaflet';
import 'leaflet/dist/leaflet.css';  // Import Leaflet CSS for markers and controls
import type { RootState } from '../store/store';
import { setView } from '../store/mapSlice';

const MapView: React.FC = () => {
  const dispatch = useDispatch();
  // Get the currently selected itinerary points and map center/zoom from the store
  const points = useSelector((state: RootState) => {
    const selectedId = state.itineraries.selectedId;
    const itinerary = state.itineraries.list.find(it => it.id === selectedId);
    return itinerary ? itinerary.points : [];
  });
  const center = useSelector((state: RootState) => state.map.center);
  const zoom = useSelector((state: RootState) => state.map.zoom);

  const mapContainerRef = useRef<HTMLDivElement>(null);
  const mapRef = useRef<LeafletMap | null>(null);
  const markersRef = useRef<LeafletMarker[]>([]);

  useEffect(() => {
    if (mapRef.current || !mapContainerRef.current) {
      return; // Map already initialized or container not ready
    }
    // Create the map in the container with initial center and zoom from Redux state
    mapRef.current = L.map(mapContainerRef.current).setView(center, zoom);

    // Add OpenStreetMap tile layer
    L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      maxZoom: 19,
      attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>'
    }).addTo(mapRef.current);

    // Optional: We could store the Leaflet map instance or any other setup here.
  }, []);  // run once on mount

  useEffect(() => {
    if (!mapRef.current) return;
    // Remove existing markers from the map
    markersRef.current.forEach(marker => marker.remove());
    markersRef.current = [];

    // If there are no points (e.g., no itinerary selected), nothing to do
    if (points.length === 0) return;

    // Add a marker for each point of interest
    points.forEach(pt => {
      const marker = L.marker([pt.lat, pt.lng]).addTo(mapRef.current!);
      marker.bindPopup(
        `<strong>${pt.name}</strong><br/>
         Best time: ${pt.bestTime}<br/>
         Photo tip: ${pt.photoTips}`
      ); // Attach popup with info
      markersRef.current.push(marker);
    });

    // After adding markers, adjust the map view to show them (optional):
    // Here we simply center on the first point and keep a default zoom.
    const first = points[0];
    mapRef.current.setView([first.lat, first.lng], 13);
    // Alternatively, to automatically fit all markers in view, you could use:
    // const bounds = L.latLngBounds(points.map(p => [p.lat, p.lng]));
    // mapRef.current.fitBounds(bounds);
    // And update the Redux store with the new view if desired.
    
    // Update the Redux store's map center/zoom state to match the new view
    dispatch(setView({ center: [first.lat, first.lng], zoom: 13 }));
  }, [points, dispatch]);

  return <div id="map" ref={mapContainerRef} style={{ width: '100%', height: '500px' }} />;
};

export default MapView;