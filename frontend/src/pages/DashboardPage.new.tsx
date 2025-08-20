import { Settings, LogOut, Users, Gamepad2 } from "lucide-react";
import { useState } from "react";
import { useAuth } from "../core/auth/AuthContext";
import { GameService } from "../core/game/GameService";
import { ApiService } from "../core/api/api_service";
import { ProfilePage } from "../components/auth/ProfilePage";
import { useNavigate } from "react-router-dom";
import { getRuntimeConfig } from "../config/runtimeConfig";

// Dashboard Component
export const Dashboard = () => {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [inviteCode, setInviteCode] = useState('');
  const [showJoinGame, setShowJoinGame] = useState(false);
  const [showProfile, setShowProfile] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Create GameService instance
  const apiService = new ApiService(getRuntimeConfig().apiUrl);
  const gameService = new GameService(apiService);

  const handleCreateGame = async () => {
    setLoading(true);
    setError(null);
    try {
      await gameService.createGame();
      navigate('/lobby');
    } catch (error: any) {
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  const handleJoinGame = async () => {
    if (!inviteCode.trim()) return;
    
    setLoading(true);
    setError(null);
    try {
      await gameService.joinGame(inviteCode.trim());
      setInviteCode('');
      setShowJoinGame(false);
      navigate('/lobby');
    } catch (error: any) {
      setError(error.message);
    } finally {
      setLoading(false);
    }
  };

  if (showProfile) {
    return <ProfilePage onBack={() => setShowProfile(false)} />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-900 via-purple-900 to-indigo-900">
      {/* Header */}
      <div className="flex justify-between items-center p-6">
        <h1 className="text-3xl font-bold text-white">StellarLight</h1>
        <div className="flex items-center space-x-4">
          <button
            onClick={() => setShowProfile(true)}
            className="p-2 text-white hover:text-blue-300 transition-colors"
            title="Profile"
          >
            <Settings className="w-6 h-6" />
          </button>
          <button
            onClick={logout}
            className="p-2 text-white hover:text-red-300 transition-colors"
            title="Logout"
          >
            <LogOut className="w-6 h-6" />
          </button>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-4xl mx-auto px-6 py-12">
        <div className="text-center mb-12">
          <h2 className="text-4xl font-bold text-white mb-4">
            Welcome, {user?.displayName || user?.email}
          </h2>
          <p className="text-xl text-blue-200">
            Ready to explore the galaxy?
          </p>
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-500/20 border border-red-500 rounded-lg text-red-200">
            {error}
          </div>
        )}

        {/* Game Actions */}
        <div className="grid md:grid-cols-2 gap-8 mb-12">
          {/* Create Game */}
          <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl p-8">
            <div className="flex items-center mb-6">
              <Gamepad2 className="w-8 h-8 text-blue-400 mr-3" />
              <h3 className="text-2xl font-semibold text-white">Create Game</h3>
            </div>
            <p className="text-blue-200 mb-6">
              Start a new game and invite friends to join your stellar empire.
            </p>
            <button
              onClick={handleCreateGame}
              disabled={loading}
              className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-800 disabled:opacity-50 text-white font-medium py-3 px-6 rounded-lg transition-colors flex items-center justify-center space-x-2"
            >
              {loading ? (
                <div className="animate-spin rounded-full h-5 w-5 border-2 border-white border-t-transparent"></div>
              ) : (
                <>
                  <Gamepad2 className="w-5 h-5" />
                  <span>Create New Game</span>
                </>
              )}
            </button>
          </div>

          {/* Join Game */}
          <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl p-8">
            <div className="flex items-center mb-6">
              <Users className="w-8 h-8 text-green-400 mr-3" />
              <h3 className="text-2xl font-semibold text-white">Join Game</h3>
            </div>
            <p className="text-blue-200 mb-6">
              Enter an invite code to join an existing game.
            </p>
            
            {showJoinGame ? (
              <div className="space-y-4">
                <input
                  type="text"
                  placeholder="Enter invite code..."
                  value={inviteCode}
                  onChange={(e) => setInviteCode(e.target.value)}
                  className="w-full bg-white/10 border border-white/30 rounded-lg px-4 py-3 text-white placeholder-white/50 focus:outline-none focus:border-green-400"
                  onKeyPress={(e) => e.key === 'Enter' && handleJoinGame()}
                />
                <div className="flex space-x-3">
                  <button
                    onClick={handleJoinGame}
                    disabled={!inviteCode.trim() || loading}
                    className="flex-1 bg-green-600 hover:bg-green-700 disabled:bg-green-800 disabled:opacity-50 text-white font-medium py-3 px-6 rounded-lg transition-colors"
                  >
                    Join Game
                  </button>
                  <button
                    onClick={() => {
                      setShowJoinGame(false);
                      setInviteCode('');
                    }}
                    className="flex-1 bg-gray-600 hover:bg-gray-700 text-white font-medium py-3 px-6 rounded-lg transition-colors"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setShowJoinGame(true)}
                className="w-full bg-green-600 hover:bg-green-700 text-white font-medium py-3 px-6 rounded-lg transition-colors flex items-center justify-center space-x-2"
              >
                <Users className="w-5 h-5" />
                <span>Join Existing Game</span>
              </button>
            )}
          </div>
        </div>

        {/* Recent Games / Stats could go here */}
        <div className="bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl p-8">
          <h3 className="text-xl font-semibold text-white mb-4">Game Statistics</h3>
          <div className="grid grid-cols-3 gap-6 text-center">
            <div>
              <div className="text-2xl font-bold text-blue-400">0</div>
              <div className="text-blue-200">Games Played</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-green-400">0</div>
              <div className="text-blue-200">Victories</div>
            </div>
            <div>
              <div className="text-2xl font-bold text-purple-400">0</div>
              <div className="text-blue-200">Empires Built</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};
