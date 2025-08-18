package initialize

import (
	"server/config"
	"server/internal/logger"

	"gorm.io/gorm"
)

func InitializeTables(db *gorm.DB, config config.Config, log logger.Logger) error {
	log = log.Function("InitializeTables")
	log.Info("Initializing essential production data")

	log.Info("Table initialization complete")
	return nil
}

