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
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/smarty/assertions"
	mappingpb "go.packetbroker.org/api/mapping/v2"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	pkgpacketbroker "go.thethings.network/lorawan-stack/v3/pkg/packetbroker"
	. "go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	timeout     = (1 << 7) * test.Delay
	testOptions = []Option{
		WithTestAuthenticator(&ttnpb.PacketBrokerNetworkIdentifier{
			NetId:    0x000013,
			TenantId: "foo-tenant",
		}),
	}
)

func TestComponent(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := componenttest.NewComponent(t, &component.Config{})

	test.Must(New(c, &Config{
		AuthenticationMode: "oauth2",
	}, testOptions...))
	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_PACKET_BROKER_AGENT)
}

func TestForwarder(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()

	a := assertions.New(t)

	// Elements for the valid test cases.
	userFoo := &ttnpb.UserIdentifiers{UserId: "userfoo"}
	userBar := &ttnpb.UserIdentifiers{UserId: "userbar"}
	orgID := &ttnpb.OrganizationIdentifiers{OrganizationId: "foo-org"}
	_, err := is.UserRegistry().Create(ctx, &ttnpb.CreateUserRequest{
		User: &ttnpb.User{
			Ids:                            userFoo,
			PrimaryEmailAddress:            "usr-foo@example.com",
			PrimaryEmailAddressValidatedAt: timestamppb.New(time.Now()),
		},
	})
	a.So(err, should.BeNil)
	_, err = is.UserRegistry().Create(ctx, &ttnpb.CreateUserRequest{
		User: &ttnpb.User{
			Ids:                            userBar,
			PrimaryEmailAddress:            "usr-bar@example.com",
			PrimaryEmailAddressValidatedAt: timestamppb.New(time.Now()),
		},
	})
	a.So(err, should.BeNil)
	_, err = is.OrganizationRegistry().Create(ctx, &ttnpb.CreateOrganizationRequest{
		Organization: &ttnpb.Organization{
			Ids:                   orgID,
			AdministrativeContact: userFoo.GetOrganizationOrUserIdentifiers(),
			TechnicalContact:      userBar.GetOrganizationOrUserIdentifiers(),
		},
	})
	a.So(err, should.BeNil)

	// Elements for the non valid test cases.
	userEmailNotValidated := &ttnpb.UserIdentifiers{UserId: "usr-email-not-validated"}
	orgEmailNotValidated := &ttnpb.OrganizationIdentifiers{OrganizationId: "org-email-not-validated"}
	_, err = is.UserRegistry().Create(ctx, &ttnpb.CreateUserRequest{User: &ttnpb.User{Ids: userEmailNotValidated}})
	a.So(err, should.BeNil)
	_, err = is.OrganizationRegistry().Create(ctx, &ttnpb.CreateOrganizationRequest{
		Organization: &ttnpb.Organization{
			Ids:                   orgEmailNotValidated,
			AdministrativeContact: userEmailNotValidated.GetOrganizationOrUserIdentifiers(),
			TechnicalContact:      userEmailNotValidated.GetOrganizationOrUserIdentifiers(),
		},
	})
	a.So(err, should.BeNil)

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
			Cluster: cluster.Config{IdentityServer: isAddr},
		},
	})

	_, iamAddr := mustServePBIAM(ctx, t)
	_, cpAddr := mustServePBControlPane(ctx, t)
	dp, dpAddr := mustServePBDataPlane(ctx, t)
	mp, mpAddr := mustServePBMapper(ctx, t)

	gs := test.Must(mock.NewGatewayServer(c))
	tokenKey := bytes.Repeat([]byte{0x42}, 16)
	blockCipher := test.Must(aes.NewCipher(tokenKey))
	aead := test.Must(cipher.NewGCM(blockCipher))
	test.Must(New(c, &Config{
		IAMAddress:          iamAddr.String(),
		ControlPlaneAddress: cpAddr.String(),
		DataPlaneAddress:    dpAddr.String(),
		MapperAddress:       mpAddr.String(),
		NetID:               types.NetID{0x0, 0x0, 0x13},
		TenantID:            "foo-tenant",
		ClusterID:           "test",
		Forwarder: ForwarderConfig{
			Enable: true,
			WorkerPool: WorkerPoolConfig{
				Limit: 1,
			},
			TokenKey:          tokenKey,
			TokenAEAD:         aead,
			IncludeGatewayEUI: true,
			IncludeGatewayID:  true,
			HashGatewayID:     true,
			GatewayOnlineTTL:  10 * time.Minute,
		},
	}, testOptions...))
	componenttest.StartComponent(t, c)
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_GATEWAY_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_PACKET_BROKER_AGENT)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	receivedAt := time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC)

	t.Run("Uplink", func(t *testing.T) {
		for i, tc := range []struct {
			GatewayMessage      *ttnpb.GatewayUplinkMessage
			RoutedUplinkMessage *packetbroker.RoutedUplinkMessage
		}{
			{
				GatewayMessage: &ttnpb.GatewayUplinkMessage{
					Message: &ttnpb.UplinkMessage{
						RawPayload: []byte{0x40, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
						ReceivedAt: timestamppb.New(receivedAt),
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &ttnpb.GatewayIdentifiers{
									GatewayId: "foo-gateway",
									Eui:       types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}.Bytes(),
								},
								ChannelRssi: -42,
								Rssi:        -42,
								Snr:         10.5,
								Location: &ttnpb.Location{
									Latitude:  52.5,
									Longitude: 4.8,
									Altitude:  2,
								},
								UplinkToken: []byte("test-token"),
								Timestamp:   123456,
							},
						},
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
							Frequency: 869525000,
						},
					},
					BandId: "EU_863_870",
				},
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					ForwarderNetId:     0x000013,
					ForwarderTenantId:  "foo-tenant",
					ForwarderClusterId: "test",
					Message: &packetbroker.UplinkMessage{
						GatewayId: &packetbroker.GatewayIdentifier{
							Eui: &wrapperspb.UInt64Value{
								Value: 0x1122334455667788,
							},
							Id: &packetbroker.GatewayIdentifier_Hash{
								Hash: []byte{0xc7, 0x4a, 0x72, 0x7c, 0xe5, 0x01, 0xe9, 0xc1, 0x20, 0x6b, 0xb2, 0x81, 0x82, 0xeb, 0x06, 0x91, 0x7f, 0x94, 0x43, 0x54, 0x30, 0x90, 0x78, 0x0f, 0x3a, 0x39, 0x3d, 0xeb, 0xad, 0x91, 0xad, 0x96},
							},
						},
						ForwarderReceiveTime: timestamppb.New(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC)),
						DataRate:             packetbroker.NewLoRaDataRate(7, 125000, band.Cr4_5),
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
								Hash:   []byte{0x76, 0x9f, 0xce, 0x31, 0xe8, 0x1a, 0x90, 0xa1, 0x17, 0x07, 0x69, 0x18, 0x3b, 0x24, 0x0f, 0xd9, 0x8b, 0x7f, 0x38, 0xc7, 0x86, 0xb3, 0xd4, 0xe3, 0x8d, 0xae, 0xe1, 0x73, 0xe3, 0xa4, 0xcf, 0xbd},
								Length: 15,
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
					Message: &ttnpb.UplinkMessage{
						RawPayload: []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0x42, 0x22, 0x11, 0x1, 0x2, 0x3, 0x4},
						ReceivedAt: timestamppb.New(receivedAt),
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIds: &ttnpb.GatewayIdentifiers{
									GatewayId: "foo-gateway",
									Eui:       types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}.Bytes(),
								},
								ChannelRssi: 4.2,
								Rssi:        4.2,
								Snr:         -5.5,
								UplinkToken: []byte("test-token"),
								Timestamp:   123456,
							},
						},
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
							Frequency: 868300000,
						},
					},
					BandId: "EU_863_870",
				},
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					ForwarderNetId:     0x000013,
					ForwarderTenantId:  "foo-tenant",
					ForwarderClusterId: "test",
					Message: &packetbroker.UplinkMessage{
						GatewayId: &packetbroker.GatewayIdentifier{
							Eui: &wrapperspb.UInt64Value{
								Value: 0x1122334455667788,
							},
							Id: &packetbroker.GatewayIdentifier_Hash{
								Hash: []byte{0xc7, 0x4a, 0x72, 0x7c, 0xe5, 0x01, 0xe9, 0xc1, 0x20, 0x6b, 0xb2, 0x81, 0x82, 0xeb, 0x06, 0x91, 0x7f, 0x94, 0x43, 0x54, 0x30, 0x90, 0x78, 0x0f, 0x3a, 0x39, 0x3d, 0xeb, 0xad, 0x91, 0xad, 0x96},
							},
						},
						ForwarderReceiveTime: timestamppb.New(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC)),
						DataRate:             packetbroker.NewLoRaDataRate(9, 125000, band.Cr4_5),
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
								Hash:   []byte{0xce, 0xb5, 0x2a, 0x44, 0x27, 0xb9, 0x4d, 0x8a, 0xff, 0x4c, 0x6d, 0x20, 0xf5, 0x7d, 0x81, 0x66, 0x62, 0x9e, 0x6a, 0x26, 0xe6, 0x4c, 0x5f, 0x77, 0x2f, 0x70, 0xa7, 0xac, 0x34, 0x6a, 0x38, 0x81},
								Length: 23,
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
					t.Fatal("Expected uplink message from Home Network")
				}
				pbMsg.Message.GatewayUplinkToken = nil // JWE, tested by TestWrapGatewayUplinkToken
				a.So(pbMsg, should.Resemble, tc.RoutedUplinkMessage)
			})
		}
	})

	t.Run("Downlink", func(t *testing.T) {
		a := assertions.New(t)

		plainToken := test.Must(proto.Marshal(&ttnpb.PacketBrokerAgentGatewayUplinkToken{
			GatewayUid: unique.ID(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}),
			Token:      []byte{0x1, 0x2, 0x3, 0x4},
		}))
		tokenNonce := make([]byte, aead.NonceSize())
		test.Must(io.ReadFull(rand.Reader, tokenNonce))
		token := test.Must(proto.Marshal(&ttnpb.PacketBrokerAgentEncryptedPayload{
			Ciphertext: aead.Seal(nil, tokenNonce, plainToken, nil),
			Nonce:      tokenNonce,
		}))

		dp.ForwarderDown <- &packetbroker.RoutedDownlinkMessage{
			ForwarderNetId:      0x000013,
			ForwarderTenantId:   "foo-tenant",
			ForwarderClusterId:  "test",
			HomeNetworkNetId:    0x000042,
			HomeNetworkTenantId: "foo-tenant",
			Id:                  "test",
			Message: &packetbroker.DownlinkMessage{
				PhyPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
				Region:     packetbroker.Region_EU_863_870,
				Class:      packetbroker.DownlinkMessageClass_CLASS_A,
				Priority:   packetbroker.DownlinkMessagePriority_NORMAL,
				Rx1: &packetbroker.DownlinkMessage_RXSettings{
					Frequency: 868100000,
					DataRate:  packetbroker.NewLoRaDataRate(7, 125000, ""),
				},
				Rx2: &packetbroker.DownlinkMessage_RXSettings{
					Frequency: 869525000,
					DataRate:  packetbroker.NewLoRaDataRate(12, 125000, ""),
				},
				Rx1Delay:           durationpb.New(5 * time.Second),
				GatewayUplinkToken: token,
			},
		}

		var gtwMsg *ttnpb.DownlinkMessage
		select {
		case gtwMsg = <-gs.Downlink:
		case <-time.After(timeout):
			t.Fatal("Expected downlink message from Home Network")
		}
		a.So(gtwMsg, should.Resemble, &ttnpb.DownlinkMessage{
			RawPayload:     []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
			CorrelationIds: gtwMsg.CorrelationIds,
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: &ttnpb.TxRequest{
					Class: ttnpb.Class_CLASS_A,
					DownlinkPaths: []*ttnpb.DownlinkPath{
						{
							Path: &ttnpb.DownlinkPath_UplinkToken{
								UplinkToken: []byte{0x1, 0x2, 0x3, 0x4},
							},
						},
					},
					Priority: ttnpb.TxSchedulePriority_NORMAL,
					Rx1DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 7,
								Bandwidth:       125000,
							},
						},
					},
					Rx1Frequency: 868100000,
					Rx1Delay:     ttnpb.RxDelay_RX_DELAY_5,
					Rx2DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 12,
								Bandwidth:       125000,
							},
						},
					},
					Rx2Frequency: 869525000,
				},
			},
		})

		var stateChange *packetbroker.DownlinkMessageDeliveryStateChange
		select {
		case stateChange = <-dp.ForwarderDownStateChange:
		case <-time.After(timeout):
			t.Fatal("Expected downlink message delivery state change from Home Network")
		}
		a.So(stateChange.GetSuccess(), should.NotBeNil)
	})

	t.Run("Update gateway", func(t *testing.T) {
		a := assertions.New(t)

		updateCh := make(chan *mappingpb.UpdateGatewayRequest, 1)
		mp.UpdateGatewayHandler = func(ctx context.Context, req *mappingpb.UpdateGatewayRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
			updateCh <- req
			return ttnpb.Empty, nil
		}

		for _, tc := range []struct {
			name               string
			fieldMask          *fieldmaskpb.FieldMask
			setGatewayContacts func(*ttnpb.PacketBrokerGateway)
			validateResponse   func(*mappingpb.UpdateGatewayRequest)
		}{
			{
				name: "No Admin|tech ContactInfo present",
				fieldMask: ttnpb.FieldMask(
					"antennas",
					"frequency_plan_ids",
					"ids",
					"location_public",
					"online",
					"status_public",
				),

				setGatewayContacts: func(*ttnpb.PacketBrokerGateway) {},
				validateResponse: func(up *mappingpb.UpdateGatewayRequest) {
					a.So(up.AdministrativeContact.GetValue().GetEmail(), should.BeEmpty)
					a.So(up.TechnicalContact.GetValue().GetEmail(), should.BeEmpty)
					a.So(up.FrequencyPlan.GetLoraMultiSfChannels(), should.HaveLength, 8)
					a.So(up.Online.GetValue(), should.BeTrue)
					a.So(up.GatewayLocation.GetLocation().GetTerrestrial().GetAntennaCount().GetValue(), should.Equal, 1)
				},
			},
			{
				name: "Valid Admin|tech ContactInfo present",
				fieldMask: ttnpb.FieldMask(
					"administrative_contact",
					"antennas",
					"frequency_plan_ids",
					"ids",
					"location_public",
					"online",
					"status_public",
					"technical_contact",
				),
				setGatewayContacts: func(gtw *ttnpb.PacketBrokerGateway) {
					gtw.AdministrativeContact = userFoo.GetOrganizationOrUserIdentifiers()
					gtw.TechnicalContact = orgID.GetOrganizationOrUserIdentifiers()
				},
				validateResponse: func(up *mappingpb.UpdateGatewayRequest) {
					a.So(up.AdministrativeContact.GetValue().GetName(), should.Equal, userFoo.UserId)
					a.So(up.AdministrativeContact.GetValue().GetEmail(), should.Equal, "usr-foo@example.com")
					a.So(up.TechnicalContact.GetValue().GetName(), should.Equal, userBar.UserId)
					a.So(up.TechnicalContact.GetValue().GetEmail(), should.Equal, "usr-bar@example.com")
					a.So(up.FrequencyPlan.GetLoraMultiSfChannels(), should.HaveLength, 8)
					a.So(up.Online.GetValue(), should.BeTrue)
					a.So(up.GatewayLocation.GetLocation().GetTerrestrial().GetAntennaCount().GetValue(), should.Equal, 1)
				},
			},
			{
				name: "Email not validated in Admin|tech ContactInfo",
				fieldMask: ttnpb.FieldMask(
					"administrative_contact",
					"antennas",
					"frequency_plan_ids",
					"ids",
					"location_public",
					"online",
					"status_public",
					"technical_contact",
				),
				setGatewayContacts: func(gtw *ttnpb.PacketBrokerGateway) {
					gtw.AdministrativeContact = orgEmailNotValidated.GetOrganizationOrUserIdentifiers()
					gtw.TechnicalContact = orgEmailNotValidated.GetOrganizationOrUserIdentifiers()
				},
				validateResponse: func(up *mappingpb.UpdateGatewayRequest) {
					a.So(up.AdministrativeContact.GetValue().GetEmail(), should.BeEmpty)
					a.So(up.TechnicalContact.GetValue().GetEmail(), should.BeEmpty)
					a.So(up.FrequencyPlan.GetLoraMultiSfChannels(), should.HaveLength, 8)
					a.So(up.Online.GetValue(), should.BeTrue)
					a.So(up.GatewayLocation.GetLocation().GetTerrestrial().GetAntennaCount().GetValue(), should.Equal, 1)
				},
			},
		} {
			tc := tc
			t.Run(tc.name, func(t *testing.T) { // nolint:paralleltest
				//  If validateUpdate or setRequestContacts are not set, the tests will panic.
				if tc.validateResponse == nil {
					t.Fatal("Test has to contain a validateUpdate function")
				}
				if tc.setGatewayContacts == nil {
					t.Fatal("Test has to contain a setRequestContacts function")
				}

				req := &ttnpb.UpdatePacketBrokerGatewayRequest{
					Gateway: &ttnpb.PacketBrokerGateway{
						Ids: &ttnpb.PacketBrokerGateway_GatewayIdentifiers{
							GatewayId: "foo-gateway",
							Eui:       types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}.Bytes(),
						},
						FrequencyPlanIds: []string{"EU_863_870"},
						Antennas: []*ttnpb.GatewayAntenna{
							{
								Location: &ttnpb.Location{
									Latitude:  4.85464,
									Longitude: 52.34562,
									Altitude:  16,
									Accuracy:  10,
									Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
								},
							},
						},
						StatusPublic:   true,
						LocationPublic: true,
						Online:         true,
					},
					FieldMask: tc.fieldMask,
				}

				tc.setGatewayContacts(req.Gateway)

				res, err := gs.UpdateGateway(ctx, req)
				a.So(err, should.BeNil)
				a.So(res.OnlineTtl.AsDuration(), should.NotBeZeroValue)

				select {
				case update := <-updateCh:
					tc.validateResponse(update)
				case <-time.After(timeout):
					t.Fatal("Expected gateway update timeout")
				}
			})
		}
	})
}

func TestHomeNetwork(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})

	_, iamAddr := mustServePBIAM(ctx, t)
	_, cpAddr := mustServePBControlPane(ctx, t)
	dp, dpAddr := mustServePBDataPlane(ctx, t)

	ns := test.Must(mock.NewNetworkServer(c))
	test.Must(New(c, &Config{
		IAMAddress:          iamAddr.String(),
		ControlPlaneAddress: cpAddr.String(),
		DataPlaneAddress:    dpAddr.String(),
		NetID:               types.NetID{0x0, 0x0, 0x13},
		TenantID:            "foo-tenant",
		ClusterID:           "test",
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
			// With location information and without fully defined data rate.
			{
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					ForwarderNetId:       0x000042,
					ForwarderTenantId:    "foo-tenant",
					ForwarderClusterId:   "test",
					HomeNetworkNetId:     0x000013,
					HomeNetworkTenantId:  "foo-tenant",
					HomeNetworkClusterId: "test",
					Id:                   "test",
					Message: &packetbroker.UplinkMessage{
						GatewayId: &packetbroker.GatewayIdentifier{
							Eui: &wrapperspb.UInt64Value{
								Value: 0x1122334455667788,
							},
							Id: &packetbroker.GatewayIdentifier_Plain{
								Plain: "foo-gateway",
							},
						},
						DataRate:             packetbroker.NewLoRaDataRate(7, 125000, band.Cr4_5),
						ForwarderReceiveTime: timestamppb.New(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC)),
						Frequency:            869525000,
						CodingRate:           band.Cr4_5,
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
												{
													Index: 1,
													Value: &packetbroker.TerrestrialGatewayAntennaSignalQuality{
														ChannelRssi:     -43,
														Snr:             10.6,
														FrequencyOffset: 1,
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
							GatewayIds:   pkgpacketbroker.GatewayIdentifiers,
							AntennaIndex: 0,
							PacketBroker: &ttnpb.PacketBrokerMetadata{
								MessageId:           "test",
								ForwarderNetId:      types.NetID{0x0, 0x0, 0x42}.Bytes(),
								ForwarderTenantId:   "foo-tenant",
								ForwarderClusterId:  "test",
								ForwarderGatewayEui: types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}.Bytes(),
								ForwarderGatewayId: &wrapperspb.StringValue{
									Value: "foo-gateway",
								},
								HomeNetworkNetId:     types.NetID{0x0, 0x0, 0x13}.Bytes(),
								HomeNetworkTenantId:  "foo-tenant",
								HomeNetworkClusterId: "test",
							},
							ChannelRssi: -42,
							Rssi:        -42,
							Snr:         10.5,
							Location: &ttnpb.Location{
								Latitude:  52.5,
								Longitude: 4.8,
								Altitude:  2,
							},
							UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &ttnpb.PacketBrokerAgentUplinkToken{
								ForwarderNetId:     []byte{0x0, 0x0, 0x42},
								ForwarderTenantId:  "foo-tenant",
								ForwarderClusterId: "test",
							})),
						},
						{
							GatewayIds:   pkgpacketbroker.GatewayIdentifiers,
							AntennaIndex: 1,
							PacketBroker: &ttnpb.PacketBrokerMetadata{
								MessageId:           "test",
								ForwarderNetId:      types.NetID{0x0, 0x0, 0x42}.Bytes(),
								ForwarderTenantId:   "foo-tenant",
								ForwarderClusterId:  "test",
								ForwarderGatewayEui: types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}.Bytes(),
								ForwarderGatewayId: &wrapperspb.StringValue{
									Value: "foo-gateway",
								},
								HomeNetworkNetId:     types.NetID{0x0, 0x0, 0x13}.Bytes(),
								HomeNetworkTenantId:  "foo-tenant",
								HomeNetworkClusterId: "test",
							},
							ChannelRssi:     -43,
							Rssi:            -43,
							Snr:             10.6,
							FrequencyOffset: 1,
							UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &ttnpb.PacketBrokerAgentUplinkToken{
								ForwarderNetId:     []byte{0x0, 0x0, 0x42},
								ForwarderTenantId:  "foo-tenant",
								ForwarderClusterId: "test",
							})),
						},
					},
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
						Frequency: 869525000,
					},
				},
			},
			// Without location and with fully described data rate.
			{
				RoutedUplinkMessage: &packetbroker.RoutedUplinkMessage{
					ForwarderNetId:       0x000042,
					ForwarderTenantId:    "foo-tenant",
					ForwarderClusterId:   "test",
					HomeNetworkNetId:     0x000013,
					HomeNetworkTenantId:  "foo-tenant",
					HomeNetworkClusterId: "test",
					Id:                   "test",
					Message: &packetbroker.UplinkMessage{
						GatewayId: &packetbroker.GatewayIdentifier{
							Eui: &wrapperspb.UInt64Value{
								Value: 0x1122334455667788,
							},
							Id: &packetbroker.GatewayIdentifier_Plain{
								Plain: "foo-gateway",
							},
						},
						DataRate:             packetbroker.NewLoRaDataRate(9, 125000, band.Cr4_5),
						ForwarderReceiveTime: timestamppb.New(time.Date(2020, time.March, 24, 12, 0, 0, 0, time.UTC)),
						Frequency:            869525000,
						CodingRate:           band.Cr4_5,
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
							GatewayIds: pkgpacketbroker.GatewayIdentifiers,
							PacketBroker: &ttnpb.PacketBrokerMetadata{
								MessageId:           "test",
								ForwarderNetId:      types.NetID{0x0, 0x0, 0x42}.Bytes(),
								ForwarderTenantId:   "foo-tenant",
								ForwarderClusterId:  "test",
								ForwarderGatewayEui: types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}.Bytes(),
								ForwarderGatewayId: &wrapperspb.StringValue{
									Value: "foo-gateway",
								},
								HomeNetworkNetId:     types.NetID{0x0, 0x0, 0x13}.Bytes(),
								HomeNetworkTenantId:  "foo-tenant",
								HomeNetworkClusterId: "test",
							},
							ChannelRssi: 4.2,
							Rssi:        4.2,
							Snr:         -5.5,
							UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &ttnpb.PacketBrokerAgentUplinkToken{
								ForwarderNetId:     []byte{0x0, 0x0, 0x42},
								ForwarderTenantId:  "foo-tenant",
								ForwarderClusterId: "test",
							})),
						},
					},
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
						Frequency: 869525000,
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
				a.So(nsMsg.CorrelationIds, should.HaveLength, 1)
				a.So(*ttnpb.StdTime(nsMsg.ReceivedAt), should.HappenBetween, before, time.Now()) // Packet Broker Agent sets local time on receive.

				expected := ttnpb.Clone(tc.UplinkMessage)
				expected.CorrelationIds = nsMsg.CorrelationIds
				expected.ReceivedAt = nsMsg.ReceivedAt
				for _, md := range expected.RxMetadata {
					md.ReceivedAt = tc.RoutedUplinkMessage.Message.GetForwarderReceiveTime()
				}

				a.So(nsMsg, should.Resemble, expected)

				var stateChange *packetbroker.UplinkMessageDeliveryStateChange
				select {
				case stateChange = <-dp.HomeNetworkUpStateChange:
				case <-time.After(timeout):
					t.Fatal("Expected uplink message delivery state change from Forwarder")
				}
				a.So(stateChange.Error, should.BeNil)
			})
		}
	})

	t.Run("Downlink", func(t *testing.T) {
		a := assertions.New(t)

		nsMsg := &ttnpb.DownlinkMessage{
			RawPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: &ttnpb.TxRequest{
					FrequencyPlanId: test.EUFrequencyPlanID,
					Class:           ttnpb.Class_CLASS_A,
					DownlinkPaths: []*ttnpb.DownlinkPath{
						{
							Path: &ttnpb.DownlinkPath_UplinkToken{
								UplinkToken: test.Must(WrapUplinkTokens([]byte("test-token"), nil, &ttnpb.PacketBrokerAgentUplinkToken{
									ForwarderNetId:     []byte{0x0, 0x0, 0x42},
									ForwarderTenantId:  "foo-tenant",
									ForwarderClusterId: "test",
								})),
							},
						},
					},
					Priority: ttnpb.TxSchedulePriority_NORMAL,
					Rx1DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								Bandwidth:       125000,
								SpreadingFactor: 7,
								CodingRate:      band.Cr4_5,
							},
						},
					},
					Rx1Frequency: 868100000,
					Rx1Delay:     ttnpb.RxDelay_RX_DELAY_5,
					Rx2DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								Bandwidth:       125000,
								SpreadingFactor: 12,
								CodingRate:      band.Cr4_5,
							},
						},
					},
					Rx2Frequency: 869525000,
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
			ForwarderNetId:       0x000042,
			ForwarderClusterId:   "test",
			ForwarderTenantId:    "foo-tenant",
			HomeNetworkNetId:     0x000013,
			HomeNetworkTenantId:  "foo-tenant",
			HomeNetworkClusterId: "test",
			Message: &packetbroker.DownlinkMessage{
				Region:     packetbroker.Region_EU_863_870,
				PhyPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x01, 0x01, 0x00, 0x42, 0x1, 0x42, 0x1, 0x2, 0x3, 0x4},
				Class:      packetbroker.DownlinkMessageClass_CLASS_A,
				Priority:   packetbroker.DownlinkMessagePriority_NORMAL,
				Rx1: &packetbroker.DownlinkMessage_RXSettings{
					Frequency: 868100000,
					DataRate:  packetbroker.NewLoRaDataRate(7, 125000, band.Cr4_5),
				},
				Rx2: &packetbroker.DownlinkMessage_RXSettings{
					Frequency: 869525000,
					DataRate:  packetbroker.NewLoRaDataRate(12, 125000, band.Cr4_5),
				},
				Rx1Delay:           durationpb.New(5 * time.Second),
				GatewayUplinkToken: []byte(`test-token`),
			},
		})
	})
}
