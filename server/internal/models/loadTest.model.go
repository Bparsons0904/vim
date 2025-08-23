package models

type LoadTest struct {
	BaseUUIDModel
	Rows         int     `gorm:"not null"                        json:"rows"`
	Columns      int     `gorm:"not null"                        json:"columns"`
	DateColumns  int     `gorm:"not null"                        json:"dateColumns"` // Number of date columns populated (0-10)
	Method       string  `gorm:"type:varchar(20);not null"       json:"method"` // 'brute_force' or 'optimized'
	Status       string  `gorm:"type:varchar(20);not null"       json:"status"` // 'running', 'completed', 'failed'
	CSVGenTime   *int    `gorm:"type:int"                        json:"csvGenTime"`   // milliseconds
	ParseTime    *int    `gorm:"type:int"                        json:"parseTime"`    // milliseconds
	InsertTime   *int    `gorm:"type:int"                        json:"insertTime"`   // milliseconds
	TotalTime    *int    `gorm:"type:int"                        json:"totalTime"`    // milliseconds
	ErrorMessage *string `gorm:"type:text"                       json:"errorMessage,omitempty"`
}

type CreateLoadTestRequest struct {
	Rows        int    `json:"rows" validate:"required,min=1"`
	Columns     int    `json:"columns" validate:"required,min=1"`  // Total columns (210 max: 200 regular + 10 date)
	DateColumns int    `json:"dateColumns" validate:"min=0,max=10"` // Number of date columns to populate (0-10)
	Method      string `json:"method" validate:"required,oneof=brute_force optimized"`
}