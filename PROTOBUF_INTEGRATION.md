# Protobuf WebSocket Integration

This document outlines the refactored WebSocket messaging system using Protocol Buffers for efficient binary communication between the client and server.

## Architecture Overview

### Message Separation
- **Commands** (Client → Server): Defined in `client_commands.proto`
- **Messages** (Server → Client): Defined in `server_messages.proto`

### Key Components

#### 1. Protocol Buffer Schemas

**Client Commands** (`shared/client_commands.proto`):
```proto
message ClientCommand {
  string player_id = 1;
  int64 timestamp = 2;
  
  oneof command {
    LobbyCommand lobby_command = 10;
    GameCommand game_command = 20;
    ChatCommand chat_command = 30;
  }
}
```

**Server Messages** (`shared/server_messages.proto`):
```proto
message ServerMessage {
  int64 timestamp = 1;
  
  oneof message {
    LobbyMessage lobby_message = 10;
    GameMessage game_message = 20;
    ChatMessage chat_message = 30;
    SystemMessage system_message = 40;
    ErrorMessage error_message = 50;
  }
}
```

#### 2. Frontend Services

**ProtobufWebSocketService** (`frontend/src/core/messaging/ProtobufWebSocketService.ts`):
- Handles binary WebSocket communication
- Encodes/decodes protobuf messages
- Provides type-safe message handling
- Supports command sending and message receiving

**LobbyService** (`frontend/src/core/lobby/LobbyService.ts`):
- Dedicated lobby state management
- Event-driven architecture for lobby updates
- Handles player actions (ready, color, settings)
- Provides reactive lobby state for UI components

**GameService** (`frontend/src/core/game/GameService.ts`):
- Overall game session management
- Integrates with API service for HTTP operations
- Coordinates with LobbyService for lobby functionality

#### 3. React Integration

**GameContext** (`frontend/src/core/game/GameContext.tsx`):
- Provides unified access to game and lobby services
- Manages WebSocket connections
- Handles service lifecycle and cleanup
- Exposes reactive state for React components

## Message Flow

### Lobby Operations
1. **Join Lobby**: Client sends `JoinLobbyCommand` → Server responds with `LobbyStateMessage`
2. **Set Ready**: Client sends `SetReadyCommand` → Server broadcasts `PlayerUpdatedMessage`
3. **Change Color**: Client sends `SetColorCommand` → Server broadcasts `PlayerUpdatedMessage`
4. **Update Settings**: Client sends `UpdateSettingsCommand` → Server broadcasts `LobbySettingsUpdatedMessage`
5. **Start Game**: Client sends `StartGameCommand` → Server sends `GameStartingMessage` → `GameLoadingMessage`

### Binary Message Format
```
[4 bytes: message length]
[1 byte: message type]
[remaining bytes: protobuf message data]
```

Message Types:
- `0x01`: Command (Client → Server)
- `0x02`: Message (Server → Client)
- `0x03`: Ping
- `0x04`: Pong

## Usage Examples

### Sending Lobby Commands
```typescript
// Set player ready status
lobbyService.setReady(true);

// Change player color
lobbyService.setColor('#3b82f6');

// Update game settings (host only)
lobbyService.updateSettings({
  numStars: 150,
  shape: 'spiral',
  maxHyperlanes: 5,
  hyperlaneConnectivity: 3
});
```

### Subscribing to Lobby Events
```typescript
// Subscribe to lobby state changes
const unsubscribe = lobbyService.onLobbyEvent((event) => {
  switch (event.type) {
    case 'lobby_state':
      console.log('Lobby state updated:', event.data);
      break;
    case 'player_joined':
      console.log('Player joined:', event.data);
      break;
    case 'game_starting':
      console.log('Game starting!');
      break;
  }
});

// Cleanup
unsubscribe();
```

### Using in React Components
```tsx
const LobbyPage = () => {
  const { lobbyService, lobbyState } = useGame();
  
  useEffect(() => {
    if (!lobbyService) return;
    
    const unsubscribe = lobbyService.onLobbyEvent((event) => {
      // Handle lobby events
    });
    
    return unsubscribe;
  }, [lobbyService]);
  
  return (
    <div>
      <h1>Players: {lobbyState?.players.length}</h1>
      <button onClick={() => lobbyService?.setReady(true)}>
        Ready Up
      </button>
    </div>
  );
};
```

## Benefits

1. **Type Safety**: Full TypeScript support with generated protobuf types
2. **Efficiency**: Binary encoding reduces message size and improves performance
3. **Separation of Concerns**: Clear distinction between commands and messages
4. **Scalability**: Easy to extend with new message types
5. **Reliability**: Structured message format with validation
6. **Maintainability**: Centralized message definitions and handlers

## Development Workflow

1. **Update Schemas**: Modify `.proto` files in `shared/` directory
2. **Generate Types**: Run `npm run proto:generate` in frontend
3. **Update Services**: Modify service classes to handle new message types
4. **Test Integration**: Use frontend dev server to test WebSocket communication

## Next Steps

- [ ] Backend protobuf integration (Go implementation)
- [ ] Game state synchronization using protobuf messages
- [ ] Chat system implementation
- [ ] Real-time game updates during gameplay
- [ ] Performance optimization and monitoring
