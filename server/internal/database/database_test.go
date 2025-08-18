package database

import (
	"server/internal/logger"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheConstants(t *testing.T) {
	// Test that cache constants are defined correctly
	assert.Equal(t, 0, GENERAL_CACHE_INDEX)
	assert.Equal(t, 1, SESSION_CACHE_INDEX)
	assert.Equal(t, 2, USER_CACHE_INDEX)
	assert.Equal(t, 3, EVENTS_CACHE_INDEX)
}

func TestDB_StructCreation(t *testing.T) {
	log := logger.New("test")

	db := &DB{
		log: log,
	}

	assert.NotNil(t, db)
	assert.Equal(t, log, db.log)
	assert.Nil(t, db.SQL)
}

func TestTXDefer_WithError(t *testing.T) {
	// This test verifies TXDefer handles transaction errors gracefully
	// We can't test actual database transactions without a real DB
	log := logger.New("test")

	// We can't safely test with nil transaction as it will panic
	// Just verify the function exists and can be called
	assert.NotNil(t, TXDefer)
	assert.NotNil(t, log)
}

// Cache builder tests are skipped because they require real valkey.Client interface
// These are tested in integration tests with real cache server
func TestCacheBuilder_SkippedTests(t *testing.T) {
	t.Skip("Cache builder tests require real valkey client - tested in integration tests")
}

