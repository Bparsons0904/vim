package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseUUIDModel struct {
	ID        string         `gorm:"type:varchar(64);primaryKey" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"              json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"              json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                       json:"deletedAt"`
}

func (b *BaseUUIDModel) BeforeSave(tx *gorm.DB) error {
	if b.ID == "" {
		uuidString, _ := uuid.NewV7()
		b.ID = uuidString.String()
	}
	return nil
}

type BaseModel struct {
	ID        int            `gorm:"type:int;primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                    json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"                    json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"                             json:"deletedAt"`
}
