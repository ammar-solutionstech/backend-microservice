package services

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"

	"backend/services/container-management/internal/config"
	"backend/services/container-management/internal/models"
)

// ContainerOrchestrator manages Docker container creation and configuration for client-containers
type ContainerOrchestrator struct {
	dockerClient *DockerClient
	config       *config.Config
	db           interface{} // Will be used for database operations if needed
}

// NewContainerOrchestrator creates a new container orchestrator
func NewContainerOrchestrator(dockerClient *DockerClient, cfg *config.Config) *ContainerOrchestrator {
	return &ContainerOrchestrator{
		dockerClient: dockerClient,
		config:       cfg,
	}
}

// GenerateContainerName generates a unique container name
func (o *ContainerOrchestrator) GenerateContainerName(containerID string) string {
	return fmt.Sprintf("client-container-%s", containerID)
}

// ConfigureContainerEnvironment generates environment variables for client-container
func (o *ContainerOrchestrator) ConfigureContainerEnvironment(orgID int, containerID string, containerData *models.ClientContainer, certSerial string) []string {
	env := []string{
		fmt.Sprintf("CLIENT_CONTAINER_DB_HOST=%s", o.config.ClientContainerDBHost),
		fmt.Sprintf("CLIENT_CONTAINER_DB_PORT=%s", o.config.ClientContainerDBPort),
		fmt.Sprintf("CLIENT_CONTAINER_DB_NAME=%s", o.config.ClientContainerDBName),
		fmt.Sprintf("CLIENT_CONTAINER_DB_USER=%s", o.config.ClientContainerDBUser),
		fmt.Sprintf("CLIENT_CONTAINER_DB_PASSWORD=%s", o.config.ClientContainerDBPassword),
		fmt.Sprintf("CLIENT_CONTAINER_PORT=%s", o.config.ClientContainerPort),
		fmt.Sprintf("CLIENT_CONTAINER_GRPC_PORT=%s", o.config.ClientContainerGRPCPort),
		fmt.Sprintf("ORGANIZATION_ID=%d", orgID),
		fmt.Sprintf("CONTAINER_ID=%s", containerID),
		fmt.Sprintf("CONTAINER_CERT_SERIAL=%s", certSerial),
		fmt.Sprintf("CONTAINER_MGMT_SERVICE_URL=%s", o.config.ContainerMgmtServiceURL),
		fmt.Sprintf("NOTIFICATION_SERVICE_GRPC=%s", o.config.NotificationServiceGRPC),
		fmt.Sprintf("CONTAINER_CERT_PATH=/certs/container.crt"),
		fmt.Sprintf("CONTAINER_KEY_PATH=/certs/container.key"),
		fmt.Sprintf("AGENT_CA_CERT_PATH=/certs/agent-ca.crt"),
	}

	return env
}

// ConfigureContainerVolumes generates volume mounts for client-container
func (o *ContainerOrchestrator) ConfigureContainerVolumes(containerID string) []mount.Mount {
	volumes := []mount.Mount{
		{
			Type:     mount.TypeBind,
			Source:   filepath.Join(o.config.CertificatesPath, containerID),
			Target:   "/certs",
			ReadOnly: false,
		},
		{
			Type:   mount.TypeVolume,
			Source: fmt.Sprintf("client-container-%s-data", containerID),
			Target: "/data",
		},
	}

	return volumes
}

// ConfigureContainerNetwork generates network configuration
func (o *ContainerOrchestrator) ConfigureContainerNetwork() *network.NetworkingConfig {
	netConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			o.config.DockerNetwork: {
				NetworkID: o.config.DockerNetwork,
			},
		},
	}

	return netConfig
}

// GenerateContainerConfig generates complete Docker container configuration
func (o *ContainerOrchestrator) GenerateContainerConfig(orgID int, containerID string, containerData *models.ClientContainer, certSerial string) (container.Config, container.HostConfig, error) {
	_ = o.GenerateContainerName(containerID) // Container name is set in CreateContainer

	// Container configuration
	containerConfig := container.Config{
		Image: fmt.Sprintf("%s:%s", o.config.ClientContainerImage, o.config.ClientContainerImageTag),
		Env:   o.ConfigureContainerEnvironment(orgID, containerID, containerData, certSerial),
		Labels: map[string]string{
			"org.itaas.container-id":    containerID,
			"org.itaas.organization-id": fmt.Sprintf("%d", orgID),
			"org.itaas.service":         "client-container",
		},
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%s/tcp", o.config.ClientContainerPort)):     {},
			nat.Port(fmt.Sprintf("%s/tcp", o.config.ClientContainerGRPCPort)): {},
		},
	}

	// Host configuration
	hostConfig := container.HostConfig{
		Mounts:      o.ConfigureContainerVolumes(containerID),
		NetworkMode: container.NetworkMode(o.config.DockerNetwork),
		RestartPolicy: container.RestartPolicy{
			Name:              "unless-stopped",
			MaximumRetryCount: 0,
		},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%s/tcp", o.config.ClientContainerPort)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: o.config.ClientContainerPort,
				},
			},
			nat.Port(fmt.Sprintf("%s/tcp", o.config.ClientContainerGRPCPort)): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: o.config.ClientContainerGRPCPort,
				},
			},
		},
	}

	return containerConfig, hostConfig, nil
}

// CreateClientContainer creates and starts a client-container Docker container
func (o *ContainerOrchestrator) CreateClientContainer(ctx context.Context, orgID int, containerID string, containerData *models.ClientContainer, certSerial string) (string, error) {
	// Generate container configuration
	containerConfig, hostConfig, err := o.GenerateContainerConfig(orgID, containerID, containerData, certSerial)
	if err != nil {
		return "", fmt.Errorf("failed to generate container config: %v", err)
	}

	containerName := o.GenerateContainerName(containerID)

	// Create container
	dockerContainerID, err := o.dockerClient.CreateContainer(ctx, containerConfig, hostConfig, containerName)
	if err != nil {
		return "", fmt.Errorf("failed to create Docker container: %v", err)
	}

	// Start container
	if err := o.dockerClient.StartContainer(ctx, dockerContainerID); err != nil {
		// If start fails, try to remove the container
		o.dockerClient.RemoveContainer(ctx, dockerContainerID, true)
		return "", fmt.Errorf("failed to start Docker container: %v", err)
	}

	return dockerContainerID, nil
}
