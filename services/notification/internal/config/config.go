package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

	// RabbitMQ
	RabbitMQURL      string
	RabbitMQUser     string
	RabbitMQPassword string

	// SMTP
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPTLS      bool
}

func Load() *Config {
	// Try service-specific .env first
	_ = godotenv.Load(".env")
	// Fall back to root .env if exists
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env")
	}

	cfg := &Config{
		DBHost:          getEnv("NOTIFICATION_DB_HOST", "localhost"),
		DBPort:          getEnv("NOTIFICATION_DB_PORT", "5432"),
		DBName:          getEnv("NOTIFICATION_DB_NAME", "notification_db"),
		DBUser:          getEnv("NOTIFICATION_DB_USER", "postgres"),
		DBPassword:      getEnv("NOTIFICATION_DB_PASSWORD", "postgres"),
		Port:            getEnv("NOTIFICATION_PORT", "8003"),
		GRPCPort:        getEnv("NOTIFICATION_GRPC_PORT", "9003"),
		
		// gRPC mTLS Server Configuration
		GRPCMTLSCACert:        getEnv("GRPC_MTLS_CA_CERT", "./certs/notification-ca.crt"),
		GRPCMTLSServerCert:    getEnv("GRPC_MTLS_SERVER_CERT", "./certs/notification-server.crt"),
		GRPCMTLSServerKey:     getEnv("GRPC_MTLS_SERVER_KEY", "./certs/notification-server.key"),
		GRPCMTLSServerKeyPass: getEnv("GRPC_MTLS_SERVER_KEY_PASSWORD", ""),
		
		RabbitMQURL:     getEnv("RABBITMQ_URL", "amqp://localhost:5672"),
		RabbitMQUser:    getEnv("RABBITMQ_USER", "rabbitmq"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "rabbitmq"),
		SMTPHost:        getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:        parseInt(getEnv("SMTP_PORT", "587")),
		SMTPUser:        getEnv("SMTP_USER", ""),
		SMTPPassword:    getEnv("SMTP_PASSWORD", ""),
		SMTPTLS:         getEnv("SMTP_TLS", "true") == "true",
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

	log.Println("Connected to Notification PostgreSQL database")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 587
	}
	return val
}

