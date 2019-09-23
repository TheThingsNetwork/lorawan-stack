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
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var timeout = (1 << 3) * test.Delay

func timePtr(t time.Time) *time.Time { return &t }

func Test(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	gs := mock.NewServer(c)

	ids := ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}
	antennaGain := float32(3)
	gtw := &ttnpb.Gateway{
		GatewayIdentifiers: ids,
		FrequencyPlanID:    "EU_863_870",
		Antennas: []ttnpb.GatewayAntenna{
			{
				Gain: antennaGain,
			},
		},
	}
	gs.RegisterGateway(ctx, ids, gtw)

	gtwCtx := rights.NewContext(ctx, rights.Rights{
		GatewayRights: map[string]*ttnpb.Rights{
			unique.ID(ctx, ids): ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_LINK),
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

	{
		frontend.Up <- &ttnpb.UplinkMessage{
			RxMetadata: []*ttnpb.RxMetadata{
				{
					AntennaIndex: 0,
					Timestamp:    100,
				},
			},
		}
		select {
		case up := <-conn.Up():
			tokenIDs, timestamp, err := io.ParseUplinkToken(up.RxMetadata[0].UplinkToken)
			a.So(err, should.BeNil)
			a.So(tokenIDs.GatewayIdentifiers, should.Resemble, ids)
			a.So(tokenIDs.AntennaIndex, should.Equal, 0)
			a.So(timestamp, should.Equal, 100)
		case <-time.After(timeout):
			t.Fatalf("Expected uplink message time-out")
		}
		total, t, ok := conn.UpStats()
		a.So(ok, should.BeTrue)
		a.So(total, should.Equal, 1)
		a.So(time.Since(t), should.BeLessThan, timeout)
	}

	{
		frontend.Status <- &ttnpb.GatewayStatus{}
		select {
		case <-conn.Status():
		case <-time.After(timeout):
			t.Fatalf("Expected status message time-out")
		}
		last, t, ok := conn.StatusStats()
		a.So(ok, should.BeTrue)
		a.So(last, should.NotBeNil)
		a.So(time.Since(t), should.BeLessThan, timeout)
	}

	{
		frontend.TxAck <- &ttnpb.TxAcknowledgment{}
		select {
		case <-conn.TxAck():
		case <-time.After(timeout):
			t.Fatalf("Expected Tx acknowledgement time-out")
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
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}}, 100),
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
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}}, 100),
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_A,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx1DataRateIndex: 5,
						Rx1Frequency:     868100000,
					},
				},
			},
			ExpectedEIRP: 16.15 - antennaGain,
		},
		{
			Name: "ConflictClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}}, 100), // Same as previous.
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_A,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx1DataRateIndex: 5,         // Same as previous.
						Rx1Frequency:     868100000, // Same as previous.
					},
				},
			},
			ErrorAssertion: errors.IsAborted,
			RxErrorAssertion: []func(error) bool{
				errors.IsResourceExhausted,  // Rx1 conflicts with previous.
				errors.IsFailedPrecondition, // Rx2 not provided.
			},
		},
		{
			Name: "NoUplinkTokenClassA",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_Fixed{
					Fixed: &ttnpb.GatewayAntennaIdentifiers{
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "foo-gateway",
						},
					},
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_A,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx1DataRateIndex: 5,         // Same as previous.
						Rx1Frequency:     868100000, // Same as previous.
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "ValidClassC/UplinkToken",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}}, 100),
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_C,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRateIndex: 5,
						Rx2Frequency:     869525000,
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
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "foo-gateway",
						},
					},
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_C,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRateIndex: 5,
						Rx2Frequency:     869525000,
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
						GatewayIdentifiers: ttnpb.GatewayIdentifiers{
							GatewayID: "foo-gateway",
						},
					},
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_C,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRateIndex: 5,
						Rx2Frequency:     869525000,
						AbsoluteTime:     timePtr(time.Unix(100, 0)), // The mock front-end uses Unix epoch as start time.
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
						Class:            ttnpb.CLASS_C,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRateIndex: 5,
						Rx2Frequency:     869525000,
						AbsoluteTime:     timePtr(time.Unix(100, 0)), // The mock front-end uses Unix epoch as start time.
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "InvalidDataRate",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}}, 100),
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: []byte{0x01},
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_C,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRateIndex: 10, // This one doesn't exist in the band.
						Rx2Frequency:     869525000,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "TooLong",
			Path: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gateway"}}, 100),
				},
			},
			Message: &ttnpb.DownlinkMessage{
				RawPayload: bytes.Repeat([]byte{0x01}, 80),
				Settings: &ttnpb.DownlinkMessage_Request{
					Request: &ttnpb.TxRequest{
						Class:            ttnpb.CLASS_C,
						Priority:         ttnpb.TxSchedulePriority_NORMAL,
						Rx2DataRateIndex: 0,
						Rx2Frequency:     869525000,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			_, err := conn.ScheduleDown(tc.Path, tc.Message)
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
