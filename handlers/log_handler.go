package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"program-manager/services"
)

// LogHandler handles HTTP requests for program logs
type LogHandler struct {
	logService *services.LogService
}

// NewLogHandler creates a new LogHandler instance
func NewLogHandler(logService *services.LogService) *LogHandler {
	return &LogHandler{
		logService: logService,
	}
}

// GetProgramLogs handles GET /api/programs/:id/logs
func (lh *LogHandler) GetProgramLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	programID := r.URL.Path[len("/api/programs/"):len(r.URL.Path)-len("/logs")]
	if programID == "" {
		writeError(w, http.StatusBadRequest, "Program ID is required")
		return
	}

	// Get limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := lh.logService.GetLogs(programID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get logs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

// GetAllLogs handles GET /api/logs
func (lh *LogHandler) GetAllLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	logs, err := lh.logService.GetAllLogs(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get logs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

// ClearProgramLogs handles DELETE /api/programs/:id/logs
func (lh *LogHandler) ClearProgramLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	programID := r.URL.Path[len("/api/programs/"):len(r.URL.Path)-len("/logs")]
	if programID == "" {
		writeError(w, http.StatusBadRequest, "Program ID is required")
		return
	}

	if err := lh.logService.ClearLogs(programID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to clear logs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Logs cleared successfully"})
}

// ClearAllLogs handles DELETE /api/logs
func (lh *LogHandler) ClearAllLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := lh.logService.ClearAllLogs(); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to clear logs: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "All logs cleared successfully"})
}

// writeJSON is a helper function to write JSON responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError is a helper function to write error responses
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}