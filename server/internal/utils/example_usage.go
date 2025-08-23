package utils

import (
	"fmt"
	"log"
)

// ExampleUsage demonstrates how to use the date validation and faker utilities
func ExampleUsage() {
	fmt.Println("=== Date Validation and Conversion Example ===")
	
	// Create a new DateUtils instance
	dateUtils := NewDateUtils()
	
	// Example 1: Validate and convert various date formats
	testDates := []string{
		"2023-01-15",
		"01/15/2023",
		"15/01/2023",
		"January 15, 2023",
		"2023-01-15T10:30:00Z",
		"1673827200", // Unix timestamp
		"invalid-date",
	}
	
	fmt.Println("\n1. Validating various date formats:")
	for _, date := range testDates {
		result := dateUtils.GetValidator().ValidateAndConvert(date)
		if result.IsValid {
			fmt.Printf("✓ '%s' -> %s (format: %s)\n", 
				date, result.StandardFormat, result.DetectedFormat)
		} else {
			fmt.Printf("✗ '%s' -> INVALID\n", date)
		}
	}
	
	// Example 2: Convert a single date to all supported formats
	fmt.Println("\n2. Converting '2023-01-15T10:30:00Z' to all formats:")
	conversionResult := dateUtils.ConvertToAllFormats("2023-01-15T10:30:00Z")
	if conversionResult.Success {
		for format, converted := range conversionResult.ConvertedValues {
			fmt.Printf("  %s: %s\n", format, converted)
		}
	}
	
	// Example 3: Generate test data
	fmt.Println("\n3. Generating test data:")
	faker := dateUtils.GetFaker()
	faker.SetSeed(42) // For reproducible results
	
	// Generate mixed formats
	mixedDates := faker.GenerateMixedFormats(5)
	fmt.Println("  Mixed format dates:")
	for i, date := range mixedDates {
		result := dateUtils.DetectDateFormat(date)
		fmt.Printf("    %d. %s (format: %s)\n", i+1, date, result)
	}
	
	// Generate specific format
	isodates := faker.GenerateSpecificFormat(FormatISO8601Date, 3)
	fmt.Println("  ISO dates:")
	for i, date := range isodates {
		fmt.Printf("    %d. %s\n", i+1, date)
	}
	
	// Example 4: Normalize dates to standard format
	fmt.Println("\n4. Normalizing dates to standard format:")
	datesToNormalize := []string{
		"01/15/2023",
		"15/01/2023",
		"January 15, 2023",
		"2023-01-15",
	}
	
	for _, date := range datesToNormalize {
		normalized, err := dateUtils.NormalizeDate(date)
		if err != nil {
			fmt.Printf("✗ '%s' -> ERROR: %v\n", date, err)
		} else {
			fmt.Printf("✓ '%s' -> %s\n", date, normalized)
		}
	}
	
	// Example 5: Generate comprehensive test dataset
	fmt.Println("\n5. Generating comprehensive test dataset:")
	testData := dateUtils.GenerateTestData(3)
	
	fmt.Printf("  Valid dates (%d): %v\n", len(testData.ValidDates), testData.ValidDates)
	fmt.Printf("  Invalid dates (%d): %v\n", len(testData.InvalidDates[:3]), testData.InvalidDates[:3])
	fmt.Printf("  Edge cases (%d): %v\n", len(testData.EdgeCases[:3]), testData.EdgeCases[:3])
	
	// Example 6: Validate test data and get statistics
	fmt.Println("\n6. Validation statistics:")
	stats := dateUtils.ValidateTestData(testData)
	fmt.Printf("  Total processed: %d\n", stats.TotalProcessed)
	fmt.Printf("  Success rate: %.1f%%\n", stats.GetSuccessRate())
	fmt.Printf("  Format distribution:\n")
	for format, count := range stats.FormatStats {
		fmt.Printf("    %s: %d\n", format, count)
	}
	if len(stats.Errors) > 0 {
		fmt.Printf("  Errors: %v\n", stats.Errors[:min(3, len(stats.Errors))])
	}
}

// ExamplePerformanceTesting demonstrates usage for performance testing scenarios
func ExamplePerformanceTesting() {
	fmt.Println("\n=== Performance Testing Example ===")
	
	dateUtils := NewDateUtils()
	faker := dateUtils.GetFaker()
	
	// Generate large dataset for performance testing
	options := FakeDataOptions{
		Count:       1000,
		StartYear:   2000,
		EndYear:     2024,
		IncludeTime: true,
		FormatMix:   true,
	}
	
	fmt.Printf("Generating %d test dates...\n", options.Count)
	dates := faker.GenerateFakeDates(options)
	
	// Validate all dates and measure performance
	fmt.Println("Validating all dates...")
	validCount := 0
	for _, date := range dates {
		if dateUtils.IsValidDate(date) {
			validCount++
		}
	}
	
	fmt.Printf("Validation complete: %d/%d dates valid (%.1f%%)\n", 
		validCount, len(dates), float64(validCount)/float64(len(dates))*100)
	
	// Convert all to standard format
	fmt.Println("Converting all dates to standard format...")
	normalized := dateUtils.NormalizeDates(dates)
	normalizedCount := 0
	for _, norm := range normalized {
		if norm != "" {
			normalizedCount++
		}
	}
	
	fmt.Printf("Normalization complete: %d/%d dates normalized\n", 
		normalizedCount, len(dates))
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RunExamples runs all example functions - useful for manual testing
func RunExamples() {
	fmt.Println("Running Date Utils Examples...")
	
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Example execution failed: %v", r)
		}
	}()
	
	ExampleUsage()
	ExamplePerformanceTesting()
	
	fmt.Println("\nExamples completed successfully!")
}