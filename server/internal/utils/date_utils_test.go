package utils

import (
	"testing"
	"time"
)

func TestDateValidator_ValidateAndConvert(t *testing.T) {
	validator := NewDateValidator()
	
	testCases := []struct {
		input          string
		shouldBeValid  bool
		expectedFormat DateFormat
	}{
		{"2023-01-15", true, FormatISO8601Date},
		{"01/15/2023", true, FormatUSDate},
		{"15/01/2023", true, FormatEuropeanDate},
		{"15-01-2023", true, FormatDashDate},
		{"1673827200", true, FormatUnixTime}, // Unix timestamp for 2023-01-15
		{"2023-01-15T10:30:00Z", true, FormatISO8601},
		{"January 15, 2023", true, FormatMonthDay},
		{"invalid-date", false, ""},
		{"13/32/2023", false, ""},
		{"", false, ""},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := validator.ValidateAndConvert(tc.input)
			
			if result.IsValid != tc.shouldBeValid {
				t.Errorf("Expected IsValid=%v for input '%s', got %v", 
					tc.shouldBeValid, tc.input, result.IsValid)
			}
			
			if tc.shouldBeValid && result.DetectedFormat != tc.expectedFormat {
				t.Errorf("Expected format %s for input '%s', got %s", 
					tc.expectedFormat, tc.input, result.DetectedFormat)
			}
			
			if tc.shouldBeValid && result.StandardFormat == "" {
				t.Errorf("Expected non-empty standard format for valid input '%s'", tc.input)
			}
		})
	}
}

func TestDateFaker_GenerateFakeDates(t *testing.T) {
	faker := NewDateFaker()
	faker.SetSeed(42) // Set seed for reproducible tests
	
	options := FakeDataOptions{
		Count:       10,
		StartYear:   2020,
		EndYear:     2023,
		IncludeTime: true,
		FormatMix:   true,
	}
	
	dates := faker.GenerateFakeDates(options)
	
	if len(dates) != options.Count {
		t.Errorf("Expected %d dates, got %d", options.Count, len(dates))
	}
	
	// Validate that all generated dates are valid
	validator := NewDateValidator()
	for i, date := range dates {
		result := validator.ValidateAndConvert(date)
		if !result.IsValid {
			t.Errorf("Generated date at index %d is invalid: %s", i, date)
		}
	}
}

func TestDateUtils_ConvertToAllFormats(t *testing.T) {
	utils := NewDateUtils()
	
	testInput := "2023-01-15T10:30:00Z"
	result := utils.ConvertToAllFormats(testInput)
	
	if !result.Success {
		t.Fatalf("Conversion failed: %s", result.ErrorMessage)
	}
	
	if !result.IsValid {
		t.Fatal("Input should be valid")
	}
	
	// Check that we have conversions for multiple formats
	if len(result.ConvertedValues) == 0 {
		t.Fatal("No converted values returned")
	}
	
	// Verify some specific conversions
	if isoDate, exists := result.ConvertedValues[FormatISO8601Date]; !exists || isoDate == "" {
		t.Error("ISO8601 date conversion missing or empty")
	}
	
	if unixTime, exists := result.ConvertedValues[FormatUnixTime]; !exists || unixTime == "" {
		t.Error("Unix timestamp conversion missing or empty")
	}
}

func TestDateUtils_NormalizeDate(t *testing.T) {
	utils := NewDateUtils()
	
	testCases := []struct {
		input       string
		shouldError bool
	}{
		{"01/15/2023", false},
		{"15/01/2023", false},
		{"2023-01-15", false},
		{"January 15, 2023", false},
		{"invalid-date", true},
		{"13/32/2023", true},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result, err := utils.NormalizeDate(tc.input)
			
			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error for input '%s', but got none", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input '%s': %v", tc.input, err)
				}
				if result == "" {
					t.Errorf("Expected non-empty result for valid input '%s'", tc.input)
				}
			}
		})
	}
}

func TestDateUtils_GenerateTestData(t *testing.T) {
	utils := NewDateUtils()
	
	testData := utils.GenerateTestData(5)
	
	if len(testData.ValidDates) != 5 {
		t.Errorf("Expected 5 valid dates, got %d", len(testData.ValidDates))
	}
	
	if len(testData.InvalidDates) == 0 {
		t.Error("Expected some invalid dates for testing")
	}
	
	if len(testData.EdgeCases) == 0 {
		t.Error("Expected some edge cases for testing")
	}
	
	if len(testData.FormatExamples) == 0 {
		t.Error("Expected format examples for testing")
	}
}

func TestDateUtils_CompareTimestamps(t *testing.T) {
	utils := NewDateUtils()
	
	date1 := "2023-01-15T10:00:00Z"
	date2 := "2023-01-15T11:00:00Z"
	
	duration, err := utils.CompareTimestamps(date1, date2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expectedDuration := time.Hour
	if duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, duration)
	}
	
	// Test with invalid date
	_, err = utils.CompareTimestamps("invalid-date", date2)
	if err == nil {
		t.Error("Expected error for invalid date, got none")
	}
}

func TestDateFaker_GenerateEdgeCases(t *testing.T) {
	faker := NewDateFaker()
	edgeCases := faker.GenerateEdgeCases()
	
	if len(edgeCases) == 0 {
		t.Error("Expected edge cases to be generated")
	}
	
	// Validate edge cases are parseable
	validator := NewDateValidator()
	validCount := 0
	for _, edgeCase := range edgeCases {
		if result := validator.ValidateAndConvert(edgeCase); result.IsValid {
			validCount++
		}
	}
	
	// Most edge cases should be valid (they're designed to be valid but edge cases)
	if float64(validCount)/float64(len(edgeCases)) < 0.8 {
		t.Errorf("Expected most edge cases to be valid, got %d/%d valid", validCount, len(edgeCases))
	}
}

func TestValidationStats_GetSuccessRate(t *testing.T) {
	stats := ValidationStats{
		TotalProcessed: 100,
		ValidCount:     85,
		InvalidCount:   15,
	}
	
	expectedRate := 85.0
	actualRate := stats.GetSuccessRate()
	
	if actualRate != expectedRate {
		t.Errorf("Expected success rate %f, got %f", expectedRate, actualRate)
	}
	
	// Test zero division
	emptyStats := ValidationStats{}
	if rate := emptyStats.GetSuccessRate(); rate != 0.0 {
		t.Errorf("Expected 0.0 for empty stats, got %f", rate)
	}
}