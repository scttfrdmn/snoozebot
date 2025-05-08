package plugin

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// Here we provide implementations for the gRPC server and client

// GRPCCloudProviderServer is the gRPC server that GRPCCloudProviderClient talks to.
type GRPCCloudProviderServer struct {
	// This is the real implementation
	Impl CloudProvider
	UnimplementedCloudProviderServer
}

func (m *GRPCCloudProviderServer) GetInstanceInfo(ctx context.Context, req *GetInstanceInfoRequest) (*GetInstanceInfoResponse, error) {
	info, err := m.Impl.GetInstanceInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Create an InstanceInfo protobuf message
	instance := &InstanceInfo{
		Id:         info.ID,
		Name:       info.Name,
		Type:       info.Type,
		Region:     info.Region,
		Zone:       info.Zone,
		State:      info.State,
		LaunchTime: timestamppb.New(info.LaunchTime),
	}
	
	// Create response with the instance field set
	return &GetInstanceInfoResponse{
		Instance: instance,
	}, nil
}

func (m *GRPCCloudProviderServer) StopInstance(ctx context.Context, req *StopInstanceRequest) (*StopInstanceResponse, error) {
	err := m.Impl.StopInstance(ctx)
	if err != nil {
		return &StopInstanceResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &StopInstanceResponse{
		Success: true,
	}, nil
}

func (m *GRPCCloudProviderServer) StartInstance(ctx context.Context, req *StartInstanceRequest) (*StartInstanceResponse, error) {
	err := m.Impl.StartInstance(ctx)
	if err != nil {
		return &StartInstanceResponse{
			Success:      false,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &StartInstanceResponse{
		Success: true,
	}, nil
}

func (m *GRPCCloudProviderServer) GetProviderName(ctx context.Context, req *GetProviderNameRequest) (*GetProviderNameResponse, error) {
	return &GetProviderNameResponse{
		ProviderName: m.Impl.GetProviderName(),
	}, nil
}

func (m *GRPCCloudProviderServer) GetProviderVersion(ctx context.Context, req *GetProviderVersionRequest) (*GetProviderVersionResponse, error) {
	return &GetProviderVersionResponse{
		ProviderVersion: m.Impl.GetProviderVersion(),
	}, nil
}

func (m *GRPCCloudProviderServer) ListInstances(ctx context.Context, req *ListInstancesRequest) (*ListInstancesResponse, error) {
	instances, err := m.Impl.ListInstances(ctx)
	if err != nil {
		return nil, err
	}

	protoInstances := make([]*InstanceInfo, len(instances))
	for i, instance := range instances {
		protoInstances[i] = &InstanceInfo{
			Id:         instance.ID,
			Name:       instance.Name,
			Type:       instance.Type,
			Region:     instance.Region,
			Zone:       instance.Zone,
			State:      instance.State,
			LaunchTime: timestamppb.New(instance.LaunchTime),
		}
	}

	return &ListInstancesResponse{
		Instances: protoInstances,
	}, nil
}

func (m *GRPCCloudProviderServer) Shutdown(ctx context.Context, req *ShutdownRequest) (*ShutdownResponse, error) {
	// Call the implementation's Shutdown method
	m.Impl.Shutdown()
	
	return &ShutdownResponse{
		Success: true,
	}, nil
}

// GRPCCloudProviderClient is an implementation of CloudProvider that talks over gRPC.
type GRPCCloudProviderClient struct {
	client CloudProviderClient
}

func (m *GRPCCloudProviderClient) GetInstanceInfo(ctx context.Context) (*CloudInstanceInfo, error) {
	resp, err := m.client.GetInstanceInfo(ctx, &GetInstanceInfoRequest{})
	if err != nil {
		return nil, err
	}

	if resp.Instance == nil {
		return nil, fmt.Errorf("received nil instance info from server")
	}

	var launchTime time.Time
	if resp.Instance.LaunchTime != nil {
		launchTime = resp.Instance.LaunchTime.AsTime()
	}

	return &CloudInstanceInfo{
		ID:         resp.Instance.Id,
		Name:       resp.Instance.Name,
		Type:       resp.Instance.Type,
		Region:     resp.Instance.Region,
		Zone:       resp.Instance.Zone,
		State:      resp.Instance.State,
		LaunchTime: launchTime,
	}, nil
}

func (m *GRPCCloudProviderClient) StopInstance(ctx context.Context) error {
	resp, err := m.client.StopInstance(ctx, &StopInstanceRequest{})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.ErrorMessage)
	}

	return nil
}

func (m *GRPCCloudProviderClient) StartInstance(ctx context.Context) error {
	resp, err := m.client.StartInstance(ctx, &StartInstanceRequest{})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.ErrorMessage)
	}

	return nil
}

func (m *GRPCCloudProviderClient) GetProviderName() string {
	resp, err := m.client.GetProviderName(context.Background(), &GetProviderNameRequest{})
	if err != nil {
		return "unknown"
	}

	return resp.ProviderName
}

func (m *GRPCCloudProviderClient) GetProviderVersion() string {
	resp, err := m.client.GetProviderVersion(context.Background(), &GetProviderVersionRequest{})
	if err != nil {
		return "unknown"
	}

	return resp.ProviderVersion
}

func (m *GRPCCloudProviderClient) ListInstances(ctx context.Context) ([]*CloudInstanceInfo, error) {
	resp, err := m.client.ListInstances(ctx, &ListInstancesRequest{})
	if err != nil {
		return nil, err
	}

	instances := make([]*CloudInstanceInfo, len(resp.Instances))
	for i, instance := range resp.Instances {
		var launchTime time.Time
		if instance.LaunchTime != nil {
			launchTime = instance.LaunchTime.AsTime()
		}

		instances[i] = &CloudInstanceInfo{
			ID:         instance.Id,
			Name:       instance.Name,
			Type:       instance.Type,
			Region:     instance.Region,
			Zone:       instance.Zone,
			State:      instance.State,
			LaunchTime: launchTime,
		}
	}

	return instances, nil
}

func (m *GRPCCloudProviderClient) Shutdown() {
	// Send shutdown signal to server
	_, err := m.client.Shutdown(context.Background(), &ShutdownRequest{})
	if err != nil {
		// Just log error, as this is a best-effort operation
		fmt.Printf("Error during shutdown: %v\n", err)
	}
}