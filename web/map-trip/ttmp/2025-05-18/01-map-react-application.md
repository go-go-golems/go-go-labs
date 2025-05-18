Great! I’ll put together a full guide for building a San Francisco tourism itinerary viewer using React, TypeScript, Vite, Bun, RTK Toolkit, and Leaflet. It will use JSON input for itinerary data, support multiple preselected itineraries, and include interactive map features like markers and popups.

I’ll let you know once the guide is ready.


# Building a San Francisco Itinerary Map App (React + TypeScript + Bun + Vite + RTK + Leaflet)

## Introduction

In this tutorial, we will build a **small web app for tourist itineraries in San Francisco** using a modern web stack: **React** (with **TypeScript**), **Vite** (for fast bundling), **Bun** (as our package manager and runtime), **Redux Toolkit (RTK)** for state management, and **Leaflet** for interactive maps. The app will allow users to choose from multiple pre-defined itineraries and see points of interest (like the Golden Gate Bridge or Palace of Fine Arts) plotted on a map. Clicking a map marker will show a popup with details such as the **best time to visit** and **photography tips** for that location. Users can switch between itineraries using a dropdown menu, and the map will update to show the selected tour. We’ll cover everything from project setup to displaying the map and wiring up state management, with code snippets and clear structure.

**Key Features We'll Implement:**

* Load itinerary data from a local JSON file.
* Support multiple itineraries (with a default selection on load).
* Display points of interest on a Leaflet map as markers.
* Show popups on marker clicks with details (best visiting time, photo tips).
* Allow switching between itineraries via a dropdown.
* Use Redux Toolkit to manage the selected itinerary and current map view state.
* Simple, flexible styling (no strict design – just enough to make it usable).

Let's dive in!

## Setting Up the Project (Bun + Vite + React/TypeScript)

First, ensure you have **Bun** installed. Bun is a fast JavaScript runtime that also acts as a package manager. With Bun ready, we can scaffold a new project using Vite's React template.

**1. Initialize a Vite React + TypeScript project using Bun.** Open a terminal and run:

```bash
bun create vite my-itinerary-app
```

When prompted, select **React** as the framework and **TypeScript + SWC** as the variant. This will scaffold a new Vite project in the `my-itinerary-app` directory with React and TypeScript set up.

**2. Install dependencies.** Navigate into the project folder and install the initial packages:

```bash
cd my-itinerary-app
bun install
```

This will install React, ReactDOM, and other dependencies that the Vite template needs.

**3. Add additional libraries.** Our app needs Redux Toolkit, React-Redux, and Leaflet. We’ll also add type definitions for Leaflet. Using Bun, you can add multiple packages at once:

```bash
bun add @reduxjs/toolkit react-redux leaflet @types/leaflet
```

This will update `package.json` and install these packages to your project. Redux Toolkit provides simple setup for Redux, React-Redux lets our React components access the store, and Leaflet is the mapping library (with `@types/leaflet` for TypeScript support).

**4. Update the dev script to use Bun.** By default, the Vite dev server might run under Node. We want to use Bun as the runtime for speed. Open `package.json` and modify the `"dev"` script to use Bun’s executor (`bunx`):

```json
"scripts": {
  "dev": "bunx --bun vite",
  "build": "vite build",
  "serve": "vite preview"
}
```

This change ensures the Vite CLI runs with Bun. Now you can start the development server by running:

```bash
bun run dev
```

This will launch Vite’s dev server (through Bun) and you can open `http://localhost:5173` (or the port Vite specifies) to view the app in the browser.

At this point, we have a fresh React+TS project running. Next, we'll organize our project structure and then start adding our specific features.

## Project Structure and Files

Before writing code, let's outline the structure of our project and important files we will create:

```
my-itinerary-app/
├─ bun.lockb                      # Bun lockfile for dependencies
├─ package.json                   # Project metadata and scripts
├─ tsconfig.json                  # TypeScript configuration
├─ vite.config.ts                 # Vite configuration (default from template)
├─ index.html                     # HTML template for Vite
├─ public/
│   └─ itineraries.json           # **Local JSON file with itinerary data** (we will create this)
└─ src/
    ├─ main.tsx                   # Application entry point
    ├─ App.tsx                    # Root App component
    ├─ App.css                    # Global/app styles (if needed)
    ├─ store/
    │   ├─ store.ts               # Redux store configuration
    │   ├─ itinerarySlice.ts      # Redux Toolkit slice for itineraries
    │   └─ mapSlice.ts            # Redux Toolkit slice for map view state
    ├─ components/
    │   ├─ MapView.tsx            # Leaflet Map component
    │   └─ ItinerarySelector.tsx  # Dropdown component for switching itineraries
    └─ types.d.ts                 # (Optional) Custom type definitions for our data structures
```

Let's go through each major piece and implement it step by step.

## Sample Data: The Itineraries JSON

Our app will load itinerary data from a **local JSON file**. For simplicity, we'll include this file in the Vite project's `public/` directory so that it can be fetched easily (files in `public` are served as static assets). Create a file at `public/itineraries.json` with the following example content:

```json
[
  {
    "id": "classic_sf",
    "title": "Classic San Francisco",
    "points": [
      {
        "name": "Golden Gate Bridge",
        "lat": 37.8199,
        "lng": -122.4786,
        "bestTime": "Sunset (for golden light and fewer crowds)",
        "photoTips": "Capture the bridge from Battery Spencer or Fort Point for iconic angles"
      },
      {
        "name": "Palace of Fine Arts",
        "lat": 37.8028,
        "lng": -122.4487,
        "bestTime": "Morning (soft light and fewer people)",
        "photoTips": "Use the lagoon for reflections of the rotunda in your composition"
      },
      {
        "name": "Fisherman's Wharf",
        "lat": 37.8083,
        "lng": -122.4156,
        "bestTime": "Evening (for lively atmosphere and sunset views)",
        "photoTips": "Try a long exposure to capture light trails of the carousel and waterfront"
      }
    ]
  },
  {
    "id": "scenic_views",
    "title": "SF Scenic Views",
    "points": [
      {
        "name": "Twin Peaks",
        "lat": 37.7529,
        "lng": -122.4476,
        "bestTime": "Sunrise or night (for panoramic cityscape views)",
        "photoTips": "Bring a tripod for steady shots of the city lights after dark"
      },
      {
        "name": "Coit Tower",
        "lat": 37.8024,
        "lng": -122.4060,
        "bestTime": "Early morning (to avoid crowds, with soft light)",
        "photoTips": "Shoot from the base for dramatic perspective, or capture the skyline from the top"
      },
      {
        "name": "Baker Beach",
        "lat": 37.7932,
        "lng": -122.4840,
        "bestTime": "Sunset (for colorful skies behind the Golden Gate Bridge)",
        "photoTips": "Walk to the north end of the beach for the best angle of the bridge; use a wide lens"
      }
    ]
  }
]
```

Each itinerary has an `id`, a human-friendly `title`, and an array of `points`. Each point of interest includes a name, latitude `lat` and longitude `lng` (for map markers), and additional info like `bestTime` to visit and `photoTips`. Feel free to adjust or add itineraries and points as needed. These two itineraries will serve as our default options.

## Setting Up Redux Toolkit (Store and Slices)

Next, we'll set up state management using **Redux Toolkit**. We want to manage two pieces of state globally:

* The list of itineraries (loaded from the JSON) and the currently selected itinerary.
* The map view state (for example, the current map center and zoom level).

Using Redux Toolkit's **slices**, we can define this state and the reducers (operations) that modify it.

**1. Define types for our data (optional but recommended in TypeScript).** In a new file `src/types.d.ts` or alongside our slice file, define interfaces for Itinerary and Point of Interest:

```ts
// types.d.ts (or put in itinerarySlice.ts for self-containment)
interface PointOfInterest {
  name: string;
  lat: number;
  lng: number;
  bestTime: string;
  photoTips: string;
}

interface Itinerary {
  id: string;
  title: string;
  points: PointOfInterest[];
}
```

These will help TypeScript ensure we handle the data correctly.

**2. Create the itinerary slice.** In `src/store/itinerarySlice.ts`, create a slice to hold all itineraries and the selected itinerary id:

```ts
import { createSlice, PayloadAction } from '@reduxjs/toolkit';

interface ItineraryState {
  list: Itinerary[];        // all itineraries loaded
  selectedId: string | null; // id of the currently selected itinerary
}

const initialState: ItineraryState = {
  list: [],
  selectedId: null
};

const itinerarySlice = createSlice({
  name: 'itineraries',
  initialState,
  reducers: {
    setItineraries(state, action: PayloadAction<Itinerary[]>) {
      state.list = action.payload;
      // Auto-select the first itinerary by default, if not already selected
      if (state.list.length > 0 && state.selectedId === null) {
        state.selectedId = state.list[0].id;
      }
    },
    selectItinerary(state, action: PayloadAction<string>) {
      state.selectedId = action.payload;
    }
  }
});

export const { setItineraries, selectItinerary } = itinerarySlice.actions;
export default itinerarySlice.reducer;
```

This slice provides two reducer actions:

* `setItineraries` – to set the list of itineraries (when we load the JSON) and initialize `selectedId` to the first itinerary if none is selected yet.
* `selectItinerary` – to switch the current itinerary by id.

**3. Create the map slice.** In `src/store/mapSlice.ts`, define a slice for the map's view state (center coordinates and zoom level):

```ts
import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { LatLngExpression } from 'leaflet';  // Leaflet's type for [lat, lng] tuple

interface MapState {
  center: LatLngExpression;  // [lat, lng] tuple
  zoom: number;
}

const initialState: MapState = {
  center: [37.7749, -122.4194], // default center (San Francisco city)
  zoom: 12                     // default zoom level
};

const mapSlice = createSlice({
  name: 'map',
  initialState,
  reducers: {
    setView(state, action: PayloadAction<{ center: LatLngExpression; zoom: number }>) {
      state.center = action.payload.center;
      state.zoom = action.payload.zoom;
    }
  }
});

export const { setView } = mapSlice.actions;
export default mapSlice.reducer;
```

We initialize the map center to coordinates around downtown San Francisco (latitude 37.7749 N, longitude -122.4194 W) and a reasonable zoom level 12. The `setView` action will allow us to update the center/zoom, for example when the user switches itineraries (we may recenter the map).

**4. Configure the Redux store.** In `src/store/store.ts`, set up the Redux store and combine our slices:

```ts
import { configureStore } from '@reduxjs/toolkit';
import itineraryReducer from './itinerarySlice';
import mapReducer from './mapSlice';

export const store = configureStore({
  reducer: {
    itineraries: itineraryReducer,
    map: mapReducer
  }
});

// Helper types for useSelector and useDispatch hooks in TypeScript:
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

We combine the two slices under the keys `itineraries` and `map` in the state. The helper types `RootState` and `AppDispatch` will be useful for TypeScript when we use the Redux hooks in our components.

**5. Provide the store to the React app.** Open `src/main.tsx` (the entry file that renders `<App />` to the DOM) and wrap the `<App />` with Redux’s `<Provider>`:

```tsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { store } from './store/store';
import App from './App';
import './App.css';  // (if using a global CSS file for basic styling)

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <Provider store={store}>
      <App />
    </Provider>
  </React.StrictMode>
);
```

Now our React app can access the Redux store.

## Creating the Leaflet Map Component

Now for the core of the app: the map. We will create a `MapView` component that initializes a Leaflet map and places markers for the points of interest of the selected itinerary. We’ll also ensure that clicking a marker shows a popup with the location’s details.

**1. Import Leaflet and required styles.** Leaflet requires its CSS to be loaded for the map and markers to display correctly. Vite can bundle CSS from node modules, so we can import it in our component. Also, we'll need React hooks and Redux hooks:

```tsx
// src/components/MapView.tsx
import { useEffect, useRef } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import L, { Map as LeafletMap, Marker as LeafletMarker } from 'leaflet';
import 'leaflet/dist/leaflet.css';  // Import Leaflet CSS for markers and controls:contentReference[oaicite:4]{index=4}
import { RootState } from '../store/store';
import { setView } from '../store/mapSlice';
```

Here we import `leaflet` as `L` and also bring in its types (`LeafletMap` and `LeafletMarker`) for clarity. We also import the CSS directly, which is needed to style the map (alternatively, you could include the Leaflet CSS via a `<link>` in `index.html`, but importing in the bundle is convenient for React apps).

**2. Define the MapView component structure.** This component will render a `<div>` that Leaflet takes over to draw the map. We will use a `ref` to reference this div in JavaScript, since Leaflet manipulates the DOM directly. We also retrieve from Redux the current itinerary’s points and the desired map center/zoom:

```tsx
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
```

* `points` – an array of PointOfInterest for the selected itinerary (or empty if none selected yet).
* `center` and `zoom` – current map view parameters from Redux.
* `mapContainerRef` – attached to the map `<div>`, so we can pass it to Leaflet.
* `mapRef` – will hold the Leaflet Map instance once created (to avoid reinitializing on every render).
* `markersRef` – will track any marker objects we add, so we can remove them when updating the map.

**3. Initialize the Leaflet map (once).** We only want to create the map when the component first mounts. In a `useEffect` with an empty dependency array, we will create the map:

```tsx
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
    }).addTo(mapRef.current); //:contentReference[oaicite:6]{index=6}

    // Optional: We could store the Leaflet map instance or any other setup here.
  }, []);  // run once on mount
```

This effect runs only on first render (`[]` means no state dependencies). It creates a `L.map` on our container and sets the view to the `center` and `zoom` from the store. We then add a tile layer using free OpenStreetMap tiles (with attribution as required).

> **Note:** Using OpenStreetMap’s tile server is fine for development or small-scale use, but be mindful of their usage policy. The code above includes the required attribution string.

**4. Update markers whenever the selected itinerary changes.** When the `points` array updates (e.g., user selects a different itinerary or when data is first loaded), we want to refresh the markers on the map. We do this in another `useEffect` that depends on `points`:

```tsx
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
      ); // Attach popup with info:contentReference[oaicite:9]{index=9}
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
```

This effect will run whenever `points` changes. It first clears any existing markers from the map (using `marker.remove()` to remove them from the map, and resetting our `markersRef`). Then, for each point in the new itinerary, it creates a Leaflet marker at the given latitude/longitude and binds a popup. The popup content is a small HTML snippet with the location name in bold and the best time and photo tip on separate lines. We used `marker.bindPopup(...)` to attach the info – by default, Leaflet will show this popup when the marker is clicked.

After placing markers, we recenter the map to the first point in the list and set a zoom level (13, a neighborhood-level zoom). This ensures that when an itinerary is selected, the map moves to show the new points. In the code comments, we note an alternative: using `fitBounds` to automatically adjust the view to include all markers, which is great if points are spread out. That is optional; for simplicity we center on the first point.

We also dispatch `setView` to update the Redux store with the new center and zoom (so the state stays in sync with what we did to the map). This might be useful if other components need to know the current map view or if we wanted to store view state for other reasons.

**5. Render the map container.** Finally, our component should return the container div for the map. We must give this div a size – without an explicit height, the map won’t be visible. We can use inline styles or CSS classes. For this example, we’ll set a height of 500px via style:

```tsx
  return <div id="map" ref={mapContainerRef} style={{ width: '100%', height: '500px' }} />;
};

export default MapView;
```

We gave it an `id="map"` (not strictly required, but useful for debugging) and a height of 500px. You can adjust this, or use CSS to make the map fill the screen or any container. Just remember: *a Leaflet map container must have a defined height!* If you prefer, add CSS in `App.css` such as `#map { height: 100vh; width: 100%; }` to make it full viewport height, for example.

Our `MapView` component is ready. It initializes the map once, and updates markers whenever the selected itinerary’s points change.

## Building the Itinerary Selector UI

We need a UI control to let the user switch between the available itineraries. A simple approach is to use a dropdown (`<select>`) that lists the itinerary titles. When the user selects one, we dispatch the `selectItinerary` action (and maybe recenter the map).

Create `src/components/ItinerarySelector.tsx`:

```tsx
import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { RootState } from '../store/store';
import { selectItinerary, setItineraries, setView } from '../store';  // we'll adjust import paths as needed
import { Itinerary } from '../types.d';

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
```

This component uses `useSelector` to get the list of itineraries and the currently selected itinerary ID from Redux. It then renders a labeled `<select>` dropdown. Each `<option>` displays an itinerary’s title and uses its `id` as the value. The `value` of the `<select>` is bound to `selectedId` (so the dropdown reflects the current selection).

When the user changes selection, `handleSelectChange` is called. We:

* Get the selected `id` from the event.
* Dispatch `selectItinerary(newId)` to update the Redux state for the current itinerary.
* Then find the corresponding itinerary object from the list to retrieve its first point, and dispatch `setView` to recenter the map to that point (with a preset zoom). This ensures the map moves focus to the newly selected itinerary.

*(We could rely on the effect inside MapView to auto-fit markers, but since we want Redux to be the source of truth for map view as well, we handle it here. It also makes the center update happen immediately.)*

## Wiring Everything Together in the App

Now we have all the pieces: the Redux store (with slices and initial state), the MapView component, and the ItinerarySelector component. We need to put them together in our main `App` component and trigger the data loading from the JSON file.

Open `src/App.tsx` and set up the layout:

```tsx
import React, { useEffect } from 'react';
import { useDispatch } from 'react-redux';
import MapView from './components/MapView';
import ItinerarySelector from './components/ItinerarySelector';
import { setItineraries, setView } from './store/itinerarySlice'; // adjust import paths as needed

const App: React.FC = () => {
  const dispatch = useDispatch();

  useEffect(() => {
    // Fetch the itineraries JSON on app load
    fetch('/itineraries.json')
      .then(response => response.json())
      .then((data) => {
        dispatch(setItineraries(data));
        // Optionally, set the map view to the first itinerary’s first point:
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
```

Let's break down what happens in `App`:

* We use `useEffect` to load the data exactly once when the component mounts. We fetch `/itineraries.json` (this will load the file we placed in `public/itineraries.json`). Once the JSON is retrieved, we dispatch `setItineraries(data)` to save it in Redux. We also immediately set the map view to the first point of the first itinerary (if available) by dispatching `setView` – this ensures the map is centered appropriately on initial load.
* We include a basic header and instructions.
* We render the `<ItinerarySelector />` so the user can change the itinerary.
* We render the `<MapView />` below, wrapped in a container with some top margin.

The inline styles are just for basic spacing. You can replace those with classes and define styles in `App.css` as needed.

Now the pieces are connected: when `setItineraries` runs, the Redux store gets the data and sets a default selected itinerary. The `ItinerarySelector` will reflect this in its dropdown (because `selectedId` is set). The `MapView` will also react: its `points` from the store will update, triggering the effect to add markers for that itinerary. We also set the initial map view via `setView` so that the map centers on the first location.

## Running and Testing the App

Start the development server if it's not already running:

```bash
bun run dev
```

Open the app in your browser (typically at **[http://localhost:5173](http://localhost:5173)** for Vite). You should see the title and dropdown at the top, and the map rendered below.

* Initially, the map should load centered on the first itinerary's first point (e.g., Golden Gate Bridge for "Classic San Francisco"). Markers for Golden Gate Bridge, Palace of Fine Arts, and Fisherman's Wharf will be on the map. Try clicking each marker – a popup should appear with the *best time to visit* and *photo tip* for that spot.
* Use the **Select Itinerary** dropdown to switch to "SF Scenic Views". The map will update: markers for Twin Peaks, Coit Tower, and Baker Beach will appear, and the map recenters (in this example, likely on Twin Peaks). Click those markers to see their info popups.
* The dropdown selection is managed by Redux state (`selectedId`), and the map view recentering is also managed via state (`map.center`, `map.zoom`), demonstrating a clear separation of concerns.

If everything is set up correctly, you now have a working app displaying San Francisco itineraries on a map!

## Conclusion

In this guide, we built a complete React application using TypeScript and modern tools to display tourist itineraries on an interactive Leaflet map. We used **Bun + Vite** for a fast development environment, **Redux Toolkit** to manage application state (selected itinerary and map view), and **Leaflet** to handle map rendering and interactivity. By structuring the app into small pieces (data, slices, components), we created a flexible foundation that can be extended with more data or features (for example, adding user-added markers, routing between points, or styling improvements).

Key takeaways and possible next steps:

* **Bun** made project setup and running lightning-fast, and works seamlessly with Vite.
* **Redux Toolkit** allowed us to set up global state with minimal boilerplate. We cleanly separated the itinerary data logic and map view logic into slices.
* **Leaflet** provides powerful mapping capabilities in the browser. We used OpenStreetMap tiles and included proper attribution. We demonstrated adding markers and popups for our points of interest. You can further customize markers (colors, icons) or add more layers (routes, polygons, etc.) as needed.
* The app is styled minimally; you can enhance it with CSS or a component library for a better UI if desired (for example, making the dropdown prettier, adding a sidebar list of locations, etc.).

With this foundation, you can easily swap in your own city’s itineraries or add more details. Happy coding, and enjoy exploring San Francisco (virtually)! 🚀
