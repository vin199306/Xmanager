// Package utils provides utility functions for the program manager
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"program-manager/models"
)

// ProcessManager handles process management for programs
type ProcessManager struct {
	logDir string
}

// NewProcessManager creates a new ProcessManager instance
func NewProcessManager(logDir string) *ProcessManager {
	return &ProcessManager{
		logDir: logDir,
	}
}

// StartProcess starts a new process with the given command and working directory
// Returns the process ID and any error encountered
func (pm *ProcessManager) StartProcess(command, workingDir string) (int, error) {
	// Parse the command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return 0, fmt.Errorf("empty command")
	}

	cmdName := parts[0]
	var cmdArgs []string
	if len(parts) > 1 {
		cmdArgs = parts[1:]
	}

	// Create log files
	timestamp := time.Now().Format("20060102_150405")
	logFileName := fmt.Sprintf("process_%s.log", timestamp)
	logPath := filepath.Join(pm.logDir, logFileName)

	logFile, err := os.Create(logPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create log file: %v", err)
	}
	defer logFile.Close()

	// Create the command
	cmd := exec.Command(cmdName, cmdArgs...)
	
	// Set working directory
	if workingDir != "" {
		if _, err := os.Stat(workingDir); err == nil {
			cmd.Dir = workingDir
		}
	}

	// Set output to log file
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Ensure the process runs in a new process group to prevent
	// it from being terminated when the parent exits
	setProcessGroup(cmd)

	// Start the process
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %v", err)
	}

	// Wait a moment to ensure the process is actually running
	time.Sleep(100 * time.Millisecond)
	
	// Verify the process is actually running
	if !pm.IsProcessRunning(cmd.Process.Pid) {
		return 0, fmt.Errorf("process started but failed to remain running")
	}

	return cmd.Process.Pid, nil
}

// StopProcess stops a process by its PID with enhanced safety checks
func (pm *ProcessManager) StopProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid process ID")
	}

	// Check if process exists before attempting to kill
	if !pm.IsProcessRunning(pid) {
		return fmt.Errorf("process with PID %d not found", pid)
	}

	return killProcess(pid)
}

// IsProcessRunning checks if a process with the given PID is running
func (pm *ProcessManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	return isProcessRunningOS(pid)
}

// GetProcessStatus returns the status of a process
func (pm *ProcessManager) GetProcessStatus(pid int) string {
	if pm.IsProcessRunning(pid) {
		return "running"
	}
	return "stopped"
}

// StartProgram starts a program with the given configuration
func (pm *ProcessManager) StartProgram(program *models.Program) error {
	if program.Command == "" {
		return fmt.Errorf("program command is required")
	}

	// Ensure log directory exists
	if err := os.MkdirAll(pm.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Parse command
	parts := strings.Fields(program.Command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmdName := parts[0]
	var cmdArgs []string
	if len(parts) > 1 {
		cmdArgs = parts[1:]
	}

	// Create log files
	timestamp := time.Now().Format("20060102_150405")
	logFileName := fmt.Sprintf("program_%s_%s.log", program.Name, timestamp)
	logPath := filepath.Join(pm.logDir, logFileName)

	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	// Create the command
	cmd := exec.Command(cmdName, cmdArgs...)
	
	// Set working directory
	if program.WorkingDir != "" {
		if _, err := os.Stat(program.WorkingDir); err == nil {
			cmd.Dir = program.WorkingDir
		}
	}

	// Set output to log file
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Ensure the process runs in a new process group
	setProcessGroup(cmd)

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Wait a moment to ensure the process is actually running
	time.Sleep(500 * time.Millisecond)
	
	// Verify the process is actually running
	if !pm.IsProcessRunning(cmd.Process.Pid) {
		return fmt.Errorf("process started but failed to remain running")
	}

	// Update program status
	program.Status = "running"
	program.PID = cmd.Process.Pid
	
	return nil
}

// StopProgram stops a running program
func (pm *ProcessManager) StopProgram(program *models.Program) error {
	if program.PID <= 0 {
		program.Status = "stopped"
		program.PID = 0
		return nil
	}

	// Check if process is actually running
	if !pm.IsProcessRunning(program.PID) {
		program.Status = "stopped"
		program.PID = 0
		return nil
	}

	// Stop the process
	if err := pm.StopProcess(program.PID); err != nil {
		return fmt.Errorf("failed to stop program: %w", err)
	}

	// Update program status
	program.Status = "stopped"
	program.PID = 0

	return nil
}

// GetProgramStatus returns the actual status of a program by checking process
func (pm *ProcessManager) GetProgramStatus(program *models.Program) (string, int, error) {
	if program.PID <= 0 {
		return "stopped", 0, nil
	}

	// Check if process is running
	if !pm.IsProcessRunning(program.PID) {
		return "stopped", 0, nil
	}

	// Validate the process is actually the expected program
	if !validateProcess(program.PID, program.Command) {
		return "stopped", 0, nil
	}

	return "running", program.PID, nil
}



// ProcessInfo contains basic information about a process
type ProcessInfo struct {
	Name      string
	PID       int
	MemoryKB  int64
	MemoryMB  float64
}

// GetProcessMemory gets memory usage for a specific process
func (pm *ProcessManager) GetProcessMemory(pid int) (int64, float64, error) {
	return getProcessMemoryOS(pid)
}

// GetSystemMemoryInfo returns system memory information
func (pm *ProcessManager) GetSystemMemoryInfo() (int64, int64, int64, float64, error) {
	return getSystemMemoryInfoOS()
}

// GetProcessMemoryInfo returns process memory information
func (pm *ProcessManager) GetProcessMemoryInfo(pid int) map[string]interface{} {
	if pid <= 0 {
		return map[string]interface{}{
			"memory_kb": int64(0),
			"memory_mb": 0.0,
			"error":     "invalid PID",
		}
	}

	memoryKB, memoryMB, err := getProcessMemoryOS(pid)
	if err != nil {
		return map[string]interface{}{
			"memory_kb": int64(0),
			"memory_mb": 0.0,
			"error":     err.Error(),
		}
	}

	return map[string]interface{}{
		"memory_kb": memoryKB,
		"memory_mb": memoryMB,
	}
}