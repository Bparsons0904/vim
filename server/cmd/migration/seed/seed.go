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
			FirstName:   "Bob",
			LastName:    "Parsons",
			DisplayName: "Bob Parsons",
			Email:       stringPtr("bob.parsons@example.com"),
			Login:       "deadstyle",
			Password:    "password",
			IsAdmin:     true,
		}, {
			FirstName:   "Bob",
			LastName:    "Covell",
			DisplayName: "Bitch Ass",
			Email:       stringPtr("bob.covell@example.com"),
			Login:       "bobb",
			Password:    "password",
			IsAdmin:     true,
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
