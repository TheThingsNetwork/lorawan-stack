// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package deviceclaimingserver_test

import (
	"context"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/gateways"
	dcstypes "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	claimAuthCode = []byte("test-code")
	userID        = ttnpb.UserIdentifiers{
		UserId: "test-user",
	}
	authorizedMD = rpcmetadata.MD{
		AuthType:  "Bearer",
		AuthValue: "foo",
	}
	authorizedCallOpt = grpc.PerRPCCredentials(authorizedMD)
)

func TestGatewayClaimingServer(t *testing.T) { //nolint:paralleltest
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	t.Cleanup(func() {
		cancelCtx()
	})

	supportedEUI := types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x01}

	unAuthorizedCallOpt := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:  "Bearer",
		AuthValue: "invalid-key",
	})

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})

	mockGatewayClaimer := &MockGatewayClaimer{
		IsManagedGatewayFunc: func(_ context.Context, e types.EUI64) (bool, error) {
			return e.Equal(supportedEUI), nil
		},
	}
	mockUpstream, err := gateways.NewUpstream(
		ctx,
		c,
		gateways.Config{},
		gateways.WithClaimer(
			"mock",
			[]dcstypes.EUI64Range{
				dcstypes.RangeFromEUI64Prefix(types.EUI64Prefix{
					EUI64:  types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x00},
					Length: 48,
				}),
			},
			mockGatewayClaimer,
		),
	)
	if err != nil {
		t.FailNow()
	}

	existingEUI := types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0xFF}
	mockGatewayRegistry := &mockGatewayRegistry{
		authorizedMD: authorizedMD,
		gateways: []*ttnpb.Gateway{
			{
				Ids: &ttnpb.GatewayIdentifiers{
					GatewayId: "test-gateway",
					Eui:       existingEUI.Bytes(),
				},
			},
		},
	}

	test.Must(deviceclaimingserver.New(c,
		&deviceclaimingserver.Config{},
		deviceclaimingserver.WithGatewayClaimingServer(
			mockUpstream,
			mockGatewayRegistry,
			c,
		),
	))
	componenttest.StartComponent(t, c)
	t.Cleanup(func() {
		c.Close()
	})

	// Wait for server to be ready.
	time.Sleep(timeout)

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)
	gclsClient := ttnpb.NewGatewayClaimingServerClient(c.LoopbackConn())

	// Check that AuthorizeGateway and UnauthorizeGateway are not implemented.
	_, err = gclsClient.AuthorizeGateway(ctx, &ttnpb.AuthorizeGatewayRequest{ // nolint:staticcheck
		GatewayIds: &ttnpb.GatewayIdentifiers{
			GatewayId: "test-gateway",
		},
		ApiKey: "foo",
	}, authorizedCallOpt)
	a.So(errors.IsUnimplemented(err), should.BeTrue)

	_, err = gclsClient.UnauthorizeGateway(ctx, &ttnpb.GatewayIdentifiers{ // nolint:staticcheck
		GatewayId: "test-gateway",
	}, authorizedCallOpt)
	a.So(errors.IsUnimplemented(err), should.BeTrue)

	// Test GetInfoByGatewayEUI
	_, err = gclsClient.GetInfoByGatewayEUI(
		ctx,
		&ttnpb.GetInfoByGatewayEUIRequest{
			Eui: types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x00}.Bytes(),
		},
	)
	a.So(errors.IsUnauthenticated(err), should.BeTrue)

	unsupportedEUI := types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x90, 0x00, 0x00}
	resp, err := gclsClient.GetInfoByGatewayEUI(
		ctx,
		&ttnpb.GetInfoByGatewayEUIRequest{
			Eui: unsupportedEUI.Bytes(),
		},
		authorizedCallOpt,
	)
	a.So(err, should.BeNil)
	a.So(resp.Eui, should.Resemble, unsupportedEUI.Bytes())
	a.So(resp.SupportsClaiming, should.BeFalse)
	a.So(resp.IsManaged, should.BeFalse)

	resp, err = gclsClient.GetInfoByGatewayEUI(
		ctx,
		&ttnpb.GetInfoByGatewayEUIRequest{
			Eui: supportedEUI.Bytes(),
		},
		authorizedCallOpt,
	)
	a.So(err, should.BeNil)
	a.So(resp.Eui, should.Resemble, supportedEUI.Bytes())
	a.So(resp.SupportsClaiming, should.BeTrue)
	a.So(resp.IsManaged, should.BeTrue)

	// Test claiming
	for _, tc := range []struct {
		Name           string
		Req            *ttnpb.ClaimGatewayRequest
		CallOpt        grpc.CallOption
		ClaimFunc      func(context.Context, types.EUI64, string, string) (*dcstypes.GatewayMetadata, error)
		CreateFunc     func(context.Context, *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error)
		UnclaimFunc    func(context.Context, types.EUI64) error
		ErrorAssertion func(error) bool
	}{
		{
			Name: "Claim/EmptyRequest",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/NilCollaborator",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: nil,
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/InvalidGatewayID",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "&-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/GatewayEUIAlreadyExists",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         existingEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsAlreadyExists,
		},
		{
			Name: "Claim/EUINotRegisteredForClaiming",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         unsupportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/ClaimFailed",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt: authorizedCallOpt,
			ClaimFunc: func(_ context.Context, _ types.EUI64, _, _ string) (*dcstypes.GatewayMetadata, error) {
				return nil, errClaim.New()
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/CreateFailed",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt: authorizedCallOpt,
			ClaimFunc: func(context.Context, types.EUI64, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(context.Context, *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, errCreate.New()
			},
			UnclaimFunc: func(_ context.Context, eui types.EUI64) error {
				if eui.Equal(supportedEUI) {
					return nil
				}
				return errUnclaim.New()
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/CreateFailedWithUnclaimFailed",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt: authorizedCallOpt,
			ClaimFunc: func(context.Context, types.EUI64, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(context.Context, *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, errCreate.New()
			},
			UnclaimFunc: func(context.Context, types.EUI64) error {
				return errUnclaim.New()
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/SuccessfullyClaimedAndCreated",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "test-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			ClaimFunc: func(context.Context, types.EUI64, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			CallOpt: authorizedCallOpt,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.ClaimFunc != nil {
				mockGatewayClaimer.ClaimFunc = tc.ClaimFunc
			}
			if tc.UnclaimFunc != nil {
				mockGatewayClaimer.UnclaimFunc = tc.UnclaimFunc
			}
			if tc.CreateFunc != nil {
				mockGatewayRegistry.createFunc = tc.CreateFunc
			}

			_, err := gclsClient.Claim(ctx, tc.Req, tc.CallOpt)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			}
		})
	}

	// Test unclaiming.
	for _, tc := range []struct { //nolint:paralleltest
		Name           string
		Req            *ttnpb.GatewayIdentifiers
		CallOpt        grpc.CallOption
		GetFunc        func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
		UnclaimFunc    func(context.Context, types.EUI64) error
		ErrorAssertion func(error) bool
	}{
		{
			Name:           "Unclaim/EmptyRequest",
			Req:            &ttnpb.GatewayIdentifiers{},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Unclaim/NoGatewayRights",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "test-gateway",
			},
			CallOpt:        unAuthorizedCallOpt,
			ErrorAssertion: errors.IsUnauthenticated,
		},
		{
			Name: "Unclaim/InvalidGatewayID",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "test-gateway*W(&$@#)",
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Unclaim/NoGatewayEUI",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "test-gateway",
			},
			GetFunc: func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
				return &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{
						GatewayId: "test-gateway",
					},
				}, nil
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Unclaim/EUINotRegisteredForClaiming",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "unsupported-eui",
			},
			GetFunc: func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
				return &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{
						GatewayId: "test-gateway",
						Eui:       unsupportedEUI.Bytes(),
					},
					GatewayServerAddress: "test.example.com",
				}, nil
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Unclaim/Failed",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "test-gateway",
			},
			GetFunc: func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
				return &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{
						GatewayId: "test-gateway",
						Eui:       supportedEUI.Bytes(),
					},
					GatewayServerAddress: "test.example.com",
				}, nil
			},
			UnclaimFunc: func(context.Context, types.EUI64) error {
				return errUnclaim.New()
			},
			CallOpt:        authorizedCallOpt,
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Unclaim/Success",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "test-gateway",
			},
			GetFunc: func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
				return &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{
						GatewayId: "test-gateway",
						Eui:       supportedEUI.Bytes(),
					},
					GatewayServerAddress: "test.example.com",
				}, nil
			},
			UnclaimFunc: func(context.Context, types.EUI64) error {
				return nil
			},
			CallOpt: authorizedCallOpt,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			if tc.UnclaimFunc != nil {
				mockGatewayClaimer.UnclaimFunc = tc.UnclaimFunc
			}
			if tc.GetFunc != nil {
				mockGatewayRegistry.getFunc = tc.GetFunc
			}
			_, err := gclsClient.Unclaim(ctx, tc.Req, tc.CallOpt)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			}
		})
	}
}
