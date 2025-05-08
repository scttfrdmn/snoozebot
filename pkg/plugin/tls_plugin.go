package plugin

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	plugintls "github.com/scttfrdmn/snoozebot/pkg/plugin/tls"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// SecurePlugin is a wrapper for a plugin with TLS support
type SecurePlugin struct {
	PluginName     string
	PluginCmd      *exec.Cmd
	TLSConfig      *tls.Config
	Logger         hclog.Logger
	Client         *plugin.Client
	CloudProvider  CloudProvider
}

// TLSOptions contains TLS configuration options
type TLSOptions struct {
	Enabled   bool   // Whether to use TLS
	CertDir   string // Directory for certificates
	CACert    string // CA certificate file
	CertFile  string // Certificate file
	KeyFile   string // Private key file
	SkipVerify bool   // Skip certificate verification (not recommended for production)
}

// NewSecurePlugin creates a new secure plugin
func NewSecurePlugin(pluginName string, pluginPath string, tlsOptions *TLSOptions, logger hclog.Logger) (*SecurePlugin, error) {
	if logger == nil {
		logger = hclog.NewNullLogger()
	}

	// Create plugin command
	pluginCmd := exec.Command(pluginPath)

	var tlsConfig *tls.Config
	var err error

	// Configure TLS if enabled
	if tlsOptions != nil && tlsOptions.Enabled {
		if tlsOptions.CertFile != "" && tlsOptions.KeyFile != "" {
			// Load custom TLS config from certificate files
			tlsConfig, err = plugintls.LoadTLSConfig(tlsOptions.CertFile, tlsOptions.KeyFile, tlsOptions.CACert)
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS config: %w", err)
			}
		} else if tlsOptions.CertDir != "" {
			// Create TLS manager
			tlsManager, err := plugintls.NewTLSManager(tlsOptions.CertDir)
			if err != nil {
				return nil, fmt.Errorf("failed to create TLS manager: %w", err)
			}

			// Initialize TLS manager
			if err := tlsManager.Initialize(); err != nil {
				return nil, fmt.Errorf("failed to initialize TLS manager: %w", err)
			}

			// Get TLS config
			tlsConfig, err = tlsManager.GetClientTLSConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to get client TLS config: %w", err)
			}
		} else {
			return nil, fmt.Errorf("TLS is enabled but no certificate files or directory provided")
		}

		// Skip verification if requested
		if tlsOptions.SkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
	}

	return &SecurePlugin{
		PluginName: pluginName,
		PluginCmd:  pluginCmd,
		TLSConfig:  tlsConfig,
		Logger:     logger,
	}, nil
}

// verifyClientCertificate verifies the client certificate against the CA
func verifyClientCertificate(tlsConfig *tls.Config, caFile string, pluginName string, logger hclog.Logger) error {
	// Skip verification if requested
	if tlsConfig.InsecureSkipVerify {
		logger.Warn("TLS certificate verification disabled - INSECURE")
		return nil
	}

	// Read CA certificate
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Create cert pool and add CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return fmt.Errorf("failed to append CA certificate to cert pool")
	}

	// Set CA cert pool in TLS config
	tlsConfig.RootCAs = certPool
	
	// Set the server name to the plugin name for verification
	// This helps ensure the plugin's certificate is issued for the correct name
	tlsConfig.ServerName = pluginName
	
	logger.Info("TLS certificate verification enabled for plugin", 
		"plugin", pluginName, 
		"ca", caFile,
		"server_name", tlsConfig.ServerName)
	return nil
}

// Start starts the plugin
func (p *SecurePlugin) Start() error {
	p.Logger.Info("Starting secure plugin", "name", p.PluginName)
	
	// Create plugin client configuration
	clientConfig := &plugin.ClientConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &CloudProviderPlugin{Impl: nil},
		},
		Cmd:              p.PluginCmd,
		Logger:           p.Logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	}

	// Add TLS configuration if available
	if p.TLSConfig != nil {
		p.Logger.Info("Using TLS configuration", "skip_verify", p.TLSConfig.InsecureSkipVerify)
		
		// Find CA certificate path
		var caFile string
		if p.PluginCmd != nil && p.PluginCmd.Env != nil {
			for _, env := range p.PluginCmd.Env {
				if strings.HasPrefix(env, "SNOOZEBOT_TLS_CA_FILE=") {
					caFile = strings.TrimPrefix(env, "SNOOZEBOT_TLS_CA_FILE=")
					break
				}
			}
		}
		
		// Verify client certificate if CA file is available
		if caFile != "" && !p.TLSConfig.InsecureSkipVerify {
			if err := verifyClientCertificate(p.TLSConfig, caFile, p.PluginName, p.Logger); err != nil {
				return fmt.Errorf("certificate verification failed: %w", err)
			}
		}
		
		clientConfig.TLSConfig = p.TLSConfig
		clientConfig.SecureConfig = &plugin.SecureConfig{
			TLSConfig: p.TLSConfig,
		}
	}

	// Create plugin client
	p.Logger.Info("Creating plugin client")
	client := plugin.NewClient(clientConfig)
	p.Client = client

	// Connect to the plugin
	p.Logger.Info("Connecting to plugin")
	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Request the plugin
	p.Logger.Info("Dispensing plugin")
	raw, err := rpcClient.Dispense("cloud_provider")
	if err != nil {
		return fmt.Errorf("failed to dispense plugin: %w", err)
	}

	// Cast to CloudProvider interface
	cp, ok := raw.(CloudProvider)
	if !ok {
		return fmt.Errorf("plugin does not implement CloudProvider interface")
	}

	p.CloudProvider = cp
	p.Logger.Info("Plugin started successfully", 
		"name", p.PluginName, 
		"provider", cp.GetProviderName(), 
		"version", cp.GetProviderVersion())
	return nil
}

// Stop stops the plugin
func (p *SecurePlugin) Stop() {
	if p.Client != nil {
		p.Client.Kill()
	}
}

// GetCloudProvider returns the cloud provider
func (p *SecurePlugin) GetCloudProvider() CloudProvider {
	return p.CloudProvider
}

// GRPCCloudProviderPlugin is a wrapper for the CloudProviderPlugin
// that adds TLS support to the gRPC connection
type GRPCCloudProviderPlugin struct {
	CloudProviderPlugin
	TLSConfig *tls.Config
}

// GRPCServer registers the plugin for serving with a gRPC server
// with TLS support
func (p *GRPCCloudProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	RegisterCloudProviderServer(s, &GRPCCloudProviderServer{Impl: p.Impl})
	return nil
}

// GRPCClient returns the client for this plugin with TLS support
func (p *GRPCCloudProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCCloudProviderClient{client: NewCloudProviderClient(c)}, nil
}

// verifyServerCertificate verifies the server certificate against the CA
func verifyServerCertificate(tlsConfig *tls.Config, caFile string, logger hclog.Logger) error {
	// Skip verification if requested
	if tlsConfig.InsecureSkipVerify {
		logger.Warn("TLS certificate verification disabled - INSECURE")
		return nil
	}

	// Read CA certificate
	caPEM, err := os.ReadFile(caFile)
	if err != nil {
		return fmt.Errorf("failed to read CA certificate: %w", err)
	}

	// Create cert pool and add CA certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caPEM) {
		return fmt.Errorf("failed to append CA certificate to cert pool")
	}

	// Set cert pool in TLS config
	tlsConfig.RootCAs = certPool
	tlsConfig.ClientCAs = certPool
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	logger.Info("TLS certificate verification enabled with CA", "ca", caFile)
	return nil
}

// ServePluginWithTLS serves a plugin with TLS support
func ServePluginWithTLS(pluginImpl CloudProvider, tlsOptions *TLSOptions, logger hclog.Logger) {
	if logger == nil {
		logger = hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Info,
			Output:     os.Stderr,
			JSONFormat: true,
		})
	}

	var tlsConfig *tls.Config
	var err error

	// Configure TLS if enabled
	if tlsOptions != nil && tlsOptions.Enabled {
		if tlsOptions.CertFile != "" && tlsOptions.KeyFile != "" {
			// Load custom TLS config from certificate files
			tlsConfig, err = plugintls.LoadTLSConfig(tlsOptions.CertFile, tlsOptions.KeyFile, tlsOptions.CACert)
			if err != nil {
				logger.Error("Failed to load TLS config", "error", err)
				os.Exit(1)
			}

			// Verify server certificate against CA if CA certificate is provided
			if tlsOptions.CACert != "" && !tlsOptions.SkipVerify {
				if err := verifyServerCertificate(tlsConfig, tlsOptions.CACert, logger); err != nil {
					logger.Error("Failed to verify server certificate", "error", err)
					os.Exit(1)
				}
			}
		} else if tlsOptions.CertDir != "" {
			// Create TLS manager
			tlsManager, err := plugintls.NewTLSManager(tlsOptions.CertDir)
			if err != nil {
				logger.Error("Failed to create TLS manager", "error", err)
				os.Exit(1)
			}

			// Initialize TLS manager
			if err := tlsManager.Initialize(); err != nil {
				logger.Error("Failed to initialize TLS manager", "error", err)
				os.Exit(1)
			}

			// Get TLS config
			tlsConfig, err = tlsManager.GetServerTLSConfig()
			if err != nil {
				logger.Error("Failed to get server TLS config", "error", err)
				os.Exit(1)
			}

			// Verify server certificate against CA if not skipping verification
			if !tlsOptions.SkipVerify {
				caCertFile := filepath.Join(tlsOptions.CertDir, "ca", "cert.pem")
				if err := verifyServerCertificate(tlsConfig, caCertFile, logger); err != nil {
					logger.Error("Failed to verify server certificate", "error", err)
					os.Exit(1)
				}
			}
		} else {
			logger.Error("TLS is enabled but no certificate files or directory provided")
			os.Exit(1)
		}
	}

	// Create plugin server configuration
	serveConfig := &plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"cloud_provider": &CloudProviderPlugin{Impl: pluginImpl},
		},
		GRPCServer: func(opts []grpc.ServerOption) *grpc.Server {
			// Add TLS credentials if available
			if tlsConfig != nil {
				logger.Info("Adding TLS credentials to gRPC server")
				opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
			}
			return grpc.NewServer(opts...)
		},
		Logger: logger,
	}

	// Serve the plugin
	logger.Info("Serving plugin with TLS")
	plugin.Serve(serveConfig)
}

// Creates and returns a gRPC server with TLS support
func DefaultGRPCServerWithTLS(tlsConfig *tls.Config) func([]grpc.ServerOption) *grpc.Server {
	return func(opts []grpc.ServerOption) *grpc.Server {
		// Add TLS credentials if available
		if tlsConfig != nil {
			opts = append(opts, grpc.Creds(credentials.NewTLS(tlsConfig)))
		}
		return grpc.NewServer(opts...)
	}
}