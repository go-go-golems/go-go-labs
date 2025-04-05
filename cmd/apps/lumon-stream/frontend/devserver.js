// This file configures the development server for the React frontend
// It will be used to serve the React app in development mode

const { serve } = require("bun");

async function startDevServer() {
  const server = serve({
    port: 3000,
    fetch(req) {
      // Forward API requests to the backend
      const url = new URL(req.url);
      if (url.pathname.startsWith('/api')) {
        // Proxy to backend
        return fetch(`http://localhost:8080${url.pathname}${url.search}`, {
          method: req.method,
          headers: req.headers,
          body: req.body
        });
      }
      
      // Serve static files from the public directory
      return new Response("Development server running");
    },
  });
  
  console.log(`Development server running at http://localhost:${server.port}`);
}

// Export the server function for use in scripts
module.exports = { startDevServer };
