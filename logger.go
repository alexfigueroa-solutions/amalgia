package main

import (
	"io"
	"log"
	"os"
)

// InitializeLogger sets up logging to both stderr and a log file.
func InitializeLogger() *log.Logger {
	// Open a log file for writing
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Create a multi-writer to write to both stderr and the log file
	multiWriter := io.MultiWriter(os.Stderr, logFile)

	// Create a new logger
	logger := log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	return logger
}
