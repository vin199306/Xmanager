//go:build linux

// Package utils provides Linux-specific process management utilities
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// killProcess kills a process and its children on Linux
func killProcess(pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid process ID")
	}

	// Get process group ID for the process
	pgid, err := getProcessGroupID(pid)
	if err != nil {
		// If we can't get the process group, just kill the individual process
		return killIndividualProcess(pid)
	}

	// Kill the entire process group
	if err := syscall.Kill(-pgid, syscall.SIGTERM); err != nil {
		// Fallback to individual process kill
		return killIndividualProcess(pid)
	}

	// Wait for the process to terminate
	for i := 0; i < 30; i++ {
		if !isProcessRunningOS(pid) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	// If SIGTERM didn't work, try SIGKILL
	if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil {
		return killIndividualProcess(pid)
	}

	// Wait for the process to terminate
	for i := 0; i < 20; i++ {
		if !isProcessRunningOS(pid) {
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Try to kill any remaining child processes
	killChildProcesses(pgid)

	return fmt.Errorf("process %d did not terminate after kill signal", pid)
}

// killIndividualProcess kills a single process
func killIndividualProcess(pid int) error {
	// Try SIGTERM first
	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send SIGTERM: %v", err)
	}

	// Wait for the process to terminate
	for i := 0; i < 20; i++ {
		if !isProcessRunningOS(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	// If SIGTERM didn't work, try SIGKILL
	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		return fmt.Errorf("failed to send SIGKILL: %v", err)
	}

	// Wait for the process to terminate
	for i := 0; i < 10; i++ {
		if !isProcessRunningOS(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("process %d did not terminate after kill signal", pid)
}

// getProcessGroupID gets the process group ID for a given PID
func getProcessGroupID(pid int) (int, error) {
	pgidFile := fmt.Sprintf("/proc/%d/stat", pid)
	data, err := os.ReadFile(pgidFile)
	if err != nil {
		return 0, err
	}

	// Parse the stat file to get the PGID (4th field)
	fields := strings.Fields(string(data))
	if len(fields) < 5 {
		return 0, fmt.Errorf("invalid stat file format")
	}

	pgid, err := strconv.Atoi(fields[4])
	if err != nil {
		return 0, fmt.Errorf("invalid PGID: %v", err)
	}

	return pgid, nil
}

// killChildProcesses kills all child processes of a given process group
func killChildProcesses(pgid int) {
	// List all processes and find children
	procs, _ := listProcessesOS()
	for _, proc := range procs {
		if proc.PID <= 0 {
			continue
		}
		
		// Get the PPID for this process
		ppidFile := fmt.Sprintf("/proc/%d/stat", proc.PID)
		data, err := os.ReadFile(ppidFile)
		if err != nil {
			continue
		}

		fields := strings.Fields(string(data))
		if len(fields) < 4 {
			continue
		}

		ppid, err := strconv.Atoi(fields[3])
		if err != nil {
			continue
		}

		// If this is a child of our process group, kill it
		if ppid == pgid || ppid == -pgid {
			syscall.Kill(proc.PID, syscall.SIGKILL)
		}
	}
}

// isProcessRunningOS checks if a process is running on Linux
func isProcessRunningOS(pid int) bool {
	if pid <= 0 {
		return false
	}

	// Check if the process exists and is not a zombie
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return false
	}

	// Parse the status file to get the state
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "State:") {
			state := strings.TrimSpace(strings.TrimPrefix(line, "State:"))
			// Process is running if state is R (running), S (sleeping), or D (disk sleep)
			// Exclude Z (zombie) and X (dead)
			return !strings.Contains(state, "Z") && !strings.Contains(state, "X")
		}
	}

	return false
}

// validateProcess checks if the process matches the expected command
func validateProcess(pid int, expectedCommand string) bool {
	if pid <= 0 {
		return false
	}

	// Read the command line from /proc/pid/cmdline
	cmdlineFile := fmt.Sprintf("/proc/%d/cmdline", pid)
	data, err := os.ReadFile(cmdlineFile)
	if err != nil {
		return false
	}

	// cmdline contains null-separated arguments, convert to space-separated
	cmdline := strings.ReplaceAll(string(data), "\x00", " ")
	cmdline = strings.TrimSpace(cmdline)

	if cmdline == "" {
		return false
	}

	// Compare with expected command
	expectedParts := strings.Fields(expectedCommand)
	if len(expectedParts) == 0 {
		return false
	}

	// Check if the actual command contains the expected command
	return strings.Contains(cmdline, expectedParts[0])
}

// listProcessesOS lists all running processes on Linux
func listProcessesOS() ([]ProcessInfo, error) {
	procs := make([]ProcessInfo, 0)
	
	// Read /proc directory
	procDir, err := os.Open("/proc")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc: %v", err)
	}
	defer procDir.Close()

	// Read all entries
	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc entries: %v", err)
	}

	// Process each entry
	for _, entry := range entries {
		// Skip non-numeric entries
		pid, err := strconv.Atoi(entry)
		if err != nil {
			continue
		}

		// Get process info
		info, err := getProcessInfo(pid)
		if err != nil {
			continue
		}

		procs = append(procs, info)
	}

	return procs, nil
}

// getProcessInfo gets information about a specific process
func getProcessInfo(pid int) (ProcessInfo, error) {
	info := ProcessInfo{PID: pid}

	// Read cmdline to get the process name
	cmdlineFile := fmt.Sprintf("/proc/%d/cmdline", pid)
		data, err := os.ReadFile(cmdlineFile)
		if err != nil {
			return info, err
		}

	// Get the executable name
	cmdline := strings.ReplaceAll(string(data), "\x00", " ")
	parts := strings.Fields(cmdline)
	if len(parts) > 0 {
		info.Name = filepath.Base(parts[0])
	}

	// Read memory info from status file
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
		statusData, err := os.ReadFile(statusFile)
		if err != nil {
			// 如果status文件读取失败，尝试从statm获取内存信息
			memoryKB, memoryMB, _ := getProcessMemoryFromStatm(pid)
			info.MemoryKB = memoryKB
			info.MemoryMB = memoryMB
			return info, nil
		}

	// Parse VmRSS for memory usage
	lines := strings.Split(string(statusData), "\n")
	foundMemory := false
	for _, line := range lines {
		if strings.HasPrefix(line, "VmRSS:") {
			var vmRSS string
			fmt.Sscanf(line, "VmRSS: %s", &vmRSS)
			if strings.HasSuffix(vmRSS, "kB") {
				vmRSS = strings.TrimSuffix(vmRSS, "kB")
				vmRSS = strings.TrimSpace(vmRSS)
				if kb, err := strconv.ParseInt(vmRSS, 10, 64); err == nil {
					info.MemoryKB = kb
					info.MemoryMB = float64(kb) / 1024.0
					foundMemory = true
				}
			}
			break
		}
	}

	// 如果VmRSS没有找到，尝试从statm获取
	if !foundMemory {
		memoryKB, memoryMB, _ := getProcessMemoryFromStatm(pid)
		info.MemoryKB = memoryKB
		info.MemoryMB = memoryMB
	}

	return info, nil
}

// getProcessMemoryOS gets memory usage for a specific process on Linux
func getProcessMemoryOS(pid int) (int64, float64, error) {
	statusFile := fmt.Sprintf("/proc/%d/status", pid)
	data, err := os.ReadFile(statusFile)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read status file: %v", err)
	}

	var memoryKB int64
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "VmRSS:") {
			var vmRSS string
			fmt.Sscanf(line, "VmRSS: %s", &vmRSS)
			if strings.HasSuffix(vmRSS, "kB") {
				vmRSS = strings.TrimSuffix(vmRSS, "kB")
				vmRSS = strings.TrimSpace(vmRSS)
				if kb, err := strconv.ParseInt(vmRSS, 10, 64); err == nil {
					memoryKB = kb
				} else {
					// 如果解析失败，尝试从/proc/[pid]/statm获取内存信息
					return getProcessMemoryFromStatm(pid)
				}
			}
			break
		}
	}

	// 如果VmRSS没有找到，尝试从statm获取
	if memoryKB == 0 {
		return getProcessMemoryFromStatm(pid)
	}

	memoryMB := float64(memoryKB) / 1024.0
	return memoryKB, memoryMB, nil
}

// getProcessMemoryFromStatm gets memory usage from /proc/[pid]/statm as fallback
func getProcessMemoryFromStatm(pid int) (int64, float64, error) {
	statmFile := fmt.Sprintf("/proc/%d/statm", pid)
	data, err := os.ReadFile(statmFile)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read statm file: %v", err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("invalid statm file format")
	}

	// 第二个字段是常驻内存页数
	residentPages, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse resident pages: %v", err)
	}

	// 通常一页是4KB
	memoryKB := residentPages * 4
	memoryMB := float64(memoryKB) / 1024.0
	return memoryKB, memoryMB, nil
}

// getSystemMemoryInfoOS gets system memory information on Linux
func getSystemMemoryInfoOS() (totalPhysical int64, available int64, used int64, usedPercent float64, err error) {
	meminfoFile := "/proc/meminfo"
	data, err := os.ReadFile(meminfoFile)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to read meminfo: %v", err)
	}

	var total, free, buffers, cached int64
	lines := strings.Split(string(data), "\n")
	
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		value, _ := strconv.ParseInt(fields[1], 10, 64)
		
		switch fields[0] {
		case "MemTotal:":
			total = value
		case "MemFree:":
			free = value
		case "Buffers:":
			buffers = value
		case "Cached:":
			cached = value
		}
	}

	if total == 0 {
		return 0, 0, 0, 0, fmt.Errorf("failed to parse memory information")
	}

	// Calculate available memory (free + buffers + cached)
	available = free + buffers + cached
	used = total - available
	usedPercent = (float64(used) / float64(total)) * 100.0

	return total, available, used, usedPercent, nil
}

// setProcessGroup sets the process group ID for the command
func setProcessGroup(cmd *exec.Cmd) {
	// On Linux, we use Setpgid to put the process in a new process group
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}
}