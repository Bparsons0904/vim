package controllers

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/repositories"
	"server/internal/utils"
	"sync"
	"time"

	"github.com/google/uuid"
)

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
	return &WorkerConfig{
		NumWorkers:    6,    // 6 concurrent workers
		BatchSize:     2000, // Records per batch
		BufferSize:    3,    // Buffer size for batch channel
		BatchesPerTxn: 1,    // Start with 1 batch per transaction, can be tuned
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
	c.log.Info("CSV parsing completed", "duration", parseTime, "recordsPerSecond", float64(totalRecords)/parseTime.Seconds())

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
	
	rowsPerSecond := int(float64(finalProcessed) / (float64(insertTime) / 1000))
	
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

	var currentBatch []*TestData
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

		// Convert CSV row to TestData
		testData, err := c.csvRowToTestData(row, headers, loadTestID)
		if err != nil {
			// Skip invalid rows but continue processing
			skippedRows++
			continue
		}

		currentBatch = append(currentBatch, testData)

		// Send batch when it's full
		if len(currentBatch) >= config.BatchSize {
			batchData := &BatchData{
				Records:  currentBatch,
				BatchNum: batchNum,
			}
			
			sendStart := time.Now()
			select {
			case batchChan <- batchData:
				sendTime := time.Since(sendStart)
				if sendTime > 10*time.Millisecond {
					c.log.Info("Slow batch send", "batchNum", batchNum, "sendTime", sendTime)
				}
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
			rowsPerSec := float64(rowsRead) / elapsed.Seconds()
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
	c.log.Info("CSV parsing completed", 
		"totalRows", rowsRead,
		"skippedRows", skippedRows,
		"batchesSent", batchNum,
		"parseTime", parseTime,
		"rowsPerSec", float64(rowsRead)/parseTime.Seconds())

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
			return
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


// csvRowToTestData converts a CSV row to TestData struct
func (c *OptimizedLoadTestController) csvRowToTestData(row, headers []string, loadTestID uuid.UUID) (*TestData, error) {
	if len(row) != len(headers) {
		return nil, fmt.Errorf("row length %d doesn't match headers length %d", len(row), len(headers))
	}

	testData := &TestData{
		BaseUUIDModel: BaseUUIDModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		LoadTestID: loadTestID,
	}

	// Map CSV columns to struct fields
	for i, header := range headers {
		value := row[i]
		if value == "" {
			continue // Skip empty values
		}

		switch header {
		case "first_name":
			testData.FirstName = &value
		case "last_name":
			testData.LastName = &value
		case "email":
			testData.Email = &value
		case "phone":
			testData.Phone = &value
		case "address_line_1":
			testData.AddressLine1 = &value
		case "address_line_2":
			testData.AddressLine2 = &value
		case "city":
			testData.City = &value
		case "state":
			testData.State = &value
		case "zip_code":
			testData.ZipCode = &value
		case "country":
			testData.Country = &value
		case "social_security_no":
			testData.SocialSecurityNo = &value
		case "employer":
			testData.Employer = &value
		case "job_title":
			testData.JobTitle = &value
		case "department":
			testData.Department = &value
		case "salary":
			testData.Salary = &value
		case "insurance_plan_id":
			testData.InsurancePlanID = &value
		case "insurance_carrier":
			testData.InsuranceCarrier = &value
		case "policy_number":
			testData.PolicyNumber = &value
		case "group_number":
			testData.GroupNumber = &value
		case "member_id":
			testData.MemberID = &value
		case "birth_date", "start_date", "end_date":
			// Validate and parse date fields
			validationResult := c.dateUtils.GetValidator().ValidateAndConvert(value)
			if validationResult.IsValid {
				dateStr := validationResult.ParsedTime.Format("2006-01-02")
				switch header {
				case "birth_date":
					testData.BirthDate = &dateStr
				case "start_date":
					testData.StartDate = &dateStr
				case "end_date":
					testData.EndDate = &dateStr
				}
			}
		}
	}

	return testData, nil
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