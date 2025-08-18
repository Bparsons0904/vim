package database

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"server/config"
	logg "server/internal/logger"
	"time"

	"github.com/valkey-io/valkey-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type CacheClient valkey.Client

type Cache struct {
	General CacheClient
	Session CacheClient
	User    CacheClient
	Events  CacheClient
	Story   CacheClient
}

type DB struct {
	SQL   *gorm.DB
	Cache Cache
	log   logg.Logger
}

func New(config config.Config) (DB, error) {
	log := logg.New("database").Function("New")

	log.Info("Initializing database")
	db := &DB{log: log}

	err := db.initializeDB(config)
	if err != nil {
		return DB{}, log.Err("failed to initialize database", err)
	}

	err = db.initializeCacheDB(config)
	if err != nil {
		return DB{}, log.Err("failed to initialize cache database", err)
	}

	return *db, nil
}

func TXDefer(tx *gorm.DB, log logg.Logger) {
	if tx.Error != nil {
		log.Er("failed to commit transaction", tx.Error)
		tx.Rollback()
	} else {
		err := tx.Commit().Error
		if err != nil {
			log.Er("failed to commit transaction", err)
		} else {
			log.Info("committed transaction")
		}
	}
}

func (s *DB) initializeDB(config config.Config) error {
	gormLogger := logger.New(
		slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo),
		logger.Config{
			SlowThreshold:             1 * time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      false,
			Colorful:                  false,
		},
	)

	gormConfig := &gorm.Config{
		Logger:                                   gormLogger,
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
		CreateBatchSize:                          100,
	}

	return s.initializeSQLiteDB(gormConfig, config)
}

func (s *DB) initializeSQLiteDB(gormConfig *gorm.Config, config config.Config) error {
	log := s.log.Function("initializeSQLiteDB")

	dbPath := config.DatabaseDbPath
	if dbPath == "" {
		return log.Error("database path is empty", "dbPath", dbPath)
	}

	dir := filepath.Dir(dbPath)
	log.Info("Creating database directory", "dir", dir)
	if err := os.MkdirAll("data", 0755); err != nil {
		return log.Err("failed to create database directory", err, "dir", dir)
	}

	log.Info("Connecting with GORM", "dbPath", dbPath)
	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return log.Err("failed to open database with GORM", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return log.Err("failed to get database from GORM", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return log.Err("failed to ping database through GORM", err)
	}

	log.Info("Successfully connected with GORM")
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	s.SQL = db

	return nil
}

func (s *DB) Close() (err error) {
	if s.SQL != nil {
		sqlDB, err := s.SQL.DB()
		if err == nil {
			if err := sqlDB.Close(); err != nil {
				_ = s.log.Err("failed to close database", err)
			}
		}
	}

	if s.Cache.General != nil {
		s.Cache.General.Close()
	}

	if s.Cache.Session != nil {
		s.Cache.Session.Close()
	}

	if s.Cache.Events != nil {
		s.Cache.Events.Close()
	}

	return
}

func (s *DB) SQLWithContext(ctx context.Context) *gorm.DB {
	return s.SQL.WithContext(ctx)
}

func (s *DB) FlushAllCaches() error {
	log := s.log.Function("FlushAllCaches")
	log.Info("Flushing all cache databases")
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	cacheClients := []struct {
		client CacheClient
		name   string
	}{
		{s.Cache.General, "General"},
		{s.Cache.Session, "Session"},
		{s.Cache.User, "User"},
		{s.Cache.Events, "Events"},
		{s.Cache.Story, "Story"},
	}
	
	for _, cache := range cacheClients {
		if cache.client != nil {
			if err := cache.client.Do(ctx, cache.client.B().Flushdb().Build()).Error(); err != nil {
				log.Er("Failed to flush cache database", err, "cache", cache.name)
				return err
			}
			log.Info("Successfully flushed cache database", "cache", cache.name)
		}
	}
	
	log.Info("All cache databases flushed successfully")
	return nil
}
