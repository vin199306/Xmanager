package models

import (
	"time"
)

// Program represents a managed program
type Program struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Command        string    `json:"command"`
	WorkingDir     string    `json:"working_dir,omitempty"`
	Description    string    `json:"description,omitempty"`
	Status         string    `json:"status"`
	PID            int       `json:"pid"`
	Port           int       `json:"port,omitempty"`             // 程序运行的端口
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	AutoStart      bool      `json:"auto_start,omitempty"`
	RestartPolicy  string    `json:"restart_policy,omitempty"`
}

// ProgramManager manages a collection of programs
type ProgramManager struct {
	Programs map[string]Program `json:"programs"`
	Version  string            `json:"version"`
}

// NewProgramManager creates a new ProgramManager instance
func NewProgramManager() *ProgramManager {
	return &ProgramManager{
		Programs: make(map[string]Program),
		Version:  "1.0",
	}
}

// AddProgram adds a new program
func (pm *ProgramManager) AddProgram(program Program) {
	program.CreatedAt = time.Now()
	program.UpdatedAt = time.Now()
	pm.Programs[program.ID] = program
}

// GetProgramByID retrieves a program by ID
func (pm *ProgramManager) GetProgramByID(id string) (Program, bool) {
	program, exists := pm.Programs[id]
	return program, exists
}

// UpdateProgram updates an existing program
func (pm *ProgramManager) UpdateProgram(id string, program Program) bool {
	if _, exists := pm.Programs[id]; !exists {
		return false
	}
	program.UpdatedAt = time.Now()
	pm.Programs[id] = program
	return true
}

// RemoveProgram removes a program
func (pm *ProgramManager) RemoveProgram(id string) bool {
	if _, exists := pm.Programs[id]; !exists {
		return false
	}
	delete(pm.Programs, id)
	return true
}

// GetAllPrograms returns all programs
func (pm *ProgramManager) GetAllPrograms() []Program {
	programs := make([]Program, 0, len(pm.Programs))
	for _, program := range pm.Programs {
		programs = append(programs, program)
	}
	return programs
}

// GetRunningPrograms returns all running programs
func (pm *ProgramManager) GetRunningPrograms() []Program {
	var running []Program
	for _, program := range pm.Programs {
		if program.Status == "running" {
			running = append(running, program)
		}
	}
	return running
}

// GetStoppedPrograms returns all stopped programs
func (pm *ProgramManager) GetStoppedPrograms() []Program {
	var stopped []Program
	for _, program := range pm.Programs {
		if program.Status != "running" {
			stopped = append(stopped, program)
		}
	}
	return stopped
}

// ProcessInfo represents information about a system process
type ProcessInfo struct {
	PID     int    `json:"pid"`
	Name    string `json:"name"`
	Command string `json:"command"`
}