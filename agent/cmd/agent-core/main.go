package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"backend/agent/config"
	"backend/agent/internal/core"
	"backend/agent/internal/utils"
)

func main() {
	// Parse command line flags
	var configPath string
	var containerID string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.StringVar(&containerID, "container-id", "", "Container ID for device registration")
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
	logger, err := utils.NewLogger("agent-core", cfg.LogDir, "agent-core.log", cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	logger.Info("Starting agent core service", map[string]interface{}{
		"version": utils.GetVersion(),
	})

	// Create agent core service
	agentService, err := core.NewService(cfg, logger)
	if err != nil {
		logger.Error("Failed to create agent service", err)
		os.Exit(1)
	}

	// If container ID provided, attempt registration
	if containerID != "" {
		logger.Info("Registering device", map[string]interface{}{
			"container_id": containerID,
		})

		if err := agentService.RegisterDevice(containerID); err != nil {
			logger.Error("Device registration failed", err)
			os.Exit(1)
		}

		logger.Info("Device registration initiated. Verification code required.")
		// Note: In production, verification would be handled separately
	}

	// Start agent service
	if err := agentService.Start(); err != nil {
		logger.Error("Failed to start agent service", err)
		os.Exit(1)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down agent core service")

	// Stop agent service
	if err := agentService.Stop(); err != nil {
		logger.Error("Failed to stop agent service", err)
		os.Exit(1)
	}

	logger.Info("Agent core service stopped")
}

