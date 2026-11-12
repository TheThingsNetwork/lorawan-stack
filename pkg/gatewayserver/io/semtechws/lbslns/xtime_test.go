// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

package lbslns_test

import (
	"encoding/json"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws/lbslns"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// blockingWindowFrontend is a frontend which uses the same duty cycle style as the LoRa Basics Station frontend.
type blockingWindowFrontend struct{}

func (blockingWindowFrontend) Protocol() string            { return "ws" }
func (blockingWindowFrontend) SupportsDownlinkClaim() bool { return false }
func (blockingWindowFrontend) DutyCycleStyle() scheduling.DutyCycleStyle {
	return scheduling.DutyCycleStyleBlockingWindow
}

// TestNonExtendedXTimeRollover verifies that gateways which report only the 32-bit concentrator timestamp as
// xtime (without session and without accounting for rollovers) can receive downlinks after a rollover.
func TestNonExtendedXTimeRollover(t *testing.T) {
	t.Parallel()

	a, ctx := test.New(t)
	ctx = semtechws.NewContextWithSession(ctx, &semtechws.Session{})

	gtwIDs := &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}
	conn, err := io.NewConnection(
		ctx,
		blockingWindowFrontend{},
		&ttnpb.Gateway{
			Ids:              gtwIDs,
			FrequencyPlanId:  test.EUFrequencyPlanID,
			FrequencyPlanIds: []string{test.EUFrequencyPlanID},
		},
		test.FrequencyPlanStore,
		true,
		nil,
		&ttnpb.GatewayRemoteAddress{Ip: "127.0.0.1"},
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	formatter := lbslns.NewFormatter(time.Second)
	tokens := &io.DownlinkTokens{}

	// uplinkAndDownlink handles an uplink with the given xtime, schedules a class A downlink as a response and
	// returns the xtime of the downlink message sent to the gateway.
	uplinkAndDownlink := func(t *testing.T, xTime int64, fCnt uint) int64 {
		t.Helper()
		a, _ := test.New(t)

		now := time.Now()
		raw, err := json.Marshal(lbslns.UplinkDataFrame{
			MHdr:       0x40,
			DevAddr:    0x11223344,
			FCnt:       fCnt,
			FPort:      102,
			FRMPayload: "d917d93add387645",
			MIC:        1605724433,
			RadioMetaData: lbslns.RadioMetaData{
				DataRate:  5,
				Frequency: 868100000,
				UpInfo: lbslns.UpInfo{
					RCtx:    263,
					XTime:   xTime,
					RSSI:    -117,
					SNR:     -13.15,
					GPSTime: semtechws.TimeToGPSTime(now),
				},
			},
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		if _, err := formatter.HandleUp(ctx, raw, gtwIDs, conn, now, tokens); !a.So(err, should.BeNil) {
			t.FailNow()
		}

		var up *ttnpb.GatewayUplinkMessage
		select {
		case up = <-conn.Up():
		default:
			t.Fatal("Expected uplink message")
		}

		_, _, _, err = conn.ScheduleDown(
			&ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: up.Message.RxMetadata[0].UplinkToken,
				},
			},
			&ttnpb.DownlinkMessage{
				RawPayload: []byte{0x60, 0x44, 0x33, 0x22, 0x11, 0x00, 0x00, 0x00},
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
									CodingRate:      band.Cr4_5,
								},
							},
						},
						Rx1Frequency:    868100000,
						FrequencyPlanId: test.EUFrequencyPlanID,
					},
				},
				CorrelationIds: []string{"correlation"},
			},
		)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		var down *ttnpb.DownlinkMessage
		select {
		case down = <-conn.Down():
		default:
			t.Fatal("Expected downlink message")
		}
		raw, err = formatter.FromDownlink(ctx, down, band.EU_863_870, time.Now(), tokens)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		var dnmsg lbslns.DownlinkMessage
		if !a.So(json.Unmarshal(raw, &dnmsg), should.BeNil) || !a.So(dnmsg.TimestampDownlinkMessage, should.NotBeNil) {
			t.FailNow()
		}
		a.So(dnmsg.RxDelay, should.Equal, 1)
		return dnmsg.XTime
	}

	// Uplink and downlink before the concentrator timestamp rolls over.
	a.So(uplinkAndDownlink(t, 1<<32-60000000, 1), should.Equal, 1<<32-60000000)
	// Uplink and downlink after the concentrator timestamp rolled over.
	// The downlink must not be blocked by the downlink before the rollover, and the xtime must be the 32-bit
	// concentrator timestamp of the uplink.
	a.So(uplinkAndDownlink(t, 780557, 2), should.Equal, 780557)
}
