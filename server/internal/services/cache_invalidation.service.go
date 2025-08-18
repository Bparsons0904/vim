package services

import (
	"server/internal/events"
	"server/internal/logger"
	// . "server/internal/models"
)

type CacheInvalidationService struct {
	eventBus *events.EventBus
	log      logger.Logger
}

func NewCacheInvalidationService(
	eventBus *events.EventBus,
) *CacheInvalidationService {
	return &CacheInvalidationService{
		eventBus: eventBus,
		log:      logger.New("CacheInvalidationService"),
	}
}

