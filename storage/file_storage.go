package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"


	"program-manager/models"
)

// FileStorage handles file-based data persistence
type FileStorage struct {
	filePath string
	mutex    sync.RWMutex
}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
		mutex:    sync.RWMutex{},
	}
}

// SavePrograms saves the program manager data to JSON file with atomic write
func (fs *FileStorage) SavePrograms(pm *models.ProgramManager) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// Create directory if it doesn't exist
	if err := fs.ensureDirectory(); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Create a buffer to reduce memory allocations
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(pm); err != nil {
		return fmt.Errorf("failed to encode data: %v", err)
	}

	// Atomic write: write to temp file then rename
	tempPath := fs.filePath + ".tmp"
	if err := os.WriteFile(tempPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %v", err)
	}

	if err := os.Rename(tempPath, fs.filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}

// LoadPrograms loads program manager data from JSON file
func (fs *FileStorage) LoadPrograms() (*models.ProgramManager, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	// Check if file exists
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		// Use embedded default data will be handled by the caller
		// For now, return empty manager to avoid import cycle
		return models.NewProgramManager(), nil
	}

	// Read file
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Unmarshal JSON
	var pm models.ProgramManager
	if err := json.Unmarshal(data, &pm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %v", err)
	}

	// Initialize empty map if nil
	if pm.Programs == nil {
		pm.Programs = make(map[string]models.Program)
	}

	return &pm, nil
}

// ExportPrograms exports programs to a JSON file
func (fs *FileStorage) ExportPrograms(exportPath string, pm *models.ProgramManager) error {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	data, err := json.MarshalIndent(pm, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	if err := os.WriteFile(exportPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write export file: %v", err)
	}

	return nil
}

// ImportPrograms imports programs from a JSON file
func (fs *FileStorage) ImportPrograms(importPath string) (*models.ProgramManager, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	// Check if import file exists
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("import file does not exist: %s", importPath)
	}

	// Read import file
	data, err := os.ReadFile(importPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %v", err)
	}

	// Unmarshal JSON
	var pm models.ProgramManager
	if err := json.Unmarshal(data, &pm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal import data: %v", err)
	}

	// Initialize empty map if nil
	if pm.Programs == nil {
		pm.Programs = make(map[string]models.Program)
	}

	return &pm, nil
}

// ensureDirectory creates the directory for the file if it doesn't exist
func (fs *FileStorage) ensureDirectory() error {
	// Get directory path from file path
	dir := filepath.Dir(fs.filePath)
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0755)
}

// GetFilePath returns the file path used by this storage
func (fs *FileStorage) GetFilePath() string {
	return fs.filePath
}