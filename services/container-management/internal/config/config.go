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
	Port string

	// Certificate Service Connection (mTLS)
	CertificateServiceURL       string
	CertificateServiceCA        string
	CertificateServiceCert      string
	CertificateServiceKey       string
	CertificateServiceKeyPass   string
	CertificateServiceTLSConfig *tls.Config

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

	// gRPC mTLS Client Configuration (optional, for Certificate service)
	CertificateServiceGRPCMTLSCA      string
	CertificateServiceGRPCMTLSClientCert string
	CertificateServiceGRPCMTLSClientKey  string

	// Bootstrap Token
	BootstrapTokenSecret string

	// CSR Validation Rules
	CSRRequiredOrg     string
	CSRRequiredCountry string

	// Docker Configuration
	DockerHost string

	// Client Container Configuration
	ClientContainerImage      string
	ClientContainerImageTag   string
	ClientContainerPort       string
	ClientContainerGRPCPort   string
	ClientContainerDBHost     string
	ClientContainerDBPort     string
	ClientContainerDBName     string
	ClientContainerDBUser     string
	ClientContainerDBPassword string
	ContainerMgmtServiceURL   string
	NotificationServiceGRPC   string
	CertificatesPath          string
	DockerNetwork             string
}

func Load() *Config {
	// Try service-specific .env first, then root .env as fallback
	_ = godotenv.Load(".env")
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env") // Root .env as fallback
	}

	cfg := &Config{
		DBHost:                    getEnv("CONTAINER_MGMT_DB_HOST", "localhost"),
		DBPort:                    getEnv("CONTAINER_MGMT_DB_PORT", "5432"),
		DBName:                    getEnv("CONTAINER_MGMT_DB_NAME", "container_mgmt_db"),
		DBUser:                    getEnv("CONTAINER_MGMT_DB_USER", "postgres"),
		DBPassword:                getEnv("CONTAINER_MGMT_DB_PASSWORD", "postgres"),
		Port:                      getEnv("CONTAINER_MGMT_PORT", "8005"),
		CertificateServiceURL:     getEnv("CERTIFICATE_SERVICE_URL", "https://certificate-service:8004"),
		CertificateServiceCA:      getEnv("CERTIFICATE_SERVICE_CA", "./certs/certificate-ca.crt"),
		CertificateServiceCert:    getEnv("CERTIFICATE_SERVICE_CERT", "./certs/container-management-client.crt"),
		CertificateServiceKey:     getEnv("CERTIFICATE_SERVICE_KEY", "./certs/container-management-client.key"),
		CertificateServiceKeyPass:  getEnv("CERTIFICATE_SERVICE_KEY_PASSWORD", ""),
		MTLSCACert:                getEnv("MTLS_CA_CERT", "./certs/container-management-ca.crt"),
		MTLSServerCert:            getEnv("MTLS_SERVER_CERT", "./certs/container-management-server.crt"),
		MTLSServerKey:             getEnv("MTLS_SERVER_KEY", "./certs/container-management-server.key"),
		MTLSServerKeyPass:         getEnv("MTLS_SERVER_KEY_PASSWORD", ""),
		
		// gRPC mTLS Server Configuration
		GRPCMTLSCACert:        getEnv("GRPC_MTLS_CA_CERT", "./certs/container-management-ca.crt"),
		GRPCMTLSServerCert:    getEnv("GRPC_MTLS_SERVER_CERT", "./certs/container-management-server.crt"),
		GRPCMTLSServerKey:     getEnv("GRPC_MTLS_SERVER_KEY", "./certs/container-management-server.key"),
		GRPCMTLSServerKeyPass: getEnv("GRPC_MTLS_SERVER_KEY_PASSWORD", ""),
		
		// Certificate Service gRPC mTLS Client Configuration (optional)
		CertificateServiceGRPCMTLSCA:           getEnv("CERTIFICATE_SERVICE_GRPC_MTLS_CA", "./certs/certificate-ca.crt"),
		CertificateServiceGRPCMTLSClientCert:   getEnv("CERTIFICATE_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/container-management-client.crt"),
		CertificateServiceGRPCMTLSClientKey:    getEnv("CERTIFICATE_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/container-management-client.key"),
		
		BootstrapTokenSecret:      getEnv("BOOTSTRAP_TOKEN_SECRET", ""),
		CSRRequiredOrg:            getEnv("CSR_REQUIRED_ORG", ""),
		CSRRequiredCountry:        getEnv("CSR_REQUIRED_COUNTRY", ""),
		DockerHost:                getEnv("DOCKER_HOST", "npipe:////./pipe/docker_engine"),
		ClientContainerImage:      getEnv("CLIENT_CONTAINER_IMAGE", "backend_v1-client-container"),
		ClientContainerImageTag:   getEnv("CLIENT_CONTAINER_IMAGE_TAG", "latest"),
		ClientContainerPort:       getEnv("CLIENT_CONTAINER_PORT", "8006"),
		ClientContainerGRPCPort:   getEnv("CLIENT_CONTAINER_GRPC_PORT", "9006"),
		ClientContainerDBHost:     getEnv("CLIENT_CONTAINER_DB_HOST", "localhost"),
		ClientContainerDBPort:     getEnv("CLIENT_CONTAINER_DB_PORT", "5432"),
		ClientContainerDBName:     getEnv("CLIENT_CONTAINER_DB_NAME", "client_container_db"),
		ClientContainerDBUser:     getEnv("CLIENT_CONTAINER_DB_USER", "postgres"),
		ClientContainerDBPassword: getEnv("CLIENT_CONTAINER_DB_PASSWORD", "postgres"),
		ContainerMgmtServiceURL:   getEnv("CONTAINER_MGMT_SERVICE_URL", "https://container-management-service:8005"),
		NotificationServiceGRPC:   getEnv("NOTIFICATION_SERVICE_GRPC", "notification-service:9003"),
		CertificatesPath:          getEnv("CERTIFICATES_PATH", "./certs"),
		DockerNetwork:             getEnv("DOCKER_NETWORK", "microservices-network"),
	}

	if cfg.BootstrapTokenSecret == "" {
		log.Printf("Warning: BOOTSTRAP_TOKEN_SECRET is not set. Bootstrap token security may be compromised.")
	}

	// Initialize database
	cfg.DB = initDB(cfg)

	// Initialize mTLS configs
	cfg.CertificateServiceTLSConfig = initCertificateServiceTLS(cfg)
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

	log.Println("Connected to Container Management PostgreSQL database")
	return db
}

func initCertificateServiceTLS(cfg *Config) *tls.Config {
	//log.Printf("Certificate Service CA: %s , Certificate Service Cert: %s , Certificate Service Key: %s", cfg.CertificateServiceCA, cfg.CertificateServiceCert, cfg.CertificateServiceKey)
	if cfg.CertificateServiceCA == "" || cfg.CertificateServiceCert == "" || cfg.CertificateServiceKey == "" {
		log.Printf("Warning: Certificate service mTLS certificates not configured. mTLS will not be used.")
		return nil
	}

	// Load CA cert for server verification
	caCert, err := os.ReadFile(cfg.CertificateServiceCA)
	if err != nil {
		log.Fatalf("failed to read certificate service CA cert: %v", err)
	}
	//log.Printf("Certificate Service CA:%s", string(caCert))

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		log.Fatalf("failed to parse certificate service CA cert")
	}

	// Load client certificate and key
	var cert tls.Certificate
	if cfg.CertificateServiceKeyPass != "" {
		// If password-protected key, we need to decrypt it
		// For now, assume key is not password-protected or use a library
		cert, err = tls.LoadX509KeyPair(cfg.CertificateServiceCert, cfg.CertificateServiceKey)
		if err != nil {
			log.Fatalf("failed to load certificate service client cert/key: %v", err)
		}
	} else {
		cert, err = tls.LoadX509KeyPair(cfg.CertificateServiceCert, cfg.CertificateServiceKey)
		if err != nil {
			log.Fatalf("failed to load certificate service client cert/key: %v", err)
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientAuth:   tls.RequireAnyClientCert, //.RequireAndVerifyClientCert,
		ClientCAs:    caCertPool,
		//ServerName:         "", // Will be set based on URL
		InsecureSkipVerify: true,
	}
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
		Certificates:       []tls.Certificate{cert},
		ClientAuth:         tls.RequireAnyClientCert, //RequireAndVerifyClientCert,
		ClientCAs:          caCertPool,
		InsecureSkipVerify: true,
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
