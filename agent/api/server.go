package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/agent/provider"
	"github.com/scttfrdmn/snoozebot/agent/store"
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol"
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol/gen"
	"github.com/scttfrdmn/snoozebot/pkg/notification"
	
	"google.golang.org/grpc"
)

// Server handles the HTTP API for the agent
type Server struct {
	store                  store.Store
	pluginsDir             string
	configDir              string
	pluginManager          provider.PluginManager
	authenticatedManager   *provider.PluginManagerWithAuth
	logger                 hclog.Logger
	notificationManager    *notification.Manager
}

// NewServer creates a new API server
func NewServer(store store.Store, pluginsDir string, configDir string) *Server {
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "snoozebot-agent",
		Output: log.Writer(),
		Level:  hclog.Info,
	})

	// Create the base plugin manager
	baseManager := provider.NewPluginManager(pluginsDir, logger)
	
	// Create the authenticated plugin manager
	authenticatedManager, err := provider.NewPluginManagerWithAuth(baseManager, configDir, logger.Named("auth"))
	if err != nil {
		logger.Error("Failed to create authenticated plugin manager", "error", err)
		// Fall back to base manager if authentication fails
		return &Server{
			store:         store,
			pluginsDir:    pluginsDir,
			configDir:     configDir,
			pluginManager: baseManager,
			logger:        logger,
		}
	}

	// Initialize notification manager
	notificationConfigPath := filepath.Join(configDir, "notifications.yaml")
	notificationManager, err := notification.InitManagerFromConfig(notificationConfigPath, logger)
	if err != nil {
		logger.Error("Failed to initialize notification manager", "error", err)
		// Continue without notifications if it fails
		notificationManager = notification.NewManager(logger)
	}

	return &Server{
		store:                store,
		pluginsDir:           pluginsDir,
		configDir:            configDir,
		pluginManager:        baseManager,
		authenticatedManager: authenticatedManager,
		logger:               logger,
		notificationManager:  notificationManager,
	}
}

// StartGRPCServer starts the gRPC server for agent communication
func (s *Server) StartGRPCServer(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server
	grpcServer := grpc.NewServer()
	
	// Register the service
	agentServer := NewGRPCServer(s.store, s.pluginManager)
	gen.RegisterSnoozeAgentServer(grpcServer, agentServer)
	
	// Start the server in a goroutine
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()
	
	log.Printf("gRPC server started on %s", address)
	return nil
}

// Router returns the HTTP router for the API server
func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/instances/register", s.handleRegisterInstance)
	mux.HandleFunc("/api/instances/unregister", s.handleUnregisterInstance)
	mux.HandleFunc("/api/instances/idle", s.handleIdleNotification)
	mux.HandleFunc("/api/instances/heartbeat", s.handleHeartbeat)
	mux.HandleFunc("/api/instances/state", s.handleStateChange)
	mux.HandleFunc("/api/instances", s.handleListInstances)

	// Management routes (for admin UI)
	mux.HandleFunc("/api/admin/instances", s.handleAdminListInstances)
	mux.HandleFunc("/api/admin/instances/", s.handleAdminGetInstance)
	mux.HandleFunc("/api/admin/actions", s.handleAdminScheduleAction)
	
	// Plugin management routes
	mux.HandleFunc("/api/plugins", s.handleListPlugins)
	mux.HandleFunc("/api/plugins/discover", s.handleDiscoverPlugins)
	mux.HandleFunc("/api/plugins/load", s.handleLoadPlugin)
	mux.HandleFunc("/api/plugins/unload", s.handleUnloadPlugin)
	mux.HandleFunc("/api/plugins/", s.handleGetPluginInfo)
	
	// Authentication routes
	if s.authenticatedManager != nil {
		mux.HandleFunc("/api/auth/status", s.handleAuthStatus)
		mux.HandleFunc("/api/auth/enable", s.handleEnableAuth)
		mux.HandleFunc("/api/auth/disable", s.handleDisableAuth)
		mux.HandleFunc("/api/auth/apikey", s.handleGenerateAPIKey)
		mux.HandleFunc("/api/auth/apikey/revoke", s.handleRevokeAPIKey)
	}

	return mux
}

// handleRegisterInstance handles instance registration
func (s *Server) handleRegisterInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var registration protocol.InstanceRegistration
	if err := json.NewDecoder(r.Body).Decode(&registration); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Register the instance
	if err := s.store.RegisterInstance(registration); err != nil {
		http.Error(w, fmt.Sprintf("Failed to register instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := protocol.RegistrationResponse{
		Success:           true,
		AgentID:           "agent-1", // In a real implementation, this would be a unique ID
		HeartbeatInterval: 30 * time.Second,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUnregisterInstance handles instance unregistration
func (s *Server) handleUnregisterInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		InstanceID string `json:"instance_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Unregister the instance
	if err := s.store.UnregisterInstance(request.InstanceID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unregister instance: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// handleIdleNotification handles idle notifications from instances
func (s *Server) handleIdleNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var notification protocol.IdleNotification
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Get the instance
	instance, err := s.store.GetInstance(notification.InstanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Instance not found: %v", err), http.StatusNotFound)
		return
	}

	// Update idle state
	if err := s.store.UpdateIdleState(
		notification.InstanceID,
		true,
		notification.IdleSince,
		notification.IdleDuration,
	); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update idle state: %v", err), http.StatusInternalServerError)
		return
	}

	// Update resource usage
	if err := s.store.UpdateResourceUsage(notification.InstanceID, notification.ResourceUsage); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update resource usage: %v", err), http.StatusInternalServerError)
		return
	}

	// Send idle notification if we have a notification manager
	if s.notificationManager != nil {
		// Get instance name from metadata or use ID if not available
		instanceName := notification.InstanceID
		if name, ok := instance.Registration.Metadata["name"]; ok && name != "" {
			instanceName = name
		}

		// Send notification about idle instance
		go s.notificationManager.NotifyIdle(
			context.Background(),
			notification.InstanceID,
			instanceName,
			instance.Registration.Provider,
			instance.Registration.Region,
			notification.IdleDuration,
		)
	}

	// Determine action to take
	var response protocol.IdleNotificationResponse

	// In a real implementation, we would check policies and decide whether to stop the instance
	// For now, we'll use a simple rule: if idle for more than the naptime, stop the instance
	if notification.IdleDuration >= instance.Registration.NapTime {
		response = protocol.IdleNotificationResponse{
			Action: "stop",
			Reason: fmt.Sprintf("Instance has been idle for %s (threshold: %s)",
				notification.IdleDuration, instance.Registration.NapTime),
			ScheduledAction: &protocol.ScheduledAction{
				Action:        "stop",
				ScheduledTime: time.Now().Add(5 * time.Minute), // Schedule stop in 5 minutes
				Reason:        "Idle timeout",
			},
		}

		// Add scheduled action to the instance
		if err := s.store.AddScheduledAction(notification.InstanceID, *response.ScheduledAction); err != nil {
			http.Error(w, fmt.Sprintf("Failed to add scheduled action: %v", err), http.StatusInternalServerError)
			return
		}

		// Send scheduled action notification if we have a notification manager
		if s.notificationManager != nil {
			// Get instance name from metadata or use ID if not available
			instanceName := notification.InstanceID
			if name, ok := instance.Registration.Metadata["name"]; ok && name != "" {
				instanceName = name
			}

			// Send notification about scheduled action
			go s.notificationManager.NotifyScheduledAction(
				context.Background(),
				notification.InstanceID,
				instanceName,
				instance.Registration.Provider,
				instance.Registration.Region,
				response.ScheduledAction.Action,
				response.ScheduledAction.ScheduledTime,
				response.ScheduledAction.Reason,
			)
		}
	} else {
		response = protocol.IdleNotificationResponse{
			Action: "wait",
			Reason: fmt.Sprintf("Instance has been idle for %s, but threshold is %s",
				notification.IdleDuration, instance.Registration.NapTime),
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHeartbeat handles heartbeats from instances
func (s *Server) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var heartbeat protocol.Heartbeat
	if err := json.NewDecoder(r.Body).Decode(&heartbeat); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Update last heartbeat time
	if err := s.store.UpdateLastHeartbeat(heartbeat.InstanceID, heartbeat.Timestamp); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update heartbeat: %v", err), http.StatusInternalServerError)
		return
	}

	// Update resource usage if provided
	if heartbeat.ResourceUsage != nil {
		if err := s.store.UpdateResourceUsage(heartbeat.InstanceID, heartbeat.ResourceUsage); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update resource usage: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// Get any commands for the instance
	// In a real implementation, this would check for scheduled actions
	response := protocol.HeartbeatResponse{
		Acknowledged: true,
		Commands:     make([]protocol.InstanceCommand, 0),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleStateChange handles state changes from instances
func (s *Server) handleStateChange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stateChange protocol.InstanceStateChange
	if err := json.NewDecoder(r.Body).Decode(&stateChange); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Update instance state
	if err := s.store.UpdateInstanceState(stateChange.InstanceID, stateChange.CurrentState); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update instance state: %v", err), http.StatusInternalServerError)
		return
	}

	// Send state change notification if we have a notification manager
	if s.notificationManager != nil {
		// Get the instance to get additional information
		instance, err := s.store.GetInstance(stateChange.InstanceID)
		if err == nil { // Don't fail if we can't get the instance details
			// Get instance name from metadata or use ID if not available
			instanceName := stateChange.InstanceID
			if name, ok := instance.Registration.Metadata["name"]; ok && name != "" {
				instanceName = name
			}

			// Send notification about state change
			go s.notificationManager.NotifyStateChange(
				context.Background(),
				stateChange.InstanceID,
				instanceName,
				instance.Registration.Provider,
				instance.Registration.Region,
				stateChange.PreviousState,
				stateChange.CurrentState,
				stateChange.Reason,
			)
		}
	}

	// Return success response
	response := protocol.InstanceStateChangeResponse{
		Acknowledged: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleListInstances handles listing instances
func (s *Server) handleListInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get state filter from query params
	state := r.URL.Query().Get("state")

	var instances map[string]*store.InstanceState
	var err error

	if state != "" {
		instances, err = s.store.GetInstancesByState(state)
	} else {
		instances, err = s.store.GetAllInstances()
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get instances: %v", err), http.StatusInternalServerError)
		return
	}

	// Return instances
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

// handleAdminListInstances handles admin listing of instances
func (s *Server) handleAdminListInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real implementation, this would require authentication

	instances, err := s.store.GetAllInstances()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get instances: %v", err), http.StatusInternalServerError)
		return
	}

	// Return instances with additional details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instances)
}

// handleAdminGetInstance handles admin getting a specific instance
func (s *Server) handleAdminGetInstance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real implementation, this would require authentication

	// Extract instance ID from URL path
	path := r.URL.Path
	if len(path) <= len("/api/admin/instances/") {
		http.Error(w, "Invalid instance ID", http.StatusBadRequest)
		return
	}
	instanceID := path[len("/api/admin/instances/"):]

	instance, err := s.store.GetInstance(instanceID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Instance not found: %v", err), http.StatusNotFound)
		return
	}

	// Return instance
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(instance)
}

// handleAdminScheduleAction handles admin scheduling an action for an instance
func (s *Server) handleAdminScheduleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// In a real implementation, this would require authentication

	var request struct {
		InstanceID      string                     `json:"instance_id"`
		ScheduledAction protocol.ScheduledAction `json:"scheduled_action"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Add scheduled action
	if err := s.store.AddScheduledAction(request.InstanceID, request.ScheduledAction); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add scheduled action: %v", err), http.StatusInternalServerError)
		return
	}

	// Send scheduled action notification if we have a notification manager
	if s.notificationManager != nil {
		// Get the instance to get additional information
		instance, err := s.store.GetInstance(request.InstanceID)
		if err == nil { // Don't fail if we can't get the instance details
			// Get instance name from metadata or use ID if not available
			instanceName := request.InstanceID
			if name, ok := instance.Registration.Metadata["name"]; ok && name != "" {
				instanceName = name
			}

			// Send notification about scheduled action
			go s.notificationManager.NotifyScheduledAction(
				context.Background(),
				request.InstanceID,
				instanceName,
				instance.Registration.Provider,
				instance.Registration.Region,
				request.ScheduledAction.Action,
				request.ScheduledAction.ScheduledTime,
				request.ScheduledAction.Reason,
			)
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}