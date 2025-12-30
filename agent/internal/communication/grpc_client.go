package communication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/agent/config"
	"backend/agent/internal/utils"
	agentpb "backend/agent/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// Client is a gRPC client for agent communication
type Client struct {
	config           *config.Config
	logger           *utils.Logger
	conn             *grpc.ClientConn
	client           agentpb.AgentServiceClient
	reconnectManager *ReconnectManager
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewClient creates a new gRPC client
func NewClient(cfg *config.Config, logger *utils.Logger) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := &Client{
		config:           cfg,
		logger:           logger,
		reconnectManager: NewReconnectManager(logger),
		ctx:              ctx,
		cancel:           cancel,
	}

	// Connect with retry
	if err := client.connect(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to connect: %v", err)
	}

	// Start connection monitoring
	go client.monitorConnection()

	return client, nil
}

// connect establishes a gRPC connection
func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing connection if any
	if c.conn != nil {
		c.conn.Close()
	}

	var opts []grpc.DialOption

	// Configure mTLS if certificates are available
	if c.config.ClientCertPath != "" && c.config.ClientKeyPath != "" {
		creds, err := LoadMTLSCredentials(c.config)
		if err != nil {
			c.logger.Warn("Failed to load mTLS credentials, using insecure", map[string]interface{}{
				"error": err.Error(),
			})
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(creds))
			c.logger.Info("gRPC client configured with mTLS")
		}
	} else {
		c.logger.Warn("mTLS not configured, using insecure connection")
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add keepalive settings
	opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}))

	// Connect
	conn, err := grpc.NewClient(c.config.BackendGRPCURL, opts...)
	if err != nil {
		return fmt.Errorf("failed to create gRPC client: %v", err)
	}

	c.conn = conn
	c.client = agentpb.NewAgentServiceClient(conn)

	c.logger.Info("gRPC client connected", map[string]interface{}{
		"url": c.config.BackendGRPCURL,
	})

	return nil
}

// monitorConnection monitors the connection and reconnects if needed
func (c *Client) monitorConnection() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				c.logger.Warn("Connection lost, attempting reconnect")
				c.reconnect()
				continue
			}

			// Check connection state
			state := conn.GetState()
			if state != connectivity.Ready {
				c.logger.Warn("Connection not ready, attempting reconnect", map[string]interface{}{
					"state": state.String(),
				})
				c.reconnect()
			}
		}
	}
}

// reconnect attempts to reconnect with retry
func (c *Client) reconnect() {
	_, err := c.reconnectManager.ConnectWithRetry(c.ctx, func() (*grpc.ClientConn, error) {
		if err := c.connect(); err != nil {
			return nil, err
		}
		return c.conn, nil
	})

	if err != nil {
		c.logger.Error("Failed to reconnect", err)
	}
}

// GetClient returns the gRPC service client
func (c *Client) GetClient() agentpb.AgentServiceClient {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.client
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	c.cancel()
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// IsConnected checks if the client is connected
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.conn == nil {
		return false
	}

	state := c.conn.GetState()
	return state == connectivity.Ready
}
