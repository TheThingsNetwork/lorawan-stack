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
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/iotest"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mqtt"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/proto"
)

func TestAuthentication(t *testing.T) {
	var (
		registeredGatewayID  = &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
		registeredGatewayUID = unique.ID(test.Context(), registeredGatewayID)
		registeredGatewayKey = "test-key"
		timeout              = (1 << 4) * test.Delay
	)

	a, ctx := test.New(t)
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, "Bearer", registeredGatewayKey, testGtw, testRights...)

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
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	gs := mock.NewServer(c, is)
	lis, err := net.Listen("tcp", ":0")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	go Serve(ctx, gs, lis, NewProtobuf(ctx), "tcp")

	for _, tc := range []struct {
		UID string
		Key string
		OK  bool
	}{
		{
			UID: registeredGatewayUID,
			Key: registeredGatewayKey,
			OK:  true,
		},
		{
			UID: registeredGatewayUID,
			Key: "invalid-key",
			OK:  false,
		},
		{
			UID: "invalid-gateway",
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
			if tc.OK {
				if !token.WaitTimeout(timeout) {
					t.Fatal("Connection timeout")
				}
				if !a.So(token.Error(), should.BeNil) {
					t.FailNow()
				}
			} else if token.Wait() && !a.So(token.Error(), should.NotBeNil) {
				t.FailNow()
			}
			client.Disconnect(uint(timeout / time.Millisecond))
		})
	}
}

func TestFrontend(t *testing.T) {
	t.Parallel()
	timeout := (1 << 4) * test.Delay
	iotest.Frontend(t, iotest.FrontendConfig{
		DetectsInvalidMessages: false,
		SupportsStatus:         true,
		DetectsDisconnect:      true,
		IsAuthenticated:        true,
		DeduplicatesUplinks:    false,
		CustomGatewayServerConfig: func(config *gatewayserver.Config) {
			config.MQTT.Listen = ":1882"
		},
		Link: func(
			ctx context.Context,
			t *testing.T,
			gs *gatewayserver.GatewayServer,
			ids *ttnpb.GatewayIdentifiers,
			key string,
			upCh <-chan *ttnpb.GatewayUp,
			downCh chan<- *ttnpb.GatewayDown,
		) error {
			if ids.GatewayId == "" {
				t.SkipNow()
			}
			ctx, cancel := errorcontext.New(ctx)
			clientOpts := mqtt.NewClientOptions()
			clientOpts.AddBroker("tcp://0.0.0.0:1882")
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
			// Write upstream.
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case up := <-upCh:
						for _, msg := range up.UplinkMessages {
							buf, err := proto.Marshal(msg)
							if err != nil {
								cancel(err)
								return
							}
							if token := client.Publish(fmt.Sprintf("v3/%v/up", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
								cancel(token.Error())
								return
							}
						}
						if up.GatewayStatus != nil {
							buf, err := proto.Marshal(up.GatewayStatus)
							if err != nil {
								cancel(err)
								return
							}
							if token := client.Publish(fmt.Sprintf("v3/%v/status", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
								cancel(token.Error())
								return
							}
						}
						if up.TxAcknowledgment != nil {
							buf, err := proto.Marshal(up.TxAcknowledgment)
							if err != nil {
								cancel(err)
								return
							}
							if token := client.Publish(fmt.Sprintf("v3/%v/down/ack", unique.ID(ctx, ids)), 1, false, buf); token.Wait() && token.Error() != nil {
								cancel(token.Error())
								return
							}
						}
					}
				}
			}()
			// Read downstream.
			token := client.Subscribe(fmt.Sprintf("v3/%v/down", unique.ID(ctx, ids)), 1, func(_ mqtt.Client, raw mqtt.Message) {
				var msg ttnpb.GatewayDown
				if err := proto.Unmarshal(raw.Payload(), &msg); err != nil {
					cancel(err)
					return
				}
				downCh <- &msg
			})
			if token.Wait() && token.Error() != nil {
				return token.Error()
			}
			<-ctx.Done()
			return ctx.Err()
		},
	})
}
