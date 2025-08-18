package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gr4vediggr/stellarlight/internal/auth"
	"github.com/gr4vediggr/stellarlight/internal/config"
	"github.com/gr4vediggr/stellarlight/internal/database"
	"github.com/gr4vediggr/stellarlight/internal/game/lobby"
	"github.com/gr4vediggr/stellarlight/internal/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/sync/errgroup"
)

var ()

type app struct {
	config config.Config

	server *echo.Echo
	db     *pgx.Conn
	// Dependencies
	authService  *auth.AuthService
	hub          *websocket.Hub
	lobbyManager *lobby.Manager // Assuming you have a lobby manager
}

func (a *app) initialize() error {
	var err error
	a.config, err = config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	a.db, err = pgx.Connect(context.Background(), a.config.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	userRepo := database.NewPostgresUserStore(a.db)
	a.authService = auth.NewService(userRepo, a.config.JWTSecret)

	a.lobbyManager = lobby.NewManager()
	a.hub = websocket.NewHub(a.authService, a.lobbyManager)
	a.server, err = a.setupHttpServer()
	if err != nil {
		return fmt.Errorf("failed to setup HTTP server: %w", err)
	}

	return nil
}

func main() {
	app := &app{}
	if err := app.initialize(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		addr := net.JoinHostPort("", strconv.Itoa(app.config.Port))

		log.Printf("Starting server on %s", addr)
		var err error

		log.Printf("Using TLS with cert: %s, key: %s", app.config.TLS.CertFile, app.config.TLS.KeyFile)
		err = app.server.StartTLS(addr, app.config.TLS.CertFile, app.config.TLS.KeyFile)

		if err != nil {
			log.Printf("Server failed to start: %v", err)
			return err
		}
		return nil
	})

	<-ctx.Done()

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := app.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}

	err := eg.Wait()
	if err != nil {
		log.Printf("Error during shutdown: %v", err)
	} else {
		log.Println("Server shutdown gracefully")
	}

}

func (app *app) setupHttpServer() (*echo.Echo, error) {
	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}

	// Enable CORS for all origins and methods
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: app.config.AllowedOrigins,
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

	// Routes
	app.setupRouter(e)

	return e, nil
}

func (app *app) setupRouter(e *echo.Echo) {
	h := auth.NewHandler(app.authService)

	apiGroup := e.Group("/api")

	authGroup := apiGroup.Group("/auth")
	{
		authGroup.POST("/register", h.Register)
		authGroup.POST("/login", h.Login)
		authGroup.POST("/refresh", h.RefreshToken)
	}

	userGroup := apiGroup.Group("/users", auth.RequireAuth(app.authService))
	{
		userGroup.PUT("/update-profile", h.UpdateProfile)
	}

	handler := websocket.NewHandler(app.hub, app.authService)

	// Add other routes
	e.GET("/ws", handler.HandleWebSocket)
}

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
