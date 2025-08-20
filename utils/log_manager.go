package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogManager provides centralized log management
type LogManager struct {
	mu       sync.RWMutex
	logsDir  string
	maxSize  int64 // 最大日志文件大小（字节）
	maxFiles int   // 最大日志文件数量
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	ProgramID string    `json:"program_id,omitempty"`
	Source    string    `json:"source,omitempty"`
}

// ProgramLog represents program-specific logs
var LogMgr = NewLogManager("data/logs")

// NewLogManager creates a new log manager
func NewLogManager(logsDir string) *LogManager {
	// 确保日志目录存在
	_ = os.MkdirAll(logsDir, 0755)

	return &LogManager{
		logsDir:  logsDir,
		maxSize:  10 * 1024 * 1024, // 10MB
		maxFiles: 10,
	}
}

// WriteLog writes a log entry
func (lm *LogManager) WriteLog(level string, message string, programID string, source string) error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		ProgramID: programID,
		Source:    source,
	}

	// 写入系统日志
	if err := lm.writeSystemLog(entry); err != nil {
		return fmt.Errorf("failed to write system log: %v", err)
	}

	// 写入程序特定日志
	if programID != "" {
		if err := lm.writeProgramLog(programID, entry); err != nil {
			return fmt.Errorf("failed to write program log: %v", err)
		}
	}

	return nil
}

// GetProgramLogs retrieves logs for a specific program
func (lm *LogManager) GetProgramLogs(programID string, limit int) ([]LogEntry, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	logFile := filepath.Join(lm.logsDir, fmt.Sprintf("%s.log", programID))
	
	// 检查文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return []LogEntry{}, nil
	}

	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	// 读取日志文件
	var entries []LogEntry
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		entries = append(entries, entry)
	}

	// 限制返回的条目数量
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	return entries, nil
}

// GetSystemLogs retrieves system logs
func (lm *LogManager) GetSystemLogs(limit int) ([]LogEntry, error) {
	lm.mu.RLock()
	defer lm.mu.RUnlock()

	systemLog := filepath.Join(lm.logsDir, "system.log")
	
	// 检查文件是否存在
	if _, err := os.Stat(systemLog); os.IsNotExist(err) {
		return []LogEntry{}, nil
	}

	file, err := os.Open(systemLog)
	if err != nil {
		return nil, fmt.Errorf("failed to open system log: %v", err)
	}
	defer file.Close()

	// 读取系统日志
	var entries []LogEntry
	decoder := json.NewDecoder(file)
	
	for decoder.More() {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		entries = append(entries, entry)
	}

	// 限制返回的条目数量
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	return entries, nil
}

// CleanOldLogs cleans up old log files
func (lm *LogManager) CleanOldLogs() error {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	files, err := filepath.Glob(filepath.Join(lm.logsDir, "*.log"))
	if err != nil {
		return fmt.Errorf("failed to list log files: %v", err)
	}

	// 按修改时间排序并删除旧文件
	// 这里简化处理：只保留最近的maxFiles个文件
	if len(files) > lm.maxFiles {
		// 在实际应用中，应该按文件大小和修改时间排序
		for i := 0; i < len(files)-lm.maxFiles; i++ {
			_ = os.Remove(files[i])
		}
	}

	return nil
}

// writeSystemLog writes to the system log file
func (lm *LogManager) writeSystemLog(entry LogEntry) error {
	systemLog := filepath.Join(lm.logsDir, "system.log")
	return lm.writeLogEntry(systemLog, entry)
}

// writeProgramLog writes to a program-specific log file
func (lm *LogManager) writeProgramLog(programID string, entry LogEntry) error {
	programLog := filepath.Join(lm.logsDir, fmt.Sprintf("%s.log", programID))
	return lm.writeLogEntry(programLog, entry)
}

// writeLogEntry writes a log entry to a specific file
func (lm *LogManager) writeLogEntry(filePath string, entry LogEntry) error {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}
	defer file.Close()

	// 检查文件大小
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 如果文件太大，进行轮转
	if info.Size() > lm.maxSize {
		if err := lm.rotateLog(filePath); err != nil {
			return fmt.Errorf("failed to rotate log: %v", err)
		}
		// 重新打开文件
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to reopen log file: %v", err)
		}
	}

	// 写入日志条目
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(entry); err != nil {
		return fmt.Errorf("failed to encode log entry: %v", err)
	}

	return nil
}

// rotateLog rotates a log file
func (lm *LogManager) rotateLog(filePath string) error {
	// 创建备份文件名
	backupPath := fmt.Sprintf("%s.%d", filePath, time.Now().Unix())
	
	// 重命名当前文件
	if err := os.Rename(filePath, backupPath); err != nil {
		return fmt.Errorf("failed to rename log file: %v", err)
	}

	// 压缩备份文件（可选）
	// 这里简化处理

	return nil
}

// Log writes an info log
func Log(message string, programID string) error {
	return LogMgr.WriteLog("INFO", message, programID, "system")
}

// LogErrorSimple writes an error log (renamed to avoid conflict)
func LogErrorSimple(message string, programID string) error {
	return LogMgr.WriteLog("ERROR", message, programID, "system")
}

// LogWarning writes a warning log
func LogWarning(message string, programID string) error {
	return LogMgr.WriteLog("WARNING", message, programID, "system")
}

// LogDebug writes a debug log
func LogDebug(message string, programID string) error {
	return LogMgr.WriteLog("DEBUG", message, programID, "system")
}

// LogOperation writes an operation log
func LogOperation(source string, message string, programID string, context map[string]interface{}) {
	fullMessage := message
	if len(context) > 0 {
		contextJSON, _ := json.Marshal(context)
		fullMessage = fmt.Sprintf("%s | Context: %s", message, string(contextJSON))
	}
	_ = LogMgr.WriteLog("INFO", fullMessage, programID, source)
}

// LogError writes an error log
func LogError(source string, message string, programID string, err error, context map[string]interface{}) {
	fullMessage := message
	if err != nil {
		fullMessage = fmt.Sprintf("%s: %v", message, err)
	}
	if len(context) > 0 {
		contextJSON, _ := json.Marshal(context)
		fullMessage = fmt.Sprintf("%s | Context: %s", fullMessage, string(contextJSON))
	}
	_ = LogMgr.WriteLog("ERROR", fullMessage, programID, source)
}

// LogRequest logs HTTP request information
func LogRequest(method string, path string, clientIP string, statusCode int, duration time.Duration, context map[string]interface{}) {
	message := fmt.Sprintf("HTTP %s %s - %d (%v)", method, path, statusCode, duration)
	_ = LogMgr.WriteLog("INFO", message, "", "http")
}