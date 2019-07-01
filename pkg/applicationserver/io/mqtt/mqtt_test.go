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

package mqtt_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	. "go.thethings.network/lorawan-stack/pkg/applicationserver/io/mqtt"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredApplicationID  = ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}
	registeredApplicationUID = unique.ID(test.Context(), registeredApplicationID)
	registeredApplicationKey = "test-key"
	registeredDeviceID       = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		DeviceID:               "test-device",
	}

	timeout = 10 * test.Delay
)

func TestAuthentication(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := startMockIS(ctx)
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	})
	test.Must(nil, c.Start())
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	as := mock.NewServer()
	lis, err := net.Listen("tcp", ":0")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	Start(c.FillContext(ctx), as, lis, JSON, "tcp")

	for _, tc := range []struct {
		UID string
		Key string
		OK  bool
	}{
		{
			UID: registeredApplicationUID,
			Key: registeredApplicationKey,
			OK:  true,
		},
		{
			UID: registeredApplicationUID,
			Key: "invalid-key",
			OK:  false,
		},
		{
			UID: "invalid-application",
			Key: "invalid-key",
			OK:  false,
		},
	} {
		t.Run(fmt.Sprintf("%v:%v", tc.UID, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			clientOpts := mqtt.NewClientOptions()
			clientOpts.AddBroker(fmt.Sprintf("tcp://%v", lis.Addr()))
			clientOpts.SetUsername(tc.UID)
			clientOpts.SetPassword(tc.Key)
			client := mqtt.NewClient(clientOpts)
			token := client.Connect()
			if ok := token.WaitTimeout(timeout); tc.OK {
				if a.So(ok, should.BeTrue) && a.So(token.Error(), should.BeNil) {
					client.Disconnect(uint(timeout / time.Millisecond))
				}
			} else {
				a.So(ok, should.BeTrue)
				a.So(token.Error(), should.NotBeNil)
			}
		})
	}
}

func TestTraffic(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := startMockIS(ctx)
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
			},
		},
	})
	test.Must(nil, c.Start())
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	as := mock.NewServer()
	lis, err := net.Listen("tcp", ":0")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	Start(c.FillContext(ctx), as, lis, JSON, "tcp")

	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker(fmt.Sprintf("tcp://%v", lis.Addr()))
	clientOpts.SetUsername(registeredApplicationUID)
	clientOpts.SetPassword(registeredApplicationKey)
	client := mqtt.NewClient(clientOpts)
	client.Connect()

	var sub *io.Subscription
	select {
	case sub = <-as.Subscriptions():
	case <-time.After(timeout):
		t.Fatal("Connection timeout")
	}
	defer client.Disconnect(100)

	t.Run("Upstream", func(t *testing.T) {
		for _, tc := range []struct {
			Topic   string
			Message *ttnpb.ApplicationUp
			OK      bool
		}{
			{
				Topic: "#",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{FRMPayload: []byte{0x1, 0x1, 0x1}},
					},
				},
				OK: true,
			},
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/up", unique.ID(ctx, registeredDeviceID.ApplicationIdentifiers), registeredDeviceID.DeviceID),
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{FRMPayload: []byte{0x2, 0x2, 0x2}},
					},
				},
				OK: true,
			},
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/join", unique.ID(ctx, registeredDeviceID.ApplicationIdentifiers), registeredDeviceID.DeviceID),
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{FRMPayload: []byte{0x3, 0x3, 0x3}},
					},
				},
				OK: false, // Invalid topic
			},
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/up", "invalid-application", "invalid-device"),
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{FRMPayload: []byte{0x4, 0x4, 0x4}},
					},
				},
				OK: false, // Invalid application ID
			},
		} {
			t.Run(tc.Topic, func(t *testing.T) {
				a := assertions.New(t)

				upCh := make(chan *ttnpb.ApplicationUp)
				handler := func(_ mqtt.Client, msg mqtt.Message) {
					up := &ttnpb.ApplicationUp{}
					err := jsonpb.TTN().Unmarshal(msg.Payload(), up)
					a.So(err, should.BeNil)
					upCh <- up
				}
				if token := client.Subscribe(tc.Topic, 1, handler); !a.So(token.WaitTimeout(timeout), should.BeTrue) {
					t.FailNow()
				}
				defer func() {
					token := client.Unsubscribe(tc.Topic)
					if !a.So(token.WaitTimeout(timeout), should.BeTrue) {
						t.FailNow()
					}
				}()

				err := sub.SendUp(ctx, tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				select {
				case up := <-upCh:
					if tc.OK {
						a.So(up, should.Resemble, tc.Message)
					} else {
						t.Fatalf("Expected no upstream message but have %v", up)
					}
				case <-time.After(timeout):
					if tc.OK {
						t.Fatal("Receive expected upstream timeout")
					}
				}
			})
		}
	})

	t.Run("Downstream", func(t *testing.T) {
		for _, tc := range []struct {
			Topic    string
			IDs      ttnpb.EndDeviceIdentifiers
			Message  *ttnpb.ApplicationDownlinks
			Expected []*ttnpb.ApplicationDownlink
		}{
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/down/push", unique.ID(ctx, registeredDeviceID.ApplicationIdentifiers), registeredDeviceID.DeviceID),
				IDs:   registeredDeviceID,
				Message: &ttnpb.ApplicationDownlinks{
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x1, 0x1, 0x1},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x1, 0x1, 0x1},
					},
				},
			},
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/down/replace", unique.ID(ctx, registeredDeviceID.ApplicationIdentifiers), registeredDeviceID.DeviceID),
				IDs:   registeredDeviceID,
				Message: &ttnpb.ApplicationDownlinks{
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x2, 0x2, 0x2},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2},
					},
				},
			},
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/down/push", unique.ID(ctx, registeredDeviceID.ApplicationIdentifiers), "invalid-device"),
				IDs:   registeredDeviceID,
				Message: &ttnpb.ApplicationDownlinks{
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x3, 0x3, 0x3},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2}, // Do not expect a change.
					},
				},
			},
			{
				Topic: fmt.Sprintf("v3/%v/devices/%v/down/push", "invalid-application", "invalid-device"),
				IDs:   registeredDeviceID,
				Message: &ttnpb.ApplicationDownlinks{
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x4, 0x4, 0x4},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2}, // Do not expect a change.
					},
				},
			},
		} {
			tcok := t.Run(tc.Topic, func(t *testing.T) {
				a := assertions.New(t)
				buf, err := jsonpb.TTN().Marshal(tc.Message)
				a.So(err, should.BeNil)
				if token := client.Publish(tc.Topic, 1, false, buf); !a.So(token.WaitTimeout(timeout), should.BeTrue) {
					t.FailNow()
				}
				res, err := as.DownlinkQueueList(ctx, tc.IDs)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Expected)
			})
			if !tcok {
				t.FailNow()
			}
		}
	})
}
