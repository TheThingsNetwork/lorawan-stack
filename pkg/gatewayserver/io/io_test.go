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

package io_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/mock"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

var timeout = (1 << 3) * test.Delay

func assertStatsIncludePaths(a *assertions.Assertion, conn *io.Connection, paths []string) {
	_, statsPaths := conn.Stats()
	for _, path := range paths {
		a.So(statsPaths, should.Contain, path)
	}
}

func TestFlow(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	is, _, closeIS := mockis.New(ctx)
	defer closeIS()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})
	gs := mock.NewServer(c, is)

	ids := ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}
	antennaGain := float32(3)
	gtw := &ttnpb.Gateway{
		Ids:             &ids,
		FrequencyPlanId: "EU_863_870",
		Antennas: []*ttnpb.GatewayAntenna{
			{
				Gain: antennaGain,
			},
		},
	}
	gs.RegisterGateway(ctx, ids, gtw)

	gtwCtx := rights.NewContext(ctx, rights.Rights{
		GatewayRights: map[string]*ttnpb.Rights{
			unique.ID(ctx, ids): ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_LINK),
		},
	})
	frontend, err := mock.ConnectFrontend(gtwCtx, ids, gs)
	if err != nil {
		panic(err)
	}
	conn := gs.GetConnection(ctx, ids)

	a.So(conn.Context(), should.HaveParentContextOrEqual, gtwCtx)
	a.So(time.Since(conn.ConnectTime()), should.BeLessThan, timeout)
	a.So(conn.Gateway(), should.Resemble, gtw)
	a.So(conn.Frontend().Protocol(), should.Equal, "mock")
	a.So(conn.PrimaryFrequencyPlan(), should.NotBeNil)
	a.So(conn.PrimaryFrequencyPlan().BandID, should.Equal, "EU_863_870")

	_, paths := conn.Stats()
	a.So(paths, should.Resemble, []string{"connected_at", "disconnected_at", "protocol"})

	{
		frontend.Up <- &ttnpb.UplinkMessage{
			RxMetadata: []*ttnpb.RxMetadata{
				{
					GatewayIds: &ttnpb.GatewayIdentifiers{
						GatewayId: "test-gateway",
					},
					AntennaIndex: 0,
					Timestamp:    100,
				},
			},
			Settings: &ttnpb.TxSettings{DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}}},
		}
		select {
		case up := <-conn.Up():
			token, err := io.ParseUplinkToken(up.Message.RxMetadata[0].UplinkToken)
			a.So(err, should.BeNil)
			a.So(token.Ids.GatewayIds, should.Resemble, &ids)
			a.So(token.Ids.AntennaIndex, should.Equal, 0)
			a.So(token.Timestamp, should.Equal, 100)
		case <-time.After(timeout):
			t.Fatalf("Expected uplink message time-out")
		}
		time.Sleep(timeout / 2)
		total, t, ok := conn.UpStats()
		a.So(ok, should.BeTrue)
		a.So(total, should.Equal, 1)
		a.So(time.Since(t), should.BeLessThan, timeout)
		assertStatsIncludePaths(a, conn, []string{"last_uplink_received_at", "uplink_count"})
	}

	{
		frontend.Status <- &ttnpb.GatewayStatus{
			Time: ttnpb.ProtoTimePtr(time.Now()),
		}
		select {
		case <-conn.Status():
		case <-time.After(timeout):
			t.Fatalf("Expected status message time-out")
		}
		time.Sleep(timeout / 2)
		last, t, ok := conn.StatusStats()
		a.So(ok, should.BeTrue)
		a.So(last, should.NotBeNil)
		a.So(time.Since(t), should.BeLessThan, timeout)
		assertStatsIncludePaths(a, conn, []string{"last_status_received_at", "last_status"})
	}

	{
		frontend.TxAck <- &ttnpb.TxAcknowledgment{}
		select {
		case <-conn.TxAck():
		case <-time.After(timeout):
			t.Fatalf("Expected Tx acknowledgment time-out")
		}
	}

	received := 0
	for _, tc := range []struct {
		Name             string
		Path             *ttnpb.DownlinkPath
		Message          *ttnpb.DownlinkMessage
		ErrorAssertion   func(error) bool
		RxErrorAssertion []func(error) bool
		ExpectedEIRP     float32
	}{
		{
			Name: "NoRequest",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}},
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
						Frequency: 868100000,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "ValidClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}},
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
					},
				},
			},
			ExpectedEIRP: 16.15 - antennaGain,
		},
		{
			Name: "ConflictClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}},
						100,
						100000,
						time.Unix(0, 100*1000),
						nil,
					), // Same as previous.
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
						}, // Same as previous.
						Rx1Frequency: 868100000, // Same as previous.
					},
				},
			},
			ErrorAssertion: errors.IsAborted,
			RxErrorAssertion: []func(error) bool{
				errors.IsAlreadyExists,      // Rx1 conflicts with previous.
				errors.IsFailedPrecondition, // Rx2 not provided.
			},
		},
		{
			Name: "NoUplinkTokenClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_Fixed{
					Fixed: &ttnpb.GatewayAntennaIdentifiers{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "foo-gateway",
						},
					},
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
						}, // Same as previous.
						Rx1Frequency:    868100000, // Same as previous.
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "NoRx1DelayClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}},
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
						Rx1DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx1Frequency:    868100000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "ValidClassC/UplinkToken",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}},
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
						Class:    ttnpb.Class_CLASS_C,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx2Frequency:    869525000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ExpectedEIRP: 29.15 - antennaGain,
		},
		{
			Name: "ValidClassC/FixedPath",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_Fixed{
					Fixed: &ttnpb.GatewayAntennaIdentifiers{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "foo-gateway",
						},
					},
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_C,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx2Frequency:    869525000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ExpectedEIRP: 29.15 - antennaGain,
		},
		{
			Name: "ValidClassC/AbsoluteTime",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_Fixed{
					Fixed: &ttnpb.GatewayAntennaIdentifiers{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "foo-gateway",
						},
					},
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_C,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx2Frequency:    869525000,
						AbsoluteTime:    ttnpb.ProtoTimePtr(time.Unix(100, 0)), // The mock front-end uses Unix epoch as start time.
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ExpectedEIRP: 29.15 - antennaGain,
		},
		{
			Name: "NoPathClassC",
			Path: &ttnpb.DownlinkPath{},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_C,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 7,
									Bandwidth:       125000,
								},
							},
						},
						Rx2Frequency:    869525000,
						AbsoluteTime:    ttnpb.ProtoTimePtr(time.Unix(100, 0)), // The mock front-end uses Unix epoch as start time.
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "TooLong",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gateway"}},
						100,
						100000,
						time.Unix(0, 100*1000),
						nil,
					),
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: bytes.Repeat([]byte{0x01}, 80),
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:    ttnpb.Class_CLASS_C,
						Priority: ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 12,
									Bandwidth:       125000,
								},
							},
						},
						Rx2Frequency:    869525000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			hasRX1 := tc.Message.GetRequest().GetRx1Frequency() != 0
			hasRX2 := tc.Message.GetRequest().GetRx2Frequency() != 0

			rx1, rx2, _, err := conn.ScheduleDown(tc.Path, tc.Message)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
				for i, assert := range tc.RxErrorAssertion {
					details := errors.Details(err)[0].(*ttnpb.ScheduleDownlinkErrorDetails)
					if !a.So(details, should.NotBeNil) {
						t.FailNow()
					}
					errDetail := ttnpb.ErrorDetailsFromProto(details.PathErrors[i])
					if !a.So(assert(errDetail), should.BeTrue) {
						t.Fatalf("Unexpected Rx window %d error: %v", i+1, errDetail)
					}
				}
				return
			} else if tc.ErrorAssertion != nil {
				t.Fatal("Expected error but got none")
			}
			a.So(rx1, should.Equal, hasRX1)
			a.So(rx2, should.Equal, hasRX2)

			received++
			select {
			case msg := <-frontend.Down:
				scheduled := msg.GetScheduled()
				a.So(scheduled, should.NotBeNil)
				a.So(scheduled.Downlink.TxPower, should.Equal, tc.ExpectedEIRP)
			case <-time.After(timeout):
				t.Fatalf("Expected downlink message timeout")
			}
			total, last, ok := conn.DownStats()
			a.So(ok, should.BeTrue)
			a.So(total, should.Equal, received)
			a.So(time.Since(last), should.BeLessThan, timeout)
			assertStatsIncludePaths(a, conn, []string{"downlink_count", "last_downlink_received_at", "sub_bands"})
		})
	}
}

func TestSubBandEIRPOverride(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	is, _, closeIS := mockis.New(ctx)
	defer closeIS()

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			FrequencyPlans: config.FrequencyPlansConfig{
				ConfigSource: "static",
				Static:       test.StaticFrequencyPlans,
			},
		},
	})

	gs := mock.NewServer(c, is)

	ids := ttnpb.GatewayIdentifiers{GatewayId: "bar-gateway"}
	antennaGain := float32(3)
	gtw := &ttnpb.Gateway{
		Ids:             &ids,
		FrequencyPlanId: "AS_923_925_AU", // Overrides maximum EIRP to 30 dBm in 915.0 – 928.0 MHz sub-band.
		Antennas: []*ttnpb.GatewayAntenna{
			{
				Gain: antennaGain,
			},
		},
	}
	gs.RegisterGateway(ctx, ids, gtw)

	gtwCtx := rights.NewContext(ctx, rights.Rights{
		GatewayRights: map[string]*ttnpb.Rights{
			unique.ID(ctx, ids): ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_LINK),
		},
	})
	frontend, err := mock.ConnectFrontend(gtwCtx, ids, gs)
	if err != nil {
		panic(err)
	}
	conn := gs.GetConnection(ctx, ids)

	a.So(conn.Context(), should.HaveParentContextOrEqual, gtwCtx)
	a.So(time.Since(conn.ConnectTime()), should.BeLessThan, timeout)
	a.So(conn.Gateway(), should.Resemble, gtw)
	a.So(conn.Frontend().Protocol(), should.Equal, "mock")

	// Sync the clock.
	{
		frontend.Up <- &ttnpb.UplinkMessage{
			RxMetadata: []*ttnpb.RxMetadata{
				{
					GatewayIds: &ttnpb.GatewayIdentifiers{
						GatewayId: "test-gateway",
					},
					AntennaIndex: 0,
					Timestamp:    100,
				},
			},
			Settings: &ttnpb.TxSettings{DataRate: &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}}},
		}
		select {
		case up := <-conn.Up():
			token, err := io.ParseUplinkToken(up.Message.RxMetadata[0].UplinkToken)
			a.So(err, should.BeNil)
			a.So(token.Ids.GatewayIds, should.Resemble, &ids)
			a.So(token.Ids.AntennaIndex, should.Equal, 0)
			a.So(token.Timestamp, should.Equal, 100)
		case <-time.After(timeout):
			t.Fatalf("Expected uplink message time-out")
		}
		total, t, ok := conn.UpStats()
		a.So(ok, should.BeTrue)
		a.So(total, should.Equal, 1)
		a.So(time.Since(t), should.BeLessThan, timeout)
	}

	received := 0
	for _, tc := range []struct {
		Name         string
		Path         *ttnpb.DownlinkPath
		Message      *ttnpb.DownlinkMessage
		ExpectedEIRP float32
	}{
		{
			Name: "ValidClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(
						&ttnpb.GatewayAntennaIdentifiers{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "bar-gateway"}},
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
						Rx1Frequency:    923200000,
						FrequencyPlanId: "AS_923_925_AU",
					},
				},
			},
			ExpectedEIRP: 30 - antennaGain,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			_, _, _, err := conn.ScheduleDown(tc.Path, tc.Message)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			received++
			select {
			case msg := <-frontend.Down:
				scheduled := msg.GetScheduled()
				a.So(scheduled, should.NotBeNil)
				a.So(scheduled.Downlink.TxPower, should.Equal, tc.ExpectedEIRP)
			case <-time.After(timeout):
				t.Fatalf("Expected downlink message timeout")
			}
			total, last, ok := conn.DownStats()
			a.So(ok, should.BeTrue)
			a.So(total, should.Equal, received)
			a.So(time.Since(last), should.BeLessThan, timeout)
		})
	}
}

func TestUniqueUplinkMessagesByRSSI(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []*ttnpb.UplinkMessage
		out  []*ttnpb.UplinkMessage
	}{
		{
			name: "nil",
		},
		{
			name: "one",
			in: []*ttnpb.UplinkMessage{{
				RawPayload: []byte{1, 2, 3, 4, 5},
				Settings:   &ttnpb.TxSettings{Frequency: 1000000},
				RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -20, AntennaIndex: 0}},
			}},
			out: []*ttnpb.UplinkMessage{{
				RawPayload: []byte{1, 2, 3, 4, 5},
				Settings:   &ttnpb.TxSettings{Frequency: 1000000},
				RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -20, AntennaIndex: 0}},
			}},
		},
		{
			name: "deduplicate",
			in: []*ttnpb.UplinkMessage{
				{
					RawPayload: []byte{1, 2, 3, 4},
					Settings:   &ttnpb.TxSettings{Frequency: 1200000},
				},
				{
					RawPayload: []byte{1, 2, 3, 4, 5},
					Settings:   &ttnpb.TxSettings{Frequency: 1200000},
					RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -40, AntennaIndex: 0}},
				},
				{
					RawPayload: []byte{1, 2, 3, 4, 5},
					Settings:   &ttnpb.TxSettings{Frequency: 1000000},
					RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -20, AntennaIndex: 0}},
				},
				{
					RawPayload: []byte{1, 2, 3, 4, 5},
					Settings:   &ttnpb.TxSettings{Frequency: 1100000},
					RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -100, AntennaIndex: 0}},
				},
				{
					RawPayload: []byte{1, 2, 3, 4, 5, 6},
					Settings:   &ttnpb.TxSettings{Frequency: 1000000},
					RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -10, AntennaIndex: 0}},
				},
			},
			out: []*ttnpb.UplinkMessage{
				{
					RawPayload: []byte{1, 2, 3, 4},
					Settings:   &ttnpb.TxSettings{Frequency: 1200000},
				},
				{
					RawPayload: []byte{1, 2, 3, 4, 5},
					Settings:   &ttnpb.TxSettings{Frequency: 1000000},
					RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -20, AntennaIndex: 0}},
				},
				{
					RawPayload: []byte{1, 2, 3, 4, 5, 6},
					Settings:   &ttnpb.TxSettings{Frequency: 1000000},
					RxMetadata: []*ttnpb.RxMetadata{{Snr: 10, Rssi: -10, AntennaIndex: 0}},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			assertions.New(t).So(io.UniqueUplinkMessagesByRSSI(tc.in), should.Resemble, tc.out)
		})
	}
}
