package utils

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// CSVProgressCallback is a function type for CSV generation progress updates
type CSVProgressCallback func(phase string, progress float64, message string)

// CSVGenerationConfig holds configuration for CSV generation
type CSVGenerationConfig struct {
	LoadTestID       uuid.UUID
	Rows             int
	DateColumns      int
	TempDir          string
	FilePrefix       string
	Context          context.Context
	ProgressCallback CSVProgressCallback // Optional progress callback
}

// CSVGenerationResult holds the result of CSV generation
type CSVGenerationResult struct {
	FilePath       string
	GenerationTime int // milliseconds
}

// High-performance data sets - optimized for maximum speed
var performanceDataSets = struct {
	FirstNames        []string
	LastNames         []string
	EmailFirsts       []string
	EmailDomains      []string
	StreetNumbers     []int
	StreetNames       []string
	Cities            []string
	States            []string
	Countries         []string
	Companies         []string
	JobTitles         []string
	Departments       []string
	InsuranceCarriers []string
}{
	FirstNames: []string{
		"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Mary", "James", "Jennifer",
	},
	LastNames: []string{
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez",
	},
	EmailFirsts: []string{
		"john", "jane", "mike", "sarah", "david", "lisa", "bob", "mary", "james", "jen",
	},
	EmailDomains: []string{
		"gmail.com", "yahoo.com", "outlook.com", "company.com", "example.org",
	},
	StreetNumbers: []int{123, 456, 789, 1011, 1234, 5678},
	StreetNames: []string{
		"Main St", "Oak Ave", "First St", "Park Blvd", "Elm St", "Cedar Ave",
	},
	Cities: []string{
		"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia",
	},
	States:    []string{"CA", "TX", "NY", "FL", "IL", "PA"},
	Countries: []string{
		"United States", "Canada", "Mexico", "United Kingdom", "Germany", "France",
	},
	Companies: []string{
		"Acme Corp", "Tech Solutions", "Global Industries", "Metro Systems", "Alpha Enterprises", "Beta LLC",
	},
	JobTitles: []string{
		"Engineer", "Manager", "Analyst", "Coordinator", "Specialist", "Director",
	},
	Departments:       []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations"},
	InsuranceCarriers: []string{"Blue Cross", "Aetna", "Cigna", "United", "Humana", "Kaiser"},
}

// Known date columns for all methods
var allDateColumns = []string{
	"birth_date",
	"start_date",
	"end_date",
	"created_at",
	"updated_at",
}

// Meaningful columns for all methods
var meaningfulColumns = []string{
	"first_name",
	"last_name",
	"email",
	"phone",
	"address_line_1",
	"address_line_2",
	"city",
	"state",
	"zip_code",
	"country",
	"social_security_no",
	"employer",
	"job_title",
	"department",
	"salary",
	"insurance_plan_id",
	"insurance_carrier",
	"policy_number",
	"group_number",
	"member_id",
}

// GeneratePerformanceCSVWithDuplication creates optimized CSV using base generation + doubling strategy
func GeneratePerformanceCSVWithDuplication(config CSVGenerationConfig) (CSVGenerationResult, error) {
	startTime := time.Now()

	// Send initial progress
	if config.ProgressCallback != nil {
		config.ProgressCallback("csv_generation", 0, "Starting optimized CSV generation with duplication...")
	}

	// Calculate optimal base size to minimize truncation waste
	baseRows := calculateOptimalBaseSize(config.Rows)
	
	// Create temp directory
	tempDir := config.TempDir
	if tempDir == "" {
		tempDir = "/tmp/load_tests"
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return CSVGenerationResult{}, fmt.Errorf("failed to create temp directory: %w", err)
	}

	baseCsvPath := filepath.Join(tempDir, "base_"+config.LoadTestID.String()+".csv")
	
	// Generate file path
	prefix := config.FilePrefix
	if prefix == "" {
		prefix = "performance_test"
	}
	finalCsvPath := filepath.Join(tempDir, prefix+"_"+config.LoadTestID.String()+".csv")

	// Step 1: Generate optimized base dataset
	if config.ProgressCallback != nil {
		config.ProgressCallback("csv_generation", 10, fmt.Sprintf("Generating base dataset (%d rows)...", baseRows))
	}
	_, err := generateOptimizedBaseCSVWithProgress(config.Context, baseCsvPath, baseRows, config.ProgressCallback)
	if err != nil {
		return CSVGenerationResult{}, fmt.Errorf("failed to generate base CSV: %w", err)
	}

	// Step 2: Scale up using file doubling strategy
	if config.ProgressCallback != nil {
		config.ProgressCallback("csv_generation", 50, fmt.Sprintf("Scaling to %d rows using duplication...", config.Rows))
	}
	_, err = scaleCSVFileByDoublingWithProgress(config.Context, baseCsvPath, finalCsvPath, baseRows, config.Rows, config.ProgressCallback)
	if err != nil {
		return CSVGenerationResult{}, fmt.Errorf("failed to scale CSV file: %w", err)
	}

	// Cleanup base file
	if err := os.Remove(baseCsvPath); err != nil {
		// Non-fatal error, just log it
		fmt.Printf("Warning: Failed to cleanup base CSV file: %v\n", err)
	}

	// Send completion progress
	if config.ProgressCallback != nil {
		config.ProgressCallback("csv_generation", 100, fmt.Sprintf("CSV generation complete: %d rows", config.Rows))
	}

	totalTime := int(time.Since(startTime).Milliseconds())

	return CSVGenerationResult{
		FilePath:       finalCsvPath,
		GenerationTime: totalTime,
	}, nil
}

// GeneratePerformanceCSV creates a high-performance CSV file using the fastest method
func GeneratePerformanceCSV(config CSVGenerationConfig) (CSVGenerationResult, error) {
	// For small files, use direct generation; for large files, use duplication optimization
	if config.Rows <= 100000 {
		return generateDirectCSV(config)
	}
	return GeneratePerformanceCSVWithDuplication(config)
}

// generateDirectCSV creates CSV directly without duplication (for smaller datasets)
func generateDirectCSV(config CSVGenerationConfig) (CSVGenerationResult, error) {
	startTime := time.Now()

	// Send initial progress
	if config.ProgressCallback != nil {
		config.ProgressCallback("csv_generation", 0, "Starting CSV generation...")
	}

	// Randomly select date columns to populate
	selectedDateColumns := selectRandomDateColumns(allDateColumns, config.DateColumns)

	// Create temp directory
	tempDir := config.TempDir
	if tempDir == "" {
		tempDir = "/tmp/load_tests"
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return CSVGenerationResult{}, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Generate file path
	prefix := config.FilePrefix
	if prefix == "" {
		prefix = "performance_test"
	}
	csvPath := filepath.Join(tempDir, prefix+"_"+config.LoadTestID.String()+".csv")

	file, err := os.Create(csvPath)
	if err != nil {
		return CSVGenerationResult{}, fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Generate headers - combine all columns and randomize
	var allColumns []string
	allColumns = append(allColumns, allDateColumns...)
	allColumns = append(allColumns, meaningfulColumns...)

	// Shuffle columns for variety
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(allColumns), func(i, j int) {
		allColumns[i], allColumns[j] = allColumns[j], allColumns[i]
	})

	if err := writer.Write(allColumns); err != nil {
		return CSVGenerationResult{}, fmt.Errorf("failed to write headers: %w", err)
	}

	// Pre-compute maps for performance
	selectedDateColumnMap := make(map[string]bool)
	for _, col := range selectedDateColumns {
		selectedDateColumnMap[col] = true
	}

	allDateColumnMap := make(map[string]bool)
	for _, col := range allDateColumns {
		allDateColumnMap[col] = true
	}

	// Progress tracking variables
	progressInterval := config.Rows / 20 // Update progress 20 times
	if progressInterval < 1000 {
		progressInterval = 1000 // At least every 1000 rows
	}

	// Generate data rows with high-performance optimizations
	for i := 0; i < config.Rows; i++ {
		// Check for cancellation every 10,000 rows for better performance
		if config.Context != nil && i > 0 && i%10000 == 0 {
			select {
			case <-config.Context.Done():
				return CSVGenerationResult{}, fmt.Errorf("CSV generation cancelled: %w", config.Context.Err())
			default:
			}
		}

		// Send progress updates
		if config.ProgressCallback != nil && i > 0 && i%progressInterval == 0 {
			progress := float64(i) / float64(config.Rows) * 100
			message := fmt.Sprintf("Generated %d/%d rows", i, config.Rows)
			config.ProgressCallback("csv_generation", progress, message)
		}

		row := generatePerformanceDataRow(allColumns, selectedDateColumnMap, allDateColumnMap, rng)
		if err := writer.Write(row); err != nil {
			return CSVGenerationResult{}, fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	// Send completion progress
	if config.ProgressCallback != nil {
		config.ProgressCallback("csv_generation", 100, fmt.Sprintf("CSV generation complete: %d rows", config.Rows))
	}

	generationTime := int(time.Since(startTime).Milliseconds())

	return CSVGenerationResult{
		FilePath:       csvPath,
		GenerationTime: generationTime,
	}, nil
}

// selectRandomDateColumns randomly selects date columns for population
func selectRandomDateColumns(dateColumns []string, count int) []string {
	if count <= 0 {
		return []string{}
	}
	if count >= len(dateColumns) {
		return dateColumns
	}

	columns := make([]string, len(dateColumns))
	copy(columns, dateColumns)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(columns), func(i, j int) {
		columns[i], columns[j] = columns[j], columns[i]
	})

	return columns[:count]
}

// generatePerformanceDataRow creates a data row optimized for maximum performance
func generatePerformanceDataRow(
	headers []string,
	selectedDateColumnMap, allDateColumnMap map[string]bool,
	rng *rand.Rand,
) []string {
	row := make([]string, len(headers))

	for i, header := range headers {
		if allDateColumnMap[header] {
			if selectedDateColumnMap[header] {
				row[i] = generatePerformanceDateValue(rng)
			} else {
				row[i] = ""
			}
		} else {
			row[i] = generatePerformanceMeaningfulColumnValue(header, rng)
		}
	}

	return row
}

// calculateOptimalBaseSize determines the optimal base size for duplication strategy
func calculateOptimalBaseSize(targetRows int) int {
	// Use powers of 2 for efficient doubling
	baseSize := 1000 // Start with 1K rows
	
	// Find the largest power of 2 base that when doubled repeatedly can reach target
	for baseSize*16 < targetRows { // Allow for at least 4 doublings
		baseSize *= 2
	}
	
	// Cap the base size to prevent excessive memory usage
	if baseSize > 50000 {
		baseSize = 50000
	}
	
	return baseSize
}

// generateOptimizedBaseCSV creates the initial base CSV file with optimized content
func generateOptimizedBaseCSV(ctx context.Context, csvPath string, rows int) (int, error) {
	startTime := time.Now()
	
	file, err := os.Create(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create base CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers - combine all date and meaningful columns
	var headers []string
	headers = append(headers, allDateColumns...)
	headers = append(headers, meaningfulColumns...)

	// Shuffle headers for variety
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(headers), func(i, j int) {
		headers[i], headers[j] = headers[j], headers[i]
	})

	if err := writer.Write(headers); err != nil {
		return 0, fmt.Errorf("failed to write headers: %w", err)
	}

	// Pre-compute maps for performance (populate all date columns for base)
	allDateColumnMap := make(map[string]bool)
	for _, col := range allDateColumns {
		allDateColumnMap[col] = true
	}

	// Generate rows with high variety for effective duplication
	for i := 0; i < rows; i++ {
		// Check for cancellation periodically
		if ctx != nil && i > 0 && i%1000 == 0 {
			select {
			case <-ctx.Done():
				return 0, fmt.Errorf("base CSV generation cancelled: %w", ctx.Err())
			default:
			}
		}

		row := generatePerformanceDataRow(headers, allDateColumnMap, allDateColumnMap, rng)
		if err := writer.Write(row); err != nil {
			return 0, fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return int(time.Since(startTime).Milliseconds()), nil
}

// generateOptimizedBaseCSVWithProgress creates the initial base CSV file with progress tracking
func generateOptimizedBaseCSVWithProgress(ctx context.Context, csvPath string, rows int, progressCallback CSVProgressCallback) (int, error) {
	startTime := time.Now()
	
	file, err := os.Create(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create base CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers - combine all date and meaningful columns
	var headers []string
	headers = append(headers, allDateColumns...)
	headers = append(headers, meaningfulColumns...)

	// Shuffle headers for variety
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(headers), func(i, j int) {
		headers[i], headers[j] = headers[j], headers[i]
	})

	if err := writer.Write(headers); err != nil {
		return 0, fmt.Errorf("failed to write headers: %w", err)
	}

	// Pre-compute maps for performance (populate all date columns for base)
	allDateColumnMap := make(map[string]bool)
	for _, col := range allDateColumns {
		allDateColumnMap[col] = true
	}

	// Progress tracking for base generation
	progressInterval := rows / 10 // Update progress 10 times during base generation
	if progressInterval < 500 {
		progressInterval = 500 // At least every 500 rows
	}

	// Generate rows with high variety for effective duplication
	for i := 0; i < rows; i++ {
		// Check for cancellation periodically
		if ctx != nil && i > 0 && i%1000 == 0 {
			select {
			case <-ctx.Done():
				return 0, fmt.Errorf("base CSV generation cancelled: %w", ctx.Err())
			default:
			}
		}

		// Send progress updates for base generation (10-40% of total)
		if progressCallback != nil && i > 0 && i%progressInterval == 0 {
			baseProgress := float64(i) / float64(rows) * 30 // Base generation is 30% of total
			totalProgress := 10 + baseProgress // Start at 10%, go to 40%
			message := fmt.Sprintf("Base generation: %d/%d rows", i, rows)
			progressCallback("csv_generation", totalProgress, message)
		}

		row := generatePerformanceDataRow(headers, allDateColumnMap, allDateColumnMap, rng)
		if err := writer.Write(row); err != nil {
			return 0, fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return int(time.Since(startTime).Milliseconds()), nil
}

// scaleCSVFileByDoublingWithProgress scales up CSV content using file doubling strategy with progress updates
func scaleCSVFileByDoublingWithProgress(ctx context.Context, basePath, finalPath string, baseRows, targetRows int, progressCallback CSVProgressCallback) (int, error) {
	startTime := time.Now()
	
	if targetRows <= baseRows {
		// No scaling needed, just copy the file
		if progressCallback != nil {
			progressCallback("csv_generation", 90, "No scaling needed, copying file...")
		}
		return copyFile(basePath, finalPath)
	}
	
	// Read the base file content (excluding headers)
	if progressCallback != nil {
		progressCallback("csv_generation", 45, "Reading base CSV content...")
	}
	baseContent, headers, err := readCSVContent(basePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read base CSV: %w", err)
	}
	
	// Calculate how many times we need to duplicate
	currentRows := baseRows
	multiplier := 1
	
	// Build the final content by doubling
	finalContent := make([][]string, 0, targetRows)
	finalContent = append(finalContent, baseContent...)
	
	// Calculate total doubling iterations needed for progress tracking
	totalIterations := 0
	tempRows := baseRows
	for tempRows < targetRows {
		tempRows *= 2
		totalIterations++
	}

	iteration := 0
	
	// Double the content until we have enough rows
	for currentRows < targetRows {
		// Check for cancellation
		if ctx != nil {
			select {
			case <-ctx.Done():
				return 0, fmt.Errorf("CSV scaling cancelled: %w", ctx.Err())
			default:
			}
		}
		
		iteration++
		
		// Send progress updates for scaling (50-90% of total)
		if progressCallback != nil {
			scaleProgress := float64(iteration) / float64(totalIterations) * 40 // Scaling is 40% of total
			totalProgress := 50 + scaleProgress // Start at 50%, go to 90%
			message := fmt.Sprintf("Scaling iteration %d/%d (rows: %d)", iteration, totalIterations, currentRows*2)
			progressCallback("csv_generation", totalProgress, message)
		}
		
		// Double the existing content
		contentToAdd := make([][]string, len(finalContent))
		copy(contentToAdd, finalContent)
		
		// Add some variation to avoid exact duplicates
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(multiplier)))
		for i := range contentToAdd {
			// Slightly modify some fields to add variety (e.g., increment numbers)
			for j, cell := range contentToAdd[i] {
				if len(cell) > 0 && isNumericField(headers[j]) {
					contentToAdd[i][j] = addVariationToNumericField(cell, rng)
				}
			}
		}
		
		finalContent = append(finalContent, contentToAdd...)
		currentRows = len(finalContent)
		multiplier++
		
		// Safety check to prevent infinite loop
		if multiplier > 20 {
			break
		}
	}
	
	// Truncate to exact target size
	if len(finalContent) > targetRows {
		finalContent = finalContent[:targetRows]
	}
	
	// Write the final file
	if progressCallback != nil {
		progressCallback("csv_generation", 95, "Writing final CSV file...")
	}
	err = writeCSVContent(finalPath, headers, finalContent)
	if err != nil {
		return 0, fmt.Errorf("failed to write scaled CSV: %w", err)
	}
	
	return int(time.Since(startTime).Milliseconds()), nil
}

// scaleCSVFileByDoubling scales up CSV content using file doubling strategy
func scaleCSVFileByDoubling(ctx context.Context, basePath, finalPath string, baseRows, targetRows int) (int, error) {
	startTime := time.Now()
	
	if targetRows <= baseRows {
		// No scaling needed, just copy the file
		return copyFile(basePath, finalPath)
	}
	
	// Read the base file content (excluding headers)
	baseContent, headers, err := readCSVContent(basePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read base CSV: %w", err)
	}
	
	// Calculate how many times we need to duplicate
	currentRows := baseRows
	multiplier := 1
	
	// Build the final content by doubling
	finalContent := make([][]string, 0, targetRows)
	finalContent = append(finalContent, baseContent...)
	
	// Double the content until we have enough rows
	for currentRows < targetRows {
		// Check for cancellation
		if ctx != nil {
			select {
			case <-ctx.Done():
				return 0, fmt.Errorf("CSV scaling cancelled: %w", ctx.Err())
			default:
			}
		}
		
		// Double the existing content
		contentToAdd := make([][]string, len(finalContent))
		copy(contentToAdd, finalContent)
		
		// Add some variation to avoid exact duplicates
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(multiplier)))
		for i := range contentToAdd {
			// Slightly modify some fields to add variety (e.g., increment numbers)
			for j, cell := range contentToAdd[i] {
				if len(cell) > 0 && isNumericField(headers[j]) {
					contentToAdd[i][j] = addVariationToNumericField(cell, rng)
				}
			}
		}
		
		finalContent = append(finalContent, contentToAdd...)
		currentRows = len(finalContent)
		multiplier++
		
		// Safety check to prevent infinite loop
		if multiplier > 20 {
			break
		}
	}
	
	// Truncate to exact target size
	if len(finalContent) > targetRows {
		finalContent = finalContent[:targetRows]
	}
	
	// Write the final file
	err = writeCSVContent(finalPath, headers, finalContent)
	if err != nil {
		return 0, fmt.Errorf("failed to write scaled CSV: %w", err)
	}
	
	return int(time.Since(startTime).Milliseconds()), nil
}

// readCSVContent reads CSV file and returns content and headers separately
func readCSVContent(csvPath string) ([][]string, []string, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read headers
	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read headers: %w", err)
	}
	
	// Read all content
	content, err := reader.ReadAll()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read CSV content: %w", err)
	}
	
	return content, headers, nil
}

// writeCSVContent writes headers and content to CSV file
func writeCSVContent(csvPath string, headers []string, content [][]string) error {
	file, err := os.Create(csvPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write headers
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// Write content
	for i, row := range content {
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	return nil
}

// copyFile copies a file from source to destination
func copyFile(src, dst string) (int, error) {
	startTime := time.Now()
	
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return 0, err
	}

	return int(time.Since(startTime).Milliseconds()), nil
}

// isNumericField checks if a field typically contains numeric data
func isNumericField(fieldName string) bool {
	numericFields := map[string]bool{
		"salary":            true,
		"zip_code":          true,
		"policy_number":     true,
		"member_id":         true,
		"insurance_plan_id": true,
		"group_number":      true,
	}
	return numericFields[fieldName]
}

// addVariationToNumericField adds small random variation to numeric fields
func addVariationToNumericField(value string, rng *rand.Rand) string {
	// Simple variation - just append a random digit occasionally
	if rng.Intn(10) < 3 { // 30% chance of variation
		return value + fmt.Sprint(rng.Intn(10))
	}
	return value
}

// generatePerformanceDateValue creates a date value optimized for maximum performance
func generatePerformanceDateValue(rng *rand.Rand) string {
	// Fewer formats for better performance
	formats := []string{
		"2006-01-02", "01/02/2006", "01-02-2006", "2006/01/02",
	}

	year := 2020 + rng.Intn(5)
	month := rng.Intn(12) + 1
	day := rng.Intn(28) + 1

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	format := formats[rng.Intn(len(formats))]

	return date.Format(format)
}

// generatePerformanceMeaningfulColumnValue creates fake data optimized for maximum performance
func generatePerformanceMeaningfulColumnValue(columnName string, rng *rand.Rand) string {
	switch columnName {
	case "first_name":
		return performanceDataSets.FirstNames[rng.Intn(len(performanceDataSets.FirstNames))]
	case "last_name":
		return performanceDataSets.LastNames[rng.Intn(len(performanceDataSets.LastNames))]
	case "email":
		return performanceDataSets.EmailFirsts[rng.Intn(len(performanceDataSets.EmailFirsts))] +
			fmt.Sprint(rng.Intn(99)) + "@" +
			performanceDataSets.EmailDomains[rng.Intn(len(performanceDataSets.EmailDomains))]
	case "phone":
		return fmt.Sprintf("%03d%03d%04d", rng.Intn(800)+200, rng.Intn(800)+200, rng.Intn(10000))
	case "address_line_1":
		return fmt.Sprint(
			performanceDataSets.StreetNumbers[rng.Intn(len(performanceDataSets.StreetNumbers))],
		) + " " + performanceDataSets.StreetNames[rng.Intn(len(performanceDataSets.StreetNames))]
	case "address_line_2":
		if rng.Intn(4) == 0 { // 25% chance for better performance
			return "Apt " + fmt.Sprint(rng.Intn(99)+1)
		}
		return ""
	case "city":
		return performanceDataSets.Cities[rng.Intn(len(performanceDataSets.Cities))]
	case "state":
		return performanceDataSets.States[rng.Intn(len(performanceDataSets.States))]
	case "zip_code":
		return fmt.Sprintf("%05d", rng.Intn(99999))
	case "country":
		return performanceDataSets.Countries[rng.Intn(len(performanceDataSets.Countries))]
	case "social_security_no":
		return fmt.Sprintf("***%04d", rng.Intn(10000))
	case "employer":
		return performanceDataSets.Companies[rng.Intn(len(performanceDataSets.Companies))]
	case "job_title":
		return performanceDataSets.JobTitles[rng.Intn(len(performanceDataSets.JobTitles))]
	case "department":
		return performanceDataSets.Departments[rng.Intn(len(performanceDataSets.Departments))]
	case "salary":
		return fmt.Sprint((rng.Intn(100) + 40) * 1000)
	case "insurance_plan_id":
		return fmt.Sprintf("P%03d", rng.Intn(999)+1)
	case "insurance_carrier":
		return performanceDataSets.InsuranceCarriers[rng.Intn(len(performanceDataSets.InsuranceCarriers))]
	case "policy_number":
		return fmt.Sprintf("POL%d", rng.Intn(999999)+100000)
	case "group_number":
		return fmt.Sprintf("G%03d", rng.Intn(999)+1)
	case "member_id":
		return fmt.Sprintf("M%d", rng.Intn(999999)+10000)
	default:
		return fmt.Sprint(rng.Intn(9999))
	}
}