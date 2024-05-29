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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/grpc"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/iotest"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestAuthentication(t *testing.T) {
	var (
		registeredGatewayID  = &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
		registeredGatewayKey = "test-key"
		timeout              = (1 << 4) * test.Delay
	)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayKey, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)
	srv := New(gs)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewGtwGsClient(c.LoopbackConn())

	for _, tc := range []struct {
		ID  *ttnpb.GatewayIdentifiers
		Key string
		OK  bool
	}{
		{
			ID:  registeredGatewayID,
			Key: registeredGatewayKey,
			OK:  true,
		},
		{
			ID:  registeredGatewayID,
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  &ttnpb.GatewayIdentifiers{GatewayId: "invalid-gateway"},
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  &ttnpb.GatewayIdentifiers{Eui: eui.Bytes()},
			Key: "invalid-key",
			OK:  false,
		},
	} {
		t.Run(fmt.Sprintf("%v:%v", tc.ID.GatewayId, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			ctx = rpcmetadata.MD{
				ID: tc.ID.GatewayId,
			}.ToOutgoingContext(ctx)
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
				_, err = client.LinkGateway(ctx, creds)
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

func TestFrontend(t *testing.T) {
	t.Parallel()
	iotest.Frontend(t, iotest.FrontendConfig{
		SupportsStatus:      true,
		IsAuthenticated:     true,
		DetectsDisconnect:   true,
		DeduplicatesUplinks: true,
		Link: func(
			ctx context.Context,
			t *testing.T,
			gs *gatewayserver.GatewayServer,
			ids *ttnpb.GatewayIdentifiers,
			key string,
			upCh <-chan *ttnpb.GatewayUp,
			downCh chan<- *ttnpb.GatewayDown,
		) error {
			md := rpcmetadata.MD{
				ID:            ids.GatewayId,
				AuthType:      "Bearer",
				AuthValue:     key,
				AllowInsecure: true,
			}
			client := ttnpb.NewGtwGsClient(gs.LoopbackConn())
			_, err := client.GetConcentratorConfig(ctx, ttnpb.Empty, grpc.PerRPCCredentials(md))
			if err != nil {
				return err
			}
			link, err := client.LinkGateway(ctx, grpc.PerRPCCredentials(md))
			if err != nil {
				return err
			}
			ctx, cancel := errorcontext.New(ctx)
			// Write upstream.
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case msg := <-upCh:
						if err := link.Send(msg); err != nil {
							cancel(err)
							return
						}
					}
				}
			}()
			// Read downstream.
			go func() {
				for {
					msg, err := link.Recv()
					if err != nil {
						cancel(err)
						return
					}
					downCh <- msg
				}
			}()
			<-ctx.Done()
			return ctx.Err()
		},
	})
}

func TestConcentratorConfig(t *testing.T) {
	var (
		registeredGatewayID  = &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
		registeredGatewayKey = "test-key"
	)

	a, ctx := test.New(t)
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayKey, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)
	srv := New(gs)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewGtwGsClient(c.LoopbackConn())

	ctx = rpcmetadata.MD{
		ID: registeredGatewayID.GatewayId,
	}.ToOutgoingContext(ctx)
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredGatewayKey,
		AllowInsecure: true,
	})

	_, err := client.GetConcentratorConfig(ctx, ttnpb.Empty, creds)
	a.So(err, should.BeNil)
}

type mockMQTTConfigProvider struct {
	config.MQTT
}

func (p mockMQTTConfigProvider) GetMQTTConfig(context.Context) (*config.MQTT, error) {
	return &p.MQTT, nil
}

func TestMQTTConfig(t *testing.T) {
	var (
		registeredGatewayID  = &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
		registeredGatewayUID = unique.ID(test.Context(), registeredGatewayID)
		registeredGatewayKey = "test-key"
	)

	a, ctx := test.New(t)
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	testGtw := mockis.DefaultGateway(registeredGatewayID, false, false)
	is.GatewayRegistry().Add(ctx, registeredGatewayID, registeredGatewayKey, testGtw, testRights...)

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
	gs := mock.NewServer(c, is)
	srv := New(gs,
		WithMQTTConfigProvider(&mockMQTTConfigProvider{
			MQTT: config.MQTT{
				PublicAddress:    "example.com:1883",
				PublicTLSAddress: "example.com:8883",
			},
		}),
		WithMQTTV2ConfigProvider(&mockMQTTConfigProvider{
			MQTT: config.MQTT{
				PublicAddress:    "v2.example.com:1883",
				PublicTLSAddress: "v2.example.com:8883",
			},
		}),
	)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewGtwGsClient(c.LoopbackConn())

	ctx = rpcmetadata.MD{
		ID: registeredGatewayID.GatewayId,
	}.ToOutgoingContext(ctx)
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredGatewayKey,
		AllowInsecure: true,
	})

	info, err := client.GetMQTTConnectionInfo(ctx, registeredGatewayID, creds)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(info, should.Resemble, &ttnpb.MQTTConnectionInfo{
		Username:         registeredGatewayUID,
		PublicAddress:    "example.com:1883",
		PublicTlsAddress: "example.com:8883",
	})

	infov2, err := client.GetMQTTV2ConnectionInfo(ctx, registeredGatewayID, creds)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(infov2, should.Resemble, &ttnpb.MQTTConnectionInfo{
		Username:         registeredGatewayUID,
		PublicAddress:    "v2.example.com:1883",
		PublicTlsAddress: "v2.example.com:8883",
	})
}
