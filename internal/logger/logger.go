// Package logger provides structured logging with file output and in-memory storage
package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents log severity levels
type Level string

const (
	DEBUG Level = "DEBUG"
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
)

// Entry represents a single log entry
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Logger provides thread-safe logging to file and in-memory storage
type Logger struct {
	mu       sync.RWMutex
	file     *os.File
	logger   *log.Logger
	entries  []Entry
	maxSize  int
	minLevel Level
}

// New creates a new logger that writes to the specified file
func New(filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:     file,
		logger:   log.New(file, "", 0), // No prefix, we handle formatting
		entries:  make([]Entry, 0, 1000),
		maxSize:  1000, // Keep last 1000 entries in memory
		minLevel: INFO,
	}, nil
}

// log writes a message to file and stores it in memory if level is enabled
func (l *Logger) log(level Level, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now()

	// Create entry for in-memory storage
	entry := Entry{
		Timestamp: timestamp,
		Level:     string(level),
		Message:   msg,
	}

	// Thread-safe write to file
	l.mu.Lock()
	// Write formatted line to file
	logLine := fmt.Sprintf("[%s] [%s] %s\n",
		timestamp.Format("2006-01-02 15:04:05"),
		level,
		msg)
	l.logger.Print(logLine)

	// Store in memory (circular buffer)
	if len(l.entries) >= l.maxSize {
		l.entries = l.entries[1:] // Remove oldest entry
	}
	l.entries = append(l.entries, entry)
	l.mu.Unlock()
}

// shouldLog checks if the given level should be logged based on minimum level
func (l *Logger) shouldLog(level Level) bool {
	levels := map[Level]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[l.minLevel]
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// GetEntries returns a copy of all log entries in memory
func (l *Logger) GetEntries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Return a copy to prevent race conditions
	entries := make([]Entry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

// SetLevel changes the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}