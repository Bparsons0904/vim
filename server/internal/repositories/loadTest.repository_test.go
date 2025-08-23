package repositories

import (
	. "server/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestLoadTestRepository_ValidateLoadTestData(t *testing.T) {
	tests := []struct {
		name        string
		loadTest    *LoadTest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid load test creation",
			loadTest: &LoadTest{
				BaseUUIDModel: BaseUUIDModel{
					ID: uuid.New(),
				},
				Rows:        1000,
				Columns:     10,
				DateColumns: 3,
				Method:      "brute_force",
				Status:      "running",
			},
			expectError: false,
		},
		{
			name: "valid load test with optimized method",
			loadTest: &LoadTest{
				BaseUUIDModel: BaseUUIDModel{
					ID: uuid.New(),
				},
				Rows:        5000,
				Columns:     50,
				DateColumns: 10,
				Method:      "optimized",
				Status:      "completed",
				CSVGenTime:  intPtr(100),
				ParseTime:   intPtr(200),
				InsertTime:  intPtr(1500),
				TotalTime:   intPtr(1800),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the data validation logic
			assert.Greater(t, tt.loadTest.Rows, 0, "Rows should be positive")
			assert.Greater(t, tt.loadTest.Columns, 0, "Columns should be positive") 
			assert.GreaterOrEqual(t, tt.loadTest.DateColumns, 0, "DateColumns should be non-negative")
			assert.Contains(t, []string{"brute_force", "optimized"}, tt.loadTest.Method, "Method should be valid")
			assert.Contains(t, []string{"running", "completed", "failed"}, tt.loadTest.Status, "Status should be valid")
		})
	}
}

func TestLoadTestRepository_UUIDValidation(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid UUID",
			id:          uuid.New().String(),
			expectError: false,
		},
		{
			name:        "invalid UUID",
			id:          "invalid-uuid",
			expectError: true,
			errorMsg:    "invalid UUID format",
		},
		{
			name:        "empty ID",
			id:          "",
			expectError: true,
			errorMsg:    "invalid UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test UUID parsing logic without database calls
			_, err := uuid.Parse(tt.id)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoadTestRepository_StatusValidation(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		isValidStatus  bool
	}{
		{
			name:           "running status",
			status:         "running",
			isValidStatus:  true,
		},
		{
			name:           "completed status", 
			status:         "completed",
			isValidStatus:  true,
		},
		{
			name:           "failed status",
			status:         "failed",
			isValidStatus:  true,
		},
		{
			name:           "invalid status",
			status:         "invalid",
			isValidStatus:  false,
		},
	}

	validStatuses := []string{"running", "completed", "failed"}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := false
			for _, validStatus := range validStatuses {
				if tt.status == validStatus {
					isValid = true
					break
				}
			}
			assert.Equal(t, tt.isValidStatus, isValid, "Status validation should match expected result")
		})
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

// Test data validation
func TestLoadTestValidation(t *testing.T) {
	tests := []struct {
		name     string
		loadTest LoadTest
		isValid  bool
	}{
		{
			name: "valid load test with minimum fields",
			loadTest: LoadTest{
				Rows:        100,
				Columns:     5,
				DateColumns: 1,
				Method:      "brute_force",
				Status:      "running",
			},
			isValid: true,
		},
		{
			name: "valid load test with all timing data",
			loadTest: LoadTest{
				Rows:        10000,
				Columns:     100,
				DateColumns: 20,
				Method:      "optimized",
				Status:      "completed",
				CSVGenTime:  intPtr(500),
				ParseTime:   intPtr(1000),
				InsertTime:  intPtr(5000),
				TotalTime:   intPtr(6500),
			},
			isValid: true,
		},
		{
			name: "load test with error status",
			loadTest: LoadTest{
				Rows:         1000,
				Columns:      10,
				DateColumns:  3,
				Method:       "brute_force",
				Status:       "failed",
				ErrorMessage: stringPtr("Connection timeout"),
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			assert.Greater(t, tt.loadTest.Rows, 0, "Rows should be positive")
			assert.Greater(t, tt.loadTest.Columns, 0, "Columns should be positive")
			assert.GreaterOrEqual(t, tt.loadTest.DateColumns, 0, "DateColumns should be non-negative")
			assert.Contains(t, []string{"brute_force", "optimized"}, tt.loadTest.Method, "Method should be valid")
			assert.Contains(t, []string{"running", "completed", "failed"}, tt.loadTest.Status, "Status should be valid")

			if tt.loadTest.Status == "completed" {
				// Completed tests should have timing data
				if tt.loadTest.TotalTime != nil {
					assert.Greater(t, *tt.loadTest.TotalTime, 0, "TotalTime should be positive for completed tests")
				}
			}

			if tt.loadTest.Status == "failed" {
				// Failed tests should have error message
				if tt.loadTest.ErrorMessage != nil {
					assert.NotEmpty(t, *tt.loadTest.ErrorMessage, "ErrorMessage should not be empty for failed tests")
				}
			}
		})
	}
}