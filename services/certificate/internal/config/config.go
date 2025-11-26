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

	// step-ca Connection
	StepCAURL          string
	StepCAToken        string
	StepCAProvisioner  string
	StepCAUseMTLS      bool
	StepCARootCA       string
	StepCACert         string
	StepCAKey          string
	StepCAManageDocker bool

	// JWT settings (for Gateway authentication)
	JWTSecret string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		DBHost:             getEnv("CERTIFICATE_DB_HOST", "localhost"),
		DBPort:             getEnv("CERTIFICATE_DB_PORT", "5432"),
		DBName:             getEnv("CERTIFICATE_DB_NAME", "certificate_db"),
		DBUser:             getEnv("CERTIFICATE_DB_USER", "postgres"),
		DBPassword:         getEnv("CERTIFICATE_DB_PASSWORD", "postgres"),
		Port:               getEnv("CERTIFICATE_PORT", "8004"),
		GRPCPort:           getEnv("CERTIFICATE_GRPC_PORT", "9004"),
		StepCAURL:          getEnv("STEP_CA_URL", "https://step-ca:9000"),
		StepCAToken:        getEnv("STEP_CA_TOKEN", ""),
		StepCAProvisioner:  getEnv("STEP_CA_PROVISIONER", ""),
		StepCAUseMTLS:      getBoolEnv("STEP_CA_USE_MTLS", false),
		StepCARootCA:       getEnv("STEP_CA_ROOT_CA", ""),
		StepCACert:         getEnv("STEP_CA_CERT", ""),
		StepCAKey:          getEnv("STEP_CA_KEY", ""),
		StepCAManageDocker: getBoolEnv("STEP_CA_MANAGE_DOCKER", false),
		JWTSecret:          getEnv("JWT_SECRET", ""),
	}

	if cfg.StepCAToken == "" {
		log.Printf("Warning: STEP_CA_TOKEN is not set. step-ca authentication may fail.")
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

	log.Println("Connected to Certificate PostgreSQL database")
	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		val, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return val
	}
	return defaultValue
}
