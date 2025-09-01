package utils

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"server/internal/logger"
)

// CSVGeneratorConfig holds configuration for CSV generation
type CSVGeneratorConfig struct {
	TargetRows    int
	TestID        string
	Headers       []string
	DateColumns   []string
	TempDir       string
	WSManager     WebSocketManager
	Logger        logger.Logger
}

// WebSocketManager interface for sending progress updates
type WebSocketManager interface {
	SendLoadTestProgress(testID string, progress map[string]any)
}

// CSVGenerator handles optimized CSV file generation using doubling strategy
type CSVGenerator struct {
	config *CSVGeneratorConfig
	log    logger.Logger
}

// NewCSVGenerator creates a new CSV generator instance
func NewCSVGenerator(config *CSVGeneratorConfig) *CSVGenerator {
	return &CSVGenerator{
		config: config,
		log:    config.Logger.Function("CSVGenerator"),
	}
}

// GenerateOptimizedCSVFile creates a CSV file using base dataset + doubling strategy
func (g *CSVGenerator) GenerateOptimizedCSVFile(ctx context.Context) (string, int, error) {
	log := g.log.Function("GenerateOptimizedCSVFile")
	startTime := time.Now()
	
	// Smart base size selection to avoid truncation
	baseRows := g.calculateOptimalBaseSize(g.config.TargetRows)
	
	log.Info("Starting optimized CSV generation with smart base sizing",
		"targetRows", g.config.TargetRows,
		"optimalBaseRows", baseRows,
		"willRequireTruncation", g.willRequireTruncation(baseRows, g.config.TargetRows))

	// Create temp directory
	tempDir := g.config.TempDir
	if tempDir == "" {
		tempDir = "/tmp/load_tests"
	}
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}

	baseCsvPath := filepath.Join(tempDir, "base_"+g.config.TestID+".csv")
	finalCsvPath := filepath.Join(tempDir, "test_"+g.config.TestID+".csv")

	// Step 1: Generate optimized base dataset
	log.Info("Generating base dataset", "rows", baseRows)
	
	baseGenTime, err := g.generateOptimizedBaseCSVWithProgress(ctx, baseCsvPath, baseRows)
	if err != nil {
		return "", 0, fmt.Errorf("failed to generate base CSV: %w", err)
	}

	// Step 2: Scale up using file doubling  
	log.Info("Scaling dataset using doubling strategy", "from", baseRows, "to", g.config.TargetRows)
	
	// Send scaling progress
	if g.config.WSManager != nil {
		g.config.WSManager.SendLoadTestProgress(g.config.TestID, map[string]any{
			"phase":           "csv_generation",
			"overallProgress": 15,
			"phaseProgress":   50,
			"currentPhase":    "Scaling CSV File",
			"message":         fmt.Sprintf("Scaling from %d to %d rows...", baseRows, g.config.TargetRows),
		})
	}
	
	scaleTime, err := g.scaleCSVFileByDoublingWithProgress(ctx, baseCsvPath, finalCsvPath, baseRows, g.config.TargetRows)
	if err != nil {
		return "", 0, fmt.Errorf("failed to scale CSV file: %w", err)
	}

	// Cleanup base file
	if err := os.Remove(baseCsvPath); err != nil {
		log.Warn("Failed to cleanup base CSV file", "error", err)
	}

	totalTime := int(time.Since(startTime).Milliseconds())
	log.Info("Optimized CSV generation completed",
		"finalPath", finalCsvPath,
		"rows", g.config.TargetRows,
		"baseGenTime", baseGenTime,
		"scaleTime", scaleTime,
		"totalTime", totalTime)

	return finalCsvPath, totalTime, nil
}

// calculateOptimalBaseSize determines the best base size to minimize/eliminate truncation
func (g *CSVGenerator) calculateOptimalBaseSize(targetRows int) int {
	// Define common target patterns and their optimal bases
	commonTargets := map[int]int{
		100000:   100000,  // 100K: exact match
		500000:   62500,   // 500K: 62.5K × 8 = 500K
		1000000:  125000,  // 1M: 125K × 8 = 1M  
		5000000:  78125,   // 5M: 78.125K × 64 = 5M
		10000000: 156250,  // 10M: 156.25K × 64 = 10M
		50000000: 781250,  // 50M: 781.25K × 64 = 50M
		100000000: 781250, // 100M: 781.25K × 128 = 100M
	}
	
	// Check for exact matches first
	if baseSize, exists := commonTargets[targetRows]; exists {
		return baseSize
	}
	
	// For other values, find the largest power of 2 that divides evenly or gets close
	// Start with 100k as our baseline and find the best divisor
	baseOptions := []int{62500, 78125, 100000, 125000, 156250, 781250}
	
	bestBase := 100000
	bestRatio := float64(targetRows) / float64(bestBase)
	
	for _, base := range baseOptions {
		ratio := float64(targetRows) / float64(base)
		powerOf2 := math.Pow(2, math.Round(math.Log2(ratio)))
		
		// Check how close this gets us to the target
		projectedRows := int(float64(base) * powerOf2)
		currentDistance := int(math.Abs(float64(targetRows - projectedRows)))
		
		bestProjected := int(float64(bestBase) * math.Pow(2, math.Round(math.Log2(bestRatio))))
		bestDistance := int(math.Abs(float64(targetRows - bestProjected)))
		
		if currentDistance < bestDistance {
			bestBase = base
			bestRatio = ratio
		}
	}
	
	return bestBase
}

// willRequireTruncation checks if the doubling strategy will need truncation
func (g *CSVGenerator) willRequireTruncation(baseRows, targetRows int) bool {
	if targetRows <= baseRows {
		return false
	}
	
	ratio := float64(targetRows) / float64(baseRows)
	powerOf2 := math.Pow(2, math.Round(math.Log2(ratio)))
	projectedRows := int(float64(baseRows) * powerOf2)
	
	return projectedRows != targetRows
}

// generateOptimizedBaseCSVWithProgress creates base CSV with WebSocket progress updates
func (g *CSVGenerator) generateOptimizedBaseCSVWithProgress(ctx context.Context, csvPath string, rows int) (int, error) {
	log := g.log.Function("generateOptimizedBaseCSVWithProgress")
	startTime := time.Now()

	// Use provided headers or default ones
	headers := g.config.Headers
	if len(headers) == 0 {
		headers = []string{
			"birth_date", "start_date", "end_date", "created_at", "updated_at",
			"first_name", "last_name", "email", "phone", "address_line_1",
			"address_line_2", "city", "state", "zip_code", "country",
			"social_security_no", "employer", "job_title", "department", "salary",
			"insurance_plan_id", "insurance_carrier", "policy_number", "group_number", "member_id",
		}
	}

	// Pre-generate common values for maximum speed
	preGeneratedRows := g.preGenerateOptimizedRows(rows, headers)
	if ctx.Err() != nil {
		return 0, fmt.Errorf("base CSV generation cancelled: %w", ctx.Err())
	}

	file, err := os.Create(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create base CSV file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn("failed to close base CSV file", "error", err)
		}
	}()

	// Use buffered writer for better performance
	const BufferSize = 64 * 1024 // 64KB buffer
	writer := bufio.NewWriterSize(file, BufferSize)
	defer writer.Flush()

	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write headers
	if err := csvWriter.Write(headers); err != nil {
		return 0, fmt.Errorf("failed to write headers: %w", err)
	}

	// Write pre-generated rows
	for i, rowData := range preGeneratedRows {
		// Check for cancellation periodically
		if i%10000 == 0 {
			if ctx.Err() != nil {
				return 0, fmt.Errorf("base CSV generation cancelled: %w", ctx.Err())
			}
		}

		// Parse the row data and write as CSV
		rowFields := strings.Split(rowData, ",")
		if err := csvWriter.Write(rowFields); err != nil {
			return 0, fmt.Errorf("failed to write row %d: %w", i, err)
		}
	}

	totalTime := int(time.Since(startTime).Milliseconds())
	log.Info("Base CSV generation completed", "rows", rows, "time", totalTime)
	
	return totalTime, nil
}

// preGenerateOptimizedRows creates optimized CSV row strings in memory
func (g *CSVGenerator) preGenerateOptimizedRows(rows int, headers []string) []string {
	log := g.log.Function("preGenerateOptimizedRows")
	
	// Pre-allocate slice for better performance
	result := make([]string, 0, rows)
	
	// Use single RNG instance for all data generation
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	
	// Pre-generate value pools for ultra-fast lookups
	datePool := g.generateDatePool(1000, rng)
	namePool := g.generateNamePool(1000, rng)
	emailPool := g.generateEmailPool(1000, rng)
	addressPool := g.generateAddressPool(1000, rng)
	companyPool := g.generateCompanyPool(500, rng)

	// Create column type mappings
	dateColumnMap := make(map[string]bool)
	for _, col := range g.config.DateColumns {
		dateColumnMap[col] = true
	}

	for i := 0; i < rows; i++ {
		var fields []string
		
		for _, header := range headers {
			var value string
			
			if dateColumnMap[header] {
				value = datePool[rng.Intn(len(datePool))]
			} else {
				switch header {
				case "first_name", "last_name":
					value = namePool[rng.Intn(len(namePool))]
				case "email":
					value = emailPool[rng.Intn(len(emailPool))]
				case "phone":
					value = fmt.Sprintf("555-%03d-%04d", rng.Intn(1000), rng.Intn(10000))
				case "address_line_1", "address_line_2", "city", "state", "zip_code", "country":
					value = addressPool[rng.Intn(len(addressPool))]
				case "social_security_no":
					value = fmt.Sprintf("%03d-%02d-%04d", rng.Intn(1000), rng.Intn(100), rng.Intn(10000))
				case "employer", "job_title", "department":
					value = companyPool[rng.Intn(len(companyPool))]
				case "salary":
					value = strconv.Itoa(30000 + rng.Intn(170000))
				case "insurance_plan_id", "member_id":
					value = uuid.New().String()
				case "insurance_carrier":
					carriers := []string{"Blue Cross", "Aetna", "Cigna", "United Health", "Humana"}
					value = carriers[rng.Intn(len(carriers))]
				case "policy_number", "group_number":
					value = fmt.Sprintf("%08d", rng.Intn(100000000))
				default:
					value = fmt.Sprintf("data_%d", rng.Intn(10000))
				}
			}
			
			fields = append(fields, value)
		}
		
		result = append(result, strings.Join(fields, ","))
	}

	log.Info("Pre-generated rows completed", "rows", rows)
	return result
}

// scaleCSVFileByDoublingWithProgress scales CSV with WebSocket progress updates
func (g *CSVGenerator) scaleCSVFileByDoublingWithProgress(ctx context.Context, basePath, finalPath string, baseRows, targetRows int) (int, error) {
	log := g.log.Function("scaleCSVFileByDoublingWithProgress")
	startTime := time.Now()
	
	if targetRows <= baseRows {
		// Just copy the base file if target is smaller/equal
		if g.config.WSManager != nil {
			g.config.WSManager.SendLoadTestProgress(g.config.TestID, map[string]any{
				"phase":           "csv_scaling",
				"overallProgress": 20,
				"phaseProgress":   100,
				"currentPhase":    "File Copy (No Scaling Needed)",
				"message":         "Target size equal to base size, copying file...",
			})
		}
		return g.copyFile(basePath, finalPath)
	}

	log.Info("Starting file doubling strategy with progress tracking",
		"baseRows", baseRows, "targetRows", targetRows)

	currentPath := basePath
	currentRows := baseRows
	iteration := 0

	// Calculate total doubling iterations needed
	ratio := float64(targetRows) / float64(baseRows)
	totalIterations := int(math.Log2(ratio))

	for currentRows < targetRows {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("CSV scaling cancelled: %w", ctx.Err())
		default:
		}

		iteration++
		nextRows := currentRows * 2
		
		// Don't exceed target
		if nextRows > targetRows {
			nextRows = targetRows
		}

		tempPath := filepath.Join(filepath.Dir(currentPath), fmt.Sprintf("temp_%d_%s.csv", iteration, g.config.TestID))

		log.Info("Doubling file", "iteration", iteration, "from", currentRows, "to", nextRows)

		// Send progress update for significant iterations only
		if iteration%2 == 0 || nextRows == targetRows {
			progressPercent := int(float64(iteration) / float64(totalIterations) * 100)
			if progressPercent > 100 {
				progressPercent = 100
			}
			
			if g.config.WSManager != nil {
				g.config.WSManager.SendLoadTestProgress(g.config.TestID, map[string]any{
					"phase":           "csv_scaling",
					"overallProgress": 20,
					"phaseProgress":   progressPercent,
					"currentPhase":    "Scaling CSV File",
					"message":         fmt.Sprintf("Iteration %d: %d → %d rows", iteration, currentRows, nextRows),
				})
			}
		}

		if err := g.doubleFile(ctx, currentPath, tempPath, nextRows); err != nil {
			return 0, fmt.Errorf("failed to double file at iteration %d: %w", iteration, err)
		}

		// Clean up previous iteration (except base file)
		if currentPath != basePath {
			if err := os.Remove(currentPath); err != nil {
				log.Warn("Failed to cleanup temp file", "path", currentPath, "error", err)
			}
		}

		currentPath = tempPath
		currentRows = nextRows

		if currentRows >= targetRows {
			break
		}
	}

	// Move final file to target location
	if err := os.Rename(currentPath, finalPath); err != nil {
		return 0, fmt.Errorf("failed to move final file: %w", err)
	}

	totalTime := int(time.Since(startTime).Milliseconds())
	log.Info("File doubling completed",
		"finalRows", currentRows,
		"iterations", iteration,
		"totalTime", totalTime)

	return totalTime, nil
}

// Helper methods for file operations
func (g *CSVGenerator) copyFile(src, dst string) (int, error) {
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

func (g *CSVGenerator) doubleFile(ctx context.Context, srcPath, dstPath string, targetRows int) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	const BufferSize = 64 * 1024 // 64KB buffer
	srcReader := bufio.NewReaderSize(srcFile, BufferSize)
	dstWriter := bufio.NewWriterSize(dstFile, BufferSize)
	defer dstWriter.Flush()

	// Copy header first
	header, _, err := srcReader.ReadLine()
	if err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	
	if _, err := dstWriter.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := dstWriter.WriteString("\n"); err != nil {
		return fmt.Errorf("failed to write header newline: %w", err)
	}

	// Read all data rows into memory first
	var dataRows [][]byte
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		
		line, _, err := srcReader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read line: %w", err)
		}
		
		// Make a copy of the line since ReadLine reuses the buffer
		lineCopy := make([]byte, len(line))
		copy(lineCopy, line)
		dataRows = append(dataRows, lineCopy)
	}

	// Write rows repeatedly until we reach target
	rowsWritten := 0
	for rowsWritten < targetRows {
		for _, row := range dataRows {
			if rowsWritten >= targetRows {
				break
			}
			
			if _, err := dstWriter.Write(row); err != nil {
				return fmt.Errorf("failed to write row: %w", err)
			}
			if _, err := dstWriter.WriteString("\n"); err != nil {
				return fmt.Errorf("failed to write newline: %w", err)
			}
			rowsWritten++
		}
	}

	return nil
}

// Value generation helper methods
func (g *CSVGenerator) generateDatePool(size int, rng *rand.Rand) []string {
	pool := make([]string, size)
	baseDate := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	
	for i := 0; i < size; i++ {
		randomDays := rng.Intn(365 * 50) // Random date within ~50 years
		date := baseDate.AddDate(0, 0, randomDays)
		pool[i] = date.Format("2006-01-02")
	}
	
	return pool
}

func (g *CSVGenerator) generateNamePool(size int, rng *rand.Rand) []string {
	firstNames := []string{"John", "Jane", "Bob", "Alice", "Charlie", "Diana", "Frank", "Grace", "Henry", "Ivy"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
	
	pool := make([]string, size)
	for i := 0; i < size; i++ {
		if rng.Float32() < 0.5 {
			pool[i] = firstNames[rng.Intn(len(firstNames))]
		} else {
			pool[i] = lastNames[rng.Intn(len(lastNames))]
		}
	}
	
	return pool
}

func (g *CSVGenerator) generateEmailPool(size int, rng *rand.Rand) []string {
	domains := []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "company.com"}
	pool := make([]string, size)
	
	for i := 0; i < size; i++ {
		username := fmt.Sprintf("user%d", rng.Intn(100000))
		domain := domains[rng.Intn(len(domains))]
		pool[i] = username + "@" + domain
	}
	
	return pool
}

func (g *CSVGenerator) generateAddressPool(size int, rng *rand.Rand) []string {
	addresses := []string{"Main St", "Oak Ave", "Park Dr", "First St", "Second Ave", "Elm St", "Pine Rd", "Cedar Ln"}
	cities := []string{"Springfield", "Franklin", "Georgetown", "Madison", "Washington", "Lincoln", "Jefferson", "Adams"}
	states := []string{"CA", "NY", "TX", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
	countries := []string{"USA", "Canada", "Mexico", "UK", "Germany", "France", "Italy", "Spain", "Japan", "Australia"}
	
	pool := make([]string, size)
	for i := 0; i < size; i++ {
		switch rng.Intn(4) {
		case 0: // Address
			pool[i] = fmt.Sprintf("%d %s", rng.Intn(9999)+1, addresses[rng.Intn(len(addresses))])
		case 1: // City
			pool[i] = cities[rng.Intn(len(cities))]
		case 2: // State
			pool[i] = states[rng.Intn(len(states))]
		case 3: // Country
			pool[i] = countries[rng.Intn(len(countries))]
		}
	}
	
	return pool
}

func (g *CSVGenerator) generateCompanyPool(size int, rng *rand.Rand) []string {
	companies := []string{"Tech Corp", "Data Systems", "Software Inc", "Cloud Solutions", "Digital Works", "Innovation Labs"}
	departments := []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations", "Support", "Research"}
	titles := []string{"Manager", "Director", "Analyst", "Specialist", "Coordinator", "Associate", "Lead", "Senior"}
	
	pool := make([]string, size)
	for i := 0; i < size; i++ {
		switch rng.Intn(3) {
		case 0:
			pool[i] = companies[rng.Intn(len(companies))]
		case 1:
			pool[i] = departments[rng.Intn(len(departments))]
		case 2:
			pool[i] = titles[rng.Intn(len(titles))]
		}
	}
	
	return pool
}