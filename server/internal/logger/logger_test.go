package logger

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew_Success(t *testing.T) {
	logger := New("test-package")

	assert.NotNil(t, logger)
	assert.IsType(t, &SlogLogger{}, logger)
}

func TestWith_ChainMethod(t *testing.T) {
	logger := New("test")

	newLogger := logger.With("key1", "value1")

	assert.NotNil(t, newLogger)
	assert.IsType(t, &SlogLogger{}, newLogger)
}

func TestFile_Method(t *testing.T) {
	logger := New("test")

	fileLogger := logger.File("user.controller.go")

	assert.NotNil(t, fileLogger)
	assert.IsType(t, &SlogLogger{}, fileLogger)
}

func TestFunction_Method(t *testing.T) {
	logger := New("test")

	funcLogger := logger.Function("handleLogin")

	assert.NotNil(t, funcLogger)
	assert.IsType(t, &SlogLogger{}, funcLogger)
}

func TestTimer_Functionality(t *testing.T) {
	logger := New("test")

	done := logger.Timer("test operation")

	assert.NotNil(t, done)
	assert.IsType(t, func() {}, done)

	// Complete the timer (should not panic)
	done()
}

func TestError_Methods(t *testing.T) {
	logger := New("test")

	// Test Error method
	err := logger.Error("test error message")

	assert.Error(t, err)
	assert.Equal(t, "test error message", err.Error())
}

func TestErr_Method(t *testing.T) {
	logger := New("test")

	originalErr := errors.New("original error")
	returnedErr := logger.Err("context message", originalErr)

	assert.Error(t, returnedErr)
	assert.Equal(t, originalErr, returnedErr)
}

func TestEr_Method(t *testing.T) {
	logger := New("test")

	originalErr := errors.New("test error")

	// Er method should not return anything (void)
	logger.Er("error occurred", originalErr)

	// No assertion needed as it's a void method, just verify it doesn't panic
}

func TestErrMsg_Method(t *testing.T) {
	logger := New("test")

	err := logger.ErrMsg("simple error message")

	assert.Error(t, err)
	assert.Equal(t, "simple error message", err.Error())
}

func TestLoggerInterface_Implementation(t *testing.T) {
	// Verify SlogLogger implements Logger interface
	logger := New("test")

	assert.NotNil(t, logger)

	// Test key interface methods are callable
	err := logger.Error("error test")
	assert.Error(t, err)

	chainedLogger := logger.With("test", "value")
	assert.NotNil(t, chainedLogger)

	fileLogger := logger.File("test.go")
	assert.NotNil(t, fileLogger)

	funcLogger := logger.Function("testFunc")
	assert.NotNil(t, funcLogger)

	timer := logger.Timer("test timer")
	assert.NotNil(t, timer)
	timer()
}

// Negative Test Cases

func TestErr_NilError(t *testing.T) {
	logger := New("test")

	returnedErr := logger.Err("message", nil)

	assert.Nil(t, returnedErr) // Should return the nil error
}

func TestEr_NilError(t *testing.T) {
	logger := New("test")

	// Should not panic with nil error
	logger.Er("message", nil)
}

func TestStep_Method(t *testing.T) {
	// Create a logger that we can capture output from
	var capturedLogs []string
	handler := &testHandler{logs: &capturedLogs}
	logger := &SlogLogger{logger: slog.New(handler)}

	logger.Step("test step message")

	assert.Len(t, capturedLogs, 1)
	assert.Contains(t, capturedLogs[0], "test step message")
}

func TestDebug_Method(t *testing.T) {
	var capturedLogs []string
	handler := &testHandler{logs: &capturedLogs}
	logger := &SlogLogger{logger: slog.New(handler)}

	logger.Debug("debug message", "key", "value")

	assert.Len(t, capturedLogs, 1)
	assert.Contains(t, capturedLogs[0], "debug message")
	assert.Contains(t, capturedLogs[0], "key")
	assert.Contains(t, capturedLogs[0], "value")
}

func TestWarn_Method(t *testing.T) {
	var capturedLogs []string
	handler := &testHandler{logs: &capturedLogs}
	logger := &SlogLogger{logger: slog.New(handler)}

	logger.Warn("warning message", "key", "value")

	assert.Len(t, capturedLogs, 1)
	assert.Contains(t, capturedLogs[0], "warning message")
}

func TestInfo_Method(t *testing.T) {
	var capturedLogs []string
	handler := &testHandler{logs: &capturedLogs}
	logger := &SlogLogger{logger: slog.New(handler)}

	logger.Info("info message", "key", "value")

	assert.Len(t, capturedLogs, 1)
	assert.Contains(t, capturedLogs[0], "info message")
}

func TestErrorf_Method(t *testing.T) {
	logger := New("test")
	err := logger.Errorf("test error", "detailed message")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "detailed message")
}

func TestErMsg_VoidMethod(t *testing.T) {
	var capturedLogs []string
	handler := &testHandler{logs: &capturedLogs}
	logger := &SlogLogger{logger: slog.New(handler)}

	logger.ErMsg("simple error message")

	assert.Len(t, capturedLogs, 1)
	assert.Contains(t, capturedLogs[0], "simple error message")
}

// Test helper to capture log output
type testHandler struct {
	logs *[]string
}

func (h *testHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *testHandler) Handle(_ context.Context, record slog.Record) error {
	// Build the full log message including attributes
	var parts []string
	parts = append(parts, record.Message)

	record.Attrs(func(attr slog.Attr) bool {
		parts = append(parts, fmt.Sprintf("%s=%v", attr.Key, attr.Value))
		return true
	})

	fullMessage := strings.Join(parts, " ")
	*h.logs = append(*h.logs, fullMessage)
	return nil
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return h
}
