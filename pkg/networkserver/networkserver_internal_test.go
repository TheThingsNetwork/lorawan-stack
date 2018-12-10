// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

var _ ttnpb.NsGsClient = &MockNsGsClient{}

type MockNsGsClient struct {
	*test.MockClientStream
	ScheduleDownlinkFunc func(ctx context.Context, in *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error)
}

func (cl *MockNsGsClient) ScheduleDownlink(ctx context.Context, in *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
	if cl.ScheduleDownlinkFunc == nil {
		return nil, nil
	}
	return cl.ScheduleDownlinkFunc(ctx, in, opts...)
}

func TestProcessDownlinkTask(t *testing.T) {
	gateways := [...]ttnpb.GatewayIdentifiers{
		{
			GatewayID: "test-gtw-0",
		},
		{
			GatewayID: "test-gtw-1",
		},
		{
			GatewayID: "test-gtw-2",
		},
		{
			GatewayID: "test-gtw-3",
		},
	}

	// NOTE: This is only valid under assumption that test.EUFrequencyPlanID uses 868,
	// and that all devices in test cases use test.EUFrequencyPlanID as the frequency plan.
	band := test.Must(band.GetByID(band.EU_863_870)).(band.Band)

	channels := [16]*ttnpb.MACParameters_Channel{
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
		ttnpb.NewPopulatedMACParameters_Channel(test.Randy, false),
	}

	rx1Downlink := func(b []byte, chIdx uint32, drIdx ttnpb.DataRateIndex, drOffset uint32, dwellTime bool, md ttnpb.TxMetadata, correlationIDs ...string) *ttnpb.DownlinkMessage {
		chIdx = test.Must(band.Rx1Channel(chIdx)).(uint32)
		drIdx = test.Must(band.Rx1DataRate(drIdx, drOffset, dwellTime)).(ttnpb.DataRateIndex)

		msg := &ttnpb.DownlinkMessage{
			RawPayload:     b,
			CorrelationIDs: correlationIDs,
			EndDeviceIDs: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
				DeviceID:               DeviceID,
				DevAddr:                &DevAddr,
			},
			Settings: ttnpb.TxSettings{
				DataRateIndex:      drIdx,
				CodingRate:         "4/5",
				InvertPolarization: true,
				ChannelIndex:       chIdx,
				Frequency:          channels[int(chIdx)].DownlinkFrequency,
				TxPower:            int32(band.DefaultMaxEIRP),
			},
			TxMetadata: md,
		}
		test.Must(nil, setDownlinkModulation(&msg.Settings, band.DataRates[int(drIdx)]))
		return msg
	}

	rx2Downlink := func(b []byte, freq uint64, drIdx ttnpb.DataRateIndex, md ttnpb.TxMetadata, correlationIDs ...string) *ttnpb.DownlinkMessage {
		msg := &ttnpb.DownlinkMessage{
			RawPayload:     b,
			CorrelationIDs: correlationIDs,
			EndDeviceIDs: &ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
				DeviceID:               DeviceID,
				DevAddr:                &DevAddr,
			},
			Settings: ttnpb.TxSettings{
				DataRateIndex:      drIdx,
				CodingRate:         "4/5",
				InvertPolarization: true,
				Frequency:          freq,
				TxPower:            int32(band.DefaultMaxEIRP),
			},
			TxMetadata: md,
		}
		test.Must(nil, setDownlinkModulation(&msg.Settings, band.DataRates[int(drIdx)]))
		return msg
	}

	type nsKey struct{}
	type deviceKey struct{}
	type popCallKey struct{}
	type setByIDCallKey struct{}
	type nsGsClientCallKey struct{}
	type scheduleDownlinkCallKey struct{}

	for _, tc := range []struct {
		Name             string
		ContextFunc      func(context.Context) context.Context
		NsGsClient       NsGsClientFunc
		PopFunc          func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
		SetByIDFunc      func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
		ContextAssertion func(ctx context.Context) bool
		ErrorAssertion   func(t *testing.T, err error) bool
	}{
		{
			Name: "1.1/Rx1/application downlink/no ADR/no uplink dwell time/no downlink dwell time",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Channels:          channels[:],
						},
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{{
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: gateways[0],
								Timestamp:          123,
								SNR:                8.1,
							},
							{
								GatewayIdentifiers: gateways[1],
								Timestamp:          124,
								SNR:                4,
							},
							{
								GatewayIdentifiers: gateways[2],
								Timestamp:          42,
								SNR:                -1,
							},
						},
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
						Settings: ttnpb.TxSettings{
							DataRateIndex:      ttnpb.DATA_RATE_0,
							CodingRate:         "4/5",
							InvertPolarization: false,
							ChannelIndex:       3,
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   "testKeyID",
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationID1", "testCorrelationID2"},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, time.Now())
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.Component.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)

				expected := CopyEndDevice(pb)
				b, err := GenerateDownlink(ctx, expected,
					band.DataRates[test.Must(band.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
					band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate downlink: '%v'", errors.Stack(err))
				}

				expected.RecentDownlinks = append(expected.RecentDownlinks,
					rx1Downlink(b, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(3 * time.Second).UnixNano()),
					}),
				)

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret.RecentDownlinks, should.HaveLength, 1) || !a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 1) {
					t.FailNow()
				}
				expected.RecentDownlinks[0].CorrelationIDs = ret.RecentDownlinks[0].CorrelationIDs
				a.So(ret, should.Resemble, expected)
				return ret, nil
			},

			NsGsClient: func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, nsGsClientCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.Component.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)

				switch uid := unique.ID(ctx, id); uid {
				case unique.ID(ctx, gateways[0]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 0)
					return nil, fmt.Errorf("`%s` gsClient error", uid)

				case unique.ID(ctx, gateways[1]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 1)
					return &MockNsGsClient{
						ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
							t := test.MustTFromContext(ctx)
							a := assertions.New(t)

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)

							b, err := GenerateDownlink(ctx, CopyEndDevice(pb),
								band.DataRates[test.Must(band.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
								band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
							)
							if !a.So(err, should.BeNil) {
								t.Fatalf("Failed to generate downlink: %s", err)
							}

							a.So(msg.CorrelationIDs, should.NotBeEmpty)
							a.So(msg, should.Resemble, rx1Downlink(b, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
								GatewayIdentifiers: gateways[1],
								Timestamp:          uint64(time.Unix(0, 124).Add(3 * time.Second).UnixNano()),
							}, msg.CorrelationIDs...))
							a.So(opts, should.Contain, ns.WithClusterAuth())
							return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
						},
					}, nil

				case unique.ID(ctx, gateways[2]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 2)
					return &MockNsGsClient{
						ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
							t := test.MustTFromContext(ctx)
							a := assertions.New(t)

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)

							b, err := GenerateDownlink(ctx, CopyEndDevice(pb),
								band.DataRates[test.Must(band.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
								band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
							)
							if !a.So(err, should.BeNil) {
								t.Fatalf("Failed to generate downlink: %s", err)
							}

							a.So(msg.CorrelationIDs, should.NotBeEmpty)
							a.So(msg, should.Resemble, rx1Downlink(b, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
								GatewayIdentifiers: gateways[2],
								Timestamp:          uint64(time.Unix(0, 42).Add(3 * time.Second).UnixNano()),
							}, msg.CorrelationIDs...))
							a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
							return ttnpb.Empty, nil
						},
					}, nil

				default:
					t.Errorf("Unknown gateway `%s` requested", uid)
				}
				return nil, nil
			},

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 3) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
			},
		},

		{
			Name: "1.1/Rx2/application downlink/no ADR/no uplink dwell time/no downlink dwell time",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						LoRaWANVersion: ttnpb.MAC_V1_1,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{{
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: gateways[0],
								Timestamp:          123,
								SNR:                8.1,
							},
							{
								GatewayIdentifiers: gateways[1],
								Timestamp:          124,
								SNR:                4,
							},
							{
								GatewayIdentifiers: gateways[2],
								Timestamp:          42,
								SNR:                -1,
							},
						},
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_UNCONFIRMED_UP,
							},
							Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
						},
						Settings: ttnpb.TxSettings{
							DataRateIndex:      ttnpb.DATA_RATE_0,
							CodingRate:         "4/5",
							InvertPolarization: false,
							ChannelIndex:       3,
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   "testKeyID",
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationID1", "testCorrelationID2"},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, time.Now())
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}
				fp := test.Must(ns.Component.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)

				expected := CopyEndDevice(pb)
				b, err := GenerateDownlink(ctx, expected,
					band.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
					band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
				)
				if !a.So(err, should.BeNil) {
					t.Fatalf("Failed to generate downlink: %s", err)
				}

				expected.RecentDownlinks = append(expected.RecentDownlinks,
					rx2Downlink(b, 42, ttnpb.DATA_RATE_1, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(3*time.Second + time.Second).UnixNano()),
					}),
				)

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret.RecentDownlinks, should.HaveLength, 1) || !a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 1) {
					t.FailNow()
				}
				expected.RecentDownlinks[0].CorrelationIDs = ret.RecentDownlinks[0].CorrelationIDs
				a.So(ret, should.Resemble, expected)
				return ret, nil
			},

			NsGsClient: func() func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				var rx2 bool
				return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, nsGsClientCallKey{}, 1)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}
					fp := test.Must(ns.Component.FrequencyPlans.GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)

					switch uid := unique.ID(ctx, id); uid {
					case unique.ID(ctx, gateways[0]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 0)
						} else {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 3)
						}
						return nil, fmt.Errorf("`%s` gsClient error", uid)

					case unique.ID(ctx, gateways[1]):
						if rx2 {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 4)
							return &MockNsGsClient{
								ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
									t := test.MustTFromContext(ctx)
									a := assertions.New(t)

									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)

									b, err := GenerateDownlink(ctx, CopyEndDevice(pb),
										band.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
										band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
									)
									if !a.So(err, should.BeNil) {
										t.Fatalf("Failed to generate downlink: %s", err)
									}

									a.So(msg.CorrelationIDs, should.NotBeEmpty)
									a.So(msg, should.Resemble, rx2Downlink(b, 42, ttnpb.DATA_RATE_1, ttnpb.TxMetadata{
										GatewayIdentifiers: gateways[1],
										Timestamp:          uint64(time.Unix(0, 124).Add(4 * time.Second).UnixNano()),
									}, msg.CorrelationIDs...))
									a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
									return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
								},
							}, nil
						}

						a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 1)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)

								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)

								b, err := GenerateDownlink(ctx, CopyEndDevice(pb),
									band.DataRates[test.Must(band.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
									band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
								)
								if !a.So(err, should.BeNil) {
									t.Fatalf("Failed to generate downlink: %s", err)
								}

								a.So(msg.CorrelationIDs, should.NotBeEmpty)
								a.So(msg, should.Resemble, rx1Downlink(b, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
									GatewayIdentifiers: gateways[1],
									Timestamp:          uint64(time.Unix(0, 124).Add(3 * time.Second).UnixNano()),
								}, msg.CorrelationIDs...))
								a.So(opts, should.Contain, ns.WithClusterAuth())
								return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
							},
						}, nil

					case unique.ID(ctx, gateways[2]):
						defer func() { rx2 = true }()

						if rx2 {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 5)
							return &MockNsGsClient{
								ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
									t := test.MustTFromContext(ctx)
									a := assertions.New(t)

									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)

									b, err := GenerateDownlink(ctx, CopyEndDevice(pb),
										band.DataRates[ttnpb.DATA_RATE_1].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
										band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
									)
									if !a.So(err, should.BeNil) {
										t.Fatalf("Failed to generate downlink: %s", err)
									}

									a.So(msg.CorrelationIDs, should.NotBeEmpty)
									a.So(msg, should.Resemble, rx2Downlink(b, 42, ttnpb.DATA_RATE_1, ttnpb.TxMetadata{
										GatewayIdentifiers: gateways[2],
										Timestamp:          uint64(time.Unix(0, 42).Add(4 * time.Second).UnixNano()),
									}, msg.CorrelationIDs...))
									a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
									return ttnpb.Empty, nil
								},
							}, nil
						}

						a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 2)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)

								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)

								b, err := GenerateDownlink(ctx, CopyEndDevice(pb),
									band.DataRates[test.Must(band.Rx1DataRate(ttnpb.DATA_RATE_0, 2, false)).(ttnpb.DataRateIndex)].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetDownlinks()),
									band.DataRates[ttnpb.DATA_RATE_0].DefaultMaxSize.PayloadSize(true, fp.DwellTime.GetUplinks()),
								)
								if !a.So(err, should.BeNil) {
									t.Fatalf("Failed to generate downlink: %s", err)
								}

								a.So(msg.CorrelationIDs, should.NotBeEmpty)
								a.So(msg, should.Resemble, rx1Downlink(b, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
									GatewayIdentifiers: gateways[2],
									Timestamp:          uint64(time.Unix(0, 42).Add(3 * time.Second).UnixNano()),
								}, msg.CorrelationIDs...))
								a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
								return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
							},
						}, nil

					default:
						t.Errorf("Unknown gateway `%s` requested", uid)
					}
					return nil, nil
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 6) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 4)
			},
		},

		{
			Name: "1.1/Rx1/join accept/non-empty application downlink queue",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Channels:          channels[:],
						},
						QueuedJoinAccept: []byte("testJoinAccept"),
						LoRaWANVersion:   ttnpb.MAC_V1_1,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{{
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: gateways[0],
								Timestamp:          123,
								SNR:                8.1,
							},
							{
								GatewayIdentifiers: gateways[1],
								Timestamp:          124,
								SNR:                4,
							},
							{
								GatewayIdentifiers: gateways[2],
								Timestamp:          42,
								SNR:                -1,
							},
						},
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_JOIN_REQUEST,
							},
							Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{}},
						},
						Settings: ttnpb.TxSettings{
							DataRateIndex:      ttnpb.DATA_RATE_0,
							CodingRate:         "4/5",
							InvertPolarization: false,
							ChannelIndex:       3,
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   "testKeyID",
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationID1", "testCorrelationID2"},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, time.Now())
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				expected := CopyEndDevice(pb)
				expected.MACState.QueuedJoinAccept = nil
				expected.RecentDownlinks = append(expected.RecentDownlinks,
					rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay1).UnixNano()),
					}),
				)

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret.RecentDownlinks, should.HaveLength, 1) || !a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 1) {
					t.FailNow()
				}
				expected.RecentDownlinks[0].CorrelationIDs = ret.RecentDownlinks[0].CorrelationIDs
				a.So(ret, should.Resemble, expected)
				return ret, nil
			},

			NsGsClient: func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, nsGsClientCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				switch uid := unique.ID(ctx, id); uid {
				case unique.ID(ctx, gateways[0]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 0)
					return nil, fmt.Errorf("`%s` gsClient error", uid)

				case unique.ID(ctx, gateways[1]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 1)
					return &MockNsGsClient{
						ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
							t := test.MustTFromContext(ctx)
							a := assertions.New(t)

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)

							a.So(msg.CorrelationIDs, should.NotBeEmpty)
							a.So(msg, should.Resemble, rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
								GatewayIdentifiers: gateways[1],
								Timestamp:          uint64(time.Unix(0, 124).Add(band.JoinAcceptDelay1).UnixNano()),
							}, msg.CorrelationIDs...))
							a.So(opts, should.Contain, ns.WithClusterAuth())
							return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
						},
					}, nil

				case unique.ID(ctx, gateways[2]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 2)
					return &MockNsGsClient{
						ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
							t := test.MustTFromContext(ctx)
							a := assertions.New(t)

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)

							a.So(msg.CorrelationIDs, should.NotBeEmpty)
							a.So(msg, should.Resemble, rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
								GatewayIdentifiers: gateways[2],
								Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay1).UnixNano()),
							}, msg.CorrelationIDs...))
							a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
							return ttnpb.Empty, nil
						},
					}, nil

				default:
					t.Errorf("Unknown gateway `%s` requested", uid)
				}
				return nil, nil
			},

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 3) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
			},
		},

		{
			Name: "1.1/Rx1/join accept/empty application downlink queue",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Channels:          channels[:],
						},
						QueuedJoinAccept: []byte("testJoinAccept"),
						LoRaWANVersion:   ttnpb.MAC_V1_1,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{{
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: gateways[0],
								Timestamp:          123,
								SNR:                8.1,
							},
							{
								GatewayIdentifiers: gateways[1],
								Timestamp:          124,
								SNR:                4,
							},
							{
								GatewayIdentifiers: gateways[2],
								Timestamp:          42,
								SNR:                -1,
							},
						},
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_JOIN_REQUEST,
							},
							Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{}},
						},
						Settings: ttnpb.TxSettings{
							DataRateIndex:      ttnpb.DATA_RATE_0,
							CodingRate:         "4/5",
							InvertPolarization: false,
							ChannelIndex:       3,
						},
					}},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, time.Now())
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				expected := CopyEndDevice(pb)
				expected.MACState.QueuedJoinAccept = nil
				expected.RecentDownlinks = append(expected.RecentDownlinks,
					rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay1).UnixNano()),
					}),
				)

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret.RecentDownlinks, should.HaveLength, 1) || !a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 1) {
					t.FailNow()
				}
				expected.RecentDownlinks[0].CorrelationIDs = ret.RecentDownlinks[0].CorrelationIDs
				a.So(ret, should.Resemble, expected)
				return ret, nil
			},

			NsGsClient: func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, nsGsClientCallKey{}, 1)

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				switch uid := unique.ID(ctx, id); uid {
				case unique.ID(ctx, gateways[0]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 0)
					return nil, fmt.Errorf("`%s` gsClient error", uid)

				case unique.ID(ctx, gateways[1]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 1)
					return &MockNsGsClient{
						ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
							t := test.MustTFromContext(ctx)
							a := assertions.New(t)

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)

							a.So(msg.CorrelationIDs, should.NotBeEmpty)
							a.So(msg, should.Resemble, rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
								GatewayIdentifiers: gateways[1],
								Timestamp:          uint64(time.Unix(0, 124).Add(band.JoinAcceptDelay1).UnixNano()),
							}, msg.CorrelationIDs...))
							a.So(opts, should.Contain, ns.WithClusterAuth())
							return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
						},
					}, nil

				case unique.ID(ctx, gateways[2]):
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 2)
					return &MockNsGsClient{
						ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
							t := test.MustTFromContext(ctx)
							a := assertions.New(t)

							defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

							a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)

							a.So(msg.CorrelationIDs, should.NotBeEmpty)
							a.So(msg, should.Resemble, rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
								GatewayIdentifiers: gateways[2],
								Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay1).UnixNano()),
							}, msg.CorrelationIDs...))
							a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
							return ttnpb.Empty, nil
						},
					}, nil

				default:
					t.Errorf("Unknown gateway `%s` requested", uid)
				}
				return nil, nil
			},

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 3) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)
			},
		},

		{
			Name: "1.1/Rx2/join accept/non-empty application downlink queue",
			ContextFunc: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, deviceKey{}, &ttnpb.EndDevice{
					FrequencyPlanID: test.EUFrequencyPlanID,
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
						DeviceID:               DeviceID,
						DevAddr:                &DevAddr,
					},
					Session: &ttnpb.Session{
						DevAddr: DevAddr,
						SessionKeys: ttnpb.SessionKeys{
							FNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: FNwkSIntKey[:],
							},
							SNwkSIntKey: &ttnpb.KeyEnvelope{
								Key: SNwkSIntKey[:],
							},
							NwkSEncKey: &ttnpb.KeyEnvelope{
								Key: NwkSEncKey[:],
							},
						},
					},
					MACState: &ttnpb.MACState{
						CurrentParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						DesiredParameters: ttnpb.MACParameters{
							Rx1Delay:          3,
							Rx1DataRateOffset: 2,
							Rx2DataRateIndex:  ttnpb.DATA_RATE_1,
							Rx2Frequency:      42,
							Channels:          channels[:],
						},
						QueuedJoinAccept: []byte("testJoinAccept"),
						LoRaWANVersion:   ttnpb.MAC_V1_1,
					},
					RecentUplinks: []*ttnpb.UplinkMessage{{
						RxMetadata: []*ttnpb.RxMetadata{
							{
								GatewayIdentifiers: gateways[0],
								Timestamp:          123,
								SNR:                8.1,
							},
							{
								GatewayIdentifiers: gateways[1],
								Timestamp:          124,
								SNR:                4,
							},
							{
								GatewayIdentifiers: gateways[2],
								Timestamp:          42,
								SNR:                -1,
							},
						},
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_JOIN_REQUEST,
							},
							Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{}},
						},
						Settings: ttnpb.TxSettings{
							DataRateIndex:      ttnpb.DATA_RATE_0,
							CodingRate:         "4/5",
							InvertPolarization: false,
							ChannelIndex:       3,
						},
					}},
					QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
						{
							SessionKeyID:   "testKeyID",
							FPort:          1,
							FCnt:           42,
							FRMPayload:     []byte("testPayload"),
							CorrelationIDs: []string{"testCorrelationID1", "testCorrelationID2"},
						},
					},
				})
			},

			PopFunc: func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
				t, ok := test.TFromContext(ctx)
				if !ok {
					// This is the Pop called by the cluster, block until test is done or the time limit exceeded
					<-ctx.Done()
					return ctx.Err()
				}
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, popCallKey{}, 1)

				err := f(ctx, ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
				}, time.Now())
				a.So(err, should.BeNil)

				return nil
			},

			SetByIDFunc: func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
				t := test.MustTFromContext(ctx)
				a := assertions.New(t)

				defer test.MustIncrementContextCounter(ctx, setByIDCallKey{}, 1)

				a.So(appID, should.Resemble, ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID})
				a.So(devID, should.Equal, DeviceID)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"frequency_plan_id",
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"recent_uplinks",
					"session",
				})

				pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
				if !a.So(ok, should.BeTrue) {
					t.Fatal("Invalid context")
				}

				expected := CopyEndDevice(pb)
				expected.MACState.QueuedJoinAccept = nil
				expected.RecentDownlinks = append(expected.RecentDownlinks,
					rx2Downlink(pb.MACState.QueuedJoinAccept, 42, ttnpb.DATA_RATE_1, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay2).UnixNano()),
					}),
				)

				ret, paths, err := f(CopyEndDevice(pb))
				a.So(err, should.BeNil)
				a.So(paths, should.HaveSameElementsDeep, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				})
				if !a.So(ret.RecentDownlinks, should.HaveLength, 1) || !a.So(ret.RecentDownlinks[0].CorrelationIDs, should.HaveLength, 1) {
					t.FailNow()
				}
				expected.RecentDownlinks[0].CorrelationIDs = ret.RecentDownlinks[0].CorrelationIDs
				a.So(ret, should.Resemble, expected)
				return ret, nil
			},

			NsGsClient: func() func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				var rx2 bool
				return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
					t := test.MustTFromContext(ctx)
					a := assertions.New(t)

					defer test.MustIncrementContextCounter(ctx, nsGsClientCallKey{}, 1)

					pb, ok := ctx.Value(deviceKey{}).(*ttnpb.EndDevice)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					ns, ok := ctx.Value(nsKey{}).(*NetworkServer)
					if !a.So(ok, should.BeTrue) {
						t.Fatal("Invalid context")
					}

					switch uid := unique.ID(ctx, id); uid {
					case unique.ID(ctx, gateways[0]):
						if !rx2 {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 0)
						} else {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 3)
						}
						return nil, fmt.Errorf("`%s` gsClient error", uid)

					case unique.ID(ctx, gateways[1]):
						if rx2 {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 4)
							return &MockNsGsClient{
								ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
									t := test.MustTFromContext(ctx)
									a := assertions.New(t)

									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 2)

									a.So(msg.CorrelationIDs, should.NotBeEmpty)
									a.So(msg, should.Resemble, rx2Downlink(pb.MACState.QueuedJoinAccept, 42, ttnpb.DATA_RATE_1, ttnpb.TxMetadata{
										GatewayIdentifiers: gateways[1],
										Timestamp:          uint64(time.Unix(0, 124).Add(band.JoinAcceptDelay2).UnixNano()),
									}, msg.CorrelationIDs...))
									a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
									return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
								},
							}, nil
						}

						a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 1)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)

								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 0)

								a.So(msg.CorrelationIDs, should.NotBeEmpty)
								a.So(msg, should.Resemble, rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
									GatewayIdentifiers: gateways[1],
									Timestamp:          uint64(time.Unix(0, 124).Add(band.JoinAcceptDelay1).UnixNano()),
								}, msg.CorrelationIDs...))
								a.So(opts, should.Contain, ns.WithClusterAuth())
								return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
							},
						}, nil

					case unique.ID(ctx, gateways[2]):
						defer func() { rx2 = true }()

						if rx2 {
							a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 5)
							return &MockNsGsClient{
								ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
									t := test.MustTFromContext(ctx)
									a := assertions.New(t)

									defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

									a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 3)

									a.So(msg.CorrelationIDs, should.NotBeEmpty)
									a.So(msg, should.Resemble, rx2Downlink(pb.MACState.QueuedJoinAccept, 42, ttnpb.DATA_RATE_1, ttnpb.TxMetadata{
										GatewayIdentifiers: gateways[2],
										Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay2).UnixNano()),
									}, msg.CorrelationIDs...))
									a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
									return ttnpb.Empty, nil
								},
							}, nil
						}

						a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 2)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								t := test.MustTFromContext(ctx)
								a := assertions.New(t)

								defer test.MustIncrementContextCounter(ctx, scheduleDownlinkCallKey{}, 1)

								a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 1)

								a.So(msg.CorrelationIDs, should.NotBeEmpty)
								a.So(msg, should.Resemble, rx1Downlink(pb.MACState.QueuedJoinAccept, 3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
									GatewayIdentifiers: gateways[2],
									Timestamp:          uint64(time.Unix(0, 42).Add(band.JoinAcceptDelay1).UnixNano()),
								}, msg.CorrelationIDs...))
								a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
								return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
							},
						}, nil

					default:
						t.Errorf("Unknown gateway `%s` requested", uid)
					}
					return nil, nil
				}
			}(),

			ContextAssertion: func(ctx context.Context) bool {
				a := assertions.New(test.MustTFromContext(ctx))
				return a.So(test.MustCounterFromContext(ctx, popCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, setByIDCallKey{}), should.Equal, 1) &&
					a.So(test.MustCounterFromContext(ctx, nsGsClientCallKey{}), should.Equal, 6) &&
					a.So(test.MustCounterFromContext(ctx, scheduleDownlinkCallKey{}), should.Equal, 4)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t),
					&component.Config{},
				),
				&Config{
					DeduplicationWindow: 42,
					CooldownWindow:      42,
					DownlinkTasks: &MockDownlinkTaskQueue{
						PopFunc: tc.PopFunc,
					},
					Devices: &MockDeviceRegistry{
						SetByIDFunc: tc.SetByIDFunc,
					},
				},
				WithNsGsClientFunc(tc.NsGsClient),
			)).(*NetworkServer)
			ns.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
			ns.Component.AddContextFiller(tc.ContextFunc)
			ns.Component.AddContextFiller(func(ctx context.Context) context.Context {
				return context.WithValue(ctx, nsKey{}, ns)
			})
			ns.Component.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, popCallKey{})
			})
			ns.Component.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, setByIDCallKey{})
			})
			ns.Component.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, nsGsClientCallKey{})
			})
			ns.Component.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithCounter(ctx, scheduleDownlinkCallKey{})
			})
			ns.Component.AddContextFiller(func(ctx context.Context) context.Context {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
				_ = cancel
				return ctx
			})
			test.Must(nil, ns.Start())
			defer ns.Close()

			ctx := test.ContextWithT(ns.FillContext(ns.Context()), t)

			err := ns.processDownlinkTask(ctx)
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(tc.ContextAssertion(ctx), should.BeTrue)
		})
	}
}

func TestGenerateDownlink(t *testing.T) {
	encodeMessage := func(msg *ttnpb.Message, ver ttnpb.MACVersion, confFCnt uint32) []byte {
		msg = deepcopy.Copy(msg).(*ttnpb.Message)
		mac := msg.GetMACPayload()

		if len(mac.FRMPayload) > 0 && mac.FPort == 0 {
			var key types.AES128Key
			switch ver {
			case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
				key = FNwkSIntKey
			case ttnpb.MAC_V1_1:
				key = NwkSEncKey
			default:
				panic(fmt.Errorf("unknown version %s", ver))
			}

			var err error
			mac.FRMPayload, err = crypto.EncryptDownlink(key, mac.DevAddr, mac.FCnt, mac.FRMPayload)
			if err != nil {
				t.Fatal("Failed to encrypt downlink FRMPayload")
			}
		}

		b, err := lorawan.MarshalMessage(*msg)
		if err != nil {
			t.Fatal("Failed to marshal downlink")
		}

		var key types.AES128Key
		switch ver {
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1, ttnpb.MAC_V1_0_2:
			key = FNwkSIntKey
		case ttnpb.MAC_V1_1:
			key = SNwkSIntKey
		default:
			panic(fmt.Errorf("unknown version %s", ver))
		}

		mic, err := crypto.ComputeDownlinkMIC(key, mac.DevAddr, confFCnt, b)
		if err != nil {
			t.Fatal("Failed to compute MIC")
		}
		return append(b, mic[:]...)
	}

	encodeMAC := func(cmds ...*ttnpb.MACCommand) (b []byte) {
		for _, cmd := range cmds {
			b = test.Must(lorawan.DefaultMACCommands.AppendDownlink(b, *cmd)).([]byte)
		}
		return
	}

	for _, tc := range []struct {
		Name       string
		Device     *ttnpb.EndDevice
		Context    context.Context
		Bytes      []byte
		Error      error
		DeviceDiff func(*ttnpb.EndDevice)
	}{
		{
			Name:    "1.1/no app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: *ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
				MACSettings:          &ttnpb.MACSettings{},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session:                 ttnpb.NewPopulatedSession(test.Randy, false),
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes:      nil,
			Error:      errNoDownlink,
			DeviceDiff: nil,
		},
		{
			Name:    "1.1/no app downlink/status after 1 downlink/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: *ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: 3,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					LastDevStatusFCntUp: 2,
				},
				Session: &ttnpb.Session{
					LastFCntUp: 4,
				},
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes:      nil,
			Error:      errNoDownlink,
			DeviceDiff: nil,
		},
		{
			Name:    "1.1/no app downlink/status after an hour/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: *ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: 24 * time.Hour,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				LastDevStatusReceivedAt: TimePtr(time.Now()),
				Session:                 ttnpb.NewPopulatedSession(test.Randy, false),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes:      nil,
			Error:      errNoDownlink,
			DeviceDiff: nil,
		},
		{
			Name:    "1.1/no app downlink/no MAC/ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_CONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{
							MACPayload: &ttnpb.MACPayload{
								FHDR: ttnpb.FHDR{
									FCnt: 24,
								},
							},
						},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
							},
							FCnt: 42,
						},
						FPort: 0,
					},
				},
			}, ttnpb.MAC_V1_1, 24),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.Session.LastNFCntDown++
			},
		},
		{
			Name:    "1.1/unconfirmed app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  false,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				i := len(dev.QueuedApplicationDownlinks) - 1
				dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[:i]
			},
		},
		{
			Name:    "1.1/unconfirmed app downlink/no MAC/ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  false,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_CONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{
							MACPayload: &ttnpb.MACPayload{
								FHDR: ttnpb.FHDR{
									FCnt: 24,
								},
							},
						},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 24),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				i := len(dev.QueuedApplicationDownlinks) - 1
				dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[:i]
			},
		},
		{
			Name:    "1.1/confirmed app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  true,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				i := len(dev.QueuedApplicationDownlinks) - 1
				dev.QueuedApplicationDownlinks, dev.MACState.PendingApplicationDownlink = dev.QueuedApplicationDownlinks[:i], dev.QueuedApplicationDownlinks[i]
				dev.Session.LastConfFCntDown = 42
			},
		},
		{
			Name:    "1.1/confirmed app downlink/no MAC/ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				QueuedApplicationDownlinks: []*ttnpb.ApplicationDownlink{
					{
						Confirmed:  true,
						FCnt:       42,
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
				LastDevStatusReceivedAt: TimePtr(time.Unix(42, 0)),
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_CONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{
							MACPayload: &ttnpb.MACPayload{
								FHDR: ttnpb.FHDR{
									FCnt: 24,
								},
							},
						},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_CONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: true,
							},
							FCnt: 42,
						},
						FPort:      1,
						FRMPayload: []byte("test"),
					},
				},
			}, ttnpb.MAC_V1_1, 24),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				i := len(dev.QueuedApplicationDownlinks) - 1
				dev.QueuedApplicationDownlinks, dev.MACState.PendingApplicationDownlink = dev.QueuedApplicationDownlinks[:i], dev.QueuedApplicationDownlinks[i]
				dev.Session.LastConfFCntDown = 42
			},
		},
		{
			Name:    "1.1/no app downlink/status(count)/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusCountPeriodicity: 3,
				},
				MACState: &ttnpb.MACState{
					LastDevStatusFCntUp: 4,
					LoRaWANVersion:      ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					LastFCntUp:    99,
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort: 0,
						FRMPayload: encodeMAC(
							ttnpb.CID_DEV_STATUS.MACCommand(),
						),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.PendingRequests = []*ttnpb.MACCommand{
					ttnpb.CID_DEV_STATUS.MACCommand(),
				}
				dev.Session.LastNFCntDown++
			},
		},
		{
			Name:    "1.1/no app downlink/status(time/zero time)/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: &ttnpb.MACSettings{
					StatusTimePeriodicity: time.Nanosecond,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					LastNFCntDown: 41,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: NwkSEncKey[:],
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: SNwkSIntKey[:],
						},
					},
				},
				RecentUplinks: []*ttnpb.UplinkMessage{{
					Payload: &ttnpb.Message{
						MHDR: ttnpb.MHDR{
							MType: ttnpb.MType_UNCONFIRMED_UP,
						},
						Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{}},
					},
				}},
			},
			Bytes: encodeMessage(&ttnpb.Message{
				MHDR: ttnpb.MHDR{
					MType: ttnpb.MType_UNCONFIRMED_DOWN,
					Major: ttnpb.Major_LORAWAN_R1,
				},
				Payload: &ttnpb.Message_MACPayload{
					MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: DevAddr,
							FCtrl: ttnpb.FCtrl{
								Ack: false,
							},
							FCnt: 42,
						},
						FPort: 0,
						FRMPayload: encodeMAC(
							ttnpb.CID_DEV_STATUS.MACCommand(),
						),
					},
				},
			}, ttnpb.MAC_V1_1, 0),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.MACState.PendingRequests = []*ttnpb.MACCommand{
					ttnpb.CID_DEV_STATUS.MACCommand(),
				}
				dev.Session.LastNFCntDown++
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := CopyEndDevice(tc.Device)

			b, err := generateDownlink(tc.Context, dev, math.MaxUint16, math.MaxUint16)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}

			a.So(b, should.Resemble, tc.Bytes)

			expected := CopyEndDevice(tc.Device)
			if tc.DeviceDiff != nil {
				tc.DeviceDiff(expected)
			}
			a.So(dev, should.Resemble, expected)
		})
	}
}
