import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { UserProvider } from './core/store/userStore';
import { WebSocketProvider } from './core/websocket/WebSocketContext';
import { LobbyProvider } from './core/lobby/LobbyContext';
import { GameProvider } from './core/game/NewGameContext';

// Pages
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import DashboardPage from './pages/DashboardPage';
import LobbyPage from './pages/LobbyPage';
import GamePage from './pages/GamePage';

// Protected Route Component
import ProtectedRoute from './components/auth/ProtectedRoute';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 1,
    },
  },
});

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <UserProvider>
        <Router>
          <div className="App">
            <Routes>
              {/* Public routes */}
              <Route path="/login" element={<LoginPage />} />
              <Route path="/register" element={<RegisterPage />} />
              
              {/* Protected routes */}
              <Route
                path="/dashboard"
                element={
                  <ProtectedRoute>
                    <DashboardPage />
                  </ProtectedRoute>
                }
              />
              
              {/* Lobby route with WebSocket and Lobby contexts */}
              <Route
                path="/lobby"
                element={
                  <ProtectedRoute>
                    <WebSocketProvider>
                      <LobbyProvider>
                        <LobbyPage />
                      </LobbyProvider>
                    </WebSocketProvider>
                  </ProtectedRoute>
                }
              />
              
              {/* Game route with WebSocket and Game contexts */}
              <Route
                path="/game"
                element={
                  <ProtectedRoute>
                    <WebSocketProvider>
                      <GameProvider>
                        <GamePage />
                      </GameProvider>
                    </WebSocketProvider>
                  </ProtectedRoute>
                }
              />
              
              {/* Default redirect */}
              <Route path="/" element={<Navigate to="/dashboard" replace />} />
            </Routes>
          </div>
        </Router>
      </UserProvider>
    </QueryClientProvider>
  );
}

export default App;
