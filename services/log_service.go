package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"program-manager/models"
	"sort"
	"sync"
	"time"
)

// LogService handles logging for program operations
type LogService struct {
	logsDir string
	mutex   sync.RWMutex
}

// NewLogService creates a new LogService
func NewLogService(logsDir string) *LogService {
	return &LogService{
		logsDir: logsDir,
	}
}

// Log logs a message for a specific program
func (ls *LogService) Log(programID, programName, message string, level models.LogLevel, details string) error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	logEntry := &models.ProgramLog{
		ID:        fmt.Sprintf("log_%d_%s", time.Now().UnixNano(), programID),
		ProgramID: programID,
		Level:     level,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
	}

	return ls.appendLog(logEntry)
}

// readAllLogs reads all log entries from the log file
func (ls *LogService) readAllLogs() ([]*models.ProgramLog, error) {
	logFile := filepath.Join(ls.logsDir, "program_logs.json")

	var logs []*models.ProgramLog
	
	// 检查文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return logs, nil
	}

	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read logs file: %w", err)
	}

	if len(data) == 0 {
		return logs, nil
	}

	if err := json.Unmarshal(data, &logs); err != nil {
		// 如果解析失败，记录错误并返回空切片
		return []*models.ProgramLog{}, nil
	}

	// 过滤掉nil值
	var validLogs []*models.ProgramLog
	for _, log := range logs {
		if log != nil {
			validLogs = append(validLogs, log)
		}
	}

	return validLogs, nil
}

// GetAllLogs retrieves all logs with program names
func (ls *LogService) GetAllLogs(limit int) ([]models.ProgramLogEntry, error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	allLogs, err := ls.readAllLogs()
	if err != nil {
		return nil, err
	}

	// Get program names for logs - we need to get programs from storage
	// For now, we'll use a simple approach since we don't have access to ProgramService
	programNames := make(map[string]string)
	// We'll handle program name lookup in a different way or leave as "Unknown"

	var logEntries []models.ProgramLogEntry
	for _, log := range allLogs {
		programName := programNames[log.ProgramID]
		if programName == "" {
			programName = "Unknown Program"
		}

		entry := models.ProgramLogEntry{
			ID:          log.ID,
			ProgramID:   log.ProgramID,
			ProgramName: programName,
			Level:       log.Level,
			Message:     log.Message,
			Details:     log.Details,
			Timestamp:   log.Timestamp,
		}
		logEntries = append(logEntries, entry)
	}

	// Sort by timestamp (newest first)
	sort.Slice(logEntries, func(i, j int) bool {
		return logEntries[i].Timestamp.After(logEntries[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && len(logEntries) > limit {
		logEntries = logEntries[:limit]
	}

	return logEntries, nil
}

// ClearLogs clears all logs for a specific program
func (ls *LogService) ClearLogs(programID string) error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	allLogs, err := ls.readAllLogs()
	if err != nil {
		return err
	}

	var filteredLogs []*models.ProgramLog
	for _, log := range allLogs {
		if log.ProgramID != programID {
			filteredLogs = append(filteredLogs, log)
		}
	}

	return ls.writeAllLogs(filteredLogs)
}

// ClearAllLogs clears all logs
func (ls *LogService) ClearAllLogs() error {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()

	logFile := filepath.Join(ls.logsDir, "program_logs.json")
	return os.WriteFile(logFile, []byte("[]"), 0644)
}

// appendLog appends a log entry to the log file
func (ls *LogService) appendLog(logEntry *models.ProgramLog) error {
	logFile := filepath.Join(ls.logsDir, "program_logs.json")

	// Read existing logs
	var logs []*models.ProgramLog
	if data, err := os.ReadFile(logFile); err == nil && len(data) > 0 {
		_ = json.Unmarshal(data, &logs) // 忽略解析错误，继续处理
	}

	// Append new log
	logs = append(logs, logEntry)

	// Write back to file
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	return os.WriteFile(logFile, data, 0644)
}

// writeAllLogs writes all log entries to the log file
func (ls *LogService) writeAllLogs(logs []*models.ProgramLog) error {
	logFile := filepath.Join(ls.logsDir, "program_logs.json")

	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	return os.WriteFile(logFile, data, 0644)
}

// LogProgramStart logs when a program starts
func (ls *LogService) LogProgramStart(programID, programName, command string) error {
	message := fmt.Sprintf("程序 '%s' 已启动", programName)
	details := fmt.Sprintf("命令: %s", command)
	return ls.Log(programID, programName, message, models.LogLevelInfo, details)
}

// LogProgramStop logs when a program stops
func (ls *LogService) LogProgramStop(programID, programName string) error {
	message := fmt.Sprintf("程序 '%s' 已停止", programName)
	return ls.Log(programID, programName, message, models.LogLevelInfo, "")
}

// LogProgramError logs when a program encounters an error
func (ls *LogService) LogProgramError(programID, programName, errorMsg string) error {
	message := fmt.Sprintf("程序 '%s' 发生错误", programName)
	return ls.Log(programID, programName, message, models.LogLevelError, errorMsg)
}

// LogProgramWarning logs when a program has a warning
func (ls *LogService) LogProgramWarning(programID, programName, warningMsg string) error {
	message := fmt.Sprintf("程序 '%s' 警告", programName)
	return ls.Log(programID, programName, message, models.LogLevelWarning, warningMsg)
}

// GetLogs retrieves logs for a specific program
func (ls *LogService) GetLogs(programID string, limit int) ([]models.ProgramLog, error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()

	allLogs, err := ls.readAllLogs()
	if err != nil {
		return nil, err
	}

	var programLogs []models.ProgramLog
	for _, log := range allLogs {
		if log != nil && log.ProgramID == programID {
			programLogs = append(programLogs, *log)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(programLogs, func(i, j int) bool {
		return programLogs[i].Timestamp.After(programLogs[j].Timestamp)
	})

	// Apply limit
	if limit > 0 && len(programLogs) > limit {
		programLogs = programLogs[:limit]
	}

	return programLogs, nil
}