package config

import (
	"crypto/tls"
	"crypto/x509"
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

	// Container Management Service Connection (for certificate operations)
	ContainerMgmtServiceURL      string
	ContainerMgmtServiceCA       string
	ContainerMgmtServiceCert     string
	ContainerMgmtServiceKey       string
	ContainerMgmtServiceKeyPass  string
	ContainerMgmtServiceTLSConfig *tls.Config

	// Container's own certificate (obtained from container-management)
	ContainerCertPath string
	ContainerKeyPath  string
	ContainerCert     *tls.Certificate

	// Agent certificate validation
	AgentCACertPath string
	AgentCACertPool *x509.CertPool

	// Notification Service Connection (for sending verification codes)
	NotificationServiceGRPC string

	// Verification code settings
	VerificationCodeLength int
	VerificationCodeExpiry  time.Duration

	// Telemetry retention (days)
	TelemetryRetentionDays int

	// Plugin repository settings
	PluginRepositoryURL string
}

func Load() *Config {
	_ = godotenv.Load(".env")

	cfg := &Config{
		DBHost:                    getEnv("CLIENT_CONTAINER_DB_HOST", "localhost"),
		DBPort:                    getEnv("CLIENT_CONTAINER_DB_PORT", "5432"),
		DBName:                    getEnv("CLIENT_CONTAINER_DB_NAME", "client_container_db"),
		DBUser:                    getEnv("CLIENT_CONTAINER_DB_USER", "postgres"),
		DBPassword:                getEnv("CLIENT_CONTAINER_DB_PASSWORD", "postgres"),
		Port:                      getEnv("CLIENT_CONTAINER_PORT", "8006"),
		GRPCPort:                  getEnv("CLIENT_CONTAINER_GRPC_PORT", "9006"),
		ContainerMgmtServiceURL:   getEnv("CONTAINER_MGMT_SERVICE_URL", "https://container-management-service:8005"),
		ContainerMgmtServiceCA:    getEnv("CONTAINER_MGMT_SERVICE_CA", ""),
		ContainerMgmtServiceCert:   getEnv("CONTAINER_MGMT_SERVICE_CERT", ""),
		ContainerMgmtServiceKey:    getEnv("CONTAINER_MGMT_SERVICE_KEY", ""),
		ContainerMgmtServiceKeyPass: getEnv("CONTAINER_MGMT_SERVICE_KEY_PASSWORD", ""),
		ContainerCertPath:         getEnv("CONTAINER_CERT_PATH", ""),
		ContainerKeyPath:          getEnv("CONTAINER_KEY_PATH", ""),
		AgentCACertPath:          getEnv("AGENT_CA_CERT_PATH", ""),
		NotificationServiceGRPC:   getEnv("NOTIFICATION_SERVICE_GRPC", "notification-service:9003"),
		VerificationCodeLength:    getIntEnv("VERIFICATION_CODE_LENGTH", 6),
		TelemetryRetentionDays:    getIntEnv("TELEMETRY_RETENTION_DAYS", 90),
		PluginRepositoryURL:      getEnv("PLUGIN_REPOSITORY_URL", ""),
	}

	// Parse verification code expiry
	expiryStr := getEnv("VERIFICATION_CODE_EXPIRY", "15m")
	expiry, err := time.ParseDuration(expiryStr)
	if err != nil {
		log.Printf("Warning: Invalid VERIFICATION_CODE_EXPIRY, using default 15m")
		cfg.VerificationCodeExpiry = 15 * time.Minute
	} else {
		cfg.VerificationCodeExpiry = expiry
	}

	// Initialize database
	cfg.DB = initDB(cfg)

	// Initialize TLS configs
	cfg.ContainerMgmtServiceTLSConfig = initContainerMgmtTLS(cfg)
	cfg.ContainerCert = loadContainerCertificate(cfg)
	cfg.AgentCACertPool = loadAgentCACert(cfg)

	return cfg
}

func initDB(cfg *Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	log.Println("Connected to Client Container PostgreSQL database")
	return db
}

func initContainerMgmtTLS(cfg *Config) *tls.Config {
	if cfg.ContainerMgmtServiceCA == "" || cfg.ContainerMgmtServiceCert == "" || cfg.ContainerMgmtServiceKey == "" {
		log.Printf("Warning: Container management service mTLS certificates not configured.")
		return nil
	}

	// Load CA cert for server verification
	caCert, err := os.ReadFile(cfg.ContainerMgmtServiceCA)
	if err != nil {
		log.Fatalf("failed to read container management service CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to parse container management service CA cert")
	}

	// Load client certificate and key
	cert, err := tls.LoadX509KeyPair(cfg.ContainerMgmtServiceCert, cfg.ContainerMgmtServiceKey)
	if err != nil {
		log.Fatalf("failed to load container management service client cert/key: %v", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
}

func loadContainerCertificate(cfg *Config) *tls.Certificate {
	if cfg.ContainerCertPath == "" || cfg.ContainerKeyPath == "" {
		log.Printf("Warning: Container certificate not configured.")
		return nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.ContainerCertPath, cfg.ContainerKeyPath)
	if err != nil {
		log.Fatalf("failed to load container certificate: %v", err)
	}

	return &cert
}

func loadAgentCACert(cfg *Config) *x509.CertPool {
	if cfg.AgentCACertPath == "" {
		log.Printf("Warning: Agent CA certificate not configured.")
		return nil
	}

	caCert, err := os.ReadFile(cfg.AgentCACertPath)
	if err != nil {
		log.Fatalf("failed to read agent CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to parse agent CA cert")
	}

	return caCertPool
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		val, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return val
	}
	return defaultValue
}

