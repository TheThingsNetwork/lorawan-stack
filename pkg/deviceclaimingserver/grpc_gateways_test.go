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
	"google.golang.org/protobuf/types/known/emptypb"
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

func TestGatewayClaimingServer(t *testing.T) { //nolint:paralleltest,gocyclo
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
			[]types.EUI64Range{
				types.EUI64Prefix{
					EUI64:  types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0x00},
					Length: 48,
				}.EUI64Range(),
			},
			mockGatewayClaimer,
		),
	)
	if err != nil {
		t.FailNow()
	}

	existingEUI := types.EUI64{0x58, 0xa0, 0xcb, 0xff, 0xfe, 0x80, 0x00, 0xFF}
	// Default Get behavior: the created gateway is immediately visible.
	getCreatedGatewayFunc := func(_ context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
		return &ttnpb.Gateway{Ids: req.GatewayIds}, nil
	}
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
	getGatewayCalls := 0
	for _, tc := range []struct {
		Name           string
		Req            *ttnpb.ClaimGatewayRequest
		CallOpt        grpc.CallOption
		ClaimFunc      func(context.Context, *ttnpb.GatewayIdentifiers, string, string) (*dcstypes.GatewayMetadata, error)
		CreateFunc     func(context.Context, *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error)
		GetFunc        func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
		UpdateFunc     func(context.Context, *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error)
		UnclaimFunc    func(context.Context, *ttnpb.GatewayIdentifiers) error
		PurgeFunc      func(context.Context, *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error)
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
			Name: "Claim/GatewayCreationFailed",
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
			CreateFunc: func(_ context.Context, _ *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, errCreate.New()
			},
			ErrorAssertion: errors.IsAborted,
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
			CallOpt: authorizedCallOpt,
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			PurgeFunc: func(_ context.Context, _ *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
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
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			ClaimFunc: func(_ context.Context, _ *ttnpb.GatewayIdentifiers, _, _ string) (*dcstypes.GatewayMetadata, error) {
				return nil, errClaim.New()
			},
			UpdateFunc: func(_ context.Context, in *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			PurgeFunc: func(_ context.Context, _ *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/UpdateFailed",
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
			ClaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(context.Context, *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, nil //nolint:nilnil
			},
			UpdateFunc: func(_ context.Context, _ *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, errUpdate.New()
			},
			PurgeFunc: func(_ context.Context, _ *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			UnclaimFunc: func(_ context.Context, ids *ttnpb.GatewayIdentifiers) error {
				if types.MustEUI64(ids.Eui).OrZero().Equal(supportedEUI) {
					return nil
				}
				return errUnclaim.New()
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/UpdateFailedWithUnclaimFailed",
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
			ClaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(context.Context, *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, nil //nolint:nilnil
			},
			UpdateFunc: func(_ context.Context, _ *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
				return nil, errUpdate.New()
			},
			PurgeFunc: func(_ context.Context, _ *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
				return &emptypb.Empty{}, nil
			},
			UnclaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers) error {
				return errUnclaim.New()
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/SuccessfullyClaimedAndUpdated",
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
			ClaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			UpdateFunc: func(_ context.Context, in *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			CallOpt: authorizedCallOpt,
		},
		{
			Name: "Claim/CreatedGatewayVisibleAfterReplicationLag",
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
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				getGatewayCalls = 0
				return in.Gateway, nil
			},
			GetFunc: func(_ context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
				// The created gateway becomes visible on the third attempt.
				if getGatewayCalls++; getGatewayCalls < 3 {
					return nil, errGatewayNotFound.New()
				}
				return &ttnpb.Gateway{Ids: req.GatewayIds}, nil
			},
			ClaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers, string, string) (*dcstypes.GatewayMetadata, error) {
				return &dcstypes.GatewayMetadata{}, nil
			},
			UpdateFunc: func(_ context.Context, in *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
		},
		{
			Name: "Claim/EmptyTargetGatewayIDDefaultsToEUIAndDeletesOnFailedClaim",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayServerAddress: "things.example.com",
			},
			CallOpt: authorizedCallOpt,
			ClaimFunc: func(
				_ context.Context, ids *ttnpb.GatewayIdentifiers, _, _ string,
			) (*dcstypes.GatewayMetadata, error) {
				a.So(ids.GatewayId, should.Equal, "58a0cbfffe800001")
				a.So(ids.Eui, should.Resemble, supportedEUI.Bytes())
				return nil, errClaim.New()
			},
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				a.So(in.Gateway.GetIds().GetGatewayId(), should.Equal, "58a0cbfffe800001")
				return in.Gateway, nil
			},
			PurgeFunc: func(_ context.Context, ids *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
				a.So(ids.GatewayId, should.Equal, "58a0cbfffe800001")
				a.So(ids.Eui, should.Resemble, supportedEUI.Bytes())
				return &emptypb.Empty{}, nil
			},
			ErrorAssertion: errors.IsAborted,
		},
		{
			Name: "Claim/ForwardsGatewayIdentifiers",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.GetOrganizationOrUserIdentifiers(),
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         supportedEUI.Bytes(),
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "forwarded-gateway",
				TargetGatewayServerAddress: "things.example.com",
			},
			ClaimFunc: func(
				_ context.Context, ids *ttnpb.GatewayIdentifiers, _, _ string,
			) (*dcstypes.GatewayMetadata, error) {
				a.So(ids.GatewayId, should.Equal, "forwarded-gateway")
				a.So(ids.Eui, should.Resemble, supportedEUI.Bytes())
				return &dcstypes.GatewayMetadata{}, nil
			},
			CreateFunc: func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			UpdateFunc: func(_ context.Context, in *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
				return in.Gateway, nil
			},
			CallOpt: authorizedCallOpt,
		},
	} {
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
			if tc.GetFunc != nil {
				mockGatewayRegistry.getFunc = tc.GetFunc
			} else {
				mockGatewayRegistry.getFunc = getCreatedGatewayFunc
			}
			if tc.UpdateFunc != nil {
				mockGatewayRegistry.updateFunc = tc.UpdateFunc
			}
			if tc.PurgeFunc != nil {
				mockGatewayRegistry.purgeFunc = tc.PurgeFunc
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

	t.Run("Claim/CreatedGatewayNotVisible", func(t *testing.T) { //nolint:paralleltest
		a := assertions.New(t)
		var deleted bool
		mockGatewayRegistry.createFunc = func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
			return in.Gateway, nil
		}
		// Rights on a not yet replicated gateway are masked as a permission-denied error for non-admin callers.
		mockGatewayRegistry.getFunc = func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
			return nil, errNoRights.New()
		}
		mockGatewayRegistry.purgeFunc = func(context.Context, *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
			deleted = true
			return &emptypb.Empty{}, nil
		}
		_, err := gclsClient.Claim(ctx, &ttnpb.ClaimGatewayRequest{
			Collaborator: userID.GetOrganizationOrUserIdentifiers(),
			SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
				AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
					GatewayEui:         supportedEUI.Bytes(),
					AuthenticationCode: claimAuthCode,
				},
			},
			TargetGatewayId:            "test-gateway",
			TargetGatewayServerAddress: "things.example.com",
		}, authorizedCallOpt)
		a.So(errors.IsDeadlineExceeded(err), should.BeTrue)
		a.So(deleted, should.BeTrue)
	})

	t.Run("Claim/CleanupAfterCanceledRequest", func(t *testing.T) { //nolint:paralleltest
		assertClaimCleanupNotBoundToRequest(
			ctx, t, gclsClient, mockGatewayRegistry, mockGatewayClaimer, supportedEUI,
		)
	})

	// Test unclaiming.
	for _, tc := range []struct { //nolint:paralleltest
		Name           string
		Req            *ttnpb.GatewayIdentifiers
		CallOpt        grpc.CallOption
		GetFunc        func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
		UnclaimFunc    func(context.Context, *ttnpb.GatewayIdentifiers) error
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
			UnclaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers) error {
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
			UnclaimFunc: func(context.Context, *ttnpb.GatewayIdentifiers) error {
				return nil
			},
			CallOpt: authorizedCallOpt,
		},
		{
			Name: "Unclaim/ForwardsGatewayIdentifiers",
			Req: &ttnpb.GatewayIdentifiers{
				GatewayId: "forwarded-gateway",
			},
			GetFunc: func(context.Context, *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
				return &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{
						GatewayId: "forwarded-gateway",
						Eui:       supportedEUI.Bytes(),
					},
					GatewayServerAddress: "test.example.com",
				}, nil
			},
			UnclaimFunc: func(_ context.Context, ids *ttnpb.GatewayIdentifiers) error {
				a.So(ids.GatewayId, should.Equal, "forwarded-gateway")
				a.So(ids.Eui, should.Resemble, supportedEUI.Bytes())
				return nil
			},
			CallOpt: authorizedCallOpt,
		},
	} {
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

// assertClaimCleanupNotBoundToRequest asserts that the cleanup of a partially claimed gateway is not bound to the
// request context, since cancellation of the request is itself one of the failures that the cleanup reverts.
func assertClaimCleanupNotBoundToRequest(
	ctx context.Context,
	t *testing.T,
	gclsClient ttnpb.GatewayClaimingServerClient,
	registry *mockGatewayRegistry,
	claimer *MockGatewayClaimer,
	eui types.EUI64,
) {
	t.Helper()
	a := assertions.New(t)

	callCtx, cancelCall := context.WithCancel(ctx)
	t.Cleanup(cancelCall)

	// Records the error of the context that the delete is called with, at the time of the call.
	deleteCtxErrCh := make(chan error, 1)
	registry.createFunc = func(_ context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
		return in.Gateway, nil
	}
	registry.getFunc = func(_ context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
		return &ttnpb.Gateway{Ids: req.GatewayIds}, nil
	}
	registry.purgeFunc = func(ctx context.Context, _ *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
		deleteCtxErrCh <- ctx.Err()
		return &emptypb.Empty{}, nil
	}
	claimer.ClaimFunc = func(
		reqCtx context.Context, _ *ttnpb.GatewayIdentifiers, _, _ string,
	) (*dcstypes.GatewayMetadata, error) {
		// Cancel the request and wait until the cancellation reaches the handler, so that the deferred cleanup runs
		// with an already canceled request context.
		cancelCall()
		select {
		case <-reqCtx.Done():
		case <-time.After(timeout):
			t.Error("Request context was not canceled")
		}
		return nil, reqCtx.Err()
	}

	_, err := gclsClient.Claim(callCtx, &ttnpb.ClaimGatewayRequest{
		Collaborator: userID.GetOrganizationOrUserIdentifiers(),
		SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
			AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
				GatewayEui:         eui.Bytes(),
				AuthenticationCode: claimAuthCode,
			},
		},
		TargetGatewayId:            "test-gateway",
		TargetGatewayServerAddress: "things.example.com",
	}, authorizedCallOpt)
	a.So(err, should.NotBeNil)

	select {
	case deleteCtxErr := <-deleteCtxErrCh:
		a.So(deleteCtxErr, should.BeNil)
	case <-time.After(timeout):
		t.Fatal("Created gateway was not deleted after the request was canceled")
	}
}
