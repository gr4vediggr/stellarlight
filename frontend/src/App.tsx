import { useState } from "react";
import "./styles/globals.css";
import { useAuth, AuthProvider } from "./core/auth/AuthContext";
import { Dashboard } from "./pages/DashboardPage";
import { LoginPage } from "./components/auth/LoginPage";
import { RegisterPage } from "./components/auth/Register";


// Main App Component
const App = () => {
  const [currentPage, setCurrentPage] = useState('login');
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-900 via-purple-900 to-indigo-900 flex items-center justify-center">
        <div className="text-white text-xl">Loading...</div>
      </div>
    );
  }

  if (user) {
    return <Dashboard />;
  }

  return (
    <div>
      {currentPage === 'login' ? (
        <LoginPage onSwitchToRegister={() => setCurrentPage('register')} />
      ) : (
        <RegisterPage onSwitchToLogin={() => setCurrentPage('login')} />
      )}
    </div>
  );
};

// Root App with Auth Provider
const GameApp = () => {
  return (
    <AuthProvider>
      <App />
    </AuthProvider>
  );

};

export default GameApp

