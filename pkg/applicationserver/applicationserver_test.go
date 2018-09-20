// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package applicationserver_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	// This application will be added to the Entity Registry and to the link registry of the Application Server so that it
	// links automatically on start to the Network Server.
	registeredApplicationID        = ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"}
	registeredApplicationKey       = "secret"
	registeredApplicationFormatter = ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP

	// This device gets registered in the device registry of the Application Server.
	registeredDevice = &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: registeredApplicationID,
			DeviceID:               "foo-device",
			JoinEUI:                eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			DevEUI:                 eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter: ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
			UpFormatterParameter: `function Decoder(payload, f_port) {
	var sum = 0;
	for (i = 0; i < payload.length; i++) {
		sum += payload[i];
	}
	return {
		sum: sum
	};
}`,
		},
	}

	// This device does not get registered in the device registry of the Application Server and will be created on join
	// and on uplink.
	unregisteredDeviceID = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		DeviceID:               "bar-device",
		JoinEUI:                eui64Ptr(types.EUI64{0x24, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		DevEUI:                 eui64Ptr(types.EUI64{0x24, 0x24, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
	}

	timeout = 10 * test.Delay
)

func TestApplicationServer(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	is, isAddr := startMockIS(ctx)
	js, jsAddr := startMockJS(ctx)
	ns, nsAddr := startMockNS(ctx)

	// Register the application in the Entity Registry.
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	// Register some sessions in the Join Server. Sometimes the keys are sent by the Network Server as part of the
	// join-accept, and sometimes they are not sent by the Network Server so the Application Server gets them from the
	// Join Server.
	js.add(ctx, *registeredDevice.DevEUI, "session1", ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}.
		Key:      []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
		KEKLabel: "test",
	})
	js.add(ctx, *registeredDevice.DevEUI, "session2", ttnpb.KeyEnvelope{
		// AppSKey is []byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}.
		Key:      []byte{0xa3, 0x34, 0x38, 0x1c, 0xca, 0x1c, 0x12, 0x7a, 0x5b, 0xb1, 0xa8, 0x97, 0x39, 0xc7, 0x5, 0x34, 0x91, 0x26, 0x9b, 0x21, 0x4f, 0x27, 0x80, 0x19},
		KEKLabel: "test",
	})
	js.add(ctx, *registeredDevice.DevEUI, "session3", ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42}.
		Key:      []byte{0x8c, 0xe9, 0x14, 0x4b, 0x82, 0x23, 0x8, 0x39, 0x65, 0x73, 0xd, 0x42, 0x9f, 0x2a, 0x7c, 0x9c, 0x9c, 0xbe, 0x38, 0xbe, 0x35, 0x5d, 0x44, 0xf},
		KEKLabel: "test",
	})

	deviceRegistry := newMemDeviceRegistry()
	resetDeviceRegistry := func() {
		deviceRegistry.Reset()
		deviceRegistry.Set(ctx, registeredDevice.EndDeviceIdentifiers, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, error) {
			return registeredDevice, nil
		})
	}
	linkRegistry := newMemLinkRegistry()
	linkRegistry.Set(ctx, registeredApplicationID, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, error) {
		return &ttnpb.ApplicationLink{
			DefaultFormatters: &ttnpb.MessagePayloadFormatters{
				UpFormatter:   registeredApplicationFormatter,
				DownFormatter: registeredApplicationFormatter,
			},
		}, nil
	})

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9184",
				AllowInsecureForCredentials: true,
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
				JoinServer:     jsAddr,
				NetworkServer:  nsAddr,
			},
		},
	})
	config := &applicationserver.Config{
		LinkMode: applicationserver.LinkAll,
		Devices:  deviceRegistry,
		Links:    linkRegistry,
		KeyVault: cryptoutil.NewMemKeyVault(map[string][]byte{
			"test": []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
		}),
	}
	as, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := as.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.PeerInfo_APPLICATION_SERVER)

	test.Must(nil, c.Start())
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.PeerInfo_NETWORK_SERVER)
	mustHavePeer(ctx, c, ttnpb.PeerInfo_JOIN_SERVER)
	mustHavePeer(ctx, c, ttnpb.PeerInfo_ENTITY_REGISTRY)

	for _, ptc := range []struct {
		Protocol  string
		ValidAuth func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool
		Connect   func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, upCh chan<- *ttnpb.ApplicationUp, downCh <-chan *ttnpb.ApplicationDownlink) error
	}{
		{
			Protocol: "grpc",
			ValidAuth: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool {
				return ids == registeredApplicationID && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, upCh chan<- *ttnpb.ApplicationUp, downCh <-chan *ttnpb.ApplicationDownlink) error {
				conn, err := grpc.Dial(":9184", grpc.WithInsecure(), grpc.WithBlock())
				if err != nil {
					return err
				}
				defer conn.Close()
				md := rpcmetadata.MD{
					AuthType:      "Key",
					AuthValue:     key,
					AllowInsecure: true,
				}
				client := ttnpb.NewAppAsClient(conn)
				stream, err := client.Subscribe(ctx, &ids, grpc.PerRPCCredentials(md))
				if err != nil {
					return err
				}
				errCh := make(chan error, 1)
				// Read upstream.
				go func() {
					for {
						msg, err := stream.Recv()
						if err != nil {
							errCh <- err
							return
						}
						upCh <- msg
					}
				}()
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		},
	} {
		t.Run(fmt.Sprintf("Authenticate/%v", ptc.Protocol), func(t *testing.T) {
			for _, ctc := range []struct {
				Name string
				ID   ttnpb.ApplicationIdentifiers
				Key  string
			}{
				{
					Name: "ValidIDAndKey",
					ID:   registeredApplicationID,
					Key:  registeredApplicationKey,
				},
				{
					Name: "InvalidKey",
					ID:   registeredApplicationID,
					Key:  "invalid-key",
				},
				{
					Name: "InvalidIDAndKey",
					ID:   ttnpb.ApplicationIdentifiers{ApplicationID: "invalid-gateway"},
					Key:  "invalid-key",
				},
			} {
				t.Run(ctc.Name, func(t *testing.T) {
					ctx, cancel := context.WithDeadline(ctx, time.Now().Add(timeout))
					upCh := make(chan *ttnpb.ApplicationUp)
					downCh := make(chan *ttnpb.ApplicationDownlink)
					err := ptc.Connect(ctx, t, ctc.ID, ctc.Key, upCh, downCh)
					cancel()
					if errors.IsDeadlineExceeded(err) {
						if !ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
							t.Fatal("Expected link error due to invalid auth")
						}
					} else if ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
						t.Fatalf("Expected deadline exceeded with valid auth, but have %v", err)
					}
				})
			}
		})

		t.Run(fmt.Sprintf("Traffic/%v", ptc.Protocol), func(t *testing.T) {
			// Tests change the device registry; reset it to avoid side-effects from previous protocol tests.
			resetDeviceRegistry()

			ctx, cancel := context.WithCancel(ctx)
			upCh := make(chan *ttnpb.ApplicationUp)
			downCh := make(chan *ttnpb.ApplicationDownlink)

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := ptc.Connect(ctx, t, registeredApplicationID, registeredApplicationKey, upCh, downCh)
				if !errors.IsCanceled(err) {
					t.Fatalf("Expected context canceled, but have %v", err)
				}
			}()
			// Wait for connection to establish.
			time.Sleep(timeout)

			t.Run("Upstream", func(t *testing.T) {
				for _, tc := range []struct {
					Name         string
					Message      *ttnpb.ApplicationUp
					AssertUp     func(t *testing.T, up *ttnpb.ApplicationUp)
					AssertDevice func(t *testing.T, dev *ttnpb.EndDevice)
				}{
					{
						Name: "JoinAccept/RegisteredDevice/WithoutAppSKey",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x84, 0xff, 0xff, 0xff}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: "session1",
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x84, 0xff, 0xff, 0xff}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID: "session1",
									},
								},
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice) {
							a := assertions.New(t)
							a.So(dev.Session, should.NotBeNil)
							a.So(dev.Session.SessionKeyID, should.Equal, "session1")
							a.So(dev.Session.AppSKey, should.Resemble, &ttnpb.KeyEnvelope{
								Key:      []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
								KEKLabel: "test",
							})
						},
					},
					{
						Name: "JoinAccept/RegisteredDevice/WithAppSKey",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x42, 0xff, 0xff, 0xff}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: "session2",
									AppSKey: &ttnpb.KeyEnvelope{
										Key:      []byte{0xa3, 0x34, 0x38, 0x1c, 0xca, 0x1c, 0x12, 0x7a, 0x5b, 0xb1, 0xa8, 0x97, 0x39, 0xc7, 0x5, 0x34, 0x91, 0x26, 0x9b, 0x21, 0x4f, 0x27, 0x80, 0x19},
										KEKLabel: "test",
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x42, 0xff, 0xff, 0xff}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID: "session2",
									},
								},
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice) {
							a := assertions.New(t)
							a.So(dev.Session, should.NotBeNil)
							a.So(dev.Session.SessionKeyID, should.Equal, "session2")
							a.So(dev.Session.AppSKey, should.Resemble, &ttnpb.KeyEnvelope{
								Key:      []byte{0xa3, 0x34, 0x38, 0x1c, 0xca, 0x1c, 0x12, 0x7a, 0x5b, 0xb1, 0xa8, 0x97, 0x39, 0xc7, 0x5, 0x34, 0x91, 0x26, 0x9b, 0x21, 0x4f, 0x27, 0x80, 0x19},
								KEKLabel: "test",
							})
						},
					},
					{
						Name: "UplinkMessage/CurrentSession",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x42, 0xff, 0xff, 0xff}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: "session2",
									FPort:        42,
									FCnt:         42,
									FRMPayload:   []byte{0x66, 0xd8, 0xbf},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x42, 0xff, 0xff, 0xff}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: "session2",
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x01, 0x02, 0x03},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"sum": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 6, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name: "UplinkMessage/ChangedSession",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x24, 0x24, 0xff, 0xff}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: "session3",
									FPort:        24,
									FCnt:         24,
									FRMPayload:   []byte{0x58, 0xca, 0xa1},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x24, 0x24, 0xff, 0xff}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: "session3",
										FPort:        24,
										FCnt:         24,
										FRMPayload:   []byte{0x64, 0x64, 0x64},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"sum": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 300, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice) {
							a := assertions.New(t)
							a.So(dev.Session, should.NotBeNil)
							a.So(dev.Session.SessionKeyID, should.Equal, "session3")
							a.So(dev.Session.AppSKey, should.Resemble, &ttnpb.KeyEnvelope{
								Key:      []byte{0x8c, 0xe9, 0x14, 0x4b, 0x82, 0x23, 0x8, 0x39, 0x65, 0x73, 0xd, 0x42, 0x9f, 0x2a, 0x7c, 0x9c, 0x9c, 0xbe, 0x38, 0xbe, 0x35, 0x5d, 0x44, 0xf},
								KEKLabel: "test",
							})
						},
					},
					{
						Name: "JoinAccept/UnregisteredDevice",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(unregisteredDeviceID, types.DevAddr{0x24, 0xff, 0xff, 0xff}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: "session1",
									AppSKey: &ttnpb.KeyEnvelope{
										Key:      []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
										KEKLabel: "test",
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(unregisteredDeviceID, types.DevAddr{0x24, 0xff, 0xff, 0xff}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID: "session1",
									},
								},
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice) {
							a := assertions.New(t)
							a.So(dev.Session, should.NotBeNil)
							a.So(dev.Session.SessionKeyID, should.Equal, "session1")
							a.So(dev.Session.AppSKey, should.Resemble, &ttnpb.KeyEnvelope{
								Key:      []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
								KEKLabel: "test",
							})
						},
					},
					{
						Name: "UplinkMessage/UnregisteredDevice",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(unregisteredDeviceID, types.DevAddr{0x24, 0xff, 0xff, 0xff}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: "session1",
									FPort:        42,
									FCnt:         24,
									FRMPayload:   []byte{0x39, 0xf4, 0xb1, 0xc5},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(unregisteredDeviceID, types.DevAddr{0x24, 0xff, 0xff, 0xff}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: "session1",
										FPort:        42,
										FCnt:         24,
										FRMPayload:   []byte{0x7, 0x67, 0x0, 0xe1},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"temperature_7": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 22.5, // Application's default formatter is CayenneLPP.
													},
												},
											},
										},
									},
								},
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice) {
							a := assertions.New(t)
							a.So(dev.Session, should.NotBeNil)
							a.So(dev.Session.SessionKeyID, should.Equal, "session1")
							a.So(dev.Session.AppSKey, should.Resemble, &ttnpb.KeyEnvelope{
								Key:      []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
								KEKLabel: "test",
							})
						},
					},
				} {
					tcok := t.Run(tc.Name, func(t *testing.T) {
						ns.upCh <- tc.Message
						select {
						case msg := <-upCh:
							if tc.AssertUp != nil {
								tc.AssertUp(t, msg)
							} else {
								t.Fatalf("Expected no upstream message but got %v", msg)
							}
						case <-time.After(timeout):
							if tc.AssertUp != nil {
								t.Fatal("Expected upstream timeout")
							}
						}
						if tc.AssertDevice != nil {
							dev, err := deviceRegistry.Get(ctx, tc.Message.EndDeviceIdentifiers)
							if a.So(err, should.BeNil) {
								tc.AssertDevice(t, dev)
							}
						}
					})
					if !tcok {
						t.FailNow()
					}
				}
			})

			cancel()
			wg.Wait()
		})
	}
}
