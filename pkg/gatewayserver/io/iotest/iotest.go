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

// Package iotest implements tests for Gateway Server frontends.
package iotest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/smarty/assertions"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	gsio "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	gsredis "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/upstream/mock"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// FrontendConfig is a test bench configuration.
type FrontendConfig struct {
	// SupportsStatus indicates that the frontend sends gateway status messages.
	SupportsStatus bool
	// DetectsInvalidMessages indicates that the frontend detects invalid messages.
	DetectsInvalidMessages bool
	// DetectsDisconnect indicates that the frontend detects gateway disconnections.
	DetectsDisconnect bool
	// AuthenticatesWithEUI indicates that the gateway uses the EUI to authenticate, instead of the ID and key.
	AuthenticatesWithEUI bool
	// IsAuthenticated indicates whether the gateway connection provides authentication.
	// This is typically true for all frontends except UDP which is inherently unauthenticated.
	IsAuthenticated bool
	// DeduplicatesUplinks indicates that the frontend deduplicates uplinks that are received at once by using the IO
	// middleware's UniqueUplinkMessagesByRSSI.
	DeduplicatesUplinks bool
	// CustomComponentConfig applies custom configuration for the component before it gets started.
	// This is typically used for configuring security credentials.
	CustomComponentConfig func(config *component.Config)
	// CustomGatewayServerConfig applies custom configuration for the Gateway Server before it gets started.
	// This is typically used for configuring frontend listeners.
	CustomGatewayServerConfig func(gsConfig *gatewayserver.Config)
	// Link links the gateway.
	Link func(
		ctx context.Context,
		t *testing.T,
		gs *gatewayserver.GatewayServer,
		ids *ttnpb.GatewayIdentifiers,
		key string,
		upCh <-chan *ttnpb.GatewayUp,
		downCh chan<- *ttnpb.GatewayDown,
	) error
}

// Frontend tests a frontend.
func Frontend(t *testing.T, frontend FrontendConfig) { //nolint:gocyclo
	t.Helper()

	var (
		registeredGatewayID    = "eui-aaee000000000000"
		registeredGatewayKey   = "secret"
		registeredGatewayEUI   = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		unregisteredGatewayEUI = types.EUI64{0xBB, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
		timeout                = (1 << 5) * test.Delay
		testRights             = []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK, ttnpb.Right_RIGHT_GATEWAY_STATUS_READ}
	)

	a, ctx := test.New(t)

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	ns, nsAddr := mock.StartNS(ctx)

	var statsRegistry gatewayserver.GatewayConnectionStatsRegistry
	if os.Getenv("TEST_REDIS") == "1" {
		statsRedisClient, statsFlush := test.NewRedis(ctx, "gatewayserver_test")
		defer statsFlush()
		defer statsRedisClient.Close()
		registry := &gsredis.GatewayConnectionStatsRegistry{
			Redis:   statsRedisClient,
			LockTTL: test.Delay << 10,
		}
		if err := registry.Init(ctx); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		statsRegistry = registry
	}

	componentConfig := &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	}
	if frontend.CustomComponentConfig != nil {
		frontend.CustomComponentConfig(componentConfig)
	}
	c := componenttest.NewComponent(t, componentConfig)

	gsConfig := &gatewayserver.Config{
		RequireRegisteredGateways:         true,
		UpdateGatewayLocationDebounceTime: 0,
		UpdateConnectionStatsInterval:     (1 << 4) * test.Delay,
		ConnectionStatsTTL:                (1 << 6) * test.Delay,
		ConnectionStatsDisconnectTTL:      (1 << 7) * test.Delay,
		Stats:                             statsRegistry,
		FetchGatewayInterval:              (1 << 3) * test.Delay,
		FetchGatewayJitter:                0.1,
	}
	if frontend.CustomGatewayServerConfig != nil {
		frontend.CustomGatewayServerConfig(gsConfig)
	}
	er := gatewayserver.NewIS(c)
	gs, err := gatewayserver.New(c, gsConfig,
		gatewayserver.WithRegistry(er),
	)
	a.So(err, should.BeNil)

	defer c.Close()
	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER)

	componenttest.StartComponent(t, c)

	mustHavePeer(ctx, t, c, ttnpb.ClusterRole_NETWORK_SERVER)
	mustHavePeer(ctx, t, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	ids := &ttnpb.GatewayIdentifiers{
		GatewayId: registeredGatewayID,
		Eui:       registeredGatewayEUI.Bytes(),
	}
	gtw := mockis.DefaultGateway(ids, true, true)
	is.GatewayRegistry().Add(ctx, ids, registeredGatewayKey, gtw, testRights...)

	time.Sleep(timeout) // Wait for setup to be completed.

	t.Run("Authenticate", func(t *testing.T) {
		for _, ctc := range []struct {
			Name string
			ID   *ttnpb.GatewayIdentifiers
			Key  string
		}{
			{
				Name: "ValidIDAndKey",
				ID:   &ttnpb.GatewayIdentifiers{GatewayId: registeredGatewayID},
				Key:  registeredGatewayKey,
			},
			{
				Name: "InvalidKey",
				ID:   &ttnpb.GatewayIdentifiers{GatewayId: registeredGatewayID},
				Key:  "invalid-key",
			},
			{
				Name: "InvalidIDAndKey",
				ID:   &ttnpb.GatewayIdentifiers{GatewayId: "invalid-gateway"},
				Key:  "invalid-key",
			},
			{
				Name: "RegisteredEUI",
				ID:   &ttnpb.GatewayIdentifiers{Eui: registeredGatewayEUI.Bytes()},
			},
			{
				Name: "UnregisteredEUI",
				ID:   &ttnpb.GatewayIdentifiers{Eui: unregisteredGatewayEUI.Bytes()},
			},
		} {
			ctc := ctc
			t.Run(ctc.Name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctx)
				upCh := make(chan *ttnpb.GatewayUp)
				downCh := make(chan *ttnpb.GatewayDown)

				upEvents := map[string]events.Channel{}
				for _, event := range []string{"gs.gateway.connect"} {
					upEvents[event] = make(events.Channel, 5)
				}
				defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{ //nolint:revive
					PublishFunc: func(evs ...events.Event) {
						for _, ev := range evs {
							ev := ev
							switch name := ev.Name(); name {
							case "gs.gateway.connect":
								go func() {
									upEvents[name] <- ev
								}()
							default:
								t.Logf("%s event published", name)
							}
						}
					},
				})()

				var validAuth func(*ttnpb.GatewayIdentifiers, string) bool
				if frontend.AuthenticatesWithEUI {
					validAuth = func(ids *ttnpb.GatewayIdentifiers, _ string) bool {
						return bytes.Equal(ids.Eui, registeredGatewayEUI.Bytes())
					}
				} else {
					validAuth = func(ids *ttnpb.GatewayIdentifiers, key string) bool {
						return ids.GatewayId == registeredGatewayID && key == registeredGatewayKey
					}
				}

				connectedWithInvalidAuth := make(chan struct{}, 1)
				expectedProperLink := make(chan struct{}, 1)
				go func() {
					select {
					case <-upEvents["gs.gateway.connect"]:
						if !validAuth(ctc.ID, ctc.Key) {
							connectedWithInvalidAuth <- struct{}{}
						}
					case <-time.After(timeout):
						if validAuth(ctc.ID, ctc.Key) {
							expectedProperLink <- struct{}{}
						}
					}
					time.Sleep(test.Delay)
					cancel()
				}()
				err := frontend.Link(ctx, t, gs, ctc.ID, ctc.Key, upCh, downCh)
				if !errors.IsCanceled(err) && validAuth(ctc.ID, ctc.Key) {
					t.Fatalf("Expect canceled context but have %v", err)
				}
				select {
				case <-connectedWithInvalidAuth:
					t.Fatal("Expected link error due to invalid auth")
				case <-expectedProperLink:
					t.Fatal("Expected proper link")
				default:
				}
			})
		}
	})

	// Wait for gateway disconnection to be processed.
	time.Sleep(timeout)

	t.Run("DetectDisconnect", func(t *testing.T) {
		if !frontend.DetectsDisconnect {
			t.SkipNow()
		}

		id := &ttnpb.GatewayIdentifiers{
			GatewayId: registeredGatewayID,
			Eui:       registeredGatewayEUI.Bytes(),
		}

		ctx1, fail1 := errorcontext.New(ctx)
		defer fail1(context.Canceled)
		go func() {
			upCh := make(chan *ttnpb.GatewayUp)
			downCh := make(chan *ttnpb.GatewayDown)
			err := frontend.Link(ctx1, t, gs, id, registeredGatewayKey, upCh, downCh)
			fail1(err)
		}()
		select {
		case <-ctx1.Done():
			t.Fatalf("Expected no link error on first connection but have %v", ctx1.Err())
		case <-time.After(timeout):
		}

		// Establish a second connection that lasts for a short time, just to disconnect the first one.
		ctx2, cancel2 := context.WithDeadline(ctx, time.Now().Add((1<<3)*timeout))
		upCh := make(chan *ttnpb.GatewayUp)
		downCh := make(chan *ttnpb.GatewayDown)
		err := frontend.Link(ctx2, t, gs, id, registeredGatewayKey, upCh, downCh)
		cancel2()
		if !errors.IsDeadlineExceeded(err) {
			t.Fatalf("Expected deadline exceeded on second connection but have %v", err)
		}
		select {
		case <-ctx1.Done():
			t.Logf("First connection failed when second connected with %v", ctx1.Err())
		case <-time.After((1 << 3) * timeout):
			t.Fatalf("Expected link failure on first connection when second connected")
		}
	})

	// Wait for gateway disconnection to be processed.
	time.Sleep(2 * gsConfig.ConnectionStatsDisconnectTTL)

	t.Run("Traffic", func(t *testing.T) {
		a := assertions.New(t)

		ctx, cancel := context.WithCancel(ctx)
		upCh := make(chan *ttnpb.GatewayUp)
		downCh := make(chan *ttnpb.GatewayDown)
		ids := &ttnpb.GatewayIdentifiers{
			GatewayId: registeredGatewayID,
			Eui:       registeredGatewayEUI.Bytes(),
		}
		// Setup a stats client with independent context to query whether the gateway is connected and statistics on
		// upstream and downstream.
		statsCtx := metadata.AppendToOutgoingContext(test.Context(),
			"id", ids.GatewayId,
			"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
		)
		statsClient := ttnpb.NewGsClient(gs.LoopbackConn())

		// The gateway should not be connected before testing traffic.
		t.Run("NotConnected", func(t *testing.T) {
			_, err := statsClient.GetGatewayConnectionStats(statsCtx, ids)
			if !a.So(errors.IsNotFound(err), should.BeTrue) {
				t.Fatal("Expected gateway not to be connected yet, but it is")
			}
		})

		if frontend.SupportsStatus && frontend.IsAuthenticated {
			t.Run("UpdateLocation", func(t *testing.T) {
				for _, tc := range []struct {
					Name           string
					UpdateLocation bool
					Up             *ttnpb.GatewayUp
					ExpectLocation *ttnpb.Location
				}{
					{
						Name:           "NoUpdate",
						UpdateLocation: false,
						Up: &ttnpb.GatewayUp{
							GatewayStatus: &ttnpb.GatewayStatus{
								Time: timestamppb.New(time.Unix(424242, 0)),
								AntennaLocations: []*ttnpb.Location{
									{
										Source:    ttnpb.LocationSource_SOURCE_GPS,
										Altitude:  10,
										Latitude:  12,
										Longitude: 14,
									},
								},
							},
						},
						ExpectLocation: &ttnpb.Location{
							Source: ttnpb.LocationSource_SOURCE_GPS,
						},
					},
					{
						Name:           "NoLocation",
						UpdateLocation: true,
						Up: &ttnpb.GatewayUp{
							GatewayStatus: &ttnpb.GatewayStatus{
								Time: timestamppb.New(time.Unix(424242, 0)),
							},
						},
						ExpectLocation: &ttnpb.Location{
							Source: ttnpb.LocationSource_SOURCE_GPS,
						},
					},
					{
						Name:           "Update",
						UpdateLocation: true,
						Up: &ttnpb.GatewayUp{
							GatewayStatus: &ttnpb.GatewayStatus{
								Time: timestamppb.New(time.Unix(42424242, 0)),
								AntennaLocations: []*ttnpb.Location{
									{
										Source:    ttnpb.LocationSource_SOURCE_GPS,
										Altitude:  10,
										Latitude:  12,
										Longitude: 14,
									},
								},
							},
						},
						ExpectLocation: &ttnpb.Location{
							Source:    ttnpb.LocationSource_SOURCE_GPS,
							Altitude:  10,
							Latitude:  12,
							Longitude: 14,
						},
					},
				} {
					t.Run(tc.Name, func(t *testing.T) {
						a := assertions.New(t)

						gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
							GatewayIds: ids,
						})
						a.So(err, should.BeNil)

						gtw.Antennas[0].Location = &ttnpb.Location{
							Source: ttnpb.LocationSource_SOURCE_GPS,
						}
						gtw.UpdateLocationFromStatus = tc.UpdateLocation
						gtw, err = is.GatewayRegistry().Update(ctx, &ttnpb.UpdateGatewayRequest{
							Gateway:   gtw,
							FieldMask: ttnpb.FieldMask("antennas", "update_location_from_status"),
						})
						a.So(err, should.BeNil)
						a.So(gtw.UpdateLocationFromStatus, should.Equal, tc.UpdateLocation)

						ctx, cancel := context.WithCancel(ctx)
						upCh := make(chan *ttnpb.GatewayUp)
						downCh := make(chan *ttnpb.GatewayDown)

						upEvents := map[string]events.Channel{}
						for _, event := range []string{
							"gs.gateway.connect",
							"gs.gateway.disconnect",
						} {
							upEvents[event] = make(events.Channel, 5)
						}
						defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{ //nolint:revive
							PublishFunc: func(evs ...events.Event) {
								for _, ev := range evs {
									ev := ev
									t.Logf("%s event published", ev.Name())
									switch name := ev.Name(); name {
									case "gs.gateway.connect":
										go func() {
											upEvents[name] <- ev
										}()
									case "gs.gateway.disconnect":
										go func() {
											upEvents[name] <- ev
										}()
									}
								}
							},
						})()

						wg := &sync.WaitGroup{}
						wg.Add(1)
						var linkErr error
						go func() {
							defer wg.Done()
							linkErr = frontend.Link(ctx, t, gs, ids, registeredGatewayKey, upCh, downCh)
						}()

						select {
						case <-upEvents["gs.gateway.connect"]:
						case <-time.After(timeout):
							t.Fatal("Expected gateway to be connected, but it is not")
						}

						select {
						case upCh <- tc.Up:
						case <-time.After(timeout):
							t.Fatalf("Failed to send message to upstream channel")
						}

						time.Sleep(timeout)
						gtw, err = is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
							GatewayIds: ids,
						})
						a.So(err, should.BeNil)
						a.So(gtw.Antennas[0].Location, should.Resemble, tc.ExpectLocation)

						cancel()
						wg.Wait()
						if !errors.IsCanceled(linkErr) {
							t.Fatalf("Expected context canceled, but have %v", linkErr)
						}
						select {
						case <-upEvents["gs.gateway.disconnect"]:
						case <-time.After(timeout):
							t.Fatal("Expected gateway to be disconnected, but it has not disconnected")
						}
					})
				}
			})
		}

		t.Run("Disconnection", func(t *testing.T) {
			for _, tc := range []struct {
				Name               string
				AntennaGain        float32
				ExpectDisconnected bool
			}{
				{
					Name:               "NoDisconnect",
					AntennaGain:        0,
					ExpectDisconnected: false,
				},
				{
					Name:               "Disconnect",
					AntennaGain:        3,
					ExpectDisconnected: true,
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					a := assertions.New(t)

					ctx, cancel := context.WithCancel(ctx)
					upCh := make(chan *ttnpb.GatewayUp)
					downCh := make(chan *ttnpb.GatewayDown)

					upEvents := map[string]events.Channel{}
					for _, event := range []string{
						"gs.gateway.connect",
						"gs.gateway.disconnect",
					} {
						upEvents[event] = make(events.Channel, 5)
					}
					defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{ //nolint:revive
						PublishFunc: func(evs ...events.Event) {
							for _, ev := range evs {
								ev := ev
								t.Logf("%s event published", ev.Name())
								switch name := ev.Name(); name {
								case "gs.gateway.connect":
									go func() {
										upEvents[name] <- ev
									}()
								case "gs.gateway.disconnect":
									go func() {
										upEvents[name] <- ev
									}()
								}
							}
						},
					})()

					wg := &sync.WaitGroup{}
					wg.Add(1)
					var linkErr error
					go func() {
						defer wg.Done()
						linkErr = frontend.Link(ctx, t, gs, ids, registeredGatewayKey, upCh, downCh)
					}()

					select {
					case <-upEvents["gs.gateway.connect"]:
					case <-time.After(timeout):
						t.Fatal("Expected gateway to be connected, but it is not")
					}

					gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
						GatewayIds: ids,
						FieldMask:  ttnpb.FieldMask("antennas"),
					})
					a.So(err, should.BeNil)
					gtw.Antennas[0].Gain = tc.AntennaGain
					_, err = is.GatewayRegistry().Update(ctx, &ttnpb.UpdateGatewayRequest{
						Gateway:   gtw,
						FieldMask: ttnpb.FieldMask("antennas"),
					})
					a.So(err, should.BeNil)

					select {
					case <-upEvents["gs.gateway.disconnect"]:
						if !tc.ExpectDisconnected {
							t.Fatal("Expected gateway not to be disconnected, but it has disconnected")
						}
					case <-time.After(timeout):
						if tc.ExpectDisconnected {
							t.Fatal("Expected gateway to be disconnected, but it has not disconnected")
						}
					}

					_, connected := gs.GetConnection(ctx, ids)
					if !a.So(connected, should.Equal, !tc.ExpectDisconnected) {
						t.Fatal("Expected gateway to be disconnected, but it is not")
					}

					cancel()
					wg.Wait()
					if !tc.ExpectDisconnected {
						if !errors.IsCanceled(linkErr) {
							t.Fatalf("Expected context canceled, but have %v", linkErr)
						}
						select {
						case <-upEvents["gs.gateway.disconnect"]:
						case <-time.After(timeout):
							t.Fatal("Expected gateway to be disconnected, but it has not disconnected")
						}
					}
				})
			}
		})

		t.Run("LocationMetadata", func(t *testing.T) {
			location := &ttnpb.Location{
				Source:    ttnpb.LocationSource_SOURCE_GPS,
				Altitude:  10,
				Latitude:  12,
				Longitude: 14,
			}
			up := &ttnpb.GatewayUp{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{
									Lora: &ttnpb.LoRaDataRate{
										SpreadingFactor: 7,
										Bandwidth:       250000,
										CodingRate:      band.Cr4_5,
									},
								},
							},
							Frequency: 867900000,
							Timestamp: 100,
						},
						RxMetadata: []*ttnpb.RxMetadata{
							{
								AntennaIndex: 0,
								GatewayIds:   ids,
								Timestamp:    100,
								Rssi:         -69,
								ChannelRssi:  -69,
								Snr:          11,
							},
						},
					},
				},
			}
			for _, locationPublic := range []bool{false, true} {
				t.Run(fmt.Sprintf("LocationPublic=%v", locationPublic), func(t *testing.T) {
					mockGtw := mockis.DefaultGateway(ids, false, false)
					mockGtw.LocationPublic = locationPublic
					mockGtw.Antennas[0].Location = location
					is.GatewayRegistry().Add(ctx, ids, registeredGatewayKey, mockGtw, testRights...)

					for _, locationInRxMetadata := range []bool{false, true} {
						t.Run(fmt.Sprintf("RxMetadata=%v", locationInRxMetadata), func(t *testing.T) {
							if !locationInRxMetadata && locationPublic {
								// Disabled, because this is inconsistent amongst frontends
								// - gRPC and MQTT: location is in RxMetadata
								// - UDP and BasicStation: location is not in RxMetadata
								t.SkipNow()
							}
							a := assertions.New(t)

							ctx, cancel := context.WithCancel(ctx)
							upCh := make(chan *ttnpb.GatewayUp)
							downCh := make(chan *ttnpb.GatewayDown)

							upEvents := map[string]events.Channel{}
							for _, event := range []string{
								"gs.gateway.connect",
								"gs.gateway.disconnect",
							} {
								upEvents[event] = make(events.Channel, 5)
							}
							defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{ //nolint:revive
								PublishFunc: func(evs ...events.Event) {
									for _, ev := range evs {
										ev := ev
										switch name := ev.Name(); name {
										case "gs.gateway.connect":
											go func() {
												upEvents[name] <- ev
											}()
										case "gs.gateway.disconnect":
											go func() {
												upEvents[name] <- ev
											}()
										default:
											t.Logf("%s event published", name)
										}
									}
								},
							})()

							wg := &sync.WaitGroup{}
							wg.Add(1)
							var linkErr error
							go func() {
								defer wg.Done()
								linkErr = frontend.Link(ctx, t, gs, ids, registeredGatewayKey, upCh, downCh)
							}()

							select {
							case <-upEvents["gs.gateway.connect"]:
							case <-time.After(timeout):
							}

							if locationInRxMetadata {
								up.UplinkMessages[0].RxMetadata[0].Location = location
							} else {
								up.UplinkMessages[0].RxMetadata[0].Location = nil
							}
							up.UplinkMessages[0].RawPayload = randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6)

							select {
							case upCh <- up:
							case <-time.After(timeout):
								t.Fatalf("Failed to send message to upstream channel")
							}

							select {
							case msg := <-ns.Up():
								if a.So(len(msg.RxMetadata), should.Equal, 1) {
									if locationPublic {
										a.So(msg.RxMetadata[0].Location, should.Resemble, location)
									} else {
										a.So(msg.RxMetadata[0].Location, should.BeNil)
									}
								}
							case <-time.After(2 * timeout):
								t.Fatalf("Failed to get message")
							}

							cancel()
							wg.Wait()
							if !errors.IsCanceled(linkErr) {
								t.Fatalf("Expected context canceled, but have %v", linkErr)
							}

							select {
							case <-upEvents["gs.gateway.disconnect"]:
							case <-time.After(timeout):
							}
						})
					}
				})
			}
		})

		// Wait for gateway disconnection to be processed.
		time.Sleep((1 << 5) * test.Delay)

		wg := &sync.WaitGroup{}
		wg.Add(1)
		var linkErr error
		go func() {
			defer wg.Done()
			linkErr = frontend.Link(ctx, t, gs, ids, registeredGatewayKey, upCh, downCh)
		}()

		// Wait for gateway connection to be established.
		time.Sleep((1 << 5) * test.Delay)

		// Expected location for RxMetadata
		gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIds: ids,
		})
		a.So(err, should.BeNil)
		location := gtw.Antennas[0].Location

		duplicatePayload := randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6)

		t.Run("Upstream", func(t *testing.T) {
			uplinkCount := 0
			for _, tc := range []struct {
				Name                         string
				Up                           *ttnpb.GatewayUp
				Received                     []uint32 // Timestamps of uplink messages in Up that are received.
				Dropped                      []uint32 // Timestamps of uplink messages in Up that are dropped.
				PublicLocation               bool     // If gateway location is public, it should be in RxMetadata
				UplinkCount                  int      // Number of expected uplinks
				RepeatUpEvent                bool     // Expect event for repeated uplinks
				SkipIfDetectsInvalidMessages bool     // Skip this test if the frontend detects invalid messages
			}{
				{
					Name: "GatewayStatus",
					Up: &ttnpb.GatewayUp{
						GatewayStatus: &ttnpb.GatewayStatus{
							Time: timestamppb.New(time.Unix(424242, 0)),
						},
					},
				},
				{
					Name: "TxAck",
					Up: &ttnpb.GatewayUp{
						TxAcknowledgment: &ttnpb.TxAcknowledgment{
							Result: ttnpb.TxAcknowledgment_SUCCESS,
						},
					},
				},
				{
					Name: "CRCFailure",
					Up: &ttnpb.GatewayUp{
						UplinkMessages: []*ttnpb.UplinkMessage{
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 7,
												Bandwidth:       250000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 867900000,
									Timestamp: 100,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   100,
										Rssi:        -69,
										ChannelRssi: -69,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 2, 2),
								CrcStatus:  wrapperspb.Bool(false),
							},
						},
					},
					Received:                     []uint32{100},
					Dropped:                      []uint32{100},
					SkipIfDetectsInvalidMessages: true,
				},
				{
					Name: "OneValidLoRa",
					Up: &ttnpb.GatewayUp{
						UplinkMessages: []*ttnpb.UplinkMessage{
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 7,
												Bandwidth:       250000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 867900000,
									Timestamp: 200,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   200,
										Rssi:        -69,
										ChannelRssi: -69,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
							},
						},
					},
					Received: []uint32{200},
				},
				{
					Name: "OneValidLoRaAndTwoRepeated",
					Up: &ttnpb.GatewayUp{
						UplinkMessages: []*ttnpb.UplinkMessage{
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 7,
												Bandwidth:       250000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 867900000,
									Timestamp: 301,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   301,
										Rssi:        -42,
										ChannelRssi: -42,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: duplicatePayload,
							},
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 7,
												Bandwidth:       250000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 867900000,
									Timestamp: 300,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   300,
										Rssi:        -69,
										ChannelRssi: -69,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: duplicatePayload,
							},
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 7,
												Bandwidth:       250000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 867900000,
									Timestamp: 300,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   300,
										Rssi:        -69,
										ChannelRssi: -69,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: duplicatePayload,
							},
						},
					},
					Received:      []uint32{301},
					UplinkCount:   1,
					RepeatUpEvent: true,
				},
				{
					Name: "OneValidFSK",
					Up: &ttnpb.GatewayUp{
						UplinkMessages: []*ttnpb.UplinkMessage{
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Fsk{
											Fsk: &ttnpb.FSKDataRate{
												BitRate: 50000,
											},
										},
									},
									Frequency: 867900000,
									Timestamp: 400,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   400,
										Rssi:        -69,
										ChannelRssi: -69,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
							},
						},
					},
					Received: []uint32{400},
				},
				{
					Name: "OneGarbageWithStatus",
					Up: &ttnpb.GatewayUp{
						UplinkMessages: []*ttnpb.UplinkMessage{
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 9,
												Bandwidth:       125000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 868500000,
									Timestamp: 500,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   500,
										Rssi:        -112,
										ChannelRssi: -112,
										Snr:         2,
										Location:    location,
									},
								},
								RawPayload: []byte{0xff, 0x02, 0x03}, // Garbage; doesn't get forwarded.
							},
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 7,
												Bandwidth:       125000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 868100000,
									Timestamp: 501,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   501,
										Rssi:        -69,
										ChannelRssi: -69,
										Snr:         11,
										Location:    location,
									},
								},
								RawPayload: randomUpDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
							},
							{
								Settings: &ttnpb.TxSettings{
									DataRate: &ttnpb.DataRate{
										Modulation: &ttnpb.DataRate_Lora{
											Lora: &ttnpb.LoRaDataRate{
												SpreadingFactor: 12,
												Bandwidth:       125000,
												CodingRate:      band.Cr4_5,
											},
										},
									},
									Frequency: 867700000,
									Timestamp: 502,
								},
								RxMetadata: []*ttnpb.RxMetadata{
									{
										GatewayIds:  ids,
										Timestamp:   502,
										Rssi:        -36,
										ChannelRssi: -36,
										Snr:         5,
										Location:    location,
									},
								},
								RawPayload: randomJoinRequestPayload(
									types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
									types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
								),
							},
						},
						GatewayStatus: &ttnpb.GatewayStatus{
							Time: timestamppb.New(time.Unix(4242424, 0)),
						},
					},
					Received: []uint32{501, 502},
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					a := assertions.New(t)

					if tc.SkipIfDetectsInvalidMessages && frontend.DetectsInvalidMessages {
						t.Skip("Skipping test case because gateway detects invalid messages")
					}

					upEvents := map[string]events.Channel{}
					for _, event := range []string{
						"gs.up.receive",
						"gs.down.tx.success",
						"gs.down.tx.fail",
						"gs.status.receive",
						"gs.io.up.repeat",
					} {
						upEvents[event] = make(events.Channel, 5)
					}
					defer test.SetDefaultEventsPubSub(&test.MockEventPubSub{ //nolint:revive
						PublishFunc: func(evs ...events.Event) {
							for _, ev := range evs {
								ev := ev
								switch name := ev.Name(); name {
								case "gs.up.receive", "gs.down.tx.success", "gs.down.tx.fail", "gs.status.receive", "gs.io.up.repeat":
									go func() {
										upEvents[name] <- ev
									}()
								default:
									t.Logf("%s event published", name)
								}
							}
						},
					})()

					select {
					case upCh <- tc.Up:
					case <-time.After(timeout):
						t.Fatalf("Failed to send message to upstream channel")
					}
					if tc.UplinkCount > 0 {
						uplinkCount += tc.UplinkCount
					} else if frontend.DetectsInvalidMessages {
						uplinkCount += len(tc.Received)
					} else {
						uplinkCount += len(tc.Up.UplinkMessages)
					}

					if tc.RepeatUpEvent && !frontend.DeduplicatesUplinks {
						select {
						case evt := <-upEvents["gs.io.up.repeat"]:
							a.So(evt.Name(), should.Equal, "gs.io.up.repeat")
						case <-time.After(timeout):
							t.Fatal("Expected repeat uplink event timeout")
						}
					}

					received := make(map[uint32]struct{})
					forwarded := make(map[uint32]struct{})
					for _, t := range tc.Received {
						received[t] = struct{}{}
						forwarded[t] = struct{}{}
					}
					for _, t := range tc.Dropped {
						delete(forwarded, t)
					}
					for len(received) > 0 {
						if len(forwarded) > 0 {
							select {
							case msg := <-ns.Up():
								var expected *ttnpb.UplinkMessage
								for _, up := range tc.Up.UplinkMessages {
									if ts := up.Settings.Timestamp; ts == msg.Settings.Timestamp {
										if _, ok := forwarded[ts]; !ok {
											t.Fatalf("Not expecting message %v", msg)
										}
										expected = up
										delete(forwarded, ts)
										break
									}
								}
								if expected == nil {
									t.Fatalf("Received unexpected message with timestamp %d", msg.Settings.Timestamp)
								}
								a.So(time.Since(*ttnpb.StdTime(msg.ReceivedAt)), should.BeLessThan, timeout)
								a.So(msg.Settings, should.Resemble, expected.Settings)
								a.So(len(msg.RxMetadata), should.Equal, len(expected.RxMetadata))
								for i, md := range msg.RxMetadata {
									a.So(md.UplinkToken, should.NotBeEmpty)
									md.UplinkToken = nil
									md.ReceivedAt = nil
									a.So(md, should.Resemble, expected.RxMetadata[i])
								}
								a.So(msg.RawPayload, should.Resemble, expected.RawPayload)
							case <-time.After(timeout):
								t.Fatal("Expected uplink timeout")
							}
						}
						select {
						case evt := <-upEvents["gs.up.receive"]:
							a.So(evt.Name(), should.Equal, "gs.up.receive")
							msg := evt.Data().(*ttnpb.GatewayUplinkMessage) //nolint:revive
							delete(received, msg.Message.Settings.Timestamp)
						case <-time.After(timeout):
							t.Fatal("Expected uplink event timeout")
						}
					}
					if expected := tc.Up.TxAcknowledgment; expected != nil {
						select {
						case <-upEvents["gs.down.tx.success"]:
						case evt := <-upEvents["gs.down.tx.fail"]:
							received, ok := evt.Data().(ttnpb.TxAcknowledgment_Result)
							if !ok {
								t.Fatal("No acknowledgment attached to the downlink emission fail event")
							}
							a.So(received, should.Resemble, expected.Result)
						case <-time.After(timeout):
							t.Fatal("Expected Tx acknowledgment event timeout")
						}
						select {
						case ack := <-ns.TxAck():
							if txAck := ack.GetTxAck(); a.So(txAck, should.NotBeNil) {
								a.So(txAck.Result, should.Resemble, expected.Result)
								a.So(txAck.DownlinkMessage, should.Resemble, expected.DownlinkMessage)
							}
						case <-time.After(timeout):
							t.Fatal("Expected Tx acknowledgment event timeout")
						}
					}
					if tc.Up.GatewayStatus != nil && frontend.SupportsStatus {
						select {
						case <-upEvents["gs.status.receive"]:
						case <-time.After(timeout):
							t.Fatal("Expected gateway status event timeout")
						}
					}

					time.Sleep(timeout)

					conn, ok := gs.GetConnection(ctx, ids)
					a.So(ok, should.BeTrue)

					stats, paths := conn.Stats()
					a.So(stats, should.NotBeNil)
					a.So(paths, should.NotBeEmpty)

					stats, err := statsClient.GetGatewayConnectionStats(statsCtx, ids)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					a.So(stats.UplinkCount, should.Equal, uplinkCount)

					if tc.Up.GatewayStatus != nil && frontend.SupportsStatus {
						if !a.So(stats.LastStatus, should.NotBeNil) {
							t.FailNow()
						}
						a.So(stats.LastStatus.Time, should.Resemble, tc.Up.GatewayStatus.Time)
					}
				})
			}
		})

		t.Run("Downstream", func(t *testing.T) {
			ctx := clusterauth.NewContext(test.Context(), nil)
			downlinkCount := 0
			for _, tc := range []struct {
				Name                     string
				Message                  *ttnpb.DownlinkMessage
				ErrorAssertion           func(error) bool
				RxWindowDetailsAssertion []func(error) bool
			}{
				{
					Name: "InvalidSettingsType",
					Message: &ttnpb.DownlinkMessage{
						RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
						Settings: &ttnpb.DownlinkMessage_Scheduled{
							Scheduled: &ttnpb.TxSettings{
								DataRate: &ttnpb.DataRate{
									Modulation: &ttnpb.DataRate_Lora{
										Lora: &ttnpb.LoRaDataRate{
											SpreadingFactor: 12,
											Bandwidth:       125000,
											CodingRate:      band.Cr4_5,
										},
									},
								},
								Frequency: 869525000,
								Downlink: &ttnpb.TxSettings_Downlink{
									TxPower: 10,
								},
								Timestamp: 100,
							},
						},
					},
					ErrorAssertion: errors.IsInvalidArgument, // Network Server may send Tx request only.
				},
				{
					Name: "NotConnected",
					Message: &ttnpb.DownlinkMessage{
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.Class_CLASS_C,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_Fixed{
											Fixed: &ttnpb.GatewayAntennaIdentifiers{
												GatewayIds: &ttnpb.GatewayIdentifiers{
													GatewayId: "not-connected",
												},
											},
										},
									},
								},
								FrequencyPlanId: test.EUFrequencyPlanID,
							},
						},
					},
					ErrorAssertion: errors.IsAborted, // The gateway is not connected.
				},
				{
					Name: "ValidClassA",
					Message: &ttnpb.DownlinkMessage{
						RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.Class_CLASS_A,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: gsio.MustUplinkToken(
												&ttnpb.GatewayAntennaIdentifiers{
													GatewayIds: &ttnpb.GatewayIdentifiers{
														GatewayId: registeredGatewayID,
													},
												},
												10000000,
												10000000000,
												time.Unix(0, 10000000*1000),
												nil,
											),
										},
									},
								},
								Priority: ttnpb.TxSchedulePriority_NORMAL,
								Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
								Rx1DataRate: &ttnpb.DataRate{
									Modulation: &ttnpb.DataRate_Lora{
										Lora: &ttnpb.LoRaDataRate{
											SpreadingFactor: 7,
											Bandwidth:       125000,
											CodingRate:      band.Cr4_5,
										},
									},
								},
								Rx1Frequency:    868100000,
								FrequencyPlanId: test.EUFrequencyPlanID,
							},
						},
					},
				},
				{
					Name: "ValidClassAWithoutFrequencyPlanInTxRequest",
					Message: &ttnpb.DownlinkMessage{
						RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x01, 0xff, 0xff}, 1, 6),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.Class_CLASS_A,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: gsio.MustUplinkToken(
												&ttnpb.GatewayAntennaIdentifiers{
													GatewayIds: &ttnpb.GatewayIdentifiers{
														GatewayId: registeredGatewayID,
													},
												},
												20000000,
												20000000000,
												time.Unix(0, 20000000*1000),
												nil,
											),
										},
									},
								},
								Priority: ttnpb.TxSchedulePriority_NORMAL,
								Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
								Rx1DataRate: &ttnpb.DataRate{
									Modulation: &ttnpb.DataRate_Lora{
										Lora: &ttnpb.LoRaDataRate{
											SpreadingFactor: 7,
											Bandwidth:       125000,
											CodingRate:      band.Cr4_5,
										},
									},
								},
								Rx1Frequency: 868100000,
							},
						},
					},
				},
				{
					Name: "ConflictClassA",
					Message: &ttnpb.DownlinkMessage{
						RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.Class_CLASS_A,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_UplinkToken{
											UplinkToken: gsio.MustUplinkToken(
												&ttnpb.GatewayAntennaIdentifiers{
													GatewayIds: &ttnpb.GatewayIdentifiers{
														GatewayId: registeredGatewayID,
													},
												},
												10000000,
												10000000000,
												time.Unix(0, 10000000*1000),
												nil,
											),
										},
									},
								},
								Priority: ttnpb.TxSchedulePriority_NORMAL,
								Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
								Rx1DataRate: &ttnpb.DataRate{
									Modulation: &ttnpb.DataRate_Lora{
										Lora: &ttnpb.LoRaDataRate{
											SpreadingFactor: 7,
											Bandwidth:       125000,
											CodingRate:      band.Cr4_5,
										},
									},
								},
								Rx1Frequency:    868100000,
								FrequencyPlanId: test.EUFrequencyPlanID,
							},
						},
					},
					ErrorAssertion: errors.IsAborted,
					RxWindowDetailsAssertion: []func(error) bool{
						errors.IsAlreadyExists,      // Rx1 conflicts with previous.
						errors.IsFailedPrecondition, // Rx2 not provided.
					},
				},
				{
					Name: "ValidClassC",
					Message: &ttnpb.DownlinkMessage{
						RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 1, 6),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.Class_CLASS_C,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_Fixed{
											Fixed: &ttnpb.GatewayAntennaIdentifiers{
												GatewayIds: &ttnpb.GatewayIdentifiers{
													GatewayId: registeredGatewayID,
												},
											},
										},
									},
								},
								Priority: ttnpb.TxSchedulePriority_NORMAL,
								Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
								Rx1DataRate: &ttnpb.DataRate{
									Modulation: &ttnpb.DataRate_Lora{
										Lora: &ttnpb.LoRaDataRate{
											SpreadingFactor: 7,
											Bandwidth:       125000,
											CodingRate:      band.Cr4_5,
										},
									},
								},
								Rx1Frequency:    868100000,
								FrequencyPlanId: test.EUFrequencyPlanID,
							},
						},
					},
				},
				{
					Name: "ValidClassCWithoutFrequencyPlanInTxRequest",
					Message: &ttnpb.DownlinkMessage{
						RawPayload: randomDownDataPayload(types.DevAddr{0x26, 0x02, 0xff, 0xff}, 42, 2),
						Settings: &ttnpb.DownlinkMessage_Request{
							Request: &ttnpb.TxRequest{
								Class: ttnpb.Class_CLASS_C,
								DownlinkPaths: []*ttnpb.DownlinkPath{
									{
										Path: &ttnpb.DownlinkPath_Fixed{
											Fixed: &ttnpb.GatewayAntennaIdentifiers{
												GatewayIds: &ttnpb.GatewayIdentifiers{
													GatewayId: registeredGatewayID,
												},
											},
										},
									},
								},
								Priority: ttnpb.TxSchedulePriority_NORMAL,
								Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
								Rx1DataRate: &ttnpb.DataRate{
									Modulation: &ttnpb.DataRate_Lora{
										Lora: &ttnpb.LoRaDataRate{
											SpreadingFactor: 7,
											Bandwidth:       125000,
											CodingRate:      band.Cr4_5,
										},
									},
								},
								Rx1Frequency: 868100000,
							},
						},
					},
				},
			} {
				t.Run(tc.Name, func(t *testing.T) {
					a := assertions.New(t)

					_, err := gs.ScheduleDownlink(ctx, tc.Message)
					if err != nil {
						if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
							t.Fatalf("Unexpected error: %v", err)
						}
						if tc.RxWindowDetailsAssertion != nil {
							a.So(err, should.HaveSameErrorDefinitionAs, gatewayserver.ErrSchedule)
							if !a.So(errors.Details(err), should.HaveLength, 1) {
								t.FailNow()
							}
							details := errors.Details(err)[0].(*ttnpb.ScheduleDownlinkErrorDetails) //nolint:revive
							if !a.So(details, should.NotBeNil) || !a.So(details.PathErrors, should.HaveLength, 1) {
								t.FailNow()
							}
							errSchedulePathCause := errors.Cause(ttnpb.ErrorDetailsFromProto(details.PathErrors[0]))
							a.So(errors.IsAborted(errSchedulePathCause), should.BeTrue)
							for i, assert := range tc.RxWindowDetailsAssertion {
								if !a.So(errors.Details(errSchedulePathCause), should.HaveLength, 1) {
									t.FailNow()
								}
								errSchedulePathCauseDetails := errors.Details(errSchedulePathCause)[0].(*ttnpb.ScheduleDownlinkErrorDetails) //nolint:revive,lll
								if !a.So(errSchedulePathCauseDetails, should.NotBeNil) {
									t.FailNow()
								}
								if i >= len(errSchedulePathCauseDetails.PathErrors) {
									t.Fatalf("Expected error in Rx window %d", i+1)
								}
								errRxWindow := ttnpb.ErrorDetailsFromProto(errSchedulePathCauseDetails.PathErrors[i])
								if !a.So(assert(errRxWindow), should.BeTrue) {
									t.Fatalf("Unexpected Rx window %d error: %v", i+1, errRxWindow)
								}
							}
						}
						return
					} else if tc.ErrorAssertion != nil {
						t.Fatalf("Expected error")
					}
					downlinkCount++

					select {
					case msg := <-downCh:
						settings := msg.DownlinkMessage.GetScheduled()
						a.So(settings, should.NotBeNil)
					case <-time.After(timeout):
						t.Fatal("Expected downlink timeout")
					}

					time.Sleep(2 * timeout)

					conn, ok := gs.GetConnection(ctx, ids)
					a.So(ok, should.BeTrue)

					stats, paths := conn.Stats()
					a.So(stats, should.NotBeNil)
					a.So(paths, should.NotBeEmpty)

					stats, err = statsClient.GetGatewayConnectionStats(statsCtx, ids)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					a.So(stats.DownlinkCount, should.Equal, downlinkCount)
				})
			}
		})

		cancel()
		wg.Wait()
		if !errors.IsCanceled(linkErr) {
			t.Fatalf("Expected context canceled, but have %v", linkErr)
		}

		// Wait for disconnection to be processed.
		time.Sleep(2 * gsConfig.ConnectionStatsDisconnectTTL)

		// After canceling the context and awaiting the link, the connection should be gone.
		t.Run("Disconnected", func(t *testing.T) {
			_, err := statsClient.GetGatewayConnectionStats(statsCtx, ids)
			if !a.So(errors.IsNotFound(err), should.BeTrue) {
				t.Fatalf("Expected gateway to be disconnected, but it's not")
			}
		})
	})

	t.Run("Shutdown", func(t *testing.T) {
		ids := &ttnpb.GatewayIdentifiers{
			GatewayId: registeredGatewayID,
			Eui:       registeredGatewayEUI.Bytes(),
		}

		md := rpcmetadata.MD{
			ID:            ids.GatewayId,
			AuthType:      "Bearer",
			AuthValue:     registeredGatewayKey,
			AllowInsecure: true,
		}
		_, err = ttnpb.NewGtwGsClient(gs.LoopbackConn()).LinkGateway(ctx, grpc.PerRPCCredentials(md))
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		gs.Close()
		time.Sleep(2 * gsConfig.ConnectionStatsTTL)

		_, err = statsRegistry.Get(ctx, ids)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}
