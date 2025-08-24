package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseUUIDModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                        json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                        json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}

type BaseModel struct {
	ID        int            `gorm:"type:int;primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                    json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                    json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}
