package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Use a buffered writer for better performance
var stdoutWriter = zapcore.AddSync(os.Stdout)

var (
	logger   *zap.Logger
	detailed bool
	logFile  *os.File
)

// InitLogger initializes the logger
func InitLogger() {
	InitLoggerWithFile("")
}

// InitLoggerWithFile initializes the logger with optional file output
func InitLoggerWithFile(logFilePath string) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "", // Disable caller
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("15:04:05.000"),
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	// Always write to stdout
	cores := []zapcore.Core{
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			stdoutWriter,
			zapcore.InfoLevel,
		),
	}

	// If log file is specified, also write to file
	if logFilePath != "" {
		// If path doesn't contain directory separator, put it in logs/ directory
		if !filepath.IsAbs(logFilePath) && filepath.Dir(logFilePath) == "." {
			logFilePath = filepath.Join("logs", logFilePath)
		}

		// Create logs directory if it doesn't exist
		dir := filepath.Dir(logFilePath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				os.Stderr.WriteString("Warning: Failed to create log directory: " + err.Error() + "\n")
			}
		}

		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			// If we can't open the file, log to stderr and continue with stdout only
			os.Stderr.WriteString("Warning: Failed to open log file: " + err.Error() + "\n")
		} else {
			// File encoder - plain text format for readability
			fileEncoderConfig := zapcore.EncoderConfig{
				TimeKey:        "time",
				LevelKey:       "level",
				MessageKey:     "msg",
				CallerKey:      "",                                                     // Disable caller
				EncodeLevel:    zapcore.LowercaseLevelEncoder,                          // No colors in file
				EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"), // Full date for files
				EncodeDuration: zapcore.StringDurationEncoder,
			}
			cores = append(cores, zapcore.NewCore(
				zapcore.NewConsoleEncoder(fileEncoderConfig), // Plain text console format
				zapcore.AddSync(logFile),
				zapcore.InfoLevel,
			))
		}
	}

	// Create logger with multiple cores (stdout + optional file)
	logger = zap.New(zapcore.NewTee(cores...))
}

// SetDetailed sets the detailed logging mode
func SetDetailed(d bool) {
	detailed = d
}

// IsDetailed returns whether detailed logging is enabled
func IsDetailed() bool {
	return detailed
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// Sync flushes any buffered log entries
func Sync() error {
	var err error
	if logger != nil {
		err = logger.Sync()
	}
	if logFile != nil {
		if syncErr := logFile.Sync(); syncErr != nil && err == nil {
			err = syncErr
		}
		if closeErr := logFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	return err
}
