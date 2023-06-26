// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mock

import (
	"context"
	"testing"

	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PBDataPlane is a mock Packet Broker Data Plane.
type PBDataPlane struct {
	*grpc.Server
	ForwarderUp              chan *packetbroker.RoutedUplinkMessage
	ForwarderDown            chan *packetbroker.RoutedDownlinkMessage
	ForwarderDownStateChange chan *packetbroker.DownlinkMessageDeliveryStateChange
	HomeNetworkDown          chan *packetbroker.RoutedDownlinkMessage
	HomeNetworkUp            chan *packetbroker.RoutedUplinkMessage
	HomeNetworkUpStateChange chan *packetbroker.UplinkMessageDeliveryStateChange
}

// NewPBDataPlane instantiates a new mock Packet Broker Data Plane.
func NewPBDataPlane(tb testing.TB) *PBDataPlane {
	dp := &PBDataPlane{
		Server: grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				ctx = test.ContextWithTB(ctx, tb)
				return handler(ctx, req)
			}),
		),
		ForwarderUp:              make(chan *packetbroker.RoutedUplinkMessage),
		ForwarderDown:            make(chan *packetbroker.RoutedDownlinkMessage),
		ForwarderDownStateChange: make(chan *packetbroker.DownlinkMessageDeliveryStateChange),
		HomeNetworkDown:          make(chan *packetbroker.RoutedDownlinkMessage),
		HomeNetworkUp:            make(chan *packetbroker.RoutedUplinkMessage),
		HomeNetworkUpStateChange: make(chan *packetbroker.UplinkMessageDeliveryStateChange),
	}
	routingpb.RegisterForwarderDataServer(dp.Server, &routerForwarderServer{
		upCh:     dp.ForwarderUp,
		downCh:   dp.ForwarderDown,
		reportCh: dp.ForwarderDownStateChange,
	})
	routingpb.RegisterHomeNetworkDataServer(dp.Server, &routerHomeNetworkServer{
		downCh:   dp.HomeNetworkDown,
		upCh:     dp.HomeNetworkUp,
		reportCh: dp.HomeNetworkUpStateChange,
	})
	return dp
}

type routerForwarderServer struct {
	routingpb.UnimplementedForwarderDataServer
	upCh     chan *packetbroker.RoutedUplinkMessage
	downCh   chan *packetbroker.RoutedDownlinkMessage
	reportCh chan *packetbroker.DownlinkMessageDeliveryStateChange
}

func (s *routerForwarderServer) Publish(ctx context.Context, req *routingpb.PublishUplinkMessageRequest) (*routingpb.PublishUplinkMessageResponse, error) {
	s.upCh <- &packetbroker.RoutedUplinkMessage{
		ForwarderNetId:     req.ForwarderNetId,
		ForwarderTenantId:  req.ForwarderTenantId,
		ForwarderClusterId: req.ForwarderClusterId,
		Message:            req.Message,
	}
	return &routingpb.PublishUplinkMessageResponse{
		Id: "test",
	}, nil
}

func (s *routerForwarderServer) Subscribe(req *routingpb.SubscribeForwarderRequest, res routingpb.ForwarderData_SubscribeServer) error {
	for {
		select {
		case <-res.Context().Done():
			return nil
		case msg := <-s.downCh:
			if err := res.Send(msg); err != nil {
				return err
			}
		}
	}
}

func (s *routerForwarderServer) ReportDownlinkMessageDeliveryState(ctx context.Context, req *routingpb.DownlinkMessageDeliveryStateChangeRequest) (*emptypb.Empty, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.reportCh <- req.StateChange:
	}
	return ttnpb.Empty, nil
}

type routerHomeNetworkServer struct {
	routingpb.UnimplementedHomeNetworkDataServer
	downCh   chan *packetbroker.RoutedDownlinkMessage
	upCh     chan *packetbroker.RoutedUplinkMessage
	reportCh chan *packetbroker.UplinkMessageDeliveryStateChange
}

func (s *routerHomeNetworkServer) Publish(ctx context.Context, req *routingpb.PublishDownlinkMessageRequest) (*routingpb.PublishDownlinkMessageResponse, error) {
	down := &packetbroker.RoutedDownlinkMessage{
		ForwarderNetId:       req.ForwarderNetId,
		ForwarderTenantId:    req.ForwarderTenantId,
		ForwarderClusterId:   req.ForwarderClusterId,
		HomeNetworkNetId:     req.HomeNetworkNetId,
		HomeNetworkTenantId:  req.HomeNetworkTenantId,
		HomeNetworkClusterId: req.HomeNetworkClusterId,
		Message:              req.Message,
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.downCh <- down:
	}
	return &routingpb.PublishDownlinkMessageResponse{
		Id: "test",
	}, nil
}

func (s *routerHomeNetworkServer) Subscribe(req *routingpb.SubscribeHomeNetworkRequest, res routingpb.HomeNetworkData_SubscribeServer) error {
	for {
		select {
		case <-res.Context().Done():
			return nil
		case msg := <-s.upCh:
			if err := res.Send(msg); err != nil {
				return err
			}
		}
	}
}

func (s *routerHomeNetworkServer) ReportUplinkMessageDeliveryState(ctx context.Context, req *routingpb.UplinkMessageDeliveryStateChangeRequest) (*emptypb.Empty, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.reportCh <- req.StateChange:
	}
	return ttnpb.Empty, nil
}
