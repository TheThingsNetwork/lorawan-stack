// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package managed

import (
	"context"

	northboundv1 "go.thethings.industries/pkg/api/gen/tti/gateway/controller/northbound/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttgc"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type managedGCSServer struct {
	Component
	ttnpb.UnsafeManagedGatewayConfigurationServiceServer

	gatewayEUIs []types.EUI64Prefix
	client      *ttgc.Client
}

var (
	_ ttnpb.ManagedGatewayConfigurationServiceServer = (*managedGCSServer)(nil)

	errGatewayNotManaged = errors.DefineNotFound(
		"gateway_not_managed", "gateway `{gateway_id}` is not managed by The Things Gateway Controller",
	)
)

// managedGatewayID looks up the gateway EUI for the given gateway identifiers and returns a copy of the gateway
// identifiers with the EUI filled.
// If the EUI does not match the prefix configured for The Things Gateway Controller, this method returns NotFound.
func (s *managedGCSServer) managedGatewayID(
	ctx context.Context, ids *ttnpb.GatewayIdentifiers,
) (*ttnpb.GatewayIdentifiers, error) {
	cc, err := s.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	callOpt, err := rpcmetadata.WithForwardedAuth(ctx, s.AllowInsecureForCredentials())
	if err != nil {
		return nil, err
	}
	gtw, err := ttnpb.NewGatewayRegistryClient(cc).Get(ctx, &ttnpb.GetGatewayRequest{
		GatewayIds: ids,
		FieldMask:  &fieldmaskpb.FieldMask{Paths: []string{"ids.eui"}},
	}, callOpt)
	if err != nil {
		return nil, err
	}
	if len(gtw.Ids.Eui) == 0 {
		return nil, errGatewayNotManaged.WithAttributes("gateway_id", ids.GatewayId)
	}
	var (
		matchesPrefix bool
		eui           = types.MustEUI64(gtw.Ids.Eui).OrZero()
	)
	for _, prefix := range s.gatewayEUIs {
		if prefix.Matches(eui) {
			matchesPrefix = true
			break
		}
	}
	if !matchesPrefix {
		return nil, errGatewayNotManaged.WithAttributes("gateway_id", ids.GatewayId)
	}
	return &ttnpb.GatewayIdentifiers{
		GatewayId: ids.GatewayId,
		Eui:       gtw.Ids.Eui,
	}, nil
}

// Get implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (s *managedGCSServer) Get(
	ctx context.Context,
	req *ttnpb.GetGatewayRequest,
) (*ttnpb.ManagedGateway, error) {
	if err := rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_INFO); err != nil {
		return nil, err
	}
	ids, err := s.managedGatewayID(ctx, req.GatewayIds)
	if err != nil {
		return nil, err
	}
	res, err := northboundv1.NewGatewayServiceClient(s.client).Get(ctx, &northboundv1.GatewayServiceGetRequest{
		Domain:    s.client.Domain(ctx),
		GatewayId: types.MustEUI64(ids.Eui).OrZero().MarshalNumber(),
	})
	if err != nil {
		return nil, err
	}
	paths := ttnpb.AddFields(req.FieldMask.GetPaths(), "ids")
	return toManagedGateway(ids, res.Gateway, paths)
}

// ScanWiFiAccessPoints implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (s *managedGCSServer) ScanWiFiAccessPoints(
	ctx context.Context,
	req *ttnpb.GatewayIdentifiers,
) (*ttnpb.ManagedGatewayWiFiAccessPoints, error) {
	if err := rights.RequireGateway(ctx, req, ttnpb.Right_RIGHT_GATEWAY_INFO); err != nil {
		return nil, err
	}
	ids, err := s.managedGatewayID(ctx, req)
	if err != nil {
		return nil, err
	}
	res, err := northboundv1.NewGatewayServiceClient(s.client).ScanWiFiAccessPoints(
		ctx,
		&northboundv1.GatewayServiceScanWiFiAccessPointsRequest{
			Domain:    s.client.Domain(ctx),
			GatewayId: types.MustEUI64(ids.Eui).OrZero().MarshalNumber(),
		},
	)
	if err != nil {
		return nil, err
	}
	ret := &ttnpb.ManagedGatewayWiFiAccessPoints{
		AccessPoints: make([]*ttnpb.ManagedGatewayWiFiAccessPoint, len(res.AccessPoints)),
	}
	for i, ap := range res.AccessPoints {
		ret.AccessPoints[i] = toWiFiAccessPoint(ap)
	}
	return ret, nil
}

// StreamEvents implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (s *managedGCSServer) StreamEvents(
	req *ttnpb.GatewayIdentifiers,
	stream ttnpb.ManagedGatewayConfigurationService_StreamEventsServer,
) error {
	ctx := stream.Context()
	if err := rights.RequireGateway(ctx, req,
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
		ttnpb.Right_RIGHT_GATEWAY_LOCATION_READ,
	); err != nil {
		return err
	}
	ids, err := s.managedGatewayID(ctx, req)
	if err != nil {
		return err
	}
	sub, err := northboundv1.NewGatewayServiceClient(s.client).Subscribe(
		ctx,
		&northboundv1.GatewayServiceSubscribeRequest{
			Domain:    s.client.Domain(ctx),
			GatewayId: types.MustEUI64(ids.Eui).OrZero().MarshalNumber(),
		},
	)
	if err != nil {
		return err
	}
	for {
		msg, err := sub.Recv()
		if err != nil {
			return err
		}
		evt := toEvent(ids, msg)
		if evt == nil {
			continue
		}
		if err := stream.Send(evt); err != nil {
			return err
		}
	}
}

// Update implements ttnpb.ManagedGatewayConfigurationServiceServer.
func (s *managedGCSServer) Update(
	ctx context.Context,
	req *ttnpb.UpdateManagedGatewayRequest,
) (*ttnpb.ManagedGateway, error) {
	reqGtw := req.GetGateway()
	if err := rights.RequireGateway(ctx, reqGtw.GetIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	ids, err := s.managedGatewayID(ctx, reqGtw.Ids)
	if err != nil {
		return nil, err
	}
	innerReq := &northboundv1.GatewayServiceUpdateRequest{
		Domain:    s.client.Domain(ctx),
		GatewayId: types.MustEUI64(ids.Eui).OrZero().MarshalNumber(),
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "ethernet_profile_id") {
		innerReq.EthernetProfileId = fromProfileIDOrNil(reqGtw.EthernetProfileId)
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "wifi_profile_id") {
		innerReq.WifiProfileId = fromProfileIDOrNil(reqGtw.WifiProfileId)
	}
	_, err = northboundv1.NewGatewayServiceClient(s.client).Update(ctx, innerReq)
	if err != nil {
		return nil, err
	}
	ret := &ttnpb.ManagedGateway{}
	if ret.SetFields(reqGtw, req.FieldMask.GetPaths()...) != nil {
		return nil, err
	}
	return ret, nil
}
