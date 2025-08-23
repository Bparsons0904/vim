package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"server/internal/logger"
	"server/internal/repositories"
	. "server/internal/models"
	"server/internal/utils"
	"time"

	"github.com/google/uuid"
)

type LoadTestController struct {
	loadTestRepo repositories.LoadTestRepository
	testDataRepo repositories.TestDataRepository
	dateUtils    *utils.DateUtils
	log          logger.Logger
	wsManager    WSManager
}

// WSManager interface for WebSocket operations to avoid import cycles
type WSManager interface {
	SendLoadTestProgress(testID string, data map[string]any)
	SendLoadTestComplete(testID string, testResult map[string]any)
	SendLoadTestError(testID string, errorMsg string)
}

func NewLoadTestController(
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
	wsManager WSManager,
) *LoadTestController {
	return &LoadTestController{
		loadTestRepo: loadTestRepo,
		testDataRepo: testDataRepo,
		dateUtils:    utils.NewDateUtils(),
		log:          logger.New("loadTestController"),
		wsManager:    wsManager,
	}
}

// GetKnownDateColumns returns the list of known date column names that need validation
func GetKnownDateColumns() []string {
	return []string{
		"birth_date",
		"start_date", 
		"end_date",
		"plan_start_date",
		"plan_end_date",
		"created_at",
		"updated_at", 
		"effective_date",
		"expiry_date",
		"last_login_date",
	}
}


// CreateAndRunTest creates a new load test and starts the performance test
func (c *LoadTestController) CreateAndRunTest(ctx context.Context, req *CreateLoadTestRequest) (*LoadTest, error) {
	log := c.log.Function("CreateAndRunTest")
	
	// Create the LoadTest record  
	loadTest := &LoadTest{
		BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
		Rows:          req.Rows,
		Columns:       req.Columns,
		DateColumns:   req.DateColumns,
		Method:        req.Method,
		Status:        "running",
	}

	if err := c.loadTestRepo.Create(ctx, loadTest); err != nil {
		_ = log.Err("failed to create load test", err, "loadTest", loadTest)
		return nil, fmt.Errorf("failed to create load test: %w", err)
	}

	// Process the load test asynchronously
	go c.processLoadTest(ctx, loadTest)

	log.Info("load test created and started", "loadTestId", loadTest.ID, "method", loadTest.Method)
	return loadTest, nil
}

// GetLoadTestByID retrieves a load test by ID
func (c *LoadTestController) GetLoadTestByID(ctx context.Context, id string) (*LoadTest, error) {
	return c.loadTestRepo.GetByID(ctx, id)
}

// GetAllLoadTests retrieves all load tests
func (c *LoadTestController) GetAllLoadTests(ctx context.Context) ([]*LoadTest, error) {
	return c.loadTestRepo.GetAll(ctx)
}

// processLoadTest handles the actual load test processing
func (c *LoadTestController) processLoadTest(ctx context.Context, loadTest *LoadTest) {
	log := c.log.Function("processLoadTest")
	testID := loadTest.ID.String()

	// Send initial progress
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":            "csv_generation",
		"overallProgress":  0,
		"phaseProgress":    0,
		"currentPhase":     "Generating CSV Data",
		"rowsProcessed":    0,
		"rowsPerSecond":    0,
		"eta":              "Calculating...",
		"message":          "Starting CSV generation...",
	})

	// Step 1: Generate CSV file
	csvPath, csvGenTime, err := c.generateCSVFileWithProgress(loadTest)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "CSV generation failed", err)
		c.wsManager.SendLoadTestError(testID, "CSV generation failed: "+err.Error())
		return
	}

	// Progress update: CSV generation complete
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":            "parsing",
		"overallProgress":  25,
		"phaseProgress":    0,
		"currentPhase":     "Parsing and Validating",
		"rowsProcessed":    0,
		"rowsPerSecond":    0,
		"eta":              "Calculating...",
		"message":          "Starting data parsing and validation...",
	})

	// Step 2: Parse and validate CSV data
	testData, parseTime, err := c.parseAndValidateCSVWithProgress(csvPath, loadTest)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "CSV parsing failed", err)
		c.wsManager.SendLoadTestError(testID, "CSV parsing failed: "+err.Error())
		return
	}

	// Progress update: Parsing complete
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":            "insertion",
		"overallProgress":  85,
		"phaseProgress":    0,
		"currentPhase":     "Inserting into Database",
		"rowsProcessed":    0,
		"rowsPerSecond":    0,
		"eta":              "Calculating...",
		"message":          fmt.Sprintf("Starting database insertion using %s method...", loadTest.Method),
	})

	// Step 3: Insert data using specified method
	insertTime, err := c.insertTestDataWithProgress(ctx, testData, loadTest.Method, testID)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "Data insertion failed", err)
		c.wsManager.SendLoadTestError(testID, "Data insertion failed: "+err.Error())
		return
	}

	// Step 4: Update load test with completion data
	totalTime := csvGenTime + parseTime + insertTime
	loadTest.CSVGenTime = &csvGenTime
	loadTest.ParseTime = &parseTime
	loadTest.InsertTime = &insertTime
	loadTest.TotalTime = &totalTime
	loadTest.Status = "completed"

	if err := c.loadTestRepo.Update(ctx, loadTest); err != nil {
		_ = log.Err("failed to update completed load test", err, "loadTestId", loadTest.ID)
	}

	// Send completion notification
	c.wsManager.SendLoadTestComplete(testID, map[string]any{
		"id":           loadTest.ID.String(),
		"rows":         loadTest.Rows,
		"columns":      loadTest.Columns,
		"dateColumns":  loadTest.DateColumns,
		"method":       loadTest.Method,
		"status":       "completed",
		"csvGenTime":   csvGenTime,
		"parseTime":    parseTime,
		"insertTime":   insertTime,
		"totalTime":    totalTime,
	})

	log.Info("load test completed successfully", 
		"loadTestId", loadTest.ID,
		"totalTime", totalTime,
		"method", loadTest.Method)
}

// generateCSVFileWithProgress creates a CSV file with the specified dimensions and known date columns
func (c *LoadTestController) generateCSVFileWithProgress(loadTest *LoadTest) (string, int, error) {
	log := c.log.Function("generateCSVFile")
	startTime := time.Now()
	
	allDateColumns := GetKnownDateColumns()
	
	// Randomly select which date columns to populate for this test
	selectedDateColumns := c.selectRandomDateColumns(allDateColumns, loadTest.DateColumns)
	
	log.Info("CSV generation started", 
		"rows", loadTest.Rows, 
		"columns", loadTest.Columns,
		"totalDateColumns", len(allDateColumns),
		"selectedDateColumns", len(selectedDateColumns),
		"selectedColumns", selectedDateColumns)
	
	// Create temp directory if it doesn't exist
	tempDir := "/tmp/load_tests"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	// Generate file path
	csvPath := filepath.Join(tempDir, "load_test_"+loadTest.ID.String()+".csv")
	
	// Create CSV file
	file, err := os.Create(csvPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn("failed to close CSV file", "error", err)
		}
	}()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// Generate column headers in random order (includes all date columns, but only selected ones will have data)
	headers := c.generateRandomizedHeaders(loadTest.Columns, allDateColumns)
	if err := writer.Write(headers); err != nil {
		return "", 0, fmt.Errorf("failed to write headers: %w", err)
	}
	
	// Generate data rows
	for i := 0; i < loadTest.Rows; i++ {
		row := c.generateDataRow(headers, selectedDateColumns)
		if err := writer.Write(row); err != nil {
			return "", 0, fmt.Errorf("failed to write row %d: %w", i, err)
		}
		
		// Log progress for large datasets
		if i > 0 && i%10000 == 0 {
			log.Debug("CSV generation progress", "rowsGenerated", i, "totalRows", loadTest.Rows)
		}
	}
	
	csvGenTime := int(time.Since(startTime).Milliseconds())
	log.Info("CSV generation completed", 
		"csvPath", csvPath,
		"rows", loadTest.Rows,
		"columns", len(headers),
		"generationTimeMs", csvGenTime)
	
	return csvPath, csvGenTime, nil
}

// generateRandomizedHeaders creates a randomized list of column headers
func (c *LoadTestController) generateRandomizedHeaders(totalColumns int, knownDateColumns []string) []string {
	var allColumns []string
	
	// Add known date columns
	allColumns = append(allColumns, knownDateColumns...)
	
	// Add regular columns (Col1, Col2, etc.)
	regularColumnsNeeded := totalColumns - len(knownDateColumns)
	if regularColumnsNeeded < 0 {
		regularColumnsNeeded = 0
	}
	
	for i := 1; i <= regularColumnsNeeded; i++ {
		allColumns = append(allColumns, fmt.Sprintf("col%d", i))
	}
	
	// Shuffle the columns randomly
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(allColumns), func(i, j int) {
		allColumns[i], allColumns[j] = allColumns[j], allColumns[i]
	})
	
	return allColumns
}

// selectRandomDateColumns randomly selects a subset of date columns to populate
func (c *LoadTestController) selectRandomDateColumns(allDateColumns []string, count int) []string {
	if count <= 0 {
		return []string{}
	}
	if count >= len(allDateColumns) {
		return allDateColumns
	}
	
	// Create a copy to avoid modifying the original slice
	columns := make([]string, len(allDateColumns))
	copy(columns, allDateColumns)
	
	// Shuffle the columns
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(columns), func(i, j int) {
		columns[i], columns[j] = columns[j], columns[i]
	})
	
	// Return the first 'count' columns
	return columns[:count]
}

// generateDataRow creates a single data row with appropriate data for each column
func (c *LoadTestController) generateDataRow(headers []string, selectedDateColumns []string) []string {
	row := make([]string, len(headers))
	
	// Create a map for quick lookup of selected date columns (ones that should have data)
	selectedDateColumnMap := make(map[string]bool)
	for _, col := range selectedDateColumns {
		selectedDateColumnMap[col] = true
	}
	
	// Create a map for all known date columns (to distinguish from regular columns)
	allDateColumns := GetKnownDateColumns()
	allDateColumnMap := make(map[string]bool)
	for _, col := range allDateColumns {
		allDateColumnMap[col] = true
	}
	
	for i, header := range headers {
		if allDateColumnMap[header] {
			// This is a date column - populate only if it's selected
			if selectedDateColumnMap[header] {
				// Generate date data in various formats
				row[i] = c.generateDateValue()
			} else {
				// Leave empty for unselected date columns
				row[i] = ""
			}
		} else {
			// Generate random string data for regular columns
			row[i] = c.generateStringValue()
		}
	}
	
	return row
}

// generateDateValue creates a random date in various formats using our date utils
func (c *LoadTestController) generateDateValue() string {
	// Generate fake dates using our date faker
	faker := c.dateUtils.GetFaker()
	faker.SetSeed(time.Now().UnixNano())
	
	// Generate a single date in mixed formats
	dates := faker.GenerateMixedFormats(1)
	if len(dates) > 0 {
		return dates[0]
	}
	
	// Fallback to a simple date format if faker fails
	return "2023-01-15"
}

// generateStringValue creates random string data for regular columns
func (c *LoadTestController) generateStringValue() string {
	// Generate random strings of varying lengths
	prefixes := []string{"data", "value", "item", "record", "entry", "field", "content", "text"}
	numbers := []string{"001", "002", "123", "456", "789", "999", "abc", "xyz"}
	
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	prefix := prefixes[r.Intn(len(prefixes))]
	number := numbers[r.Intn(len(numbers))]
	
	return fmt.Sprintf("%s_%s_%d", prefix, number, r.Intn(10000))
}

// parseAndValidateCSVWithProgress reads the CSV file and validates only the populated date columns
func (c *LoadTestController) parseAndValidateCSVWithProgress(csvPath string, loadTest *LoadTest) ([]*TestData, int, error) {
	log := c.log.Function("parseAndValidateCSV")
	startTime := time.Now()
	
	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warn("failed to close CSV file", "error", err)
		}
	}()
	
	reader := csv.NewReader(file)
	
	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to read CSV headers: %w", err)
	}
	
	// Create header index map for efficient lookups
	headerIndex := make(map[string]int)
	for i, header := range headers {
		headerIndex[header] = i
	}
	
	log.Info("CSV parsing started", 
		"csvPath", csvPath,
		"totalColumns", len(headers),
		"expectedRows", loadTest.Rows)
	
	var testData []*TestData
	var validationErrors []string
	rowCount := 0
	
	// Process each data row
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, 0, fmt.Errorf("failed to read CSV row %d: %w", rowCount, err)
		}
		
		rowCount++
		
		// Create TestData record
		data := &TestData{
			BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
			LoadTestID:    loadTest.ID,
		}
		
		// Parse and validate the row
		if err := c.parseAndValidateRow(record, headers, headerIndex, data); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("row %d: %v", rowCount, err))
			// Continue processing even with validation errors
		}
		
		testData = append(testData, data)
		
		// Log progress for large datasets
		if rowCount > 0 && rowCount%10000 == 0 {
			log.Debug("CSV parsing progress", "rowsParsed", rowCount, "expectedRows", loadTest.Rows)
		}
	}
	
	parseTime := int(time.Since(startTime).Milliseconds())
	
	log.Info("CSV parsing completed", 
		"csvPath", csvPath,
		"rowsParsed", rowCount,
		"validationErrors", len(validationErrors),
		"parseTimeMs", parseTime)
	
	// Log validation errors if any (but don't fail)
	if len(validationErrors) > 0 {
		log.Warn("validation errors encountered", 
			"errorCount", len(validationErrors),
			"firstFewErrors", validationErrors[:min(5, len(validationErrors))])
	}
	
	return testData, parseTime, nil
}

// parseAndValidateRow parses a single CSV row into a TestData struct with validation
func (c *LoadTestController) parseAndValidateRow(record []string, headers []string, headerIndex map[string]int, data *TestData) error {
	knownDateColumns := GetKnownDateColumns()
	var validationErrors []string
	
	// Map all columns from CSV to TestData struct
	for i, header := range headers {
		if i >= len(record) {
			continue // Skip if record is shorter than headers
		}
		
		value := record[i]
		
		// Handle date columns with validation
		if c.isKnownDateColumn(header, knownDateColumns) {
			// Only validate if the value is not empty
			if value != "" {
				if !c.validateDateValue(value) {
					validationErrors = append(validationErrors, fmt.Sprintf("invalid date in %s: %s", header, value))
				}
			}
			
			// Set the value regardless of validation (store what we got)
			if err := c.setDateColumnValue(data, header, value); err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("failed to set %s: %v", header, err))
			}
		} else {
			// Handle regular columns (no validation needed)
			if err := c.setRegularColumnValue(data, header, value); err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("failed to set %s: %v", header, err))
			}
		}
	}
	
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation errors: %v", validationErrors)
	}
	
	return nil
}

// isKnownDateColumn checks if a column name is a known date column
func (c *LoadTestController) isKnownDateColumn(columnName string, knownDateColumns []string) bool {
	for _, dateCol := range knownDateColumns {
		if columnName == dateCol {
			return true
		}
	}
	return false
}

// validateDateValue validates a date string using our date utils
func (c *LoadTestController) validateDateValue(value string) bool {
	result := c.dateUtils.GetValidator().ValidateAndConvert(value)
	return result.IsValid
}

// setDateColumnValue sets a date column value in the TestData struct
func (c *LoadTestController) setDateColumnValue(data *TestData, columnName, value string) error {
	// Handle empty values
	var valuePtr *string
	if value != "" {
		valuePtr = &value
	}
	
	switch columnName {
	case "birth_date":
		data.BirthDate = valuePtr
	case "start_date":
		data.StartDate = valuePtr
	case "end_date":
		data.EndDate = valuePtr
	case "plan_start_date":
		data.PlanStartDate = valuePtr
	case "plan_end_date":
		data.PlanEndDate = valuePtr
	case "created_at":
		data.CreatedAt = valuePtr
	case "updated_at":
		data.UpdatedAt = valuePtr
	case "effective_date":
		data.EffectiveDate = valuePtr
	case "expiry_date":
		data.ExpiryDate = valuePtr
	case "last_login_date":
		data.LastLoginDate = valuePtr
	default:
		return fmt.Errorf("unknown date column: %s", columnName)
	}
	return nil
}

// setRegularColumnValue sets a regular column value in the TestData struct
func (c *LoadTestController) setRegularColumnValue(data *TestData, columnName, value string) error {
	// Handle empty values
	var valuePtr *string
	if value != "" {
		valuePtr = &value
	}
	
	// Parse column number from colX format
	if len(columnName) > 3 && columnName[:3] == "col" {
		// For regular columns like col1, col2, etc.
		// We'll use reflection or a switch statement based on the column number
		switch columnName {
		case "col1":
			data.Col1 = valuePtr
		case "col2":
			data.Col2 = valuePtr
		case "col3":
			data.Col3 = valuePtr
		case "col4":
			data.Col4 = valuePtr
		case "col5":
			data.Col5 = valuePtr
		// Add more cases as needed... this is where we'd handle col6 through col200
		// For now, we'll just handle the first few and ignore the rest for the demo
		default:
			// For columns beyond col5, we'll just ignore them for now
			// In a real implementation, you'd either extend the TestData model
			// or use a dynamic approach like storing in a map
			return nil
		}
	}
	return nil
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// insertTestDataWithProgress performs the actual data insertion using the specified method
func (c *LoadTestController) insertTestDataWithProgress(ctx context.Context, testData []*TestData, method string, testID string) (int, error) {
	log := c.log.Function("insertTestData")
	startTime := time.Now()
	
	log.Info("data insertion started", "method", method, "recordCount", len(testData))
	
	switch method {
	case "brute_force":
		return c.insertBruteForceWithProgress(ctx, testData, startTime, testID)
	case "optimized":
		return c.insertOptimizedWithProgress(ctx, testData, startTime, testID)
	default:
		return 0, fmt.Errorf("unknown insertion method: %s", method)
	}
}

// insertBruteForceWithProgress performs individual inserts for each record
func (c *LoadTestController) insertBruteForceWithProgress(ctx context.Context, testData []*TestData, startTime time.Time, testID string) (int, error) {
	log := c.log.Function("insertBruteForce")
	
	successCount := 0
	errorCount := 0
	
	for i, data := range testData {
		if err := c.testDataRepo.Create(ctx, data); err != nil {
			log.Warn("failed to insert record", "recordIndex", i, "error", err)
			errorCount++
		} else {
			successCount++
		}
		
		// Send progress updates every 500 records or for smaller datasets every 50
		progressInterval := 500
		if len(testData) < 5000 {
			progressInterval = 50
		}
		
		if i > 0 && i%progressInterval == 0 {
			elapsed := time.Since(startTime)
			rowsPerSecond := int(float64(i) / elapsed.Seconds())
			remaining := len(testData) - i
			eta := "Calculating..."
			if rowsPerSecond > 0 {
				etaSeconds := remaining / rowsPerSecond
				if etaSeconds < 60 {
					eta = fmt.Sprintf("%ds", etaSeconds)
				} else {
					eta = fmt.Sprintf("%dm", etaSeconds/60)
				}
			}
			
			progress := float64(i) / float64(len(testData)) * 100
			c.wsManager.SendLoadTestProgress(testID, map[string]any{
				"phase":            "insertion",
				"overallProgress":  85 + (progress * 0.15), // 85% to 100%
				"phaseProgress":    progress,
				"currentPhase":     "Inserting into Database",
				"rowsProcessed":    i,
				"rowsPerSecond":    rowsPerSecond,
				"eta":              eta,
				"message":          fmt.Sprintf("Inserting records using brute force method (%d/%d)...", i, len(testData)),
			})
			
			log.Debug("brute force insertion progress", 
				"recordsProcessed", i+1,
				"totalRecords", len(testData),
				"successCount", successCount,
				"errorCount", errorCount)
		}
	}
	
	insertTime := int(time.Since(startTime).Milliseconds())
	
	log.Info("brute force insertion completed", 
		"totalRecords", len(testData),
		"successCount", successCount,
		"errorCount", errorCount,
		"insertTimeMs", insertTime)
	
	if errorCount > 0 {
		return insertTime, fmt.Errorf("brute force insertion completed with %d errors out of %d records", errorCount, len(testData))
	}
	
	return insertTime, nil
}

// insertOptimizedWithProgress performs batch inserts for better performance
func (c *LoadTestController) insertOptimizedWithProgress(ctx context.Context, testData []*TestData, startTime time.Time, testID string) (int, error) {
	log := c.log.Function("insertOptimized")
	
	// Use default batch size of 1000, or configure based on data size
	batchSize := 1000
	if len(testData) < 100 {
		batchSize = len(testData) // Use smaller batches for small datasets
	}
	
	// Send progress update for optimized batch insertion start
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":            "insertion",
		"overallProgress":  90,
		"phaseProgress":    50,
		"currentPhase":     "Inserting into Database",
		"rowsProcessed":    0,
		"rowsPerSecond":    0,
		"eta":              "Calculating...",
		"message":          fmt.Sprintf("Performing optimized batch insertion (%d records in batches of %d)...", len(testData), batchSize),
	})
	
	if err := c.testDataRepo.CreateBatch(ctx, testData, batchSize); err != nil {
		insertTime := int(time.Since(startTime).Milliseconds())
		_ = log.Err("optimized batch insertion failed", err, "recordCount", len(testData), "batchSize", batchSize)
		return insertTime, fmt.Errorf("batch insertion failed: %w", err)
	}
	
	insertTime := int(time.Since(startTime).Milliseconds())
	rowsPerSecond := int(float64(len(testData)) / (float64(insertTime) / 1000))
	
	// Send final progress update
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":            "insertion",
		"overallProgress":  100,
		"phaseProgress":    100,
		"currentPhase":     "Database Insertion Complete",
		"rowsProcessed":    len(testData),
		"rowsPerSecond":    rowsPerSecond,
		"eta":              "Done",
		"message":          fmt.Sprintf("Successfully inserted %d records using optimized batch method", len(testData)),
	})
	
	log.Info("optimized batch insertion completed", 
		"totalRecords", len(testData),
		"batchSize", batchSize,
		"insertTimeMs", insertTime)
	
	return insertTime, nil
}

// updateLoadTestError updates the load test with error information
func (c *LoadTestController) updateLoadTestError(ctx context.Context, loadTest *LoadTest, message string, err error) {
	log := c.log.Function("updateLoadTestError")
	
	errorMsg := message + ": " + err.Error()
	loadTest.Status = "failed"
	loadTest.ErrorMessage = &errorMsg
	
	if updateErr := c.loadTestRepo.Update(ctx, loadTest); updateErr != nil {
		_ = log.Err("failed to update load test error", updateErr, "loadTestId", loadTest.ID)
	}
	
	_ = log.Err(message, err, "loadTestId", loadTest.ID)
}

