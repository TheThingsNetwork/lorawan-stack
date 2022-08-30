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

package grpc_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	. "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/grpc"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors"
	"go.thethings.network/lorawan-stack/v3/pkg/messageprocessors/cayennelpp"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	registeredApplicationID  = &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}
	registeredApplicationUID = unique.ID(test.Context(), registeredApplicationID)
	registeredApplicationKey = "test-key"

	timeout = (1 << 6) * test.Delay
)

func TestAuthentication(t *testing.T) {
	t.Parallel()
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	srv := New(as)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewAppAsClient(c.LoopbackConn())

	//nolint:paralleltest
	for _, tc := range []struct {
		ID  *ttnpb.ApplicationIdentifiers
		Key string
		OK  bool
	}{
		{
			ID:  registeredApplicationID,
			Key: registeredApplicationKey,
			OK:  true,
		},
		{
			ID:  registeredApplicationID,
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  &ttnpb.ApplicationIdentifiers{ApplicationId: "invalid-application"},
			Key: "invalid-key",
			OK:  false,
		},
	} {
		t.Run(fmt.Sprintf("%v:%v", tc.ID.ApplicationId, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			creds := grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     tc.Key,
				AllowInsecure: true,
			})

			wg := sync.WaitGroup{}
			wg.Add(1)
			var err error
			go func() {
				defer wg.Done()
				_, err = client.Subscribe(ctx, tc.ID, creds)
			}()
			wg.Wait()

			if tc.OK && err != nil && !a.So(errors.IsCanceled(err), should.BeTrue) {
				t.Fatalf("Unexpected link error: %v", err)
			}
			if !tc.OK && !a.So(errors.IsCanceled(err), should.BeFalse) {
				t.FailNow()
			}
		})
	}
}

type erroredApplicationUp struct {
	*ttnpb.ApplicationUp
	error
}

func TestTraffic(t *testing.T) {
	t.Parallel()
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	srv := New(as)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewAppAsClient(c.LoopbackConn())

	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})
	badCreds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     "barfoo",
		AllowInsecure: true,
	})

	upCh := make(chan erroredApplicationUp, 10)
	stream, err := client.Subscribe(ctx, registeredApplicationID, creds)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	go func() {
		for ctx.Err() == nil {
			up, err := stream.Recv()
			upCh <- erroredApplicationUp{up, err}
		}
	}()

	var sub *io.Subscription
	select {
	case sub = <-as.Subscriptions():
	case <-time.After(timeout):
		t.Fatal("Subscription timeout")
	}

	//nolint:paralleltest
	t.Run("Upstream", func(t *testing.T) {
		a := assertions.New(t)

		up := &ttnpb.ApplicationUp{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: registeredApplicationID,
				DeviceId:       "foo-device",
			},
			Up: &ttnpb.ApplicationUp_UplinkMessage{
				UplinkMessage: &ttnpb.ApplicationUplink{
					FrmPayload: []byte{0x01, 0x02, 0x03},
				},
			},
		}
		if err := sub.Publish(ctx, up); !a.So(err, should.BeNil) {
			t.FailNow()
		}

		select {
		case actual := <-upCh:
			a.So(actual.ApplicationUp, should.Resemble, up)
			a.So(actual.error, should.BeNil)
		case <-time.After(timeout):
			t.Fatal("Receive expected upstream message timeout")
		}
	})

	//nolint:paralleltest
	t.Run("Downstream", func(t *testing.T) {
		a := assertions.New(t)
		ids := ttnpb.EndDeviceIdentifiers{
			ApplicationIds: registeredApplicationID,
			DeviceId:       "foo-device",
		}

		// List: unauthorized.
		{
			_, err := client.DownlinkQueueList(ctx, &ids, badCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// List: happy flow; no items.
		{
			res, err := client.DownlinkQueueList(ctx, &ids, creds)
			a.So(err, should.BeNil)
			a.So(res.Downlinks, should.HaveLength, 0)
		}

		// Push: unauthorized.
		{
			_, err := client.DownlinkQueuePush(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: &ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      1,
						FrmPayload: []byte{0x01, 0x01, 0x01},
					},
				},
			}, badCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Push and assert content: happy flow.
		{
			_, err := client.DownlinkQueuePush(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: &ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						SessionKeyId:   []byte{0x11, 0x22, 0x33, 0x44},
						FPort:          1,
						FrmPayload:     []byte{0x01, 0x01, 0x01},
						Confirmed:      true,
						CorrelationIds: []string{"test"},
					},
					{
						FPort:      2,
						FrmPayload: []byte{0x02, 0x02, 0x02},
					},
				},
			}, creds)
			a.So(err, should.BeNil)
		}
		{
			_, err := client.DownlinkQueuePush(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: &ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      3,
						FrmPayload: []byte{0x03, 0x03, 0x03},
					},
				},
			}, creds)
			a.So(err, should.BeNil)
		}
		{
			res, err := client.DownlinkQueueList(ctx, &ids, creds)
			a.So(err, should.BeNil)
			a.So(res.Downlinks, should.HaveLength, 3)
			a.So(res.Downlinks, should.Resemble, []*ttnpb.ApplicationDownlink{
				{
					SessionKeyId:   []byte{0x11, 0x22, 0x33, 0x44},
					FPort:          1,
					Confirmed:      true,
					FrmPayload:     []byte{0x01, 0x01, 0x01},
					CorrelationIds: []string{"test"},
				},
				{
					FPort:      2,
					FrmPayload: []byte{0x02, 0x02, 0x02},
				},
				{
					FPort:      3,
					FrmPayload: []byte{0x03, 0x03, 0x03},
				},
			})
		}

		// Replace: unauthorized.
		{
			_, err := client.DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: &ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      4,
						FrmPayload: []byte{0x04, 0x04, 0x04},
					},
				},
			}, badCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Replace and assert content: happy flow.
		{
			_, err := client.DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIds: &ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      4,
						FrmPayload: []byte{0x04, 0x04, 0x04},
						Confirmed:  true,
					},
				},
			}, creds)
			a.So(err, should.BeNil)
		}
		{
			res, err := client.DownlinkQueueList(ctx, &ids, creds)
			a.So(err, should.BeNil)
			a.So(res.Downlinks, should.HaveLength, 1)
			a.So(res.Downlinks, should.Resemble, []*ttnpb.ApplicationDownlink{
				{
					FPort:      4,
					FrmPayload: []byte{0x04, 0x04, 0x04},
					Confirmed:  true,
				},
			})
		}
	})
}

type mockMQTTConfigProvider struct {
	config.MQTT
}

func (p mockMQTTConfigProvider) GetMQTTConfig(context.Context) (*config.MQTT, error) {
	return &p.MQTT, nil
}

func TestMQTTConfig(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	srv := New(as, WithMQTTConfigProvider(&mockMQTTConfigProvider{
		MQTT: config.MQTT{
			PublicAddress:    "example.com:1883",
			PublicTLSAddress: "example.com:8883",
		},
	}))
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewAppAsClient(c.LoopbackConn())

	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})

	info, err := client.GetMQTTConnectionInfo(ctx, registeredApplicationID, creds)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(info, should.Resemble, &ttnpb.MQTTConnectionInfo{
		Username:         registeredApplicationUID,
		PublicAddress:    "example.com:1883",
		PublicTlsAddress: "example.com:8883",
	})
}

func TestSimulateUplink(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})

	registeredDeviceID := &ttnpb.EndDeviceIdentifiers{
		DeviceId:       "dev1",
		ApplicationIds: registeredApplicationID,
		DevEui:         types.EUI64{0x00, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x01}.Bytes(),
	}

	as := mock.NewServer(c)
	f := &mockFetcher{}
	srv := New(as, WithGetEndDeviceIdentifiers(f.Get))
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewAppAsClient(c.LoopbackConn())
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})

	upCh := make(chan erroredApplicationUp, 1)
	streamCtx := test.Context()
	stream, err := client.Subscribe(streamCtx, registeredApplicationID, creds)
	if err != nil {
		t.Fatalf("Failed to subscribe: %s\n", err)
	}
	go func() {
		for streamCtx.Done() == nil {
			up, err := stream.Recv()
			upCh <- erroredApplicationUp{up, err}
		}
	}()

	<-as.Subscriptions()
	//nolint:paralleltest
	for _, tc := range []struct {
		name              string
		up                *ttnpb.ApplicationUp
		setup             func(f *mockFetcher)
		expectIdentifiers *ttnpb.EndDeviceIdentifiers
	}{
		{
			name: "Fetch",
			up: &ttnpb.ApplicationUp{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       registeredDeviceID.DeviceId,
					ApplicationIds: registeredApplicationID,
				},
				Up: &ttnpb.ApplicationUp_ServiceData{
					ServiceData: &ttnpb.ApplicationServiceData{},
				},
			},
			setup: func(f *mockFetcher) {
				f.ids = registeredDeviceID
				f.err = nil
			},
			expectIdentifiers: registeredDeviceID,
		},
		{
			name: "FetchError",
			up: &ttnpb.ApplicationUp{
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId:       registeredDeviceID.DeviceId,
					ApplicationIds: registeredApplicationID,
				},
				Up: &ttnpb.ApplicationUp_ServiceData{
					ServiceData: &ttnpb.ApplicationServiceData{},
				},
			},
			setup: func(f *mockFetcher) {
				f.ids = nil
				f.err = fmt.Errorf("mock error")
			},
			expectIdentifiers: &ttnpb.EndDeviceIdentifiers{
				DeviceId:       registeredDeviceID.DeviceId,
				ApplicationIds: registeredApplicationID,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(f)

			_, err = client.SimulateUplink(ctx, tc.up, creds)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			select {
			case up := <-upCh:
				if err := up.error; err != nil {
					t.Fatalf("Received unexpected error: %s\n", err)
				}
				a.So(f.calledWithIdentifiers, should.Resemble, tc.up.EndDeviceIds)
				a.So(up.EndDeviceIds, should.Resemble, tc.expectIdentifiers)
			case <-time.After(timeout):
				t.Fatal("Timed out waiting for simulated uplink")
			}
		})
	}
}

func TestMessageProcessors(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
		},
	})
	as := mock.NewServer(c)
	srv := New(as, WithPayloadProcessor(&messageprocessors.MapPayloadProcessor{
		ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP: cayennelpp.New(),
	}))
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewAppAsClient(c.LoopbackConn())

	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})

	{
		resp, err := client.EncodeDownlink(ctx, &ttnpb.EncodeDownlinkRequest{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: registeredApplicationID,
				DeviceId:       "foobar",
			},
			Downlink: &ttnpb.ApplicationDownlink{
				DecodedPayload: &pbtypes.Struct{
					Fields: map[string]*pbtypes.Value{
						"value_2": {
							Kind: &pbtypes.Value_NumberValue{
								NumberValue: -50.51,
							},
						},
					},
				},
				FPort: 1,
			},
			Formatter: ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
		}, creds)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		if a.So(resp.Downlink, should.NotBeNil) {
			a.So(resp.Downlink.FPort, should.Equal, 1)
			a.So(resp.Downlink.FrmPayload, should.Resemble, []byte{2, 236, 69})
		}
	}

	{
		resp, err := client.DecodeUplink(ctx, &ttnpb.DecodeUplinkRequest{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: registeredApplicationID,
				DeviceId:       "foobar",
			},
			Uplink: &ttnpb.ApplicationUplink{
				FrmPayload: []byte{1, 0, 255},
				RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}},
					},
				},
				FPort: 1,
			},
			Formatter: ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
		}, creds)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		if a.So(resp.Uplink, should.NotBeNil) {
			a.So(resp.Uplink.FPort, should.Equal, 1)
			a.So(resp.Uplink.DecodedPayload, should.Resemble, &pbtypes.Struct{
				Fields: map[string]*pbtypes.Value{
					"digital_in_1": {
						Kind: &pbtypes.Value_NumberValue{NumberValue: 255},
					},
				},
			})
		}
	}

	{
		resp, err := client.DecodeDownlink(ctx, &ttnpb.DecodeDownlinkRequest{
			EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: registeredApplicationID,
				DeviceId:       "foobar",
			},
			Downlink: &ttnpb.ApplicationDownlink{
				FrmPayload: []byte{2, 236, 69},
				FPort:      1,
			},
			Formatter: ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP,
		}, creds)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		if a.So(resp.Downlink, should.NotBeNil) {
			a.So(resp.Downlink.FPort, should.Equal, 1)
			a.So(resp.Downlink.DecodedPayload, should.Resemble, &pbtypes.Struct{
				Fields: map[string]*pbtypes.Value{
					"value_2": {
						Kind: &pbtypes.Value_NumberValue{
							NumberValue: -50.51,
						},
					},
				},
			})
		}
	}
}
