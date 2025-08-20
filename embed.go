package main

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed web/*
var webFS embed.FS

//go:embed programs.json
var defaultData []byte

// getWebHandler returns http.Handler for web files
func getWebHandler() http.Handler {
	// Create a sub-filesystem for the web directory
	webSubFS, err := fs.Sub(webFS, "web")
	if err != nil {
		// Log the error and return a simple 404 handler instead of panicking
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Web files not found"))
		})
	}
	return http.FileServer(http.FS(webSubFS))
}

// getDefaultData returns the default programs.json content
func getDefaultData() []byte {
	return defaultData
}