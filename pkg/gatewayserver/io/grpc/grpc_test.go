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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io"
	. "go.thethings.network/lorawan-stack/pkg/gatewayserver/io/grpc"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredGatewayID  = ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}
	registeredGatewayUID = unique.ID(test.Context(), registeredGatewayID)
	registeredGatewayKey = "test-key"

	timeout = 10 * test.Delay
)

func TestAuthentication(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx = newContextWithRightsFetcher(ctx)

	gs := mock.NewServer()
	srv := New(gs)

	eui := types.EUI64{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01}

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
			ID:  ttnpb.GatewayIdentifiers{GatewayID: "invalid-gateway"},
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  ttnpb.GatewayIdentifiers{EUI: &eui},
			Key: "invalid-key",
			OK:  false,
		},
	} {
		t.Run(fmt.Sprintf("%v:%v", tc.ID.GatewayID, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancelCtx := context.WithCancel(ctx)
			stream := &mockGtwGsLinkServerStream{
				MockServerStream: &test.MockServerStream{
					MockStream: &test.MockStream{
						ContextFunc: contextWithKey(ctx, tc.ID, tc.Key),
					},
				},
				RecvFunc: func() (*ttnpb.GatewayUp, error) {
					<-ctx.Done()
					return nil, ctx.Err()
				},
			}

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				err := srv.LinkGateway(stream)
				if tc.OK && !a.So(errors.IsCanceled(err), should.BeTrue) {
					t.Fatalf("Unexpected link error: %v", err)
				}
				if !tc.OK && !a.So(errors.IsCanceled(err), should.BeFalse) {
					t.FailNow()
				}
				wg.Done()
			}()

			cancelCtx()
			wg.Wait()
		})
	}
}

func TestTraffic(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx = newContextWithRightsFetcher(ctx)
	ctx, cancelCtx := context.WithCancel(ctx)

	gs := mock.NewServer()
	srv := New(gs)

	upCh := make(chan *ttnpb.GatewayUp)
	downCh := make(chan *ttnpb.DownlinkMessage)

	stream := &mockGtwGsLinkServerStream{
		MockServerStream: &test.MockServerStream{
			MockStream: &test.MockStream{
				ContextFunc: contextWithKey(ctx, registeredGatewayID, registeredGatewayKey),
			},
		},
		RecvFunc: func() (*ttnpb.GatewayUp, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case msg := <-upCh:
				return msg, nil
			}
		},
		SendFunc: func(msg *ttnpb.GatewayDown) error {
			downCh <- msg.DownlinkMessage
			return nil
		},
	}

	go func() {
		if err := srv.LinkGateway(stream); err != nil {
			if !a.So(errors.IsCanceled(err), should.BeTrue) {
				t.Fatalf("Expected context cancellation but got: %v", err)
			}
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
					IP: []string{"1.1.1.1"},
				},
			},
			{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x01},
					},
				},
			},
			{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x02},
					},
				},
				GatewayStatus: &ttnpb.GatewayStatus{
					IP: []string{"2.2.2.2"},
				},
			},
			{
				UplinkMessages: []*ttnpb.UplinkMessage{
					{
						RawPayload: []byte{0x03},
					},
					{
						RawPayload: []byte{0x04},
					},
					{
						RawPayload: []byte{0x05},
					},
				},
				GatewayStatus: &ttnpb.GatewayStatus{
					IP: []string{"3.3.3.3"},
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
						a.So(up, should.Resemble, tc.UplinkMessages[ups])
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
	})

	t.Run("Downstream", func(t *testing.T) {
		for i, tc := range []struct {
			Path           *ttnpb.DownlinkPath
			Message        *ttnpb.DownlinkMessage
			ErrorAssertion func(error) bool
		}{
			{
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID}, 100),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
					Settings: &ttnpb.DownlinkMessage_Request{
						Request: &ttnpb.TxRequest{
							Class:            ttnpb.CLASS_A,
							Priority:         ttnpb.TxSchedulePriority_NORMAL,
							Rx1Delay:         ttnpb.RX_DELAY_1,
							Rx1DataRateIndex: 5,
							Rx1Frequency:     868100000,
							Rx2DataRateIndex: 0,
							Rx2Frequency:     869525000,
						},
					},
				},
			},
			{
				Path: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_UplinkToken{
						UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID}, 100),
					},
				},
				Message: &ttnpb.DownlinkMessage{
					RawPayload: []byte{0x01},
					Settings: &ttnpb.DownlinkMessage_Scheduled{
						Scheduled: &ttnpb.TxSettings{
							DataRate: ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_LoRa{
									LoRa: &ttnpb.LoRaDataRate{
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
						UplinkToken: io.MustUplinkToken(ttnpb.GatewayAntennaIdentifiers{GatewayIdentifiers: registeredGatewayID}, 100),
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

				_, err := conn.SendDown(tc.Path, tc.Message)
				if err != nil && (tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue)) {
					t.Fatalf("Unexpected error: %v", err)
				}
				select {
				case down := <-downCh:
					if tc.ErrorAssertion == nil {
						a.So(down, should.Resemble, tc.Message)
					} else {
						t.Fatalf("Unexpected message: %v", down)
					}
				case <-time.After(timeout):
					if tc.ErrorAssertion == nil {
						t.Fatal("Receive expected downlink timeout")
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
	ctx = newContextWithRightsFetcher(ctx)

	gs := mock.NewServer()
	srv := New(gs)

	ctx = contextWithKey(ctx, registeredGatewayID, registeredGatewayKey)()

	_, err := srv.GetConcentratorConfig(ctx, ttnpb.Empty)
	a.So(err, should.BeNil)
}
