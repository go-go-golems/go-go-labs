import React from 'react';
import './App.css';
import StreamInfoDisplay from './components/StreamInfoDisplay';
import TailwindDemo from './components/TailwindDemo';

function App() {
  return (
    <div className="App">
      <TailwindDemo />
      <StreamInfoDisplay />
    </div>
  );
}

export default App;
