package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"server/internal/utils"
	"time"
)

type LoadTestController struct {
	loadTestRepo        repositories.LoadTestRepository
	testDataRepo        repositories.TestDataRepository
	plaidController     *PlaidController
	optimizedController *OptimizedOnlyController
	ludicrousController *LudicrousOnlyController
	dateUtils           *utils.DateUtils
	log                 logger.Logger
	wsManager           WSManager
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
	db database.DB,
	wsManager WSManager,
	config config.Config,
	plaidController *PlaidController,
) *LoadTestController {
	optimizedController := NewOptimizedOnlyController(
		loadTestRepo,
		testDataRepo,
		db,
		wsManager,
		config,
	)
	ludicrousController := NewLudicrousOnlyController(
		loadTestRepo,
		testDataRepo,
		db,
		wsManager,
		config,
	)

	return &LoadTestController{
		loadTestRepo:        loadTestRepo,
		testDataRepo:        testDataRepo,
		plaidController:     plaidController,
		optimizedController: optimizedController,
		ludicrousController: ludicrousController,
		dateUtils:           utils.NewDateUtils(),
		log:                 logger.New("loadTestController"),
		wsManager:           wsManager,
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
func (c *LoadTestController) CreateAndRunTest(
	ctx context.Context,
	req *CreateLoadTestRequest,
) (*LoadTest, error) {
	log := c.log.Function("CreateAndRunTest")

	// Delegate to dedicated controllers for optimized and ludicrous methods
	if req.Method == "optimized" {
		log.Info("delegating to optimized controller", "method", req.Method)
		return c.optimizedController.CreateAndRunTest(ctx, req)
	}

	if req.Method == "ludicrous" {
		log.Info("delegating to ludicrous controller", "method", req.Method)
		return c.ludicrousController.CreateAndRunTest(ctx, req)
	}

	// Create the LoadTest record with fixed column structure
	// We use a fixed structure: 5 date columns + 20 regular columns = 25 total
	const FixedTotalColumns = 25
	const FixedDateColumns = 5 // We populate all 5 available date columns

	loadTest := &LoadTest{
		Rows:        req.Rows,
		Columns:     FixedTotalColumns, // Override: always use 25 columns
		DateColumns: FixedDateColumns,  // Override: always populate 6 date columns
		Method:      req.Method,
		Status:      "running",
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

// PerformanceSummary represents performance metrics grouped by test method
type PerformanceSummary struct {
	Method        string `json:"method"`
	TestCount     int    `json:"testCount"`
	AvgRowsPerSec int    `json:"avgRowsPerSec"`
	MaxRowsPerSec int    `json:"maxRowsPerSec"`
	MinRowsPerSec int    `json:"minRowsPerSec"`
	P95RowsPerSec int    `json:"p95RowsPerSec"`
	AvgTotalTime  int    `json:"avgTotalTime"` // milliseconds
	MaxTotalTime  int    `json:"maxTotalTime"` // milliseconds
	MinTotalTime  int    `json:"minTotalTime"` // milliseconds
}

// GetPerformanceSummary retrieves performance statistics grouped by test method
func (c *LoadTestController) GetPerformanceSummary(
	ctx context.Context,
) ([]*PerformanceSummary, error) {
	log := c.log.Function("GetPerformanceSummary")

	// Get all completed load tests (not limited to recent 10)
	allTests, err := c.loadTestRepo.GetAllForSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get load tests: %w", err)
	}

	// Filter completed tests and group by method
	methodGroups := make(map[string][]*LoadTest)
	for _, test := range allTests {
		if test.Status == "completed" && test.TotalTime != nil && *test.TotalTime > 0 {
			methodGroups[test.Method] = append(methodGroups[test.Method], test)
		}
	}

	var summaries []*PerformanceSummary

	// Calculate statistics for each method
	for method, tests := range methodGroups {
		if len(tests) == 0 {
			continue
		}

		summary := &PerformanceSummary{
			Method:    method,
			TestCount: len(tests),
		}

		// Calculate rows per second for each test and collect stats
		var rowsPerSecondValues []int
		var totalTimeValues []int

		for _, test := range tests {
			if test.TotalTime != nil && *test.TotalTime > 0 {
				rowsPerSec := int(float64(test.Rows) / (float64(*test.TotalTime) / 1000.0))
				rowsPerSecondValues = append(rowsPerSecondValues, rowsPerSec)
				totalTimeValues = append(totalTimeValues, *test.TotalTime)
			}
		}

		if len(rowsPerSecondValues) > 0 {
			// Sort for percentile calculations
			sortedRPS := make([]int, len(rowsPerSecondValues))
			copy(sortedRPS, rowsPerSecondValues)
			for i := 0; i < len(sortedRPS); i++ {
				for j := i + 1; j < len(sortedRPS); j++ {
					if sortedRPS[i] > sortedRPS[j] {
						sortedRPS[i], sortedRPS[j] = sortedRPS[j], sortedRPS[i]
					}
				}
			}

			sortedTimes := make([]int, len(totalTimeValues))
			copy(sortedTimes, totalTimeValues)
			for i := 0; i < len(sortedTimes); i++ {
				for j := i + 1; j < len(sortedTimes); j++ {
					if sortedTimes[i] > sortedTimes[j] {
						sortedTimes[i], sortedTimes[j] = sortedTimes[j], sortedTimes[i]
					}
				}
			}

			// Calculate averages
			sumRPS := 0
			sumTime := 0
			for i, rps := range rowsPerSecondValues {
				sumRPS += rps
				sumTime += totalTimeValues[i]
			}
			summary.AvgRowsPerSec = sumRPS / len(rowsPerSecondValues)
			summary.AvgTotalTime = sumTime / len(totalTimeValues)

			// Min/Max
			summary.MinRowsPerSec = sortedRPS[0]
			summary.MaxRowsPerSec = sortedRPS[len(sortedRPS)-1]
			summary.MinTotalTime = sortedTimes[0]
			summary.MaxTotalTime = sortedTimes[len(sortedTimes)-1]

			// P95 (95th percentile)
			p95Index := int(float64(len(sortedRPS)) * 0.95)
			if p95Index >= len(sortedRPS) {
				p95Index = len(sortedRPS) - 1
			}
			summary.P95RowsPerSec = sortedRPS[p95Index]
		}

		summaries = append(summaries, summary)
	}

	log.Info("performance summary calculated", "methodCount", len(summaries))

	return summaries, nil
}

// processLoadTest handles the actual load test processing
func (c *LoadTestController) processLoadTest(ctx context.Context, loadTest *LoadTest) {
	log := c.log.Function("processLoadTest")
	testID := loadTest.ID.String()

	// Send initial progress
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "csv_generation",
		"overallProgress": 0,
		"phaseProgress":   0,
		"currentPhase":    "Generating CSV Data",
		"rowsProcessed":   0,
		"rowsPerSecond":   0,
		"eta":             "Calculating...",
		"message":         "Starting CSV generation...",
	})

	// Step 1: Generate CSV file using shared utility
	csvConfig := utils.CSVGenerationConfig{
		LoadTestID:  loadTest.ID,
		Rows:        loadTest.Rows,
		DateColumns: loadTest.DateColumns,
		FilePrefix:  "load_test",
		Context:     ctx,
	}
	csvResult, err := utils.GeneratePerformanceCSV(csvConfig)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "CSV generation failed", err)
		c.wsManager.SendLoadTestError(testID, "CSV generation failed: "+err.Error())
		return
	}
	csvPath := csvResult.FilePath
	csvGenTime := csvResult.GenerationTime
	defer os.Remove(csvPath)

	if loadTest.Method == "plaid" {
		// Go directly to plaid COPY streaming insertion
		c.wsManager.SendLoadTestProgress(testID, map[string]any{
			"phase":           "insertion",
			"overallProgress": 25,
			"phaseProgress":   0,
			"currentPhase":    "Plaid COPY Insertion",
			"rowsProcessed":   0,
			"rowsPerSecond":   0,
			"eta":             "Calculating...",
			"message":         "Starting Plaid PostgreSQL COPY streaming insertion...",
		})

		timingResult, err := c.plaidController.RunPlaidCopy(
			ctx,
			csvPath,
			loadTest.ID,
			loadTest.Rows,
		)
		if err != nil {
			c.updateLoadTestError(ctx, loadTest, "Plaid COPY insertion failed", err)
			c.wsManager.SendLoadTestError(testID, "Plaid COPY insertion failed: "+err.Error())
			return
		}

		// Update load test with completion data
		loadTest.CSVGenTime = &csvGenTime
		loadTest.ParseTime = &timingResult.ParseTime
		loadTest.InsertTime = &timingResult.InsertTime
		loadTest.TotalTime = &timingResult.TotalTime
		loadTest.Status = "completed"

		if err := c.loadTestRepo.Update(ctx, loadTest); err != nil {
			_ = log.Err("failed to update completed load test", err, "loadTestId", loadTest.ID)
		}

		// Send completion notification
		c.wsManager.SendLoadTestComplete(testID, map[string]any{
			"id":          loadTest.ID.String(),
			"rows":        loadTest.Rows,
			"columns":     loadTest.Columns,
			"dateColumns": loadTest.DateColumns,
			"method":      loadTest.Method,
			"status":      "completed",
			"csvGenTime":  csvGenTime,
			"parseTime":   timingResult.ParseTime,
			"insertTime":  timingResult.InsertTime,
			"totalTime":   timingResult.TotalTime,
		})

		log.Info("plaid load test completed successfully",
			"loadTestId", loadTest.ID,
			"totalTime", timingResult.TotalTime,
			"method", loadTest.Method)
		return
	}

	// Traditional flow for brute_force and batched methods
	// Progress update: CSV generation complete
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "parsing",
		"overallProgress": 25,
		"phaseProgress":   0,
		"currentPhase":    "Parsing and Validating",
		"rowsProcessed":   0,
		"rowsPerSecond":   0,
		"eta":             "Calculating...",
		"message":         "Starting data parsing and validation...",
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
		"phase":           "insertion",
		"overallProgress": 85,
		"phaseProgress":   0,
		"currentPhase":    "Inserting into Database",
		"rowsProcessed":   0,
		"rowsPerSecond":   0,
		"eta":             "Calculating...",
		"message": fmt.Sprintf(
			"Starting database insertion using %s method...",
			loadTest.Method,
		),
	})

	// Step 3: Insert data using specified method
	insertTime, err := c.insertTestDataWithProgress(ctx, testData, loadTest.Method, testID)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "Data insertion failed", err)
		c.wsManager.SendLoadTestError(testID, "Data insertion failed: "+err.Error())
		return
	}

	// Step 4: Update load test with completion data
	// Note: totalTime excludes CSV generation as that's test setup, not performance measurement
	totalTime := parseTime + insertTime
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
		"id":          loadTest.ID.String(),
		"rows":        loadTest.Rows,
		"columns":     loadTest.Columns,
		"dateColumns": loadTest.DateColumns,
		"method":      loadTest.Method,
		"status":      "completed",
		"csvGenTime":  csvGenTime,
		"parseTime":   parseTime,
		"insertTime":  insertTime,
		"totalTime":   totalTime,
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

	// Create single RNG instance for better performance
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Pre-compute column maps for better performance
	selectedDateColumnMap := make(map[string]bool)
	for _, col := range selectedDateColumns {
		selectedDateColumnMap[col] = true
	}

	allDateColumnMap := make(map[string]bool)
	for _, col := range allDateColumns {
		allDateColumnMap[col] = true
	}

	// Generate data rows with optimized approach
	for i := 0; i < loadTest.Rows; i++ {
		row := c.generateDataRowOptimized(headers, selectedDateColumnMap, allDateColumnMap, rng)
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
func (c *LoadTestController) generateRandomizedHeaders(
	totalColumns int,
	knownDateColumns []string,
) []string {
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

// generateDataRowOptimized creates a single data row with pre-computed maps for better performance
func (c *LoadTestController) generateDataRowOptimized(
	headers []string,
	selectedDateColumnMap, allDateColumnMap map[string]bool,
	rng *rand.Rand,
) []string {
	row := make([]string, len(headers))

	for i, header := range headers {
		if allDateColumnMap[header] {
			// This is a date column - populate only if it's selected
			if selectedDateColumnMap[header] {
				// Generate date data in various formats
				row[i] = c.generateDateValueOptimized(rng)
			} else {
				// Leave empty for unselected date columns
				row[i] = ""
			}
		} else {
			// Generate appropriate data for meaningful columns
			row[i] = c.generateMeaningfulColumnValueOptimized(header, rng)
		}
	}

	return row
}

// generateDataRow creates a single data row with appropriate data for each column (kept for backward compatibility)
func (c *LoadTestController) generateDataRow(
	headers []string,
	selectedDateColumns []string,
) []string {
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

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

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
			row[i] = c.generateMeaningfulColumnValueOptimized(header, rng)
		}
	}

	return row
}

// generateDateValueOptimized creates a random date optimized for performance
func (c *LoadTestController) generateDateValueOptimized(rng *rand.Rand) string {
	// Pre-defined date formats and patterns for better performance
	formats := []string{
		"01/02/2006", "1/2/2006", "01/02/06", "1/2/06",
		"01-02-2006", "1-2-2006", "01-02-06", "1-2-06",
		"2006-01-02", "2006/01/02",
	}

	// Generate random date between 2020 and 2024
	year := 2020 + rng.Intn(5)
	month := rng.Intn(12) + 1
	day := rng.Intn(28) + 1 // Use 28 to avoid month-specific logic

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	format := formats[rng.Intn(len(formats))]

	return date.Format(format)
}

// generateDateValue creates a random date in various formats using our date utils (kept for backward compatibility)
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

// Pre-defined data sets for optimized generation (defined at package level for better performance)
var (
	firstNames = []string{
		"John",
		"Jane",
		"Michael",
		"Sarah",
		"David",
		"Lisa",
		"Robert",
		"Mary",
		"James",
		"Jennifer",
		"William",
		"Patricia",
		"Richard",
		"Linda",
		"Charles",
		"Barbara",
		"Joseph",
		"Elizabeth",
		"Thomas",
		"Dorothy",
	}
	lastNames = []string{
		"Smith",
		"Johnson",
		"Williams",
		"Brown",
		"Jones",
		"Garcia",
		"Miller",
		"Davis",
		"Rodriguez",
		"Martinez",
		"Hernandez",
		"Lopez",
		"Gonzalez",
		"Wilson",
		"Anderson",
		"Thomas",
		"Taylor",
		"Moore",
		"Jackson",
		"Martin",
	}
	emailFirsts = []string{
		"john",
		"jane",
		"mike",
		"sarah",
		"david",
		"lisa",
		"bob",
		"mary",
		"james",
		"jen",
		"alex",
		"chris",
		"pat",
		"sam",
		"casey",
		"jordan",
		"taylor",
		"morgan",
		"riley",
		"drew",
	}
	emailDomains = []string{
		"gmail.com",
		"yahoo.com",
		"outlook.com",
		"company.com",
		"example.org",
		"hotmail.com",
		"aol.com",
		"icloud.com",
		"protonmail.com",
		"mail.com",
	}
	streetNumbers = []int{
		123,
		456,
		789,
		1011,
		1234,
		5678,
		2468,
		1357,
		9876,
		5432,
		1111,
		2222,
		3333,
		4444,
		5555,
		6666,
		7777,
		8888,
		9999,
		1024,
	}
	streetNames = []string{
		"Main St",
		"Oak Ave",
		"First St",
		"Park Blvd",
		"Elm St",
		"Cedar Ave",
		"Pine Rd",
		"Maple Dr",
		"Washington St",
		"Lincoln Ave",
		"Jefferson Blvd",
		"Adams Way",
		"Madison Ct",
		"Monroe Dr",
		"Jackson St",
		"Franklin Ave",
		"Roosevelt Rd",
		"Kennedy Blvd",
		"Wilson St",
		"Taylor Ave",
	}
	cities = []string{
		"New York",
		"Los Angeles",
		"Chicago",
		"Houston",
		"Phoenix",
		"Philadelphia",
		"San Antonio",
		"San Diego",
		"Dallas",
		"Austin",
		"Jacksonville",
		"Fort Worth",
		"Columbus",
		"Indianapolis",
		"Charlotte",
		"San Francisco",
		"Seattle",
		"Denver",
		"Washington",
		"Boston",
	}
	states = []string{
		"CA",
		"TX",
		"NY",
		"FL",
		"IL",
		"PA",
		"OH",
		"GA",
		"NC",
		"MI",
		"WA",
		"AZ",
		"MA",
		"VA",
		"CO",
		"MD",
		"MN",
		"MO",
		"WI",
		"OR",
	}
	countries = []string{
		"United States",
		"Canada",
		"Mexico",
		"United Kingdom",
		"Germany",
		"France",
		"Japan",
		"Australia",
		"Italy",
		"Spain",
		"Netherlands",
		"Sweden",
		"Norway",
		"Denmark",
		"Finland",
		"Belgium",
		"Switzerland",
		"Austria",
		"Portugal",
		"Ireland",
	}
	companies = []string{
		"Acme Corp",
		"Tech Solutions",
		"Global Industries",
		"Metro Systems",
		"Alpha Enterprises",
		"Beta LLC",
		"Gamma Inc",
		"Delta Group",
		"Omega Solutions",
		"Prime Industries",
		"Elite Systems",
		"Superior Corp",
		"Advanced Tech",
		"Innovative Solutions",
		"Dynamic Systems",
		"Strategic Industries",
		"Premier Group",
		"Excellence Corp",
		"Quality Solutions",
		"Performance Systems",
	}
	jobTitles = []string{
		"Software Engineer",
		"Manager",
		"Analyst",
		"Coordinator",
		"Specialist",
		"Director",
		"Associate",
		"Consultant",
		"Administrator",
		"Developer",
		"Designer",
		"Architect",
		"Lead",
		"Senior",
		"Junior",
		"Principal",
		"Staff",
		"Team Lead",
		"Project Manager",
		"Product Manager",
	}
	departments = []string{
		"Engineering",
		"Sales",
		"Marketing",
		"HR",
		"Finance",
		"Operations",
		"IT",
		"Customer Service",
		"Legal",
		"Research",
		"Development",
		"Quality Assurance",
		"Security",
		"Procurement",
		"Training",
		"Facilities",
		"Compliance",
		"Strategy",
		"Business Development",
		"Analytics",
	}
	insuranceCarriers = []string{
		"Blue Cross Blue Shield",
		"Aetna",
		"Cigna",
		"UnitedHealthcare",
		"Humana",
		"Kaiser Permanente",
		"Anthem",
		"Molina Healthcare",
		"Centene",
		"WellCare",
		"Independence Blue Cross",
		"Highmark",
		"BCBS of Michigan",
		"Florida Blue",
		"Premera Blue Cross",
		"CareFirst",
		"Excellus",
		"HCSC",
		"Caresource",
		"BCBS of North Carolina",
	}
)

// generateMeaningfulColumnValueOptimized creates appropriate fake data with pre-allocated data and shared RNG
func (c *LoadTestController) generateMeaningfulColumnValueOptimized(
	columnName string,
	rng *rand.Rand,
) string {
	switch columnName {
	case "first_name":
		return firstNames[rng.Intn(len(firstNames))]
	case "last_name":
		return lastNames[rng.Intn(len(lastNames))]
	case "email":
		return emailFirsts[rng.Intn(len(emailFirsts))] + fmt.Sprint(
			rng.Intn(999),
		) + "@" + emailDomains[rng.Intn(len(emailDomains))]
	case "phone":
		return fmt.Sprintf(
			"(%03d) %03d-%04d",
			rng.Intn(800)+200,
			rng.Intn(800)+200,
			rng.Intn(10000),
		)
	case "address_line_1":
		return fmt.Sprint(
			streetNumbers[rng.Intn(len(streetNumbers))],
		) + " " + streetNames[rng.Intn(len(streetNames))]
	case "address_line_2":
		if rng.Intn(3) == 0 { // 33% chance of having apartment/unit
			return "Apt " + fmt.Sprint(rng.Intn(999)+1)
		}
		return "" // Often empty
	case "city":
		return cities[rng.Intn(len(cities))]
	case "state":
		return states[rng.Intn(len(states))]
	case "zip_code":
		return fmt.Sprintf("%05d", rng.Intn(99999))
	case "country":
		return countries[rng.Intn(len(countries))]
	case "social_security_no":
		return fmt.Sprintf("***-**-%04d", rng.Intn(10000)) // Masked for privacy
	case "employer":
		return companies[rng.Intn(len(companies))]
	case "job_title":
		return jobTitles[rng.Intn(len(jobTitles))]
	case "department":
		return departments[rng.Intn(len(departments))]
	case "salary":
		return "$" + fmt.Sprint((rng.Intn(100)+40)*1000) // $40k to $140k
	case "insurance_plan_id":
		return fmt.Sprintf("PLAN-%03d", rng.Intn(999)+1)
	case "insurance_carrier":
		return insuranceCarriers[rng.Intn(len(insuranceCarriers))]
	case "policy_number":
		return fmt.Sprintf("POL-%d-%d", rng.Intn(9999)+1000, rng.Intn(999999)+100000)
	case "group_number":
		return fmt.Sprintf("GRP%03d", rng.Intn(999)+1)
	case "member_id":
		return fmt.Sprintf("MEM%d%03d", rng.Intn(99)+10, rng.Intn(999)+1)
	default:
		// Fallback for unknown columns
		return "data_" + fmt.Sprint(rng.Intn(9999))
	}
}

// generateMeaningfulColumnValue creates appropriate fake data for meaningful columns (kept for backward compatibility)
func (c *LoadTestController) generateMeaningfulColumnValue(columnName string) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(1000))))

	switch columnName {
	case "first_name":
		names := []string{
			"John",
			"Jane",
			"Michael",
			"Sarah",
			"David",
			"Lisa",
			"Robert",
			"Mary",
			"James",
			"Jennifer",
		}
		return names[r.Intn(len(names))]
	case "last_name":
		surnames := []string{
			"Smith",
			"Johnson",
			"Williams",
			"Brown",
			"Jones",
			"Garcia",
			"Miller",
			"Davis",
			"Rodriguez",
			"Martinez",
		}
		return surnames[r.Intn(len(surnames))]
	case "email":
		first := []string{
			"john",
			"jane",
			"mike",
			"sarah",
			"david",
			"lisa",
			"bob",
			"mary",
			"james",
			"jen",
		}
		domains := []string{"gmail.com", "yahoo.com", "outlook.com", "company.com", "example.org"}
		return fmt.Sprintf(
			"%s%d@%s",
			first[r.Intn(len(first))],
			r.Intn(999),
			domains[r.Intn(len(domains))],
		)
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
		cities := []string{
			"New York",
			"Los Angeles",
			"Chicago",
			"Houston",
			"Phoenix",
			"Philadelphia",
			"San Antonio",
			"San Diego",
			"Dallas",
			"Austin",
		}
		return cities[r.Intn(len(cities))]
	case "state":
		states := []string{"CA", "TX", "NY", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
		return states[r.Intn(len(states))]
	case "zip_code":
		return fmt.Sprintf("%05d", r.Intn(99999))
	case "country":
		countries := []string{
			"United States",
			"Canada",
			"Mexico",
			"United Kingdom",
			"Germany",
			"France",
			"Japan",
			"Australia",
		}
		return countries[r.Intn(len(countries))]
	case "social_security_no":
		return fmt.Sprintf("***-**-%04d", r.Intn(10000)) // Masked for privacy
	case "employer":
		companies := []string{
			"Acme Corp",
			"Tech Solutions",
			"Global Industries",
			"Metro Systems",
			"Alpha Enterprises",
			"Beta LLC",
			"Gamma Inc",
			"Delta Group",
		}
		return companies[r.Intn(len(companies))]
	case "job_title":
		titles := []string{
			"Software Engineer",
			"Manager",
			"Analyst",
			"Coordinator",
			"Specialist",
			"Director",
			"Associate",
			"Consultant",
			"Administrator",
		}
		return titles[r.Intn(len(titles))]
	case "department":
		departments := []string{
			"Engineering",
			"Sales",
			"Marketing",
			"HR",
			"Finance",
			"Operations",
			"IT",
			"Customer Service",
			"Legal",
		}
		return departments[r.Intn(len(departments))]
	case "salary":
		return fmt.Sprintf("$%d", (r.Intn(100)+40)*1000) // $40k to $140k
	case "insurance_plan_id":
		return fmt.Sprintf("PLAN-%03d", r.Intn(999)+1)
	case "insurance_carrier":
		carriers := []string{
			"Blue Cross Blue Shield",
			"Aetna",
			"Cigna",
			"UnitedHealthcare",
			"Humana",
			"Kaiser Permanente",
			"Anthem",
		}
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
func (c *LoadTestController) parseAndValidateCSVWithProgress(
	csvPath string,
	loadTest *LoadTest,
) ([]*TestData, int, error) {
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
			LoadTestID: loadTest.ID,
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
		log.Warn(
			"validation errors encountered",
			"errorCount",
			len(validationErrors),
			"errorRate",
			fmt.Sprintf("%.2f%%", float64(len(validationErrors))/float64(rowCount)*100),
			"firstFewErrors",
			validationErrors[:min(5, len(validationErrors))],
		)
	}

	return testData, parseTime, nil
}

// parseAndValidateRow parses a single CSV row into a TestData struct with validation
func (c *LoadTestController) parseAndValidateRow(
	record []string,
	headers []string,
	headerIndex map[string]int,
	data *TestData,
) error {
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
				validationErrors = append(
					validationErrors,
					fmt.Sprintf("failed to set %s: %v", header, err),
				)
			}

			// Additional validation reporting for non-empty values
			if value != "" {
				if _, isValid := c.validateAndNormalizeDateValue(value); !isValid {
					validationErrors = append(
						validationErrors,
						fmt.Sprintf("invalid date in %s: %s", header, value),
					)
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
		// CreatedAt field removed with BaseModel - ignoring this column
	case "updated_at":
		// UpdatedAt field removed with BaseModel - ignoring this column
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
func (c *LoadTestController) insertTestDataWithProgress(
	ctx context.Context,
	testData []*TestData,
	method string,
	testID string,
) (int, error) {
	log := c.log.Function("insertTestData")
	startTime := time.Now()

	log.Info("data insertion started", "method", method, "recordCount", len(testData))

	switch method {
	case "brute_force":
		return c.insertBruteForceWithProgress(ctx, testData, startTime, testID)
	case "batched":
		return c.insertBatchedWithProgress(ctx, testData, startTime, testID)
	case "plaid":
		// Plaid method should not reach this point as it uses streaming insertion
		return 0, fmt.Errorf("plaid method should use streaming insertion, not traditional flow")
	default:
		return 0, fmt.Errorf("unknown insertion method: %s", method)
	}
}

// insertBruteForceWithProgress performs individual inserts for each record
func (c *LoadTestController) insertBruteForceWithProgress(
	ctx context.Context,
	testData []*TestData,
	startTime time.Time,
	testID string,
) (int, error) {
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
				"phase":           "insertion",
				"overallProgress": 85 + (progress * 0.15), // 85% to 100%
				"phaseProgress":   progress,
				"currentPhase":    "Inserting into Database",
				"rowsProcessed":   i,
				"rowsPerSecond":   rowsPerSecond,
				"eta":             eta,
				"message": fmt.Sprintf(
					"Inserting records using brute force method (%d/%d)...",
					i,
					len(testData),
				),
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
		return insertTime, fmt.Errorf(
			"brute force insertion completed with %d errors out of %d records",
			errorCount,
			len(testData),
		)
	}

	return insertTime, nil
}

// insertBatchedWithProgress performs batch inserts with progress monitoring
func (c *LoadTestController) insertBatchedWithProgress(
	ctx context.Context,
	testData []*TestData,
	startTime time.Time,
	testID string,
) (int, error) {
	log := c.log.Function("insertBatched")

	// Use default batch size of 2000, or configure based on data size
	batchSize := 2000
	if len(testData) < 100 {
		batchSize = len(testData) // Use smaller batches for small datasets
	}

	totalRecords := len(testData)
	loadTestUUID := testData[0].LoadTestID // All records have the same LoadTestID

	// Send initial progress update
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 85,
		"phaseProgress":   0,
		"currentPhase":    "Inserting into Database",
		"rowsProcessed":   0,
		"rowsPerSecond":   0,
		"eta":             "Calculating...",
		"message": fmt.Sprintf(
			"Starting batched insertion (%d records in batches of %d)...",
			totalRecords,
			batchSize,
		),
	})

	// Start progress monitoring goroutine
	progressDone := make(chan bool)
	go c.monitorInsertProgress(
		ctx,
		loadTestUUID.String(),
		totalRecords,
		startTime,
		testID,
		progressDone,
	)

	// Perform the actual batch insertion (single call - let GORM handle internal batching)
	err := c.testDataRepo.CreateBatch(ctx, testData, batchSize)

	// Stop progress monitoring
	close(progressDone)

	insertTime := int(time.Since(startTime).Milliseconds())

	if err != nil {
		_ = log.Err(
			"batched insertion failed",
			err,
			"recordCount",
			len(testData),
			"batchSize",
			batchSize,
		)
		return insertTime, fmt.Errorf("batch insertion failed: %w", err)
	}

	rowsPerSecond := int(float64(totalRecords) / (float64(insertTime) / 1000))

	// Send final progress update
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 100,
		"phaseProgress":   100,
		"currentPhase":    "Database Insertion Complete",
		"rowsProcessed":   totalRecords,
		"rowsPerSecond":   rowsPerSecond,
		"eta":             "Done",
		"message": fmt.Sprintf(
			"Successfully inserted all %d records using batched method",
			totalRecords,
		),
	})

	log.Info("batched insertion completed",
		"totalRecords", totalRecords,
		"batchSize", batchSize,
		"insertTimeMs", insertTime,
		"rowsPerSecond", rowsPerSecond)

	return insertTime, nil
}

// monitorInsertProgress monitors the insertion progress by periodically checking the database count
func (c *LoadTestController) monitorInsertProgress(
	ctx context.Context,
	loadTestID string,
	totalRecords int,
	startTime time.Time,
	testID string,
	done chan bool,
) {
	log := c.log.Function("monitorInsertProgress")

	ticker := time.NewTicker(2 * time.Second) // Check every 2 seconds
	defer ticker.Stop()

	lastCount := int64(0)
	lastCheckTime := startTime

	for {
		select {
		case <-done:
			log.Debug("Progress monitoring stopped", "testID", testID)
			return
		case <-ticker.C:
			// Get current count of inserted records
			count, err := c.testDataRepo.CountByLoadTestID(ctx, loadTestID)
			if err != nil {
				log.Warn(
					"Failed to get count for progress monitoring",
					"error",
					err,
					"testID",
					testID,
				)
				continue
			}

			log.Debug("Progress monitoring check",
				"testID", testID,
				"currentCount", count,
				"lastCount", lastCount,
				"totalRecords", totalRecords)

			currentTime := time.Now()
			elapsed := currentTime.Sub(startTime)

			// Calculate metrics
			overallProgress := float64(count) / float64(totalRecords)
			phaseProgress := overallProgress * 100
			finalOverallProgress := 85 + (overallProgress * 15) // 85% to 100%

			// Calculate rows per second (overall rate)
			rowsPerSecond := int(float64(count) / elapsed.Seconds())
			if elapsed.Seconds() < 1 {
				rowsPerSecond = 0 // Avoid division by very small numbers
			}

			// Calculate ETA based on current progress
			eta := "Calculating..."
			if count > lastCount && rowsPerSecond > 0 {
				remaining := totalRecords - int(count)
				if remaining > 0 {
					etaSeconds := remaining / rowsPerSecond
					if etaSeconds < 60 {
						eta = fmt.Sprintf("%ds", etaSeconds)
					} else if etaSeconds < 3600 {
						eta = fmt.Sprintf("%dm %ds", etaSeconds/60, etaSeconds%60)
					} else {
						eta = fmt.Sprintf("%dh %dm", etaSeconds/3600, (etaSeconds%3600)/60)
					}
				} else {
					eta = "Done"
				}
			}

			// Only send update if there's been progress or it's been a while
			timeSinceLastCheck := currentTime.Sub(lastCheckTime)
			if count > lastCount || timeSinceLastCheck > 2*time.Second {
				c.wsManager.SendLoadTestProgress(testID, map[string]any{
					"phase":           "insertion",
					"overallProgress": finalOverallProgress,
					"phaseProgress":   phaseProgress,
					"currentPhase":    "Inserting into Database",
					"rowsProcessed":   int(count),
					"rowsPerSecond":   rowsPerSecond,
					"eta":             eta,
					"message": fmt.Sprintf(
						"Inserted %d/%d records (%s elapsed)...",
						count,
						totalRecords,
						elapsed.Round(time.Second),
					),
				})

				lastCount = count
				lastCheckTime = currentTime

				log.Debug("Progress update sent",
					"insertedCount", count,
					"totalRecords", totalRecords,
					"progress", fmt.Sprintf("%.1f%%", phaseProgress),
					"rowsPerSecond", rowsPerSecond)
			}

			// If we've reached the total, we can exit early
			if count >= int64(totalRecords) {
				log.Debug("All records inserted, stopping progress monitoring", "testID", testID)
				return
			}
		}
	}
}

// updateLoadTestError updates the load test with error information
func (c *LoadTestController) updateLoadTestError(
	ctx context.Context,
	loadTest *LoadTest,
	message string,
	err error,
) {
	log := c.log.Function("updateLoadTestError")

	errorMsg := message + ": " + err.Error()
	loadTest.Status = "failed"
	loadTest.ErrorMessage = &errorMsg

	if updateErr := c.loadTestRepo.Update(ctx, loadTest); updateErr != nil {
		_ = log.Err("failed to update load test error", updateErr, "loadTestId", loadTest.ID)
	}

	_ = log.Err(message, err, "loadTestId", loadTest.ID)
}
