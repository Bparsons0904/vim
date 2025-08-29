package controllers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"server/internal/utils"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type LudicrousOnlyController struct {
	loadTestRepo repositories.LoadTestRepository
	testDataRepo repositories.TestDataRepository
	dateUtils    *utils.DateUtils
	log          logger.Logger
	wsManager    WSManager
	db           database.DB
}

func NewLudicrousOnlyController(
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
	db database.DB,
	wsManager WSManager,
	config config.Config,
) *LudicrousOnlyController {
	return &LudicrousOnlyController{
		loadTestRepo: loadTestRepo,
		testDataRepo: testDataRepo,
		dateUtils:    utils.NewDateUtils(),
		log:          logger.New("ludicrousOnlyController"),
		wsManager:    wsManager,
		db:           db,
	}
}

// CreateAndRunTest creates a new ludicrous speed load test and starts the performance test
func (c *LudicrousOnlyController) CreateAndRunTest(
	ctx context.Context,
	req *CreateLoadTestRequest,
) (*LoadTest, error) {
	log := c.log.Function("CreateAndRunTest")

	// Force method to ludicrous
	const FixedTotalColumns = 25
	const FixedDateColumns = 5

	loadTest := &LoadTest{
		// Let GORM handle ID generation automatically
		Rows:        req.Rows,
		Columns:     FixedTotalColumns,
		DateColumns: FixedDateColumns,
		Method:      "ludicrous", // Force ludicrous method
		Status:      "running",
	}

	if err := c.loadTestRepo.Create(ctx, loadTest); err != nil {
		_ = log.Err("failed to create load test", err, "loadTest", loadTest)
		return nil, fmt.Errorf("failed to create load test: %w", err)
	}

	// Process the load test asynchronously with recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("processLoadTest goroutine panicked", "panic", r, "loadTestId", loadTest.ID)
				c.updateLoadTestError(ctx, loadTest, "Internal processing error", fmt.Errorf("goroutine panic: %v", r))
				c.wsManager.SendLoadTestError(loadTest.ID.String(), fmt.Sprintf("Internal processing error: %v", r))
			}
		}()
		c.processLoadTest(ctx, loadTest)
	}()

	log.Info("ludicrous speed load test created and started", "loadTestId", loadTest.ID)
	return loadTest, nil
}

// processLoadTest handles the ludicrous speed load test processing
func (c *LudicrousOnlyController) processLoadTest(ctx context.Context, loadTest *LoadTest) {
	log := c.log.Function("processLoadTest")
	testID := loadTest.ID.String()
	
	// Create a derived context with timeout for the entire operation
	processCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	
	// Check for cancellation before starting
	select {
	case <-processCtx.Done():
		c.updateLoadTestError(ctx, loadTest, "Process cancelled before start", processCtx.Err())
		c.wsManager.SendLoadTestError(testID, "Load test cancelled: "+processCtx.Err().Error())
		return
	default:
	}

	// Send initial progress
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "csv_generation",
		"overallProgress": 0,
		"phaseProgress":   0,
		"currentPhase":    "Generating CSV Data",
		"rowsProcessed":   0,
		"rowsPerSecond":   0,
		"eta":             "Calculating...",
		"message":         "Starting CSV generation for ludicrous speed method...",
	})

	// Step 1: Generate CSV file using shared utility
	csvConfig := utils.CSVGenerationConfig{
		LoadTestID:  loadTest.ID,
		Rows:        loadTest.Rows,
		DateColumns: loadTest.DateColumns,
		FilePrefix:  "ludicrous_test",
		Context:     processCtx,
	}
	csvResult, err := utils.GeneratePerformanceCSV(csvConfig)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "CSV generation failed", err)
		c.wsManager.SendLoadTestError(testID, "CSV generation failed: "+err.Error())
		return
	}
	csvPath := csvResult.FilePath
	csvGenTime := csvResult.GenerationTime
	
	// Check for cancellation after CSV generation
	select {
	case <-processCtx.Done():
		c.updateLoadTestError(ctx, loadTest, "Process cancelled after CSV generation", processCtx.Err())
		c.wsManager.SendLoadTestError(testID, "Load test cancelled: "+processCtx.Err().Error())
		return
	default:
	}

	// Step 2: Ludicrous speed streaming insertion
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 25,
		"phaseProgress":   0,
		"currentPhase":    "Ludicrous Speed Insertion",
		"rowsProcessed":   0,
		"rowsPerSecond":   0,
		"eta":             "Calculating...",
		"message":         "Starting ludicrous speed streaming insertion...",
	})

	// Start timing for parse + insert only (excluding CSV generation)
	parseInsertStartTime := time.Now()
	timingResult, err := c.insertLudicrousStreaming(
		processCtx,
		csvPath,
		loadTest.ID,
		loadTest.Rows,
		parseInsertStartTime,
		testID,
	)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "Ludicrous insertion failed", err)
		c.wsManager.SendLoadTestError(testID, "Ludicrous insertion failed: "+err.Error())
		return
	}

	// Update load test with completion data
	insertTime := timingResult.InsertTime
	parseTime := timingResult.ParseTime
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

	log.Info("ludicrous speed load test completed successfully",
		"loadTestId", loadTest.ID,
		"totalTime", totalTime)
}

// generateLudicrousCSVFile creates a CSV file optimized for ludicrous speed method
func (c *LudicrousOnlyController) generateLudicrousCSVFile(
	ctx context.Context,
	loadTest *LoadTest,
) (string, int, error) {
	log := c.log.Function("generateLudicrousCSVFile")
	startTime := time.Now()

	// Known date columns for ludicrous method
	allDateColumns := []string{
		"birth_date",
		"start_date",
		"end_date",
		"created_at",
		"updated_at",
	}

	// Meaningful columns for ludicrous method
	meaningfulColumns := []string{
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

	// Randomly select date columns to populate
	selectedDateColumns := c.selectRandomDateColumns(allDateColumns, loadTest.DateColumns)

	log.Info("Ludicrous CSV generation started",
		"rows", loadTest.Rows,
		"columns", loadTest.Columns,
		"selectedDateColumns", len(selectedDateColumns))

	// Create temp directory
	tempDir := "/tmp/load_tests"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}

	csvPath := filepath.Join(tempDir, "ludicrous_test_"+loadTest.ID.String()+".csv")

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

	// Generate headers - combine all columns and randomize
	var allColumns []string
	allColumns = append(allColumns, allDateColumns...)
	allColumns = append(allColumns, meaningfulColumns...)

	// Shuffle columns for ludicrous method
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(allColumns), func(i, j int) {
		allColumns[i], allColumns[j] = allColumns[j], allColumns[i]
	})

	if err := writer.Write(allColumns); err != nil {
		return "", 0, fmt.Errorf("failed to write headers: %w", err)
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

	// Generate data rows with ludicrous speed optimizations
	for i := 0; i < loadTest.Rows; i++ {
		// Check for cancellation every 10,000 rows for better performance
		if i > 0 && i%10000 == 0 {
			select {
			case <-ctx.Done():
				return "", 0, fmt.Errorf("CSV generation cancelled: %w", ctx.Err())
			default:
			}
		}
		
		row := c.generateLudicrousDataRow(allColumns, selectedDateColumnMap, allDateColumnMap, r)
		if err := writer.Write(row); err != nil {
			return "", 0, fmt.Errorf("failed to write row %d: %w", i, err)
		}

		// Less frequent progress logging for better performance
		if i > 0 && i%25000 == 0 {
			log.Debug(
				"Ludicrous CSV generation progress",
				"rowsGenerated",
				i,
				"totalRows",
				loadTest.Rows,
			)
		}
	}

	csvGenTime := int(time.Since(startTime).Milliseconds())
	log.Info("Ludicrous CSV generation completed",
		"csvPath", csvPath,
		"rows", loadTest.Rows,
		"generationTimeMs", csvGenTime)

	return csvPath, csvGenTime, nil
}

// selectRandomDateColumns randomly selects date columns for ludicrous method
func (c *LudicrousOnlyController) selectRandomDateColumns(
	allDateColumns []string,
	count int,
) []string {
	if count <= 0 {
		return []string{}
	}
	if count >= len(allDateColumns) {
		return allDateColumns
	}

	columns := make([]string, len(allDateColumns))
	copy(columns, allDateColumns)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(columns), func(i, j int) {
		columns[i], columns[j] = columns[j], columns[i]
	})

	return columns[:count]
}

// generateLudicrousDataRow creates a data row optimized for ludicrous speed method
func (c *LudicrousOnlyController) generateLudicrousDataRow(
	headers []string,
	selectedDateColumnMap, allDateColumnMap map[string]bool,
	rng *rand.Rand,
) []string {
	row := make([]string, len(headers))

	for i, header := range headers {
		if allDateColumnMap[header] {
			if selectedDateColumnMap[header] {
				row[i] = c.generateLudicrousDateValue(rng)
			} else {
				row[i] = ""
			}
		} else {
			row[i] = c.generateLudicrousMeaningfulColumnValue(header, rng)
		}
	}

	return row
}

// generateLudicrousDateValue creates a date value optimized for ludicrous speed
func (c *LudicrousOnlyController) generateLudicrousDateValue(rng *rand.Rand) string {
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

// Ludicrous speed data sets - optimized for maximum performance
var ludicrousDataSets = struct {
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
	},
	LastNames: []string{
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
	},
	EmailFirsts: []string{
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
	},
	EmailDomains: []string{
		"gmail.com",
		"yahoo.com",
		"outlook.com",
		"company.com",
		"example.org",
	},
	StreetNumbers: []int{123, 456, 789, 1011, 1234, 5678},
	StreetNames: []string{
		"Main St",
		"Oak Ave",
		"First St",
		"Park Blvd",
		"Elm St",
		"Cedar Ave",
	},
	Cities: []string{
		"New York",
		"Los Angeles",
		"Chicago",
		"Houston",
		"Phoenix",
		"Philadelphia",
	},
	States: []string{"CA", "TX", "NY", "FL", "IL", "PA"},
	Countries: []string{
		"United States",
		"Canada",
		"Mexico",
		"United Kingdom",
		"Germany",
		"France",
	},
	Companies: []string{
		"Acme Corp",
		"Tech Solutions",
		"Global Industries",
		"Metro Systems",
		"Alpha Enterprises",
		"Beta LLC",
	},
	JobTitles: []string{
		"Engineer",
		"Manager",
		"Analyst",
		"Coordinator",
		"Specialist",
		"Director",
	},
	Departments:       []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations"},
	InsuranceCarriers: []string{"Blue Cross", "Aetna", "Cigna", "United", "Humana", "Kaiser"},
}

// generateLudicrousMeaningfulColumnValue creates fake data optimized for ludicrous speed
func (c *LudicrousOnlyController) generateLudicrousMeaningfulColumnValue(
	columnName string,
	rng *rand.Rand,
) string {
	switch columnName {
	case "first_name":
		return ludicrousDataSets.FirstNames[rng.Intn(len(ludicrousDataSets.FirstNames))]
	case "last_name":
		return ludicrousDataSets.LastNames[rng.Intn(len(ludicrousDataSets.LastNames))]
	case "email":
		return ludicrousDataSets.EmailFirsts[rng.Intn(len(ludicrousDataSets.EmailFirsts))] +
			fmt.Sprint(rng.Intn(99)) + "@" +
			ludicrousDataSets.EmailDomains[rng.Intn(len(ludicrousDataSets.EmailDomains))]
	case "phone":
		return fmt.Sprintf("%03d%03d%04d", rng.Intn(800)+200, rng.Intn(800)+200, rng.Intn(10000))
	case "address_line_1":
		return fmt.Sprint(
			ludicrousDataSets.StreetNumbers[rng.Intn(len(ludicrousDataSets.StreetNumbers))],
		) +
			" " + ludicrousDataSets.StreetNames[rng.Intn(len(ludicrousDataSets.StreetNames))]
	case "address_line_2":
		if rng.Intn(4) == 0 { // 25% chance for better performance
			return "Apt " + fmt.Sprint(rng.Intn(99)+1)
		}
		return ""
	case "city":
		return ludicrousDataSets.Cities[rng.Intn(len(ludicrousDataSets.Cities))]
	case "state":
		return ludicrousDataSets.States[rng.Intn(len(ludicrousDataSets.States))]
	case "zip_code":
		return fmt.Sprintf("%05d", rng.Intn(99999))
	case "country":
		return ludicrousDataSets.Countries[rng.Intn(len(ludicrousDataSets.Countries))]
	case "social_security_no":
		return fmt.Sprintf("***%04d", rng.Intn(10000))
	case "employer":
		return ludicrousDataSets.Companies[rng.Intn(len(ludicrousDataSets.Companies))]
	case "job_title":
		return ludicrousDataSets.JobTitles[rng.Intn(len(ludicrousDataSets.JobTitles))]
	case "department":
		return ludicrousDataSets.Departments[rng.Intn(len(ludicrousDataSets.Departments))]
	case "salary":
		return fmt.Sprint((rng.Intn(100) + 40) * 1000)
	case "insurance_plan_id":
		return fmt.Sprintf("P%03d", rng.Intn(999)+1)
	case "insurance_carrier":
		return ludicrousDataSets.InsuranceCarriers[rng.Intn(len(ludicrousDataSets.InsuranceCarriers))]
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

// LudicrousTimingResult contains timing breakdown specific to ludicrous method
type LudicrousTimingResult struct {
	ParseTime  int
	InsertTime int
}

// insertLudicrousStreaming performs ludicrous speed streaming insertion
func (c *LudicrousOnlyController) insertLudicrousStreaming(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (LudicrousTimingResult, error) {
	log := c.log.Function("insertLudicrousStreaming")

	// Ludicrous speed configuration - maximize performance
	numWorkers := runtime.NumCPU() * 2 // Double the workers
	batchSize := 2000                  // Larger batches
	bufferSize := numWorkers * 8       // Larger buffer

	log.Info("Starting ludicrous speed streaming insertion",
		"totalRecords", totalRecords,
		"workers", numWorkers,
		"batchSize", batchSize)

	file, err := os.Open(csvPath)
	if err != nil {
		return LudicrousTimingResult{}, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Initialize progress tracking
	progress := &Progress{
		TotalRecords:     totalRecords,
		TotalBatches:     (totalRecords + batchSize - 1) / batchSize,
		StartTime:        startTime,
		RecordsProcessed: 0,
		BatchesProcessed: 0,
	}

	// Create channels for producer-consumer pattern
	batchChan := make(chan *BatchData, bufferSize)
	errorChan := make(chan error, numWorkers)

	// Start progress monitoring
	progressDone := make(chan bool)
	go c.monitorLudicrousProgress(progress, testID, progressDone)

	// Start worker goroutines
	var workerWG sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWG.Add(1)
		go c.ludicrousWorker(ctx, i, batchChan, errorChan, progress, batchSize, &workerWG)
	}

	// Start CSV parser
	parserDone := make(chan error, 1)
	parseStartTime := time.Now()
	go c.parseLudicrousCSVStreaming(file, loadTestID, batchChan, parserDone, batchSize)

	// Wait for parser or worker errors with comprehensive error handling
	var parseErr error
	var allErrors []error
	
	// Wait for parser with timeout
	parseComplete := false
	for !parseComplete {
		select {
		case parseErr = <-parserDone:
			parseComplete = true
			if parseErr != nil {
				allErrors = append(allErrors, fmt.Errorf("CSV parsing failed: %w", parseErr))
			}
		case workerErr := <-errorChan:
			if workerErr != nil {
				allErrors = append(allErrors, fmt.Errorf("worker failed: %w", workerErr))
			}
		case <-ctx.Done():
			allErrors = append(allErrors, fmt.Errorf("operation cancelled: %w", ctx.Err()))
			parseComplete = true
		case <-time.After(10 * time.Minute):
			allErrors = append(allErrors, fmt.Errorf("parsing timeout exceeded"))
			parseComplete = true
		}
		
		// If we have critical errors, stop immediately
		if len(allErrors) > 0 {
			break
		}
	}
	
	// Handle any accumulated errors
	if len(allErrors) > 0 {
		close(batchChan)
		progressDone <- true
		
		// Combine all errors into a single error message
		var errorMsgs []string
		for _, err := range allErrors {
			errorMsgs = append(errorMsgs, err.Error())
		}
		return LudicrousTimingResult{}, fmt.Errorf("multiple errors occurred: %s", strings.Join(errorMsgs, "; "))
	}
	parseTime := time.Since(parseStartTime)

	// Close batch channel and wait for workers
	close(batchChan)
	workerWG.Wait()
	close(errorChan)
	progressDone <- true

	// Collect all remaining worker errors
	var remainingErrors []error
	for err := range errorChan {
		if err != nil {
			remainingErrors = append(remainingErrors, err)
		}
	}
	
	// If we have any worker errors, return them
	if len(remainingErrors) > 0 {
		var errorMsgs []string
		for _, err := range remainingErrors {
			errorMsgs = append(errorMsgs, err.Error())
		}
		return LudicrousTimingResult{}, fmt.Errorf("worker errors: %s", strings.Join(errorMsgs, "; "))
	}

	totalParseInsertTime := int(time.Since(startTime).Milliseconds())

	// Send final progress update
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 100,
		"phaseProgress":   100,
		"currentPhase":    "Ludicrous Speed Complete",
		"rowsProcessed":   progress.RecordsProcessed,
		"rowsPerSecond":   int(float64(progress.RecordsProcessed) / (float64(totalParseInsertTime) / 1000)),
		"eta":             "Done",
		"message": fmt.Sprintf(
			"Successfully inserted %d records using ludicrous speed",
			progress.RecordsProcessed,
		),
	})

	// Calculate actual insert time (total - parse time)
	actualInsertTime := totalParseInsertTime - int(parseTime.Milliseconds())

	log.Info("ludicrous speed streaming insertion completed",
		"totalRecords", progress.RecordsProcessed,
		"parseTimeMs", parseTime.Milliseconds(),
		"insertTimeMs", actualInsertTime)

	return LudicrousTimingResult{
		ParseTime:  int(parseTime.Milliseconds()),
		InsertTime: actualInsertTime,
	}, nil
}

// parseLudicrousCSVStreaming parses CSV for ludicrous speed method
func (c *LudicrousOnlyController) parseLudicrousCSVStreaming(
	file *os.File,
	loadTestID uuid.UUID,
	batchChan chan<- *BatchData,
	done chan<- error,
	batchSize int,
) {
	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		done <- fmt.Errorf("failed to read CSV headers: %w", err)
		return
	}

	// Create simplified setters for ludicrous speed parsing
	setters := make([]func(td *TestData, val string), len(headers))
	for i, header := range headers {
		switch header {
		case "first_name":
			setters[i] = func(td *TestData, val string) { td.FirstName = &val }
		case "last_name":
			setters[i] = func(td *TestData, val string) { td.LastName = &val }
		case "email":
			setters[i] = func(td *TestData, val string) { td.Email = &val }
		case "phone":
			setters[i] = func(td *TestData, val string) { td.Phone = &val }
		case "address_line_1":
			setters[i] = func(td *TestData, val string) { td.AddressLine1 = &val }
		case "city":
			setters[i] = func(td *TestData, val string) { td.City = &val }
		case "state":
			setters[i] = func(td *TestData, val string) { td.State = &val }
		case "zip_code":
			setters[i] = func(td *TestData, val string) { td.ZipCode = &val }
		case "employer":
			setters[i] = func(td *TestData, val string) { td.Employer = &val }
		case "birth_date", "start_date", "end_date":
			h := header
			setters[i] = func(td *TestData, val string) {
				if val != "" {
					// Simplified date validation for ludicrous speed
					if len(val) >= 8 { // Basic length check
						switch h {
						case "birth_date":
							td.BirthDate = &val
						case "start_date":
							td.StartDate = &val
						case "end_date":
							td.EndDate = &val
						}
					}
				}
			}
		default:
			setters[i] = func(td *TestData, val string) {}
		}
	}

	currentBatch := make([]*TestData, 0, batchSize)
	batchNum := 0

	for {
		row, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			done <- fmt.Errorf("failed to read CSV row: %w", err)
			return
		}

		testData := &TestData{
			// Let database handle ID generation automatically
			LoadTestID: loadTestID,
		}

		// Use setters to populate struct - simplified for speed
		for i, value := range row {
			if i < len(setters) && value != "" {
				setters[i](testData, value)
			}
		}

		currentBatch = append(currentBatch, testData)

		// Send batch when full
		if len(currentBatch) >= batchSize {
			batchData := &BatchData{
				Records:  currentBatch,
				BatchNum: batchNum,
			}
			batchChan <- batchData
			currentBatch = make([]*TestData, 0, batchSize)
			batchNum++
		}
	}

	// Send remaining records
	if len(currentBatch) > 0 {
		batchData := &BatchData{
			Records:  currentBatch,
			BatchNum: batchNum,
		}
		batchChan <- batchData
	}

	done <- nil
}

func (c *LudicrousOnlyController) ludicrousWorker(
	ctx context.Context,
	workerID int,
	batchChan <-chan *BatchData,
	errorChan chan<- error,
	progress *Progress,
	batchSize int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Get the underlying SQL connection
	gormDB := c.db.SQLWithContext(ctx)
	sqlDB, err := gormDB.DB()
	if err != nil {
		errorChan <- fmt.Errorf("ludicrous worker %d failed to get SQL DB: %w", workerID, err)
		return
	}

	for batch := range batchChan {
		err := c.insertBatchWithRawSQLLudicrous(ctx, sqlDB, batch.Records)
		if err != nil {
			errorChan <- fmt.Errorf("ludicrous worker %d failed to insert batch %d: %w", workerID, batch.BatchNum, err)
			return
		}

		// Update progress
		progress.mu.Lock()
		progress.RecordsProcessed += len(batch.Records)
		progress.BatchesProcessed++
		progress.mu.Unlock()
	}

	errorChan <- nil
}

// insertBatchWithRawSQLLudicrous uses raw SQL optimized for ludicrous speed
func (c *LudicrousOnlyController) insertBatchWithRawSQLLudicrous(
	ctx context.Context,
	sqlDB *sql.DB,
	records []*TestData,
) error {
	if len(records) == 0 {
		return nil
	}

	// Simplified SQL for ludicrous speed - only essential columns
	const baseSQL = `INSERT INTO test_data (
		load_test_id, birth_date, start_date, end_date,
		first_name, last_name, email, phone, address_line1, city, state, zip_code, employer
	) VALUES `

	// Build VALUES clauses
	var valueClauses []string
	args := make(
		[]interface{},
		0,
		len(records)*13,
	) // 13 columns (removed id, created_at, updated_at)

	for i, record := range records {
		// Build placeholders
		placeholders := make([]string, 13)
		for j := 0; j < 13; j++ {
			placeholders[j] = fmt.Sprintf("$%d", i*13+j+1)
		}
		valueClauses = append(valueClauses, "("+strings.Join(placeholders, ", ")+")")

		// Helper function to safely dereference string pointers
		getStringValue := func(ptr *string) interface{} {
			if ptr == nil {
				return nil
			}
			return *ptr
		}

		args = append(args,
			record.LoadTestID,
			getStringValue(record.BirthDate),
			getStringValue(record.StartDate),
			getStringValue(record.EndDate),
			getStringValue(record.FirstName),
			getStringValue(record.LastName),
			getStringValue(record.Email),
			getStringValue(record.Phone),
			getStringValue(record.AddressLine1),
			getStringValue(record.City),
			getStringValue(record.State),
			getStringValue(record.ZipCode),
			getStringValue(record.Employer),
		)
	}

	// Combine SQL
	finalSQL := baseSQL + strings.Join(valueClauses, ", ")

	// Execute with context
	_, err := sqlDB.ExecContext(ctx, finalSQL, args...)
	if err != nil {
		return fmt.Errorf(
			"ludicrous raw SQL batch insert failed (records: %d): %w",
			len(records),
			err,
		)
	}

	return nil
}

// monitorLudicrousProgress sends progress updates for ludicrous speed method
func (c *LudicrousOnlyController) monitorLudicrousProgress(
	progress *Progress,
	testID string,
	done <-chan bool,
) {
	ticker := time.NewTicker(2 * time.Second) // Less frequent updates for better performance
	heartbeatTicker := time.NewTicker(10 * time.Second) // Heartbeat every 10 seconds
	defer ticker.Stop()
	defer heartbeatTicker.Stop()
	
	lastRecordsProcessed := 0
	stuckCount := 0

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			c.sendProgressUpdate(progress, testID, &lastRecordsProcessed, &stuckCount)
		case <-heartbeatTicker.C:
			// Send heartbeat with stall detection
			c.sendHeartbeat(progress, testID, &lastRecordsProcessed, &stuckCount)
		}
	}
}

func (c *LudicrousOnlyController) sendProgressUpdate(
	progress *Progress,
	testID string,
	lastRecordsProcessed *int,
	stuckCount *int,
) {
	log := c.log.Function("sendProgressUpdate")
	
	progress.mu.RLock()
	defer progress.mu.RUnlock()

	elapsed := time.Since(progress.StartTime)
	overallProgress := float64(progress.RecordsProcessed) / float64(progress.TotalRecords)
	phaseProgress := overallProgress * 100

	var eta string
	if progress.RecordsProcessed > 0 && overallProgress > 0 {
		totalEstimated := time.Duration(float64(elapsed) / overallProgress)
		remaining := totalEstimated - elapsed
		if remaining > 0 {
			eta = fmt.Sprintf("%ds", int(remaining.Seconds()))
		} else {
			eta = "Almost done"
		}
	} else {
		eta = "Calculating..."
	}

	rowsPerSecond := 0
	if elapsed.Seconds() > 0 {
		rowsPerSecond = int(float64(progress.RecordsProcessed) / elapsed.Seconds())
	}

	// Check for progress stalls
	if progress.RecordsProcessed == *lastRecordsProcessed {
		*stuckCount++
		if *stuckCount >= 5 { // 10 seconds without progress
			log.Warn("Progress appears stalled", "testID", testID, "stuckCount", *stuckCount)
		}
	} else {
		*stuckCount = 0
		*lastRecordsProcessed = progress.RecordsProcessed
	}

	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 85 + (15 * overallProgress),
		"phaseProgress":   phaseProgress,
		"currentPhase":    "Ludicrous Speed Insertion",
		"rowsProcessed":   progress.RecordsProcessed,
		"rowsPerSecond":   rowsPerSecond,
		"eta":             eta,
		"isStalled":       *stuckCount >= 5,
		"message": fmt.Sprintf(
			"Ludicrous speed: %d/%d records (%d batches)",
			progress.RecordsProcessed,
			progress.TotalRecords,
			progress.BatchesProcessed,
		),
	})
}

func (c *LudicrousOnlyController) sendHeartbeat(
	progress *Progress,
	testID string,
	lastRecordsProcessed *int,
	stuckCount *int,
) {
	log := c.log.Function("sendHeartbeat")
	
	progress.mu.RLock()
	currentRecords := progress.RecordsProcessed
	totalRecords := progress.TotalRecords
	elapsed := time.Since(progress.StartTime)
	progress.mu.RUnlock()

	// Check if we're truly stuck
	if currentRecords == *lastRecordsProcessed {
		*stuckCount++
		log.Warn("Heartbeat: No progress detected", 
			"testID", testID, 
			"stuckDuration", (*stuckCount)*10, 
			"currentRecords", currentRecords)
	} else {
		if *stuckCount > 0 {
			log.Info("Heartbeat: Progress resumed", "testID", testID, "currentRecords", currentRecords)
		}
		*stuckCount = 0
		*lastRecordsProcessed = currentRecords
	}

	// Send heartbeat message
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"currentPhase":    "Ludicrous Speed Insertion",
		"rowsProcessed":   currentRecords,
		"totalRecords":    totalRecords,
		"elapsed":         int(elapsed.Seconds()),
		"isHeartbeat":     true,
		"isStalled":       *stuckCount >= 3, // 30 seconds
		"message": fmt.Sprintf(
			"Heartbeat: Processing %d/%d records (alive for %ds)",
			currentRecords,
			totalRecords,
			int(elapsed.Seconds()),
		),
	})
}

// updateLoadTestError updates load test with error information for ludicrous speed method
func (c *LudicrousOnlyController) updateLoadTestError(
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

// GetLoadTestByID retrieves a load test by ID
func (c *LudicrousOnlyController) GetLoadTestByID(
	ctx context.Context,
	id string,
) (*LoadTest, error) {
	return c.loadTestRepo.GetByID(ctx, id)
}

// GetAllLoadTests retrieves all load tests
func (c *LudicrousOnlyController) GetAllLoadTests(ctx context.Context) ([]*LoadTest, error) {
	return c.loadTestRepo.GetAll(ctx)
}

// GetPerformanceSummary retrieves performance statistics for ludicrous tests only
func (c *LudicrousOnlyController) GetPerformanceSummary(
	ctx context.Context,
) ([]*PerformanceSummary, error) {
	log := c.log.Function("GetPerformanceSummary")

	allTests, err := c.loadTestRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get load tests: %w", err)
	}

	// Filter for ludicrous method only
	var ludicrousTests []*LoadTest
	for _, test := range allTests {
		if test.Status == "completed" && test.Method == "ludicrous" && test.TotalTime != nil &&
			*test.TotalTime > 0 {
			ludicrousTests = append(ludicrousTests, test)
		}
	}

	if len(ludicrousTests) == 0 {
		log.Info("no completed ludicrous tests found")
		return []*PerformanceSummary{}, nil
	}

	summary := &PerformanceSummary{
		Method:    "ludicrous",
		TestCount: len(ludicrousTests),
	}

	var rowsPerSecondValues []int
	var totalTimeValues []int

	for _, test := range ludicrousTests {
		// Use parseTime + insertTime to exclude CSV generation time
		if test.ParseTime != nil && test.InsertTime != nil && (*test.ParseTime + *test.InsertTime) > 0 {
			parseInsertTime := *test.ParseTime + *test.InsertTime
			rowsPerSec := int(float64(test.Rows) / (float64(parseInsertTime) / 1000.0))
			rowsPerSecondValues = append(rowsPerSecondValues, rowsPerSec)
			totalTimeValues = append(totalTimeValues, parseInsertTime)
		}
	}

	if len(rowsPerSecondValues) > 0 {
		// Simple sorting for percentiles
		for i := 0; i < len(rowsPerSecondValues); i++ {
			for j := i + 1; j < len(rowsPerSecondValues); j++ {
				if rowsPerSecondValues[i] > rowsPerSecondValues[j] {
					rowsPerSecondValues[i], rowsPerSecondValues[j] = rowsPerSecondValues[j], rowsPerSecondValues[i]
				}
			}
		}

		for i := 0; i < len(totalTimeValues); i++ {
			for j := i + 1; j < len(totalTimeValues); j++ {
				if totalTimeValues[i] > totalTimeValues[j] {
					totalTimeValues[i], totalTimeValues[j] = totalTimeValues[j], totalTimeValues[i]
				}
			}
		}

		// Calculate stats
		sumRPS := 0
		sumTime := 0
		for i, rps := range rowsPerSecondValues {
			sumRPS += rps
			sumTime += totalTimeValues[i]
		}
		summary.AvgRowsPerSec = sumRPS / len(rowsPerSecondValues)
		summary.AvgTotalTime = sumTime / len(totalTimeValues)

		summary.MinRowsPerSec = rowsPerSecondValues[0]
		summary.MaxRowsPerSec = rowsPerSecondValues[len(rowsPerSecondValues)-1]
		summary.MinTotalTime = totalTimeValues[0]
		summary.MaxTotalTime = totalTimeValues[len(totalTimeValues)-1]

		p95Index := int(float64(len(rowsPerSecondValues)) * 0.95)
		if p95Index >= len(rowsPerSecondValues) {
			p95Index = len(rowsPerSecondValues) - 1
		}
		summary.P95RowsPerSec = rowsPerSecondValues[p95Index]
	}

	log.Info("ludicrous performance summary calculated", "testCount", len(ludicrousTests))

	return []*PerformanceSummary{summary}, nil
}
