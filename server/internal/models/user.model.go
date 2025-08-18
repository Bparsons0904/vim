package models

import (
	"gorm.io/gorm"
)

type User struct {
	BaseUUIDModel
	FirstName   string  `gorm:"type:text"                      json:"firstName"`
	LastName    string  `gorm:"type:text"                      json:"lastName"`
	DisplayName string  `gorm:"type:text"                      json:"displayName"`
	Email       *string `gorm:"type:text"                      json:"email"`
	Login       string  `gorm:"type:text;uniqueIndex;not null" json:"login"`
	Password    string  `gorm:"type:text;not null"             json:"-"`
	IsAdmin     bool    `gorm:"type:bool;default:false"        json:"isAdmin"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.DisplayName == "" {
		u.DisplayName = u.FirstName + " " + u.LastName
	}
	return nil
}
