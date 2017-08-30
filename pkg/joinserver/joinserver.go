// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package networkserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// JoinServer implements the join server component.
//
// The join server exposes the NsJs and DeviceManagement services.
type JoinServer struct {
	*component.Component
}

// HandleJoin is called by the network server to join a device
func (js *JoinServer) HandleJoin(ctx context.Context, joinRequest *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// ListDevices is called by clients to list devices that match a filter.
func (js *JoinServer) ListDevices(ctx context.Context, filter *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevices, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// GetDevice is called by clients to get a device.
func (js *JoinServer) GetDevice(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// SetDevice is called by clients to create or update a device.
func (js *JoinServer) SetDevice(ctx context.Context, dev *ttnpb.EndDevice) (*types.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DeleteDevice is called by clients delete a device.
func (js *JoinServer) DeleteDevice(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers) (*types.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (js *JoinServer) RegisterServices(s *grpc.Server) {
	ttnpb.RegisterNsJsServer(s, js)
	ttnpb.RegisterDeviceManagementServer(s, js)
}

func (js *JoinServer) RegisterHandlers(s *runtime.ServeMux, conn *grpc.ClientConn) {

}
