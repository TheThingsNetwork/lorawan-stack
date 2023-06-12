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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	. "go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver"
	"go.thethings.network/lorawan-stack/v3/pkg/deviceclaimingserver/enddevices"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var timeout = (1 << 5) * test.Delay

func mustHavePeer(ctx context.Context, c *component.Component, role ttnpb.ClusterRole) {
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	panic("could not connect to peer")
}

var (
	registeredApplicationIDs = &ttnpb.ApplicationIdentifiers{
		ApplicationId: "test-application",
	}
	registeredApplicationKey = "test-key"
	registeredEndDeviceID    = "test-end-device"
	registeredJoinEUI        = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0C}
	unRegisteredJoinEUI      = types.EUI64{0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D}
	registeredDevEUI         = types.EUI64{0x00, 0x04, 0xA3, 0x0B, 0x00, 0x1C, 0x05, 0x30}
)

func TestEndDeviceClaimingServer(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	t.Cleanup(func() {
		cancelCtx()
	})

	is, isAddr, closeIS := mockis.New(ctx)
	t.Cleanup(closeIS)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
	mockUpstream, err := enddevices.NewUpstream(
		ctx,
		c,
		enddevices.Config{},
		enddevices.WithClaimer("test", &MockClaimer{
			JoinEUI: registeredJoinEUI,
		}),
	)
	a.So(err, should.BeNil)
	dcs, err := New(c, &Config{}, WithEndDeviceClaimingUpstream(mockUpstream))
	test.Must(dcs, err)

	componenttest.StartComponent(t, c)
	t.Cleanup(func() {
		c.Close()
	})

	// Wait for server to be ready.
	time.Sleep(timeout)

	mustHavePeer(ctx, c, ttnpb.ClusterRole_DEVICE_CLAIMING_SERVER)
	edcsClient := ttnpb.NewEndDeviceClaimingServerClient(c.LoopbackConn())

	authorizedCallOpt = grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:  "Bearer",
		AuthValue: registeredApplicationKey,
	})

	// Register entities.
	is.ApplicationRegistry().Add(
		ctx,
		registeredApplicationIDs,
		registeredApplicationKey,
		ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	)
	is.EndDeviceRegistry().Add(
		ctx,
		&ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: registeredApplicationIDs,
				DeviceId:       registeredEndDeviceID,
				JoinEui:        registeredJoinEUI.Bytes(),
				DevEui:         registeredDevEUI.Bytes(),
			},
		},
	)

	// GetInfoByJoinEUI.
	resp, err := edcsClient.GetInfoByJoinEUI(ctx, &ttnpb.GetInfoByJoinEUIRequest{
		JoinEui: registeredJoinEUI.Bytes(),
	})
	a.So(err, should.BeNil)
	a.So(resp, should.NotBeNil)
	a.So(resp.JoinEui, should.Resemble, registeredJoinEUI.Bytes())
	a.So(resp.SupportsClaiming, should.BeTrue)

	resp, err = edcsClient.GetInfoByJoinEUI(ctx, &ttnpb.GetInfoByJoinEUIRequest{
		JoinEui: unRegisteredJoinEUI.Bytes(),
	})
	a.So(err, should.BeNil)
	a.So(resp, should.NotBeNil)
	a.So(resp.JoinEui, should.Resemble, unRegisteredJoinEUI.Bytes())
	a.So(resp.SupportsClaiming, should.BeFalse)

	// Claim end device.
	for _, tc := range []struct {
		Name           string
		Req            *ttnpb.ClaimEndDeviceRequest
		ErrorAssertion func(err error) bool
	}{
		{
			Name:           "EmptyRequest",
			Req:            &ttnpb.ClaimEndDeviceRequest{},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "NilTargetApplicationIds",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_QrCode{
					QrCode: []byte("URN:LW:DP:42FFFFFFFFFFFFFF:4242FFFFFFFFFFFF:42FFFF42:%V0102"),
				},
				TargetApplicationIds: nil,
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "NilSource",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: nil,
				TargetApplicationIds: &ttnpb.ApplicationIdentifiers{
					ApplicationId: "target-app",
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "NoEUIs",
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
		{
			Name: "ValidDevice",
			Req: &ttnpb.ClaimEndDeviceRequest{
				SourceDevice: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers_{
					AuthenticatedIdentifiers: &ttnpb.ClaimEndDeviceRequest_AuthenticatedIdentifiers{
						JoinEui:            registeredJoinEUI.Bytes(),
						DevEui:             registeredDevEUI.Bytes(),
						AuthenticationCode: "TEST1234",
					},
				},
				TargetApplicationIds: registeredApplicationIDs,
				TargetDeviceId:       "target-device",
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			_, err := edcsClient.Claim(ctx, tc.Req, authorizedCallOpt)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			}
		})
	}

	// GetClaimStatus.
	status, err := edcsClient.GetClaimStatus(ctx, &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: registeredApplicationIDs,
		DeviceId:       registeredEndDeviceID,
	}, authorizedCallOpt)
	a.So(err, should.BeNil)
	a.So(status, should.NotBeNil)

	// Unclaim.
	_, err = edcsClient.Unclaim(ctx, &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: registeredApplicationIDs,
		DeviceId:       registeredEndDeviceID,
	}, authorizedCallOpt)
	a.So(err, should.BeNil)
}
