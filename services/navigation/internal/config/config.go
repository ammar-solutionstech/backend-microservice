package config

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"

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
	_ = godotenv.Load(".env")
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env")
	}

	cfg := &Config{
		DBHost:                getEnv("NAVIGATION_DB_HOST", "localhost"),
		DBPort:                getEnv("NAVIGATION_DB_PORT", "5432"),
		DBName:                getEnv("NAVIGATION_DB_NAME", "navigation_db"),
		DBUser:                getEnv("NAVIGATION_DB_USER", "postgres"),
		DBPassword:            getEnv("NAVIGATION_DB_PASSWORD", "postgres"),
		Port:                  getEnv("NAVIGATION_PORT", "8009"),
		GRPCPort:              getEnv("NAVIGATION_GRPC_PORT", "9009"),
		GRPCMTLSCACert:        getEnv("GRPC_MTLS_CA_CERT", "./certs/navigation-ca.crt"),
		GRPCMTLSServerCert:    getEnv("GRPC_MTLS_SERVER_CERT", "./certs/navigation-server.crt"),
		GRPCMTLSServerKey:     getEnv("GRPC_MTLS_SERVER_KEY", "./certs/navigation-server.key"),
		GRPCMTLSServerKeyPass: getEnv("GRPC_MTLS_SERVER_KEY_PASSWORD", ""),
	}

	cfg.DB = initDB(cfg)

	if cfg.GRPCMTLSCACert != "" && cfg.GRPCMTLSServerCert != "" && cfg.GRPCMTLSServerKey != "" {
		cfg.GRPCMTLSTLSConfig = initGRPCTLS(cfg)
	} else {
		log.Println("Warning: gRPC mTLS certificates not configured. gRPC server will start without TLS.")
	}

	return cfg
}

func initGRPCTLS(cfg *Config) *tls.Config {
	return nil
}

func initDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	log.Println("Connected to Navigation PostgreSQL database")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

