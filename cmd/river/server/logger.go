package server

import (
	"log/slog"
	"os"
	"path/filepath"
)

type Logger struct {
	ls      []*slog.Logger
	logfile *os.File
}

// Debug logs at LevelDebug.
func (l *Logger) Debug(msg string, args ...any) {
	for _, logger := range l.ls {
		logger.Debug(msg, args...)
	}
}

// Info logs at LevelInfo.
func (l *Logger) Info(msg string, args ...any) {
	for _, logger := range l.ls {
		logger.Info(msg, args...)
	}
}

// Warn logs at LevelWarn.
func (l *Logger) Warn(msg string, args ...any) {
	for _, logger := range l.ls {
		logger.Warn(msg, args...)
	}
}

// Error logs at LevelError.
func (l *Logger) Error(msg string, args ...any) {
	for _, logger := range l.ls {
		logger.Error(msg, args...)
	}
}

func (l *Logger) Close() error {
	return l.logfile.Close()
}

func newLogger(logfile string, loglevel string) (*Logger, error) {
	var lv slog.LevelVar
	if err := lv.UnmarshalText([]byte(loglevel)); err != nil {
		return nil, err
	}

	slog.Default()

	if err := os.MkdirAll(filepath.Dir(logfile), 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		return nil, err
	}
	return &Logger{
		ls: []*slog.Logger{
			slog.Default(),
			slog.New(slog.NewTextHandler(file, &slog.HandlerOptions{
				AddSource:   false,
				Level:       &lv,
				ReplaceAttr: nil,
			})),
		},
		logfile: file,
	}, nil
}
