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
	"fmt"
	"os"
	"testing"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	ttnpbv2 "go.thethings.network/lorawan-stack-legacy/v2/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	gsredis "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/redis"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

func mustHavePeer(ctx context.Context, t *testing.T, c *component.Component, role ttnpb.ClusterRole) {
	t.Helper()
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond)
		if _, err := c.GetPeer(ctx, role, nil); err == nil {
			return
		}
	}
	t.Fatal("Could not connect to peer")
}

func TestUpdateVersionInfo(t *testing.T) { //nolint:paralleltest
	var (
		timeout              = (1 << 4) * test.Delay
		testRights           = []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK, ttnpb.Right_RIGHT_GATEWAY_STATUS_READ}
		registeredGatewayID  = "eui-aaee000000000000"
		registeredGatewayKey = "secret"
		registeredGatewayEUI = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	)

	a, ctx := test.New(t)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	ids := &ttnpb.GatewayIdentifiers{
		GatewayId: registeredGatewayID,
		Eui:       registeredGatewayEUI.Bytes(),
	}
	gtw := mockis.DefaultGateway(ids, true, true)
	is.GatewayRegistry().Add(ctx, ids, "Bearer", registeredGatewayKey, gtw, testRights...)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	defer c.Close()

	gsConfig := &gatewayserver.Config{
		FetchGatewayInterval:   (1 << 5) * test.Delay,
		FetchGatewayJitter:     0.1,
		UpdateVersionInfoDelay: test.Delay,
		MQTTV2: config.MQTT{
			Listen: ":1881",
		},
	}

	er := gatewayserver.NewIS(c)
	gs, err := gatewayserver.New(c, gsConfig,
		gatewayserver.WithRegistry(er),
	)
	if !a.So(err, should.BeNil) {
		t.Fatalf("Failed to setup server :%v", err)
	}
	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER)
	a.So(err, should.BeNil)

	componenttest.StartComponent(t, c)
	time.Sleep(timeout) // Wait for component to start.

	mustHavePeer(ctx, t, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	gtwIDs := &ttnpb.GatewayIdentifiers{
		GatewayId: registeredGatewayID,
		Eui:       registeredGatewayEUI.Bytes(),
	}

	mockGtw := mockis.DefaultGateway(gtwIDs, true, true)
	is.GatewayRegistry().Add(ctx, gtwIDs, "Bearer", registeredGatewayKey, mockGtw, testRights...)
	time.Sleep(timeout) // Wait for setup to be completed.

	link := func(ctx context.Context, ids *ttnpb.GatewayIdentifiers, key string, statCh <-chan *ttnpbv2.StatusMessage) error {
		ctx, cancel := errorcontext.New(ctx)
		clientOpts := mqtt.NewClientOptions()
		clientOpts.AddBroker("tcp://0.0.0.0:1881")
		clientOpts.SetUsername(unique.ID(ctx, ids))
		clientOpts.SetPassword(key)
		clientOpts.SetAutoReconnect(false)
		clientOpts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
			cancel(err)
		})
		client := mqtt.NewClient(clientOpts)
		if token := client.Connect(); !token.WaitTimeout(timeout) {
			return context.DeadlineExceeded
		} else if err := token.Error(); err != nil {
			return err
		}
		defer client.Disconnect(uint(timeout / time.Millisecond))
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case stat := <-statCh:
					buf, err := proto.Marshal(stat)
					if err != nil {
						cancel(err)
						return
					}
					if token := client.Publish(fmt.Sprintf("%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
						cancel(token.Error())
						return
					}
				}
			}
		}()
		<-ctx.Done()
		return ctx.Err()
	}

	statCh := make(chan *ttnpbv2.StatusMessage)
	go link(ctx, ids, registeredGatewayKey, statCh)

	for _, tc := range []struct {
		Name               string
		Stat               *ttnpbv2.StatusMessage
		ExpectedAttributes map[string]string
	}{
		{
			Name: "FirstStat",
			Stat: &ttnpbv2.StatusMessage{
				Platform: "The Things Gateway v1 - BL r9-12345678 (2006-01-02T15:04:05Z) - Firmware v1.2.3-12345678 (2006-01-02T15:04:05Z)",
			},
			ExpectedAttributes: map[string]string{
				"model":    "The Things Kickstarter Gateway v1",
				"firmware": "v1.2.3-12345678",
			},
		},
		{
			Name: "SubsequentStatNoUpdate",
			Stat: &ttnpbv2.StatusMessage{
				Platform: "The Things Gateway v1 - BL r9-12345678 (2006-01-02T15:04:05Z) - Firmware v2.0.0-00000000 (2006-01-02T15:04:05Z)",
			},
			ExpectedAttributes: map[string]string{
				"model":    "The Things Kickstarter Gateway v1",
				"firmware": "v1.2.3-12345678",
			},
		},
	} {
		t.Run(fmt.Sprintf("UpdateVersionInfo/%s", tc.Name), func(t *testing.T) {
			select {
			case statCh <- tc.Stat:
			case <-time.After(timeout):
				t.Fatalf("Failed to send status message to upstream channel")
			}
			time.Sleep(timeout)
			gtw, err := is.GatewayRegistry().Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: ids,
			})
			a.So(err, should.BeNil)
			a.So(gtw.Attributes, should.Resemble, tc.ExpectedAttributes)
		})
	}

	// Test Disconnection on delete.
	// Setup a stats client with independent context to query whether the gateway is connected and statistics on
	// upstream and downstream.
	statsCtx := metadata.AppendToOutgoingContext(test.Context(),
		"id", ids.GatewayId,
		"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
	)
	statsClient := ttnpb.NewGsClient(gs.LoopbackConn())

	stat, err := statsClient.GetGatewayConnectionStats(statsCtx, gtwIDs)
	a.So(err, should.BeNil)
	a.So(stat, should.NotBeNil)

	// Delete and wait for fetch interval.
	is.GatewayRegistry().Delete(ctx, gtwIDs)
	time.Sleep(2 * gsConfig.FetchGatewayInterval)

	stat, err = statsClient.GetGatewayConnectionStats(statsCtx, gtwIDs)
	a.So(errors.IsNotFound(err), should.BeTrue)
	a.So(stat, should.BeNil)

	gs.Close()
	time.Sleep(timeout)
}

func TestBatchGetStatus(t *testing.T) { // nolint:paralleltest
	var (
		timeout              = (1 << 4) * test.Delay
		testRights           = []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_LINK, ttnpb.Right_RIGHT_GATEWAY_STATUS_READ}
		registeredGatewayID  = "eui-aaee000000000000"
		registeredGatewayKey = "secret"
		registeredGatewayEUI = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	)

	a, ctx := test.New(t)

	for _, tc := range []struct { //nolint:paralleltest
		Name      string
		WithRedis bool
	}{
		{
			Name:      "Redis",
			WithRedis: true,
		},
		{
			Name: "NilRegistry",
		},
	} {
		t.Run(fmt.Sprintf("BatchGetStatus/%s", tc.Name), func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			is, isAddr, closeIS := mockis.New(ctx)
			defer closeIS()

			c := componenttest.NewComponent(t, &component.Config{
				ServiceBase: config.ServiceBase{
					GRPC: config.GRPC{
						Listen:                      ":0",
						AllowInsecureForCredentials: true,
					},
					Cluster: cluster.Config{
						IdentityServer: isAddr,
					},
					FrequencyPlans: config.FrequencyPlansConfig{
						ConfigSource: "static",
						Static:       test.StaticFrequencyPlans,
					},
				},
			})
			defer c.Close()

			gsConfig := &gatewayserver.Config{
				FetchGatewayInterval:   (1 << 5) * test.Delay,
				FetchGatewayJitter:     0.1,
				UpdateVersionInfoDelay: test.Delay,
				MQTTV2: config.MQTT{
					Listen: ":1881",
				},
			}

			if tc.WithRedis && os.Getenv("TEST_REDIS") == "1" {
				statsRedisClient, statsFlush := test.NewRedis(ctx, "gatewayserver_test")
				defer statsFlush()
				defer statsRedisClient.Close()
				registry := &gsredis.GatewayConnectionStatsRegistry{
					Redis:   statsRedisClient,
					LockTTL: timeout,
				}
				if err := registry.Init(ctx); err != nil {
					t.Fatalf("Failed to setup stats registry :%v", err)
				}
				gsConfig.Stats = registry
			}

			er := gatewayserver.NewIS(c)
			gs, err := gatewayserver.New(c, gsConfig,
				gatewayserver.WithRegistry(er),
			)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to setup server :%v", err)
			}
			roles := gs.Roles()
			a.So(len(roles), should.Equal, 1)
			a.So(roles[0], should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER)
			a.So(err, should.BeNil)

			componenttest.StartComponent(t, c)
			time.Sleep(timeout) // Wait for component to start.

			mustHavePeer(ctx, t, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

			linkFn := func(ctx context.Context, t *testing.T,
				ids *ttnpb.GatewayIdentifiers, key string, statCh <-chan *ttnpbv2.StatusMessage,
			) error {
				t.Helper()
				ctx, cancel := errorcontext.New(ctx)
				clientOpts := mqtt.NewClientOptions()
				clientOpts.AddBroker("tcp://0.0.0.0:1881")
				clientOpts.SetUsername(unique.ID(ctx, ids))
				clientOpts.SetPassword(key)
				clientOpts.SetAutoReconnect(false)
				clientOpts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
					cancel(err)
				})
				client := mqtt.NewClient(clientOpts)
				if token := client.Connect(); !token.WaitTimeout(timeout) {
					return context.DeadlineExceeded
				} else if err := token.Error(); err != nil {
					return err
				}
				defer client.Disconnect(uint(timeout / time.Millisecond))
				go func() {
					for {
						select {
						case <-ctx.Done():
							return
						case stat := <-statCh:
							buf, err := proto.Marshal(stat)
							if err != nil {
								cancel(err)
								return
							}
							if token := client.Publish(
								fmt.Sprintf("%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
								cancel(token.Error())
								return
							}
						}
					}
				}()
				<-ctx.Done()
				return ctx.Err()
			}

			gtwIDs1 := &ttnpb.GatewayIdentifiers{
				GatewayId: registeredGatewayID,
				Eui:       registeredGatewayEUI.Bytes(),
			}

			gtwIDs2 := &ttnpb.GatewayIdentifiers{
				GatewayId: "eui-aaee000000000001",
			}

			statsClient := ttnpb.NewGsClient(gs.LoopbackConn())
			statsCtx := metadata.AppendToOutgoingContext(test.Context(),
				"id", gtwIDs1.GatewayId,
				"authorization", fmt.Sprintf("Bearer %v", registeredGatewayKey),
			)

			request := &ttnpb.BatchGetGatewayConnectionStatsRequest{
				GatewayIds: []*ttnpb.GatewayIdentifiers{
					gtwIDs1,
					gtwIDs2,
				},
			}

			// Get Stats before creation.
			res, err := statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.NotBeNil)
			a.So(res, should.BeNil)

			mockGtw1 := mockis.DefaultGateway(gtwIDs1, true, true)
			is.GatewayRegistry().Add(ctx, gtwIDs1, "Bearer", registeredGatewayKey, mockGtw1, testRights...)
			time.Sleep(timeout) // Wait for setup to be completed.

			mockGtw2 := mockis.DefaultGateway(gtwIDs2, true, true)
			is.GatewayRegistry().Add(ctx, gtwIDs2, "Bearer", registeredGatewayKey, mockGtw2, testRights...)
			time.Sleep(timeout) // Wait for setup to be completed.

			// Invalid batch
			res, err = statsClient.BatchGetGatewayConnectionStats(
				statsCtx,
				&ttnpb.BatchGetGatewayConnectionStatsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtwIDs1,
						{
							GatewayId: "unknown",
						},
					},
				},
			)
			a.So(err, should.NotBeNil)
			a.So(res, should.BeNil)

			// Get Stats before connection.
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
			a.So(len(res.Entries), should.Equal, 0)

			// Connect first gateway.
			statCh1 := make(chan *ttnpbv2.StatusMessage)
			go func() {
				_ = linkFn(ctx, t, gtwIDs1, registeredGatewayKey, statCh1)
			}()
			time.Sleep(timeout) // Wait for connection to be completed.

			// Get Stats
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
			a.So(len(res.Entries), should.Equal, 1)
			a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)

			// Connect second gateway.
			ctxWithCancel, gtwConnCancel := context.WithCancel(ctx)
			statCh2 := make(chan *ttnpbv2.StatusMessage)
			go func() {
				_ = linkFn(ctxWithCancel, t, gtwIDs2, registeredGatewayKey, statCh2)
			}()
			time.Sleep(timeout) // Wait for connection to be completed.

			// Get Stats
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)
			a.So(len(res.Entries), should.Equal, 2)
			a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)
			a.So(res.Entries[gtwIDs2.GatewayId], should.NotBeNil)

			// Disconnect second gateway.
			gtwConnCancel()
			time.Sleep(timeout) // Wait for connection to be closed.

			cfg, err := gs.GetConfig(ctx)
			if !a.So(err, should.BeNil) {
				t.Fatalf("Failed to get Gateway Server configuration :%v", err)
			}
			res, err = statsClient.BatchGetGatewayConnectionStats(statsCtx, request)
			a.So(err, should.BeNil)
			a.So(res, should.NotBeNil)

			if cfg.Stats != nil {
				// Only stats entries in the registry are persisted until Redis TTL after disconnection.
				// These entries will have the `disconnected_at` field set.
				a.So(len(res.Entries), should.Equal, 2)
				a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)
				a.So(res.Entries[gtwIDs2.GatewayId], should.NotBeNil)
			} else {
				a.So(len(res.Entries), should.Equal, 1)
				a.So(res.Entries[gtwIDs1.GatewayId], should.NotBeNil)
			}
			// Close the Gateway Server.
			gs.Close()
			time.Sleep(timeout)
		})
	}
}
