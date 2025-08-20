package services

import (
	"fmt"
	"os"
	"path/filepath"
	"program-manager/models"
	"program-manager/storage"
	"program-manager/utils"
	"time"
)

// ProgramService handles program management business logic
type ProgramService struct {
	storage        *storage.FileStorage
	processManager *utils.ProcessManager
	LogService     *LogService
}

// NewProgramService creates a new ProgramService instance
func NewProgramService(storage *storage.FileStorage, processManager *utils.ProcessManager) *ProgramService {
	logsDir := filepath.Join("data", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		logsDir = "logs"
	}
	
	return &ProgramService{
		storage:        storage,
		processManager: processManager,
		LogService:     NewLogService(logsDir),
	}
}

// GetAllPrograms retrieves all programs
func (ps *ProgramService) GetAllPrograms() ([]models.Program, error) {
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}
	
	programs := make([]models.Program, 0, len(pm.Programs))
	for _, program := range pm.Programs {
		programs = append(programs, program)
	}
	return programs, nil
}

// GetProgramByID retrieves a specific program by ID
func (ps *ProgramService) GetProgramByID(id string) (*models.Program, error) {
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		return nil, fmt.Errorf("failed to load programs: %v", err)
	}

	program, found := pm.GetProgramByID(id)
	if !found {
		return nil, fmt.Errorf("program not found: %s", id)
	}

	return &program, nil
}

// AddProgram adds a new program with validation
func (ps *ProgramService) AddProgram(program models.Program) error {
	utils.LogOperation("SERVICE", "Adding new program", program.ID, map[string]interface{}{
		"name":        program.Name,
		"command":     program.Command,
		"port":        program.Port,
		"working_dir": program.WorkingDir,
	})

	// Validate program
	if program.Name == "" {
		utils.LogError("SERVICE", "Program name validation failed", program.ID, fmt.Errorf("empty name"), nil)
		return fmt.Errorf("program name is required")
	}
	if program.Command == "" {
		utils.LogError("SERVICE", "Program command validation failed", program.ID, fmt.Errorf("empty command"), nil)
		return fmt.Errorf("program command is required")
	}
	if err := utils.CommandValidator.ValidateCommand(program.Command); err != nil {
		utils.LogError("SERVICE", "Command validation failed", program.ID, err, map[string]interface{}{
			"command": program.Command,
		})
		return fmt.Errorf("invalid command: %v", err)
	}
	if program.WorkingDir != "" {
		if err := utils.CommandValidator.ValidateWorkingDir(program.WorkingDir); err != nil {
			utils.LogError("SERVICE", "Working directory validation failed", program.ID, err, map[string]interface{}{
				"working_dir": program.WorkingDir,
			})
			return fmt.Errorf("invalid working directory: %v", err)
		}
	}

	// Load existing programs
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		utils.LogError("SERVICE", "Failed to load programs for add operation", program.ID, err, nil)
		return fmt.Errorf("failed to load programs: %v", err)
	}

	// Check for duplicate name
	for _, existing := range pm.Programs {
		if existing.Name == program.Name && existing.ID != program.ID {
			utils.LogError("SERVICE", "Duplicate program name", program.ID, fmt.Errorf("name '%s' already exists", program.Name), nil)
			return fmt.Errorf("program with name '%s' already exists", program.Name)
		}
	}

	// Set timestamps
	now := time.Now()
	if program.CreatedAt.IsZero() {
		program.CreatedAt = now
	}
	program.UpdatedAt = now

	// Add program
	pm.AddProgram(program)

	// Save to storage
	if err := ps.storage.SavePrograms(pm); err != nil {
		utils.LogError("SERVICE", "Failed to save program", program.ID, err, nil)
		return fmt.Errorf("failed to save program: %v", err)
	}

	utils.LogOperation("SERVICE", "Program added successfully", program.ID, map[string]interface{}{
		"program_name": program.Name,
	})

	return nil
}

// UpdateProgram updates an existing program
func (ps *ProgramService) UpdateProgram(id string, program models.Program) error {
	// Validate program
	if program.Name == "" {
		return fmt.Errorf("program name is required")
	}
	if program.Command == "" {
		return fmt.Errorf("program command is required")
	}
	if err := utils.CommandValidator.ValidateCommand(program.Command); err != nil {
		return fmt.Errorf("invalid command: %v", err)
	}
	if program.WorkingDir != "" {
		if err := utils.CommandValidator.ValidateWorkingDir(program.WorkingDir); err != nil {
			return fmt.Errorf("invalid working directory: %v", err)
		}
	}

	// Load existing programs
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		return fmt.Errorf("failed to load programs: %v", err)
	}

	// Check for duplicate name (excluding current program)
	for _, existing := range pm.Programs {
		if existing.Name == program.Name && existing.ID != id {
			return fmt.Errorf("program with name '%s' already exists", program.Name)
		}
	}

	// Update program
	program.UpdatedAt = time.Now()
	if !pm.UpdateProgram(id, program) {
		return fmt.Errorf("program not found: %s", id)
	}

	// Save to storage
	if err := ps.storage.SavePrograms(pm); err != nil {
		return fmt.Errorf("failed to save program: %v", err)
	}

	return nil
}

// DeleteProgram removes a program
func (ps *ProgramService) DeleteProgram(id string) error {
	// Load existing programs
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		return fmt.Errorf("failed to load programs: %v", err)
	}

	// Check if program exists
	program, exists := pm.GetProgramByID(id)
	if !exists {
		return fmt.Errorf("program not found: %s", id)
	}

	// Stop program if running
	if program.Status == "running" {
		if err := ps.StopProgram(&program); err != nil {
			return fmt.Errorf("failed to stop program before deletion: %v", err)
		}
	}

	// Remove program
	if !pm.RemoveProgram(id) {
		return fmt.Errorf("program not found: %s", id)
	}

	// Save to storage
	if err := ps.storage.SavePrograms(pm); err != nil {
		return fmt.Errorf("failed to save programs: %v", err)
	}

	return nil
}

// StartProgram starts a program
func (ps *ProgramService) StartProgram(program *models.Program) error {
	// Check if program is already running
	if program.Status == "running" {
		return fmt.Errorf("program is already running")
	}

	// Validate program
	if program.Command == "" {
		return fmt.Errorf("program command is required")
	}

	// Load programs to get the latest version
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		return fmt.Errorf("failed to load programs: %v", err)
	}

	// Get the program from storage to ensure we have the latest version
	storedProgram, exists := pm.GetProgramByID(program.ID)
	if !exists {
		return fmt.Errorf("program not found: %s", program.ID)
	}

	// Start the program
	if err := ps.processManager.StartProgram(&storedProgram); err != nil {
		return fmt.Errorf("failed to start program: %v", err)
	}

	// Update the program in storage
	storedProgram.UpdatedAt = time.Now()
	if !pm.UpdateProgram(program.ID, storedProgram) {
		return fmt.Errorf("program not found: %s", program.ID)
	}

	// Save to storage
	if err := ps.storage.SavePrograms(pm); err != nil {
		return fmt.Errorf("failed to save program: %v", err)
	}

	// Update the original program object
	*program = storedProgram

	return nil
}

// StopProgram stops a program
func (ps *ProgramService) StopProgram(program *models.Program) error {
	// Load programs to get the latest version
	pm, err := ps.storage.LoadPrograms()
	if err != nil {
		return fmt.Errorf("failed to load programs: %v", err)
	}

	// Get the program from storage to ensure we have the latest version
	storedProgram, exists := pm.GetProgramByID(program.ID)
	if !exists {
		return fmt.Errorf("program not found: %s", program.ID)
	}

	// Stop the program
	if err := ps.processManager.StopProgram(&storedProgram); err != nil {
		return fmt.Errorf("failed to stop program: %v", err)
	}

	// Update the program in storage
	storedProgram.UpdatedAt = time.Now()
	if !pm.UpdateProgram(program.ID, storedProgram) {
		return fmt.Errorf("program not found: %s", program.ID)
	}

	// Save to storage
	if err := ps.storage.SavePrograms(pm); err != nil {
		return fmt.Errorf("failed to save program: %v", err)
	}

	// Update the original program object
	*program = storedProgram

	return nil
}

// GetProgramStatus gets the actual status of a program
func (ps *ProgramService) GetProgramStatus(program *models.Program) (string, int, error) {
	return ps.processManager.GetProgramStatus(program)
}

// BatchStartPrograms starts multiple programs
func (ps *ProgramService) BatchStartPrograms(ids []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	errorCount := 0

	for _, id := range ids {
		program, err := ps.GetProgramByID(id)
		if err != nil {
			results[id] = map[string]string{"error": err.Error()}
			errorCount++
			continue
		}

		if err := ps.StartProgram(program); err != nil {
			results[id] = map[string]string{"error": err.Error()}
			errorCount++
		} else {
			results[id] = map[string]string{"status": "started"}
			successCount++
		}
	}

	results["summary"] = map[string]interface{}{
		"total":   len(ids),
		"success": successCount,
		"errors":  errorCount,
	}

	return results, nil
}

// BatchStopPrograms stops multiple programs
func (ps *ProgramService) BatchStopPrograms(ids []string) (map[string]interface{}, error) {
	results := make(map[string]interface{})
	successCount := 0
	errorCount := 0

	for _, id := range ids {
		program, err := ps.GetProgramByID(id)
		if err != nil {
			results[id] = map[string]string{"error": err.Error()}
			errorCount++
			continue
		}

		if err := ps.StopProgram(program); err != nil {
			results[id] = map[string]string{"error": err.Error()}
			errorCount++
		} else {
			results[id] = map[string]string{"status": "stopped"}
			successCount++
		}
	}

	results["summary"] = map[string]interface{}{
		"total":   len(ids),
		"success": successCount,
		"errors":  errorCount,
	}

	return results, nil
}

// ValidateAndFixStatuses validates and fixes program statuses
func (ps *ProgramService) ValidateAndFixStatuses() error {
	programs, err := ps.GetAllPrograms()
	if err != nil {
		return fmt.Errorf("failed to get programs: %v", err)
	}

	for _, program := range programs {
		needsUpdate := false
		originalStatus := program.Status
		originalPID := program.PID

		// Check if process actually exists
		if program.Status == "running" && program.PID > 0 {
			// Check if process is actually running
			isRunning := ps.processManager.IsProcessRunning(program.PID)
			if !isRunning {
				program.Status = "stopped"
				program.PID = 0
				needsUpdate = true
			}
		} else if program.Status == "running" && program.PID == 0 {
			// Inconsistent state - mark as stopped
			program.Status = "stopped"
			needsUpdate = true
		}

		// Update if needed
		if needsUpdate {
			if err := ps.UpdateProgram(program.ID, program); err != nil {
				utils.LogWarning(fmt.Sprintf("Failed to update program %s status: %v", program.ID, err), program.ID)
			} else {
				utils.Log(fmt.Sprintf("Fixed program %s status: %s -> %s, PID: %d -> %d", 
					program.ID, originalStatus, program.Status, originalPID, program.PID), program.ID)
			}
		}
	}

	return nil
}