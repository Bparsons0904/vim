package models

import "github.com/google/uuid"

type LoadTest struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuidv7()" json:"id"`
	Rows         int       `gorm:"not null"                              json:"rows"`
	Columns      int       `gorm:"not null"                              json:"columns"`
	DateColumns  int       `gorm:"not null"                              json:"dateColumns"` // Number of date columns populated (0-10)
	Method       string    `gorm:"type:varchar(20);not null"             json:"method"`      // 'brute_force', 'batched', 'plaid', 'optimized', or 'ludicrous'
	Status       string    `gorm:"type:varchar(20);not null"             json:"status"`      // 'running', 'completed', 'failed'
	CSVGenTime   *int      `gorm:"type:int"                              json:"csvGenTime"`  // milliseconds
	ParseTime    *int      `gorm:"type:int"                              json:"parseTime"`   // milliseconds
	InsertTime   *int      `gorm:"type:int"                              json:"insertTime"`  // milliseconds
	TotalTime    *int      `gorm:"type:int"                              json:"totalTime"`   // milliseconds
	ErrorMessage *string   `gorm:"type:text"                             json:"errorMessage,omitempty"`
}

type CreateLoadTestRequest struct {
	Rows   int    `json:"rows"   validate:"required,min=1"`
	Method string `json:"method" validate:"required,oneof=brute_force batched plaid"`
	// Note: Columns and DateColumns are ignored - we use a fixed structure:
	// - 5 date columns (birth_date, start_date, end_date, created_at, updated_at)
	// - 20 regular columns (col1-col20)
	// - Total: 25 columns
}

