package controllers

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"server/config"
	"server/internal/logger"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/lib/pq"
)

type PlaidTimingResult struct {
	ParseTime  int
	InsertTime int
	TotalTime  int
}

type PlaidController struct {
	log       logger.Logger
	config    config.Config
	wsManager WSManager
	db        *sql.DB
}

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

	// Set connection pool settings
	db.SetMaxOpenConns(25) // Sensible default
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PlaidController{
		log:       logger.New("plaidController"),
		config:    config,
		wsManager: wsManager,
		db:        db,
	}, nil
}

func (c *PlaidController) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

func (c *PlaidController) RunPlaidCopy(ctx context.Context, csvPath string, loadTestID uuid.UUID, totalRecords int) (PlaidTimingResult, error) {
	file, err := os.Open(csvPath)
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	timingResult, err := c.executeStreamingCopy(ctx, c.db, file, loadTestID, totalRecords, loadTestID.String())
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("streaming COPY failed: %w", err)
	}

	return timingResult, nil
}

func (c *PlaidController) executeStreamingCopy(
	ctx context.Context,
	db *sql.DB,
	file *os.File,
	loadTestID uuid.UUID,
	totalRecords int,
	testID string,
) (PlaidTimingResult, error) {
	startTime := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	copyColumns := []string{
		"load_test_id", "birth_date", "start_date", "end_date",
		"first_name", "last_name", "email", "phone", "address_line1", "address_line2",
		"city", "state", "zip_code", "country", "social_security_no",
		"employer", "job_title", "department", "salary",
		"insurance_plan_id", "insurance_carrier", "policy_number",
		"group_number", "member_id",
	}
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("test_data", copyColumns...))
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("failed to prepare COPY statement: %w", err)
	}
	defer stmt.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("failed to read CSV headers: %w", err)
	}

	headerIndex := make(map[string]int, len(headers))
	for i, h := range headers {
		headerIndex[h] = i
	}

	values := make([]interface{}, len(copyColumns))

	var totalParseTime int64
	var totalInsertTime int64
	rowCount := 0
	lastUpdateTime := time.Now()

	for {
		parseStart := time.Now()
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return PlaidTimingResult{}, fmt.Errorf("failed to read CSV row: %w", err)
		}

		loadTestIDStr := loadTestID.String()
		
		getValue := func(key string) string {
			if idx, ok := headerIndex[key]; ok && idx < len(record) {
				return record[idx]
			}
			return ""
		}

		values[0] = loadTestIDStr
		values[1] = c.getNormalizedDateOrNull(getValue("birth_date"))
		values[2] = c.getNormalizedDateOrNull(getValue("start_date"))
		values[3] = c.getNormalizedDateOrNull(getValue("end_date"))
		values[4] = c.getValueOrNull(getValue("first_name"))
		values[5] = c.getValueOrNull(getValue("last_name"))
		values[6] = c.getValueOrNull(getValue("email"))
		values[7] = c.getValueOrNull(getValue("phone"))
		values[8] = c.getValueOrNull(getValue("address_line1"))
		values[9] = c.getValueOrNull(getValue("address_line2"))
		values[10] = c.getValueOrNull(getValue("city"))
		values[11] = c.getValueOrNull(getValue("state"))
		values[12] = c.getValueOrNull(getValue("zip_code"))
		values[13] = c.getValueOrNull(getValue("country"))
		values[14] = c.getValueOrNull(getValue("social_security_no"))
		values[15] = c.getValueOrNull(getValue("employer"))
		values[16] = c.getValueOrNull(getValue("job_title"))
		values[17] = c.getValueOrNull(getValue("department"))
		values[18] = c.getValueOrNull(getValue("salary"))
		values[19] = c.getValueOrNull(getValue("insurance_plan_id"))
		values[20] = c.getValueOrNull(getValue("insurance_carrier"))
		values[21] = c.getValueOrNull(getValue("policy_number"))
		values[22] = c.getValueOrNull(getValue("group_number"))
		values[23] = c.getValueOrNull(getValue("member_id"))
		totalParseTime += time.Since(parseStart).Nanoseconds()

		insertStart := time.Now()
		_, err = stmt.Exec(values...)
		if err != nil {
			return PlaidTimingResult{}, fmt.Errorf("failed to execute COPY for record %d: %w", rowCount+1, err)
		}
		totalInsertTime += time.Since(insertStart).Nanoseconds()

		rowCount++

		if time.Since(lastUpdateTime) > 2*time.Second {
			elapsed := time.Since(startTime)
			rowsPerSecond := int(float64(rowCount) / elapsed.Seconds())
			progress := float64(rowCount) / float64(totalRecords) * 100
			c.wsManager.SendLoadTestProgress(testID, map[string]any{
				"phase":            "insertion",
				"overallProgress":  25 + (progress * 0.75),
				"phaseProgress":    progress,
				"currentPhase":     "Plaid COPY Insertion",
				"rowsProcessed":    rowCount,
				"rowsPerSecond":    rowsPerSecond,
				"eta":              "Calculating...",
				"message":          fmt.Sprintf("Streaming records to database (%d/%d)...", rowCount, totalRecords),
			})
			lastUpdateTime = time.Now()
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return PlaidTimingResult{}, fmt.Errorf("failed to finalize COPY operation: %w", err)
	}

	parseTimeMs := int(totalParseTime / 1e6)
	insertTimeMs := int(totalInsertTime / 1e6)

	return PlaidTimingResult{
		ParseTime:  parseTimeMs,
		InsertTime: insertTimeMs,
		TotalTime:  parseTimeMs + insertTimeMs,
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
		// This is a simplified date validation. A more robust solution would be needed for production.
		if t, err := time.Parse("2006-01-02", value); err == nil {
			return t
		}
		if t, err := time.Parse("01/02/2006", value); err == nil {
			return t
		}
	}
	return nil
}
