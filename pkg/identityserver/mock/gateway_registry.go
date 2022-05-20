// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package mockis

import (
	"context"
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc/metadata"
)

var errNoGatewayRights = errors.DefinePermissionDenied("no_gateway_rights", "no gateway rights")

// DefaultGateway generates a gateway with values that is adequate for most test cases.
func DefaultGateway(ids *ttnpb.GatewayIdentifiers, locationPublic, updateLocationFromStatus bool) *ttnpb.Gateway {
	return &ttnpb.Gateway{
		Ids:              ids,
		FrequencyPlanId:  test.EUFrequencyPlanID,
		FrequencyPlanIds: []string{test.EUFrequencyPlanID},
		Antennas: []*ttnpb.GatewayAntenna{
			{
				Location: &ttnpb.Location{
					Source: ttnpb.LocationSource_SOURCE_REGISTRY,
				},
			},
		},
		GatewayServerAddress:     "mockgatewayserver",
		LocationPublic:           locationPublic,
		UpdateLocationFromStatus: updateLocationFromStatus,
	}
}

type mockISGatewayRegistry struct {
	ttnpb.GatewayRegistryServer
	ttnpb.GatewayAccessServer
	gateways      map[string]*ttnpb.Gateway
	gatewayAuths  map[string][]string
	gatewayRights map[string]authKeyToRights

	registeredGateway *ttnpb.GatewayIdentifiers
}

func (is *mockISGatewayRegistry) SetRegisteredGateway(gtwIDs *ttnpb.GatewayIdentifiers) {
	is.registeredGateway = gtwIDs
}

func (is *mockISGatewayRegistry) Add(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string, gtw *ttnpb.Gateway, rights ...ttnpb.Right) {
	uid := unique.ID(ctx, ids)
	is.gateways[uid] = gtw

	var bearerKey string
	if key != "" {
		bearerKey = fmt.Sprintf("Bearer %v", key)
		is.gatewayAuths[uid] = []string{bearerKey}
	}
	if is.gatewayRights[uid] == nil {
		is.gatewayRights[uid] = make(authKeyToRights)
	}
	is.gatewayRights[uid][bearerKey] = rights
}

func (is *mockISGatewayRegistry) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	uid := unique.ID(ctx, req.GetGatewayIds())
	gtw, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound.New()
	}
	if gtw == nil {
		return nil, errNoGatewayRights.New()
	}
	return gtw, nil
}

func (is *mockISGatewayRegistry) Update(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
	uid := unique.ID(ctx, req.Gateway.GetIds())
	gtw, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound.New()
	}
	gtw.SetFields(req.Gateway, req.FieldMask.GetPaths()...)
	return gtw, nil
}

func (is *mockISGatewayRegistry) Delete(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	uid := unique.ID(ctx, ids)
	_, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound.New()
	}
	is.gateways[uid] = nil
	return nil, nil
}

func (is *mockISGatewayRegistry) ListRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (res *ttnpb.Rights, err error) {
	res = &ttnpb.Rights{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	authorization, ok := md["authorization"]
	if !ok || len(authorization) == 0 {
		return
	}

	uid := unique.ID(ctx, *ids)
	auths, ok := is.gatewayAuths[uid]
	if !ok {
		return
	}
	for _, auth := range auths {
		if auth == authorization[0] && is.gatewayRights[uid] != nil {
			res.Rights = append(res.Rights, is.gatewayRights[uid][auth]...)
		}
	}
	return
}

func (is *mockISGatewayRegistry) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	if is.registeredGateway == nil {
		return nil, errNotFound.New()
	}
	if types.MustEUI64(req.Eui).OrZero().Equal(*is.registeredGateway.Eui) {
		return is.registeredGateway, nil
	}
	return nil, errNotFound.New()
}
