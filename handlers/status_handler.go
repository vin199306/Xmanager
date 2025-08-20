package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"program-manager/services"
)

// StatusHandler handles status-related HTTP requests
type StatusHandler struct {
	statusService *services.StatusService
}

// NewStatusHandler creates a new StatusHandler instance
func NewStatusHandler(statusService *services.StatusService) *StatusHandler {
	return &StatusHandler{
		statusService: statusService,
	}
}

// GetProgramStatus handles GET /api/programs/{id}/status
func (sh *StatusHandler) GetProgramStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	program, err := sh.statusService.GetProgramStatus(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to get program status: %v", err), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(program)
}

// GetAllProgramsStatus handles GET /api/programs/status
func (sh *StatusHandler) GetAllProgramsStatus(w http.ResponseWriter, r *http.Request) {
	programs, err := sh.statusService.GetAllProgramsStatus()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get programs status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

// RefreshAllStatuses handles POST /api/programs/refresh
func (sh *StatusHandler) RefreshAllStatuses(w http.ResponseWriter, r *http.Request) {
	if err := sh.statusService.RefreshAllStatuses(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to refresh statuses: %v", err), http.StatusInternalServerError)
		return
	}

	programs, err := sh.statusService.GetAllProgramsStatus()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get updated programs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

// GetRunningPrograms handles GET /api/programs/running
func (sh *StatusHandler) GetRunningPrograms(w http.ResponseWriter, r *http.Request) {
	programs, err := sh.statusService.GetRunningPrograms()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get running programs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

// GetStoppedPrograms handles GET /api/programs/stopped
func (sh *StatusHandler) GetStoppedPrograms(w http.ResponseWriter, r *http.Request) {
	programs, err := sh.statusService.GetStoppedPrograms()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get stopped programs: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}

// GetProcessInfo handles GET /api/processes/{pid}/info
func (sh *StatusHandler) GetProcessInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pidStr := vars["pid"]

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		http.Error(w, "Invalid process ID", http.StatusBadRequest)
		return
	}

	info, err := sh.statusService.GetProcessInfo(pid)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get process info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetSystemMemoryInfo handles GET /api/system/memory
func (sh *StatusHandler) GetSystemMemoryInfo(w http.ResponseWriter, r *http.Request) {
	info, err := sh.statusService.GetSystemMemoryInfo()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get system memory info: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

// GetProgramsWithMemory handles GET /api/programs/memory
func (sh *StatusHandler) GetProgramsWithMemory(w http.ResponseWriter, r *http.Request) {
	programs, err := sh.statusService.GetProgramsWithMemory()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get programs with memory: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(programs)
}