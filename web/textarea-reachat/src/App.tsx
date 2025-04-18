import React from 'react';
import { ThemeProvider, theme } from 'reablocks';
import ChatWithTextarea from "./ChatWithTextarea";
import "./index.css"; // Import Tailwind and Reachat styles

function App() {
  return (
    <ThemeProvider theme={theme}>
      <main className="max-w-3xl mx-auto py-10 px-4"> {/* Added padding */}
        <h1 className="text-2xl font-bold mb-4">Blog post draft</h1>
        <ChatWithTextarea />
        {/* Add more instances or other content here */}
        {/* <h2 className="text-xl font-semibold mt-8 mb-4">Another Text Area</h2> */}
        {/* <ChatWithTextarea /> */}
      </main>
    </ThemeProvider>
  );
}

export default App; 