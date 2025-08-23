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
		"created_at",
		"updated_at", 
	}
}

// GetMeaningfulColumns returns the list of meaningful column names for demographics, employment, and insurance
func GetMeaningfulColumns() []string {
	return []string{
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
}


// CreateAndRunTest creates a new load test and starts the performance test
func (c *LoadTestController) CreateAndRunTest(ctx context.Context, req *CreateLoadTestRequest) (*LoadTest, error) {
	log := c.log.Function("CreateAndRunTest")
	
	// Create the LoadTest record with fixed column structure
	// We use a fixed structure: 5 date columns + 20 regular columns = 25 total
	const FixedTotalColumns = 25
	const FixedDateColumns = 5 // We populate all 5 available date columns
	
	loadTest := &LoadTest{
		BaseUUIDModel: BaseUUIDModel{ID: uuid.New()},
		Rows:          req.Rows,
		Columns:       FixedTotalColumns,     // Override: always use 25 columns
		DateColumns:   FixedDateColumns,      // Override: always populate 6 date columns
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
	
	// Add meaningful columns (demographics, employment, insurance)
	meaningfulColumns := GetMeaningfulColumns()
	allColumns = append(allColumns, meaningfulColumns...)
	
	// Verify we have exactly the expected number of columns
	if len(allColumns) != totalColumns {
		c.log.Warn("column count mismatch", 
			"expected", totalColumns, 
			"actual", len(allColumns),
			"dateColumns", len(knownDateColumns),
			"meaningfulColumns", len(meaningfulColumns))
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
			// Generate appropriate data for meaningful columns
			row[i] = c.generateMeaningfulColumnValue(header)
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

// generateMeaningfulColumnValue creates appropriate fake data for meaningful columns
func (c *LoadTestController) generateMeaningfulColumnValue(columnName string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(1000))))
	
	switch columnName {
	case "first_name":
		names := []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Mary", "James", "Jennifer"}
		return names[r.Intn(len(names))]
	case "last_name":
		surnames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez"}
		return surnames[r.Intn(len(surnames))]
	case "email":
		first := []string{"john", "jane", "mike", "sarah", "david", "lisa", "bob", "mary", "james", "jen"}
		domains := []string{"gmail.com", "yahoo.com", "outlook.com", "company.com", "example.org"}
		return fmt.Sprintf("%s%d@%s", first[r.Intn(len(first))], r.Intn(999), domains[r.Intn(len(domains))])
	case "phone":
		return fmt.Sprintf("(%03d) %03d-%04d", r.Intn(800)+200, r.Intn(800)+200, r.Intn(10000))
	case "address_line_1":
		numbers := []int{123, 456, 789, 1011, 1234, 5678}
		streets := []string{"Main St", "Oak Ave", "First St", "Park Blvd", "Elm St", "Cedar Ave"}
		return fmt.Sprintf("%d %s", numbers[r.Intn(len(numbers))], streets[r.Intn(len(streets))])
	case "address_line_2":
		if r.Intn(3) == 0 { // 33% chance of having apartment/unit
			return fmt.Sprintf("Apt %d", r.Intn(999)+1)
		}
		return "" // Often empty
	case "city":
		cities := []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "Austin"}
		return cities[r.Intn(len(cities))]
	case "state":
		states := []string{"CA", "TX", "NY", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
		return states[r.Intn(len(states))]
	case "zip_code":
		return fmt.Sprintf("%05d", r.Intn(99999))
	case "country":
		countries := []string{"United States", "Canada", "Mexico", "United Kingdom", "Germany", "France", "Japan", "Australia"}
		return countries[r.Intn(len(countries))]
	case "social_security_no":
		return fmt.Sprintf("***-**-%04d", r.Intn(10000)) // Masked for privacy
	case "employer":
		companies := []string{"Acme Corp", "Tech Solutions", "Global Industries", "Metro Systems", "Alpha Enterprises", "Beta LLC", "Gamma Inc", "Delta Group"}
		return companies[r.Intn(len(companies))]
	case "job_title":
		titles := []string{"Software Engineer", "Manager", "Analyst", "Coordinator", "Specialist", "Director", "Associate", "Consultant", "Administrator"}
		return titles[r.Intn(len(titles))]
	case "department":
		departments := []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations", "IT", "Customer Service", "Legal"}
		return departments[r.Intn(len(departments))]
	case "salary":
		return fmt.Sprintf("$%d", (r.Intn(100)+40)*1000) // $40k to $140k
	case "insurance_plan_id":
		return fmt.Sprintf("PLAN-%03d", r.Intn(999)+1)
	case "insurance_carrier":
		carriers := []string{"Blue Cross Blue Shield", "Aetna", "Cigna", "UnitedHealthcare", "Humana", "Kaiser Permanente", "Anthem"}
		return carriers[r.Intn(len(carriers))]
	case "policy_number":
		return fmt.Sprintf("POL-%d-%d", r.Intn(9999)+1000, r.Intn(999999)+100000)
	case "group_number":
		return fmt.Sprintf("GRP%03d", r.Intn(999)+1)
	case "member_id":
		return fmt.Sprintf("MEM%d%03d", r.Intn(99)+10, r.Intn(999)+1)
	default:
		// Fallback for unknown columns
		return fmt.Sprintf("data_%d", r.Intn(9999))
	}
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
			"errorRate", fmt.Sprintf("%.2f%%", float64(len(validationErrors))/float64(rowCount)*100),
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
		
		// Handle date columns with validation and normalization
		if c.isKnownDateColumn(header, knownDateColumns) {
			// Set the value with validation and normalization
			if err := c.setDateColumnValue(data, header, value); err != nil {
				validationErrors = append(validationErrors, fmt.Sprintf("failed to set %s: %v", header, err))
			}
			
			// Additional validation reporting for non-empty values
			if value != "" {
				if _, isValid := c.validateAndNormalizeDateValue(value); !isValid {
					validationErrors = append(validationErrors, fmt.Sprintf("invalid date in %s: %s", header, value))
				}
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

// validateAndNormalizeDateValue validates and normalizes a date string to ISO8601 format with UTC timezone
func (c *LoadTestController) validateAndNormalizeDateValue(value string) (string, bool) {
	if value == "" {
		return "", true // Empty values are valid
	}
	
	result := c.dateUtils.GetValidator().ValidateAndConvert(value)
	if !result.IsValid {
		return "", false
	}
	
	// Convert to UTC and format as ISO8601 with timezone (RFC3339)
	normalized := result.ParsedTime.UTC().Format(time.RFC3339)
	return normalized, true
}

// validateDateValue validates a date string using our date utils (keeping for backward compatibility)
func (c *LoadTestController) validateDateValue(value string) bool {
	result := c.dateUtils.GetValidator().ValidateAndConvert(value)
	return result.IsValid
}

// setDateColumnValue sets a date column value in the TestData struct with normalization
func (c *LoadTestController) setDateColumnValue(data *TestData, columnName, value string) error {
	// Handle empty values and normalize non-empty ones
	var valuePtr *string
	if value != "" {
		normalized, isValid := c.validateAndNormalizeDateValue(value)
		if !isValid {
			// Log the invalid date but store the original value for debugging
			c.log.Warn("invalid date detected, storing original value", 
				"column", columnName, 
				"originalValue", value)
			valuePtr = &value
		} else {
			valuePtr = &normalized
		}
	}
	
	switch columnName {
	case "birth_date":
		data.BirthDate = valuePtr
	case "start_date":
		data.StartDate = valuePtr
	case "end_date":
		data.EndDate = valuePtr
	case "created_at":
		data.CreatedAt = valuePtr
	case "updated_at":
		data.UpdatedAt = valuePtr
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
	
	// Map meaningful column names to TestData fields
	switch columnName {
	case "first_name":
		data.FirstName = valuePtr
	case "last_name":
		data.LastName = valuePtr
	case "email":
		data.Email = valuePtr
	case "phone":
		data.Phone = valuePtr
	case "address_line_1":
		data.AddressLine1 = valuePtr
	case "address_line_2":
		data.AddressLine2 = valuePtr
	case "city":
		data.City = valuePtr
	case "state":
		data.State = valuePtr
	case "zip_code":
		data.ZipCode = valuePtr
	case "country":
		data.Country = valuePtr
	case "social_security_no":
		data.SocialSecurityNo = valuePtr
	case "employer":
		data.Employer = valuePtr
	case "job_title":
		data.JobTitle = valuePtr
	case "department":
		data.Department = valuePtr
	case "salary":
		data.Salary = valuePtr
	case "insurance_plan_id":
		data.InsurancePlanID = valuePtr
	case "insurance_carrier":
		data.InsuranceCarrier = valuePtr
	case "policy_number":
		data.PolicyNumber = valuePtr
	case "group_number":
		data.GroupNumber = valuePtr
	case "member_id":
		data.MemberID = valuePtr
	default:
		// Unknown column - log warning but continue
		return fmt.Errorf("unknown meaningful column: %s", columnName)
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
	case "batched", "optimized": // Support both names for backward compatibility
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

