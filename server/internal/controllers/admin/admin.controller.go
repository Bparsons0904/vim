package adminController

import (
	"context"
	"server/config"
	"server/internal/events"
	"server/internal/logger"
	"server/internal/repositories"
	"time"

	. "server/internal/models"

	"github.com/google/uuid"
)

type AdminController struct {
	userRepo repositories.UserRepository
	Config   config.Config
	log      logger.Logger
	eventBus *events.EventBus
}

func New(
	eventBus *events.EventBus,
	userRepo repositories.UserRepository,
	config config.Config,
) *AdminController {
	return &AdminController{
		userRepo: userRepo,
		Config:   config,
		log:      logger.New("AdminController"),
		eventBus: eventBus,
	}
}

type Message struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Channel   string         `json:"channel,omitempty"`
	Action    string         `json:"action,omitempty"`
	UserID    string         `json:"userId,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
}

func (c *AdminController) SendBroadcast(ctx context.Context, user User, message string) {
	log := c.log.Function("SendBroadcast")

	event := events.Event{
		ID:        uuid.New().String(),
		Type:      "admin",
		Channel:   "admin",
		UserID:    user.ID,
		Data:      map[string]any{"message": message},
		Timestamp: time.Now(),
	}

	log.Info("Broadcasting user login event", "message", message, "userID", user.ID)
	if err := c.eventBus.Publish("broadcast", event); err != nil {
		log.Er("failed to publish event", err, "event", event)
		return
	}
}
