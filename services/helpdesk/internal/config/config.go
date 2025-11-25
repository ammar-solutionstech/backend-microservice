package config

import (
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
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

	// RabbitMQ
	RabbitMQURL      string
	RabbitMQUser     string
	RabbitMQPassword string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		DBHost:          getEnv("HELPDESK_DB_HOST", "localhost"),
		DBPort:          getEnv("HELPDESK_DB_PORT", "5432"),
		DBName:          getEnv("HELPDESK_DB_NAME", "helpdesk_db"),
		DBUser:          getEnv("HELPDESK_DB_USER", "postgres"),
		DBPassword:      getEnv("HELPDESK_DB_PASSWORD", "postgres"),
		Port:            getEnv("HELPDESK_PORT", "8002"),
		GRPCPort:        getEnv("HELPDESK_GRPC_PORT", "9002"),
		AuthServiceGRPC: getEnv("AUTH_SERVICE_GRPC", "localhost:9001"),
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

