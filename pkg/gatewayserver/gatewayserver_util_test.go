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

package gatewayserver_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc/metadata"
)

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

type mockIS struct {
	ttnpb.GatewayRegistryServer
	ttnpb.GatewayAccessServer
	gateways     map[string]*ttnpb.Gateway
	gatewayAuths map[string][]string
}

func startMockIS(ctx context.Context) (*mockIS, string) {
	is := &mockIS{
		gateways:     make(map[string]*ttnpb.Gateway),
		gatewayAuths: make(map[string][]string),
	}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterGatewayRegistryServer(srv.Server, is)
	ttnpb.RegisterGatewayAccessServer(srv.Server, is)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return is, lis.Addr().String()
}

func (is *mockIS) add(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string, locationPublic bool, updateLocationFromStatus bool) {
	uid := unique.ID(ctx, ids)
	is.gateways[uid] = &ttnpb.Gateway{
		Ids:              &ids,
		FrequencyPlanId:  test.EUFrequencyPlanID,
		FrequencyPlanIds: []string{test.EUFrequencyPlanID},
		Antennas: []*ttnpb.GatewayAntenna{
			{
				Location: &ttnpb.Location{
					Source: ttnpb.LocationSource_SOURCE_REGISTRY,
				},
			},
		},
		LocationPublic:           locationPublic,
		UpdateLocationFromStatus: updateLocationFromStatus,
	}
	if key != "" {
		is.gatewayAuths[uid] = []string{fmt.Sprintf("Bearer %v", key)}
	}
}

var (
	errNotFound        = errors.DefineNotFound("not_found", "not found")
	errNoGatewayRights = errors.DefinePermissionDenied("no_gateway_rights", "no gateway rights")
)

func (is *mockIS) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	uid := unique.ID(ctx, req.GetGatewayIds())
	gtw, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound.New()
	}
	if gtw == nil {
		return nil, errNoGatewayRights.New() // This simulates the behaviour of the IS with a deleted gateway.
	}
	return gtw, nil
}

func (is *mockIS) Update(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
	uid := unique.ID(ctx, req.Gateway.GetIds())
	gtw, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound.New()
	}
	gtw.SetFields(req.Gateway, req.FieldMask.GetPaths()...)
	return gtw, nil
}

func (is *mockIS) Delete(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	uid := unique.ID(ctx, ids)
	_, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound.New()
	}
	is.gateways[uid] = nil
	return nil, nil
}

func (is *mockIS) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	if req.Eui.Equal(registeredGatewayEUI) {
		return &ttnpb.GatewayIdentifiers{
			GatewayId: registeredGatewayID,
			Eui:       &registeredGatewayEUI,
		}, nil
	}
	return nil, errNotFound.New()
}

func (is *mockIS) ListRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (res *ttnpb.Rights, err error) {
	res = &ttnpb.Rights{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	authorization, ok := md["authorization"]
	if !ok || len(authorization) == 0 {
		return
	}
	auths, ok := is.gatewayAuths[unique.ID(ctx, *ids)]
	if !ok {
		return
	}
	for _, auth := range auths {
		if auth == authorization[0] {
			res.Rights = append(res.Rights, ttnpb.RIGHT_GATEWAY_LINK, ttnpb.RIGHT_GATEWAY_STATUS_READ)
		}
	}
	return
}

func randomJoinRequestPayload(joinEUI, devEUI types.EUI64) []byte {
	var nwkKey types.AES128Key
	rand.Read(nwkKey[:])
	var devNonce types.DevNonce
	rand.Read(devNonce[:])

	msg := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: ttnpb.MType_JOIN_REQUEST,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			Payload: &ttnpb.Message_JoinRequestPayload{
				JoinRequestPayload: &ttnpb.JoinRequestPayload{
					JoinEui:  joinEUI,
					DevEui:   devEUI,
					DevNonce: devNonce,
				},
			},
		},
	}
	buf, err := lorawan.MarshalMessage(*msg.Payload)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeJoinRequestMIC(nwkKey, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}

func randomUpDataPayload(devAddr types.DevAddr, fPort uint32, size int) []byte {
	var fNwkSIntKey, sNwkSIntKey, appSKey types.AES128Key
	rand.Read(fNwkSIntKey[:])
	rand.Read(sNwkSIntKey[:])
	rand.Read(appSKey[:])

	pld := &ttnpb.MACPayload{
		FHdr: &ttnpb.FHDR{
			DevAddr: devAddr,
			FCnt:    42,
		},
		FPort:      fPort,
		FrmPayload: random.Bytes(size),
	}
	buf, err := crypto.EncryptUplink(appSKey, devAddr, pld.FHdr.FCnt, pld.FrmPayload, false)
	if err != nil {
		panic(err)
	}
	pld.FrmPayload = buf

	msg := &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: ttnpb.MType_UNCONFIRMED_UP,
				Major: ttnpb.Major_LORAWAN_R1,
			},
			Payload: &ttnpb.Message_MacPayload{
				MacPayload: pld,
			},
		},
	}
	buf, err = lorawan.MarshalMessage(*msg.Payload)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeUplinkMIC(sNwkSIntKey, fNwkSIntKey, 0, 5, 0, devAddr, pld.FHdr.FCnt, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}

func randomDownDataPayload(devAddr types.DevAddr, fPort uint32, size int) []byte {
	var sNwkSIntKey, appSKey types.AES128Key
	rand.Read(sNwkSIntKey[:])
	rand.Read(appSKey[:])

	pld := &ttnpb.MACPayload{
		FHdr: &ttnpb.FHDR{
			DevAddr: devAddr,
			FCnt:    42,
		},
		FPort:      fPort,
		FrmPayload: random.Bytes(size),
	}
	buf, err := crypto.EncryptDownlink(appSKey, devAddr, pld.FHdr.FCnt, pld.FrmPayload, false)
	if err != nil {
		panic(err)
	}
	pld.FrmPayload = buf

	msg := ttnpb.Message{
		MHdr: &ttnpb.MHDR{
			MType: ttnpb.MType_UNCONFIRMED_DOWN,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MacPayload{
			MacPayload: pld,
		},
	}
	buf, err = lorawan.MarshalMessage(msg)
	if err != nil {
		panic(err)
	}
	mic, err := crypto.ComputeDownlinkMIC(sNwkSIntKey, devAddr, 0, pld.FHdr.FCnt, buf)
	if err != nil {
		panic(err)
	}
	return append(buf, mic[:]...)
}
