package controllers

import (
	"context"
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
	"sync"
	"time"

	"github.com/google/uuid"
)

type OptimizedOnlyController struct {
	loadTestRepo repositories.LoadTestRepository
	testDataRepo repositories.TestDataRepository
	dateUtils    *utils.DateUtils
	log          logger.Logger
	wsManager    WSManager
	db           database.DB
}

func NewOptimizedOnlyController(
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
	db database.DB,
	wsManager WSManager,
	config config.Config,
) *OptimizedOnlyController {
	return &OptimizedOnlyController{
		loadTestRepo: loadTestRepo,
		testDataRepo: testDataRepo,
		dateUtils:    utils.NewDateUtils(),
		log:          logger.New("optimizedOnlyController"),
		wsManager:    wsManager,
		db:           db,
	}
}

// CreateAndRunTest creates a new optimized load test and starts the performance test
func (c *OptimizedOnlyController) CreateAndRunTest(ctx context.Context, req *CreateLoadTestRequest) (*LoadTest, error) {
	log := c.log.Function("CreateAndRunTest")
	
	// Force method to optimized
	const FixedTotalColumns = 25
	const FixedDateColumns = 5
	
	loadTest := &LoadTest{
		// Let GORM handle ID generation automatically
		Rows:          req.Rows,
		Columns:       FixedTotalColumns,
		DateColumns:   FixedDateColumns,
		Method:        "optimized", // Force optimized method
		Status:        "running",
	}

	if err := c.loadTestRepo.Create(ctx, loadTest); err != nil {
		_ = log.Err("failed to create load test", err, "loadTest", loadTest)
		return nil, fmt.Errorf("failed to create load test: %w", err)
	}

	// Process the load test asynchronously
	go c.processLoadTest(ctx, loadTest)

	log.Info("optimized load test created and started", "loadTestId", loadTest.ID)
	return loadTest, nil
}

// processLoadTest handles the optimized load test processing
func (c *OptimizedOnlyController) processLoadTest(ctx context.Context, loadTest *LoadTest) {
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
		"message":          "Starting CSV generation for optimized method...",
	})

	// Step 1: Generate CSV file using shared utility
	csvConfig := utils.CSVGenerationConfig{
		LoadTestID:  loadTest.ID,
		Rows:        loadTest.Rows,
		DateColumns: loadTest.DateColumns,
		FilePrefix:  "optimized_test",
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

	// Step 2: Optimized streaming insertion
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":            "insertion",
		"overallProgress":  25,
		"phaseProgress":    0,
		"currentPhase":     "Optimized Streaming Insertion",
		"rowsProcessed":    0,
		"rowsPerSecond":    0,
		"eta":              "Calculating...",
		"message":          "Starting optimized streaming insertion...",
	})
	
	// Start timing for parse + insert only (excluding CSV generation)
	parseInsertStartTime := time.Now()
	timingResult, err := c.insertOptimizedStreaming(ctx, csvPath, loadTest.ID, loadTest.Rows, parseInsertStartTime, testID)
	if err != nil {
		c.updateLoadTestError(ctx, loadTest, "Optimized insertion failed", err)
		c.wsManager.SendLoadTestError(testID, "Optimized insertion failed: "+err.Error())
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
	
	log.Info("optimized load test completed successfully", 
		"loadTestId", loadTest.ID,
		"totalTime", totalTime)
}

// generateOptimizedCSVFile creates a CSV file optimized for the optimized method
func (c *OptimizedOnlyController) generateOptimizedCSVFile(loadTest *LoadTest) (string, int, error) {
	log := c.log.Function("generateOptimizedCSVFile")
	startTime := time.Now()
	
	// Known date columns for optimized method
	allDateColumns := []string{
		"birth_date",
		"start_date", 
		"end_date",
	}
	
	// Meaningful columns for optimized method
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
	
	log.Info("Optimized CSV generation started", 
		"rows", loadTest.Rows, 
		"columns", loadTest.Columns,
		"selectedDateColumns", len(selectedDateColumns))
	
	// Create temp directory
	tempDir := "/tmp/load_tests"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", 0, fmt.Errorf("failed to create temp directory: %w", err)
	}
	
	csvPath := filepath.Join(tempDir, "optimized_test_"+loadTest.ID.String()+".csv")
	
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
	
	// Shuffle columns
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
	
	// Generate data rows
	for i := 0; i < loadTest.Rows; i++ {
		row := c.generateOptimizedDataRow(allColumns, selectedDateColumnMap, allDateColumnMap, r)
		if err := writer.Write(row); err != nil {
			return "", 0, fmt.Errorf("failed to write row %d: %w", i, err)
		}
		
		if i > 0 && i%10000 == 0 {
			log.Debug("Optimized CSV generation progress", "rowsGenerated", i, "totalRows", loadTest.Rows)
		}
	}
	
	csvGenTime := int(time.Since(startTime).Milliseconds())
	log.Info("Optimized CSV generation completed", 
		"csvPath", csvPath,
		"rows", loadTest.Rows,
		"generationTimeMs", csvGenTime)
	
	return csvPath, csvGenTime, nil
}

// selectRandomDateColumns randomly selects date columns for optimized method
func (c *OptimizedOnlyController) selectRandomDateColumns(allDateColumns []string, count int) []string {
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

// generateOptimizedDataRow creates a data row optimized for the optimized method
func (c *OptimizedOnlyController) generateOptimizedDataRow(headers []string, selectedDateColumnMap, allDateColumnMap map[string]bool, rng *rand.Rand) []string {
	row := make([]string, len(headers))
	
	for i, header := range headers {
		if allDateColumnMap[header] {
			if selectedDateColumnMap[header] {
				row[i] = c.generateOptimizedDateValue(rng)
			} else {
				row[i] = ""
			}
		} else {
			row[i] = c.generateOptimizedMeaningfulColumnValue(header, rng)
		}
	}
	
	return row
}

// generateOptimizedDateValue creates a date value optimized for the optimized method
func (c *OptimizedOnlyController) generateOptimizedDateValue(rng *rand.Rand) string {
	formats := []string{
		"01/02/2006", "1/2/2006", "01/02/06", "1/2/06",
		"01-02-2006", "1-2-2006", "01-02-06", "1-2-06",
		"2006-01-02", "2006/01/02",
	}
	
	year := 2020 + rng.Intn(5)
	month := rng.Intn(12) + 1
	day := rng.Intn(28) + 1
	
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	format := formats[rng.Intn(len(formats))]
	
	return date.Format(format)
}

// Optimized data sets for better performance
var optimizedDataSets = struct {
	FirstNames      []string
	LastNames       []string
	EmailFirsts     []string
	EmailDomains    []string
	StreetNumbers   []int
	StreetNames     []string
	Cities          []string
	States          []string
	Countries       []string
	Companies       []string
	JobTitles       []string
	Departments     []string
	InsuranceCarriers []string
}{
	FirstNames: []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Mary", "James", "Jennifer", "William", "Patricia", "Richard", "Linda", "Charles", "Barbara", "Joseph", "Elizabeth", "Thomas", "Dorothy"},
	LastNames: []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin"},
	EmailFirsts: []string{"john", "jane", "mike", "sarah", "david", "lisa", "bob", "mary", "james", "jen", "alex", "chris", "pat", "sam", "casey", "jordan", "taylor", "morgan", "riley", "drew"},
	EmailDomains: []string{"gmail.com", "yahoo.com", "outlook.com", "company.com", "example.org", "hotmail.com", "aol.com", "icloud.com", "protonmail.com", "mail.com"},
	StreetNumbers: []int{123, 456, 789, 1011, 1234, 5678, 2468, 1357, 9876, 5432, 1111, 2222, 3333, 4444, 5555, 6666, 7777, 8888, 9999, 1024},
	StreetNames: []string{"Main St", "Oak Ave", "First St", "Park Blvd", "Elm St", "Cedar Ave", "Pine Rd", "Maple Dr", "Washington St", "Lincoln Ave", "Jefferson Blvd", "Adams Way", "Madison Ct", "Monroe Dr", "Jackson St", "Franklin Ave", "Roosevelt Rd", "Kennedy Blvd", "Wilson St", "Taylor Ave"},
	Cities: []string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego", "Dallas", "Austin", "Jacksonville", "Fort Worth", "Columbus", "Indianapolis", "Charlotte", "San Francisco", "Seattle", "Denver", "Washington", "Boston"},
	States: []string{"CA", "TX", "NY", "FL", "IL", "PA", "OH", "GA", "NC", "MI", "WA", "AZ", "MA", "VA", "CO", "MD", "MN", "MO", "WI", "OR"},
	Countries: []string{"United States", "Canada", "Mexico", "United Kingdom", "Germany", "France", "Japan", "Australia", "Italy", "Spain", "Netherlands", "Sweden", "Norway", "Denmark", "Finland", "Belgium", "Switzerland", "Austria", "Portugal", "Ireland"},
	Companies: []string{"Acme Corp", "Tech Solutions", "Global Industries", "Metro Systems", "Alpha Enterprises", "Beta LLC", "Gamma Inc", "Delta Group", "Omega Solutions", "Prime Industries", "Elite Systems", "Superior Corp", "Advanced Tech", "Innovative Solutions", "Dynamic Systems", "Strategic Industries", "Premier Group", "Excellence Corp", "Quality Solutions", "Performance Systems"},
	JobTitles: []string{"Software Engineer", "Manager", "Analyst", "Coordinator", "Specialist", "Director", "Associate", "Consultant", "Administrator", "Developer", "Designer", "Architect", "Lead", "Senior", "Junior", "Principal", "Staff", "Team Lead", "Project Manager", "Product Manager"},
	Departments: []string{"Engineering", "Sales", "Marketing", "HR", "Finance", "Operations", "IT", "Customer Service", "Legal", "Research", "Development", "Quality Assurance", "Security", "Procurement", "Training", "Facilities", "Compliance", "Strategy", "Business Development", "Analytics"},
	InsuranceCarriers: []string{"Blue Cross Blue Shield", "Aetna", "Cigna", "UnitedHealthcare", "Humana", "Kaiser Permanente", "Anthem", "Molina Healthcare", "Centene", "WellCare", "Independence Blue Cross", "Highmark", "BCBS of Michigan", "Florida Blue", "Premera Blue Cross", "CareFirst", "Excellus", "HCSC", "Caresource", "BCBS of North Carolina"},
}

// generateOptimizedMeaningfulColumnValue creates appropriate fake data optimized for the optimized method
func (c *OptimizedOnlyController) generateOptimizedMeaningfulColumnValue(columnName string, rng *rand.Rand) string {
	switch columnName {
	case "first_name":
		return optimizedDataSets.FirstNames[rng.Intn(len(optimizedDataSets.FirstNames))]
	case "last_name":
		return optimizedDataSets.LastNames[rng.Intn(len(optimizedDataSets.LastNames))]
	case "email":
		return optimizedDataSets.EmailFirsts[rng.Intn(len(optimizedDataSets.EmailFirsts))] + 
			fmt.Sprint(rng.Intn(999)) + "@" + 
			optimizedDataSets.EmailDomains[rng.Intn(len(optimizedDataSets.EmailDomains))]
	case "phone":
		return fmt.Sprintf("(%03d) %03d-%04d", rng.Intn(800)+200, rng.Intn(800)+200, rng.Intn(10000))
	case "address_line_1":
		return fmt.Sprint(optimizedDataSets.StreetNumbers[rng.Intn(len(optimizedDataSets.StreetNumbers))]) + 
			" " + optimizedDataSets.StreetNames[rng.Intn(len(optimizedDataSets.StreetNames))]
	case "address_line_2":
		if rng.Intn(3) == 0 {
			return "Apt " + fmt.Sprint(rng.Intn(999)+1)
		}
		return ""
	case "city":
		return optimizedDataSets.Cities[rng.Intn(len(optimizedDataSets.Cities))]
	case "state":
		return optimizedDataSets.States[rng.Intn(len(optimizedDataSets.States))]
	case "zip_code":
		return fmt.Sprintf("%05d", rng.Intn(99999))
	case "country":
		return optimizedDataSets.Countries[rng.Intn(len(optimizedDataSets.Countries))]
	case "social_security_no":
		return fmt.Sprintf("***-**-%04d", rng.Intn(10000))
	case "employer":
		return optimizedDataSets.Companies[rng.Intn(len(optimizedDataSets.Companies))]
	case "job_title":
		return optimizedDataSets.JobTitles[rng.Intn(len(optimizedDataSets.JobTitles))]
	case "department":
		return optimizedDataSets.Departments[rng.Intn(len(optimizedDataSets.Departments))]
	case "salary":
		return "$" + fmt.Sprint((rng.Intn(100)+40)*1000)
	case "insurance_plan_id":
		return fmt.Sprintf("PLAN-%03d", rng.Intn(999)+1)
	case "insurance_carrier":
		return optimizedDataSets.InsuranceCarriers[rng.Intn(len(optimizedDataSets.InsuranceCarriers))]
	case "policy_number":
		return fmt.Sprintf("POL-%d-%d", rng.Intn(9999)+1000, rng.Intn(999999)+100000)
	case "group_number":
		return fmt.Sprintf("GRP%03d", rng.Intn(999)+1)
	case "member_id":
		return fmt.Sprintf("MEM%d%03d", rng.Intn(99)+10, rng.Intn(999)+1)
	default:
		return "data_" + fmt.Sprint(rng.Intn(9999))
	}
}

// OptimizedTimingResult contains timing breakdown specific to optimized method
type OptimizedTimingResult struct {
	ParseTime  int
	InsertTime int
}

// insertOptimizedStreaming performs optimized streaming insertion
func (c *OptimizedOnlyController) insertOptimizedStreaming(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (OptimizedTimingResult, error) {
	log := c.log.Function("insertOptimizedStreaming")
	
	// Optimized configuration
	numWorkers := runtime.NumCPU()
	batchSize := 2000
	bufferSize := numWorkers * 4
	
	log.Info("Starting optimized streaming insertion",
		"totalRecords", totalRecords,
		"workers", numWorkers,
		"batchSize", batchSize)

	file, err := os.Open(csvPath)
	if err != nil {
		return OptimizedTimingResult{}, fmt.Errorf("failed to open CSV file: %w", err)
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
	go c.monitorOptimizedProgress(progress, testID, progressDone)

	// Start worker goroutines
	var workerWG sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		workerWG.Add(1)
		go c.optimizedWorker(ctx, i, batchChan, errorChan, progress, batchSize, &workerWG)
	}

	// Start CSV parser
	parserDone := make(chan error, 1)
	parseStartTime := time.Now()
	go c.parseOptimizedCSVStreaming(file, loadTestID, batchChan, parserDone, batchSize)

	// Wait for parser
	var parseErr error
	select {
	case parseErr = <-parserDone:
		if parseErr != nil {
			close(batchChan)
			progressDone <- true
			return OptimizedTimingResult{}, fmt.Errorf("CSV parsing failed: %w", parseErr)
		}
	case workerErr := <-errorChan:
		if workerErr != nil {
			close(batchChan)
			progressDone <- true
			return OptimizedTimingResult{}, fmt.Errorf("worker failed: %w", workerErr)
		}
	}
	parseTime := time.Since(parseStartTime)

	// Close batch channel and wait for workers
	close(batchChan)
	workerWG.Wait()
	close(errorChan)
	progressDone <- true

	// Check for worker errors
	for err := range errorChan {
		if err != nil {
			return OptimizedTimingResult{}, fmt.Errorf("worker error: %w", err)
		}
	}

	totalParseInsertTime := int(time.Since(startTime).Milliseconds())

	// Send final progress update
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 100,
		"phaseProgress":   100,
		"currentPhase":    "Optimized Insertion Complete",
		"rowsProcessed":   progress.RecordsProcessed,
		"rowsPerSecond":   int(float64(progress.RecordsProcessed) / (float64(totalParseInsertTime) / 1000)),
		"eta":             "Done",
		"message":         fmt.Sprintf("Successfully inserted %d records using optimized method", progress.RecordsProcessed),
	})

	// Calculate actual insert time (total - parse time)
	actualInsertTime := totalParseInsertTime - int(parseTime.Milliseconds())

	log.Info("optimized streaming insertion completed",
		"totalRecords", progress.RecordsProcessed,
		"parseTimeMs", parseTime.Milliseconds(),
		"insertTimeMs", actualInsertTime)

	return OptimizedTimingResult{
		ParseTime:  int(parseTime.Milliseconds()),
		InsertTime: actualInsertTime,
	}, nil
}

// parseOptimizedCSVStreaming parses CSV for optimized method
func (c *OptimizedOnlyController) parseOptimizedCSVStreaming(
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

	// Create setters for optimized parsing
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
		case "address_line_2":
			setters[i] = func(td *TestData, val string) { td.AddressLine2 = &val }
		case "city":
			setters[i] = func(td *TestData, val string) { td.City = &val }
		case "state":
			setters[i] = func(td *TestData, val string) { td.State = &val }
		case "zip_code":
			setters[i] = func(td *TestData, val string) { td.ZipCode = &val }
		case "country":
			setters[i] = func(td *TestData, val string) { td.Country = &val }
		case "social_security_no":
			setters[i] = func(td *TestData, val string) { td.SocialSecurityNo = &val }
		case "employer":
			setters[i] = func(td *TestData, val string) { td.Employer = &val }
		case "job_title":
			setters[i] = func(td *TestData, val string) { td.JobTitle = &val }
		case "department":
			setters[i] = func(td *TestData, val string) { td.Department = &val }
		case "salary":
			setters[i] = func(td *TestData, val string) { td.Salary = &val }
		case "insurance_plan_id":
			setters[i] = func(td *TestData, val string) { td.InsurancePlanID = &val }
		case "insurance_carrier":
			setters[i] = func(td *TestData, val string) { td.InsuranceCarrier = &val }
		case "policy_number":
			setters[i] = func(td *TestData, val string) { td.PolicyNumber = &val }
		case "group_number":
			setters[i] = func(td *TestData, val string) { td.GroupNumber = &val }
		case "member_id":
			setters[i] = func(td *TestData, val string) { td.MemberID = &val }
		case "birth_date", "start_date", "end_date":
			h := header
			setters[i] = func(td *TestData, val string) {
				if val != "" {
					validationResult := c.dateUtils.GetValidator().ValidateAndConvert(val)
					if validationResult.IsValid {
						dateStr := validationResult.ParsedTime.Format("2006-01-02")
						switch h {
						case "birth_date":
							td.BirthDate = &dateStr
						case "start_date":
							td.StartDate = &dateStr
						case "end_date":
							td.EndDate = &dateStr
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
			// Let GORM handle ID automatically
			LoadTestID: loadTestID,
		}

		// Use setters to populate struct
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

// optimizedWorker processes batches for optimized method
func (c *OptimizedOnlyController) optimizedWorker(
	ctx context.Context,
	workerID int,
	batchChan <-chan *BatchData,
	errorChan chan<- error,
	progress *Progress,
	batchSize int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for batch := range batchChan {
		// Use GORM batch insert optimized for this method
		db := c.db.SQLWithContext(ctx)
		err := db.CreateInBatches(batch.Records, batchSize).Error
		if err != nil {
			errorChan <- fmt.Errorf("optimized worker %d failed to insert batch %d: %w", workerID, batch.BatchNum, err)
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

// monitorOptimizedProgress sends progress updates for optimized method
func (c *OptimizedOnlyController) monitorOptimizedProgress(
	progress *Progress,
	testID string,
	done <-chan bool,
) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			progress.mu.RLock()

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

			progress.mu.RUnlock()

			c.wsManager.SendLoadTestProgress(testID, map[string]any{
				"phase":           "insertion",
				"overallProgress": 85 + (15 * overallProgress),
				"phaseProgress":   phaseProgress,
				"currentPhase":    "Optimized Streaming Insertion",
				"rowsProcessed":   progress.RecordsProcessed,
				"rowsPerSecond":   rowsPerSecond,
				"eta":             eta,
				"message": fmt.Sprintf(
					"Optimized processing: %d/%d records (%d batches)",
					progress.RecordsProcessed,
					progress.TotalRecords,
					progress.BatchesProcessed,
				),
			})
		}
	}
}

// updateLoadTestError updates load test with error information for optimized method
func (c *OptimizedOnlyController) updateLoadTestError(ctx context.Context, loadTest *LoadTest, message string, err error) {
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
func (c *OptimizedOnlyController) GetLoadTestByID(ctx context.Context, id string) (*LoadTest, error) {
	return c.loadTestRepo.GetByID(ctx, id)
}

// GetAllLoadTests retrieves all load tests
func (c *OptimizedOnlyController) GetAllLoadTests(ctx context.Context) ([]*LoadTest, error) {
	return c.loadTestRepo.GetAll(ctx)
}

// GetPerformanceSummary retrieves performance statistics for optimized tests only
func (c *OptimizedOnlyController) GetPerformanceSummary(ctx context.Context) ([]*PerformanceSummary, error) {
	log := c.log.Function("GetPerformanceSummary")
	
	allTests, err := c.loadTestRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get load tests: %w", err)
	}
	
	// Filter for optimized method only
	var optimizedTests []*LoadTest
	for _, test := range allTests {
		if test.Status == "completed" && test.Method == "optimized" && test.TotalTime != nil && *test.TotalTime > 0 {
			optimizedTests = append(optimizedTests, test)
		}
	}
	
	if len(optimizedTests) == 0 {
		log.Info("no completed optimized tests found")
		return []*PerformanceSummary{}, nil
	}
	
	summary := &PerformanceSummary{
		Method:    "optimized",
		TestCount: len(optimizedTests),
	}
	
	var rowsPerSecondValues []int
	var totalTimeValues []int
	
	for _, test := range optimizedTests {
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
	
	log.Info("optimized performance summary calculated", "testCount", len(optimizedTests))
	
	return []*PerformanceSummary{summary}, nil
}