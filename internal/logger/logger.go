package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type Level string

const (
	DEBUG Level = "DEBUG"
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
)

type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type Logger struct {
	mu       sync.RWMutex
	file     *os.File
	logger   *log.Logger
	entries  []Entry
	maxSize  int
	minLevel Level
}

func New(filename string) (*Logger, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &Logger{
		file:     file,
		logger:   log.New(file, "", 0),
		entries:  make([]Entry, 0, 1000),
		maxSize:  1000,
		minLevel: INFO,
	}, nil
}

func (l *Logger) log(level Level, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now()

	// Create entry
	entry := Entry{
		Timestamp: timestamp,
		Level:     string(level),
		Message:   msg,
	}

	// Thread-safe write to file
	l.mu.Lock()
	logLine := fmt.Sprintf("[%s] [%s] %s\n",
		timestamp.Format("2006-01-02 15:04:05"),
		level,
		msg)
	l.logger.Print(logLine)

	// Store in memory (circular buffer)
	if len(l.entries) >= l.maxSize {
		l.entries = l.entries[1:]
	}
	l.entries = append(l.entries, entry)
	l.mu.Unlock()
}

func (l *Logger) shouldLog(level Level) bool {
	levels := map[Level]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
	}
	return levels[level] >= levels[l.minLevel]
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

func (l *Logger) GetEntries() []Entry {
	l.mu.RLock()
	defer l.mu.RUnlock()
	
	// Return a copy to prevent race conditions
	entries := make([]Entry, len(l.entries))
	copy(entries, l.entries)
	return entries
}

func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}