package api

import (
	"context"
	"testing"
	"time"

	"github.com/scttfrdmn/snoozebot/agent/provider"
	"github.com/scttfrdmn/snoozebot/pkg/common/protocol/gen"
)

// mockInstanceStore is a mock implementation of instance store for testing
type mockInstanceStore struct {
	instances map[string]*instance
}

type instance struct {
	Registration instanceRegistration
	State        string
}

type instanceRegistration struct {
	InstanceID string
	Provider   string
}

func newMockInstanceStore() *mockInstanceStore {
	return &mockInstanceStore{
		instances: make(map[string]*instance),
	}
}

func (m *mockInstanceStore) GetInstance(instanceID string) (*instance, error) {
	instance, ok := m.instances[instanceID]
	if !ok {
		return nil, nil
	}
	return instance, nil
}

func (m *mockInstanceStore) UpdateInstanceState(instanceID, state string) error {
	instance, ok := m.instances[instanceID]
	if !ok {
		return nil
	}
	instance.State = state
	return nil
}

func (m *mockInstanceStore) AddInstance(instanceID, provider string) {
	m.instances[instanceID] = &instance{
		Registration: instanceRegistration{
			InstanceID: instanceID,
			Provider:   provider,
		},
		State: "running",
	}
}

// TestGetInstanceInfo tests the GetInstanceInfo gRPC handler
func TestGetInstanceInfo(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{}

	// Create a mock instance store
	mockIS := newMockInstanceStore()
	mockIS.AddInstance("test-instance", "mock")

	// Create a gRPC server with mocks
	s := &GRPCServer{
		pluginManager: mockPM,
		instanceStore: mockIS,
	}

	// Create a request
	req := &gen.GetInstanceInfoRequest{
		InstanceId: "test-instance",
	}

	// Call the handler
	resp, err := s.GetInstanceInfo(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to get instance info: %v", err)
	}

	// Check that the plugin was loaded
	if len(mockPM.loadedPlugins) != 1 || mockPM.loadedPlugins[0] != "mock" {
		t.Errorf("Expected 'mock' plugin to be loaded, got %v", mockPM.loadedPlugins)
	}

	// Check the response
	if resp.Id != "test-instance" {
		t.Errorf("Expected instance ID 'test-instance', got '%s'", resp.Id)
	}
	if resp.Provider != "mock" {
		t.Errorf("Expected provider 'mock', got '%s'", resp.Provider)
	}
}

// TestStopInstance tests the StopInstance gRPC handler
func TestStopInstance(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{}

	// Create a mock instance store
	mockIS := newMockInstanceStore()
	mockIS.AddInstance("test-instance", "mock")

	// Create a gRPC server with mocks
	s := &GRPCServer{
		pluginManager: mockPM,
		instanceStore: mockIS,
	}

	// Create a request
	req := &gen.StopInstanceRequest{
		InstanceId: "test-instance",
	}

	// Call the handler
	resp, err := s.StopInstance(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to stop instance: %v", err)
	}

	// Check that the plugin was loaded
	if len(mockPM.loadedPlugins) != 1 || mockPM.loadedPlugins[0] != "mock" {
		t.Errorf("Expected 'mock' plugin to be loaded, got %v", mockPM.loadedPlugins)
	}

	// Check the response
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Check that the instance state was updated
	instance, _ := mockIS.GetInstance("test-instance")
	if instance.State != "stopping" {
		t.Errorf("Expected instance state 'stopping', got '%s'", instance.State)
	}
}

// TestStartInstance tests the StartInstance gRPC handler
func TestStartInstance(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{}

	// Create a mock instance store
	mockIS := newMockInstanceStore()
	mockIS.AddInstance("test-instance", "mock")
	mockIS.UpdateInstanceState("test-instance", "stopped")

	// Create a gRPC server with mocks
	s := &GRPCServer{
		pluginManager: mockPM,
		instanceStore: mockIS,
	}

	// Create a request
	req := &gen.StartInstanceRequest{
		InstanceId: "test-instance",
	}

	// Call the handler
	resp, err := s.StartInstance(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to start instance: %v", err)
	}

	// Check that the plugin was loaded
	if len(mockPM.loadedPlugins) != 1 || mockPM.loadedPlugins[0] != "mock" {
		t.Errorf("Expected 'mock' plugin to be loaded, got %v", mockPM.loadedPlugins)
	}

	// Check the response
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	// Check that the instance state was updated
	instance, _ := mockIS.GetInstance("test-instance")
	if instance.State != "starting" {
		t.Errorf("Expected instance state 'starting', got '%s'", instance.State)
	}
}

// TestListCloudProviders tests the ListCloudProviders gRPC handler
func TestListCloudProviders(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{
		loadedPlugins: []string{"aws", "gcp"},
	}

	// Create a gRPC server with mocks
	s := &GRPCServer{
		pluginManager: mockPM,
	}

	// Create a request
	req := &gen.ListCloudProvidersRequest{}

	// Call the handler
	resp, err := s.ListCloudProviders(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to list cloud providers: %v", err)
	}

	// Check the response
	if len(resp.Providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(resp.Providers))
	}

	// Check that both providers are in the response
	found1 := false
	found2 := false
	for _, p := range resp.Providers {
		if p.Name == "aws" {
			found1 = true
		} else if p.Name == "gcp" {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Expected to find 'aws' in the response")
	}
	if !found2 {
		t.Error("Expected to find 'gcp' in the response")
	}
}