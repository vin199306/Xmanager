// Package services provides business logic for the program manager
package services

import (
	"fmt"
	"program-manager/models"
	"program-manager/storage"
	"program-manager/utils"
	"sync"
)

// StatusService handles program status monitoring and validation
type StatusService struct {
	storage        *storage.FileStorage
	processManager *utils.ProcessManager
	cache          map[string]*models.Program
	cacheMutex     sync.RWMutex
}

// NewStatusService creates a new StatusService instance
func NewStatusService(storage *storage.FileStorage, processManager *utils.ProcessManager) *StatusService {
	return &StatusService{
		storage:        storage,
		processManager: processManager,
		cache:          make(map[string]*models.Program),
	}
}

// GetProgramStatus gets the status of a specific program with enhanced validation
func (ss *StatusService) GetProgramStatus(id string) (*models.Program, error) {
	pm, err := ss.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}

	program, found := pm.GetProgramByID(id)
	if !found {
		return nil, fmt.Errorf("program not found: %s", id)
	}

	// Use enhanced validation to check actual status
	actualStatus, actualPID, err := ss.processManager.GetProgramStatus(&program)
	if err != nil {
		// Log error but continue with basic check
		fmt.Printf("Warning: enhanced validation failed: %v\n", err)
		actualStatus = ss.processManager.GetProcessStatus(program.PID)
		actualPID = program.PID
	}
	
	// Update status if it doesn't match
	if actualStatus != program.Status || actualPID != program.PID {
		program.Status = actualStatus
		program.PID = actualPID
		program.UpdatedAt = utils.GetCurrentTime()

		// Update the stored status
			pm.UpdateProgram(id, program)
			_ = ss.storage.SavePrograms(pm)
	}

	return &program, nil
}

// GetAllProgramsStatus gets the status of all programs with enhanced validation
func (ss *StatusService) GetAllProgramsStatus() ([]models.Program, error) {
	pm, err := ss.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}

	programs := pm.GetAllPrograms()
	var result []models.Program

	for _, program := range programs {
		// Use enhanced validation to check actual status
		actualStatus, actualPID, err := ss.processManager.GetProgramStatus(&program)
		if err != nil {
			// Continue with basic check if enhanced validation fails
			actualStatus = ss.processManager.GetProcessStatus(program.PID)
			actualPID = program.PID
		}

		// Update status if it doesn't match
		if actualStatus != program.Status || actualPID != program.PID {
			program.Status = actualStatus
			program.PID = actualPID
			program.UpdatedAt = utils.GetCurrentTime()

			// Update the stored status
			pm.UpdateProgram(program.ID, program)
			_ = ss.storage.SavePrograms(pm)
		}
		result = append(result, program)
	}

	return result, nil
}

// RefreshAllStatuses refreshes the status of all programs with enhanced validation
func (ss *StatusService) RefreshAllStatuses() error {
	pm, err := ss.storage.LoadPrograms()
	if err != nil {
		return fmt.Errorf("failed to load programs: %v", err)
	}

	// Refresh status for all programs
	updated := false
	for _, program := range pm.Programs {
		// Use enhanced validation to check actual status
		actualStatus, actualPID, err := ss.processManager.GetProgramStatus(&program)
		if err != nil {
			// Log error but continue with basic check
			fmt.Printf("Warning: enhanced validation failed: %v\n", err)
			actualStatus = ss.processManager.GetProcessStatus(program.PID)
			actualPID = program.PID
		}

		// Update status if it doesn't match
		if actualStatus != program.Status || actualPID != program.PID {
			program.Status = actualStatus
			program.PID = actualPID
			program.UpdatedAt = utils.GetCurrentTime()
			updated = true
		}

		// Update the stored status
		pm.UpdateProgram(program.ID, program)
		if err := ss.storage.SavePrograms(pm); err != nil {
			// Log error but don't fail the status check
			fmt.Printf("Warning: failed to update program status: %v\n", err)
		}
	}

	// Save updated statuses if any changed
	if updated {
		if err := ss.storage.SavePrograms(pm); err != nil {
			return fmt.Errorf("failed to save updated statuses: %v", err)
		}
	}

	return nil
}

// GetRunningPrograms returns all running programs
func (ss *StatusService) GetRunningPrograms() ([]models.Program, error) {
	pm, err := ss.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}

	programs := pm.GetAllPrograms()
	var running []models.Program

	for _, program := range programs {
		// Refresh the actual status
		actualStatus := ss.processManager.GetProcessStatus(program.PID)
		if actualStatus != program.Status {
			program.Status = actualStatus
			if actualStatus == "stopped" {
				program.PID = 0
			}

			// Update the stored status
			pm.UpdateProgram(program.ID, program)
			_ = ss.storage.SavePrograms(pm)
		}

		if program.Status == "running" {
			running = append(running, program)
		}
	}

	return running, nil
}

// GetStoppedPrograms returns all stopped programs
func (ss *StatusService) GetStoppedPrograms() ([]models.Program, error) {
	pm, err := ss.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}

	programs := pm.GetAllPrograms()
	var stopped []models.Program

	for _, program := range programs {
		// Refresh the actual status
		actualStatus := ss.processManager.GetProcessStatus(program.PID)
		if actualStatus != program.Status {
			program.Status = actualStatus
			if actualStatus == "stopped" {
				program.PID = 0
			}

			// Update the stored status
			pm.UpdateProgram(program.ID, program)
			_ = ss.storage.SavePrograms(pm)
		}

		if program.Status == "stopped" {
			stopped = append(stopped, program)
		}
	}

	return stopped, nil
}

// GetProcessInfo returns detailed process information including memory usage
func (ss *StatusService) GetProcessInfo(pid int) (map[string]interface{}, error) {
	isRunning := ss.processManager.IsProcessRunning(pid)
	
	info := map[string]interface{}{
		"pid":     pid,
		"running": isRunning,
	}

	if isRunning {
		info["status"] = "running"
		
		// Get memory usage information
		memoryKB, memoryMB, err := ss.processManager.GetProcessMemory(pid)
		if err == nil {
			info["memoryKB"] = memoryKB
			info["memoryMB"] = memoryMB
			info["memoryDisplay"] = fmt.Sprintf("%.2f MB", memoryMB)
		} else {
			info["memoryKB"] = 0
			info["memoryMB"] = 0.0
			info["memoryDisplay"] = "N/A"
		}
	} else {
		info["status"] = "stopped"
		info["memoryKB"] = 0
		info["memoryMB"] = 0.0
		info["memoryDisplay"] = "N/A"
	}

	return info, nil
}

// GetSystemMemoryInfo returns system memory information
func (ss *StatusService) GetSystemMemoryInfo() (map[string]interface{}, error) {
	totalPhysical, available, used, usedPercent, err := ss.processManager.GetSystemMemoryInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get system memory info: %v", err)
	}

	info := map[string]interface{}{
		"totalPhysical": totalPhysical,
		"totalPhysicalGB": float64(totalPhysical) / (1024 * 1024 * 1024),
		"available": available,
		"availableGB": float64(available) / (1024 * 1024 * 1024),
		"used": used,
		"usedGB": float64(used) / (1024 * 1024 * 1024),
		"usedPercent": usedPercent,
		"usedPercentDisplay": fmt.Sprintf("%.1f%%", usedPercent),
		"totalPhysicalDisplay": fmt.Sprintf("%.2f GB", float64(totalPhysical)/(1024*1024*1024)),
		"availableDisplay": fmt.Sprintf("%.2f GB", float64(available)/(1024*1024*1024)),
		"usedDisplay": fmt.Sprintf("%.2f GB", float64(used)/(1024*1024*1024)),
	}

	return info, nil
}

// GetProgramsWithMemory returns all programs with their memory usage
func (ss *StatusService) GetProgramsWithMemory() ([]map[string]interface{}, error) {
	pm, err := ss.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}

	programs := pm.GetAllPrograms()
	var result []map[string]interface{}

	for _, program := range programs {
		// Refresh the actual status
		actualStatus := ss.processManager.GetProcessStatus(program.PID)
		if actualStatus != program.Status {
			program.Status = actualStatus
			if actualStatus == "stopped" {
				program.PID = 0
			}

			// Update the stored status
			pm.UpdateProgram(program.ID, program)
			_ = ss.storage.SavePrograms(pm)
		}

		// Create program info with memory data
		programInfo := map[string]interface{}{
			"id":          program.ID,
			"name":        program.Name,
			"command":     program.Command,
			"status":      program.Status,
			"pid":         program.PID,
			"memoryUsage": "0 KB",
		}

		// Add memory info if running
		if program.Status == "running" && program.PID > 0 {
			memoryKB, memoryMB, err := ss.processManager.GetProcessMemory(program.PID)
			if err == nil && memoryKB > 0 {
				programInfo["memoryUsage"] = fmt.Sprintf("%.2f MB", memoryMB)
				programInfo["memory_kb"] = memoryKB
				programInfo["memory_mb"] = memoryMB
			} else {
				programInfo["memoryUsage"] = "0 KB"
				programInfo["memory_kb"] = 0
				programInfo["memory_mb"] = 0.0
			}
		}

		result = append(result, programInfo)
	}

	return result, nil
}