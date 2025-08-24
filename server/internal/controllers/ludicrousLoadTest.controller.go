package controllers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
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

// ludicrousTestDataPool helps reuse TestData objects to reduce GC pressure.
var ludicrousTestDataPool = sync.Pool{
	New: func() any {
		return new(TestData)
	},
}

type LudicrousLoadTestController struct {
	db           database.DB
	loadTestRepo repositories.LoadTestRepository
	testDataRepo repositories.TestDataRepository
	dateUtils    *utils.DateUtils
	log          logger.Logger
	wsManager    WSManager
}

// LudicrousConfig holds configuration for ludicrous speed insertion
type LudicrousConfig struct {
	NumWorkers int
	BatchSize  int
	BufferSize int
}

// LudicrousBatchData represents a batch of records ready for insertion
type LudicrousBatchData struct {
	Records  []*TestData
	BatchNum int
}

// LudicrousProgress represents the current state of the insertion process
type LudicrousProgress struct {
	RecordsProcessed int
	BatchesProcessed int
	TotalRecords     int
	TotalBatches     int
	StartTime        time.Time
	mu               sync.RWMutex
}

func NewLudicrousLoadTestController(
	db database.DB,
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
	wsManager WSManager,
) *LudicrousLoadTestController {
	return &LudicrousLoadTestController{
		db:           db,
		loadTestRepo: loadTestRepo,
		testDataRepo: testDataRepo,
		dateUtils:    utils.NewDateUtils(),
		log:          logger.New("ludicrousLoadTestController"),
		wsManager:    wsManager,
	}
}

// LudicrousWorkerConfig provides configuration optimized for ludicrous speed
func LudicrousWorkerConfig() *LudicrousConfig {
	numWorkers := runtime.NumCPU()
	return &LudicrousConfig{
		NumWorkers: numWorkers,
		BatchSize:  3000, // Larger batches for raw SQL
		BufferSize: numWorkers * 4,
	}
}

// InsertLudicrousSpeed performs streaming CSV parsing with concurrent batch processing using optimized raw SQL
func (c *LudicrousLoadTestController) InsertLudicrousSpeed(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (int, error) {
	config := LudicrousWorkerConfig()

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Initialize progress tracking
	progress := &LudicrousProgress{
		TotalRecords:     totalRecords,
		TotalBatches:     (totalRecords + config.BatchSize - 1) / config.BatchSize,
		StartTime:        startTime,
		RecordsProcessed: 0,
		BatchesProcessed: 0,
	}

	// Drop indexes for datasets >= 100k records
	if totalRecords >= 100000 {
		c.dropIndexesTemporarily(ctx)
	}

	// Create channels for producer-consumer pattern
	batchChan := make(chan *LudicrousBatchData, config.BufferSize)
	errorChan := make(chan error, config.NumWorkers)

	// Start progress monitoring goroutine
	progressDone := make(chan bool)
	go c.monitorLudicrousProgress(progress, testID, progressDone)

	// Start worker goroutines
	var workerWG sync.WaitGroup
	for i := 0; i < config.NumWorkers; i++ {
		workerWG.Add(1)
		go c.ludicrousInsertWorker(ctx, i, batchChan, errorChan, progress, config, &workerWG)
	}

	// Start CSV parser (producer)
	parserDone := make(chan error, 1)
	go c.parseCSVLudicrous(file, loadTestID, batchChan, parserDone, config)

	// Wait for either parser to finish or worker error
	select {
	case err := <-parserDone:
		if err != nil {
			close(batchChan)
			progressDone <- true
			return 0, fmt.Errorf("CSV parsing failed: %w", err)
		}
	case err := <-errorChan:
		if err != nil {
			close(batchChan)
			progressDone <- true
			return 0, fmt.Errorf("worker failed during parsing: %w", err)
		}
	}

	// Close batch channel to signal workers to finish
	close(batchChan)

	// Wait for all workers to complete
	workerWG.Wait()
	close(errorChan)

	// Stop progress monitoring
	progressDone <- true

	// Check for any remaining worker errors
	for err := range errorChan {
		if err != nil {
			return 0, fmt.Errorf("worker error: %w", err)
		}
	}

	// Recreate indexes (only if we dropped them)
	if totalRecords >= 100000 {
		c.recreateIndexes(ctx)
	}

	insertTime := int(time.Since(startTime).Milliseconds())

	// Send final progress update
	progress.mu.RLock()
	finalProcessed := progress.RecordsProcessed
	progress.mu.RUnlock()

	var rowsPerSecond int
	if insertTime > 0 {
		rowsPerSecond = int(float64(finalProcessed) / (float64(insertTime) / 1000))
	}

	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 100,
		"phaseProgress":   100,
		"currentPhase":    "Ludicrous Speed Complete",
		"rowsProcessed":   finalProcessed,
		"rowsPerSecond":   rowsPerSecond,
		"eta":             "Done",
		"message": fmt.Sprintf(
			"Successfully inserted %d records using ludicrous speed method",
			finalProcessed,
		),
	})

	return insertTime, nil
}

// parseCSVLudicrous reads CSV file and feeds batches to workers with minimal overhead
func (c *LudicrousLoadTestController) parseCSVLudicrous(
	file *os.File,
	loadTestID uuid.UUID,
	batchChan chan<- *LudicrousBatchData,
	done chan<- error,
	config *LudicrousConfig,
) {
	defer func() {
		if r := recover(); r != nil {
			done <- fmt.Errorf("parser panic: %v", r)
		}
	}()

	reader := csv.NewReader(file)

	// Read header
	headers, err := reader.Read()
	if err != nil {
		done <- fmt.Errorf("failed to read CSV headers: %w", err)
		return
	}

	// Create setter functions based on header order
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
			// Need to capture header for the closure
			h := header
			setters[i] = func(td *TestData, val string) {
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
				// Note: Invalid dates are not stored (matches Optimized behavior)
			}
		default:
			setters[i] = func(td *TestData, val string) {}
		}
	}

	currentBatch := make([]*TestData, 0, config.BatchSize)
	batchNum := 0
	nowStr := time.Now().Format(time.RFC3339)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			done <- fmt.Errorf("failed to read CSV row: %w", err)
			return
		}

		// Get a TestData object from the pool and reset it
		testData := ludicrousTestDataPool.Get().(*TestData)
		*testData = TestData{}

		// Initialize base fields
		testData.ID = uuid.New()
		testData.CreatedAt = &nowStr
		testData.UpdatedAt = &nowStr
		testData.LoadTestID = loadTestID

		// Use the setters to populate the struct
		for i, value := range row {
			if value != "" {
				setters[i](testData, value)
			}
		}

		currentBatch = append(currentBatch, testData)

		// Send batch when it's full
		if len(currentBatch) >= config.BatchSize {
			batchData := &LudicrousBatchData{
				Records:  currentBatch,
				BatchNum: batchNum,
			}

			select {
			case batchChan <- batchData:
				currentBatch = make([]*TestData, 0, config.BatchSize)
				batchNum++
			case <-time.After(30 * time.Second):
				done <- fmt.Errorf("timeout sending batch to workers")
				return
			}
		}
	}

	// Send remaining records in final batch
	if len(currentBatch) > 0 {
		batchData := &LudicrousBatchData{
			Records:  currentBatch,
			BatchNum: batchNum,
		}
		batchChan <- batchData
	}

	done <- nil
}

// ludicrousInsertWorker processes batches from the channel using raw SQL
func (c *LudicrousLoadTestController) ludicrousInsertWorker(
	ctx context.Context,
	workerID int,
	batchChan <-chan *LudicrousBatchData,
	errorChan chan<- error,
	progress *LudicrousProgress,
	config *LudicrousConfig,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Get the underlying *sql.DB connection
	gormDB := c.db.SQLWithContext(ctx)
	sqlDB, err := gormDB.DB()
	if err != nil {
		errorChan <- fmt.Errorf("worker %d failed to get SQL DB: %w", workerID, err)
		return
	}

	for batch := range batchChan {
		if err := c.insertBatchLudicrous(ctx, batch.Records, sqlDB); err != nil {
			errorChan <- fmt.Errorf("worker %d ludicrous insert failed on batch %d: %w", workerID, batch.BatchNum, err)
			// Return failed objects to the pool
			for _, record := range batch.Records {
				ludicrousTestDataPool.Put(record)
			}
			return
		}

		// Return successfully processed objects to the pool
		for _, record := range batch.Records {
			ludicrousTestDataPool.Put(record)
		}

		// Update progress
		progress.mu.Lock()
		progress.RecordsProcessed += len(batch.Records)
		progress.BatchesProcessed++
		progress.mu.Unlock()
	}

	errorChan <- nil // Signal successful completion
}

// insertBatchLudicrous uses raw SQL with minimal overhead for maximum speed
func (c *LudicrousLoadTestController) insertBatchLudicrous(
	ctx context.Context,
	records []*TestData,
	sqlDB *sql.DB,
) error {
	if len(records) == 0 {
		return nil
	}

	// Build a single INSERT statement with multiple VALUE clauses
	const baseSQL = `INSERT INTO test_data (
		id, created_at, updated_at, load_test_id, birth_date, start_date, end_date,
		first_name, last_name, email, phone, address_line_1, address_line_2,
		city, state, zip_code, country, social_security_no, employer, job_title,
		department, salary, insurance_plan_id, insurance_carrier, policy_number,
		group_number, member_id
	) VALUES `

	// Pre-allocate slices for better performance
	valueClauses := make([]string, 0, len(records))
	args := make([]interface{}, 0, len(records)*27) // 27 columns per record

	valueClause := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	
	for _, record := range records {
		valueClauses = append(valueClauses, valueClause)
		
		// Add all values in the same order as the columns
		args = append(args,
			record.ID,
			record.CreatedAt,
			record.UpdatedAt,
			record.LoadTestID,
			record.BirthDate,
			record.StartDate,
			record.EndDate,
			record.FirstName,
			record.LastName,
			record.Email,
			record.Phone,
			record.AddressLine1,
			record.AddressLine2,
			record.City,
			record.State,
			record.ZipCode,
			record.Country,
			record.SocialSecurityNo,
			record.Employer,
			record.JobTitle,
			record.Department,
			record.Salary,
			record.InsurancePlanID,
			record.InsuranceCarrier,
			record.PolicyNumber,
			record.GroupNumber,
			record.MemberID,
		)
	}

	// Combine base SQL with VALUES clauses
	finalSQL := baseSQL + strings.Join(valueClauses, ", ")

	// Execute the prepared statement
	_, err := sqlDB.ExecContext(ctx, finalSQL, args...)
	if err != nil {
		// Log first few args for debugging
		debugArgs := args
		if len(debugArgs) > 10 {
			debugArgs = args[:10]
		}
		return fmt.Errorf("SQL execution failed (records: %d, args: %d): %w. First few args: %v", 
			len(records), len(args), err, debugArgs)
	}
	return nil
}

// monitorLudicrousProgress sends real-time progress updates
func (c *LudicrousLoadTestController) monitorLudicrousProgress(
	progress *LudicrousProgress,
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

			// Calculate ETA
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

			// Calculate rows per second
			rowsPerSecond := 0
			if elapsed.Seconds() > 0 {
				rowsPerSecond = int(float64(progress.RecordsProcessed) / elapsed.Seconds())
			}

			progress.mu.RUnlock()

			c.wsManager.SendLoadTestProgress(testID, map[string]any{
				"phase":           "insertion",
				"overallProgress": 85 + (15 * overallProgress), // Scale to 85-100%
				"phaseProgress":   phaseProgress,
				"currentPhase":    "Ludicrous Speed Insertion",
				"rowsProcessed":   progress.RecordsProcessed,
				"rowsPerSecond":   rowsPerSecond,
				"eta":             eta,
				"message": fmt.Sprintf(
					"Ludicrous speed: %d/%d records (%d batches)",
					progress.RecordsProcessed,
					progress.TotalRecords,
					progress.BatchesProcessed,
				),
			})
		}
	}
}

// dropIndexesTemporarily drops indexes for better insert performance
func (c *LudicrousLoadTestController) dropIndexesTemporarily(ctx context.Context) error {
	db := c.db.SQLWithContext(ctx)

	queries := []string{
		"DROP INDEX IF EXISTS idx_test_data_load_test_id",
		"DROP INDEX IF EXISTS idx_test_data_email",
		"DROP INDEX IF EXISTS idx_test_data_phone",
	}

	for _, query := range queries {
		db.Exec(query)
	}

	return nil
}

// recreateIndexes recreates the indexes after bulk insert
func (c *LudicrousLoadTestController) recreateIndexes(ctx context.Context) error {
	db := c.db.SQLWithContext(ctx)

	queries := []string{
		"CREATE INDEX IF NOT EXISTS idx_test_data_load_test_id ON test_data(load_test_id)",
	}

	for _, query := range queries {
		db.Exec(query)
	}

	return nil
}