// This file configures the Bun build system for the React frontend
// It will be used to build the React app for production and development

const { build } = require("bun");

async function buildReact() {
  // Build the React app for production
  await build({
    entrypoints: ["./src/index.js"],
    outdir: "./build",
    minify: true,
    target: "browser",
    sourcemap: "external",
  });
  
  console.log("React build completed successfully!");
}

// Export the build function for use in scripts
module.exports = { buildReact };
