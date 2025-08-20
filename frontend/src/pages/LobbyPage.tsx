import { useState, useEffect } from 'react';
import { useGame } from '../core/game/GameContext';
import { useAuth } from '../core/auth/AuthContext';
import { Settings, Users, Play, UserCheck, UserX, Palette, Copy, ArrowLeft } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import type { Player } from '../core/game/GameService';

const PLAYER_COLORS = [
  '#3b82f6', // blue
  '#ef4444', // red
  '#10b981', // emerald
  '#f59e0b', // amber
  '#8b5cf6', // violet
  '#ec4899', // pink
  '#06b6d4', // cyan
  '#84cc16', // lime
];

export const LobbyPage = () => {
  const { user } = useAuth();
  const {
    currentSession,
    lobbyState,
    loading,
    error,
    leaveGame,
    setReady,
    setColor,
    updateSettings,
    startGame,
  } = useGame();
  
  const navigate = useNavigate();
  const [showSettings, setShowSettings] = useState(false);
  const [settings, setSettings] = useState({
    numStars: 150,
    shape: 'spiral',
    maxHyperlanes: 5,
    hyperlaneConnectivity: 3,
  });

  // If no session, redirect to dashboard
  useEffect(() => {
    console.log('🎮 LobbyPage: Checking session state:', { currentSession: !!currentSession, lobbyState: !!lobbyState, loading });
    
    // Only redirect if we definitely have no session AND we're not loading
    // The loading check ensures GameContext has finished initializing
    if (!loading && !currentSession) {
      console.log('🎮 LobbyPage: No session found, redirecting to dashboard');
      navigate('/dashboard');
    }
  }, [currentSession, lobbyState, loading, navigate]);

  // Show loading while GameContext initializes
  if (loading || (!currentSession && !lobbyState)) {
    console.log('🎮 LobbyPage: Showing loading - waiting for GameContext initialization');
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-indigo-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
          <p className="text-white">Initializing...</p>
        </div>
      </div>
    );
  }

  // If we have a session but no lobby state, show loading (waiting for automatic lobby state)
  if (currentSession && !lobbyState) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-indigo-900 flex items-center justify-center">
        <div className="text-center bg-white/10 backdrop-blur-sm border border-white/20 rounded-xl p-8 max-w-md">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
          <h2 className="text-2xl font-bold text-white mb-4">Connecting to Lobby</h2>
          <p className="text-blue-200 mb-4">
            Establishing connection to your game lobby...
          </p>
          <p className="text-sm text-blue-300 mb-6">
            Session ID: {currentSession.sessionId}
          </p>
          <button
            onClick={() => navigate('/dashboard')}
            className="w-full bg-gray-600 hover:bg-gray-700 text-white font-medium py-3 px-6 rounded-lg transition-colors"
          >
            Back to Dashboard
          </button>
        </div>
      </div>
    );
  }

  // If no session at all, this will be handled by the useEffect redirect
  if (!currentSession) {
    return null; // This should not happen due to useEffect redirect
  }

  // If we don't have lobbyState, it means we're not connected to the lobby yet
  if (!lobbyState) {
    // This case is handled above in the connect interface
    return null;
  }

  const currentPlayer = lobbyState.players.find((p: Player) => p.userId === user?.id);
  const isHost = currentPlayer?.isHost || false;
  const canStartGame = isHost && lobbyState.players.every((p: Player) => p.isReady) && lobbyState.players.length > 0;
  console.log(lobbyState.players, user?.id)
  const handleLeaveGame = async () => {
    await leaveGame();
    navigate('/dashboard');
  };

  const handleToggleReady = () => {
    console.log(currentPlayer, "cp")
    if (currentPlayer) {
      setReady(!currentPlayer.isReady);
    }
  };

  const handleColorSelect = (color: string) => {
    setColor(color);
  };

  const handleUpdateSettings = () => {
    updateSettings(settings);
    setShowSettings(false);
  };

  const copyInviteCode = async () => {
    try {
      await navigator.clipboard.writeText(currentSession.inviteCode);
      // Could show a toast notification here
    } catch (error) {
      console.error('Failed to copy invite code:', error);
    }
  };
  console.log(currentSession);
  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-blue-900 to-indigo-900 p-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-6">
          <div className="flex items-center gap-4">
            <button
              onClick={handleLeaveGame}
              className="flex items-center gap-2 px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition duration-200"
            >
              <ArrowLeft size={20} />
              Leave Game
            </button>
            <div>
              <h1 className="text-3xl font-bold text-white">Game Lobby</h1>
              <p className="text-gray-300">Waiting for players...</p>
            </div>
          </div>
          
          {isHost && (
            <button
              onClick={() => setShowSettings(true)}
              className="flex items-center gap-2 px-4 py-2 bg-white/10 text-white rounded-lg hover:bg-white/20 transition duration-200"
            >
              <Settings size={20} />
              Game Settings
            </button>
          )}
        </div>

        {/* Invite Code */}
        <div className="bg-white/10 backdrop-blur-md rounded-xl p-4 mb-6 border border-white/20">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-medium text-white mb-1">Invite Code</h3>
              <p className="text-gray-300">Share this code with friends to join the game</p>
            </div>
            <div className="flex items-center gap-3">
              <code className="text-xl font-mono bg-black/30 px-4 py-2 rounded text-white">
                {currentSession.inviteCode}
              </code>
              <button
                onClick={copyInviteCode}
                className="p-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition duration-200"
              >
                <Copy size={20} />
              </button>
            </div>
          </div>
        </div>

        <div className="grid lg:grid-cols-3 gap-6">
          {/* Players List */}
          <div className="lg:col-span-2 bg-white/10 backdrop-blur-md rounded-xl p-6 border border-white/20">
            <div className="flex items-center gap-3 mb-4">
              <Users className="w-6 h-6 text-blue-400" />
              <h2 className="text-xl font-semibold text-white">
                Players ({lobbyState.players.length})
              </h2>
            </div>
            
            <div className="space-y-3">
              {lobbyState.players.map((player: Player) => (
                <div
                  key={player.userId}
                  className="flex items-center justify-between p-3 bg-white/5 rounded-lg border border-white/10"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className="w-4 h-4 rounded-full border-2 border-white/30"
                      style={{ backgroundColor: player.color || '#6b7280' }}
                    ></div>
                    <span className="text-white font-medium">{player.displayName}</span>
                    {player.isHost && (
                      <span className="text-xs bg-yellow-500/20 text-yellow-300 px-2 py-1 rounded">
                        HOST
                      </span>
                    )}
                  </div>
                  
                  <div className="flex items-center gap-2">
                    {player.isReady ? (
                      <div className="flex items-center gap-1 text-green-400">
                        <UserCheck size={16} />
                        <span className="text-sm">Ready</span>
                      </div>
                    ) : (
                      <div className="flex items-center gap-1 text-gray-400">
                        <UserX size={16} />
                        <span className="text-sm">Not Ready</span>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Player Controls */}
          <div className="space-y-6">
            {/* Ready Status */}
            <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 border border-white/20">
              <h3 className="text-lg font-medium text-white mb-4">Your Status</h3>
              <button
                onClick={handleToggleReady}
                className={`w-full px-6 py-3 rounded-lg font-medium transition duration-200 ${
                  currentPlayer?.isReady
                    ? 'bg-red-600 hover:bg-red-700 text-white'
                    : 'bg-green-600 hover:bg-green-700 text-white'
                }`}
              >
                {currentPlayer?.isReady ? 'Cancel Ready' : 'Ready Up'}
              </button>
            </div>

            {/* Color Selection */}
            <div className="bg-white/10 backdrop-blur-md rounded-xl p-6 border border-white/20">
              <div className="flex items-center gap-2 mb-4">
                <Palette size={20} className="text-purple-400" />
                <h3 className="text-lg font-medium text-white">Choose Color</h3>
              </div>
              <div className="grid grid-cols-4 gap-2">
                {PLAYER_COLORS.map((color) => {
                  const isSelected = currentPlayer?.color === color;
                  const isUsed = lobbyState.players.some((p: any) => p.color === color && p.playerId !== user?.id);
                  
                  return (
                    <button
                      key={color}
                      onClick={() => handleColorSelect(color)}
                      disabled={isUsed}
                      className={`w-12 h-12 rounded-lg border-2 transition duration-200 ${
                        isSelected 
                          ? 'border-white shadow-lg' 
                          : isUsed 
                            ? 'border-gray-600 opacity-50 cursor-not-allowed' 
                            : 'border-gray-400 hover:border-white'
                      }`}
                      style={{ backgroundColor: color }}
                    >
                      {isSelected && <div className="w-full h-full rounded-md border-2 border-white/50"></div>}
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Start Game Button (Host Only) */}
            {isHost && (
              <button
                onClick={startGame}
                disabled={!canStartGame || loading}
                className="w-full flex items-center justify-center gap-2 px-6 py-4 bg-gradient-to-r from-green-600 to-blue-600 text-white rounded-lg hover:from-green-700 hover:to-blue-700 transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
              >
                <Play size={20} />
                {loading ? 'Starting Game...' : 'Start Game'}
              </button>
            )}
          </div>
        </div>

        {/* Error Display */}
        {error && (
          <div className="mt-6 p-4 bg-red-500/20 border border-red-500/30 rounded-lg">
            <p className="text-red-200">{error}</p>
          </div>
        )}

        {/* Game Settings Modal */}
        {showSettings && isHost && (
          <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
            <div className="bg-slate-800 rounded-xl p-6 max-w-md w-full border border-white/20">
              <h3 className="text-xl font-bold text-white mb-4">Game Settings</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1">
                    Number of Stars
                  </label>
                  <input
                    type="range"
                    min="50"
                    max="300"
                    step="25"
                    value={settings.numStars}
                    onChange={(e) => setSettings(prev => ({ ...prev, numStars: parseInt(e.target.value) }))}
                    className="w-full"
                  />
                  <span className="text-white text-sm">{settings.numStars} stars</span>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1">
                    Galaxy Shape
                  </label>
                  <select
                    value={settings.shape}
                    onChange={(e) => setSettings(prev => ({ ...prev, shape: e.target.value }))}
                    className="w-full px-3 py-2 bg-white/10 border border-white/20 rounded text-white"
                  >
                    <option value="spiral">Spiral</option>
                    <option value="elliptical">Elliptical</option>
                    <option value="irregular">Irregular</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300 mb-1">
                    Max Hyperlanes per System
                  </label>
                  <input
                    type="range"
                    min="3"
                    max="8"
                    value={settings.maxHyperlanes}
                    onChange={(e) => setSettings(prev => ({ ...prev, maxHyperlanes: parseInt(e.target.value) }))}
                    className="w-full"
                  />
                  <span className="text-white text-sm">{settings.maxHyperlanes}</span>
                </div>
              </div>

              <div className="flex gap-3 mt-6">
                <button
                  onClick={handleUpdateSettings}
                  className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition duration-200"
                >
                  Apply Settings
                </button>
                <button
                  onClick={() => setShowSettings(false)}
                  className="flex-1 px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition duration-200"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
