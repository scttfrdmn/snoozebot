package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/scttfrdmn/snoozebot/pkg/common/protocol"
)

// InstanceState contains the current state of an instance
type InstanceState struct {
	// InstanceID is the unique identifier for the instance
	InstanceID string
	
	// Registration is the registration information for the instance
	Registration protocol.InstanceRegistration
	
	// State is the current state of the instance (running, idle, stopping, stopped)
	State string
	
	// LastHeartbeat is the time of the last heartbeat from the instance
	LastHeartbeat time.Time
	
	// IdleSince is the time when the instance became idle
	IdleSince time.Time
	
	// IdleDuration is how long the instance has been idle
	IdleDuration time.Duration
	
	// ResourceUsage is the most recent resource usage report
	ResourceUsage map[string]float64
	
	// ScheduledActions is a list of actions scheduled for the instance
	ScheduledActions []protocol.ScheduledAction
}

// Store defines the interface for storing and retrieving instance state
type Store interface {
	// RegisterInstance registers a new instance
	RegisterInstance(registration protocol.InstanceRegistration) error
	
	// UnregisterInstance unregisters an instance
	UnregisterInstance(instanceID string) error
	
	// GetInstance gets the state of an instance
	GetInstance(instanceID string) (*InstanceState, error)
	
	// UpdateInstanceState updates the state of an instance
	UpdateInstanceState(instanceID string, state string) error
	
	// UpdateResourceUsage updates the resource usage for an instance
	UpdateResourceUsage(instanceID string, usage map[string]float64) error
	
	// UpdateIdleState updates the idle state of an instance
	UpdateIdleState(instanceID string, isIdle bool, since time.Time, duration time.Duration) error
	
	// UpdateLastHeartbeat updates the time of the last heartbeat from an instance
	UpdateLastHeartbeat(instanceID string, time time.Time) error
	
	// AddScheduledAction adds a scheduled action for an instance
	AddScheduledAction(instanceID string, action protocol.ScheduledAction) error
	
	// RemoveScheduledAction removes a scheduled action for an instance
	RemoveScheduledAction(instanceID string, actionIndex int) error
	
	// GetAllInstances gets all registered instances
	GetAllInstances() (map[string]*InstanceState, error)
	
	// GetInstancesByState gets all instances in a specific state
	GetInstancesByState(state string) (map[string]*InstanceState, error)
}

// MemoryStore is an in-memory implementation of the Store interface
type MemoryStore struct {
	instances map[string]*InstanceState
	mutex     sync.RWMutex
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		instances: make(map[string]*InstanceState),
	}
}

// RegisterInstance registers a new instance
func (s *MemoryStore) RegisterInstance(registration protocol.InstanceRegistration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	s.instances[registration.InstanceID] = &InstanceState{
		InstanceID:       registration.InstanceID,
		Registration:     registration,
		State:            "running",
		LastHeartbeat:    time.Now(),
		ResourceUsage:    make(map[string]float64),
		ScheduledActions: make([]protocol.ScheduledAction, 0),
	}
	
	return nil
}

// UnregisterInstance unregisters an instance
func (s *MemoryStore) UnregisterInstance(instanceID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	delete(s.instances, instanceID)
	return nil
}

// GetInstance gets the state of an instance
func (s *MemoryStore) GetInstance(instanceID string) (*InstanceState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return nil, fmt.Errorf("instance not found: %s", instanceID)
	}
	
	return instance, nil
}

// UpdateInstanceState updates the state of an instance
func (s *MemoryStore) UpdateInstanceState(instanceID string, state string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	instance.State = state
	return nil
}

// UpdateResourceUsage updates the resource usage for an instance
func (s *MemoryStore) UpdateResourceUsage(instanceID string, usage map[string]float64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	instance.ResourceUsage = usage
	return nil
}

// UpdateIdleState updates the idle state of an instance
func (s *MemoryStore) UpdateIdleState(instanceID string, isIdle bool, since time.Time, duration time.Duration) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	if isIdle {
		instance.IdleSince = since
		instance.IdleDuration = duration
	} else {
		instance.IdleSince = time.Time{}
		instance.IdleDuration = 0
	}
	
	return nil
}

// UpdateLastHeartbeat updates the time of the last heartbeat from an instance
func (s *MemoryStore) UpdateLastHeartbeat(instanceID string, t time.Time) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	instance.LastHeartbeat = t
	return nil
}

// AddScheduledAction adds a scheduled action for an instance
func (s *MemoryStore) AddScheduledAction(instanceID string, action protocol.ScheduledAction) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	instance.ScheduledActions = append(instance.ScheduledActions, action)
	return nil
}

// RemoveScheduledAction removes a scheduled action for an instance
func (s *MemoryStore) RemoveScheduledAction(instanceID string, actionIndex int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	instance, ok := s.instances[instanceID]
	if !ok {
		return fmt.Errorf("instance not found: %s", instanceID)
	}
	
	if actionIndex < 0 || actionIndex >= len(instance.ScheduledActions) {
		return fmt.Errorf("invalid action index: %d", actionIndex)
	}
	
	// Remove the action at the given index
	instance.ScheduledActions = append(
		instance.ScheduledActions[:actionIndex],
		instance.ScheduledActions[actionIndex+1:]...,
	)
	
	return nil
}

// GetAllInstances gets all registered instances
func (s *MemoryStore) GetAllInstances() (map[string]*InstanceState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Create a copy of the instances map
	instances := make(map[string]*InstanceState)
	for id, instance := range s.instances {
		instances[id] = instance
	}
	
	return instances, nil
}

// GetInstancesByState gets all instances in a specific state
func (s *MemoryStore) GetInstancesByState(state string) (map[string]*InstanceState, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	// Filter instances by state
	instances := make(map[string]*InstanceState)
	for id, instance := range s.instances {
		if instance.State == state {
			instances[id] = instance
		}
	}
	
	return instances, nil
}