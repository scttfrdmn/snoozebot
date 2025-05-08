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

	return &GetInstanceInfoResponse{
		Id:         info.ID,
		Name:       info.Name,
		Type:       info.Type,
		Region:     info.Region,
		Zone:       info.Zone,
		State:      info.State,
		LaunchTime: timestamppb.New(info.LaunchTime),
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

// GRPCCloudProviderClient is an implementation of CloudProvider that talks over gRPC.
type GRPCCloudProviderClient struct {
	client CloudProviderClient
}

func (m *GRPCCloudProviderClient) GetInstanceInfo(ctx context.Context) (*InstanceInfo, error) {
	resp, err := m.client.GetInstanceInfo(ctx, &GetInstanceInfoRequest{})
	if err != nil {
		return nil, err
	}

	var launchTime time.Time
	if resp.LaunchTime != nil {
		launchTime = resp.LaunchTime.AsTime()
	}

	return &InstanceInfo{
		ID:         resp.Id,
		Name:       resp.Name,
		Type:       resp.Type,
		Region:     resp.Region,
		Zone:       resp.Zone,
		State:      resp.State,
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