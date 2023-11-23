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
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDeviceNeedsRelayCtrlUplinkListReq(t *testing.T) {
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
			Needs: true,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				t.Helper()
				dev := ttnpb.Clone(tc.InputDevice)
				res := mac.DeviceNeedsRelayCtrlUplinkListReq(dev)
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

func TestEnqueueRelayCtrlUplinkListReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       mac.EnqueueState
	}{
		{
			Name: "remove rule",
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
										{},
										{
											LastWFCnt: 42,

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
										{},
										{
											LastWFCnt: 42,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayCtrlUplinkListReq{
							RuleIndex: 0,
							Action:    ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE,
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 48,
				MaxUpLen:   44,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayCtrlUplinkListRequest.With(events.WithData(
						&ttnpb.MACCommand_RelayCtrlUplinkListReq{
							RuleIndex: 0,
							Action:    ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE,
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
				st := mac.EnqueueRelayCtrlUplinkListReq(ctx, dev, tc.MaxDownlinkLength, tc.MaxUplinkLength)
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
				st.QueuedEvents = tc.State.QueuedEvents
				a.So(st, should.Resemble, tc.State)
			},
		})
	}
}

func TestHandleRelayCtrlUplinkListAns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RelayCtrlUplinkListAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "remove rule",
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
											LastWFCnt: 42,

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
											LastWFCnt: 42,

											DeviceId:     "bar",
											SessionKeyId: []byte{0x02, 0x03, 0x04},
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayCtrlUplinkListReq{
							RuleIndex: 0,
							Action:    ttnpb.RelayCtrlUplinkListAction_RELAY_CTRL_UPLINK_LIST_ACTION_REMOVE_TRUSTED_END_DEVICE,
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
										{},
										{
											LastWFCnt: 42,

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
											LastWFCnt: 42,

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
			Payload: &ttnpb.MACCommand_RelayCtrlUplinkListAns{
				RuleIndexAck: true,
				WFCnt:        0x11223344,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayCtrlUplinkListAccept.With(events.WithData(
					&ttnpb.MACCommand_RelayCtrlUplinkListAns{
						RuleIndexAck: true,
						WFCnt:        0x11223344,
					},
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
				evs, err := mac.HandleRelayCtrlUplinkListAns(ctx, dev, tc.Payload)
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
