package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/agent/provider"
)

// mockPluginManager is a mock implementation of provider.PluginManager for testing
type mockPluginManager struct {
	discoveredPlugins []string
	loadedPlugins     []string
	loadError         error
	unloadError       error
}

func (m *mockPluginManager) DiscoverPlugins() ([]string, error) {
	return m.discoveredPlugins, nil
}

func (m *mockPluginManager) LoadPlugin(ctx interface{}, pluginName string) (provider.CloudProvider, error) {
	if m.loadError != nil {
		return nil, m.loadError
	}
	// Add the plugin to loaded plugins
	m.loadedPlugins = append(m.loadedPlugins, pluginName)
	return &mockCloudProvider{name: pluginName}, nil
}

func (m *mockPluginManager) UnloadPlugin(pluginName string) error {
	if m.unloadError != nil {
		return m.unloadError
	}
	// Remove the plugin from loaded plugins
	for i, p := range m.loadedPlugins {
		if p == pluginName {
			m.loadedPlugins = append(m.loadedPlugins[:i], m.loadedPlugins[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockPluginManager) GetPlugin(pluginName string) (provider.CloudProvider, error) {
	for _, p := range m.loadedPlugins {
		if p == pluginName {
			return &mockCloudProvider{name: pluginName}, nil
		}
	}
	return nil, nil
}

func (m *mockPluginManager) ListPlugins() []string {
	return m.loadedPlugins
}

// mockCloudProvider is a mock implementation of provider.CloudProvider for testing
type mockCloudProvider struct {
	name string
}

func (m *mockCloudProvider) GetInstanceInfo(ctx context.Context, instanceID string) (*provider.InstanceInfo, error) {
	return nil, nil
}

func (m *mockCloudProvider) StopInstance(ctx context.Context, instanceID string) error {
	return nil
}

func (m *mockCloudProvider) StartInstance(ctx context.Context, instanceID string) error {
	return nil
}

func (m *mockCloudProvider) GetProviderName() string {
	return m.name
}

func (m *mockCloudProvider) GetProviderVersion() string {
	return "1.0.0"
}

// TestHandleListPlugins tests the handleListPlugins handler
func TestHandleListPlugins(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{
		loadedPlugins: []string{"aws", "gcp"},
	}

	// Create a server with the mock plugin manager
	s := &Server{
		pluginManager: mockPM,
		logger: hclog.NewNullLogger(),
	}

	// Create a request
	req, err := http.NewRequest("GET", "/plugins", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(s.handleListPlugins)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response struct {
		Plugins []string `json:"plugins"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check the plugins
	if len(response.Plugins) != 2 {
		t.Errorf("Expected 2 plugins, got %d", len(response.Plugins))
	}

	// Check that both plugins are in the response
	found1 := false
	found2 := false
	for _, p := range response.Plugins {
		if p == "aws" {
			found1 = true
		} else if p == "gcp" {
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

// TestHandleDiscoverPlugins tests the handleDiscoverPlugins handler
func TestHandleDiscoverPlugins(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{
		discoveredPlugins: []string{"aws", "gcp", "azure"},
	}

	// Create a server with the mock plugin manager
	s := &Server{
		pluginManager: mockPM,
		logger: hclog.NewNullLogger(),
		pluginsDir: "/plugins",
	}

	// Create a request
	req, err := http.NewRequest("GET", "/plugins/discover", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(s.handleDiscoverPlugins)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var response struct {
		Plugins []string `json:"plugins"`
	}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check the plugins
	if len(response.Plugins) != 3 {
		t.Errorf("Expected 3 plugins, got %d", len(response.Plugins))
	}

	// Check that all plugins are in the response
	found1 := false
	found2 := false
	found3 := false
	for _, p := range response.Plugins {
		if p == "aws" {
			found1 = true
		} else if p == "gcp" {
			found2 = true
		} else if p == "azure" {
			found3 = true
		}
	}

	if !found1 {
		t.Error("Expected to find 'aws' in the response")
	}
	if !found2 {
		t.Error("Expected to find 'gcp' in the response")
	}
	if !found3 {
		t.Error("Expected to find 'azure' in the response")
	}
}

// TestHandleLoadPlugin tests the handleLoadPlugin handler
func TestHandleLoadPlugin(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{}

	// Create a server with the mock plugin manager
	s := &Server{
		pluginManager: mockPM,
		logger: hclog.NewNullLogger(),
	}

	// Create a request
	reqBody := `{"plugin_name": "aws"}`
	req, err := http.NewRequest("POST", "/plugins/load", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(s.handleLoadPlugin)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the plugin was loaded
	if len(mockPM.loadedPlugins) != 1 || mockPM.loadedPlugins[0] != "aws" {
		t.Errorf("Expected 'aws' plugin to be loaded, got %v", mockPM.loadedPlugins)
	}
}

// TestHandleUnloadPlugin tests the handleUnloadPlugin handler
func TestHandleUnloadPlugin(t *testing.T) {
	// Create a mock plugin manager
	mockPM := &mockPluginManager{
		loadedPlugins: []string{"aws", "gcp"},
	}

	// Create a server with the mock plugin manager
	s := &Server{
		pluginManager: mockPM,
		logger: hclog.NewNullLogger(),
	}

	// Create a request
	reqBody := `{"plugin_name": "aws"}`
	req, err := http.NewRequest("POST", "/plugins/unload", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(s.handleUnloadPlugin)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check that the plugin was unloaded
	if len(mockPM.loadedPlugins) != 1 || mockPM.loadedPlugins[0] != "gcp" {
		t.Errorf("Expected 'aws' plugin to be unloaded, got %v", mockPM.loadedPlugins)
	}
}