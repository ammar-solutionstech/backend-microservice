package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server
	Port string

	// gRPC Services
	AuthServiceGRPC         string
	HelpDeskServiceGRPC     string
	NotificationServiceGRPC string
	InventoryServiceGRPC    string
	GeographyServiceGRPC     string
	NavigationServiceGRPC    string

	// gRPC mTLS Client Configuration
	// Auth Service
	AuthServiceGRPCMTLSCA      string
	AuthServiceGRPCMTLSClientCert string
	AuthServiceGRPCMTLSClientKey  string
	
	// Helpdesk Service
	HelpdeskServiceGRPCMTLSCA      string
	HelpdeskServiceGRPCMTLSClientCert string
	HelpdeskServiceGRPCMTLSClientKey  string
	
	// Notification Service
	NotificationServiceGRPCMTLSCA      string
	NotificationServiceGRPCMTLSClientCert string
	NotificationServiceGRPCMTLSClientKey  string
	
	// Inventory Service
	InventoryServiceGRPCMTLSCA      string
	InventoryServiceGRPCMTLSClientCert string
	InventoryServiceGRPCMTLSClientKey  string
	
	// Geography Service
	GeographyServiceGRPCMTLSCA      string
	GeographyServiceGRPCMTLSClientCert string
	GeographyServiceGRPCMTLSClientKey  string
	
	// Navigation Service
	NavigationServiceGRPCMTLSCA      string
	NavigationServiceGRPCMTLSClientCert string
	NavigationServiceGRPCMTLSClientKey  string

	// HTTP Services (for REST API proxying)
	AuthServiceHTTP string
	InventoryServiceHTTP string
	GeographyServiceHTTP string
	NavigationServiceHTTP string

}

func Load() *Config {
	// Try service-specific .env first
	_ = godotenv.Load(".env")
	// Fall back to root .env if exists
	if _, err := os.Stat("../../.env"); err == nil {
		_ = godotenv.Overload("../../.env")
	}

	cfg := &Config{
		Port:                    getEnv("GATEWAY_PORT", "8080"),
		AuthServiceGRPC:         getEnv("AUTH_SERVICE_GRPC", "localhost:9001"),
		HelpDeskServiceGRPC:     getEnv("HELPDESK_SERVICE_GRPC", "localhost:9002"),
		NotificationServiceGRPC: getEnv("NOTIFICATION_SERVICE_GRPC", "localhost:9003"),
		InventoryServiceGRPC:    getEnv("INVENTORY_SERVICE_GRPC", "localhost:9007"),
		GeographyServiceGRPC:     getEnv("GEOGRAPHY_SERVICE_GRPC", "localhost:9008"),
		NavigationServiceGRPC:    getEnv("NAVIGATION_SERVICE_GRPC", "localhost:9009"),
		
		// Auth Service gRPC mTLS
		AuthServiceGRPCMTLSCA:           getEnv("AUTH_SERVICE_GRPC_MTLS_CA", "./certs/auth-ca.crt"),
		AuthServiceGRPCMTLSClientCert:    getEnv("AUTH_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/gateway-client.crt"),
		AuthServiceGRPCMTLSClientKey:   getEnv("AUTH_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/gateway-client.key"),
		
		// Helpdesk Service gRPC mTLS
		HelpdeskServiceGRPCMTLSCA:           getEnv("HELPDESK_SERVICE_GRPC_MTLS_CA", "./certs/helpdesk-ca.crt"),
		HelpdeskServiceGRPCMTLSClientCert:   getEnv("HELPDESK_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/gateway-client.crt"),
		HelpdeskServiceGRPCMTLSClientKey:    getEnv("HELPDESK_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/gateway-client.key"),
		
		// Notification Service gRPC mTLS
		NotificationServiceGRPCMTLSCA:           getEnv("NOTIFICATION_SERVICE_GRPC_MTLS_CA", "./certs/notification-ca.crt"),
		NotificationServiceGRPCMTLSClientCert:   getEnv("NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/gateway-client.crt"),
		NotificationServiceGRPCMTLSClientKey:    getEnv("NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/gateway-client.key"),
		
		// Inventory Service gRPC mTLS
		InventoryServiceGRPCMTLSCA:           getEnv("INVENTORY_SERVICE_GRPC_MTLS_CA", "./certs/inventory-ca.crt"),
		InventoryServiceGRPCMTLSClientCert:   getEnv("INVENTORY_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/gateway-client.crt"),
		InventoryServiceGRPCMTLSClientKey:    getEnv("INVENTORY_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/gateway-client.key"),
		
		// Geography Service gRPC mTLS
		GeographyServiceGRPCMTLSCA:           getEnv("GEOGRAPHY_SERVICE_GRPC_MTLS_CA", "./certs/geography-ca.crt"),
		GeographyServiceGRPCMTLSClientCert:   getEnv("GEOGRAPHY_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/gateway-client.crt"),
		GeographyServiceGRPCMTLSClientKey:    getEnv("GEOGRAPHY_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/gateway-client.key"),
		
		// Navigation Service gRPC mTLS
		NavigationServiceGRPCMTLSCA:           getEnv("NAVIGATION_SERVICE_GRPC_MTLS_CA", "./certs/navigation-ca.crt"),
		NavigationServiceGRPCMTLSClientCert:   getEnv("NAVIGATION_SERVICE_GRPC_MTLS_CLIENT_CERT", "./certs/gateway-client.crt"),
		NavigationServiceGRPCMTLSClientKey:    getEnv("NAVIGATION_SERVICE_GRPC_MTLS_CLIENT_KEY", "./certs/gateway-client.key"),
		
		// Determine HTTP URLs based on gRPC addresses
		AuthServiceHTTP:  getServiceHTTP("AUTH_SERVICE_GRPC", "auth-service:8001", "localhost:8001"),
		InventoryServiceHTTP: getServiceHTTP("INVENTORY_SERVICE_GRPC", "inventory-service:8007", "localhost:8007"),
		GeographyServiceHTTP: getServiceHTTP("GEOGRAPHY_SERVICE_GRPC", "geography-service:8008", "localhost:8008"),
		NavigationServiceHTTP: getServiceHTTP("NAVIGATION_SERVICE_GRPC", "navigation-service:8009", "localhost:8009"),
	}

	return cfg
}

func getServiceHTTP(envKey, dockerService, localhost string) string {
	grpcAddr := getEnv(envKey, "")
	// If using Docker service name, use HTTP service name; otherwise localhost
	if strings.Contains(grpcAddr, strings.Split(dockerService, ":")[0]) {
		return "http://" + dockerService
	}
	return "http://" + localhost
}


func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
