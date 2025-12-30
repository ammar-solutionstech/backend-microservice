package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger provides structured logging
type Logger struct {
	level    LogLevel
	logFile  *os.File
	useJSON  bool
	service  string
}

// NewLogger creates a new logger instance
func NewLogger(service, logDir, logFile, level, format string) (*Logger, error) {
	logger := &Logger{
		service: service,
		useJSON: format == "json",
	}

	// Parse log level
	switch level {
	case "debug":
		logger.level = LogLevelDebug
	case "info":
		logger.level = LogLevelInfo
	case "warn":
		logger.level = LogLevelWarn
	case "error":
		logger.level = LogLevelError
	default:
		logger.level = LogLevelInfo
	}

	// Open log file if specified
	if logFile != "" {
		logPath := filepath.Join(logDir, logFile)
		file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %v", err)
		}
		logger.logFile = file
	}

	return logger, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	if l.level <= LogLevelDebug {
		l.log("DEBUG", msg, fields...)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	if l.level <= LogLevelInfo {
		l.log("INFO", msg, fields...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	if l.level <= LogLevelWarn {
		l.log("WARN", msg, fields...)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	if l.level <= LogLevelError {
		allFields := append(fields, map[string]interface{}{"error": err.Error()})
		l.log("ERROR", msg, allFields...)
	}
}

func (l *Logger) log(level, msg string, fields ...map[string]interface{}) {
	timestamp := time.Now().Format(time.RFC3339)
	
	if l.useJSON {
		entry := map[string]interface{}{
			"timestamp": timestamp,
			"level":     level,
			"service":   l.service,
			"message":   msg,
		}
		
		// Merge fields
		for _, f := range fields {
			for k, v := range f {
				entry[k] = v
			}
		}
		
		jsonData, _ := json.Marshal(entry)
		output := string(jsonData) + "\n"
		
		if l.logFile != nil {
			l.logFile.WriteString(output)
		} else {
			log.Print(output)
		}
	} else {
		// Plain text format
		output := fmt.Sprintf("[%s] [%s] [%s] %s", timestamp, level, l.service, msg)
		if len(fields) > 0 {
			for _, f := range fields {
				for k, v := range f {
					output += fmt.Sprintf(" %s=%v", k, v)
				}
			}
		}
		output += "\n"
		
		if l.logFile != nil {
			l.logFile.WriteString(output)
		} else {
			log.Print(output)
		}
	}
}

