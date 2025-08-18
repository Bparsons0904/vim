package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

type Logger interface {
	Errorf(msg string, errMessage string) error
	Error(msg string, args ...any) error
	Err(msg string, err error, args ...any) error
	ErrMsg(msg string) error
	ErMsg(msg string)
	Er(msg string, err error, args ...any)
	Step(msg string)
	Debug(msg string, args ...any)
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
	With(args ...any) Logger
	File(name string) Logger
	Function(name string) Logger
	Timer(msg string) func()
}

type SlogLogger struct {
	logger *slog.Logger
}

func New(name string) Logger {
	var handler slog.Handler

	if isTestMode() {
		handler = slog.NewTextHandler(io.Discard, nil)
	} else {
		handler = slog.Default().Handler()
	}

	return &SlogLogger{
		logger: slog.New(handler).With("package", name),
	}
}

func isTestMode() bool {
	for _, arg := range os.Args {
		if arg == "-test.v" || arg == "-test.run" || arg == "-test.bench" {
			return true
		}
	}
	return false
}

func (l *SlogLogger) With(args ...any) Logger {
	return &SlogLogger{
		logger: l.logger.With(args...),
	}
}

func (l *SlogLogger) Error(msg string, args ...any) error {
	l.logger.Error(msg, args...)
	return fmt.Errorf("%s", msg)
}

func (l *SlogLogger) File(name string) Logger {
	return l.With("file", name)
}

func (l *SlogLogger) Function(name string) Logger {
	return l.With("function", name)
}

func (l *SlogLogger) Timer(msg string) func() {
	start := time.Now()
	l.logger.Debug("Starting", "operation", msg)

	return func() {
		duration := time.Since(start)
		l.logger.Info("Timer Completed",
			"operation", msg,
			"duration_ms", duration.Milliseconds(),
			"duration", duration.String(),
		)
	}
}

func (l *SlogLogger) Errorf(msg string, errMessage string) error {
	err := fmt.Errorf("error: %s", errMessage)
	l.logger.Error(msg, "error", err)
	return err
}

func (l *SlogLogger) Er(msg string, err error, args ...any) {
	logArgs := append([]any{"error", err}, args...)
	l.logger.Error(msg, logArgs...)
}

func (l *SlogLogger) Err(msg string, err error, args ...any) error {
	logArgs := append([]any{"error", err}, args...)
	l.logger.Error(msg, logArgs...)
	return err
}

func (l *SlogLogger) ErMsg(msg string) {
	l.logger.Error(msg)
}

func (l *SlogLogger) ErrMsg(msg string) error {
	l.logger.Error(msg)
	return fmt.Errorf("%s", msg)
}

func (l *SlogLogger) Step(msg string) {
	l.logger.Info(msg)
}

func (l *SlogLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}
