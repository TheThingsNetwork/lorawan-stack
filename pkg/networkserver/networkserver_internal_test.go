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
	"sync/atomic"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

// TODO(#1008) Move eventCollector to the test package
type eventCollector events.Channel

func collectEvents(name string) eventCollector {
	collectedEvents := make(events.Channel, 32)
	events.Subscribe(name, collectedEvents)
	return eventCollector(collectedEvents)
}

// Expect n events, fail the test if not received within reasonable time.
func (ch eventCollector) expect(t *testing.T, n int) []events.Event {
	collected := make([]events.Event, 0, n)
	for i := 0; i < n; i++ {
		evt := events.Channel(ch).ReceiveTimeout(10 * time.Millisecond * test.Delay)
		if evt == nil {
			t.Fatalf("Did not receive expected event %d/%d", i+1, n)
		}
		collected = append(collected, evt)
	}
	return collected
}

const (
	RecentUplinkCount = recentUplinkCount
)

var (
	ResetMACState    = resetMACState
	GenerateDownlink = generateDownlink
	ErrNoDownlink    = errNoDownlink

	FNwkSIntKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey = types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	NwkSEncKey  = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	AppSKey     = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	DevAddr       = types.DevAddr{0x42, 0x42, 0xff, 0xff}
	DevEUI        = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	JoinEUI       = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DeviceID      = "test"
	ApplicationID = "test"
)

func SetAppQueueUpdateTimeout(d time.Duration) {
	appQueueUpdateTimeout = d
}

func TestAccumulator(t *testing.T) {
	a := assertions.New(t)

	acc := newMetadataAccumulator()
	a.So(func() { acc.Add() }, should.NotPanic)

	vals := []*ttnpb.RxMetadata{
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		nil,
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
		ttnpb.NewPopulatedRxMetadata(test.Randy, false),
	}

	acc = newMetadataAccumulator(vals...)
	a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals)
	acc.Reset()
	a.So(acc.Accumulated(), should.BeEmpty)

	acc.Add(vals[0], vals[1], vals[2])
	a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals[:3])

	for i := 2; i < len(vals); i++ {
		acc.Add(vals[i])
		a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals[:i+1])
	}
	a.So(acc.Accumulated(), should.HaveSameElementsDeep, vals)

	acc.Reset()
	a.So(acc.Accumulated(), should.BeEmpty)
}

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

func TestScheduleDownlink(t *testing.T) {
	testCtx := test.ContextWithT(test.Context(), t)
	testBytes := []byte("test-payload")

	type nsKey struct{}

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

	rx1Downlink := func(chIdx uint32, drIdx ttnpb.DataRateIndex, drOffset uint32, dwellTime bool, md ttnpb.TxMetadata) *ttnpb.DownlinkMessage {
		chIdx = test.Must(band.Rx1Channel(chIdx)).(uint32)
		drIdx = test.Must(band.Rx1DataRate(drIdx, drOffset, dwellTime)).(ttnpb.DataRateIndex)

		msg := &ttnpb.DownlinkMessage{
			RawPayload: testBytes,
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
				DeviceID:               DeviceID,
				DevAddr:                &DevAddr,
			},
			Settings: ttnpb.TxSettings{
				DataRateIndex:         drIdx,
				CodingRate:            "4/5",
				PolarizationInversion: true,
				ChannelIndex:          chIdx,
				Frequency:             channels[int(chIdx)].DownlinkFrequency,
				TxPower:               int32(band.DefaultMaxEIRP), // TODO: Rename this to EIRP(https://github.com/TheThingsIndustries/lorawan-stack/issues/848)
			},
			TxMetadata: md,
		}
		test.Must(nil, setDownlinkModulation(&msg.Settings, band.DataRates[int(drIdx)]))
		return msg
	}

	rx2Downlink := func(freq uint64, drIdx ttnpb.DataRateIndex, md ttnpb.TxMetadata) *ttnpb.DownlinkMessage {
		msg := &ttnpb.DownlinkMessage{
			RawPayload: testBytes,
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
				DeviceID:               DeviceID,
				DevAddr:                &DevAddr,
			},
			Settings: ttnpb.TxSettings{
				DataRateIndex:         drIdx,
				CodingRate:            "4/5",
				PolarizationInversion: true,
				Frequency:             freq,
				TxPower:               int32(band.DefaultMaxEIRP), // TODO: Rename this to EIRP(https://github.com/TheThingsIndustries/lorawan-stack/issues/848)
			},
			TxMetadata: md,
		}
		test.Must(nil, setDownlinkModulation(&msg.Settings, band.DataRates[int(drIdx)]))
		return msg
	}

	for _, tc := range []struct {
		Name              string
		Context           context.Context
		Device            *ttnpb.EndDevice
		Uplink            *ttnpb.UplinkMessage
		Accumulator       *metadataAccumulator
		Bytes             []byte
		IsJoinAccept      bool
		DeduplicationDone WindowEndFunc
		NsGsClient        NsGsClientFunc
		Error             error
		DeviceDiff        func(*ttnpb.EndDevice)
	}{
		{
			Name:    "1.1/Rx1",
			Context: testCtx,
			Device: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					MACParameters: ttnpb.MACParameters{
						Rx1Delay:          3,
						Rx1DataRateOffset: 2,
						Channels:          channels[:],
					},
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
			},
			Uplink: &ttnpb.UplinkMessage{
				Settings: ttnpb.TxSettings{
					DataRateIndex:         ttnpb.DATA_RATE_0,
					CodingRate:            "4/5",
					PolarizationInversion: false,
					ChannelIndex:          3,
				},
			},
			Accumulator: newMetadataAccumulator(
				&ttnpb.RxMetadata{
					GatewayIdentifiers: gateways[0],
					Timestamp:          123,
					SNR:                8.1,
				},
				&ttnpb.RxMetadata{
					GatewayIdentifiers: gateways[1],
					Timestamp:          124,
					SNR:                4,
				},
				&ttnpb.RxMetadata{
					GatewayIdentifiers: gateways[2],
					Timestamp:          42,
					SNR:                -1,
				},
			),
			Bytes:        testBytes,
			IsJoinAccept: false,
			DeduplicationDone: func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ctx, should.HaveParentContext, testCtx)
				ch := make(chan time.Time, 1)
				ch <- time.Now()
				close(ch)
				return ch
			},
			NsGsClient: func() func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				var i uint32
				return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
					defer atomic.AddUint32(&i, 1)

					a := assertions.New(test.MustTFromContext(ctx))
					a.So(ctx, should.HaveParentContext, testCtx)

					switch uid := unique.ID(ctx, id); uid {
					case unique.ID(ctx, gateways[0]):
						a.So(i, should.Equal, 0)
						return nil, fmt.Errorf("`%s` gsClient error", uid)

					case unique.ID(ctx, gateways[1]):
						a.So(i, should.Equal, 1)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								a.So(ctx, should.HaveParentContext, testCtx)
								a.So(msg, should.Resemble, rx1Downlink(3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
									GatewayIdentifiers: gateways[1],
									Timestamp:          uint64(time.Unix(0, 124).Add(3 * time.Second).UnixNano()),
								}))

								a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
								return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
							},
						}, nil

					case unique.ID(ctx, gateways[2]):
						a.So(i, should.Equal, 2)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								a.So(ctx, should.HaveParentContext, testCtx)
								a.So(msg, should.Resemble, rx1Downlink(3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
									GatewayIdentifiers: gateways[2],
									Timestamp:          uint64(time.Unix(0, 42).Add(3 * time.Second).UnixNano()),
								}))
								a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
								return ttnpb.Empty, nil
							},
						}, nil

					default:
						t.Errorf("Unknown gateway `%s` requested", uid)
					}
					return nil, nil
				}
			}(),
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.RecentDownlinks = append(dev.RecentDownlinks,
					rx1Downlink(3, ttnpb.DATA_RATE_0, 2, false, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(3 * time.Second).UnixNano()),
					}),
				)
			},
		},
		{
			Name:    "1.1/Rx2",
			Context: testCtx,
			Device: &ttnpb.EndDevice{
				FrequencyPlanID: test.EUFrequencyPlanID,
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: ApplicationID},
					DeviceID:               DeviceID,
					DevAddr:                &DevAddr,
				},
				MACState: &ttnpb.MACState{
					MACParameters: ttnpb.MACParameters{
						Rx1Delay:          3,
						Rx1DataRateOffset: 2,
						Rx2DataRateIndex:  ttnpb.DATA_RATE_3,
						Rx2Frequency:      42,
						Channels:          channels[:],
					},
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
			},
			Uplink: &ttnpb.UplinkMessage{
				Settings: ttnpb.TxSettings{
					DataRateIndex:         ttnpb.DATA_RATE_0,
					CodingRate:            "4/5",
					PolarizationInversion: false,
					ChannelIndex:          3,
				},
			},
			Accumulator: newMetadataAccumulator(
				&ttnpb.RxMetadata{
					GatewayIdentifiers: gateways[0],
					Timestamp:          123,
					SNR:                8.1,
				},
				&ttnpb.RxMetadata{
					GatewayIdentifiers: gateways[1],
					Timestamp:          124,
					SNR:                4,
				},
				&ttnpb.RxMetadata{
					GatewayIdentifiers: gateways[2],
					Timestamp:          42,
					SNR:                -1,
				},
			),
			Bytes:        testBytes,
			IsJoinAccept: false,
			DeduplicationDone: func(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(ctx, should.HaveParentContext, testCtx)
				ch := make(chan time.Time, 1)
				ch <- time.Now()
				close(ch)
				return ch
			},
			NsGsClient: func() func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
				var i uint32
				var rx2 bool
				return func(ctx context.Context, id ttnpb.GatewayIdentifiers) (ttnpb.NsGsClient, error) {
					defer atomic.AddUint32(&i, 1)

					a := assertions.New(test.MustTFromContext(ctx))
					a.So(ctx, should.HaveParentContext, testCtx)

					switch uid := unique.ID(ctx, id); uid {
					case unique.ID(ctx, gateways[0]):
						if !rx2 {
							a.So(i, should.Equal, 0)
						} else {
							a.So(i, should.Equal, 3)
						}
						return nil, fmt.Errorf("`%s` gsClient error", uid)

					case unique.ID(ctx, gateways[1]):
						md := ttnpb.TxMetadata{
							GatewayIdentifiers: gateways[1],
							Timestamp:          uint64(time.Unix(0, 124).Add(3 * time.Second).UnixNano()),
						}

						if rx2 {
							a.So(i, should.Equal, 4)
							return &MockNsGsClient{
								ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
									a.So(ctx, should.HaveParentContext, testCtx)
									a.So(msg, should.Resemble, rx2Downlink(42, ttnpb.DATA_RATE_3, ttnpb.TxMetadata{
										GatewayIdentifiers: md.GatewayIdentifiers,
										Timestamp:          md.Timestamp + uint64(time.Second.Nanoseconds()),
									}))
									a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
									return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
								},
							}, nil
						}

						a.So(i, should.Equal, 1)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								a.So(ctx, should.HaveParentContext, testCtx)
								a.So(msg, should.Resemble, rx1Downlink(3, ttnpb.DATA_RATE_0, 2, false, md))
								a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
								return nil, fmt.Errorf("`%s` ScheduleDownlink error", uid)
							},
						}, nil

					case unique.ID(ctx, gateways[2]):
						md := ttnpb.TxMetadata{
							GatewayIdentifiers: gateways[2],
							Timestamp:          uint64(time.Unix(0, 42).Add(3 * time.Second).UnixNano()),
						}

						defer func() { rx2 = true }()

						if rx2 {
							a.So(i, should.Equal, 5)
							return &MockNsGsClient{
								ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
									a.So(ctx, should.HaveParentContext, testCtx)
									a.So(msg, should.Resemble, rx2Downlink(42, ttnpb.DATA_RATE_3, ttnpb.TxMetadata{
										GatewayIdentifiers: md.GatewayIdentifiers,
										Timestamp:          md.Timestamp + uint64(time.Second.Nanoseconds()),
									}))
									a.So(opts, should.Contain, ctx.Value(nsKey{}).(*NetworkServer).WithClusterAuth())
									return ttnpb.Empty, nil
								},
							}, nil
						}

						a.So(i, should.Equal, 2)
						return &MockNsGsClient{
							ScheduleDownlinkFunc: func(ctx context.Context, msg *ttnpb.DownlinkMessage, opts ...grpc.CallOption) (*pbtypes.Empty, error) {
								a.So(ctx, should.HaveParentContext, testCtx)
								a.So(msg, should.Resemble, rx1Downlink(3, ttnpb.DATA_RATE_0, 2, false, md))
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
			Error: nil,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.RecentDownlinks = append(dev.RecentDownlinks,
					rx2Downlink(42, ttnpb.DATA_RATE_3, ttnpb.TxMetadata{
						GatewayIdentifiers: gateways[2],
						Timestamp:          uint64(time.Unix(0, 42).Add(3*time.Second + time.Second).UnixNano()),
					}),
				)
			},
		},
		// TODO: Add JoinAccept schedule test(https://github.com/TheThingsIndustries/lorawan-stack/issues/979)
	} {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := test.ContextWithT(tc.Context, t)

			a := assertions.New(t)

			ns := test.Must(New(
				component.MustNew(test.GetLogger(t),
					&component.Config{},
				),
				&Config{
					Registry:            deviceregistry.New(store.NewTypedMapStoreClient(mapstore.New())),
					JoinServers:         nil,
					DeduplicationWindow: 42,
					CooldownWindow:      42,
				},
				WithNsGsClientFunc(tc.NsGsClient),
				WithDeduplicationDoneFunc(tc.DeduplicationDone),
			)).(*NetworkServer)
			test.Must(nil, ns.Start())
			ns.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher

			ctx = context.WithValue(ctx, nsKey{}, ns)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
			up := deepcopy.Copy(tc.Uplink).(*ttnpb.UplinkMessage)
			b := deepcopy.Copy(tc.Bytes).([]byte)

			err := ns.scheduleDownlink(ctx, dev, up, tc.Accumulator, b, tc.IsJoinAccept)
			if tc.Error == nil && !a.So(err, should.BeNil) ||
				tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) {
				t.FailNow()
			}

			expected := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
			if tc.DeviceDiff != nil {
				tc.DeviceDiff(expected)
			}

			a.So(dev, should.Resemble, expected)
			a.So(up, should.Resemble, tc.Uplink)
			a.So(b, should.Resemble, tc.Bytes)
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

		b, err := msg.MarshalLoRaWAN()
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
			b = test.Must(cmd.AppendLoRaWAN(b)).([]byte)
		}
		return
	}

	for _, tc := range []struct {
		Name       string
		Device     *ttnpb.EndDevice
		Context    context.Context
		Ack        bool
		ConfFCnt   uint32
		Bytes      []byte
		Error      error
		DeviceDiff func(*ttnpb.EndDevice)
	}{
		{
			Name:    "1.1/no app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: *ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: ttnpb.NewPopulatedSession(test.Randy, false),
			},
			Ack:        false,
			ConfFCnt:   0,
			Bytes:      nil,
			Error:      errNoDownlink,
			DeviceDiff: nil,
		},
		{
			Name:    "1.1/no app downlink/status after 1 downlink/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: *ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
				MACSettings: ttnpb.MACSettings{
					StatusCountPeriodicity: 3,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				NextStatusAfter: 1,
				Session:         ttnpb.NewPopulatedSession(test.Randy, false),
			},
			Ack:   false,
			Bytes: nil,
			Error: errNoDownlink,
			DeviceDiff: func(dev *ttnpb.EndDevice) {
				dev.NextStatusAfter--
			},
		},
		{
			Name:    "1.1/no app downlink/status after an hour/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: *ttnpb.NewPopulatedEndDeviceIdentifiers(test.Randy, false),
				MACSettings: ttnpb.MACSettings{
					StatusTimePeriodicity: time.Nanosecond,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				NextStatusAt: time.Now().Add(24 * time.Hour),
				Session:      ttnpb.NewPopulatedSession(test.Randy, false),
			},
			Ack:        false,
			ConfFCnt:   0,
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
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					NextNFCntDown: 42,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
			},
			Ack:      true,
			ConfFCnt: 24,
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
				dev.Session.NextNFCntDown++
			},
		},
		{
			Name:    "1.1/unconfirmed app downlink/no MAC/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
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
			},
			Ack:      false,
			ConfFCnt: 0,
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
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
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
			},
			Ack:      true,
			ConfFCnt: 24,
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
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
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
			},
			Ack:      false,
			ConfFCnt: 0,
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
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
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
			},
			Ack:      true,
			ConfFCnt: 24,
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
				MACSettings: ttnpb.MACSettings{
					StatusCountPeriodicity: 3,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					NextNFCntDown: 42,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
				NextStatusAfter: 0,
			},
			Ack:      false,
			ConfFCnt: 0,
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
				dev.Session.NextNFCntDown++
				dev.NextStatusAfter = dev.MACSettings.StatusCountPeriodicity
			},
		},
		{
			Name:    "1.1/no app downlink/status(time/zero time)/no ack",
			Context: test.Context(),
			Device: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr: &DevAddr,
				},
				MACSettings: ttnpb.MACSettings{
					StatusTimePeriodicity: time.Nanosecond,
				},
				MACState: &ttnpb.MACState{
					LoRaWANVersion: ttnpb.MAC_V1_1,
				},
				Session: &ttnpb.Session{
					NextNFCntDown: 42,
					SessionKeys: ttnpb.SessionKeys{
						NwkSEncKey: &ttnpb.KeyEnvelope{
							Key: &NwkSEncKey,
						},
						SNwkSIntKey: &ttnpb.KeyEnvelope{
							Key: &SNwkSIntKey,
						},
					},
				},
			},
			Ack:      false,
			ConfFCnt: 0,
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
				dev.Session.NextNFCntDown++
				dev.NextStatusAt = time.Now().Add(dev.MACSettings.StatusTimePeriodicity)
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			dev := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)

			start := time.Now()
			b, err := generateDownlink(tc.Context, dev, tc.Ack, tc.ConfFCnt)
			if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
				tc.Error == nil && !a.So(err, should.BeNil) {
				t.FailNow()
			}

			a.So(b, should.Resemble, tc.Bytes)

			expected := deepcopy.Copy(tc.Device).(*ttnpb.EndDevice)
			if tc.DeviceDiff != nil {
				tc.DeviceDiff(expected)
			}

			if tc.Device.MACSettings.StatusTimePeriodicity > 0 && tc.Device.NextStatusAt.Before(time.Now()) {
				a.So([]time.Time{start, dev.NextStatusAt, expected.NextStatusAt}, should.BeChronological)
				expected.NextStatusAt = dev.NextStatusAt
			}
			a.So(dev, should.Resemble, expected)
		})
	}
}
