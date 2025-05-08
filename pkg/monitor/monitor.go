package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol"
	"github.com/scttfrdmn/snoozebot/pkg/monitor/resources"
)

// monitor implements the Monitor interface
type monitor struct {
	config            Config
	customMonitors    map[string]ResourceMonitorFunc
	idleStateHandlers []IdleStateChangeHandler
	errorHandlers     []ErrorHandler
	
	currentState      MonitorState
	ctx               context.Context
	cancel            context.CancelFunc
	running           bool
	mutex             sync.RWMutex
	wg                sync.WaitGroup
}

// newMonitor creates a new monitor instance
func newMonitor(config Config) *monitor {
	return &monitor{
		config:            config,
		customMonitors:    make(map[string]ResourceMonitorFunc),
		idleStateHandlers: make([]IdleStateChangeHandler, 0),
		errorHandlers:     make([]ErrorHandler, 0),
		currentState: MonitorState{
			IsIdle:       false,
			IdleSince:    time.Time{},
			IdleDuration: 0,
			CurrentUsage: make(map[ResourceType]*ResourceUsage),
			Connected:    false,
		},
	}
}

// Configuration methods

func (m *monitor) WithThreshold(resourceType ResourceType, threshold float64) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.config.Thresholds[resourceType] = threshold
	return m
}

func (m *monitor) WithNapTime(duration time.Duration) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.config.NapTime = duration
	return m
}

func (m *monitor) WithCheckInterval(duration time.Duration) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.config.CheckInterval = duration
	return m
}

func (m *monitor) WithAgentURL(url string) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.config.AgentURL = url
	return m
}

// Custom monitoring

func (m *monitor) AddResourceMonitor(name string, fn ResourceMonitorFunc) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.customMonitors[name] = fn
	return m
}

// Event handlers

func (m *monitor) OnIdleStateChange(fn IdleStateChangeHandler) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.idleStateHandlers = append(m.idleStateHandlers, fn)
	return m
}

func (m *monitor) OnError(fn ErrorHandler) Monitor {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.errorHandlers = append(m.errorHandlers, fn)
	return m
}

// Lifecycle

func (m *monitor) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if m.running {
		return fmt.Errorf("monitor already running")
	}
	
	m.ctx, m.cancel = context.WithCancel(ctx)
	m.running = true
	
	// Start the monitoring goroutine
	m.wg.Add(1)
	go m.monitorResources()
	
	// Start the agent connection goroutine
	m.wg.Add(1)
	go m.connectToAgent()
	
	return nil
}

func (m *monitor) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if !m.running {
		return fmt.Errorf("monitor not running")
	}
	
	m.cancel()
	m.wg.Wait()
	m.running = false
	
	return nil
}

// State

func (m *monitor) GetCurrentState() MonitorState {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Create a deep copy of the current state
	stateCopy := MonitorState{
		IsIdle:       m.currentState.IsIdle,
		IdleSince:    m.currentState.IdleSince,
		IdleDuration: m.currentState.IdleDuration,
		CurrentUsage: make(map[ResourceType]*ResourceUsage),
		Connected:    m.currentState.Connected,
	}
	
	// Copy the usage map
	for k, v := range m.currentState.CurrentUsage {
		usageCopy := *v
		stateCopy.CurrentUsage[k] = &usageCopy
	}
	
	return stateCopy
}

func (m *monitor) IsIdle() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	return m.currentState.IsIdle
}

func (m *monitor) IdleDuration() time.Duration {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	if !m.currentState.IsIdle {
		return 0
	}
	
	return time.Since(m.currentState.IdleSince)
}

// Internal methods

func (m *monitor) monitorResources() {
	defer m.wg.Done()
	
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateResourceUsage()
			m.checkIdleState()
		}
	}
}

func (m *monitor) updateResourceUsage() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Get all resource usage from the resource manager
	resourceManager, err := resources.NewMonitorManager()
	if err != nil {
		m.handleError(fmt.Errorf("failed to create resource manager: %w", err))
		return
	}
	
	// Add custom monitors to the resource manager
	for name, monitorFn := range m.customMonitors {
		resourceManager.AddCustomMonitor(name, monitorFn)
	}
	
	// Get all usage
	allUsage, err := resourceManager.GetAllUsage()
	if err != nil {
		m.handleError(fmt.Errorf("failed to get resource usage: %w", err))
		return
	}
	
	// Convert to our internal format
	for resourceType, usage := range allUsage {
		m.currentState.CurrentUsage[resourceType] = &ResourceUsage{
			Type:      resourceType,
			Value:     usage.Value,
			Timestamp: usage.Timestamp,
		}
	}
}

func (m *monitor) checkIdleState() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// Check if any resource is above its threshold
	isIdle := true
	for resourceType, usage := range m.currentState.CurrentUsage {
		threshold, ok := m.config.Thresholds[resourceType]
		if !ok {
			continue
		}
		
		if usage.Value > threshold {
			isIdle = false
			break
		}
	}
	
	// Update idle state
	now := time.Now()
	wasIdle := m.currentState.IsIdle
	
	if isIdle {
		if !wasIdle {
			// System just became idle
			m.currentState.IsIdle = true
			m.currentState.IdleSince = now
			m.currentState.IdleDuration = 0
			
			// Notify handlers
			m.notifyIdleStateChange(true, 0)
		} else {
			// System was already idle, update duration
			m.currentState.IdleDuration = now.Sub(m.currentState.IdleSince)
			
			// If we've reached the naptime threshold, notify the agent
			if m.currentState.IdleDuration >= m.config.NapTime {
				// Only notify if we're connected to an agent
				if m.currentState.Connected {
					fmt.Printf("System has been idle for %s, notifying agent\n", m.currentState.IdleDuration)
					
					// This would be better done in a separate goroutine 
					// and with a reference to the agent client
					go func() {
						// Create a new agent client just for this notification
						client := protocol.NewAgentClient(m.config.AgentURL, getInstanceID())
						err := client.Connect(context.Background())
						if err != nil {
							m.handleError(fmt.Errorf("failed to connect to agent for idle notification: %w", err))
							return
						}
						defer client.Disconnect()
						
						// Get current resource usage
						resourceUsage := make(map[string]float64)
						for k, v := range m.currentState.CurrentUsage {
							resourceUsage[string(k)] = v.Value
						}
						
						// Send idle notification
						action, err := client.SendIdleNotification(
							context.Background(),
							m.currentState.IdleSince,
							m.currentState.IdleDuration,
							resourceUsage,
						)
						
						if err != nil {
							m.handleError(fmt.Errorf("failed to send idle notification: %w", err))
							return
						}
						
						fmt.Printf("Agent action response: %s\n", action)
					}()
				} else {
					fmt.Printf("System has been idle for %s (not connected to agent)\n", m.currentState.IdleDuration)
				}
			}
		}
	} else {
		if wasIdle {
			// System was idle but is now active
			m.currentState.IsIdle = false
			m.currentState.IdleSince = time.Time{}
			m.currentState.IdleDuration = 0
			
			// Notify handlers
			m.notifyIdleStateChange(false, 0)
		}
		// else: System was already active, nothing to do
	}
}

func (m *monitor) connectToAgent() {
	defer m.wg.Done()
	
	if m.config.AgentURL == "" {
		m.mutex.Lock()
		m.currentState.Connected = false
		m.mutex.Unlock()
		
		// No agent URL configured, so we'll just run in standalone mode
		fmt.Println("Running in standalone mode (no agent URL configured)")
		
		// Wait for shutdown signal
		<-m.ctx.Done()
		return
	}
	
	// Create a new agent client
	client := protocol.NewAgentClient(m.config.AgentURL, getInstanceID())
	
	// Try to connect to the agent
	err := client.Connect(m.ctx)
	if err != nil {
		m.handleError(fmt.Errorf("failed to connect to agent: %w", err))
		m.mutex.Lock()
		m.currentState.Connected = false
		m.mutex.Unlock()
		
		// Wait for shutdown signal
		<-m.ctx.Done()
		return
	}
	
	m.mutex.Lock()
	m.currentState.Connected = true
	m.mutex.Unlock()
	
	// Register the instance with the agent
	err = m.registerWithAgent(client)
	if err != nil {
		m.handleError(fmt.Errorf("failed to register with agent: %w", err))
		client.Disconnect()
		
		m.mutex.Lock()
		m.currentState.Connected = false
		m.mutex.Unlock()
		
		// Wait for shutdown signal
		<-m.ctx.Done()
		return
	}
	
	// Set up heartbeat ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	// Set up agent communication loop
	for {
		select {
		case <-m.ctx.Done():
			// Unregister and disconnect when shutting down
			m.unregisterFromAgent(client)
			client.Disconnect()
			return
			
		case <-ticker.C:
			// Send heartbeat to agent
			err := m.sendHeartbeat(client)
			if err != nil {
				m.handleError(fmt.Errorf("heartbeat failed: %w", err))
				
				// Try to reconnect
				if !client.IsConnected() {
					err = client.Connect(m.ctx)
					if err != nil {
						m.handleError(fmt.Errorf("reconnect failed: %w", err))
						m.mutex.Lock()
						m.currentState.Connected = false
						m.mutex.Unlock()
						continue
					}
					
					// Re-register after reconnect
					err = m.registerWithAgent(client)
					if err != nil {
						m.handleError(fmt.Errorf("re-registration failed: %w", err))
						m.mutex.Lock()
						m.currentState.Connected = false
						m.mutex.Unlock()
						continue
					}
					
					m.mutex.Lock()
					m.currentState.Connected = true
					m.mutex.Unlock()
				}
			}
		}
	}
}

// registerWithAgent registers the monitor with the agent
func (m *monitor) registerWithAgent(client *protocol.AgentClient) error {
	// Get thresholds from configuration
	thresholds := make(map[string]float64)
	for k, v := range m.config.Thresholds {
		thresholds[string(k)] = v
	}
	
	// Get instance metadata
	instanceType, region, zone, provider := getInstanceMetadata()
	
	// Register with the agent
	return client.RegisterInstance(m.ctx, instanceType, region, zone, provider, thresholds, m.config.NapTime)
}

// unregisterFromAgent unregisters the monitor from the agent
func (m *monitor) unregisterFromAgent(client *protocol.AgentClient) {
	err := client.UnregisterInstance(m.ctx)
	if err != nil {
		m.handleError(fmt.Errorf("failed to unregister from agent: %w", err))
	}
}

// sendHeartbeat sends a heartbeat to the agent
func (m *monitor) sendHeartbeat(client *protocol.AgentClient) error {
	// Get current resource usage
	resourceUsage := make(map[string]float64)
	for k, v := range m.currentState.CurrentUsage {
		resourceUsage[string(k)] = v.Value
	}
	
	// Determine current state
	state := "active"
	if m.currentState.IsIdle {
		state = "idle"
	}
	
	// Send heartbeat
	commands, err := client.SendHeartbeat(m.ctx, state, resourceUsage)
	if err != nil {
		return err
	}
	
	// Process commands
	for _, command := range commands {
		m.processAgentCommand(command)
	}
	
	return nil
}

// processAgentCommand processes a command from the agent
func (m *monitor) processAgentCommand(command string) {
	switch command {
	case "ping":
		// Simple ping command - nothing to do
		fmt.Println("Received ping command from agent")
		
	case "stop":
		// Stop command - trigger stop immediately
		fmt.Println("Received stop command from agent")
		
	case "refresh":
		// Refresh command - trigger immediate resource check
		fmt.Println("Received refresh command from agent")
		m.updateResourceUsage()
		m.checkIdleState()
		
	default:
		m.handleError(fmt.Errorf("unknown command from agent: %s", command))
	}
}

// getInstanceID gets the ID of the current instance
func getInstanceID() string {
	// In a real implementation, this would get the instance ID from the cloud metadata service
	// For now, return a placeholder value
	return "instance-1234"
}

// getInstanceMetadata gets metadata about the current instance
func getInstanceMetadata() (string, string, string, string) {
	// In a real implementation, this would get instance metadata from the cloud metadata service
	// For now, return placeholder values
	instanceType := "unknown"
	region := "unknown"
	zone := "unknown"
	provider := "unknown"
	
	return instanceType, region, zone, provider
}

func (m *monitor) notifyIdleStateChange(isIdle bool, duration time.Duration) {
	// Create a copy of the handlers to avoid holding the lock during notification
	m.mutex.RLock()
	handlers := make([]IdleStateChangeHandler, len(m.idleStateHandlers))
	copy(handlers, m.idleStateHandlers)
	m.mutex.RUnlock()
	
	// Notify all handlers
	for _, handler := range handlers {
		go handler(isIdle, duration)
	}
}

func (m *monitor) handleError(err error) {
	// Create a copy of the handlers to avoid holding the lock during notification
	m.mutex.RLock()
	handlers := make([]ErrorHandler, len(m.errorHandlers))
	copy(handlers, m.errorHandlers)
	m.mutex.RUnlock()
	
	// Notify all handlers
	for _, handler := range handlers {
		go handler(err)
	}
}