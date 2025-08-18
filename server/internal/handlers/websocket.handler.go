package handlers

import (
	"server/internal/websockets"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func WebSocketHandler(router fiber.Router, wsManager *websockets.Manager) {
	router.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.Get("/ws", websocket.New(func(c *websocket.Conn) {
		wsManager.HandleWebSocket(c)
	}))
}
