// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package mac_test

import (
	"context"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type relayKeyServiceFunc func(
	context.Context, *ttnpb.ApplicationIdentifiers, []string, [][]byte,
) ([]*types.DevAddr, []*types.AES128Key, error)

var _ (mac.RelayKeyService) = relayKeyServiceFunc(nil)

// BatchDeriveRootWorSKey implements mac.RelayKeyService.
func (f relayKeyServiceFunc) BatchDeriveRootWorSKey(
	ctx context.Context, appID *ttnpb.ApplicationIdentifiers, deviceIDs []string, sessionKeyIDs [][]byte,
) ([]*types.DevAddr, []*types.AES128Key, error) {
	return f(ctx, appID, deviceIDs, sessionKeyIDs)
}

func TestDeviceNeedsRelayUpdateUplinkListReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Needs       bool
	}{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
		},
		{
			Name: "no relay",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
		},
		{
			Name: "add rule",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "update rule",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 12,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											Limits: &ttnpb.RelayUplinkForwardLimits{
												BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_12,
												ReloadRate: 24,
											},
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "remove rule",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 12,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
										{
											LastWFCnt: 12,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{},
										{
											LastWFCnt: 12,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.InputDevice)
				res := mac.DeviceNeedsRelayUpdateUplinkListReq(dev)
				if tc.Needs {
					a.So(res, should.BeTrue)
				} else {
					a.So(res, should.BeFalse)
				}
				a.So(dev, should.Resemble, tc.InputDevice)
			},
		})
	}
}

func TestEnqueueRelayUpdateUplinkListReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		RelayKeyService             relayKeyServiceFunc
		State                       mac.EnqueueState
	}{
		{
			Name: "add rule",
			InputDevice: &ttnpb.EndDevice{
				Ids: test.MakeEndDeviceIdentifiers(),
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				Ids: test.MakeEndDeviceIdentifiers(),
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayUpdateUplinkListReq{
							RuleIndex:   0,
							DevAddr:     types.DevAddr{0x42, 0x43, 0x44, 0x45}.Bytes(),
							WFCnt:       42,
							RootWorSKey: types.AES128Key{0x01, 0x02, 0x03, 0x04}.Bytes(),

							DeviceId:     "foo",
							SessionKeyId: []byte{0x01, 0x02, 0x03},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			RelayKeyService: func(
				context.Context, *ttnpb.ApplicationIdentifiers, []string, [][]byte,
			) ([]*types.DevAddr, []*types.AES128Key, error) {
				return []*types.DevAddr{{0x42, 0x43, 0x44, 0x45}}, []*types.AES128Key{{0x01, 0x02, 0x03, 0x04}}, nil
			},
			State: mac.EnqueueState{
				MaxDownLen: 23,
				MaxUpLen:   49,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayUpdateUplinkListRequest.With(events.WithData(
						&ttnpb.MACCommand_RelayUpdateUplinkListReq{
							RuleIndex: 0,
							DevAddr:   types.DevAddr{0x42, 0x43, 0x44, 0x45}.Bytes(),
							WFCnt:     42,

							DeviceId:     "foo",
							SessionKeyId: []byte{0x01, 0x02, 0x03},
						},
					)),
				},
				Ok: true,
			},
		},
		{
			Name: "update rule",
			InputDevice: &ttnpb.EndDevice{
				Ids: test.MakeEndDeviceIdentifiers(),
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 24,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 42,

											DeviceId:     "foobar",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				Ids: test.MakeEndDeviceIdentifiers(),
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 24,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 42,

											DeviceId:     "foobar",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayUpdateUplinkListReq{
							RuleIndex:   1,
							DevAddr:     types.DevAddr{0x43, 0x44, 0x45, 0x46}.Bytes(),
							WFCnt:       42,
							RootWorSKey: types.AES128Key{0x02, 0x03, 0x04, 0x05}.Bytes(),

							DeviceId:     "foobar",
							SessionKeyId: []byte{0x03, 0x04, 0x05},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			RelayKeyService: func(
				context.Context, *ttnpb.ApplicationIdentifiers, []string, [][]byte,
			) ([]*types.DevAddr, []*types.AES128Key, error) {
				return []*types.DevAddr{{0x43, 0x44, 0x45, 0x46}}, []*types.AES128Key{{0x02, 0x03, 0x04, 0x05}}, nil
			},
			State: mac.EnqueueState{
				MaxDownLen: 23,
				MaxUpLen:   49,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayUpdateUplinkListRequest.With(events.WithData(
						&ttnpb.MACCommand_RelayUpdateUplinkListReq{
							RuleIndex: 1,
							DevAddr:   types.DevAddr{0x43, 0x44, 0x45, 0x46}.Bytes(),
							WFCnt:     42,

							DeviceId:     "foobar",
							SessionKeyId: []byte{0x03, 0x04, 0x05},
						},
					)),
				},
				Ok: true,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.InputDevice)
				st := mac.EnqueueRelayUpdateUplinkListReq(
					ctx, dev, tc.MaxDownlinkLength, tc.MaxUplinkLength, tc.RelayKeyService,
				)
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
				st.QueuedEvents = tc.State.QueuedEvents
				a.So(st, should.Resemble, tc.State)
			},
		})
	}
}

func TestHandleRelayUpdateUplinkListAns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RelayUpdateUplinkListAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "add rule",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayUpdateUplinkListReq{
							RuleIndex: 0,
							DevAddr:   types.DevAddr{0x42, 0x43, 0x44, 0x45}.Bytes(),
							WFCnt:     42,

							DeviceId:     "foo",
							SessionKeyId: []byte{0x01, 0x02, 0x03},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
									},
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayUpdateUplinkListAns{},
			Events: events.Builders{
				mac.EvtReceiveRelayUpdateUplinkListAnswer.With(events.WithData(
					&ttnpb.MACCommand_RelayUpdateUplinkListAns{},
				)),
			},
		},
		{
			Name: "update rule",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 24,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 42,

											DeviceId:     "foobar",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayUpdateUplinkListReq{
							RuleIndex: 1,
							DevAddr:   types.DevAddr{0x43, 0x44, 0x45, 0x46}.Bytes(),
							WFCnt:     42,

							DeviceId:     "foobar",
							SessionKeyId: []byte{0x03, 0x04, 0x05},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 42,

											DeviceId:     "foobar",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
									},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									UplinkForwardingRules: []*ttnpb.RelayUplinkForwardingRule{
										{
											LastWFCnt: 42,

											DeviceId:     "foo",
											SessionKeyId: []byte{0x01, 0x02, 0x03},
										},
										{
											LastWFCnt: 42,

											DeviceId:     "foobar",
											SessionKeyId: []byte{0x03, 0x04, 0x05},
										},
									},
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayUpdateUplinkListAns{},
			Events: events.Builders{
				mac.EvtReceiveRelayUpdateUplinkListAnswer.With(events.WithData(
					&ttnpb.MACCommand_RelayUpdateUplinkListAns{},
				)),
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.Device)
				evs, err := mac.HandleRelayUpdateUplinkListAns(ctx, dev, tc.Payload)
				if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
					tc.Error == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(dev, should.Resemble, tc.Expected)
				a.So(evs, should.ResembleEventBuilders, tc.Events)
			},
		})
	}
}
