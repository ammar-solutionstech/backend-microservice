package config

import (
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

	// Auth Service gRPC
	AuthServiceGRPC string

	// gRPC mTLS Server Configuration
	GRPCMTLSCACert        string
	GRPCMTLSServerCert    string
	GRPCMTLSServerKey     string
	GRPCMTLSServerKeyPass string

	// gRPC mTLS Client Configuration (for Auth service)
	AuthServiceGRPCMTLSCA      string
	AuthServiceGRPCMTLSClientCert string
	AuthServiceGRPCMTLSClientKey  string

	// RabbitMQ
	RabbitMQURL      string
	RabbitMQUser     string
	RabbitMQPassword string
}

func Load() *Config {
	// Try service-specific .env first
	_ = godotenv.Load(".env")
	// Fall back to root .env if exists
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env")
	}

	cfg := &Config{
		DBHost:          getEnv("HELPDESK_DB_HOST", "localhost"),
		DBPort:          getEnv("HELPDESK_DB_PORT", "5432"),
		DBName:          getEnv("HELPDESK_DB_NAME", "helpdesk_db"),
		DBUser:          getEnv("HELPDESK_DB_USER", "postgres"),
		DBPassword:      getEnv("HELPDESK_DB_PASSWORD", "postgres"),
		Port:            getEnv("HELPDESK_PORT", "8002"),
		GRPCPort:        getEnv("HELPDESK_GRPC_PORT", "9002"),
		AuthServiceGRPC: getEnv("AUTH_SERVICE_GRPC", "localhost:9001"),
		
		// gRPC mTLS Server Configuration
		GRPCMTLSCACert:        getEnv("GRPC_MTLS_CA_CERT", "./certs/helpdesk-ca.crt"),
		GRPCMTLSServerCert:    getEnv("GRPC_MTLS_SERVER_CERT", "./certs/helpdesk-server.crt"),
		GRPCMTLSServerKey:     getEnv("GRPC_MTLS_SERVER_KEY", "./certs/helpdesk-server.key"),
		GRPCMTLSServerKeyPass: getEnv("GRPC_MTLS_SERVER_KEY_PASSWORD", ""),
		
		// Auth Service gRPC mTLS Client Configuration
		AuthServiceGRPCMTLSCA:           getEnv("AUTH_SERVICE_GRPC_MTLS_CA", "./certs/auth-ca.crt"),
		AuthServiceGRPCMTLSClientCert:   getEnv("AUTH_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/helpdesk-client.crt"),
		AuthServiceGRPCMTLSClientKey:    getEnv("AUTH_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/helpdesk-client.key"),
		
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://localhost:5672"),
		RabbitMQUser:    getEnv("RABBITMQ_USER", "rabbitmq"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "rabbitmq"),
	}

	cfg.DB = initDB(cfg)
	return cfg
}

func initDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	log.Println("Connected to Help Desk PostgreSQL database")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

