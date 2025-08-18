package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/gr4vediggr/stellarlight/internal/auth"
	"github.com/gr4vediggr/stellarlight/internal/config"
	"github.com/gr4vediggr/stellarlight/internal/database"
	"github.com/gr4vediggr/stellarlight/internal/game/session"
	"github.com/gr4vediggr/stellarlight/internal/users"
	"github.com/gr4vediggr/stellarlight/internal/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close(context.Background())

	// Initialize repositories and services
	userRepo := database.NewPostgresUserStore(db)
	authService := auth.NewService(userRepo, cfg.JWTSecret)

	// Initialize game session manager
	sessionManager := session.NewSessionManager()
	// Start cleanup routine for expired sessions
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			sessionManager.CleanupExpiredSessions()
		}
	}()

	// Initialize WebSocket hub with session manager
	hub := websocket.NewHub(sessionManager)
	go hub.Run()

	// Initialize WebSocket handler
	wsHandler := websocket.NewSessionHandler(hub, authService)

	e := setupHttpServer(cfg)

	// Auth routes
	setupAuthRoutes(e, authService)

	// Game routes
	registerGameRoutes(e, sessionManager, authService)

	// WebSocket route
	e.GET("/ws", wsHandler.HandleWebSocket)

	addr := net.JoinHostPort("", strconv.Itoa(cfg.Port))

	log.Printf("Starting server on %s", addr)

	log.Printf("Using TLS with cert: %s, key: %s", cfg.TLS.CertFile, cfg.TLS.KeyFile)

	if err := e.StartTLS(addr, cfg.TLS.CertFile, cfg.TLS.KeyFile); err != nil {
		log.Fatalf("Server failed to start: %v", err)

	}

}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
func setupHttpServer(cfg config.Config) *echo.Echo {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Enable CORS for all origins and methods
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: cfg.AllowedOrigins,
		AllowMethods: []string{
			echo.GET, echo.POST, echo.PUT, echo.DELETE, echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
	}))
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestId := c.Response().Header().Get(echo.HeaderXRequestID)
			fmt.Printf("REQUEST: uri: %v, status: %v, request-id: %v\n", v.URI, v.Status, requestId)
			return nil
		},
	}))

	return e
}

// setupAuthRoutes registers HTTP routes for authentication
func setupAuthRoutes(e *echo.Echo, authService *auth.AuthService) {
	h := auth.NewHandler(authService)

	apiGroup := e.Group("/api")
	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)
		authGroup.POST("/logout", h.Logout)
	}

	userGroup := apiGroup.Group("/users", auth.RequireAuth(authService))
	{
		userGroup.PUT("/update-profile", h.UpdateProfile)
	}
}

// getUserFromContext helper function to get user from context
func getUserFromContext(c echo.Context, authService *auth.AuthService) (*users.User, error) {
	userID, ok := c.Get("userID").(uuid.UUID)
	if !ok {
		return nil, echo.NewHTTPError(400, "Invalid user ID")
	}

	user, err := authService.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return nil, echo.NewHTTPError(400, "User not found")
	}

	return user, nil
}

// registerGameRoutes registers HTTP routes for game management
func registerGameRoutes(e *echo.Echo, sessionManager *session.SessionManager, authService *auth.AuthService) {
	gameGroup := e.Group("/api/game")
	gameGroup.Use(auth.RequireAuth(authService))

	gameGroup.POST("/create", func(c echo.Context) error {
		user, err := getUserFromContext(c, authService)
		if err != nil {
			return err
		}

		session, err := sessionManager.CreateSession(user)
		if err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]interface{}{
			"session_id":  session.ID,
			"invite_code": session.InviteCode,
			"state":       session.State,
		})
	})

	gameGroup.POST("/join", func(c echo.Context) error {
		user, err := getUserFromContext(c, authService)
		if err != nil {
			return err
		}

		var req struct {
			InviteCode string `json:"invite_code"`
		}
		if err := c.Bind(&req); err != nil {
			return c.JSON(400, map[string]string{"error": "Invalid request"})
		}

		session, err := sessionManager.JoinSession(user, req.InviteCode)
		if err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]interface{}{
			"session_id":  session.ID,
			"invite_code": session.InviteCode,
			"state":       session.State,
		})
	})

	gameGroup.POST("/leave", func(c echo.Context) error {
		user, err := getUserFromContext(c, authService)
		if err != nil {
			return err
		}

		if err := sessionManager.LeaveSession(user.ID); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]string{"message": "Left game successfully"})
	})

	gameGroup.GET("/current", func(c echo.Context) error {
		user, err := getUserFromContext(c, authService)
		if err != nil {
			return err
		}

		session, err := sessionManager.GetPlayerSession(user.ID)
		if err != nil {
			return c.JSON(404, map[string]string{"error": "Not in any game"})
		}

		return c.JSON(200, map[string]interface{}{
			"session_id":  session.ID,
			"invite_code": session.InviteCode,
			"state":       session.State,
		})
	})

	gameGroup.POST("/start", func(c echo.Context) error {
		user, err := getUserFromContext(c, authService)
		if err != nil {
			return err
		}

		session, err := sessionManager.GetPlayerSession(user.ID)
		if err != nil {
			return c.JSON(404, map[string]string{"error": "Not in any game"})
		}

		if err := session.StartGame(); err != nil {
			return c.JSON(400, map[string]string{"error": err.Error()})
		}

		return c.JSON(200, map[string]string{"message": "Game started successfully"})
	})
}
