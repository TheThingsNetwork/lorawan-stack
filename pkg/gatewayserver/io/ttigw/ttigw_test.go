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

package ttigw_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"net/http"
	"testing"

	"github.com/coder/websocket"
	lorav1 "go.thethings.industries/pkg/api/gen/tti/gateway/data/lora/v1"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/config/tlsconfig"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/iotest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

//go:embed testdata/serverca.pem
var serverCAPEM []byte

func writeMessage(ctx context.Context, conn *websocket.Conn, msg *lorav1.GatewayMessage) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageBinary, data)
}

func readMessage(ctx context.Context, conn *websocket.Conn) (*lorav1.NetworkServerMessage, error) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}
	msg := &lorav1.NetworkServerMessage{}
	if err := proto.Unmarshal(data, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func TestFrontend(t *testing.T) { //nolint:gocyclo
	t.Parallel()

	gatewayCerts := map[types.EUI64]tls.Certificate{
		{0xaa, 0xee, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}: test.Must(
			tls.LoadX509KeyPair("testdata/aaee000000000000.pem", "testdata/aaee000000000000-key.pem"),
		),
		{0xbb, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}: test.Must(
			tls.LoadX509KeyPair("testdata/bbff000000000000.pem", "testdata/bbff000000000000-key.pem"),
		),
	}

	iotest.Frontend(t, iotest.FrontendConfig{
		DropsCRCFailure:      true,
		DropsInvalidLoRaWAN:  false,
		SupportsStatus:       false,
		DetectsDisconnect:    true,
		AuthenticatesWithEUI: true,
		IsAuthenticated:      true,
		DeduplicatesUplinks:  true,
		UsesGatewayToken:     true,
		CustomComponentConfig: func(componentConfig *component.Config) {
			componentConfig.TLS = tlsconfig.Config{
				ServerAuth: tlsconfig.ServerAuth{
					Source:      "file",
					Certificate: "testdata/servercert.pem",
					Key:         "testdata/serverkey.pem",
				},
			}
			componentConfig.MTLSAuth = config.MTLSAuthConfig{
				Source:    "directory",
				Directory: "testdata/mtls",
			}
		},
		CustomGatewayServerConfig: func(gsConfig *gatewayserver.Config) {
			gsConfig.TheThingsIndustriesGateway.ListenTLS = ":8889"
		},
		Link: func(
			ctx context.Context,
			_ *testing.T,
			_ *gatewayserver.GatewayServer,
			ids *ttnpb.GatewayIdentifiers,
			_ string,
			upCh <-chan *ttnpb.GatewayUp,
			downCh chan<- *ttnpb.GatewayDown,
		) error {
			rootCAs := x509.NewCertPool()
			rootCAs.AppendCertsFromPEM(serverCAPEM)
			transport := http.DefaultTransport.(*http.Transport).Clone()
			transport.TLSClientConfig = &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{gatewayCerts[types.MustEUI64(ids.Eui).OrZero()]},
				RootCAs:      rootCAs,
			}
			conn, _, err := websocket.Dial( //nolint:bodyclose
				ctx,
				"wss://localhost:8889/api/protocols/tti/v1",
				&websocket.DialOptions{
					Subprotocols: []string{"v1.lora.data.gateway.thethings.industries"},
					HTTPClient: &http.Client{
						Transport: transport,
					},
				},
			)
			if err != nil {
				return err
			}
			defer conn.CloseNow() //nolint:errcheck

			if err := writeMessage(ctx, conn, &lorav1.GatewayMessage{
				Message: &lorav1.GatewayMessage_ClientHelloNotification{
					ClientHelloNotification: &lorav1.ClientHelloNotification{
						DeviceManufacturer: 0x42,
						DeviceModel:        "test",
						HardwareVersion:    "test",
						RuntimeVersion:     "test",
						FirmwareVersion:    "test",
					},
				},
			}); err != nil {
				return err
			}

			// First message is the server hello, ignore.
			if serverHello, err := readMessage(ctx, conn); err != nil {
				return err
			} else if serverHello.GetServerHelloNotification() == nil {
				return fmt.Errorf("expected server hello, got %T", serverHello.Message)
			}

			// Second message is the gateway configuration.
			// This is used to build some state about IF chains and TX channels that are used in uplink and downlink messages.
			gwConfigMsg, err := readMessage(ctx, conn)
			if err != nil {
				return err
			}
			gwConfig := gwConfigMsg.GetConfigureLoraGatewayRequest().GetConfig()
			if gwConfig == nil {
				return fmt.Errorf("expected configure LoRa gateway request, got %T", gwConfigMsg.Message)
			}
			var (
				multiSFIFChains = map[uint64]uint32{}
				txFrequencies   []uint64
				txBandwidths    []uint32
			)
			for _, b := range gwConfig.Boards {
				rfChainFreqs := []int64{int64(b.RfChain0.GetFrequency()), int64(b.RfChain1.GetFrequency())} //nolint:gosec
				for i, multiSF := range []*lorav1.Board_IntermediateFrequencies_MultipleSF{
					b.Ifs.GetMultipleSf0(),
					b.Ifs.GetMultipleSf1(),
					b.Ifs.GetMultipleSf2(),
					b.Ifs.GetMultipleSf3(),
					b.Ifs.GetMultipleSf4(),
					b.Ifs.GetMultipleSf5(),
					b.Ifs.GetMultipleSf6(),
					b.Ifs.GetMultipleSf7(),
				} {
					if multiSF == nil {
						continue
					}
					freq := uint64(rfChainFreqs[multiSF.RfChain] + int64(multiSF.Frequency)) //nolint:gosec
					multiSFIFChains[freq] = uint32(i)                                        //nolint:gosec
				}
			}
			for _, ch := range gwConfig.Tx {
				txFrequencies = append(txFrequencies, ch.Frequency)
				txBandwidths = append(txBandwidths, map[lorav1.Bandwidth]uint32{
					lorav1.Bandwidth_BANDWIDTH_125_KHZ: 125000,
					lorav1.Bandwidth_BANDWIDTH_250_KHZ: 250000,
					lorav1.Bandwidth_BANDWIDTH_500_KHZ: 500000,
				}[ch.Bandwidth])
			}

			replyCh := make(chan *lorav1.GatewayMessage, 1)
			wg, ctx := errgroup.WithContext(ctx)
			// Write upstream.
			wg.Go(func() error {
				for {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case msg := <-upCh:
						if len(msg.UplinkMessages) > 0 {
							messages := make([]*lorav1.UplinkMessage, 0, len(msg.UplinkMessages))
							for _, up := range msg.UplinkMessages {
								uplink := &lorav1.UplinkMessage{
									Board:     0,
									Timestamp: up.RxMetadata[0].Timestamp,
									RssiChannel: &lorav1.UplinkMessage_RssiChannelNegatedDeprecated{
										RssiChannelNegatedDeprecated: -up.RxMetadata[0].ChannelRssi,
									},
									Payload: up.RawPayload,
								}
								switch mod := up.Settings.DataRate.Modulation.(type) {
								case *ttnpb.DataRate_Lora:
									dr := &lorav1.UplinkMessage_Lora{
										RssiSignal: &lorav1.UplinkMessage_Lora_RssiSignalNegatedDeprecated{
											RssiSignalNegatedDeprecated: -up.RxMetadata[0].SignalRssi.GetValue(),
										},
										SpreadingFactor: mod.Lora.SpreadingFactor,
										CodeRate: map[string]lorav1.CodeRate{
											"4/5": lorav1.CodeRate_CODE_RATE_4_5,
											"4/6": lorav1.CodeRate_CODE_RATE_4_6,
											"4/7": lorav1.CodeRate_CODE_RATE_4_7,
											"4/8": lorav1.CodeRate_CODE_RATE_4_8,
										}[mod.Lora.CodingRate],
									}
									if up.RxMetadata[0].Snr < 0 {
										dr.Snr = &lorav1.UplinkMessage_Lora_SnrNegative{
											SnrNegative: -up.RxMetadata[0].Snr,
										}
									} else {
										dr.Snr = &lorav1.UplinkMessage_Lora_SnrPositive{
											SnrPositive: up.RxMetadata[0].Snr,
										}
									}
									uplink.DataRate = &lorav1.UplinkMessage_Lora_{
										Lora: dr,
									}
									if mod.Lora.Bandwidth == 125000 { // Assume multi-SF.
										uplink.IfChain = multiSFIFChains[up.Settings.Frequency]
									} else {
										uplink.IfChain = 9 // LoRa service channel
									}
								case *ttnpb.DataRate_Fsk:
									uplink.DataRate = &lorav1.UplinkMessage_Fsk{
										Fsk: &lorav1.UplinkMessage_FSK{},
									}
									uplink.IfChain = 8 // FSK
								}
								messages = append(messages, uplink)
							}
							if err := writeMessage(ctx, conn, &lorav1.GatewayMessage{
								Message: &lorav1.GatewayMessage_UplinkMessagesNotification{
									UplinkMessagesNotification: &lorav1.UplinkMessagesNotification{
										Messages: messages,
									},
								},
							}); err != nil {
								return err
							}
						}
						if msg.TxAcknowledgment != nil {
							switch msg.TxAcknowledgment.Result {
							case ttnpb.TxAcknowledgment_SUCCESS:
								if err := writeMessage(ctx, conn, &lorav1.GatewayMessage{
									Message: &lorav1.GatewayMessage_TransmitDownlinkResponse{},
								}); err != nil {
									return err
								}
							default:
								if err := writeMessage(ctx, conn, &lorav1.GatewayMessage{
									Message: &lorav1.GatewayMessage_ErrorNotification{
										ErrorNotification: &lorav1.ErrorNotification{
											Code: map[ttnpb.TxAcknowledgment_Result]lorav1.ErrorCode{
												ttnpb.TxAcknowledgment_TOO_LATE:  lorav1.ErrorCode_ERROR_CODE_TX_TOO_LATE,
												ttnpb.TxAcknowledgment_TOO_EARLY: lorav1.ErrorCode_ERROR_CODE_TX_TOO_EARLY,
												ttnpb.TxAcknowledgment_TX_FREQ:   lorav1.ErrorCode_ERROR_CODE_TX_FREQUENCY,
												ttnpb.TxAcknowledgment_TX_POWER:  lorav1.ErrorCode_ERROR_CODE_TX_POWER,
											}[msg.TxAcknowledgment.Result],
										},
									},
								}); err != nil {
									return err
								}
							}
						}
					case reply := <-replyCh:
						data, err := proto.Marshal(reply)
						if err != nil {
							return err
						}
						if err := conn.Write(ctx, websocket.MessageBinary, data); err != nil {
							return err
						}
					}
				}
			})
			// Read downstream.
			wg.Go(func() error {
				for {
					envelope, err := readMessage(ctx, conn)
					if err != nil {
						return err
					}
					switch msg := envelope.Message.(type) {
					case *lorav1.NetworkServerMessage_ConfigureLoraGatewayRequest:
						select {
						case <-ctx.Done():
							return ctx.Err()
						case replyCh <- &lorav1.GatewayMessage{
							TransactionId: envelope.TransactionId,
							Message: &lorav1.GatewayMessage_ConfigureLoraGatewayResponse{
								ConfigureLoraGatewayResponse: &lorav1.ConfigureLoraGatewayResponse{},
							},
						}:
						}
					case *lorav1.NetworkServerMessage_TransmitDownlinkRequest:
						var (
							frequency uint64
							bandwidth uint32
						)
						switch txCh := msg.TransmitDownlinkRequest.Message.TxChannel.(type) {
						case *lorav1.DownlinkMessage_TxChannelIndex:
							frequency = txFrequencies[txCh.TxChannelIndex]
							bandwidth = txBandwidths[txCh.TxChannelIndex]
						case *lorav1.DownlinkMessage_TxChannelConfig:
							frequency = txCh.TxChannelConfig.Frequency
							bandwidth = map[lorav1.Bandwidth]uint32{
								lorav1.Bandwidth_BANDWIDTH_125_KHZ: 125000,
								lorav1.Bandwidth_BANDWIDTH_250_KHZ: 250000,
								lorav1.Bandwidth_BANDWIDTH_500_KHZ: 500000,
							}[txCh.TxChannelConfig.Bandwidth]
						}
						scheduled := &ttnpb.TxSettings{
							Frequency: frequency,
							Timestamp: msg.TransmitDownlinkRequest.Message.Timestamp,
							Downlink: &ttnpb.TxSettings_Downlink{
								TxPower: float32(msg.TransmitDownlinkRequest.Message.TxPower) + 2.15,
							},
						}
						switch dataRate := msg.TransmitDownlinkRequest.Message.DataRate.(type) {
						case *lorav1.DownlinkMessage_Lora_:
							scheduled.DataRate = &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{
									Lora: &ttnpb.LoRaDataRate{
										Bandwidth:       bandwidth,
										SpreadingFactor: dataRate.Lora.SpreadingFactor,
										CodingRate: map[lorav1.CodeRate]string{
											lorav1.CodeRate_CODE_RATE_4_5: "4/5",
											lorav1.CodeRate_CODE_RATE_4_6: "4/6",
											lorav1.CodeRate_CODE_RATE_4_7: "4/7",
											lorav1.CodeRate_CODE_RATE_4_8: "4/8",
										}[dataRate.Lora.CodeRate],
									},
								},
							}
							scheduled.EnableCrc = dataRate.Lora.LorawanUplink
							scheduled.Downlink.InvertPolarization = !dataRate.Lora.LorawanUplink
						case *lorav1.DownlinkMessage_Fsk:
							scheduled.DataRate = &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Fsk{
									Fsk: &ttnpb.FSKDataRate{
										BitRate: dataRate.Fsk.Bitrate,
									},
								},
							}
						}
						select {
						case <-ctx.Done():
							return ctx.Err()
						case downCh <- &ttnpb.GatewayDown{
							DownlinkMessage: &ttnpb.DownlinkMessage{
								RawPayload: msg.TransmitDownlinkRequest.Message.Payload,
								Settings: &ttnpb.DownlinkMessage_Scheduled{
									Scheduled: scheduled,
								},
							},
						}:
						}
					}
				}
			})
			err = wg.Wait()
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return err
			}
		},
	})
}
