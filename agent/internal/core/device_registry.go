package core

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"os"

	"backend/agent/config"
	"backend/agent/internal/communication"
	"backend/agent/internal/security"
	"backend/agent/internal/utils"
	agentpb "backend/agent/proto"
)

// DeviceRegistry handles device registration
type DeviceRegistry struct {
	config          *config.Config
	logger          *utils.Logger
	grpcClient      *communication.Client
	credentialStore *security.CredentialStore
	deviceInfo      *utils.DeviceInfo
	deviceID        int32
}

// NewDeviceRegistry creates a new device registry
func NewDeviceRegistry(
	cfg *config.Config,
	logger *utils.Logger,
	grpcClient *communication.Client,
	credentialStore *security.CredentialStore,
) (*DeviceRegistry, error) {
	// Collect device information
	deviceInfo, err := utils.CollectDeviceInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to collect device info: %v", err)
	}

	return &DeviceRegistry{
		config:          cfg,
		logger:          logger,
		grpcClient:      grpcClient,
		credentialStore: credentialStore,
		deviceInfo:      deviceInfo,
	}, nil
}

// Register registers the device with the backend
func (dr *DeviceRegistry) Register(containerID string) error {
	dr.logger.Info("Registering device", map[string]interface{}{
		"container_id": containerID,
		"hostname":     dr.deviceInfo.Hostname,
	})

	// Generate device ID if not set
	deviceID := dr.config.DeviceID
	if deviceID == "" {
		deviceID = utils.GetDeviceID()
	}

	// Get device serial
	deviceSerial := utils.GetDeviceSerial()

	// Generate CSR
	csrPEM, privateKeyPEM, err := dr.generateCSR(deviceID, deviceSerial)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %v", err)
	}

	// Store private key securely
	if err := dr.credentialStore.Store("device_private_key", privateKeyPEM); err != nil {
		return fmt.Errorf("failed to store private key: %v", err)
	}

	// Prepare hardware info map
	hardwareInfo := make(map[string]string)
	for k, v := range dr.deviceInfo.HardwareInfo {
		hardwareInfo[k] = v
	}

	// Get primary IP address
	ipAddress := ""
	if len(dr.deviceInfo.IPAddresses) > 0 {
		ipAddress = dr.deviceInfo.IPAddresses[0]
	}

	// Call registration endpoint
	req := &agentpb.RegisterDeviceRequest{
		DeviceId:      deviceID,
		Hostname:      dr.deviceInfo.Hostname,
		OsType:        dr.deviceInfo.OSType,
		OsVersion:      dr.deviceInfo.OSVersion,
		Architecture:  dr.deviceInfo.Architecture,
		IpAddress:      ipAddress,
		DeviceSerial:  deviceSerial,
		AgentVersion:  utils.GetVersion(),
		CsrPem:        string(csrPEM),
		HardwareInfo:  hardwareInfo,
	}

	client := dr.grpcClient.GetClient()
	if client == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	ctx := context.Background()
	resp, err := client.RegisterDevice(ctx, req)
	if err != nil {
		return fmt.Errorf("registration failed: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	dr.deviceID = resp.DeviceId
	dr.logger.Info("Device registered successfully", map[string]interface{}{
		"device_id":         resp.DeviceId,
		"verification_code": resp.VerificationCode,
	})

	// Store device ID and container ID
	if err := dr.credentialStore.Store("device_id", []byte(fmt.Sprintf("%d", resp.DeviceId))); err != nil {
		dr.logger.Error("Failed to store device ID", err)
	}

	if err := dr.credentialStore.Store("container_id", []byte(containerID)); err != nil {
		dr.logger.Error("Failed to store container ID", err)
	}

	return nil
}

// Verify verifies the device with the verification code
func (dr *DeviceRegistry) Verify(verificationCode string) error {
	dr.logger.Info("Verifying device", map[string]interface{}{
		"device_id": dr.deviceID,
	})

	// Generate new CSR for certificate issuance
	deviceID := dr.config.DeviceID
	if deviceID == "" {
		deviceID = utils.GetDeviceID()
	}
	deviceSerial := utils.GetDeviceSerial()

	csrPEM, _, err := dr.generateCSR(deviceID, deviceSerial)
	if err != nil {
		return fmt.Errorf("failed to generate CSR: %v", err)
	}

	req := &agentpb.VerifyDeviceRequest{
		DeviceId:        dr.deviceID,
		VerificationCode: verificationCode,
		CsrPem:          string(csrPEM),
	}

	client := dr.grpcClient.GetClient()
	if client == nil {
		return fmt.Errorf("gRPC client not connected")
	}

	ctx := context.Background()
	resp, err := client.VerifyDevice(ctx, req)
	if err != nil {
		return fmt.Errorf("verification failed: %v", err)
	}

	if !resp.Success {
		return fmt.Errorf("verification failed: %s", resp.Message)
	}

	// Store certificate
	if err := dr.storeCertificate(resp.CertificatePem, resp.CertificateSerial); err != nil {
		return fmt.Errorf("failed to store certificate: %v", err)
	}

	// Store container ID from response
	dr.config.ContainerID = resp.ContainerId
	if err := dr.credentialStore.Store("container_id", []byte(resp.ContainerId)); err != nil {
		dr.logger.Error("Failed to store container ID", err)
	}

	dr.logger.Info("Device verified successfully", map[string]interface{}{
		"certificate_serial": resp.CertificateSerial,
		"container_id":       resp.ContainerId,
	})

	return nil
}

// generateCSR generates a certificate signing request
func (dr *DeviceRegistry) generateCSR(deviceID, deviceSerial string) ([]byte, []byte, error) {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %v", err)
	}

	// Create CSR template
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   deviceSerial,
			Organization: []string{dr.config.ContainerID},
		},
		DNSNames: []string{deviceID},
	}

	// Create CSR
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create CSR: %v", err)
	}

	// Encode CSR to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrBytes,
	})

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return csrPEM, privateKeyPEM, nil
}

// storeCertificate stores the certificate securely
func (dr *DeviceRegistry) storeCertificate(certPEM, certSerial string) error {
	// Store certificate
	certPath := fmt.Sprintf("%s/device.crt", dr.config.CertificateDir)
	if err := os.WriteFile(certPath, []byte(certPEM), 0600); err != nil {
		return fmt.Errorf("failed to write certificate: %v", err)
	}

	// Store certificate serial
	if err := dr.credentialStore.Store("certificate_serial", []byte(certSerial)); err != nil {
		return fmt.Errorf("failed to store certificate serial: %v", err)
	}

	// Update config
	dr.config.ClientCertPath = certPath

	dr.logger.Info("Certificate stored", map[string]interface{}{
		"cert_path": certPath,
		"serial":    certSerial,
	})

	return nil
}

// GetDeviceID returns the registered device ID
func (dr *DeviceRegistry) GetDeviceID() int32 {
	return dr.deviceID
}

