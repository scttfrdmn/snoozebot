package protocol

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/scttfrdmn/snoozebot/pkg/common/protocol/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentClient is a client for communicating with the remote agent
type AgentClient struct {
	conn         *grpc.ClientConn
	client       gen.SnoozeAgentClient
	instanceID   string
	agentID      string
	agentURL     string
	connected    bool
	reconnecting bool
	mutex        sync.RWMutex
}

// NewAgentClient creates a new client for communicating with the remote agent
func NewAgentClient(agentURL, instanceID string) *AgentClient {
	return &AgentClient{
		agentURL:   agentURL,
		instanceID: instanceID,
	}
}

// Connect connects to the remote agent
func (c *AgentClient) Connect(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.connected {
		return nil
	}

	// Connect to the gRPC server
	conn, err := grpc.Dial(c.agentURL, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to agent: %w", err)
	}

	c.conn = conn
	c.client = gen.NewSnoozeAgentClient(conn)
	c.connected = true

	return nil
}

// Disconnect disconnects from the remote agent
func (c *AgentClient) Disconnect() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil
	}

	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
	}

	c.connected = false
	c.conn = nil
	c.client = nil

	return nil
}

// IsConnected returns true if the client is connected to the agent
func (c *AgentClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.connected
}

// RegisterInstance registers an instance with the agent
func (c *AgentClient) RegisterInstance(ctx context.Context, instanceType, region, zone, provider string, 
	thresholds map[string]float64, napTime time.Duration) error {
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to agent")
	}

	// Convert thresholds to protocol format
	protoThresholds := make(map[string]float64)
	for k, v := range thresholds {
		protoThresholds[string(k)] = v
	}

	// Prepare the request
	req := &gen.InstanceRegistration{
		InstanceId:   c.instanceID,
		InstanceType: instanceType,
		Region:       region,
		Zone:         zone,
		Provider:     provider,
		Thresholds:   protoThresholds,
		NapTime:      int64(napTime.Seconds()),
	}

	// Send the request
	resp, err := c.client.RegisterInstance(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register instance: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Error)
	}

	c.agentID = resp.AgentId

	log.Printf("Instance registered with agent %s", c.agentID)
	return nil
}

// UnregisterInstance unregisters an instance from the agent
func (c *AgentClient) UnregisterInstance(ctx context.Context) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to agent")
	}

	// Prepare the request
	req := &gen.UnregisterRequest{
		InstanceId: c.instanceID,
	}

	// Send the request
	resp, err := c.client.UnregisterInstance(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unregister instance: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("unregistration failed: %s", resp.Error)
	}

	log.Printf("Instance unregistered from agent %s", c.agentID)
	return nil
}

// SendIdleNotification sends an idle notification to the agent
func (c *AgentClient) SendIdleNotification(ctx context.Context, idleSince time.Time, 
	idleDuration time.Duration, resourceUsage map[string]float64) (string, error) {
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return "", fmt.Errorf("not connected to agent")
	}

	// Prepare the request
	req := &gen.IdleNotificationRequest{
		InstanceId:    c.instanceID,
		IdleSince:     idleSince.Unix(),
		IdleDuration:  int64(idleDuration.Seconds()),
		ResourceUsage: resourceUsage,
	}

	// Send the request
	resp, err := c.client.SendIdleNotification(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to send idle notification: %w", err)
	}

	// Return the action to take
	return resp.Action, nil
}

// SendHeartbeat sends a heartbeat to the agent
func (c *AgentClient) SendHeartbeat(ctx context.Context, state string, 
	resourceUsage map[string]float64) ([]string, error) {
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected to agent")
	}

	// Prepare the request
	req := &gen.HeartbeatRequest{
		InstanceId:    c.instanceID,
		Timestamp:     time.Now().Unix(),
		State:         state,
		ResourceUsage: resourceUsage,
	}

	// Send the request
	resp, err := c.client.SendHeartbeat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to send heartbeat: %w", err)
	}

	// Extract commands
	commands := make([]string, len(resp.Commands))
	for i, cmd := range resp.Commands {
		commands[i] = cmd.Command
	}

	return commands, nil
}

// ReportStateChange reports a state change to the agent
func (c *AgentClient) ReportStateChange(ctx context.Context, previousState, 
	currentState, reason string) error {
	
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to agent")
	}

	// Prepare the request
	req := &gen.StateChangeRequest{
		InstanceId:    c.instanceID,
		PreviousState: previousState,
		CurrentState:  currentState,
		Timestamp:     time.Now().Unix(),
		Reason:        reason,
	}

	// Send the request
	resp, err := c.client.ReportStateChange(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to report state change: %w", err)
	}

	if !resp.Acknowledged {
		return fmt.Errorf("state change not acknowledged: %s", resp.Error)
	}

	return nil
}