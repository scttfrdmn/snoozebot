package plugin

import (
	"context"
	"fmt"
	"net/rpc"
	"time"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't terribly important. You may want to share it under a different
	// name or value.
	ProtocolVersion:  1,
	MagicCookieKey:   "SNOOZEBOT_PLUGIN",
	MagicCookieValue: "snoozebot_plugin_v1",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"cloud_provider": &CloudProviderPlugin{Impl: nil},
}

// InstanceInfo contains information about a cloud instance
type InstanceInfo struct {
	ID         string
	Name       string
	Type       string
	Region     string
	Zone       string
	State      string
	LaunchTime time.Time
}

// CloudProvider is the interface that we expose for cloud provider plugins
type CloudProvider interface {
	// GetInstanceInfo gets information about the current instance
	GetInstanceInfo(ctx context.Context) (*InstanceInfo, error)
	
	// StopInstance stops the current instance
	StopInstance(ctx context.Context) error
	
	// StartInstance starts the current instance
	StartInstance(ctx context.Context) error
	
	// GetProviderName returns the name of the cloud provider
	GetProviderName() string
	
	// GetProviderVersion returns the version of the cloud provider plugin
	GetProviderVersion() string
}

// CloudProviderPlugin is the implementation of plugin.Plugin so we can serve/consume this.
type CloudProviderPlugin struct {
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl CloudProvider
}

// Server implements plugin.Plugin interface for serving
func (p *CloudProviderPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	// We're not using this method since we're using gRPC
	return nil, fmt.Errorf("not implemented")
}

// Client implements plugin.Plugin interface for consuming
func (p *CloudProviderPlugin) Client(*plugin.MuxBroker, *rpc.Client) (interface{}, error) {
	// We're not using this method since we're using gRPC
	return nil, fmt.Errorf("not implemented")
}

// GRPCServer registers this plugin for serving with a gRPC server.
func (p *CloudProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterCloudProviderServer(s, &GRPCCloudProviderServer{Impl: p.Impl})
	return nil
}

// GRPCClient returns the client for this plugin.
func (p *CloudProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCCloudProviderClient{client: NewCloudProviderClient(c)}, nil
}