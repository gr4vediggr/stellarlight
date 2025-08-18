// src/index.js or src/main.jsx
import React from 'react';
import ReactDOM from 'react-dom/client';
import GameApp from './App'; // Make sure this path is correct

async function init() { 
  // Load runtime configuration
  await import('./config/runtimeConfig').then(module => module.loadRuntimeConfig());
const container = document.getElementById('root');
if (container) {
  const root = ReactDOM.createRoot(container);
  root.render(
    <React.StrictMode>
      <GameApp />  {/* Make sure you're rendering GameApp, not App */}
    </React.StrictMode>
  );
} else {
  throw new Error("Root container not found");
}
}

init();
