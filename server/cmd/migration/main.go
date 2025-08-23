package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"server/cmd/migration/initialize"
	"server/cmd/migration/seed"
	"server/config"
	"server/internal/database"
	"server/internal/logger"
	. "server/internal/models"
	"strconv"

	migrate "github.com/rubenv/sql-migrate"
	"gorm.io/gorm"
	_ "github.com/lib/pq"
)

const (
	MIGRATION_PATH = "cmd/migration/migrations"
	MIGRATION_DB   = "postgres"
)

var MODELS_TO_MIGRATE = []any{
	&User{},
	&LoadTest{},
	&TestData{},
}

func main() {
	log := logger.New("migrations")
	log = log.Function("main")

	config, err := config.InitConfig()
	if err != nil {
		log.Er("failed to initialize config", err)
		os.Exit(1)
	}

	db, err := database.New(config)
	if err != nil {
		log.Er("failed to create database", err)
		os.Exit(1)
	}

	// Get flags from command line
	migrationType := "up"
	if len(os.Args) > 1 {
		migrationType = os.Args[1]
	}

	switch migrationType {
	case "up":
		err = migrateUp(db.SQL, config, log)
	case "down":
		steps := 1
		if len(os.Args) > 2 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil {
				log.Er("failed to parse step", err)
				os.Exit(1)
			}
		}
		err = migrateDown(steps, config, log)
	case "seed":
		err = migrateSeed(db, config, log)
	}

	if err != nil {
		log.Er("failed to run migrations", err)
		os.Exit(1)
	}

	log.Info("Migrations complete")
}

func migrateUp(db *gorm.DB, config config.Config, log logger.Logger) error {
	log = log.Function("migrateUp")
	log.Info("Running migrations up")

	err := runMigrations(config, log, migrate.Up)
	if err != nil {
		return log.Err("failed to run migrations", err)
	}

	err = autoMigrate(db, log)
	if err != nil {
		return log.Err("failed to auto migrate", err)
	}

	err = initialize.InitializeTables(db, config, log)
	if err != nil {
		return log.Err("failed to initialize tables", err)
	}

	return nil
}

func migrateDown(steps int, config config.Config, log logger.Logger) error {
	log = log.Function("migrateDown")
	log.Info("Running migrations down")

	for range steps {
		err := runMigrations(config, log, migrate.Down)
		if err != nil {
			return log.Err("failed to run migrations", err)
		}
	}

	return nil
}

func migrateSeed(db database.DB, config config.Config, log logger.Logger) error {
	log = log.Function("migrateSeed")
	log.Info("Running seed")

	// Clean DB to get to a fresh state before seeding
	if err := cleanDatabase(db.SQL, log); err != nil {
		return log.Err("failed to clean database", err)
	}

	// Flush all cache databases
	if err := db.FlushAllCaches(); err != nil {
		return log.Err("failed to flush cache databases", err)
	}

	if err := migrateUp(db.SQL, config, log); err != nil {
		return log.Err("failed to auto migrate", err)
	}

	if err := autoMigrate(db.SQL, log); err != nil {
		return log.Err("failed to auto migrate", err)
	}

	log.Info("Seeding database")
	if err := seed.Seed(db.SQL, config, log); err != nil {
		return log.Err("failed to seed database", err)
	}

	return nil
}

func autoMigrate(db *gorm.DB, log logger.Logger) error {
	log = log.Function("autoMigrate")

	dbTables := MODELS_TO_MIGRATE

	log.Info("GORM auto-migrating tables", "tables", dbTables)
	err := db.AutoMigrate(dbTables...)
	if err != nil {
		return log.Err("failed to auto migrate", err)
	}

	return nil
}

func runMigrations(
	config config.Config,
	log logger.Logger,
	direction migrate.MigrationDirection,
) error {
	log = log.Function("runMigrations")

	if _, err := os.Stat(MIGRATION_PATH); os.IsNotExist(err) {
		log.Info("Migrations directory does not exist, skipping file-based migrations")
		return nil
	}

	files, err := filepath.Glob(filepath.Join(MIGRATION_PATH, "*.sql"))
	if err != nil {
		return log.Err("failed to check for migration files", err)
	}

	if len(files) == 0 {
		log.Info("No migration files found, skipping file-based migrations")
		return nil
	}

	migrations := &migrate.FileMigrationSource{
		Dir: MIGRATION_PATH,
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.DatabaseHost,
		config.DatabasePort,
		config.DatabaseUser,
		config.DatabasePassword,
		config.DatabaseName,
	)

	db, err := sql.Open(MIGRATION_DB, dsn)
	if err != nil {
		return log.Err("failed to open database for migrations", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Er("failed to close database", err)
		}
	}()

	n, err := migrate.Exec(db, MIGRATION_DB, migrations, direction)
	if err != nil {
		return log.Err("failed to run migrations", err)
	}

	if n == 0 {
		log.Info("No migrations to apply")
	} else {
		log.Info("Applied migrations", "migrationCount", n)
	}

	return nil
}

func cleanDatabase(db *gorm.DB, log logger.Logger) error {
	log = log.Function("cleanDatabase")
	log.Info("Cleaning database before seeding")

	// Drop all tables to get a completely fresh start
	// This is simpler and more thorough than selective deletion
	if err := db.Migrator().DropTable(MODELS_TO_MIGRATE...); err != nil {
		return log.Err("failed to drop tables", err)
	}
	log.Info("Dropped all tables successfully")

	log.Info("Database cleaned successfully")
	return nil
}
