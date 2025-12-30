package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"backend/agent/config"
	"backend/agent/internal/update"
	"backend/agent/internal/utils"
)

func main() {
	// Parse command line flags
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logger, err := utils.NewLogger("update-service", cfg.LogDir, "update-service.log", cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Starting update service", map[string]interface{}{
		"version": utils.GetVersion(),
	})

	// Create update service
	updateService, err := update.NewService(cfg, logger)
	if err != nil {
		logger.Error("Failed to create update service", err)
		os.Exit(1)
	}

	// Start update service
	if err := updateService.Start(); err != nil {
		logger.Error("Failed to start update service", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down update service")

	// Stop update service
	if err := updateService.Stop(); err != nil {
		logger.Error("Failed to stop update service", err)
		os.Exit(1)
	}

	logger.Info("Update service stopped")
}

