package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"program-manager/models"
	"strings"
	"time"
)

// ProcessManager handles process management for programs
type ProcessManager struct {
	logDir string
}

// NewProcessManager creates a new ProcessManager instance
func NewProcessManager(logDir ...string) *ProcessManager {
	pm := &ProcessManager{}
	if len(logDir) > 0 {
		pm.logDir = logDir[0]
	} else {
		pm.logDir = "logs"
	}
	return pm
}

// StartProgram starts a new program with enhanced process management
func (pm *ProcessManager) StartProgram(program *models.Program) error {
	if program.Status == "running" {
		return fmt.Errorf("program is already running")
	}

	if program.Command == "" {
		return fmt.Errorf("program command is required")
	}

	// Ensure log directory exists
	if err := os.MkdirAll(pm.logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log files
	logFileName := fmt.Sprintf("%s_%d.log", program.ID, time.Now().Unix())
	stdoutFile, err := os.Create(filepath.Join(pm.logDir, logFileName))
	if err != nil {
		return fmt.Errorf("failed to create stdout log: %w", err)
	}
	defer stdoutFile.Close()

	stderrFile, err := os.Create(filepath.Join(pm.logDir, logFileName+".err"))
	if err != nil {
		return fmt.Errorf("failed to create stderr log: %w", err)
	}
	defer stderrFile.Close()

	// Use direct command execution for Linux background processes
	parts := strings.Fields(program.Command)
	cmdName := parts[0]
	var cmdArgs []string
	if len(parts) > 1 {
		cmdArgs = parts[1:]
	}
	cmd := exec.Command(cmdName, cmdArgs...)
	
	// Set working directory
	if program.WorkingDir != "" {
		cmd.Dir = program.WorkingDir
	}

	// Set up command I/O
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile

	// Start the process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Wait and verify the process is actually running
	time.Sleep(500 * time.Millisecond)
	
	// Verify the process is running
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
	if program.Status != "running" || program.PID == 0 {
		// 如果程序标记为未运行，直接更新状态
		program.Status = "stopped"
		program.PID = 0
		return nil
	}

	// 首先检查进程是否实际存在
	if !pm.IsProcessRunning(program.PID) {
		// 进程已经不存在，更新状态
		program.Status = "stopped"
		program.PID = 0
		return nil
	}

	// 尝试优雅地结束进程
	var err error
	
	// Linux-specific process termination
	cmd := exec.Command("kill", "-TERM", fmt.Sprintf("%d", program.PID))
	err = cmd.Run()
	
	// 如果优雅关闭失败，使用强制关闭
	if err != nil {
		cmd = exec.Command("kill", "-KILL", fmt.Sprintf("%d", program.PID))
		err = cmd.Run()
		if err != nil {
			// 尝试查找并杀死子进程
			childCmd := exec.Command("pgrep", "-P", fmt.Sprintf("%d", program.PID))
			if output, childErr := childCmd.Output(); childErr == nil {
				lines := strings.Split(string(output), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" {
						childPID := strings.Trim(line, " ")
						if childPID != "" {
							childKillCmd := exec.Command("kill", "-KILL", childPID)
							childKillCmd.Run()
						}
					}
				}
			}
			
			// 最后再次尝试强制杀死主进程
			cmd = exec.Command("kill", "-KILL", fmt.Sprintf("%d", program.PID))
			_ = cmd.Run() // 忽略错误，进程可能已经退出
		}
	}

	// 等待一小段时间确保进程已退出
	for i := 0; i < 5; i++ {
		if !pm.IsProcessRunning(program.PID) {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 更新程序状态
	program.Status = "stopped"
	program.PID = 0
	
	return nil
}

// GetProgramStatus checks if a program is actually running
func (pm *ProcessManager) GetProgramStatus(program *models.Program) (string, int, error) {
	if program.PID == 0 {
		return "stopped", 0, nil
	}

	actualStatus := pm.GetProcessStatus(program.PID)
	if actualStatus == "running" {
		return "running", program.PID, nil
	}
	
	return "stopped", 0, nil
}

// GetAllProcesses returns a list of all running processes
func (pm *ProcessManager) GetAllProcesses() ([]models.ProcessInfo, error) {
	cmd := exec.Command("tasklist", "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get process list: %w", err)
	}

	var processes []models.ProcessInfo
	lines := strings.Split(string(output), "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\",\"")
		if len(parts) >= 2 {
			name := strings.Trim(parts[0], "\"")
			pidStr := strings.Trim(parts[1], "\"")
			
			var pid int
			fmt.Sscanf(pidStr, "%d", &pid)
			
			processes = append(processes, models.ProcessInfo{
				PID:     pid,
				Name:    name,
				Command: name,
			})
		}
	}

	return processes, nil
}

// ExecuteCommand executes a command and returns the output
func (pm *ProcessManager) ExecuteCommand(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	output, err := cmd.CombinedOutput()
	
	return string(output), err
}

// GetProcessStatus checks if a process is actually running (alias for GetProgramStatus)
func (pm *ProcessManager) GetProcessStatus(pid int) string {
	if pid == 0 {
		return "stopped"
	}

	// Check if process exists using kill -0 (signal 0 test)
	cmd := exec.Command("kill", "-0", fmt.Sprintf("%d", pid))
	_, err := cmd.CombinedOutput()
	if err == nil {
		return "running"
	}
	return "stopped"
}

// IsProcessRunning checks if a process with given PID is running
func (pm *ProcessManager) IsProcessRunning(pid int) bool {
	if pid == 0 {
		return false
	}
	
	status := pm.GetProcessStatus(pid)
	return status == "running"
}

// GetProcessMemory gets memory usage for a specific process
func (pm *ProcessManager) GetProcessMemory(pid int) (int64, float64, error) {
	if pid == 0 {
		return 0, 0, fmt.Errorf("invalid PID")
	}
	
	// Use PowerShell to get memory usage
	cmd := exec.Command("powershell", "-Command", 
		fmt.Sprintf("Get-Process -Id %d | Select-Object -ExpandProperty WorkingSet", pid))
	
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get memory info: %v", err)
	}
	
	var workingSet int64
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &workingSet)
	
	memoryKB := workingSet / 1024
	memoryMB := float64(workingSet) / (1024 * 1024)
	
	return memoryKB, memoryMB, nil
}

// GetSystemMemoryInfo returns system memory information
func (pm *ProcessManager) GetSystemMemoryInfo() (uint64, uint64, uint64, float64, error) {
	// Use free to get system memory info on Linux
	cmd := exec.Command("free", "-b")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to get system memory: %v", err)
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				var total, used, free uint64
				fmt.Sscanf(fields[1], "%d", &total)
				fmt.Sscanf(fields[2], "%d", &used)
				fmt.Sscanf(fields[3], "%d", &free)
				usedPercent := float64(used) / float64(total) * 100
				return total, free, used, usedPercent, nil
			}
		}
	}
	
	return 0, 0, 0, 0, fmt.Errorf("invalid memory info format")
}