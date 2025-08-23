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

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type LoadTestController struct {
	loadTestRepo repositories.LoadTestRepository
	testDataRepo repositories.TestDataRepository
	dateUtils    *utils.DateUtils
	log          logger.Logger
}

func NewLoadTestController(
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
) *LoadTestController {
	return &LoadTestController{
		loadTestRepo: loadTestRepo,
		testDataRepo: testDataRepo,
		dateUtils:    utils.NewDateUtils(),
		log:          logger.New("loadTestController"),
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

// CreateLoadTestRequest represents the request payload for creating a load test
type CreateLoadTestRequest struct {
	Rows        int    `json:"rows" validate:"required,min=1"`
	Columns     int    `json:"columns" validate:"required,min=1"`  // Total columns (210 max: 200 regular + 10 date)
	DateColumns int    `json:"dateColumns" validate:"min=0,max=10"` // Number of date columns to populate (0-10)
	Method      string `json:"method" validate:"required,oneof=brute_force optimized"`
}

// CreateLoadTest creates a new load test and starts the performance test
func (c *LoadTestController) CreateLoadTest(ctx *fiber.Ctx) error {
	log := c.log.Function("CreateLoadTest")
	
	var req CreateLoadTestRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Er("invalid request body", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	// TODO: Add request validation

	// Step 1: Create the LoadTest record  
	loadTest := &LoadTest{
		BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
		Rows:          req.Rows,
		Columns:       req.Columns,
		Method:        req.Method,
		Status:        "running",
	}
	
	// Store dateColumns for the async processor
	loadTest.DateColumns = req.DateColumns

	if err := c.loadTestRepo.Create(ctx.Context(), loadTest); err != nil {
		log.Er("failed to create load test", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create load test",
		})
	}

	// Step 2: Generate CSV data (async)
	go c.processLoadTest(ctx.Context(), loadTest)

	return ctx.JSON(fiber.Map{
		"loadTestId": loadTest.ID,
		"status":     "started",
		"message":    "Load test started successfully",
	})
}

// processLoadTest handles the actual load test processing
func (c *LoadTestController) processLoadTest(ctx context.Context, loadTest *LoadTest) {
	log := c.log.Function("processLoadTest")

	// Step 1: Generate CSV file
	csvPath, csvGenTime, err := c.generateCSVFile(loadTest)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "CSV generation failed", err)
		return
	}

	// Step 2: Parse and validate CSV data
	testData, parseTime, err := c.parseAndValidateCSV(csvPath, loadTest)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "CSV parsing failed", err)
		return
	}

	// Step 3: Insert data using specified method
	insertTime, err := c.insertTestData(ctx, testData, loadTest.Method)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "Data insertion failed", err)
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
		log.Err("failed to update completed load test", err, "loadTestId", loadTest.ID)
	}

	log.Info("load test completed successfully", 
		"loadTestId", loadTest.ID,
		"totalTime", totalTime,
		"method", loadTest.Method)
}

// generateCSVFile creates a CSV file with the specified dimensions and known date columns
func (c *LoadTestController) generateCSVFile(loadTest *LoadTest) (string, int, error) {
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
	defer file.Close()
	
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

// parseAndValidateCSV reads the CSV file and validates only the populated date columns
func (c *LoadTestController) parseAndValidateCSV(csvPath string, loadTest *LoadTest) ([]*TestData, int, error) {
	log := c.log.Function("parseAndValidateCSV")
	startTime := time.Now()
	
	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()
	
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

// insertTestData performs the actual data insertion using the specified method
func (c *LoadTestController) insertTestData(ctx context.Context, testData []*TestData, method string) (int, error) {
	log := c.log.Function("insertTestData")
	startTime := time.Now()
	
	log.Info("data insertion started", "method", method, "recordCount", len(testData))
	
	switch method {
	case "brute_force":
		return c.insertBruteForce(ctx, testData, startTime)
	case "optimized":
		return c.insertOptimized(ctx, testData, startTime)
	default:
		return 0, fmt.Errorf("unknown insertion method: %s", method)
	}
}

// insertBruteForce performs individual inserts for each record
func (c *LoadTestController) insertBruteForce(ctx context.Context, testData []*TestData, startTime time.Time) (int, error) {
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
		
		// Log progress for large datasets
		if i > 0 && i%1000 == 0 {
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

// insertOptimized performs batch inserts for better performance
func (c *LoadTestController) insertOptimized(ctx context.Context, testData []*TestData, startTime time.Time) (int, error) {
	log := c.log.Function("insertOptimized")
	
	// Use default batch size of 1000, or configure based on data size
	batchSize := 1000
	if len(testData) < 100 {
		batchSize = len(testData) // Use smaller batches for small datasets
	}
	
	if err := c.testDataRepo.CreateBatch(ctx, testData, batchSize); err != nil {
		insertTime := int(time.Since(startTime).Milliseconds())
		log.Err("optimized batch insertion failed", err, "recordCount", len(testData), "batchSize", batchSize)
		return insertTime, fmt.Errorf("batch insertion failed: %w", err)
	}
	
	insertTime := int(time.Since(startTime).Milliseconds())
	
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
		log.Err("failed to update load test error", updateErr, "loadTestId", loadTest.ID)
	}
	
	log.Err(message, err, "loadTestId", loadTest.ID)
}

// GetLoadTest retrieves a load test by ID
func (c *LoadTestController) GetLoadTest(ctx *fiber.Ctx) error {
	log := c.log.Function("GetLoadTest")
	
	id := ctx.Params("id")
	if id == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "load test ID is required",
		})
	}

	loadTest, err := c.loadTestRepo.GetByID(ctx.Context(), id)
	if err != nil {
		log.Er("load test not found", err)
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "load test not found",
		})
	}

	return ctx.JSON(loadTest)
}

// GetLoadTests retrieves all load tests
func (c *LoadTestController) GetLoadTests(ctx *fiber.Ctx) error {
	log := c.log.Function("GetLoadTests")
	
	loadTests, err := c.loadTestRepo.GetAll(ctx.Context())
	if err != nil {
		log.Er("failed to retrieve load tests", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve load tests",
		})
	}

	return ctx.JSON(loadTests)
}

// GetTestData retrieves test data for a specific load test with pagination
func (c *LoadTestController) GetTestData(ctx *fiber.Ctx) error {
	log := c.log.Function("GetTestData")
	
	loadTestID := ctx.Params("id")
	if loadTestID == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "load test ID is required",
		})
	}

	// TODO: Parse pagination parameters from query string
	offset := 0
	limit := 100
	
	testData, err := c.testDataRepo.GetByLoadTestIDPaginated(ctx.Context(), loadTestID, offset, limit)
	if err != nil {
		log.Er("failed to retrieve test data", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to retrieve test data",
		})
	}

	// Get total count for pagination
	totalCount, err := c.testDataRepo.CountByLoadTestID(ctx.Context(), loadTestID)
	if err != nil {
		log.Er("failed to count test data", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to count test data",
		})
	}

	return ctx.JSON(fiber.Map{
		"data":       testData,
		"totalCount": totalCount,
		"offset":     offset,
		"limit":      limit,
	})
}