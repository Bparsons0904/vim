package seed

import (
	"server/config"
	"server/internal/logger"
	. "server/internal/models"

	"gorm.io/gorm"
)

func stringPtr(s string) *string {
	return &s
}

func Seed(db *gorm.DB, config config.Config, log logger.Logger) error {
	log = log.Function("seed")
	log.Info("Seeding development data")

	users := []User{
		{
			FirstName:   "Admin",
			LastName:    "User",
			DisplayName: "Administrator",
			Email:       stringPtr("admin@example.com"),
			Login:       "admin",
			Password:    "password",
			IsAdmin:     true,
		}, {
			FirstName:   "Test",
			LastName:    "User",
			DisplayName: "Test User",
			Email:       stringPtr("test@example.com"),
			Login:       "test",
			Password:    "password",
			IsAdmin:     false,
		}, {
			FirstName:   "Ada",
			LastName:    "Lovelace",
			Email:       stringPtr("ada.lovelace@example.com"),
			DisplayName: "Ada Lovelace",
			Login:       "ada",
			Password:    "password",
			IsAdmin:     false,
		},
	}

	for _, user := range users {
		var existingUser User
		if err := db.First(&existingUser, "login = ?", user.Login).Error; err == nil {
			log.Info("User already exists", "user", user)
			continue
		}
		log.Info("Seeding user", "user", user)
		if err := db.Create(&user).Error; err != nil {
			log.Er("failed to create user", err, "user", user)
		}
	}

	return nil
}
