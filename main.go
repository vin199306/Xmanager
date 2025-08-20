package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"program-manager/config"
	"program-manager/handlers"
	"program-manager/middleware"
	"program-manager/services"
	"program-manager/storage"
	"program-manager/utils"
)

func main() {

	
	// Log application start
	utils.LogOperation("SYSTEM", "Program Manager starting", "", map[string]interface{}{
		"version": "1.0.0",
		"os":      "linux",
	})
	
	// Parse command line arguments
	portFlag := flag.String("p", "", "Port to run the server on (e.g., :8080)")
	flag.Parse()

	// Initialize configuration
	cfg := config.NewConfig()
	if *portFlag != "" {
		cfg.ServerPort = *portFlag
	}

	utils.LogOperation("SYSTEM", "Configuration loaded", "", map[string]interface{}{
		"port":      cfg.ServerPort,
		"data_path": cfg.GetDataFilePath(),
		"log_path":  cfg.GetLogDirectory(),
	})

	// Ensure necessary directories exist
	if err := cfg.EnsureDirectories(); err != nil {
		utils.LogError("SYSTEM", "Failed to create directories", "", err, nil)
		log.Printf("Failed to create directories: %v", err)
		os.Exit(1)
	}

	// Initialize storage
	fileStorage := storage.NewFileStorage(cfg.GetDataFilePath())
	
	// Check if data file exists, if not create with embedded default data
	if _, err := os.Stat(cfg.GetDataFilePath()); os.IsNotExist(err) {
		utils.LogOperation("SYSTEM", "Creating initial data file", cfg.GetDataFilePath(), nil)
		if err := os.WriteFile(cfg.GetDataFilePath(), []byte(`{"programs":[]}`), 0644); err != nil {
			utils.LogError("SYSTEM", "Failed to create initial data file", cfg.GetDataFilePath(), err, nil)
			log.Printf("Warning: Failed to create initial data file: %v", err)
		}
	} else {
		utils.LogOperation("SYSTEM", "Data file exists", cfg.GetDataFilePath(), nil)
	}

	// Initialize process manager
	processManager := utils.NewProcessManager(cfg.GetLogDirectory())

	// Initialize services
	programService := services.NewProgramService(fileStorage, processManager)
	statusService := services.NewStatusService(fileStorage, processManager)

	// Initialize handlers
	programHandler := handlers.NewProgramHandler(programService)
	logHandler := handlers.NewLogHandler(programService.LogService)
	statusHandler := handlers.NewStatusHandler(statusService)
	
	// Create router
	router := mux.NewRouter()

	// Add middleware
	router.Use(middleware.RequestLoggingMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.NewAPIEndpointLogger(next)
	})

	// API routes
	api := router.PathPrefix("/api").Subrouter()
	
	// Program routes
	api.HandleFunc("/programs", programHandler.GetAllPrograms).Methods("GET")
	api.HandleFunc("/programs", programHandler.CreateProgram).Methods("POST")
	api.HandleFunc("/programs/memory", statusHandler.GetProgramsWithMemory).Methods("GET")
	api.HandleFunc("/programs/{id}", programHandler.GetProgram).Methods("GET")
	api.HandleFunc("/programs/{id}", programHandler.UpdateProgram).Methods("PUT")
	api.HandleFunc("/programs/{id}", programHandler.DeleteProgram).Methods("DELETE")
	api.HandleFunc("/programs/{id}/start", programHandler.StartProgram).Methods("POST")
	api.HandleFunc("/programs/{id}/stop", programHandler.StopProgram).Methods("POST")

	// Log routes
	api.HandleFunc("/logs/{id}", logHandler.GetProgramLogs).Methods("GET")

	// Status routes
	api.HandleFunc("/status", statusHandler.GetAllProgramsStatus).Methods("GET")
	api.HandleFunc("/status/{id}", statusHandler.GetProgramStatus).Methods("GET")
	api.HandleFunc("/system/memory", statusHandler.GetSystemMemoryInfo).Methods("GET")

	// Static file serving - use embedded files for production
	// In development, check for actual files
	var staticHandler http.Handler
	
	// Check if web directory exists for development mode
	execDir, err := os.Getwd()
	if err != nil {
		execDir = "."
	}
	
	webPath := filepath.Join(execDir, "web")
	
	if _, err := os.Stat(webPath); err == nil {
		// Development mode: serve from filesystem (web)
		utils.LogOperation("SYSTEM", "Development mode: serving static files from web directory", webPath, nil)
		staticHandler = http.FileServer(http.Dir(webPath))
	} else {
		// Production mode: use embedded files
			utils.LogOperation("SYSTEM", "Production mode: serving static files from embedded files", "", nil)
			staticHandler = getWebHandler()
	}

	// Handle all other routes with the file server
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Add caching headers for static assets
		if filepath.Ext(r.URL.Path) != "" {
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
		}

		// For SPA routing: serve index.html for non-file requests
		if r.URL.Path == "/" || filepath.Ext(r.URL.Path) == "" {
			if _, err := os.Stat(filepath.Join(execDir, "web", "index.html")); err == nil {
				// Development mode: serve actual index.html
				http.ServeFile(w, r, filepath.Join(execDir, "web", "index.html"))
				return
			}
		}

		// Serve the requested file
		staticHandler.ServeHTTP(w, r)
	})

	// Log server start
	utils.LogOperation("SYSTEM", "Server starting", "", map[string]interface{}{
		"port": cfg.ServerPort,
		"host": "0.0.0.0",
	})

	// Start server
	utils.LogOperation("SYSTEM", "Server started successfully", "", map[string]interface{}{
		"address": cfg.ServerPort,
	})
	if err := http.ListenAndServe(cfg.ServerPort, router); err != nil {
		utils.LogError("SYSTEM", "Server failed to start", "", err, map[string]interface{}{
			"port": cfg.ServerPort,
		})
		log.Printf("Server failed to start: %v", err)
		os.Exit(1)
	}
}