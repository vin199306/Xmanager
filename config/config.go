package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config holds application configuration
type Config struct {
	ServerPort   string
	DataFile     string
	LogDirectory string
	TempDir      string
}

// NewConfig creates a new configuration instance with default values
func NewConfig() *Config {
	// Get executable directory for default paths
	exeDir, err := os.Getwd()
	if err != nil {
		exeDir = "."
	}

	// Use user home directory for Linux/macOS for better permissions
	var dataDir, logDir string
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			dataDir = filepath.Join(homeDir, ".program-manager")
			logDir = filepath.Join(dataDir, "logs")
		} else {
			dataDir = exeDir
			logDir = filepath.Join(exeDir, "logs")
		}
	} else {
		dataDir = exeDir
		logDir = filepath.Join(exeDir, "logs")
	}

	return &Config{
		ServerPort:   ":8081",
		DataFile:     filepath.Join(dataDir, "programs.json"),
		LogDirectory: logDir,
		TempDir:      filepath.Join(dataDir, "temp"),
	}
}

// EnsureDirectories creates necessary directories if they don't exist
func (c *Config) EnsureDirectories() error {
	// Create logs directory with appropriate permissions
	if err := os.MkdirAll(c.LogDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %v", err)
	}
	
	// Create temp directory for temporary files
	if err := os.MkdirAll(c.TempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %v", err)
	}
	
	return nil
}

// GetDataFilePath returns the full path to the data file
func (c *Config) GetDataFilePath() string {
	return c.DataFile
}

// GetLogDirectory returns the log directory path
func (c *Config) GetLogDirectory() string {
	return c.LogDirectory
}

// GetTempDir returns the temporary directory path
func (c *Config) GetTempDir() string {
	return c.TempDir
}