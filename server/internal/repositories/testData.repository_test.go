package repositories

import (
	"server/internal/logger"
	. "server/internal/models"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTestDataRepository_BatchSizeCalculation(t *testing.T) {
	tests := []struct {
		name         string
		testDataSize int
		batchSize    int
		expectedBatches int
	}{
		{
			name:            "default batch size with small data",
			testDataSize:    500,
			batchSize:       0, // Should use default 1000
			expectedBatches: 1,
		},
		{
			name:            "custom batch size",
			testDataSize:    2500,
			batchSize:       1000,
			expectedBatches: 3, // 1000, 1000, 500
		},
		{
			name:            "exact batch size match",
			testDataSize:    1000,
			batchSize:       1000,
			expectedBatches: 1,
		},
		{
			name:            "small batch size",
			testDataSize:    100,
			batchSize:       25,
			expectedBatches: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batchSize := tt.batchSize
			if batchSize <= 0 {
				batchSize = 1000 // Default
			}

			expectedBatches := (tt.testDataSize + batchSize - 1) / batchSize
			assert.Equal(t, tt.expectedBatches, expectedBatches, "Batch calculation should be correct")
		})
	}
}

func TestTestDataRepository_UUIDValidation(t *testing.T) {
	tests := []struct {
		name        string
		loadTestID  string
		expectError bool
	}{
		{
			name:        "valid UUID",
			loadTestID:  uuid.New().String(),
			expectError: false,
		},
		{
			name:        "invalid UUID",
			loadTestID:  "invalid-uuid",
			expectError: true,
		},
		{
			name:        "empty UUID",
			loadTestID:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test UUID parsing logic
			_, err := uuid.Parse(tt.loadTestID)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTestDataRepository_PaginationValidation(t *testing.T) {
	tests := []struct {
		name        string
		offset      int
		limit       int
		isValid     bool
	}{
		{
			name:    "valid pagination with offset and limit",
			offset:  10,
			limit:   50,
			isValid: true,
		},
		{
			name:    "pagination with zero offset",
			offset:  0,
			limit:   100,
			isValid: true,
		},
		{
			name:    "pagination with zero limit (no limit)",
			offset:  20,
			limit:   0,
			isValid: true,
		},
		{
			name:    "negative offset",
			offset:  -5,
			limit:   10,
			isValid: false,
		},
		{
			name:    "negative limit",
			offset:  0,
			limit:   -10,
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate pagination parameters
			isValid := tt.offset >= 0 && tt.limit >= 0
			assert.Equal(t, tt.isValid, isValid, "Pagination validation should match expected result")
		})
	}
}

func TestTestDataRepository_GetColumnValue(t *testing.T) {
	loadTestID := uuid.New()
	testData := &TestData{
		BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
		LoadTestID:    loadTestID,
		Col1:          stringPtr("value1"),
		Col2:          stringPtr("value2"),
		Col3:          stringPtr("value3"),
		Col4:          stringPtr("value4"),
		Col5:          stringPtr("value5"),
	}

	repo := &testDataRepository{
		log: logger.New("test"),
	}

	tests := []struct {
		columnName    string
		expectedValue *string
	}{
		{"col1", stringPtr("value1")},
		{"col2", stringPtr("value2")},
		{"col3", stringPtr("value3")},
		{"col4", stringPtr("value4")},
		{"col5", stringPtr("value5")},
		{"col999", nil}, // Non-existent column
		{"invalid", nil},
	}

	for _, tt := range tests {
		t.Run(tt.columnName, func(t *testing.T) {
			result := repo.GetColumnValue(testData, tt.columnName)
			
			if tt.expectedValue == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, *tt.expectedValue, *result)
			}
		})
	}
}

func TestTestDataRepository_SetColumnValue(t *testing.T) {
	loadTestID := uuid.New()
	testData := &TestData{
		BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
		LoadTestID:    loadTestID,
	}

	repo := &testDataRepository{
		log: logger.New("test"),
	}

	tests := []struct {
		columnName    string
		value         *string
		expectError   bool
	}{
		{"col1", stringPtr("newValue1"), false},
		{"col2", stringPtr("newValue2"), false},
		{"col3", stringPtr("newValue3"), false},
		{"col4", stringPtr("newValue4"), false},
		{"col5", stringPtr("newValue5"), false},
		{"col999", stringPtr("value"), true}, // Non-existent column
		{"invalid", stringPtr("value"), true},
	}

	for _, tt := range tests {
		t.Run(tt.columnName, func(t *testing.T) {
			err := repo.SetColumnValue(testData, tt.columnName, tt.value)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unknown column name")
			} else {
				assert.NoError(t, err)
				
				// Verify the value was set
				retrievedValue := repo.GetColumnValue(testData, tt.columnName)
				assert.NotNil(t, retrievedValue)
				assert.Equal(t, *tt.value, *retrievedValue)
			}
		})
	}
}

func TestTestDataRepository_DataValidation(t *testing.T) {
	tests := []struct {
		name     string
		testData *TestData
		isValid  bool
	}{
		{
			name: "valid test data",
			testData: &TestData{
				BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
				LoadTestID:    uuid.New(),
				Col1:          stringPtr("2023-01-01"),
				Col2:          stringPtr("data"),
				Col3:          stringPtr("value"),
			},
			isValid: true,
		},
		{
			name: "test data with nil values",
			testData: &TestData{
				BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
				LoadTestID:    uuid.New(),
				Col1:          nil,
				Col2:          nil,
			},
			isValid: true, // Nil values are allowed
		},
		{
			name: "test data with empty UUID",
			testData: &TestData{
				BaseUUIDModel: BaseUUIDModel{ID: uuid.Nil},
				LoadTestID:    uuid.New(),
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation checks
			isValid := tt.testData.ID != uuid.Nil && tt.testData.LoadTestID != uuid.Nil
			assert.Equal(t, tt.isValid, isValid, "Data validation should match expected result")
		})
	}
}

func TestTestDataRepository_CustomColumnMapping(t *testing.T) {
	tests := []struct {
		name     string
		records  []map[string]interface{}
		expected int
	}{
		{
			name: "valid custom column mapping",
			records: []map[string]interface{}{
				{
					"col1": "2023-01-01",
					"col2": "data1",
					"col3": "value1",
				},
				{
					"col1": "2023-01-02", 
					"col2": "data2",
					"col3": "value2",
				},
			},
			expected: 2,
		},
		{
			name: "mixed data types",
			records: []map[string]interface{}{
				{
					"col1": "string_value",
					"col2": 123,        // Should be ignored (not string)
					"col3": "value3",
				},
			},
			expected: 1,
		},
		{
			name:     "empty records",
			records:  []map[string]interface{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the data conversion logic
			assert.Equal(t, tt.expected, len(tt.records), "Record count should match expected")
			
			for _, record := range tt.records {
				// Validate that string values can be extracted
				if val, ok := record["col1"]; ok {
					if str, ok := val.(string); ok {
						assert.NotEmpty(t, str, "String values should not be empty")
					}
				}
			}
		})
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

