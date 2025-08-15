package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/gr4vediggr/stellarlight/internal/gen"
	"github.com/gr4vediggr/stellarlight/internal/resource"
	"github.com/gr4vediggr/stellarlight/internal/utils"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type app struct {
	// Define your application structure here
	config appconfig // Application configuration

	assets *resource.Assets
}

type appconfig struct {
	// Define your application configuration here
	assetFolder string
	port        int
}

func loadEnvironment() (appconfig, error) {
	// Load environment variables or configuration here

	assetFolder := os.Getenv("ASSET_FOLDER") // Example of loading an environment variable
	port := os.Getenv("PORT")
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return appconfig{}, fmt.Errorf("invalid PORT value: %w", err)
	}

	return appconfig{
		assetFolder: assetFolder,
		port:        portInt,
	}, nil
}

func loadFlags(appConfig appconfig) error {
	// Load command line flags or arguments here
	// For example, you can use flag package to parse command line arguments

	flag.StringVar(&appConfig.assetFolder, "assetfolder", appConfig.assetFolder, "Path to the asset folder")
	flag.IntVar(&appConfig.port, "port", appConfig.port, "Port to run the server on")
	flag.Parse()

	if appConfig.assetFolder == "" {
		return os.ErrInvalid // Return an error if asset folder is not set
	}

	return nil
}

func createApp() (*app, error) {
	appConfig, err := loadEnvironment()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	if err := loadFlags(appConfig); err != nil {
		return nil, fmt.Errorf("failed to load flags: %w", err)
	}

	assets, err := resource.LoadAssetsFromDirs([]string{appConfig.assetFolder})
	if err != nil {
		return nil, fmt.Errorf("failed to load assets: %w", err)
	}

	return &app{
		assets: assets,
		config: appConfig,
	}, nil
}

func main() {

	if err := run(); err != nil {
		fmt.Println("Error starting the application:", err)
		os.Exit(1)
	}
}

func run() error {
	app, err := createApp()
	if err != nil {
		return err
	}

	e := app.createServer()
	app.setupRoutes(e)

	if err := e.Start(net.JoinHostPort("", fmt.Sprintf("%d", app.config.port))); err != nil {
		return err
	}

	return nil
}

func (app *app) createServer() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	setupMiddleware(e)

	return e
}

func (app *app) setupRoutes(e *echo.Echo) {
	// Define your routes here
	e.GET("/", func(c echo.Context) error {
		return c.String(200, "Welcome to the Empire API")
	})
	e.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	e.POST("/galaxy-generate", func(c echo.Context) error {

		var genConfig gen.GalaxyGenerationConfig
		if err := c.Bind(&genConfig); err != nil {
			return err
		}

		builder := gen.GalaxyBuilder{
			StarTypes:   utils.MapToSlice(app.assets.StarTypes),
			PlanetTypes: utils.MapToSlice(app.assets.PlanetTypes),
		}
		galaxy, err := builder.GenerateGalaxy(genConfig)
		if err != nil {
			return err
		}

		return c.JSON(200, galaxy)

	})

}

func setupMiddleware(e *echo.Echo) {
	// skipper := func(c echo.Context) bool {
	// 	// Skip health check endpoint
	// 	return c.Request().URL.Path == "/health"
	// }
	e.Use(middleware.RequestID())
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogURI:    true,
		BeforeNextFunc: func(c echo.Context) {
			c.Set("customValueFromContext", 42)
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			requestId := c.Response().Header().Get(echo.HeaderXRequestID)
			fmt.Printf("REQUEST: uri: %v, status: %v, request-id: %v\n", v.URI, v.Status, requestId)
			return nil
		},
	}))
}
