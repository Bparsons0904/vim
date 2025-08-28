package controllers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"runtime"
	"server/config"
	"server/internal/logger"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// PlaidTimingResult holds the timing metrics for the operation.
type PlaidTimingResult struct {
	ParseTime  int
	InsertTime int
	TotalTime  int
}

// PlaidController manages the database operations.
type PlaidController struct {
	log       logger.Logger
	config    config.Config
	wsManager WSManager
	db        *sql.DB
}

// NewPlaidController creates a new PlaidController instance.
func NewPlaidController(config config.Config, wsManager WSManager) (*PlaidController, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseUser,
		config.DatabasePassword,
		config.DatabaseName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open raw sql with lib/pq: %w", err)
	}

	// Set connection pool settings. It's crucial to have enough connections
	// to support the concurrent workers.
	db.SetMaxOpenConns(2 * runtime.NumCPU())
	db.SetMaxIdleConns(2 * runtime.NumCPU())
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PlaidController{
		log:       logger.New("plaidController"),
		config:    config,
		wsManager: wsManager,
		db:        db,
	}, nil
}

// Close closes the database connection.
func (c *PlaidController) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// RunPlaidCopy initiates the concurrent copy operation from a CSV file.
func (c *PlaidController) RunPlaidCopy(
	ctx context.Context,
	csvPath string,
	loadTestID uuid.UUID,
	totalRecords int,
) (PlaidTimingResult, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	timingResult, err := c.executeConcurrentStreamingCopy(
		ctx,
		file,
		loadTestID,
		totalRecords,
		loadTestID.String(),
	)
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("concurrent streaming COPY failed: %w", err)
	}

	return timingResult, nil
}

// executeConcurrentStreamingCopy implements the producer-consumer pattern for bulk insertion.
// A single producer goroutine reads and parses the CSV, while multiple consumer goroutines
// handle the database COPY operations in parallel. This is a much better way to
// utilize multiple CPU cores and database connections.
func (c *PlaidController) executeConcurrentStreamingCopy(
	ctx context.Context,
	file *os.File,
	loadTestID uuid.UUID,
	totalRecords int,
	testID string,
) (PlaidTimingResult, error) {
	// Pre-defined columns for the database table.
	dbColumns := []string{
		"load_test_id", "birth_date", "start_date", "end_date",
		"first_name", "last_name", "email", "phone", "address_line1", "address_line2",
		"city", "state", "zip_code", "country", "social_security_no",
		"employer", "job_title", "department", "salary",
		"insurance_plan_id", "insurance_carrier", "policy_number",
		"group_number", "member_id",
	}

	// Producer-consumer pattern setup
	var wg sync.WaitGroup
	// Channel to send parsed rows to the workers. Using a buffered channel prevents the producer from blocking.
	recordsChan := make(chan []interface{}, 1000)
	errChan := make(chan error, 1)

	// Timing variables - use wall-clock time instead of summed worker time
	parseStartTime := time.Now()
	startTime := time.Now()
	lastUpdateTime := time.Now()

	// ------------------
	// Consumer (Worker) Goroutines
	// ------------------
	numWorkers := runtime.NumCPU()
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			workerDB, err := c.db.Conn(ctx)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("worker %d failed to get DB connection: %w", workerID, err):
				default:
				}
				return
			}
			defer workerDB.Close()

			tx, err := workerDB.BeginTx(ctx, nil)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("worker %d failed to begin transaction: %w", workerID, err):
				default:
				}
				return
			}
			defer tx.Rollback() // Rollback on return unless commit is successful

			// Using a dedicated statement for each worker's transaction.
			stmt, err := tx.PrepareContext(ctx, pq.CopyIn("test_data", dbColumns...))
			if err != nil {
				select {
				case errChan <- fmt.Errorf("worker %d failed to prepare COPY statement: %w", workerID, err):
				default:
				}
				return
			}
			defer stmt.Close()

			for record := range recordsChan {
				_, err := stmt.Exec(record...)
				if err != nil {
					select {
					case errChan <- fmt.Errorf("worker %d failed to execute COPY: %w", workerID, err):
					default:
					}
					return
				}
			}

			// Finalize the COPY operation.
			_, err = stmt.Exec()
			if err != nil {
				select {
				case errChan <- fmt.Errorf("worker %d failed to finalize COPY operation: %w", workerID, err):
				default:
				}
				return
			}

			if err := tx.Commit(); err != nil {
				select {
				case errChan <- fmt.Errorf("worker %d failed to commit transaction: %w", workerID, err):
				default:
				}
				return
			}
		}(i)
	}

	// ------------------
	// Producer (Main) Goroutine
	// ------------------
	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		close(recordsChan)
		return PlaidTimingResult{}, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	// Column mapping: Map the source CSV header names to their index in the target DB columns slice.
	// This is done once to avoid repeated map lookups inside the loop.
	headerIndexMap := make(map[string]int, len(headers))
	for i, h := range headers {
		headerIndexMap[h] = i
	}

	// Create a fast mapping from dbColumns to the CSV record's index
	columnMapping := make([]int, len(dbColumns))
	for i, col := range dbColumns {
		if idx, ok := headerIndexMap[col]; ok {
			columnMapping[i] = idx
		} else {
			// This handles columns that might not exist in the CSV, e.g., 'load_test_id'.
			columnMapping[i] = -1
		}
	}

	rowCount := 0
	var parseEndTime time.Time
	go func() {
		defer close(recordsChan)
		for {
			csvRecord, err := reader.Read()
			if err == io.EOF {
				parseEndTime = time.Now()
				break
			}
			if err != nil {
				select {
				case errChan <- fmt.Errorf("failed to read CSV row: %w", err):
				default:
				}
				return
			}

			// Create the target record with a single pass, using the pre-calculated mapping
			record := make([]interface{}, len(dbColumns))
			for i, csvIndex := range columnMapping {
				switch i {
				case 0: // "load_test_id"
					record[i] = loadTestID.String()
				case 1, 2, 3: // "birth_date", "start_date", "end_date"
					if csvIndex != -1 && csvIndex < len(csvRecord) {
						record[i] = c.getNormalizedDateOrNull(csvRecord[csvIndex])
					} else {
						record[i] = nil
					}
				default: // All other columns
					if csvIndex != -1 && csvIndex < len(csvRecord) {
						record[i] = c.getValueOrNull(csvRecord[csvIndex])
					} else {
						record[i] = nil
					}
				}
			}

			recordsChan <- record
			rowCount++

			if time.Since(lastUpdateTime) > 2*time.Second {
				elapsed := time.Since(startTime)
				rowsPerSecond := int(float64(rowCount) / elapsed.Seconds())
				progress := float64(rowCount) / float64(totalRecords) * 100
				
				// Calculate ETA
				var eta string
				if rowCount > 0 && rowsPerSecond > 0 {
					remaining := totalRecords - rowCount
					etaSeconds := remaining / rowsPerSecond
					if etaSeconds < 60 {
						eta = fmt.Sprintf("%ds", etaSeconds)
					} else if etaSeconds < 3600 {
						eta = fmt.Sprintf("%dm %ds", etaSeconds/60, etaSeconds%60)
					} else {
						eta = fmt.Sprintf("%dh %dm", etaSeconds/3600, (etaSeconds%3600)/60)
					}
				} else {
					eta = "Calculating..."
				}
				
				c.wsManager.SendLoadTestProgress(testID, map[string]any{
					"phase":           "insertion",
					"overallProgress": 25 + (progress * 0.75),
					"phaseProgress":   progress,
					"currentPhase":    "Plaid COPY Insertion",
					"rowsProcessed":   rowCount,
					"rowsPerSecond":   rowsPerSecond,
					"eta":             eta,
					"message": fmt.Sprintf(
						"Streaming records to database (%d/%d)...",
						rowCount,
						totalRecords,
					),
				})
				lastUpdateTime = time.Now()
			}
		}
	}()

	// Wait for all workers to finish
	wg.Wait()
	insertEndTime := time.Now()

	// Check if any error occurred during the process
	select {
	case err := <-errChan:
		return PlaidTimingResult{}, err
	default:
	}

	// Calculate wall-clock times
	parseTimeMs := int(parseEndTime.Sub(parseStartTime).Milliseconds())
	insertTimeMs := int(insertEndTime.Sub(parseEndTime).Milliseconds())
	totalTimeMs := int(insertEndTime.Sub(startTime).Milliseconds())

	return PlaidTimingResult{
		ParseTime:  parseTimeMs,
		InsertTime: insertTimeMs,
		TotalTime:  totalTimeMs,
	}, nil
}

func (c *PlaidController) getValueOrNull(value string) interface{} {
	if value != "" {
		return value
	}
	return nil
}

func (c *PlaidController) getNormalizedDateOrNull(value string) interface{} {
	if value != "" {
		if t, err := time.Parse("2006-01-02", value); err == nil {
			return t
		}
		if t, err := time.Parse("01/02/2006", value); err == nil {
			return t
		}
	}
	return nil
}
