package repositories

import (
	"context"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"server/internal/services"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	LOAD_TEST_CACHE_EXPIRY = 24 * time.Hour // 24 hours
)

type LoadTestRepository interface {
	GetByID(ctx context.Context, id string) (*LoadTest, error)
	Create(ctx context.Context, loadTest *LoadTest) error
	Update(ctx context.Context, loadTest *LoadTest) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]*LoadTest, error)
	GetAllForSummary(ctx context.Context) ([]*LoadTest, error)
	GetByStatus(ctx context.Context, status string) ([]*LoadTest, error)
}

type loadTestRepository struct {
	db  database.DB
	log logger.Logger
}

func NewLoadTest(db database.DB) LoadTestRepository {
	return &loadTestRepository{
		db:  db,
		log: logger.New("loadTestRepository"),
	}
}

func (r *loadTestRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := services.GetTransaction(ctx); ok {
		return tx
	}
	return r.db.SQLWithContext(ctx)
}

func (r *loadTestRepository) GetByID(ctx context.Context, id string) (*LoadTest, error) {
	log := r.log.Function("GetByID")

	var loadTest LoadTest
	if err := r.getCacheByID(ctx, id, &loadTest); err == nil {
		return &loadTest, nil
	}

	if err := r.getDBByID(ctx, id, &loadTest); err != nil {
		return nil, err
	}

	if err := r.addLoadTestToCache(ctx, &loadTest); err != nil {
		log.Warn("failed to add load test to cache", "loadTestID", id, "error", err)
	}

	return &loadTest, nil
}

func (r *loadTestRepository) Create(ctx context.Context, loadTest *LoadTest) error {
	log := r.log.Function("Create")

	if err := r.getDB(ctx).Create(loadTest).Error; err != nil {
		return log.Err("failed to create load test", err, "loadTest", loadTest)
	}

	if err := r.addLoadTestToCache(ctx, loadTest); err != nil {
		log.Warn("failed to add load test to cache", "loadTestID", loadTest.ID, "error", err)
	}

	return nil
}

func (r *loadTestRepository) Update(ctx context.Context, loadTest *LoadTest) error {
	log := r.log.Function("Update")

	if err := r.getDB(ctx).Save(loadTest).Error; err != nil {
		return log.Err("failed to update load test", err, "loadTest", loadTest)
	}

	if err := r.addLoadTestToCache(ctx, loadTest); err != nil {
		log.Warn("failed to update load test in cache", "loadTestID", loadTest.ID, "error", err)
	}

	return nil
}

func (r *loadTestRepository) Delete(ctx context.Context, id string) error {
	log := r.log.Function("Delete")

	if err := r.getDB(ctx).Delete(&LoadTest{}, "id = ?", id).Error; err != nil {
		return log.Err("failed to delete load test", err, "id", id)
	}

	if err := database.NewCacheBuilder(r.db.Cache.LoadTest, id).Delete(); err != nil {
		log.Warn("failed to remove load test from cache", "loadTestID", id, "error", err)
	}

	return nil
}

func (r *loadTestRepository) GetAll(ctx context.Context) ([]*LoadTest, error) {
	log := r.log.Function("GetAll")

	var loadTests []*LoadTest
	if err := r.getDB(ctx).Order("created_at DESC").Limit(10).Find(&loadTests).Error; err != nil {
		return nil, log.Err("failed to get all load tests", err)
	}

	return loadTests, nil
}

func (r *loadTestRepository) GetAllForSummary(ctx context.Context) ([]*LoadTest, error) {
	log := r.log.Function("GetAllForSummary")

	var loadTests []*LoadTest
	if err := r.getDB(ctx).Order("created_at DESC").Find(&loadTests).Error; err != nil {
		return nil, log.Err("failed to get all load tests for summary", err)
	}

	return loadTests, nil
}


func (r *loadTestRepository) GetByStatus(ctx context.Context, status string) ([]*LoadTest, error) {
	log := r.log.Function("GetByStatus")

	var loadTests []*LoadTest
	if err := r.getDB(ctx).Where("status = ?", status).Find(&loadTests).Error; err != nil {
		return nil, log.Err("failed to get load tests by status", err, "status", status)
	}

	return loadTests, nil
}

func (r *loadTestRepository) getCacheByID(ctx context.Context, loadTestID string, loadTest *LoadTest) error {
	found, err := database.NewCacheBuilder(r.db.Cache.LoadTest, loadTestID).Get(loadTest)
	if err != nil {
		return r.log.Function("getCacheByID").
			Err("failed to get load test from cache", err, "loadTestID", loadTestID)
	}

	if !found {
		return r.log.Function("getCacheByID").
			Error("load test not found in cache", "loadTestID", loadTestID)
	}

	return nil
}

func (r *loadTestRepository) addLoadTestToCache(ctx context.Context, loadTest *LoadTest) error {
	if err := database.NewCacheBuilder(r.db.Cache.LoadTest, loadTest.ID).
		WithStruct(loadTest).
		WithTTL(LOAD_TEST_CACHE_EXPIRY).
		WithContext(ctx).
		Set(); err != nil {
		return r.log.Function("addLoadTestToCache").
			Err("failed to add load test to cache", err, "loadTest", loadTest)
	}
	return nil
}

func (r *loadTestRepository) getDBByID(ctx context.Context, loadTestID string, loadTest *LoadTest) error {
	log := r.log.Function("getDBByID")

	id, err := uuid.Parse(loadTestID)
	if err != nil {
		return log.Err("failed to parse loadTestID", err, "loadTestID", loadTestID)
	}

	if err := r.getDB(ctx).First(loadTest, "id = ?", id).Error; err != nil {
		return log.Err("failed to get load test by id", err, "id", loadTestID)
	}

	return nil
}