// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"gopkg.in/square/go-jose.v2"
)

var (
	timeout     = (1 << 4) * test.Delay
	testOptions []Option
)

func TestComponent(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := componenttest.NewComponent(t, &component.Config{})

	test.Must(New(c, &Config{}, testOptions...))
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_PACKET_BROKER_AGENT)
}

func TestForwarder(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			TLS: tlsconfig.Config{
				Client: tlsconfig.Client{
					RootCA: "testdata/serverca.pem",
				},
			},
		},
	})
	dp, addr := mustServePBDataPlane(ctx)

	gs := test.Must(mock.NewGatewayServer(c)).(*mock.GatewayServer)
	tokenKey := bytes.Repeat([]byte{0x42}, 16)
	tokenEncrypter := test.Must(jose.NewEncrypter(jose.A128GCM, jose.Recipient{
		Algorithm: jose.A128GCMKW,
		Key:       tokenKey,
	}, nil)).(jose.Encrypter)
	test.Must(New(c, &Config{
		DataPlaneAddress: fmt.Sprintf("localhost:%d", addr.(*net.TCPAddr).Port),
		NetID:            types.NetID{0x0, 0x0, 0x13},
		TenantID:         "test",
		ClusterID:        "test",
		TLS: tlsconfig.ClientAuth{
			Source:      "file",
			Certificate: "testdata/clientcert.pem",
			Key:         "testdata/clientkey.pem",
		},
		Forwarder: ForwarderConfig{
			Enable: true,
			WorkerPool: WorkerPoolConfig{
				Limit: 1,
			},
			TokenKey:       tokenKey,
			TokenEncrypter: tokenEncrypter,
		},
	}, testOptions...))
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_GATEWAY_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_PACKET_BROKER_AGENT)

	t.Run("Uplink", func(t *testing.T) {
		for i, tc := range []struct {
			GatewayMessage      *ttnpb.GatewayUplinkMessage
			RoutedUplinkMessage *packetbroker.RoutedUplinkMessage
		}{
			{
				GatewayMessage: &ttnpb.GatewayUplinkMessage{
					UplinkMessage: &ttnpb.UplinkMessage{
						RawPayload: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
						ReceivedAt: time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC),
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
								ChannelRSSI:        -42,
								RSSI:               -42,
								SNR:                10.5,
								Location: &ttnpb.Location{
									Latitude:  52.5,
									Longitude: 4.8,
									Altitude:  2,
								},
								UplinkToken: []byte("test-token"),
								Timestamp:   123456,
							},
						},
						Settings: ttnpb.TxSettings{
							DataRate: ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_LoRa{
									LoRa: &ttnpb.LoRaDataRate{
										SpreadingFactor: 7,
										Bandwidth:       125000,
									},
								},
							},
							CodingRate:    "4/5",
							DataRateIndex: 5,
							Frequency:     869525000,
						},
					},
					BandID: "EU_863_870",
				},
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					Message: &packetbroker.UplinkMessage{
						ForwarderReceiveTime: test.Must(pbtypes.TimestampProto(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC))).(*pbtypes.Timestamp),
						DataRateIndex:        5,
						Frequency:            869525000,
						GatewayMetadata: &packetbroker.UplinkMessage_GatewayMetadata{
							Teaser: &packetbroker.GatewayMetadataTeaser{
								Value: &packetbroker.GatewayMetadataTeaser_Terrestrial_{
									Terrestrial: &packetbroker.GatewayMetadataTeaser_Terrestrial{},
								},
							},
							SignalQuality: &packetbroker.UplinkMessage_GatewayMetadata_PlainSignalQuality{
								PlainSignalQuality: &packetbroker.GatewayMetadataSignalQuality{
									Value: &packetbroker.GatewayMetadataSignalQuality_Terrestrial_{
										Terrestrial: &packetbroker.GatewayMetadataSignalQuality_Terrestrial{
											Antennas: []*packetbroker.GatewayMetadataSignalQuality_Terrestrial_Antenna{
												{
													Index: 0,
													Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi:     -42,
														Snr:             10.5,
														FrequencyOffset: 0,
													},
												},
											},
										},
									},
								},
							},
							Localization: &packetbroker.UplinkMessage_GatewayMetadata_PlainLocalization{
								PlainLocalization: &packetbroker.GatewayMetadataLocalization{
									Value: &packetbroker.GatewayMetadataLocalization_Terrestrial_{
										Terrestrial: &packetbroker.GatewayMetadataLocalization_Terrestrial{
											Antennas: []*packetbroker.GatewayMetadataLocalization_Terrestrial_Antenna{
												{
													Index: 0,
													Location: &packetbroker.Location{
														Latitude:  52.5,
														Longitude: 4.8,
														Altitude:  2,
														Accuracy:  0,
													},
													SignalQuality: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi: -42,
														Snr:         10.5,
													},
												},
											},
										},
									},
								},
							},
						},
						GatewayRegion: packetbroker.Region_EU_863_870,
						PhyPayload: &packetbroker.UplinkMessage_PHYPayload{
							Teaser: &packetbroker.PHYPayloadTeaser{
								Hash: []byte{0x76, 0x9f, 0xce, 0x31, 0xe8, 0x1a, 0x90, 0xa1, 0x17, 0x07, 0x69, 0x18, 0x3b, 0x24, 0x0f, 0xd9, 0x8b, 0x7f, 0x38, 0xc7, 0x86, 0xb3, 0xd4, 0xe3, 0x8d, 0xae, 0xe1, 0x73, 0xe3, 0xa4, 0xcf, 0xbd},
								Payload: &packetbroker.PHYPayloadTeaser_Mac{
									Mac: &packetbroker.PHYPayloadTeaser_MACPayloadTeaser{
										FOpts:            true,
										DevAddr:          0x11223344,
										FPort:            1,
										FCnt:             1,
										FrmPayloadLength: 1,
									},
								},
							},
							Value: &packetbroker.UplinkMessage_PHYPayload_Plain{
								Plain: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
							},
						},
					},
				},
			},
			{
				GatewayMessage: &ttnpb.GatewayUplinkMessage{
					UplinkMessage: &ttnpb.UplinkMessage{
						RawPayload: []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42, 0x22, 0x11, 0x1, 0x2, 0x3, 0x4},
						ReceivedAt: time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC),
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
								ChannelRSSI:        4.2,
								RSSI:               4.2,
								SNR:                -5.5,
								UplinkToken:        []byte("test-token"),
								Timestamp:          123456,
							},
						},
						Settings: ttnpb.TxSettings{
							DataRate: ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_LoRa{
									LoRa: &ttnpb.LoRaDataRate{
										SpreadingFactor: 9,
										Bandwidth:       125000,
									},
								},
							},
							CodingRate:    "4/5",
							DataRateIndex: 3,
							Frequency:     868300000,
						},
					},
					BandID: "EU_863_870",
				},
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					Message: &packetbroker.UplinkMessage{
						ForwarderReceiveTime: test.Must(pbtypes.TimestampProto(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC))).(*pbtypes.Timestamp),
						DataRateIndex:        3,
						Frequency:            868300000,
						GatewayMetadata: &packetbroker.UplinkMessage_GatewayMetadata{
							Teaser: &packetbroker.GatewayMetadataTeaser{
								Value: &packetbroker.GatewayMetadataTeaser_Terrestrial_{
									Terrestrial: &packetbroker.GatewayMetadataTeaser_Terrestrial{},
								},
							},
							SignalQuality: &packetbroker.UplinkMessage_GatewayMetadata_PlainSignalQuality{
								PlainSignalQuality: &packetbroker.GatewayMetadataSignalQuality{
									Value: &packetbroker.GatewayMetadataSignalQuality_Terrestrial_{
										Terrestrial: &packetbroker.GatewayMetadataSignalQuality_Terrestrial{
											Antennas: []*packetbroker.GatewayMetadataSignalQuality_Terrestrial_Antenna{
												{
													Index: 0,
													Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi:     4.2,
														Snr:             -5.5,
														FrequencyOffset: 0,
													},
												},
											},
										},
									},
								},
							},
						},
						GatewayRegion: packetbroker.Region_EU_863_870,
						PhyPayload: &packetbroker.UplinkMessage_PHYPayload{
							Teaser: &packetbroker.PHYPayloadTeaser{
								Hash: []byte{0xce, 0xb5, 0x2a, 0x44, 0x27, 0xb9, 0x4d, 0x8a, 0xff, 0x4c, 0x6d, 0x20, 0xf5, 0x7d, 0x81, 0x66, 0x62, 0x9e, 0x6a, 0x26, 0xe6, 0x4c, 0x5f, 0x77, 0x2f, 0x70, 0xa7, 0xac, 0x34, 0x6a, 0x38, 0x81},
								Payload: &packetbroker.PHYPayloadTeaser_JoinRequest{
									JoinRequest: &packetbroker.PHYPayloadTeaser_JoinRequestTeaser{
										JoinEui:  0x42FFFFFFFFFFFFFF,
										DevEui:   0x4242FFFFFFFFFFFF,
										DevNonce: 0x1122,
									},
								},
							},
							Value: &packetbroker.UplinkMessage_PHYPayload_Plain{
								Plain: []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42, 0x22, 0x11, 0x1, 0x2, 0x3, 0x4},
							},
						},
					},
				},
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)

				err := gs.Publish(ctx, tc.GatewayMessage)
				a.So(err, should.BeNil)

				var pbMsg *packetbroker.RoutedUplinkMessage
				select {
				case pbMsg = <-dp.ForwarderUp:
				case <-time.After(timeout):
					t.Fatal("Expected uplink message from Forwarder")
				}

				pbMsg.Message.GatewayUplinkToken = nil // JWE, tested by TestWrapGatewayUplinkToken
				a.So(pbMsg, should.Resemble, tc.RoutedUplinkMessage)
			})
		}
	})

	t.Run("Downlink", func(t *testing.T) {
		a := assertions.New(t)

		token := test.Must(json.Marshal(GatewayUplinkToken{
			GatewayID: "test-gateway",
			Token:     []byte{0x1, 0x2, 0x3, 0x4},
		})).([]byte)
		tokenObj := test.Must(tokenEncrypter.Encrypt(token)).(*jose.JSONWebEncryption)
		tokenCompact := test.Must(tokenObj.CompactSerialize()).(string)

		dp.ForwarderDown <- &packetbroker.RoutedDownlinkMessage{
			ForwarderNetId:      0x000013,
			ForwarderTenantId:   "test",
			ForwarderId:         "test",
			HomeNetworkNetId:    0x000042,
			HomeNetworkTenantId: "test",
			Id:                  "test",
			Message: &packetbroker.DownlinkMessage{
				PhyPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
				Class:      packetbroker.DownlinkMessageClass_CLASS_A,
				Priority:   packetbroker.DownlinkMessagePriority_NORMAL,
				Rx1: &packetbroker.DownlinkMessage_RXSettings{
					Frequency:     868100000,
					DataRateIndex: 5,
				},
				Rx2: &packetbroker.DownlinkMessage_RXSettings{
					Frequency:     869525000,
					DataRateIndex: 0,
				},
				Rx1Delay:           pbtypes.DurationProto(5 * time.Second),
				GatewayUplinkToken: []byte(tokenCompact),
			},
		}

		var gtwMsg *ttnpb.DownlinkMessage
		select {
		case gtwMsg = <-gs.Downlink:
		case <-time.After(timeout):
			t.Fatal("Expected downlink message from Forwarder")
		}
		a.So(gtwMsg, should.Resemble, &ttnpb.DownlinkMessage{
			RawPayload:     []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
			CorrelationIDs: gtwMsg.CorrelationIDs,
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: &ttnpb.TxRequest{
					Class: ttnpb.CLASS_A,
					DownlinkPaths: []*ttnpb.DownlinkPath{
						{
							Path: &ttnpb.DownlinkPath_UplinkToken{
								UplinkToken: []byte{0x1, 0x2, 0x3, 0x4},
							},
						},
					},
					Priority:         ttnpb.TxSchedulePriority_NORMAL,
					Rx1DataRateIndex: 5,
					Rx1Frequency:     868100000,
					Rx1Delay:         ttnpb.RX_DELAY_5,
					Rx2DataRateIndex: 0,
					Rx2Frequency:     869525000,
				},
			},
		})
	})
}

func TestHomeNetwork(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			TLS: tlsconfig.Config{
				Client: tlsconfig.Client{
					RootCA: "testdata/serverca.pem",
				},
			},
		},
	})
	dp, addr := mustServePBDataPlane(ctx)

	ns := test.Must(mock.NewNetworkServer(c)).(*mock.NetworkServer)
	test.Must(New(c, &Config{
		DataPlaneAddress: fmt.Sprintf("localhost:%d", addr.(*net.TCPAddr).Port),
		NetID:            types.NetID{0x0, 0x0, 0x13},
		TenantID:         "test",
		ClusterID:        "test",
		TLS: tlsconfig.ClientAuth{
			Source:      "file",
			Certificate: "testdata/clientcert.pem",
			Key:         "testdata/clientkey.pem",
		},
		HomeNetwork: HomeNetworkConfig{
			Enable: true,
			WorkerPool: WorkerPoolConfig{
				Limit: 1,
			},
		},
	}, testOptions...))
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_PACKET_BROKER_AGENT)

	t.Run("Uplink", func(t *testing.T) {
		for i, tc := range []struct {
			RoutedUplinkMessage *packetbroker.RoutedUplinkMessage
			UplinkMessage       *ttnpb.UplinkMessage
		}{
			{
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					ForwarderNetId:      0x000042,
					ForwarderId:         "test",
					ForwarderTenantId:   "test",
					HomeNetworkNetId:    0x000013,
					HomeNetworkTenantId: "test",
					Id:                  "test",
					Message: &packetbroker.UplinkMessage{
						DataRateIndex:        5,
						ForwarderReceiveTime: test.Must(pbtypes.TimestampProto(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC))).(*pbtypes.Timestamp),
						Frequency:            869525000,
						GatewayMetadata: &packetbroker.UplinkMessage_GatewayMetadata{
							Teaser: &packetbroker.GatewayMetadataTeaser{
								Value: &packetbroker.GatewayMetadataTeaser_Terrestrial_{
									Terrestrial: &packetbroker.GatewayMetadataTeaser_Terrestrial{},
								},
							},
							SignalQuality: &packetbroker.UplinkMessage_GatewayMetadata_PlainSignalQuality{
								PlainSignalQuality: &packetbroker.GatewayMetadataSignalQuality{
									Value: &packetbroker.GatewayMetadataSignalQuality_Terrestrial_{
										Terrestrial: &packetbroker.GatewayMetadataSignalQuality_Terrestrial{
											Antennas: []*packetbroker.GatewayMetadataSignalQuality_Terrestrial_Antenna{
												{
													Index: 0,
													Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi:     -42,
														Snr:             10.5,
														FrequencyOffset: 0,
													},
												},
											},
										},
									},
								},
							},
							Localization: &packetbroker.UplinkMessage_GatewayMetadata_PlainLocalization{
								PlainLocalization: &packetbroker.GatewayMetadataLocalization{
									Value: &packetbroker.GatewayMetadataLocalization_Terrestrial_{
										Terrestrial: &packetbroker.GatewayMetadataLocalization_Terrestrial{
											Antennas: []*packetbroker.GatewayMetadataLocalization_Terrestrial_Antenna{
												{
													Index: 0,
													Location: &packetbroker.Location{
														Latitude:  52.5,
														Longitude: 4.8,
														Altitude:  2,
														Accuracy:  0,
													},
													SignalQuality: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi: -42,
														Snr:         10.5,
													},
												},
											},
										},
									},
								},
							},
						},
						GatewayRegion:      packetbroker.Region_EU_863_870,
						GatewayUplinkToken: []byte("test-token"),
						PhyPayload: &packetbroker.UplinkMessage_PHYPayload{
							Teaser: &packetbroker.PHYPayloadTeaser{
								Hash: []byte{0x76, 0x9f, 0xce, 0x31, 0xe8, 0x1a, 0x90, 0xa1, 0x17, 0x07, 0x69, 0x18, 0x3b, 0x24, 0x0f, 0xd9, 0x8b, 0x7f, 0x38, 0xc7, 0x86, 0xb3, 0xd4, 0xe3, 0x8d, 0xae, 0xe1, 0x73, 0xe3, 0xa4, 0xcf, 0xbd},
								Payload: &packetbroker.PHYPayloadTeaser_Mac{
									Mac: &packetbroker.PHYPayloadTeaser_MACPayloadTeaser{
										FOpts:            true,
										DevAddr:          0x11223344,
										FPort:            1,
										FCnt:             1,
										FrmPayloadLength: 1,
									},
								},
							},
							Value: &packetbroker.UplinkMessage_PHYPayload_Plain{
								Plain: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
							},
						},
					},
				},
				UplinkMessage: &ttnpb.UplinkMessage{
					RawPayload: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
					RxMetadata: []*ttnpb.RxMetadata{
						{
							GatewayIdentifiers: cluster.PacketBrokerGatewayID,
							PacketBroker: &ttnpb.PacketBrokerMetadata{
								MessageID:           "test",
								ForwarderNetID:      [3]byte{0x0, 0x0, 0x42},
								ForwarderTenantID:   "test",
								ForwarderID:         "test",
								HomeNetworkNetID:    [3]byte{0x0, 0x0, 0x13},
								HomeNetworkTenantID: "test",
							},
							ChannelRSSI: -42,
							RSSI:        -42,
							SNR:         10.5,
							Location: &ttnpb.Location{
								Latitude:  52.5,
								Longitude: 4.8,
								Altitude:  2,
							},
							UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &AgentUplinkToken{
								ForwarderNetID:    [3]byte{0x0, 0x0, 0x42},
								ForwarderID:       "test",
								ForwarderTenantID: "test",
							})).([]byte),
						},
					},
					Settings: ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_LoRa{
								LoRa: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						DataRateIndex: 5,
						Frequency:     869525000,
					},
				},
			},
			{
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					ForwarderNetId:      0x000042,
					ForwarderId:         "test",
					ForwarderTenantId:   "test",
					HomeNetworkNetId:    0x000013,
					HomeNetworkTenantId: "test",
					Id:                  "test",
					Message: &packetbroker.UplinkMessage{
						DataRateIndex:        3,
						ForwarderReceiveTime: test.Must(pbtypes.TimestampProto(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC))).(*pbtypes.Timestamp),
						Frequency:            869525000,
						GatewayMetadata: &packetbroker.UplinkMessage_GatewayMetadata{
							Teaser: &packetbroker.GatewayMetadataTeaser{
								Value: &packetbroker.GatewayMetadataTeaser_Terrestrial_{
									Terrestrial: &packetbroker.GatewayMetadataTeaser_Terrestrial{},
								},
							},
							SignalQuality: &packetbroker.UplinkMessage_GatewayMetadata_PlainSignalQuality{
								PlainSignalQuality: &packetbroker.GatewayMetadataSignalQuality{
									Value: &packetbroker.GatewayMetadataSignalQuality_Terrestrial_{
										Terrestrial: &packetbroker.GatewayMetadataSignalQuality_Terrestrial{
											Antennas: []*packetbroker.GatewayMetadataSignalQuality_Terrestrial_Antenna{
												{
													Index: 0,
													Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi:     4.2,
														Snr:             -5.5,
														FrequencyOffset: 0,
													},
												},
											},
										},
									},
								},
							},
						},
						GatewayRegion:      packetbroker.Region_EU_863_870,
						GatewayUplinkToken: []byte("test-token"),
						PhyPayload: &packetbroker.UplinkMessage_PHYPayload{
							Teaser: &packetbroker.PHYPayloadTeaser{
								Hash: []byte{0x76, 0x9f, 0xce, 0x31, 0xe8, 0x1a, 0x90, 0xa1, 0x17, 0x07, 0x69, 0x18, 0x3b, 0x24, 0x0f, 0xd9, 0x8b, 0x7f, 0x38, 0xc7, 0x86, 0xb3, 0xd4, 0xe3, 0x8d, 0xae, 0xe1, 0x73, 0xe3, 0xa4, 0xcf, 0xbd},
								Payload: &packetbroker.PHYPayloadTeaser_Mac{
									Mac: &packetbroker.PHYPayloadTeaser_MACPayloadTeaser{
										FOpts:            true,
										DevAddr:          0x11223344,
										FPort:            1,
										FCnt:             1,
										FrmPayloadLength: 1,
									},
								},
							},
							Value: &packetbroker.UplinkMessage_PHYPayload_Plain{
								Plain: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
							},
						},
					},
				},
				UplinkMessage: &ttnpb.UplinkMessage{
					RawPayload: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
					RxMetadata: []*ttnpb.RxMetadata{
						{
							GatewayIdentifiers: cluster.PacketBrokerGatewayID,
							PacketBroker: &ttnpb.PacketBrokerMetadata{
								MessageID:           "test",
								ForwarderNetID:      [3]byte{0x0, 0x0, 0x42},
								ForwarderTenantID:   "test",
								ForwarderID:         "test",
								HomeNetworkNetID:    [3]byte{0x0, 0x0, 0x13},
								HomeNetworkTenantID: "test",
							},
							ChannelRSSI: 4.2,
							RSSI:        4.2,
							SNR:         -5.5,
							UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &AgentUplinkToken{
								ForwarderNetID:    [3]byte{0x0, 0x0, 0x42},
								ForwarderID:       "test",
								ForwarderTenantID: "test",
							})).([]byte),
						},
					},
					Settings: ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_LoRa{
								LoRa: &ttnpb.LoRaDataRate{
									SpreadingFactor: 9,
									Bandwidth:       125000,
								},
							},
						},
						DataRateIndex: 3,
						Frequency:     869525000,
					},
				},
			},
		} {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				a := assertions.New(t)

				dp.HomeNetworkUp <- tc.RoutedUplinkMessage

				before := time.Now()
				var nsMsg *ttnpb.UplinkMessage
				select {
				case nsMsg = <-ns.Uplink:
				case <-time.After(timeout):
					t.Fatal("Expected uplink message from Forwarder")
				}

				a.So(nsMsg.CorrelationIDs, should.HaveLength, 2)
				nsMsg.CorrelationIDs = nil
				a.So(nsMsg.ReceivedAt, should.HappenBetween, before, time.Now()) // Packet Broker Agent sets local time on receive.
				nsMsg.ReceivedAt = time.Time{}

				a.So(nsMsg, should.Resemble, tc.UplinkMessage)
			})
		}
	})

	t.Run("Downlink", func(t *testing.T) {
		a := assertions.New(t)

		nsMsg := &ttnpb.DownlinkMessage{
			RawPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: &ttnpb.TxRequest{
					Class: ttnpb.CLASS_A,
					DownlinkPaths: []*ttnpb.DownlinkPath{
						{
							Path: &ttnpb.DownlinkPath_UplinkToken{
								UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &AgentUplinkToken{
									ForwarderNetID:    [3]byte{0x0, 0x0, 0x42},
									ForwarderID:       "test",
									ForwarderTenantID: "test",
								})).([]byte),
							},
						},
					},
					Priority:         ttnpb.TxSchedulePriority_NORMAL,
					Rx1DataRateIndex: 5,
					Rx1Frequency:     868100000,
					Rx1Delay:         ttnpb.RX_DELAY_5,
					Rx2DataRateIndex: 0,
					Rx2Frequency:     869525000,
				},
			},
		}
		err := ns.Publish(ctx, nsMsg)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		var pbMsg *packetbroker.RoutedDownlinkMessage
		select {
		case pbMsg = <-dp.HomeNetworkDown:
		case <-time.After(timeout):
			t.Fatal("Expected downlink message from Forwarder")
		}

		a.So(pbMsg, should.Resemble, &packetbroker.RoutedDownlinkMessage{
			ForwarderNetId:    0x000042,
			ForwarderId:       "test",
			ForwarderTenantId: "test",
			Message: &packetbroker.DownlinkMessage{
				PhyPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
				Class:      packetbroker.DownlinkMessageClass_CLASS_A,
				Priority:   packetbroker.DownlinkMessagePriority_NORMAL,
				Rx1: &packetbroker.DownlinkMessage_RXSettings{
					Frequency:     868100000,
					DataRateIndex: 5,
				},
				Rx2: &packetbroker.DownlinkMessage_RXSettings{
					Frequency:     869525000,
					DataRateIndex: 0,
				},
				Rx1Delay:           pbtypes.DurationProto(5 * time.Second),
				GatewayUplinkToken: []byte(`test-token`),
			},
		})
	})
}
