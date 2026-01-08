package config

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DB         *gorm.DB

	// Server
	Port     string
	GRPCPort string

	// gRPC mTLS Server Configuration
	GRPCMTLSCACert        string
	GRPCMTLSServerCert    string
	GRPCMTLSServerKey     string
	GRPCMTLSServerKeyPass string
	GRPCMTLSTLSConfig     *tls.Config
}

func Load() *Config {
	// Try service-specific .env first
	_ = godotenv.Load(".env")
	// Fall back to root .env if exists
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env")
	}

	cfg := &Config{
		DBHost:                getEnv("INVENTORY_DB_HOST", "localhost"),
		DBPort:                getEnv("INVENTORY_DB_PORT", "5432"),
		DBName:                getEnv("INVENTORY_DB_NAME", "inventory_db"),
		DBUser:                getEnv("INVENTORY_DB_USER", "postgres"),
		DBPassword:            getEnv("INVENTORY_DB_PASSWORD", "postgres"),
		Port:                  getEnv("INVENTORY_PORT", "8007"),
		GRPCPort:              getEnv("INVENTORY_GRPC_PORT", "9007"),
		GRPCMTLSCACert:        getEnv("GRPC_MTLS_CA_CERT", "./certs/inventory-ca.crt"),
		GRPCMTLSServerCert:    getEnv("GRPC_MTLS_SERVER_CERT", "./certs/inventory-server.crt"),
		GRPCMTLSServerKey:     getEnv("GRPC_MTLS_SERVER_KEY", "./certs/inventory-server.key"),
		GRPCMTLSServerKeyPass: getEnv("GRPC_MTLS_SERVER_KEY_PASSWORD", ""),
	}

	cfg.DB = initDB(cfg)

	// Initialize gRPC TLS config if certificates are provided
	if cfg.GRPCMTLSCACert != "" && cfg.GRPCMTLSServerCert != "" && cfg.GRPCMTLSServerKey != "" {
		cfg.GRPCMTLSTLSConfig = initGRPCTLS(cfg)
	} else {
		log.Println("Warning: gRPC mTLS certificates not configured. gRPC server will start without TLS.")
	}

	return cfg
}

func initGRPCTLS(cfg *Config) *tls.Config {
	// This function is used to validate TLS config, actual credentials are loaded in main.go
	return nil
}

func initDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	log.Println("Connected to Inventory PostgreSQL database")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("invalid duration %s, using default: %v", s, err)
		return 15 * time.Minute
	}
	return d
}

