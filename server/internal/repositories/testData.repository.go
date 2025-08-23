package repositories

import (
	"context"
	"fmt"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/services"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TestDataRepository interface {
	GetByID(ctx context.Context, id string) (*TestData, error)
	Create(ctx context.Context, testData *TestData) error
	CreateBatch(ctx context.Context, testDataBatch []*TestData, batchSize int) error
	GetByLoadTestID(ctx context.Context, loadTestID string) ([]*TestData, error)
	GetByLoadTestIDPaginated(
		ctx context.Context,
		loadTestID string,
		offset, limit int,
	) ([]*TestData, error)
	CountByLoadTestID(ctx context.Context, loadTestID string) (int64, error)
	Delete(ctx context.Context, id string) error
	DeleteByLoadTestID(ctx context.Context, loadTestID string) error
}

type testDataRepository struct {
	db  database.DB
	log logger.Logger
}

func NewTestData(db database.DB) TestDataRepository {
	return &testDataRepository{
		db:  db,
		log: logger.New("testDataRepository"),
	}
}

func (r *testDataRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := services.GetTransaction(ctx); ok {
		return tx
	}
	return r.db.SQLWithContext(ctx)
}

func (r *testDataRepository) GetByID(ctx context.Context, id string) (*TestData, error) {
	var testData TestData
	if err := r.getDBByID(ctx, id, &testData); err != nil {
		return nil, err
	}

	return &testData, nil
}

func (r *testDataRepository) Create(ctx context.Context, testData *TestData) error {
	log := r.log.Function("Create")

	if err := r.getDB(ctx).Create(testData).Error; err != nil {
		return log.Err("failed to create test data", err, "testData", testData)
	}

	return nil
}

func (r *testDataRepository) CreateBatch(
	ctx context.Context,
	testDataBatch []*TestData,
	batchSize int,
) error {
	log := r.log.Function("CreateBatch")

	if len(testDataBatch) == 0 {
		return log.Error("empty test data batch provided")
	}

	if batchSize <= 0 {
		batchSize = 1000 // Default batch size
	}

	db := r.getDB(ctx)

	if err := db.CreateInBatches(testDataBatch, batchSize).Error; err != nil {
		return log.Err("failed to create test data batch", err,
			"totalRecords", len(testDataBatch),
			"batchSize", batchSize)
	}

	log.Info(
		"successfully inserted all test data",
		"totalRecords",
		len(testDataBatch),
		"batchSize",
		batchSize,
	)
	return nil
}

func (r *testDataRepository) GetByLoadTestID(
	ctx context.Context,
	loadTestID string,
) ([]*TestData, error) {
	log := r.log.Function("GetByLoadTestID")

	loadTestUUID, err := uuid.Parse(loadTestID)
	if err != nil {
		return nil, log.Err("failed to parse loadTestID", err, "loadTestID", loadTestID)
	}

	var testData []*TestData
	if err := r.getDB(ctx).Where("load_test_id = ?", loadTestUUID).Find(&testData).Error; err != nil {
		return nil, log.Err(
			"failed to get test data by load test ID",
			err,
			"loadTestID",
			loadTestID,
		)
	}

	return testData, nil
}

func (r *testDataRepository) GetByLoadTestIDPaginated(
	ctx context.Context,
	loadTestID string,
	offset, limit int,
) ([]*TestData, error) {
	log := r.log.Function("GetByLoadTestIDPaginated")

	loadTestUUID, err := uuid.Parse(loadTestID)
	if err != nil {
		return nil, log.Err("failed to parse loadTestID", err, "loadTestID", loadTestID)
	}

	var testData []*TestData
	query := r.getDB(ctx).Where("load_test_id = ?", loadTestUUID)

	if offset > 0 {
		query = query.Offset(offset)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&testData).Error; err != nil {
		return nil, log.Err("failed to get paginated test data by load test ID", err,
			"loadTestID", loadTestID, "offset", offset, "limit", limit)
	}

	return testData, nil
}

func (r *testDataRepository) CountByLoadTestID(
	ctx context.Context,
	loadTestID string,
) (int64, error) {
	log := r.log.Function("CountByLoadTestID")

	loadTestUUID, err := uuid.Parse(loadTestID)
	if err != nil {
		return 0, log.Err("failed to parse loadTestID", err, "loadTestID", loadTestID)
	}

	var count int64
	if err := r.getDB(ctx).Model(&TestData{}).Where("load_test_id = ?", loadTestUUID).Count(&count).Error; err != nil {
		return 0, log.Err(
			"failed to count test data by load test ID",
			err,
			"loadTestID",
			loadTestID,
		)
	}

	return count, nil
}

func (r *testDataRepository) Delete(ctx context.Context, id string) error {
	log := r.log.Function("Delete")

	if err := r.getDB(ctx).Delete(&TestData{}, "id = ?", id).Error; err != nil {
		return log.Err("failed to delete test data", err, "id", id)
	}

	return nil
}

func (r *testDataRepository) DeleteByLoadTestID(ctx context.Context, loadTestID string) error {
	log := r.log.Function("DeleteByLoadTestID")

	loadTestUUID, err := uuid.Parse(loadTestID)
	if err != nil {
		return log.Err("failed to parse loadTestID", err, "loadTestID", loadTestID)
	}

	result := r.getDB(ctx).Where("load_test_id = ?", loadTestUUID).Delete(&TestData{})
	if result.Error != nil {
		return log.Err(
			"failed to delete test data by load test ID",
			result.Error,
			"loadTestID",
			loadTestID,
		)
	}

	log.Info(
		"deleted test data records",
		"loadTestID",
		loadTestID,
		"deletedCount",
		result.RowsAffected,
	)
	return nil
}

func (r *testDataRepository) getDBByID(
	ctx context.Context,
	testDataID string,
	testData *TestData,
) error {
	log := r.log.Function("getDBByID")

	id, err := uuid.Parse(testDataID)
	if err != nil {
		return log.Err("failed to parse testDataID", err, "testDataID", testDataID)
	}

	if err := r.getDB(ctx).First(testData, "id = ?", id).Error; err != nil {
		return log.Err("failed to get test data by id", err, "id", testDataID)
	}

	return nil
}

// BatchInsertWithCustomColumns allows dynamic column mapping for performance testing scenarios
func (r *testDataRepository) BatchInsertWithCustomColumns(
	ctx context.Context,
	records []map[string]interface{},
	loadTestID uuid.UUID,
	batchSize int,
) error {
	log := r.log.Function("BatchInsertWithCustomColumns")

	if len(records) == 0 {
		return log.Error("empty records provided")
	}

	if batchSize <= 0 {
		batchSize = 1000
	}

	// Convert map records to TestData structs
	var testDataBatch []*TestData
	for _, record := range records {
		testData := &TestData{
			LoadTestID: loadTestID,
		}

		// Use reflection or manual mapping to populate fields
		// For performance, we'll manually map the most common columns
		if val, ok := record["col1"]; ok {
			if str, ok := val.(string); ok {
				testData.Col1 = &str
			}
		}
		if val, ok := record["col2"]; ok {
			if str, ok := val.(string); ok {
				testData.Col2 = &str
			}
		}
		if val, ok := record["col3"]; ok {
			if str, ok := val.(string); ok {
				testData.Col3 = &str
			}
		}
		// Continue mapping other columns as needed...

		testDataBatch = append(testDataBatch, testData)
	}

	return r.CreateBatch(ctx, testDataBatch, batchSize)
}

// GetColumnValue retrieves a specific column value from TestData by field name
func (r *testDataRepository) GetColumnValue(testData *TestData, columnName string) *string {
	switch columnName {
	case "col1":
		return testData.Col1
	case "col2":
		return testData.Col2
	case "col3":
		return testData.Col3
	case "col4":
		return testData.Col4
	case "col5":
		return testData.Col5
	// Add more cases as needed for other columns
	default:
		return nil
	}
}

// SetColumnValue sets a specific column value in TestData by field name
func (r *testDataRepository) SetColumnValue(
	testData *TestData,
	columnName string,
	value *string,
) error {
	switch columnName {
	case "col1":
		testData.Col1 = value
	case "col2":
		testData.Col2 = value
	case "col3":
		testData.Col3 = value
	case "col4":
		testData.Col4 = value
	case "col5":
		testData.Col5 = value
	// Add more cases as needed for other columns
	default:
		return fmt.Errorf("unknown column name: %s", columnName)
	}
	return nil
}

