package handlers

import (
	"server/internal/app"
	"server/internal/handlers/middleware"
	"server/internal/logger"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type Handler struct {
	middleware middleware.Middleware
	log        logger.Logger
	router     fiber.Router
}

func Router(router fiber.Router, app *app.App) (err error) {
	setupWebSocketRoute(router, app)

	api := router.Group("/api")
	HealthHandler(api, app.Config)
	NewUserHandler(*app, api).Register()
	NewLoadTestHandler(*app, api).Register()

	return nil
}

func setupWebSocketRoute(router fiber.Router, app *app.App) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		app.Websocket.HandleWebSocket(c)
	}))
}
