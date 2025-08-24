package controllers

import (
	"context"
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
	"sync"
	"time"

	"github.com/google/uuid"
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

// WorkerConfig holds configuration for the optimized insertion
type WorkerConfig struct {
	NumWorkers    int
	BatchSize     int
	BufferSize    int
	BatchesPerTxn int // Number of batches per transaction
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
		BufferSize:    numWorkers * 2, // A larger buffer can help keep workers fed
		BatchesPerTxn: 1,
	}
}

// InsertOptimizedWithProgress performs streaming CSV parsing with concurrent batch processing
func (c *OptimizedLoadTestController) InsertOptimizedWithProgress(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
	startTime time.Time,
	testID string,
) (int, error) {
	config := DefaultWorkerConfig()

	c.log.Info("Starting optimized insertion",
		"totalRecords", totalRecords,
		"workers", config.NumWorkers,
		"batchSize", config.BatchSize,
		"bufferSize", config.BufferSize)

	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %w", err)
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

	// Skip index dropping for smaller datasets (under 100k records)
	var indexDropTime time.Duration
	if totalRecords >= 100000 {
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
	c.log.Info("Started workers", "count", config.NumWorkers, "startupTime", time.Since(workerStartTime))

	// Start CSV parser (producer)
	parserDone := make(chan error, 1)
	parseStartTime := time.Now()
	go c.parseCSVStreaming(file, loadTestID, batchChan, parserDone, config)

	// Wait for parser to finish
	if err := <-parserDone; err != nil {
		close(batchChan)
		progressDone <- true
		return 0, fmt.Errorf("CSV parsing failed: %w", err)
	}
	parseTime := time.Since(parseStartTime)
	if parseTime.Seconds() > 0 {
		c.log.Info("CSV parsing completed", "duration", parseTime, "recordsPerSecond", float64(totalRecords)/parseTime.Seconds())
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
			return 0, fmt.Errorf("worker error: %w", err)
		}
	}

	// Recreate indexes (only if we dropped them)
	var indexRecreateTime time.Duration
	if totalRecords >= 100000 {
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
		"phase":            "insertion",
		"overallProgress":  100,
		"phaseProgress":    100,
		"currentPhase":     "Database Insertion Complete",
		"rowsProcessed":    progress.RecordsProcessed,
		"rowsPerSecond":    rowsPerSecond,
		"eta":              "Done",
		"message":          fmt.Sprintf("Successfully inserted %d records using optimized streaming method", progress.RecordsProcessed),
	})

	c.log.Info("optimized insertion completed",
		"totalRecords", progress.RecordsProcessed,
		"totalBatches", progress.BatchesProcessed,
		"workers", config.NumWorkers,
		"batchSize", config.BatchSize,
		"insertTimeMs", insertTime,
		"rowsPerSecond", rowsPerSecond)

	return insertTime, nil
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

	var currentBatch = make([]*TestData, 0, config.BatchSize)
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

		// Log progress every 1000 rows
		if rowsRead%1000 == 0 {
			elapsed := time.Since(parseStart)
			var rowsPerSec float64
			if elapsed.Seconds() > 0 {
				rowsPerSec = float64(rowsRead) / elapsed.Seconds()
			}
			c.log.Info("Parsing progress", "rowsRead", rowsRead, "batchesSent", batchNum, "rowsPerSec", int(rowsPerSec), "skippedRows", skippedRows)
		}
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

	workerStart := time.Now()
	batchesProcessed := 0
	totalProcessTime := time.Duration(0)

	for batch := range batchChan {
		batchStart := time.Now()

		if err := c.insertBatchWithGORM(ctx, batch.Records, config); err != nil {
			errorChan <- fmt.Errorf("worker %d failed to insert batch %d: %w", workerID, batch.BatchNum, err)
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

		batchTime := time.Since(batchStart)
		totalProcessTime += batchTime
		batchesProcessed++

		// Update progress
		progress.mu.Lock()
		progress.RecordsProcessed += len(batch.Records)
		progress.BatchesProcessed++
		progress.mu.Unlock()

		// Log slow batches
		if batchTime > 100*time.Millisecond {
			c.log.Info("Slow batch detected",
				"workerID", workerID,
				"batchNum", batch.BatchNum,
				"batchSize", len(batch.Records),
				"batchTime", batchTime)
		}
	}

	workerTotal := time.Since(workerStart)
	var avgBatchTime time.Duration
	if batchesProcessed > 0 {
		avgBatchTime = totalProcessTime / time.Duration(batchesProcessed)
	}

	c.log.Info("Worker completed",
		"workerID", workerID,
		"batchesProcessed", batchesProcessed,
		"totalTime", workerTotal,
		"totalProcessTime", totalProcessTime,
		"avgBatchTime", avgBatchTime,
		"idleTime", workerTotal-totalProcessTime)

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

	// Use GORM's native batch insert - much faster than raw SQL + transaction
	execStart := time.Now()
	err := db.CreateInBatches(records, config.BatchSize).Error
	execTime := time.Since(execStart)

	// Log slow operations
	if execTime > 50*time.Millisecond {
		c.log.Info("Slow GORM batch operation",
			"recordCount", len(records),
			"execTime", execTime,
			"batchSize", config.BatchSize)
	}

	if err != nil {
		return fmt.Errorf("GORM batch insert failed: %w", err)
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
				"phase":            "insertion",
				"overallProgress":  85 + (15 * overallProgress), // Scale to 85-100%
				"phaseProgress":    phaseProgress,
				"currentPhase":     "Optimized Streaming Insertion",
				"rowsProcessed":    progress.RecordsProcessed,
				"rowsPerSecond":    rowsPerSecond,
				"eta":              eta,
				"message":          fmt.Sprintf("Streaming processing: %d/%d records (%d batches)", progress.RecordsProcessed, progress.TotalRecords, progress.BatchesProcessed),
			})
		}
	}
}

// dropIndexesTemporarily drops indexes for better insert performance
func (c *OptimizedLoadTestController) dropIndexesTemporarily(ctx context.Context) error {
	db := c.db.SQLWithContext(ctx)
	
	// Drop indexes that might slow down bulk inserts
	// Note: Be careful with this in production - ensure indexes are recreated
	queries := []string{
		"DROP INDEX IF EXISTS idx_test_data_load_test_id",
		"DROP INDEX IF EXISTS idx_test_data_email",
		"DROP INDEX IF EXISTS idx_test_data_phone",
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
	db := c.db.SQLWithContext(ctx)
	
	// Recreate important indexes
	queries := []string{
		"CREATE INDEX IF NOT EXISTS idx_test_data_load_test_id ON test_data(load_test_id)",
		"CREATE INDEX IF NOT EXISTS idx_test_data_email ON test_data(email)",
		"CREATE INDEX IF NOT EXISTS idx_test_data_phone ON test_data(phone)",
	}

	for _, query := range queries {
		if err := db.Exec(query).Error; err != nil {
			return fmt.Errorf("failed to recreate index: %w", err)
		}
	}

	c.log.Info("Recreated indexes after bulk insert")
	return nil
}
