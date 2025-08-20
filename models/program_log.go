package models

import (
	"fmt"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelInfo    LogLevel = "INFO"
	LogLevelWarning LogLevel = "WARNING"
	LogLevelError   LogLevel = "ERROR"
	LogLevelDebug   LogLevel = "DEBUG"
)

// ProgramLog represents a log entry for a program
type ProgramLog struct {
	ID        string    `json:"id"`
	ProgramID string    `json:"program_id"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ProgramLogEntry represents a detailed log entry for API responses
type ProgramLogEntry struct {
	ID        string    `json:"id"`
	ProgramID string    `json:"program_id"`
	ProgramName string  `json:"program_name"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// NewProgramLog creates a new program log entry
func NewProgramLog(programID, message string, level LogLevel, details string) *ProgramLog {
	return &ProgramLog{
		ID:        generateID(),
		ProgramID: programID,
		Level:     level,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}
}

// generateID generates a unique ID for log entries
func generateID() string {
	timestamp := time.Now().Format("20060102150405")
	nano := time.Now().UnixNano()
	nanoStr := fmt.Sprintf("%06d", nano%1000000) // 确保至少6位数字
	return timestamp + "-" + nanoStr
}