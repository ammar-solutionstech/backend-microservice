package config

import (
	"crypto/tls"
	"crypto/x509"
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
	StepCAAudience     string
	StepCAToken        string
	StepCAProvisioner  string
	StepCAUseMTLS      bool
	StepCARootCA       string
	StepCACert         string
	StepCAKey          string
	StepCAKeyPath      string
	StepCAKeyPassword  string
	StepCAManageDocker bool

	// mTLS Server Configuration (HTTP)
	MTLSCACert        string
	MTLSServerCert    string
	MTLSServerKey     string
	MTLSServerKeyPass string
	MTLSTLSConfig     *tls.Config

	// gRPC mTLS Server Configuration
	GRPCMTLSCACert        string
	GRPCMTLSServerCert    string
	GRPCMTLSServerKey     string
	GRPCMTLSServerKeyPass string

	// JWT settings (for Gateway authentication)
	JWTSecret string
}

type Keys struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Y   string `json:"y"`
	D   string `json:"d"` // هذا هو المفتاح الخاص الفعلي
	Use string `json:"use"`
}

func Load() *Config {
	// Try service-specific .env first
	_ = godotenv.Load(".env")
	// Fall back to root .env if exists
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env")
	}

	cfg := &Config{
		DBHost:             getEnv("CERTIFICATE_DB_HOST", "localhost"),
		DBPort:             getEnv("CERTIFICATE_DB_PORT", "5432"),
		DBName:             getEnv("CERTIFICATE_DB_NAME", "certificate_db"),
		DBUser:             getEnv("CERTIFICATE_DB_USER", "postgres"),
		DBPassword:         getEnv("CERTIFICATE_DB_PASSWORD", "postgres"),
		Port:               getEnv("CERTIFICATE_PORT", "8004"),
		GRPCPort:           getEnv("CERTIFICATE_GRPC_PORT", "9004"),
		StepCAURL:          getEnv("STEP_CA_URL", "https://step-ca:9000"),
		StepCAAudience:     getEnv("STEP_CA_AUDIANCE", ""),
		StepCAToken:        getEnv("STEP_CA_TOKEN", ""),
		StepCAProvisioner:  getEnv("STEP_CA_PROVISIONER", ""),
		StepCAUseMTLS:      getBoolEnv("STEP_CA_USE_MTLS", false),
		StepCARootCA:       getEnv("STEP_CA_ROOT_CA", ""),
		StepCACert:         getEnv("STEP_CA_CERT", ""),
		StepCAKey:          getEnv("STEP_CA_KEY", ""),
		StepCAManageDocker: getBoolEnv("STEP_CA_MANAGE_DOCKER", false),
		StepCAKeyPath:      getEnv("STEP_CA_KEY_PATH", ""),
		StepCAKeyPassword:  getEnv("STEP_CA_KEY_PASSWORD", ""),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		MTLSCACert:         getEnv("MTLS_CA_CERT", "./certs/certificate-ca.crt"),
		MTLSServerCert:     getEnv("CERTIFICATE_SERVER_CERT", "./certs/certificate-server.crt"),
		MTLSServerKey:      getEnv("CERTIFICATE_SERVER_KEY", "./certs/certificate-server.key"),
		MTLSServerKeyPass:  getEnv("CERTIFICATE_SERVER_KEY_PASSWORD", ""),
		
		// gRPC mTLS Server Configuration
		GRPCMTLSCACert:        getEnv("GRPC_MTLS_CA_CERT", "./certs/certificate-ca.crt"),
		GRPCMTLSServerCert:    getEnv("GRPC_MTLS_SERVER_CERT", "./certs/certificate-server.crt"),
		GRPCMTLSServerKey:     getEnv("GRPC_MTLS_SERVER_KEY", "./certs/certificate-server.key"),
		GRPCMTLSServerKeyPass: getEnv("GRPC_MTLS_SERVER_KEY_PASSWORD", ""),
	}

	if cfg.StepCAToken == "" {
		log.Printf("Warning: STEP_CA_TOKEN is not set. step-ca authentication may fail.")
	}

	cfg.DB = initDB(cfg)

	// Initialize mTLS configs
	cfg.MTLSTLSConfig = initServerTLS(cfg)
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

func initServerTLS(cfg *Config) *tls.Config {
	if cfg.MTLSCACert == "" || cfg.MTLSServerCert == "" || cfg.MTLSServerKey == "" {
		log.Printf("Warning: Server mTLS certificates not configured. Server will not use mTLS.")
		return nil
	}

	// Load CA cert for client verification
	caCert, err := os.ReadFile(cfg.MTLSCACert)
	if err != nil {
		log.Fatalf("failed to read server mTLS CA cert: %v", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to parse server mTLS CA cert")
	}

	// Load server certificate and key
	var cert tls.Certificate
	if cfg.MTLSServerKeyPass != "" {
		// Password-protected key handling
		cert, err = tls.LoadX509KeyPair(cfg.MTLSServerCert, cfg.MTLSServerKey)
		if err != nil {
			log.Fatalf("failed to load server mTLS cert/key: %v", err)
		}
	} else {
		cert, err = tls.LoadX509KeyPair(cfg.MTLSServerCert, cfg.MTLSServerKey)
		if err != nil {
			log.Fatalf("failed to load server mTLS cert/key: %v", err)
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAnyClientCert, //tls.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
	}
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
