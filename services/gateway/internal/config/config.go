package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	// Server
	Port string

	// gRPC Services
	AuthServiceGRPC         string
	HelpDeskServiceGRPC     string
	NotificationServiceGRPC string

	// HTTP Services (for REST API proxying)
	AuthServiceHTTP string

	// Legacy Database (for temporary handlers)
	LegacyDBHost     string
	LegacyDBPort     string
	LegacyDBName     string
	LegacyDBUser     string
	LegacyDBPassword string
	LegacyDBSchema   string
	LegacyDB         *gorm.DB
}

func Load() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		Port:                    getEnv("GATEWAY_PORT", "8080"),
		AuthServiceGRPC:         getEnv("AUTH_SERVICE_GRPC", "localhost:9001"),
		HelpDeskServiceGRPC:     getEnv("HELPDESK_SERVICE_GRPC", "localhost:9002"),
		NotificationServiceGRPC: getEnv("NOTIFICATION_SERVICE_GRPC", "localhost:9003"),
		// Determine Auth Service HTTP URL based on gRPC address
		AuthServiceHTTP:  getAuthServiceHTTP(),
		LegacyDBHost:     getEnv("LEGACY_DB_HOST", "localhost"),
		LegacyDBPort:     getEnv("LEGACY_DB_PORT", "5432"),
		LegacyDBName:     getEnv("LEGACY_DB_NAME", "ITaaS"),
		LegacyDBUser:     getEnv("LEGACY_DB_USER", "postgres"),
		LegacyDBPassword: getEnv("LEGACY_DB_PASSWORD", "postgres"),
		LegacyDBSchema:   getEnv("LEGACY_DB_SCHEMA", "public"),
	}

	cfg.LegacyDB = initLegacyDB(cfg)
	return cfg
}

func getAuthServiceHTTP() string {
	grpcAddr := getEnv("AUTH_SERVICE_GRPC", "localhost:9001")
	// If using Docker service name, use HTTP service name; otherwise localhost
	if strings.Contains(grpcAddr, "auth-service") {
		return "http://auth-service:8001"
	}
	return "http://localhost:8001"
}

func initLegacyDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		cfg.LegacyDBHost, cfg.LegacyDBPort, cfg.LegacyDBUser, cfg.LegacyDBPassword, cfg.LegacyDBName, cfg.LegacyDBSchema)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Warning: failed to connect to legacy DB: %v", err)
		return nil
	}

	log.Println("Connected to Legacy PostgreSQL database")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
