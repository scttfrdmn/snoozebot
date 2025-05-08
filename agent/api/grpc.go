package api

import (
	"context"
	"fmt"
	"time"

	"github.com/scttfrdmn/snoozebot/agent/provider"
	"github.com/scttfrdmn/snoozebot/agent/store"
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol"
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol/gen"
)

// GRPCServer implements the SnoozeAgent gRPC service
type GRPCServer struct {
	gen.UnimplementedSnoozeAgentServer
	instanceStore  store.Store
	pluginManager  provider.PluginManager
	agentID        string
}

// NewGRPCServer creates a new gRPC server
func NewGRPCServer(instanceStore store.Store, pluginManager provider.PluginManager) *GRPCServer {
	return &GRPCServer{
		instanceStore:  instanceStore,
		pluginManager:  pluginManager,
		agentID:        "agent-1", // In a real implementation, this would be a unique ID
	}
}

// RegisterInstance registers a new instance with the agent
func (s *GRPCServer) RegisterInstance(ctx context.Context, req *gen.InstanceRegistration) (*gen.RegistrationResponse, error) {
	// Convert thresholds to internal format
	thresholds := make(map[string]float64)
	for k, v := range req.Thresholds {
		thresholds[k] = v
	}

	// Create registration
	registration := protocol.InstanceRegistration{
		InstanceID:   req.InstanceId,
		InstanceType: req.InstanceType,
		Region:       req.Region,
		Zone:         req.Zone,
		Provider:     req.Provider,
		Metadata:     req.Metadata,
		NapTime:      time.Duration(req.NapTime) * time.Second,
	}

	// Register the instance
	err := s.instanceStore.RegisterInstance(registration)
	if err != nil {
		return &gen.RegistrationResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Return success response
	return &gen.RegistrationResponse{
		Success:           true,
		AgentId:           s.agentID,
		HeartbeatInterval: 30, // 30 seconds
	}, nil
}

// UnregisterInstance unregisters an instance from the agent
func (s *GRPCServer) UnregisterInstance(ctx context.Context, req *gen.UnregisterRequest) (*gen.UnregisterResponse, error) {
	// Unregister the instance
	err := s.instanceStore.UnregisterInstance(req.InstanceId)
	if err != nil {
		return &gen.UnregisterResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Return success response
	return &gen.UnregisterResponse{
		Success: true,
	}, nil
}

// SendIdleNotification handles idle notifications from instances
func (s *GRPCServer) SendIdleNotification(ctx context.Context, req *gen.IdleNotificationRequest) (*gen.IdleNotificationResponse, error) {
	// Get the instance
	instance, err := s.instanceStore.GetInstance(req.InstanceId)
	if err != nil {
		return &gen.IdleNotificationResponse{
			Action: "error",
			Reason: fmt.Sprintf("Failed to get instance: %v", err),
		}, nil
	}

	// Convert idle time
	idleSince := time.Unix(req.IdleSince, 0)
	idleDuration := time.Duration(req.IdleDuration) * time.Second

	// Update idle state
	err = s.instanceStore.UpdateIdleState(
		req.InstanceId,
		true,
		idleSince,
		idleDuration,
	)
	if err != nil {
		return &gen.IdleNotificationResponse{
			Action: "error",
			Reason: fmt.Sprintf("Failed to update idle state: %v", err),
		}, nil
	}

	// Update resource usage
	err = s.instanceStore.UpdateResourceUsage(req.InstanceId, req.ResourceUsage)
	if err != nil {
		return &gen.IdleNotificationResponse{
			Action: "error",
			Reason: fmt.Sprintf("Failed to update resource usage: %v", err),
		}, nil
	}

	// Determine action to take
	var response *gen.IdleNotificationResponse

	// In a real implementation, we would check policies and decide whether to stop the instance
	// For now, we'll use a simple rule: if idle for more than the naptime, stop the instance
	if idleDuration >= instance.Registration.NapTime {
		scheduledTime := time.Now().Add(5 * time.Minute) // Schedule stop in 5 minutes

		response = &gen.IdleNotificationResponse{
			Action: "stop",
			Reason: fmt.Sprintf("Instance has been idle for %s (threshold: %s)",
				idleDuration, instance.Registration.NapTime),
			ScheduledAction: &gen.ScheduledAction{
				Action:        "stop",
				ScheduledTime: scheduledTime.Unix(),
				Reason:        "Idle timeout",
			},
		}

		// Add scheduled action to the instance
		err = s.instanceStore.AddScheduledAction(req.InstanceId, protocol.ScheduledAction{
			Action:        "stop",
			ScheduledTime: scheduledTime,
			Reason:        "Idle timeout",
		})
		if err != nil {
			return &gen.IdleNotificationResponse{
				Action: "error",
				Reason: fmt.Sprintf("Failed to add scheduled action: %v", err),
			}, nil
		}
	} else {
		response = &gen.IdleNotificationResponse{
			Action: "wait",
			Reason: fmt.Sprintf("Instance has been idle for %s, but threshold is %s",
				idleDuration, instance.Registration.NapTime),
		}
	}

	return response, nil
}

// SendHeartbeat handles heartbeats from instances
func (s *GRPCServer) SendHeartbeat(ctx context.Context, req *gen.HeartbeatRequest) (*gen.HeartbeatResponse, error) {
	// Update last heartbeat time
	err := s.instanceStore.UpdateLastHeartbeat(req.InstanceId, time.Unix(req.Timestamp, 0))
	if err != nil {
		return &gen.HeartbeatResponse{
			Acknowledged: false,
		}, nil
	}

	// Update resource usage if provided
	if len(req.ResourceUsage) > 0 {
		err = s.instanceStore.UpdateResourceUsage(req.InstanceId, req.ResourceUsage)
		if err != nil {
			return &gen.HeartbeatResponse{
				Acknowledged: false,
			}, nil
		}
	}

	// Update instance state
	err = s.instanceStore.UpdateInstanceState(req.InstanceId, req.State)
	if err != nil {
		return &gen.HeartbeatResponse{
				Acknowledged: false,
			}, nil
	}

	// Get scheduled actions for the instance
	instance, err := s.instanceStore.GetInstance(req.InstanceId)
	if err != nil {
		return &gen.HeartbeatResponse{
			Acknowledged: false,
		}, nil
	}

	// Check if any scheduled actions are due
	commands := make([]*gen.Command, 0)
	now := time.Now()

	for _, action := range instance.ScheduledActions {
		if now.After(action.ScheduledTime) {
			// Action is due, add command
			commands = append(commands, &gen.Command{
				Command: action.Action,
				Parameters: map[string]string{
					"reason": action.Reason,
				},
			})

			// Remove the action (in a real implementation, we would mark it as executed)
			// This is a simplification
		}
	}

	return &gen.HeartbeatResponse{
		Acknowledged: true,
		Commands:     commands,
	}, nil
}

// ReportStateChange handles state change reports from instances
func (s *GRPCServer) ReportStateChange(ctx context.Context, req *gen.StateChangeRequest) (*gen.StateChangeResponse, error) {
	// Update instance state
	err := s.instanceStore.UpdateInstanceState(req.InstanceId, req.CurrentState)
	if err != nil {
		return &gen.StateChangeResponse{
			Acknowledged: false,
			Error:        err.Error(),
		}, nil
	}

	// In a real implementation, we would record the state change in a history log
	// For now, just return success

	return &gen.StateChangeResponse{
		Acknowledged: true,
	}, nil
}