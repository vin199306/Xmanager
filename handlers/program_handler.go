package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"program-manager/models"
	"program-manager/services"
	"program-manager/utils"
	"github.com/gorilla/mux"
)

var startTime = time.Now()

// ProgramHandler handles program-related HTTP requests
type ProgramHandler struct {
	programService *services.ProgramService
}

// NewProgramHandler creates a new ProgramHandler instance
func NewProgramHandler(programService *services.ProgramService) *ProgramHandler {
	return &ProgramHandler{
		programService: programService,
	}
}

// HealthCheck handles GET /api/health
func (ph *ProgramHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"uptime":    time.Since(startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// GetAllPrograms handles GET /api/programs
func (ph *ProgramHandler) GetAllPrograms(w http.ResponseWriter, r *http.Request) {
	utils.LogOperation("HANDLER", "Getting all programs", "", map[string]interface{}{
		"method": r.Method,
		"path":   r.URL.Path,
	})
	
	programs, err := ph.programService.GetAllPrograms()
	if err != nil {
		utils.LogError("HANDLER", "Failed to get all programs", "", err, map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, fmt.Sprintf("Failed to get programs: %v", err), http.StatusInternalServerError)
		return
	}

	utils.LogOperation("HANDLER", "Successfully retrieved all programs", "", map[string]interface{}{
		"count": len(programs),
	})

	// Add security and caching headers
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

// GetProgram handles GET /api/programs/{id}
func (ph *ProgramHandler) GetProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	program, err := ph.programService.GetProgramByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(program)
}

// CreateProgram handles POST /api/programs with enhanced validation
func (ph *ProgramHandler) CreateProgram(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Command     string `json:"command"`
		WorkingDir  string `json:"working_dir"`
		Description string `json:"description"`
		Port        int    `json:"port"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "Program name is required", http.StatusBadRequest)
		return
	}
	if len(req.Name) < 2 || len(req.Name) > 50 {
		http.Error(w, "Program name must be 2-50 characters", http.StatusBadRequest)
		return
	}

	req.Command = strings.TrimSpace(req.Command)
	if req.Command == "" {
		http.Error(w, "Program command is required", http.StatusBadRequest)
		return
	}

	// Validate port range
	if req.Port < 0 || req.Port > 65535 {
		http.Error(w, "Port must be between 0 and 65535", http.StatusBadRequest)
		return
	}

	// Generate unique ID
	id := generateProgramID()
	
	// Check for duplicate IDs (extremely unlikely but just in case)
	existing, _ := ph.programService.GetProgramByID(id)
	if existing != nil {
		// Retry with a new ID
		id = generateProgramID()
		existing, _ = ph.programService.GetProgramByID(id)
		if existing != nil {
			http.Error(w, "Failed to generate unique program ID", http.StatusInternalServerError)
			return
		}
	}

	// Create new program
	program := models.Program{
		ID:          id,
		Name:        req.Name,
		Command:     req.Command,
		WorkingDir:  strings.TrimSpace(req.WorkingDir),
		Description: strings.TrimSpace(req.Description),
		Port:        req.Port,
		Status:      "stopped",
		PID:         0,
	}
	// Treat port 0 as "no port"
	if program.Port == 0 {
		program.Port = -1
	}

	if err := ph.programService.AddProgram(program); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to create program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(program)
}

// generateProgramID generates a unique program ID
func generateProgramID() string {
	return fmt.Sprintf("%s%x", time.Now().Format("20060102150405"), time.Now().UnixNano()%10000)
}

// UpdateProgram handles PUT /api/programs/{id}
// UpdateProgram updates an existing program
func (ph *ProgramHandler) UpdateProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		utils.LogError("HANDLER", "Missing program ID in update request", "", fmt.Errorf("empty id parameter"), nil)
		http.Error(w, "Program ID is required", http.StatusBadRequest)
		return
	}

	// Decode request body
	var req struct {
		Name        string `json:"name"`
		Command     string `json:"command"`
		WorkingDir  string `json:"working_dir"`
		Description string `json:"description"`
		Port        *int   `json:"port"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError("HANDLER", "Failed to decode update request body", id, err, nil)
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	req.Name = strings.TrimSpace(req.Name)
	req.Command = strings.TrimSpace(req.Command)
	req.Description = strings.TrimSpace(req.Description)
	req.WorkingDir = strings.TrimSpace(req.WorkingDir)

	// Get existing program to preserve other fields
	existingProgram, err := ph.programService.GetProgramByID(id)
	if err != nil {
		utils.LogError("HANDLER", "Failed to get existing program for update", id, err, nil)
		http.Error(w, fmt.Sprintf("Failed to get program: %v", err), http.StatusInternalServerError)
		return
	}

	// Update only the fields provided in the request
	existingProgram.Name = req.Name
	existingProgram.Command = req.Command
	existingProgram.WorkingDir = req.WorkingDir
	existingProgram.Description = req.Description
	if req.Port != nil {
		existingProgram.Port = *req.Port
	}

	if req.Name == "" {
		utils.LogError("HANDLER", "Program name validation failed in update", id, fmt.Errorf("empty name"), nil)
		http.Error(w, "Program name is required", http.StatusBadRequest)
		return
	}
	if req.Command == "" {
		utils.LogError("HANDLER", "Program command validation failed in update", id, fmt.Errorf("empty command"), nil)
		http.Error(w, "Program command is required", http.StatusBadRequest)
		return
	}

	// Log validation success
	utils.LogOperation("HANDLER", "Program update validation passed", id, map[string]interface{}{
		"name": req.Name,
		"cmd":  req.Command,
	})

	// Update program via service
	if err := ph.programService.UpdateProgram(id, *existingProgram); err != nil {
		utils.LogError("HANDLER", "Failed to update program", id, err, nil)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else if strings.Contains(err.Error(), "already exists") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to update program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Program updated successfully",
		"id":      id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	utils.LogOperation("HANDLER", "Program updated successfully", id, map[string]interface{}{
		"name": req.Name,
	})
}

// DeleteProgram handles DELETE /api/programs/{id}
func (ph *ProgramHandler) DeleteProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := ph.programService.DeleteProgram(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// StartProgram handles POST /api/programs/{id}/start
func (ph *ProgramHandler) StartProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get program
	program, err := ph.programService.GetProgramByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := ph.programService.StartProgram(program); err != nil {
		if strings.Contains(err.Error(), "already running") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to start program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Get updated program
	updatedProgram, err := ph.programService.GetProgramByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated program: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedProgram)
}

// StopProgram handles POST /api/programs/{id}/stop
func (ph *ProgramHandler) StopProgram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get program
	program, err := ph.programService.GetProgramByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	if err := ph.programService.StopProgram(program); err != nil {
		if strings.Contains(err.Error(), "not running") {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, fmt.Sprintf("Failed to stop program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Get updated program
	updatedProgram, err := ph.programService.GetProgramByID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated program: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedProgram)
}

// BatchStartPrograms handles POST /api/programs/start
func (ph *ProgramHandler) BatchStartPrograms(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IDs []string `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.IDs) == 0 {
		http.Error(w, "Program IDs are required", http.StatusBadRequest)
		return
	}

	results, err := ph.programService.BatchStartPrograms(request.IDs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start programs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// BatchStopPrograms handles POST /api/programs/stop
func (ph *ProgramHandler) BatchStopPrograms(w http.ResponseWriter, r *http.Request) {
	var request struct {
		IDs []string `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.IDs) == 0 {
		http.Error(w, "Program IDs are required", http.StatusBadRequest)
		return
	}

	results, err := ph.programService.BatchStopPrograms(request.IDs)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to stop programs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// ValidateProgramsHandler handles POST /api/programs/validate
func (ph *ProgramHandler) ValidateProgramsHandler(w http.ResponseWriter, r *http.Request) {
	err := ph.programService.ValidateAndFixStatuses()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to validate programs: %v", err), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"message": "Program statuses validated and fixed",
		"success": true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetProgramValidationHandler handles GET /api/programs/{id}/validate
func (ph *ProgramHandler) GetProgramValidationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	program, err := ph.programService.GetProgramByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get program: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Validate the specific program
	isValid := true
	message := "Program status is consistent"

	// Check if process actually exists
	actualStatus := "stopped"
	actualPID := 0
	if program.PID > 0 {
		// Use process manager to check if process is running
		// This is a simplified check - in real implementation, use process manager
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", program.PID))
		output, _ := cmd.CombinedOutput()
		if strings.Contains(string(output), fmt.Sprintf("%d", program.PID)) {
			actualStatus = "running"
			actualPID = program.PID
		}
	}

	if actualStatus != program.Status {
		isValid = false
		message = fmt.Sprintf("Status mismatch: stored=%s, actual=%s", program.Status, actualStatus)
		
		// Auto-fix the status
		program.Status = actualStatus
		program.PID = actualPID
		if err := ph.programService.UpdateProgram(id, *program); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update program status: %v", err), http.StatusInternalServerError)
			return
		}
	}

	result := map[string]interface{}{
		"id":      id,
		"valid":   isValid,
		"message": message,
		"program": program,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}