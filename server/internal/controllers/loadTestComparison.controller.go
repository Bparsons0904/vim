package controllers

import (
	"context"
	"fmt"
	"server/internal/database"
	"server/internal/logger"
	"server/internal/repositories"
	"time"

	"github.com/google/uuid"
)

type LoadTestComparisonController struct {
	optimizedController *OptimizedLoadTestController
	ludicrousController *LudicrousLoadTestController
	log                 logger.Logger
}

func NewLoadTestComparisonController(
	db database.DB,
	loadTestRepo repositories.LoadTestRepository,
	testDataRepo repositories.TestDataRepository,
	wsManager WSManager,
) *LoadTestComparisonController {
	return &LoadTestComparisonController{
		optimizedController: NewOptimizedLoadTestController(db, loadTestRepo, testDataRepo, wsManager),
		ludicrousController: NewLudicrousLoadTestController(db, loadTestRepo, testDataRepo, wsManager),
		log:                 logger.New("loadTestComparisonController"),
	}
}

// ComparisonResult holds the results of comparing insertion methods
type ComparisonResult struct {
	OptimizedResult *InsertionResult `json:"optimized_result"`
	LudicrousResult *InsertionResult `json:"ludicrous_result"`
	Winner          string           `json:"winner"`
	Improvement     string           `json:"improvement"`
}

// InsertionResult holds the results of a single insertion test
type InsertionResult struct {
	Method       string `json:"method"`
	TotalTimeMs  int    `json:"total_time_ms"`
	RecordsCount int    `json:"records_count"`
	RowsPerSec   int    `json:"rows_per_second"`
	Success      bool   `json:"success"`
	ErrorMsg     string `json:"error_message,omitempty"`
}

// CompareInsertionMethods runs both Optimized and Ludicrous insertion methods and compares their performance
func (c *LoadTestComparisonController) CompareInsertionMethods(
	ctx context.Context,
	csvPath string,
	totalRecords int,
	testID string,
) (*ComparisonResult, error) {
	c.log.Info("Starting insertion method comparison",
		"csvPath", csvPath,
		"totalRecords", totalRecords)

	result := &ComparisonResult{}

	// Test Optimized method (GORM)
	c.log.Info("Testing Optimized insertion method")
	optimizedResult := c.runInsertionTest(ctx, csvPath, totalRecords, testID, "optimized")
	result.OptimizedResult = optimizedResult

	// Clean up data between tests
	if err := c.cleanupTestData(ctx, optimizedResult.Method); err != nil {
		c.log.Warn("Failed to cleanup Optimized test data", "error", err)
	}

	// Wait a moment to let the database settle
	time.Sleep(1 * time.Second)

	// Test Ludicrous method (Raw SQL)
	c.log.Info("Testing Ludicrous insertion method")
	ludicrousResult := c.runInsertionTest(ctx, csvPath, totalRecords, testID, "ludicrous")
	result.LudicrousResult = ludicrousResult

	// Determine winner and calculate improvement
	if optimizedResult.Success && ludicrousResult.Success {
		if ludicrousResult.TotalTimeMs < optimizedResult.TotalTimeMs {
			result.Winner = "Ludicrous"
			improvement := float64(optimizedResult.TotalTimeMs-ludicrousResult.TotalTimeMs) / float64(optimizedResult.TotalTimeMs) * 100
			result.Improvement = fmt.Sprintf("%.1f%% faster", improvement)
		} else {
			result.Winner = "Optimized"
			improvement := float64(ludicrousResult.TotalTimeMs-optimizedResult.TotalTimeMs) / float64(ludicrousResult.TotalTimeMs) * 100
			result.Improvement = fmt.Sprintf("%.1f%% faster", improvement)
		}
	} else if optimizedResult.Success {
		result.Winner = "Optimized"
		result.Improvement = "Ludicrous failed"
	} else if ludicrousResult.Success {
		result.Winner = "Ludicrous"
		result.Improvement = "Optimized failed"
	} else {
		result.Winner = "None"
		result.Improvement = "Both methods failed"
	}

	c.log.Info("Insertion method comparison completed",
		"winner", result.Winner,
		"improvement", result.Improvement,
		"optimizedTimeMs", optimizedResult.TotalTimeMs,
		"ludicrousTimeMs", ludicrousResult.TotalTimeMs)

	return result, nil
}

// runInsertionTest runs a single insertion test with the specified method
func (c *LoadTestComparisonController) runInsertionTest(
	ctx context.Context,
	csvPath string,
	totalRecords int,
	testID string,
	method string,
) *InsertionResult {
	result := &InsertionResult{
		Method:       method,
		RecordsCount: totalRecords,
	}

	loadTestID := uuid.New()
	startTime := time.Now()

	var totalTimeMs int
	var err error

	switch method {
	case "ludicrous":
		totalTimeMs, err = c.ludicrousController.InsertLudicrousSpeed(
			ctx, csvPath, loadTestID, totalRecords, startTime, testID)
	case "optimized":
		fallthrough
	default:
		totalTimeMs, err = c.optimizedController.InsertOptimizedWithProgress(
			ctx, csvPath, loadTestID, totalRecords, startTime, testID)
	}

	result.TotalTimeMs = totalTimeMs
	if err != nil {
		result.Success = false
		result.ErrorMsg = err.Error()
		c.log.Error("Insertion test failed", "method", method, "error", err)
	} else {
		result.Success = true
		if totalTimeMs > 0 {
			result.RowsPerSec = int(float64(totalRecords) / (float64(totalTimeMs) / 1000))
		}
		c.log.Info("Insertion test completed", 
			"method", method, 
			"totalTimeMs", totalTimeMs,
			"rowsPerSec", result.RowsPerSec)
	}

	return result
}

// cleanupTestData removes test data to ensure fair comparison
func (c *LoadTestComparisonController) cleanupTestData(ctx context.Context, method string) error {
	// This would typically clean up test data from the previous run
	// For now, just log that we would clean up
	c.log.Info("Would cleanup test data for method", "method", method)
	return nil
}