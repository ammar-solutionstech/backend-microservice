package services

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"backend/services/client-container/internal/config"
	"backend/services/client-container/internal/models"

	"gorm.io/gorm"
)

// DeviceService manages device registration and certificates
type DeviceService struct {
	db                *gorm.DB
	certClient        *CertificateClient
	verificationSvc   *VerificationService
	notificationClient *NotificationClient
	config            *config.Config
}

// NewDeviceService creates a new device service
func NewDeviceService(cfg *config.Config, db *gorm.DB, certClient *CertificateClient, verificationSvc *VerificationService, notificationClient *NotificationClient) *DeviceService {
	return &DeviceService{
		db:                db,
		certClient:        certClient,
		verificationSvc:   verificationSvc,
		notificationClient: notificationClient,
		config:            cfg,
	}
}

// RegisterDevice registers a new device and generates verification code
func (s *DeviceService) RegisterDevice(containerID int, deviceInfo map[string]interface{}, csrPEM string) (*models.Device, string, error) {
	// Get container
	var container models.ClientContainer
	if err := s.db.First(&container, containerID).Error; err != nil {
		return nil, "", fmt.Errorf("container not found: %v", err)
	}

	// Extract device information
	deviceID, _ := deviceInfo["device_id"].(string)
	hostname, _ := deviceInfo["hostname"].(string)
	osType, _ := deviceInfo["os_type"].(string)
	osVersion, _ := deviceInfo["os_version"].(string)
	architecture, _ := deviceInfo["architecture"].(string)
	ipAddress, _ := deviceInfo["ip_address"].(string)
	deviceSerial, _ := deviceInfo["device_serial"].(string)

	if deviceSerial == "" {
		return nil, "", fmt.Errorf("device_serial is required")
	}

	// Check if device already exists
	var existing models.Device
	if err := s.db.Where("container_id = ? AND device_id = ?", containerID, deviceID).First(&existing).Error; err == nil {
		return nil, "", fmt.Errorf("device with ID '%s' already exists", deviceID)
	}

	// Create device record
	device := &models.Device{
		ContainerID:     containerID,
		DeviceID:        deviceID,
		Hostname:        hostname,
		OSType:          osType,
		OSVersion:       osVersion,
		Architecture:    architecture,
		IPAddress:       ipAddress,
		DeviceSerial:    deviceSerial,
		Status:          "pending", // Pending until certificate is issued
		RegisteredAt:    time.Now(),
	}

	if err := s.db.Create(device).Error; err != nil {
		return nil, "", fmt.Errorf("failed to create device: %v", err)
	}

	// Generate verification code
	code, err := s.verificationSvc.GenerateVerificationCode(device.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate verification code: %v", err)
	}

	// Send verification code to organization admin
	if err := s.notificationClient.SendVerificationCode(container.AdminEmail, func() string {
		if container.AdminPhone != nil {
			return *container.AdminPhone
		}
		return ""
	}(), code); err != nil {
		// Log error but don't fail registration
		fmt.Printf("Warning: Failed to send verification code: %v\n", err)
	}

	return device, code, nil
}

// VerifyAndIssueCertificate verifies the code and issues certificate
func (s *DeviceService) VerifyAndIssueCertificate(deviceID int, verificationCode string) (*models.AgentCertificate, error) {
	// Get device
	var device models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	// Get container
	var container models.ClientContainer
	if err := s.db.First(&container, device.ContainerID).Error; err != nil {
		return nil, fmt.Errorf("container not found: %v", err)
	}

	// Validate verification code
	valid, err := s.verificationSvc.ValidateVerificationCode(deviceID, verificationCode)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %v", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid verification code")
	}

	// Get the CSR from device (stored during registration or passed separately)
	// For now, we'll need to get it from the registration request
	// In a real implementation, you might store the CSR temporarily
	// For this implementation, we assume the CSR is provided again or stored

	// Note: CSR should be retrieved from registration or provided
	// For now, we'll need to modify this to accept CSR
	// This is a placeholder - actual implementation should store CSR during registration
	// Use VerifyAndIssueCertificateWithCSR instead which accepts CSR as parameter
	return nil, fmt.Errorf("CSR must be provided - use VerifyAndIssueCertificateWithCSR instead")
}

// VerifyAndIssueCertificateWithCSR verifies code and issues certificate with provided CSR
func (s *DeviceService) VerifyAndIssueCertificateWithCSR(deviceID int, verificationCode, csrPEM string) (*models.AgentCertificate, error) {
	// Get device
	var device models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		return nil, fmt.Errorf("device not found: %v", err)
	}

	// Get container
	var container models.ClientContainer
	if err := s.db.First(&container, device.ContainerID).Error; err != nil {
		return nil, fmt.Errorf("container not found: %v", err)
	}

	// Validate verification code
	valid, err := s.verificationSvc.ValidateVerificationCode(deviceID, verificationCode)
	if err != nil {
		return nil, fmt.Errorf("verification failed: %v", err)
	}
	if !valid {
		return nil, fmt.Errorf("invalid verification code")
	}

	// Request certificate from container-management service
	containerInfo := map[string]string{
		"container_id": container.ContainerID,
		"org_id":        fmt.Sprintf("%d", container.OrganizationID),
		"org_domain":    "", // Can be retrieved from organization if needed
	}

	certResp, err := s.certClient.RequestAgentCertificate(container.ContainerID, device.DeviceSerial, csrPEM, containerInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to request certificate: %v", err)
	}

	// Parse certificate to get expiration
	block, _ := pem.Decode([]byte(certResp.CertificatePEM))
	if block == nil {
		return nil, fmt.Errorf("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Store certificate
	agentCert := &models.AgentCertificate{
		DeviceID:          deviceID,
		CertificateSerial: certResp.SerialNumber,
		CertificatePEM:    certResp.CertificatePEM,
		IssuedAt:          cert.NotBefore,
		ExpiresAt:         cert.NotAfter,
		Status:            "active",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.db.Create(agentCert).Error; err != nil {
		return nil, fmt.Errorf("failed to store certificate: %v", err)
	}

	// Update device with certificate serial and status
	device.CertificateSerial = &certResp.SerialNumber
	device.Status = "active"
	device.UpdatedAt = time.Now()
	if err := s.db.Save(&device).Error; err != nil {
		return nil, fmt.Errorf("failed to update device: %v", err)
	}

	return agentCert, nil
}

// GetDevice retrieves a device by ID
func (s *DeviceService) GetDevice(deviceID int) (*models.Device, error) {
	var device models.Device
	if err := s.db.First(&device, deviceID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("failed to get device: %v", err)
	}

	return &device, nil
}

// GetDeviceByContainerAndDeviceID retrieves a device by container and device ID
func (s *DeviceService) GetDeviceByContainerAndDeviceID(containerID int, deviceIDStr string) (*models.Device, error) {
	var device models.Device
	if err := s.db.Where("container_id = ? AND device_id = ?", containerID, deviceIDStr).First(&device).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("failed to get device: %v", err)
	}

	return &device, nil
}

// UpdateDeviceStatus updates device status
func (s *DeviceService) UpdateDeviceStatus(deviceID int, status string) error {
	if err := s.db.Model(&models.Device{}).
		Where("id = ?", deviceID).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to update device status: %v", err)
	}

	return nil
}

// RevokeDevice revokes a device and its certificate
func (s *DeviceService) RevokeDevice(deviceID int) error {
	device, err := s.GetDevice(deviceID)
	if err != nil {
		return err
	}

	// Revoke certificate if exists
	if device.CertificateSerial != nil {
		if err := s.certClient.RevokeCertificate(*device.CertificateSerial, 0); err != nil {
			return fmt.Errorf("failed to revoke certificate: %v", err)
		}

		// Update certificate status
		s.db.Model(&models.AgentCertificate{}).
			Where("device_id = ?", deviceID).
			Updates(map[string]interface{}{
				"status":     "revoked",
				"revoked_at": time.Now(),
				"updated_at": time.Now(),
			})
	}

	// Update device status
	device.Status = "revoked"
	device.UpdatedAt = time.Now()
	if err := s.db.Save(device).Error; err != nil {
		return fmt.Errorf("failed to update device: %v", err)
	}

	return nil
}

// ListDevices lists devices for a container
func (s *DeviceService) ListDevices(containerID int, status string) ([]models.Device, error) {
	var devices []models.Device
	query := s.db.Where("container_id = ?", containerID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&devices).Error; err != nil {
		return nil, fmt.Errorf("failed to list devices: %v", err)
	}

	return devices, nil
}

// ValidateDeviceCertificate validates that certificate serial matches device serial
func (s *DeviceService) ValidateDeviceCertificate(deviceID int, certSerial string) (bool, error) {
	device, err := s.GetDevice(deviceID)
	if err != nil {
		return false, err
	}

	// Get certificate
	var agentCert models.AgentCertificate
	if err := s.db.Where("device_id = ? AND certificate_serial = ?", deviceID, certSerial).First(&agentCert).Error; err != nil {
		return false, fmt.Errorf("certificate not found")
	}

	// Parse certificate to extract device serial
	block, _ := pem.Decode([]byte(agentCert.CertificatePEM))
	if block == nil {
		return false, fmt.Errorf("failed to decode certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Check if CN or SAN contains device serial
	if cert.Subject.CommonName == device.DeviceSerial {
		return true, nil
	}

	// Check SANs
	for _, dnsName := range cert.DNSNames {
		if dnsName == device.DeviceSerial {
			return true, nil
		}
	}

	return false, fmt.Errorf("certificate serial does not match device serial")
}

// UpdateDeviceLastSeen updates the last seen timestamp for a device
func (s *DeviceService) UpdateDeviceLastSeen(deviceID int) error {
	now := time.Now()
	if err := s.db.Model(&models.Device{}).
		Where("id = ?", deviceID).
		Update("last_seen", now).Error; err != nil {
		return fmt.Errorf("failed to update last_seen: %v", err)
	}

	return nil
}

