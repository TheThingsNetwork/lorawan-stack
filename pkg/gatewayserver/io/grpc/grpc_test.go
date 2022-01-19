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
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/grpc"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var (
	registeredGatewayID  = ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
	registeredGatewayUID = unique.ID(test.Context(), registeredGatewayID)
	registeredGatewayKey = "test-key"

	timeout = (1 << 4) * test.Delay
)

func TestAuthentication(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayKey)

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
	gs := mock.NewServer(c)
	srv := New(gs)
	c.RegisterGRPC(&mockRegisterer{ctx, srv})
	componenttest.StartComponent(t, c)
	defer c.Close()

	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	client := ttnpb.NewGtwGsClient(c.LoopbackConn())

	for _, tc := range []struct {
		ID  ttnpb.GatewayIdentifiers
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
			ID:  ttnpb.GatewayIdentifiers{GatewayId: "invalid-gateway"},
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  ttnpb.GatewayIdentifiers{Eui: &eui},
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

type erroredGatewayDown struct {
	*ttnpb.GatewayDown
	error
}

func TestTraffic(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayKey)

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
	gs := mock.NewServer(c)
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

	upCh := make(chan *ttnpb.GatewayUp, 10)
	downCh := make(chan erroredGatewayDown, 10)

	stream, err := client.LinkGateway(ctx, creds)
	if err != nil {
		t.Fatalf("Failed to link gateway: %v", err)
	}
	go func() {
		for up := range upCh {
			if err := stream.Send(up); err != nil {
				panic(err)
			}
		}
	}()
	go func() {
		for ctx.Err() == nil {
			down, err := stream.Recv()
			downCh <- erroredGatewayDown{down, err}
		}
	}()

	var conn *io.Connection
	select {
	case conn = <-gs.Connections():
	case <-time.After(timeout):
		t.Fatal("Connection timeout")
	}

	t.Run("Upstream", func(t *testing.T) {
		for _, tc := range []*ttnpb.GatewayUp{
			{},
			{
				GatewayStatus: &ttnpb.GatewayStatus{
					Ip:   []string{"1.1.1.1"},
					Time: ttnpb.ProtoTimePtr(time.Now()),
				},
			},
			{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x01},
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &registeredGatewayID,
							},
						},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868500000,
							Timestamp: 42,
						},
					},
				},
			},
			{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x02},
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &registeredGatewayID,
							},
						},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868500000,
							Timestamp: 42,
						},
					},
				},
				GatewayStatus: &ttnpb.GatewayStatus{
					Ip:   []string{"2.2.2.2"},
					Time: ttnpb.ProtoTimePtr(time.Now()),
				},
			},
			{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x03},
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &registeredGatewayID,
							},
						},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868500000,
							Timestamp: 42,
						},
					},
					{
						RawPayload: []byte{0x04},
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &registeredGatewayID,
							},
						},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868500000,
							Timestamp: 42,
						},
					},
					{
						RawPayload: []byte{0x05},
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &registeredGatewayID,
							},
						},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868500000,
							Timestamp: 42,
						},
					},
				},
				GatewayStatus: &ttnpb.GatewayStatus{
					Ip:   []string{"3.3.3.3"},
					Time: ttnpb.ProtoTimePtr(time.Now()),
				},
			},
			{
				TxAcknowledgment: &ttnpb.TxAcknowledgment{
					Result: ttnpb.TxAcknowledgment_SUCCESS,
				},
			},
		} {
			t.Run(fmt.Sprintf("%v/%v", len(tc.UplinkMessages), tc.GatewayStatus != nil), func(t *testing.T) {
				a := assertions.New(t)

				upCh <- tc

				var ups int
				var needStatus bool = tc.GatewayStatus != nil
				var needTxAck bool = tc.TxAcknowledgment != nil
				for ups != len(tc.UplinkMessages) || needStatus || needTxAck {
					select {
					case up := <-conn.Up():
						expected := tc.UplinkMessages[ups]
						up.Message.ReceivedAt = expected.ReceivedAt
						up.Message.RxMetadata[0].UplinkToken = expected.RxMetadata[0].UplinkToken
						a.So(up.Message, should.Resemble, expected)
						ups++
					case status := <-conn.Status():
						a.So(needStatus, should.BeTrue)
						a.So(status, should.Resemble, tc.GatewayStatus)
						needStatus = false
					case ack := <-conn.TxAck():
						a.So(needTxAck, should.BeTrue)
						a.So(ack, should.Resemble, tc.TxAcknowledgment)
						needTxAck = false
					case <-time.After(timeout):
						t.Fatalf("Receive expected upstream timeout; ups = %v, needStatus = %t, needAck = %t", ups, needStatus, needTxAck)
					}
				}
			})
		}
		t.Run("DeduplicateByRSSI", func(t *testing.T) {
			a := assertions.New(t)
			upCh <- &ttnpb.GatewayUp{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x06},
						RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &registeredGatewayID, Rssi: -100}},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868500000,
							Timestamp: 42,
						},
					},
					{
						RawPayload: []byte{0x06},
						RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &registeredGatewayID, Rssi: -10}},
						Settings: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{
									Bandwidth:       125000,
									SpreadingFactor: 11,
								}},
							},
							EnableCrc: true,
							Frequency: 868700000,
							Timestamp: 42,
						},
					},
				},
			}
			select {
			case up := <-conn.Up():
				a.So(up.Message.RxMetadata[0].Rssi, should.Equal, -10)
				a.So(up.Message.RawPayload, should.Resemble, []byte{0x06})
				a.So(up.Message.Settings.Frequency, should.Equal, 868700000)
			case <-time.After(timeout):
				t.Fatalf("Receive unexpected upstream timeout")
			}
			select {
			case <-conn.Up():
				t.Fatalf("Received unexpected upstream message")
			case <-time.After(timeout):
			}
		})
	})

	t.Run("Downstream", func(t *testing.T) {
		for i, tc := range []struct {
			Path                *ttnpb.DownlinkPath
			Message             *ttnpb.DownlinkMessage
			ErrorAssertion      func(error) bool
			SendTxAck           bool
			TxAckErrorAssertion func(error) bool
		}{
			{
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
							100,
							100000,
							time.Unix(0, 100*1000),
							nil,
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:    ttnpb.Class_CLASS_A,
							Priority: ttnpb.TxSchedulePriority_NORMAL,
							Rx1Delay: ttnpb.RxDelay_RX_DELAY_1,
							Rx1DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{
									Lora: &ttnpb.LoRaDataRate{
										SpreadingFactor: 7,
										Bandwidth:       125000,
									},
								},
							},
							Rx1Frequency: 868100000,
							Rx2DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{
									Lora: &ttnpb.LoRaDataRate{
										SpreadingFactor: 0,
										Bandwidth:       125000,
									},
								},
							},
							Rx2Frequency:    869525000,
							FrequencyPlanId: test.EUFrequencyPlanID,
						},
					},
				},
				SendTxAck: true,
			},
			{
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
							100,
							100000,
							time.Unix(0, 100*1000),
							nil,
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
					Settings: &ttnpb.DownlinkMessage_Scheduled{
						Scheduled: &ttnpb.TxSettings{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{
									Lora: &ttnpb.LoRaDataRate{
										Bandwidth:       125000,
										SpreadingFactor: 7,
									},
								},
							},
							CodingRate: "4/5",
							Frequency:  869525000,
						},
					},
				},
				ErrorAssertion: errors.IsInvalidArgument, // Does not support scheduled downlink; the Gateway Server or gateway will take care of that.
			},
			{
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(
							&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &registeredGatewayID},
							100,
							100000,
							time.Unix(0, 100*1000),
							nil,
						),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x02},
				},
				ErrorAssertion: errors.IsInvalidArgument, // Tx request missing.
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)

				_, _, _, err := conn.ScheduleDown(tc.Path, tc.Message)
				if err != nil && (tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue)) {
					t.Fatalf("Unexpected error: %v", err)
				}
				var cids []string
				select {
				case down := <-downCh:
					if tc.ErrorAssertion == nil {
						cids = down.DownlinkMessage.CorrelationIds
						a.So(down.DownlinkMessage, should.Resemble, tc.Message)
					} else {
						t.Fatalf("Unexpected message: %v", down.DownlinkMessage)
					}
				case <-time.After(timeout):
					if tc.ErrorAssertion == nil {
						t.Fatal("Receive expected downlink timeout")
					}
				}

				if tc.ErrorAssertion != nil || !tc.SendTxAck {
					return
				}
				select {
				case upCh <- &ttnpb.GatewayUp{
					TxAcknowledgment: &ttnpb.TxAcknowledgment{
						CorrelationIds: cids,
						Result:         ttnpb.TxAcknowledgment_SUCCESS,
					},
				}:
				case <-time.After(timeout):
					if tc.TxAckErrorAssertion == nil {
						t.Fatal("Receive unexpected timeout while sending Tx acknowledgment")
					}
				}

				select {
				case ack := <-conn.TxAck():
					a.So(ack.DownlinkMessage, should.Resemble, tc.Message)
					a.So(ack.Result, should.Equal, ttnpb.TxAcknowledgment_SUCCESS)
				case <-time.After(timeout):
					if tc.TxAckErrorAssertion == nil {
						t.Fatal("Timeout waiting for Tx acknowledgment")
					}
				}
			})
		}
	})

	cancelCtx()
}

func TestConcentratorConfig(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayKey)

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
	gs := mock.NewServer(c)
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
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	is, isAddr := mock.NewIS(ctx)
	is.Add(ctx, registeredGatewayID, registeredGatewayKey)

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
	gs := mock.NewServer(c)
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

	info, err := client.GetMQTTConnectionInfo(ctx, &registeredGatewayID, creds)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(info, should.Resemble, &ttnpb.MQTTConnectionInfo{
		Username:         registeredGatewayUID,
		PublicAddress:    "example.com:1883",
		PublicTlsAddress: "example.com:8883",
	})

	infov2, err := client.GetMQTTV2ConnectionInfo(ctx, &registeredGatewayID, creds)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(infov2, should.Resemble, &ttnpb.MQTTConnectionInfo{
		Username:         registeredGatewayUID,
		PublicAddress:    "v2.example.com:1883",
		PublicTlsAddress: "v2.example.com:8883",
	})
}
