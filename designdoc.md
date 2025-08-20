# Game design 

Long term online strategy 4x game. Multiplayer with many players who play over a long period of time. 

Main points: 
- Time passes in real life and the game at a fixed rate. 
- Massively online
- Browser based.

Games can take months. As 1 hour in real life progresses the game by 1 year (could be changed). 

Should respect the players time by implementing queues, and not requiring the player to be online / watching all the time. 
Expected play time, 30m/day-1h/day should be sufficient

## The universe (game world)

Galaxy. 

Star systems. 
- Can be owned by players
- Connected to other star systems via hyperlanes. 


Planets
- Varying levels of habitability. 
- Can have resources.
- Need infrastructure to extract resources. 
- Can be terraformed

Colony
- Planets can be colonized if above certain level of habitability. 
- Owned by players.
- Have the following properties: 
   - Industries: 
        List of buildings that can be worked by the population. Each industry can have a level. 
        Limited by planet size and planet population
        Produces resources. 
   - Population: 
        Grows over time. 
        Can resettle to new colonies (automatic).

   - Defenses: 
        Can have an impact on fleet battles in the system. 

- Can be conquered. 

Fleets: 
- Consist of multiple ships. 
- If two players fleets meet in the same star system, and they are hostile, it starts a battle. 
- If sent into a hostile star system without a fleet, it starts a battle. 

Battles:
- Take place over many real time hours. 
- Can be reinforced. 

# Technical implementation

Back-end: 

- Golang microservices

Front-end:
- Web browser  (react?)

# Game systems

ColonyUpdateSystem 
- Responsible for updating the colony
- Progresses building of buildings. 
- Keeps track of what is being build. 
- Updates population, expected resource outputs 

EmpireSystem 
- Updates resources
- Manages Empire build queues (dispatches to colonies) 
    Players can queue up (even if they dont have resources yet) a build order. 
    These items will start building the moment resources become available. 

FleetBuildSystem: 
- Manages fleets in construction

FleetUpdateSystem: 
- Manages fleets in movement 

BattleSystem: 
- Checks if fleets should meet in battle. 
- For all battles, updates the battle progress. 

UserInteractionSystem: 
- If players are logged in, they should receive updates live from the game server
- They can send commands, which will be validated by the game server. 
- The game client will read from this system also about updates to the game world/empire state

MessageSystem: 
- Players should be able to message / create chats with other players. 
- There could also be a global chat. 

UserNotificationSystem: 
- When players are online, messages should come in as they happend.
- When players are offline, a log should be created for all messages and presented to them upon login. 

VisionSystem: 
- Used when players are online. Should update the vision and send items that players do not own, but can see. 

(And more..., Terraforming, Etc.)

# Persistence: PostgresDB. 



## Session flow

:::mermaid
flowchart TD
    %% Login Flow
    A[Player Opens Web Game] --> B[Login Page]
    B -->|POST /api/auth/login| C{Login Successful?}
    C -->|No| B
    C -->|Yes| D[Dashboard]

    %% Dashboard Options
    D --> E[View / Modify Profile & Stats]
    D --> F[Check Current Game<br/>GET /api/game/current]

    %% If NOT in a game
    F -->|No Current Game| G[Create New Game<br/>POST /api/game/create]
    F -->|No Current Game| H[Join Existing Game<br/>POST /api/game/join]
    F -->|No Current Game| I[Matchmaking<br/>Server Assigns Game]

    %% If already in a game
    F -->|Current Game Exists| J[Continue Game]
    F -->|Current Game Exists| K[Quit Game]

    %% Game Lobby
    G --> L[Game Lobby<br/>WebSocket Connection]
    H --> L
    I --> L
    J --> L

    %% From Lobby to Gameplay
    L --> M[Gameplay]
::: 