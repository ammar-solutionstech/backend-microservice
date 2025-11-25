package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
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

	// JWT
	JWTSecret        string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string

	// RabbitMQ
	RabbitMQURL      string
	RabbitMQUser     string
	RabbitMQPassword string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		DBHost:           getEnv("AUTH_DB_HOST", "localhost"),
		DBPort:           getEnv("AUTH_DB_PORT", "5432"),
		DBName:           getEnv("AUTH_DB_NAME", "auth_db"),
		DBUser:           getEnv("AUTH_DB_USER", "postgres"),
		DBPassword:       getEnv("AUTH_DB_PASSWORD", "postgres"),
		Port:             getEnv("AUTH_PORT", "8001"),
		GRPCPort:         getEnv("AUTH_GRPC_PORT", "9001"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRY", "7d")),
		JWTIssuer:        getEnv("JWT_ISSUER", "auth-service"),
		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://localhost:5672"),
		RabbitMQUser:     getEnv("RABBITMQ_USER", "rabbitmq"),
		RabbitMQPassword: getEnv("RABBITMQ_PASSWORD", "rabbitmq"),
	}

	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
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

	log.Println("Connected to Auth PostgreSQL database")
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

func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}
