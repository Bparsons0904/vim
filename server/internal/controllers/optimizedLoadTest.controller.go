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
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// testDataPool helps reuse TestData objects to reduce GC pressure.
var testDataPool = sync.Pool{
	New: func() any {
		return new(TestData)
	},
}

type OptimizedLoadTestController struct {
	db           database.DB
	loadTestRepo repositories.LoadTestRepository
	testDataRepo repositories.TestDataRepository
	dateUtils    *utils.DateUtils
	log          logger.Logger
	wsManager    WSManager
}

// InsertMethod defines the insertion approach
type InsertMethod string

const (
	InsertMethodGORM   InsertMethod = "gorm"
	InsertMethodRawSQL InsertMethod = "raw_sql"
)

// WorkerConfig holds configuration for the optimized insertion
type WorkerConfig struct {
	NumWorkers    int
	BatchSize     int
	BufferSize    int
	BatchesPerTxn int          // Number of batches per transaction
	InsertMethod  InsertMethod // Which insertion method to use
}

// BatchData represents a batch of records ready for insertion
type BatchData struct {
	Records  []*TestData
	BatchNum int
}

// Progress represents the current state of the insertion process
type Progress struct {
	RecordsProcessed int
	BatchesProcessed int
	TotalRecords     int
	TotalBatches     int
	StartTime        time.Time
	mu               sync.RWMutex
}

func NewOptimizedLoadTestController(
	db database.DB,
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
	wsManager WSManager,
) *OptimizedLoadTestController {
	return &OptimizedLoadTestController{
		db:           db,
		loadTestRepo: loadTestRepo,
		testDataRepo: testDataRepo,
		dateUtils:    utils.NewDateUtils(),
		log:          logger.New("optimizedLoadTestController"),
		wsManager:    wsManager,
	}
}

// DefaultWorkerConfig provides sensible defaults for the optimized insertion
func DefaultWorkerConfig() *WorkerConfig {
	numWorkers := runtime.NumCPU()
	return &WorkerConfig{
		NumWorkers:    numWorkers,
		BatchSize:     2000,
		BufferSize:    numWorkers * 4, // A larger buffer can help keep workers fed
		BatchesPerTxn: 4,
		InsertMethod:  InsertMethodGORM, // Default to GORM
	}
}

// RawSQLWorkerConfig provides configuration optimized for raw SQL insertion
func RawSQLWorkerConfig() *WorkerConfig {
	numWorkers := runtime.NumCPU()
	return &WorkerConfig{
		NumWorkers:    numWorkers,
		BatchSize:     2000, // Can be larger with raw SQL
		BufferSize:    numWorkers * 4,
		BatchesPerTxn: 4,
		InsertMethod:  InsertMethodRawSQL,
	}
}

// TimingResult contains the timing breakdown for database operations
type TimingResult struct {
	ParseTime  int // milliseconds spent parsing CSV
	InsertTime int // milliseconds spent inserting to database
}

// InsertOptimizedWithProgress performs streaming CSV parsing with concurrent batch processing using GORM
func (c *OptimizedLoadTestController) InsertOptimizedWithProgress(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (TimingResult, error) {
	config := DefaultWorkerConfig()
	config.InsertMethod = InsertMethodGORM
	return c.insertWithConfig(ctx, csvPath, loadTestID, totalRecords, startTime, testID, config)
}

// InsertOptimizedWithRawSQL performs streaming CSV parsing with concurrent batch processing using raw SQL
func (c *OptimizedLoadTestController) InsertOptimizedWithRawSQL(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (TimingResult, error) {
	config := RawSQLWorkerConfig()
	return c.insertWithConfig(ctx, csvPath, loadTestID, totalRecords, startTime, testID, config)
}

// InsertLudicrousSpeed performs streaming CSV parsing with concurrent batch processing using raw SQL with optimized settings
func (c *OptimizedLoadTestController) InsertLudicrousSpeed(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (TimingResult, error) {
	// Ludicrous speed: Aggressive settings with raw SQL and larger batches
	config := &WorkerConfig{
		NumWorkers:    runtime.NumCPU(),
		BatchSize:     2000,                 // Larger batches for ludicrous speed
		BufferSize:    runtime.NumCPU() * 6, // Larger buffer for ludicrous
		BatchesPerTxn: 1,                    // Minimal transaction overhead
		InsertMethod:  InsertMethodRawSQL,
	}
	return c.insertWithConfig(ctx, csvPath, loadTestID, totalRecords, startTime, testID, config)
}

// insertWithConfig performs the actual insertion logic with given configuration
func (c *OptimizedLoadTestController) insertWithConfig(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
	config *WorkerConfig,
) (TimingResult, error) {
	c.log.Info("Starting optimized insertion",
		"totalRecords", totalRecords,
		"workers", config.NumWorkers,
		"batchSize", config.BatchSize,
		"bufferSize", config.BufferSize,
		"insertMethod", config.InsertMethod)

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return TimingResult{}, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	// Initialize progress tracking
	progress := &Progress{
		TotalRecords:     totalRecords,
		TotalBatches:     (totalRecords + config.BatchSize - 1) / config.BatchSize,
		StartTime:        startTime,
		RecordsProcessed: 0,
		BatchesProcessed: 0,
	}

	// Skip index dropping for smaller datasets (under 10 million records)
	var indexDropTime time.Duration
	if totalRecords >= 10000000 {
		indexDropStart := time.Now()
		if err := c.dropIndexesTemporarily(ctx); err != nil {
			c.log.Warn("Failed to drop indexes, continuing anyway", "error", err)
		}
		indexDropTime = time.Since(indexDropStart)
		c.log.Info("Index drop completed", "duration", indexDropTime)
	} else {
		c.log.Info("Skipping index drop for small dataset", "records", totalRecords)
	}

	// Create channels for producer-consumer pattern
	batchChan := make(chan *BatchData, config.BufferSize)
	errorChan := make(chan error, config.NumWorkers)

	// Start progress monitoring goroutine
	progressDone := make(chan bool)
	go c.monitorOptimizedProgress(progress, testID, progressDone)

	// Start worker goroutines
	var workerWG sync.WaitGroup
	workerStartTime := time.Now()
	for i := 0; i < config.NumWorkers; i++ {
		workerWG.Add(1)
		go c.insertWorker(ctx, i, batchChan, errorChan, progress, config, &workerWG)
	}
	c.log.Info(
		"Started workers",
		"count",
		config.NumWorkers,
		"startupTime",
		time.Since(workerStartTime),
	)

	// Start CSV parser (producer)
	parserDone := make(chan error, 1)
	parseStartTime := time.Now()
	go c.parseCSVStreaming(file, loadTestID, batchChan, parserDone, config)

	// Wait for either parser to finish or worker error
	var parseErr error
	select {
	case parseErr = <-parserDone:
		if parseErr != nil {
			close(batchChan)
			progressDone <- true
			return TimingResult{}, fmt.Errorf("CSV parsing failed: %w", parseErr)
		}
	case workerErr := <-errorChan:
		if workerErr != nil {
			close(batchChan)
			progressDone <- true
			return TimingResult{}, fmt.Errorf("worker failed during parsing: %w", workerErr)
		}
	}
	parseTime := time.Since(parseStartTime)
	if parseTime.Seconds() > 0 {
		c.log.Info(
			"CSV parsing completed",
			"duration",
			parseTime,
			"recordsPerSecond",
			float64(totalRecords)/parseTime.Seconds(),
		)
	} else {
		c.log.Info("CSV parsing completed", "duration", parseTime)
	}

	// Close batch channel to signal workers to finish
	close(batchChan)

	// Wait for all workers to complete
	workersStartWait := time.Now()
	workerWG.Wait()
	workersWaitTime := time.Since(workersStartWait)
	c.log.Info("All workers completed", "waitTime", workersWaitTime)

	close(errorChan)

	// Stop progress monitoring
	progressDone <- true

	// Check for worker errors
	for err := range errorChan {
		if err != nil {
			return TimingResult{}, fmt.Errorf("worker error: %w", err)
		}
	}

	// Recreate indexes (only if we dropped them)
	var indexRecreateTime time.Duration
	if totalRecords >= 10000000 {
		indexRecreateStart := time.Now()
		if err := c.recreateIndexes(ctx); err != nil {
			c.log.Warn("Failed to recreate indexes", "error", err)
		}
		indexRecreateTime = time.Since(indexRecreateStart)
		c.log.Info("Index recreation completed", "duration", indexRecreateTime)
	} else {
		c.log.Info("Skipping index recreation for small dataset")
	}

	insertTime := int(time.Since(startTime).Milliseconds())

	// Log final timing breakdown
	progress.mu.RLock()
	finalProcessed := progress.RecordsProcessed
	finalBatches := progress.BatchesProcessed
	progress.mu.RUnlock()

	var rowsPerSecond int
	if insertTime > 0 {
		rowsPerSecond = int(float64(finalProcessed) / (float64(insertTime) / 1000))
	}

	c.log.Info("Optimized insertion timing breakdown",
		"totalTime", time.Since(startTime),
		"indexDropTime", indexDropTime,
		"parseTime", parseTime,
		"workersWaitTime", workersWaitTime,
		"indexRecreateTime", indexRecreateTime,
		"finalProcessed", finalProcessed,
		"finalBatches", finalBatches,
		"rowsPerSecond", rowsPerSecond)

	// Send final progress update
	c.wsManager.SendLoadTestProgress(testID, map[string]any{
		"phase":           "insertion",
		"overallProgress": 100,
		"phaseProgress":   100,
		"currentPhase":    "Database Insertion Complete",
		"rowsProcessed":   progress.RecordsProcessed,
		"rowsPerSecond":   rowsPerSecond,
		"eta":             "Done",
		"message": fmt.Sprintf(
			"Successfully inserted %d records using optimized streaming method",
			progress.RecordsProcessed,
		),
	})

	c.log.Info("optimized insertion completed",
		"totalRecords", progress.RecordsProcessed,
		"totalBatches", progress.BatchesProcessed,
		"workers", config.NumWorkers,
		"batchSize", config.BatchSize,
		"insertTimeMs", insertTime,
		"rowsPerSecond", rowsPerSecond)

	return TimingResult{
		ParseTime:  int(parseTime.Milliseconds()),
		InsertTime: insertTime,
	}, nil
}

// parseCSVStreaming reads CSV file and feeds batches to workers
func (c *OptimizedLoadTestController) parseCSVStreaming(
	file *os.File,
	loadTestID uuid.UUID,
	batchChan chan<- *BatchData,
	done chan<- error,
	config *WorkerConfig,
) {
	defer func() {
		if r := recover(); r != nil {
			done <- fmt.Errorf("parser panic: %v", r)
		}
	}()

	reader := csv.NewReader(file)

	// Read header
	headerStart := time.Now()
	headers, err := reader.Read()
	if err != nil {
		done <- fmt.Errorf("failed to read CSV headers: %w", err)
		return
	}
	c.log.Info("CSV headers read", "headerCount", len(headers), "readTime", time.Since(headerStart))

	// Create a slice of setter functions based on header order to avoid switch statements in the loop
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
			}
		default:
			// Create a no-op setter for columns we don't care about
			setters[i] = func(td *TestData, val string) {}
		}
	}

	currentBatch := make([]*TestData, 0, config.BatchSize)
	batchNum := 0
	rowsRead := 0
	skippedRows := 0
	parseStart := time.Now()

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			done <- fmt.Errorf("failed to read CSV row: %w", err)
			return
		}

		rowsRead++

		// Get a TestData object from the pool and reset it
		testData := testDataPool.Get().(*TestData)
		*testData = TestData{}

		// Initialize base fields
		nowStr := time.Now().Format(time.RFC3339)
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
			batchData := &BatchData{
				Records:  currentBatch,
				BatchNum: batchNum,
			}

			select {
			case batchChan <- batchData:
				// Prep for the next batch
				currentBatch = make([]*TestData, 0, config.BatchSize)
				batchNum++
			case <-time.After(30 * time.Second):
				done <- fmt.Errorf("timeout sending batch to workers")
				return
			}
		}

		// Progress logging removed for cleaner output
	}

	// Send remaining records in final batch
	if len(currentBatch) > 0 {
		batchData := &BatchData{
			Records:  currentBatch,
			BatchNum: batchNum,
		}
		batchChan <- batchData
		batchNum++
	}

	parseTime := time.Since(parseStart)
	var finalRowsPerSec float64
	if parseTime.Seconds() > 0 {
		finalRowsPerSec = float64(rowsRead) / parseTime.Seconds()
	}
	c.log.Info("CSV parsing completed",
		"totalRows", rowsRead,
		"skippedRows", skippedRows,
		"batchesSent", batchNum,
		"parseTime", parseTime,
		"rowsPerSec", finalRowsPerSec)

	done <- nil
}

// insertWorker processes batches from the channel
func (c *OptimizedLoadTestController) insertWorker(
	ctx context.Context,
	workerID int,
	batchChan <-chan *BatchData,
	errorChan chan<- error,
	progress *Progress,
	config *WorkerConfig,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for batch := range batchChan {
		var err error
		switch config.InsertMethod {
		case InsertMethodRawSQL:
			err = c.insertBatchWithRawSQL(ctx, batch.Records, config)
		case InsertMethodGORM:
			fallthrough
		default:
			err = c.insertBatchWithGORM(ctx, batch.Records, config)
		}

		if err != nil {
			errorChan <- fmt.Errorf("worker %d failed to insert batch %d using %s: %w",
				workerID, batch.BatchNum, config.InsertMethod, err)
			// Return failed objects to the pool
			for _, record := range batch.Records {
				testDataPool.Put(record)
			}
			return
		}

		// Return successfully processed objects to the pool
		for _, record := range batch.Records {
			testDataPool.Put(record)
		}

		// Update progress
		progress.mu.Lock()
		progress.RecordsProcessed += len(batch.Records)
		progress.BatchesProcessed++
		progress.mu.Unlock()
	}

	errorChan <- nil // Signal successful completion
}

// insertBatchWithGORM uses GORM's native batch insert capabilities
func (c *OptimizedLoadTestController) insertBatchWithGORM(
	ctx context.Context,
	records []*TestData,
	config *WorkerConfig,
) error {
	if len(records) == 0 {
		return nil
	}

	db := c.db.SQLWithContext(ctx)

	// Use GORM's native batch insert
	err := db.CreateInBatches(records, config.BatchSize).Error
	if err != nil {
		return fmt.Errorf("GORM batch insert failed: %w", err)
	}

	return nil
}

// insertBatchWithRawSQL uses raw SQL prepared statements for maximum performance
func (c *OptimizedLoadTestController) insertBatchWithRawSQL(
	ctx context.Context,
	records []*TestData,
	config *WorkerConfig,
) error {
	if len(records) == 0 {
		return nil
	}

	// Get the underlying *sql.DB connection
	gormDB := c.db.SQLWithContext(ctx)
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying SQL DB: %w", err)
	}

	// Build a single INSERT statement with multiple VALUE clauses
	const baseSQL = `INSERT INTO test_data (
		id, created_at, updated_at, deleted_at, load_test_id, birth_date, start_date, end_date,
		first_name, last_name, email, phone, address_line1, address_line2,
		city, state, zip_code, country, social_security_no, employer, job_title,
		department, salary, insurance_plan_id, insurance_carrier, policy_number,
		group_number, member_id
	) VALUES `

	// Build VALUES clauses and args slice
	var valueClauses []string
	args := make([]interface{}, 0, len(records)*28) // 28 columns per record (including deleted_at)

	for i, record := range records {
		// Build PostgreSQL-style placeholders ($1, $2, $3...)
		placeholders := make([]string, 28)
		for j := 0; j < 28; j++ {
			placeholders[j] = fmt.Sprintf("$%d", i*28+j+1)
		}
		valueClauses = append(valueClauses, "("+strings.Join(placeholders, ", ")+")")

		// Add all values in the same order as the columns
		// Helper function to safely dereference string pointers
		getStringValue := func(ptr *string) interface{} {
			if ptr == nil {
				return nil
			}
			return *ptr
		}

		args = append(args,
			record.ID,
			getStringValue(record.CreatedAt),
			getStringValue(record.UpdatedAt),
			nil, // deleted_at is always NULL for new records
			record.LoadTestID,
			getStringValue(record.BirthDate),
			getStringValue(record.StartDate),
			getStringValue(record.EndDate),
			getStringValue(record.FirstName),
			getStringValue(record.LastName),
			getStringValue(record.Email),
			getStringValue(record.Phone),
			getStringValue(record.AddressLine1),
			getStringValue(record.AddressLine2),
			getStringValue(record.City),
			getStringValue(record.State),
			getStringValue(record.ZipCode),
			getStringValue(record.Country),
			getStringValue(record.SocialSecurityNo),
			getStringValue(record.Employer),
			getStringValue(record.JobTitle),
			getStringValue(record.Department),
			getStringValue(record.Salary),
			getStringValue(record.InsurancePlanID),
			getStringValue(record.InsuranceCarrier),
			getStringValue(record.PolicyNumber),
			getStringValue(record.GroupNumber),
			getStringValue(record.MemberID),
		)
	}

	// Combine base SQL with VALUES clauses
	finalSQL := baseSQL + strings.Join(valueClauses, ", ")

	// Debug: Log the SQL statement structure for troubleshooting
	c.log.Debug("Raw SQL debug info",
		"recordCount", len(records),
		"argCount", len(args),
		"argsPerRecord", len(args)/len(records),
		"valueClauses", len(valueClauses),
		"firstValueClause", valueClauses[0],
		"sqlLength", len(finalSQL))

	// Execute the prepared statement
	_, err = sqlDB.ExecContext(ctx, finalSQL, args...)
	// Performance logging removed for cleaner bulk operation logs
	if err != nil {
		// Log first few args for debugging
		debugArgs := args
		if len(debugArgs) > 10 {
			debugArgs = args[:10]
		}
		return fmt.Errorf(
			"raw SQL batch insert failed (records: %d, args: %d): %w. First few args: %v",
			len(records),
			len(args),
			err,
			debugArgs,
		)
	}

	return nil
}

// insertBatchWithRawSQLAndMultipleConnections uses raw SQL with multiple database connections for maximum parallelism
func (c *OptimizedLoadTestController) insertBatchWithRawSQLAndMultipleConnections(
	ctx context.Context,
	records []*TestData,
	config *WorkerConfig,
	connectionPool *sql.DB, // Pass in a dedicated connection for this worker
) error {
	if len(records) == 0 {
		return nil
	}

	// Build a single INSERT statement with multiple VALUE clauses
	const baseSQL = `INSERT INTO test_data (
		id, created_at, updated_at, deleted_at, load_test_id, birth_date, start_date, end_date,
		first_name, last_name, email, phone, address_line1, address_line2,
		city, state, zip_code, country, social_security_no, employer, job_title,
		department, salary, insurance_plan_id, insurance_carrier, policy_number,
		group_number, member_id
	) VALUES `

	// Build VALUES clauses and args slice
	var valueClauses []string
	args := make([]interface{}, 0, len(records)*28) // 28 columns per record (including deleted_at)

	for i, record := range records {
		// Build PostgreSQL-style placeholders ($1, $2, $3...)
		placeholders := make([]string, 28)
		for j := 0; j < 28; j++ {
			placeholders[j] = fmt.Sprintf("$%d", i*28+j+1)
		}
		valueClauses = append(valueClauses, "("+strings.Join(placeholders, ", ")+")")

		// Add all values in the same order as the columns
		// Helper function to safely dereference string pointers
		getStringValue := func(ptr *string) interface{} {
			if ptr == nil {
				return nil
			}
			return *ptr
		}

		args = append(args,
			record.ID,
			getStringValue(record.CreatedAt),
			getStringValue(record.UpdatedAt),
			nil, // deleted_at is always NULL for new records
			record.LoadTestID,
			getStringValue(record.BirthDate),
			getStringValue(record.StartDate),
			getStringValue(record.EndDate),
			getStringValue(record.FirstName),
			getStringValue(record.LastName),
			getStringValue(record.Email),
			getStringValue(record.Phone),
			getStringValue(record.AddressLine1),
			getStringValue(record.AddressLine2),
			getStringValue(record.City),
			getStringValue(record.State),
			getStringValue(record.ZipCode),
			getStringValue(record.Country),
			getStringValue(record.SocialSecurityNo),
			getStringValue(record.Employer),
			getStringValue(record.JobTitle),
			getStringValue(record.Department),
			getStringValue(record.Salary),
			getStringValue(record.InsurancePlanID),
			getStringValue(record.InsuranceCarrier),
			getStringValue(record.PolicyNumber),
			getStringValue(record.GroupNumber),
			getStringValue(record.MemberID),
		)
	}

	// Combine base SQL with VALUES clauses
	finalSQL := baseSQL + strings.Join(valueClauses, ", ")

	// Execute using the dedicated connection
	_, err := connectionPool.ExecContext(ctx, finalSQL, args...)
	// Performance logging removed for cleaner bulk operation logs
	if err != nil {
		return fmt.Errorf("raw SQL multi-connection batch insert failed: %w", err)
	}

	return nil
}

// monitorOptimizedProgress sends real-time progress updates
func (c *OptimizedLoadTestController) monitorOptimizedProgress(
	progress *Progress,
	testID string,
	done <-chan bool,
) {
	ticker := time.NewTicker(1 * time.Second) // Update every second for real-time feel
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
				"currentPhase":    "Optimized Streaming Insertion",
				"rowsProcessed":   progress.RecordsProcessed,
				"rowsPerSecond":   rowsPerSecond,
				"eta":             eta,
				"message": fmt.Sprintf(
					"Streaming processing: %d/%d records (%d batches)",
					progress.RecordsProcessed,
					progress.TotalRecords,
					progress.BatchesProcessed,
				),
			})
		}
	}
}

// dropIndexesTemporarily drops indexes for better insert performance
func (c *OptimizedLoadTestController) dropIndexesTemporarily(ctx context.Context) error {
	db := c.db.SQLWithContext(ctx).
		Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})

	// Drop indexes that might slow down bulk inserts
	// Note: Be careful with this in production - ensure indexes are recreated
	queries := []string{
		"ALTER TABLE test_data DROP CONSTRAINT IF EXISTS test_data_pkey",
		"DROP INDEX IF EXISTS idx_test_data_load_test_id",
		"DROP INDEX IF EXISTS idx_test_data_deleted_at",
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			// Log warning but continue - some indexes might not exist
			c.log.Warn("Failed to drop index", "query", query, "error", err)
		}
	}

	return nil
}

// recreateIndexes recreates the indexes after bulk insert
func (c *OptimizedLoadTestController) recreateIndexes(ctx context.Context) error {
	db := c.db.SQLWithContext(ctx).
		Session(&gorm.Session{Logger: gormLogger.Default.LogMode(gormLogger.Silent)})

	// Recreate important indexes
	queries := []struct {
		sql         string
		description string
	}{
		{
			sql:         "ALTER TABLE test_data ADD CONSTRAINT test_data_pkey PRIMARY KEY (id)",
			description: "primary key constraint",
		},
		{
			sql:         "CREATE INDEX IF NOT EXISTS idx_test_data_load_test_id ON test_data(load_test_id)",
			description: "load_test_id foreign key index",
		},
	}

	for _, query := range queries {
		if err := db.Exec(query.sql).Error; err != nil {
			// For primary key, check if it already exists and continue
			if strings.Contains(err.Error(), "multiple primary keys") ||
				strings.Contains(err.Error(), "already exists") {
				c.log.Debug("Index already exists, skipping", "description", query.description)
				continue
			}
			c.log.Er("Failed to recreate index", err, "description", query.description)
			return fmt.Errorf("failed to recreate %s: %w", query.description, err)
		}
	}

	c.log.Info("Recreated indexes after bulk insert")
	return nil
}
