package app

import (
	"server/config"
	"server/internal/database"
	"server/internal/events"
	"server/internal/handlers/middleware"
	"server/internal/logger"
	"server/internal/repositories"
	"server/internal/services"
	"server/internal/websockets"

	userController "server/internal/controllers/users"
)

type App struct {
	Database   database.DB
	Middleware middleware.Middleware
	Websocket  *websockets.Manager
	EventBus   *events.EventBus
	Config     config.Config

	// Services
	TransactionService *services.TransactionService

	// Repositories
	UserRepo repositories.UserRepository

	// Controllers
	UserController *userController.UserController
}

func New() (*App, error) {
	log := logger.New("app").Function("New")

	config, err := config.InitConfig()
	if err != nil {
		return &App{}, log.Err("failed to initialize config", err)
	}

	db, err := database.New(config)
	if err != nil {
		return &App{}, log.Err("failed to create database", err)
	}

	eventBus := events.New(db.Cache.Events, config)

	// Initialize services
	transactionService := services.NewTransactionService(db)

	// Initialize repositories
	userRepo := repositories.New(db)

	// Initialize controllers with repositories and services
	middleware := middleware.New(db, eventBus, config, userRepo)
	userController := userController.New(eventBus, userRepo, config)

	websocket, err := websockets.New(db, eventBus, config)
	if err != nil {
		return &App{}, log.Err("failed to create websocket manager", err)
	}

	app := &App{
		Database:           db,
		Config:             config,
		Middleware:         middleware,
		TransactionService: transactionService,
		UserRepo:           userRepo,
		UserController:     userController,
		Websocket:          websocket,
		EventBus:           eventBus,
	}

	if err := app.validate(); err != nil {
		return &App{}, log.Err("failed to validate app", err)
	}

	return app, nil
}

func (a *App) validate() error {
	log := logger.New("app").Function("validate")
	if a.Database.SQL == nil {
		return log.ErrMsg("database is nil")
	}

	if a.Config == (config.Config{}) {
		return log.ErrMsg("config is nil")
	}

	nilChecks := []any{
		a.Websocket,
		a.EventBus,
		a.TransactionService,
		a.UserController,
		a.Middleware,
		a.UserRepo,
	}

	for _, check := range nilChecks {
		if check == nil {
			return log.ErrMsg("nil check failed")
		}
	}

	return nil
}

func (a *App) Close() (err error) {
	if a.EventBus != nil {
		if closeErr := a.EventBus.Close(); closeErr != nil {
			err = closeErr
		}
	}

	if dbErr := a.Database.Close(); dbErr != nil {
		err = dbErr
	}

	return err
}
