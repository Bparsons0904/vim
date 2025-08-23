package utils

import (
	"fmt"
	"time"
)

// DateUtils provides a comprehensive suite of date validation, conversion, and generation utilities
type DateUtils struct {
	validator *DateValidator
	faker     *DateFaker
}

// ConversionResult represents the result of a date conversion operation
type ConversionResult struct {
	ValidationResult
	ConvertedValues map[DateFormat]string
	Success         bool
	ErrorMessage    string
}

// NewDateUtils creates a new DateUtils instance with validator and faker
func NewDateUtils() *DateUtils {
	return &DateUtils{
		validator: NewDateValidator(),
		faker:     NewDateFaker(),
	}
}

// GetValidator returns the internal validator
func (du *DateUtils) GetValidator() *DateValidator {
	return du.validator
}

// GetFaker returns the internal faker
func (du *DateUtils) GetFaker() *DateFaker {
	return du.faker
}

// ConvertToAllFormats converts a date string to all supported formats
func (du *DateUtils) ConvertToAllFormats(input string) ConversionResult {
	result := ConversionResult{
		ConvertedValues: make(map[DateFormat]string),
		Success:         false,
	}

	// First validate the input
	validationResult := du.validator.ValidateAndConvert(input)
	result.ValidationResult = validationResult

	if !validationResult.IsValid {
		result.ErrorMessage = "Invalid date format: could not parse input"
		return result
	}

	// Convert to all supported formats
	formats := du.validator.GetSupportedFormats()
	for _, format := range formats {
		if format == FormatUnixTime {
			result.ConvertedValues[format] = fmt.Sprintf("%d", validationResult.ParsedTime.Unix())
		} else {
			result.ConvertedValues[format] = validationResult.ParsedTime.Format(string(format))
		}
	}

	result.Success = true
	return result
}

// ConvertToSpecificFormats converts a date string to specified formats only
func (du *DateUtils) ConvertToSpecificFormats(input string, targetFormats []DateFormat) ConversionResult {
	result := ConversionResult{
		ConvertedValues: make(map[DateFormat]string),
		Success:         false,
	}

	// First validate the input
	validationResult := du.validator.ValidateAndConvert(input)
	result.ValidationResult = validationResult

	if !validationResult.IsValid {
		result.ErrorMessage = "Invalid date format: could not parse input"
		return result
	}

	// Convert to specified formats
	for _, format := range targetFormats {
		if format == FormatUnixTime {
			result.ConvertedValues[format] = fmt.Sprintf("%d", validationResult.ParsedTime.Unix())
		} else {
			result.ConvertedValues[format] = validationResult.ParsedTime.Format(string(format))
		}
	}

	result.Success = true
	return result
}

// NormalizeDate converts any valid date to the standard format (ISO8601 by default)
func (du *DateUtils) NormalizeDate(input string) (string, error) {
	validationResult := du.validator.ValidateAndConvert(input)
	if !validationResult.IsValid {
		return "", fmt.Errorf("invalid date format: %s", input)
	}
	return validationResult.StandardFormat, nil
}

// NormalizeDates normalizes a batch of date strings
func (du *DateUtils) NormalizeDates(inputs []string) []string {
	results := make([]string, len(inputs))
	for i, input := range inputs {
		if normalized, err := du.NormalizeDate(input); err == nil {
			results[i] = normalized
		} else {
			results[i] = "" // Invalid dates become empty strings
		}
	}
	return results
}

// GenerateTestData creates a comprehensive test dataset with both valid and invalid dates
func (du *DateUtils) GenerateTestData(count int) TestDataSet {
	return TestDataSet{
		ValidDates:   du.faker.GenerateMixedFormats(count),
		InvalidDates: du.faker.GenerateInvalidDates(),
		EdgeCases:    du.faker.GenerateEdgeCases(),
		FormatExamples: du.faker.GenerateTestDataSet(),
	}
}

// TestDataSet represents a complete set of test data for date validation
type TestDataSet struct {
	ValidDates     []string                   `json:"validDates"`
	InvalidDates   []string                   `json:"invalidDates"`
	EdgeCases      []string                   `json:"edgeCases"`
	FormatExamples map[DateFormat][]string    `json:"formatExamples"`
}

// ValidateTestData validates all dates in a test dataset and returns statistics
func (du *DateUtils) ValidateTestData(testData TestDataSet) ValidationStats {
	stats := ValidationStats{
		TotalProcessed: 0,
		ValidCount:     0,
		InvalidCount:   0,
		FormatStats:    make(map[DateFormat]int),
		Errors:         []string{},
	}

	// Process valid dates
	for _, date := range testData.ValidDates {
		result := du.validator.ValidateAndConvert(date)
		stats.TotalProcessed++
		if result.IsValid {
			stats.ValidCount++
			stats.FormatStats[result.DetectedFormat]++
		} else {
			stats.InvalidCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("Expected valid but failed: %s", date))
		}
	}

	// Process invalid dates (should fail validation)
	for _, date := range testData.InvalidDates {
		result := du.validator.ValidateAndConvert(date)
		stats.TotalProcessed++
		if !result.IsValid {
			stats.ValidCount++ // Correctly identified as invalid
		} else {
			stats.InvalidCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("Expected invalid but passed: %s", date))
		}
	}

	// Process edge cases
	for _, date := range testData.EdgeCases {
		result := du.validator.ValidateAndConvert(date)
		stats.TotalProcessed++
		if result.IsValid {
			stats.ValidCount++
			stats.FormatStats[result.DetectedFormat]++
		} else {
			stats.InvalidCount++
			stats.Errors = append(stats.Errors, fmt.Sprintf("Edge case failed: %s", date))
		}
	}

	return stats
}

// ValidationStats provides statistics about validation results
type ValidationStats struct {
	TotalProcessed int                 `json:"totalProcessed"`
	ValidCount     int                 `json:"validCount"`
	InvalidCount   int                 `json:"invalidCount"`
	FormatStats    map[DateFormat]int  `json:"formatStats"`
	Errors         []string            `json:"errors"`
}

// GetSuccessRate returns the validation success rate as a percentage
func (vs *ValidationStats) GetSuccessRate() float64 {
	if vs.TotalProcessed == 0 {
		return 0.0
	}
	return (float64(vs.ValidCount) / float64(vs.TotalProcessed)) * 100.0
}

// DetectDateFormat attempts to detect the format of a date string without full parsing
func (du *DateUtils) DetectDateFormat(input string) DateFormat {
	result := du.validator.ValidateAndConvert(input)
	if result.IsValid {
		return result.DetectedFormat
	}
	return ""
}

// IsValidDate checks if a string is a valid date in any supported format
func (du *DateUtils) IsValidDate(input string) bool {
	result := du.validator.ValidateAndConvert(input)
	return result.IsValid
}

// ConvertDateRange converts a range of dates maintaining their relationships
func (du *DateUtils) ConvertDateRange(dates []string, targetFormat DateFormat) []string {
	results := make([]string, len(dates))
	for i, date := range dates {
		if conversionResult := du.ConvertToSpecificFormats(date, []DateFormat{targetFormat}); conversionResult.Success {
			results[i] = conversionResult.ConvertedValues[targetFormat]
		} else {
			results[i] = "" // Invalid dates become empty
		}
	}
	return results
}

// GetFormatInfo returns information about all supported formats
func (du *DateUtils) GetFormatInfo() map[DateFormat]string {
	formatInfo := make(map[DateFormat]string)
	formats := du.validator.GetSupportedFormats()
	
	for _, format := range formats {
		formatInfo[format] = GetFormatExample(format)
	}
	
	return formatInfo
}

// CompareTimestamps compares two date strings and returns their time difference
func (du *DateUtils) CompareTimestamps(date1, date2 string) (time.Duration, error) {
	result1 := du.validator.ValidateAndConvert(date1)
	if !result1.IsValid {
		return 0, fmt.Errorf("invalid first date: %s", date1)
	}
	
	result2 := du.validator.ValidateAndConvert(date2)
	if !result2.IsValid {
		return 0, fmt.Errorf("invalid second date: %s", date2)
	}
	
	return result2.ParsedTime.Sub(result1.ParsedTime), nil
}