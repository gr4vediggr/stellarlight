import { Settings, LogOut, Users, Gamepad2 } from "lucide-react";
import { useState } from "react";
import { useAuth } from "../core/auth/AuthContext";
import { ProfilePage } from "../components/auth/ProfilePage";

// Dashboard Component
export const Dashboard = () => {
  const { user, gameSession, wsConnected, logout, createGame, joinGame, leaveGame } = useAuth();
  const [inviteCode, setInviteCode] = useState('');
  const [showJoinGame, setShowJoinGame] = useState(false);
  const [showProfile, setShowProfile] = useState(false);

  const handleCreateGame = () => {
    createGame();
  };

  const handleJoinGame = () => {
    if (inviteCode.trim()) {
      joinGame(inviteCode.trim());
      setInviteCode('');
      setShowJoinGame(false);
    }
  };

  const handleReturnToGame = () => {
    // Navigate to game view (you would implement this routing)
    console.log('Returning to game:', gameSession.gameId);
  };

  if (showProfile) {
    return <ProfilePage onBack={() => setShowProfile(false)} />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-indigo-900 p-4">
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-3xl font-bold text-white">Game Dashboard</h1>
            <p className="text-gray-300">Welcome back, {user?.displayName}!</p>
          </div>
          <div className="flex gap-4">
            <button
              onClick={() => setShowProfile(true)}
              className="flex items-center gap-2 px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition duration-200"
            >
              <Settings size={20} />
              Profile
            </button>
            <button
              onClick={logout}
              className="flex items-center gap-2 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition duration-200"
            >
              <LogOut size={20} />
              Logout
            </button>
          </div>
        </div>

        {/* Game Status */}
        {gameSession && (
          <div className="bg-green-500/20 border border-green-500/30 rounded-xl p-6 mb-8">
            <h2 className="text-xl font-semibold text-green-100 mb-2">Currently in Game</h2>
            <p className="text-green-200 mb-4">Game ID: {gameSession.gameId}</p>
            <div className="flex gap-4">
              <button
                onClick={handleReturnToGame}
                className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition duration-200"
              >
                Return to Game
              </button>
              <button
                onClick={leaveGame}
                className="px-6 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition duration-200"
              >
                Leave Game
              </button>
            </div>
          </div>
        )}

        {/* Game Actions */}
        <div className="grid md:grid-cols-2 gap-6 mb-8">
          {/* Create Game */}
          <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 border border-white/20">
            <div className="flex items-center gap-3 mb-4">
              <Gamepad2 className="w-8 h-8 text-blue-400" />
              <h2 className="text-xl font-semibold text-white">Create New Game</h2>
            </div>
            <p className="text-gray-300 mb-4">Start a new game lobby and invite friends to join.</p>
            <button
              onClick={handleCreateGame}
              disabled={!!gameSession || !wsConnected}
              className="w-full px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-lg hover:from-blue-700 hover:to-purple-700 transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {!wsConnected ? 'Connecting...' : gameSession ? 'Already in Game' : 'Create Game'}
            </button>
          </div>

          {/* Join Game */}
          <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 border border-white/20">
            <div className="flex items-center gap-3 mb-4">
              <Users className="w-8 h-8 text-green-400" />
              <h2 className="text-xl font-semibold text-white">Join Game</h2>
            </div>
            <p className="text-gray-300 mb-4">Enter an invite code to join an existing game.</p>
            
            {showJoinGame ? (
              <div className="space-y-3">
                <input
                  type="text"
                  value={inviteCode}
                  onChange={(e) => setInviteCode(e.target.value)}
                  placeholder="Enter invite code"
                  className="w-full px-4 py-2 bg-white/10 border border-white/20 rounded-lg text-white placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-green-500"
                />
                <div className="flex gap-2">
                  <button
                    onClick={handleJoinGame}
                    disabled={!wsConnected}
                    className="flex-1 px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {!wsConnected ? 'Connecting...' : 'Join'}
                  </button>
                  <button
                    onClick={() => {
                      setShowJoinGame(false);
                      setInviteCode('');
                    }}
                    className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition duration-200"
                  >
                    Cancel
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setShowJoinGame(true)}
                disabled={!wsConnected}
                className="w-full px-6 py-3 bg-gradient-to-r from-green-600 to-blue-600 text-white rounded-lg hover:from-green-700 hover:to-blue-700 transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {!wsConnected ? 'Connecting...' : 'Join Game'}
              </button>
            )}
          </div>
        </div>

        {/* Connection Status */}
        <div className="bg-white/5 rounded-xl p-4 border border-white/10">
          <h3 className="text-lg font-medium text-white mb-2">Connection Status</h3>
          <div className="flex items-center gap-2">
            <div className={`w-3 h-3 rounded-full ${wsConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'}`}></div>
            <span className={wsConnected ? 'text-green-400' : 'text-red-400'}>
              {wsConnected ? 'Connected via WebSocket' : 'Disconnected - Attempting to reconnect...'}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
};
