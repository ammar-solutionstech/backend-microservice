package services

import (
	"context"
	"fmt"
	"time"

	cerrdefs "github.com/containerd/containerd/errdefs"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

// DockerClient wraps Docker API client for container operations
type DockerClient struct {
	client *client.Client
}

// NewDockerClient creates a new Docker client
func NewDockerClient(dockerHost string) (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %v", err)
	}

	return &DockerClient{
		client: cli,
	}, nil
}

// CreateContainer creates a Docker container with the given configuration
func (d *DockerClient) CreateContainer(ctx context.Context, config container.Config, hostConfig container.HostConfig, netConfig *network.NetworkingConfig, name string) (string, error) {
	containerBody, err := d.client.ContainerCreate(ctx, &config, &hostConfig, netConfig, nil, name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %v", err)
	}

	return containerBody.ID, nil
}

// StartContainer starts a Docker container
func (d *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	if err := d.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %v", err)
	}
	return nil
}

// StopContainer stops a Docker container
func (d *DockerClient) StopContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	if timeout == nil {
		defaultTimeout := 10 * time.Second
		timeout = &defaultTimeout
	}

	timeoutSeconds := int(timeout.Seconds())
	opts := container.StopOptions{
		Timeout: &timeoutSeconds,
	}
	if err := d.client.ContainerStop(ctx, containerID, opts); err != nil {
		return fmt.Errorf("failed to stop container: %v", err)
	}
	return nil
}

// RestartContainer restarts a Docker container
func (d *DockerClient) RestartContainer(ctx context.Context, containerID string, timeout *time.Duration) error {
	if timeout == nil {
		defaultTimeout := 10 * time.Second
		timeout = &defaultTimeout
	}

	timeoutSeconds := int(timeout.Seconds())
	opts := container.StopOptions{
		Timeout: &timeoutSeconds,
	}
	if err := d.client.ContainerRestart(ctx, containerID, opts); err != nil {
		return fmt.Errorf("failed to restart container: %v", err)
	}
	return nil
}

// RemoveContainer removes a Docker container
func (d *DockerClient) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	opts := container.RemoveOptions{
		Force: force,
	}

	if err := d.client.ContainerRemove(ctx, containerID, opts); err != nil {
		return fmt.Errorf("failed to remove container: %v", err)
	}
	return nil
}

// GetContainerStatus gets the status of a Docker container
func (d *DockerClient) GetContainerStatus(ctx context.Context, containerID string) (string, error) {
	containerJSON, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("failed to inspect container: %v", err)
	}

	// Return container state status
	return containerJSON.State.Status, nil
}

// GetContainerInfo gets full container information
func (d *DockerClient) GetContainerInfo(ctx context.Context, containerID string) (container.InspectResponse, error) {
	containerJSON, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return container.InspectResponse{}, fmt.Errorf("failed to inspect container: %v", err)
	}

	return containerJSON, nil
}

// UpdateContainer updates container configuration (resources, restart policy, etc.)
func (d *DockerClient) UpdateContainer(ctx context.Context, containerID string, updateConfig container.UpdateConfig) error {
	_, err := d.client.ContainerUpdate(ctx, containerID, updateConfig)
	if err != nil {
		return fmt.Errorf("failed to update container: %v", err)
	}
	return nil
}

// ListContainers lists containers with optional filters
func (d *DockerClient) ListContainers(ctx context.Context, filterArgs filters.Args) ([]types.Container, error) {
	opts := container.ListOptions{
		All:     true,
		Filters: filterArgs,
	}

	containers, err := d.client.ContainerList(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %v", err)
	}

	return containers, nil
}

// ContainerExists checks if a container exists
func (d *DockerClient) ContainerExists(ctx context.Context, containerID string) (bool, error) {
	_, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check container existence: %v", err)
	}
	return true, nil
}

// Close closes the Docker client connection
func (d *DockerClient) Close() error {
	return d.client.Close()
}

// GenerateNetworkName generates a unique network name for a container
func (d *DockerClient) GenerateNetworkName(containerID string) string {
	return fmt.Sprintf("client-container-%s-network", containerID)
}

// CreateNetwork creates a new Docker network
func (d *DockerClient) CreateNetwork(ctx context.Context, networkName string, labels map[string]string) (string, error) {
	networkCreate := network.CreateOptions{
		//CheckDuplicate: true,
		Driver: "bridge",
		//EnableIPv6: false,
		Labels:     labels,
		Internal:   false,
		Attachable: true,
	}

	networkResp, err := d.client.NetworkCreate(ctx, networkName, networkCreate)
	if err != nil {
		return "", fmt.Errorf("failed to create network: %v", err)
	}

	return networkResp.ID, nil
}

// NetworkExists checks if a network exists
func (d *DockerClient) NetworkExists(ctx context.Context, networkName string) (bool, error) {
	_, err := d.client.NetworkInspect(ctx, networkName, network.InspectOptions{})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check network existence: %v", err)
	}
	return true, nil
}

// RemoveNetwork removes a Docker network
func (d *DockerClient) RemoveNetwork(ctx context.Context, networkName string) error {
	if err := d.client.NetworkRemove(ctx, networkName); err != nil {
		return fmt.Errorf("failed to remove network: %v", err)
	}
	return nil
}

// ConnectContainerToNetwork connects a container to a network
func (d *DockerClient) ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error {
	networkConfig := network.EndpointSettings{}
	if err := d.client.NetworkConnect(ctx, networkName, containerID, &networkConfig); err != nil {
		return fmt.Errorf("failed to connect container to network: %v", err)
	}
	return nil
}

// DisconnectContainerFromNetwork disconnects a container from a network
func (d *DockerClient) DisconnectContainerFromNetwork(ctx context.Context, containerID, networkName string, force bool) error {
	if err := d.client.NetworkDisconnect(ctx, networkName, containerID, force); err != nil {
		return fmt.Errorf("failed to disconnect container from network: %v", err)
	}
	return nil
}
