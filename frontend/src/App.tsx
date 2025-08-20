import { useState } from "react";
import { BrowserRouter as Router, Routes, Route, Navigate } from "react-router-dom";
import "./styles/globals.css";
import { useAuth, AuthProvider } from "./core/auth/AuthContext";
import { GameProvider } from "./core/game/GameContext";
import { Dashboard } from "./pages/DashboardPage";
import { LobbyPage } from "./pages/LobbyPage";
import { LoginPage } from "./components/auth/LoginPage";
import { RegisterPage } from "./components/auth/Register";


// Protected Route Component
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-900 via-purple-900 to-indigo-900 flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  return user ? <>{children}</> : <Navigate to="/login" replace />;
};

// Auth Route Component - redirects to dashboard if already logged in
const AuthRoute = ({ children }: { children: React.ReactNode }) => {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-900 via-purple-900 to-indigo-900 flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  return user ? <Navigate to="/dashboard" replace /> : <>{children}</>;
};

// Main App Component with routing
const App = () => {
  const [authPage, setAuthPage] = useState<'login' | 'register'>('login');

  return (
    <Router>
      <Routes>
        <Route 
          path="/login" 
          element={
            <AuthRoute>
              {authPage === 'login' ? (
                <LoginPage onSwitchToRegister={() => setAuthPage('register')} />
              ) : (
                <RegisterPage onSwitchToLogin={() => setAuthPage('login')} />
              )}
            </AuthRoute>
          } 
        />
        <Route 
          path="/dashboard" 
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          } 
        />
        <Route 
          path="/lobby" 
          element={
            <ProtectedRoute>
              <GameProvider>
                <LobbyPage />
              </GameProvider>
            </ProtectedRoute>
          } 
        />
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Router>
  );
};

// Root App with Providers
const GameApp = () => {
  return (
    <AuthProvider>
      <App />
    </AuthProvider>
  );
};

export default GameApp

