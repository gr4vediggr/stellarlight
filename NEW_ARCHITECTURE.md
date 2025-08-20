# New Architecture Design

## Overview

You're absolutely right! The current architecture mixes concerns and creates confusion. Here's the new clean separation:

## New Architecture

```
WebSocketContext (Connection Management)
├── LobbyContext (Lobby State)
└── GameContext (Game State)
```

### 1. WebSocketContext
- **Single Responsibility**: Manages WebSocket connection lifecycle
- **Provides**: Connection status, message subscription, command sending
- **Used by**: Both LobbyContext and GameContext

### 2. LobbyContext  
- **Single Responsibility**: Manages lobby state and player interactions
- **Subscribes to**: Lobby messages from WebSocket
- **Manages**: Player list, ready states, game settings, lobby chat
- **Transitions to**: Game phase when game starts

### 3. GameContext
- **Single Responsibility**: Manages actual game state and gameplay
- **Subscribes to**: Game messages from WebSocket  
- **Manages**: Game state, player actions, turn management
- **Independent of**: Lobby concerns (players come from game state)

## Flow

```
Dashboard → Join Game → Lobby → Start Game → Game Page
     ↓         ↓         ↓         ↓          ↓
    None → WebSocket → Lobby → Game → Game
                   ↓    Context  Context
                 Lobby
                Context
```

## Key Benefits

1. **Clear Separation**: Each context has a single responsibility
2. **Reusable WebSocket**: One connection serves both phases  
3. **Independent State**: Lobby and Game don't interfere with each other
4. **Easy Testing**: Each context can be tested in isolation
5. **Maintainable**: Changes to lobby don't affect game logic and vice versa

## Implementation Strategy

1. **WebSocketContext**: Already created ✅
2. **LobbyContext**: Already created ✅  
3. **GameContext**: Already created ✅
4. **Update App.tsx**: Route-specific provider wrapping
5. **Update Pages**: Use appropriate contexts in each page

## Route-Specific Providers

```tsx
// Dashboard - No WebSocket needed
<Route path="/dashboard" element={<DashboardPage />} />

// Lobby - WebSocket + Lobby
<Route path="/lobby" element={
  <WebSocketProvider>
    <LobbyProvider>
      <LobbyPage />
    </LobbyProvider>
  </WebSocketProvider>
} />

// Game - WebSocket + Game  
<Route path="/game" element={
  <WebSocketProvider>
    <GameProvider>
      <GamePage />
    </GameProvider>
  </WebSocketProvider>
} />
```

This design gives you clean separation of concerns while maintaining efficient WebSocket usage!
