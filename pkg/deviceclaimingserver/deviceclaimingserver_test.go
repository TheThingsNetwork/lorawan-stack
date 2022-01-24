// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	. "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver"
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
	timeout = (1 << 5) * test.Delay
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

func TestEndDeviceClaimingServer(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
	test.Must(New(c, &Config{}))
	componenttest.StartComponent(t, c)
	defer c.Close()

	// Wait for server to be ready.
	time.Sleep(timeout)

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)
	edcsClient := ttnpb.NewEndDeviceClaimingServerClient(c.LoopbackConn())

	ids := &ttnpb.ApplicationIdentifiers{
		ApplicationId: "foo",
	}

	authorizedCallOpt := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:  "Bearer",
		AuthValue: "foo",
	})

	// Test API Validation here. Functionality is tested in the implementations.
	for _, tc := range []struct {
		Name           string
		Req            interface{}
		ErrorAssertion func(err error) bool
	}{
		{
			Name: "Authorize/NilIDs",
			Req: &ttnpb.AuthorizeApplicationRequest{
				ApplicationIds: nil,
				ApiKey:         "test",
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Authorize/EmptyAPIKey",
			Req: &ttnpb.AuthorizeApplicationRequest{
				ApplicationIds: ids,
				ApiKey:         "",
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:           "Unauthorize/EmptyAppIDs",
			Req:            &ttnpb.ApplicationIdentifiers{},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:           "Claim/EmptyRequest",
			Req:            &ttnpb.ClaimEndDeviceRequest{},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/NilTargetApplicationIds",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QrCode{
					QrCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIds: nil,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/NilSource",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: nil,
				TargetApplicationIds: &ttnpb.ApplicationIdentifiers{
					ApplicationId: "target-app",
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/NoEUIs",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{},
				},
				TargetApplicationIds: &ttnpb.ApplicationIdentifiers{
					ApplicationId: "target-app",
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			var err error
			switch req := tc.Req.(type) {
			case *ttnpb.AuthorizeApplicationRequest:
				_, err = edcsClient.AuthorizeApplication(ctx, req, authorizedCallOpt)
			case *ttnpb.ApplicationIdentifiers:
				_, err = edcsClient.UnauthorizeApplication(ctx, req, authorizedCallOpt)
			case *ttnpb.ClaimEndDeviceRequest:
				_, err = edcsClient.Claim(ctx, req, authorizedCallOpt)
			default:
				panic("invalid request type")
			}
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

var (
	claimAuthCode = []byte("test-code")
	userID        = ttnpb.UserIdentifiers{
		UserId: "test-user",
	}
	authorizedCallOpt = grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:  "Bearer",
		AuthValue: "foo",
	})
)

func TestGatewayClaimingServer(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
	test.Must(New(c, &Config{}))
	componenttest.StartComponent(t, c)
	defer c.Close()

	// Wait for server to be ready.
	time.Sleep(timeout)

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)
	gclsClient := ttnpb.NewGatewayClaimingServerClient(c.LoopbackConn())

	// Test API Validation here. Functionality is tested in the implementations.
	for _, tc := range []struct {
		Name           string
		Req            interface{}
		ErrorAssertion func(err error) bool
	}{
		{
			Name: "Authorize/NilIDs",
			Req: &ttnpb.AuthorizeGatewayRequest{
				GatewayIds: nil,
				ApiKey:     "test",
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name:           "Unauthorize/EmptyIDs",
			Req:            &ttnpb.GatewayIdentifiers{},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/EmptyRequest",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: userID.OrganizationOrUserIdentifiers(),
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/NilCollaborator",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: nil,
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         types.EUI64{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x20},
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "my-new-gateway",
				TargetGatewayServerAddress: "target-tenant.things.example.com",
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "Claim/NilCollaborator",
			Req: &ttnpb.ClaimGatewayRequest{
				Collaborator: nil,
				SourceGateway: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimGatewayRequest_AuthenticatedIdentifiers{
						GatewayEui:         types.EUI64{0x58, 0xA0, 0xCB, 0xFF, 0xFE, 0x80, 0x00, 0x20},
						AuthenticationCode: claimAuthCode,
					},
				},
				TargetGatewayId:            "my-new-gateway",
				TargetGatewayServerAddress: "target-tenant.things.example.com",
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			var err error
			switch req := tc.Req.(type) {
			case *ttnpb.AuthorizeGatewayRequest:
				_, err = gclsClient.AuthorizeGateway(ctx, req, authorizedCallOpt)
			case *ttnpb.GatewayIdentifiers:
				_, err = gclsClient.UnauthorizeGateway(ctx, req, authorizedCallOpt)
			case *ttnpb.ClaimGatewayRequest:
				_, err = gclsClient.Claim(ctx, req, authorizedCallOpt)
			default:
				panic("invalid request type")
			}
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
