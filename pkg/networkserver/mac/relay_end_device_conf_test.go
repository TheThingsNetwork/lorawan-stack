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

func TestDeviceNeedsRelayEndDeviceConfReq(t *testing.T) {
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
			Name: "disable served",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Needs: true,
		},
		{
			Name: "enable served",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "always to dynamic mode",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Dynamic{
										Dynamic: &ttnpb.RelayEndDeviceDynamicMode{
											SmartEnableLevel: ttnpb.RelaySmartEnableLevel_RELAY_SMART_ENABLE_LEVEL_32,
										},
									},
									ServingDeviceId: "foo",
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
				res := mac.DeviceNeedsRelayEndDeviceConfReq(dev)
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

func TestEnqueueRelayEndDeviceConfReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       mac.EnqueueState
	}{
		{
			Name: "enable served",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayEndDeviceConfReq{
							Configuration: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
								Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always{
									Always: &ttnpb.RelayEndDeviceAlwaysMode{},
								},
								ServingDeviceId: "foo",
							},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 43,
				MaxUpLen:   48,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayEndDeviceConfRequest.With(events.WithData(&ttnpb.MACCommand_RelayEndDeviceConfReq{
						Configuration: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
							Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always{
								Always: &ttnpb.RelayEndDeviceAlwaysMode{},
							},
							ServingDeviceId: "foo",
						},
					})),
				},
				Ok: true,
			},
		},
		{
			Name: "disable served",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayEndDeviceConfReq{}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 43,
				MaxUpLen:   48,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayEndDeviceConfRequest.With(events.WithData(&ttnpb.MACCommand_RelayEndDeviceConfReq{})),
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
				st := mac.EnqueueRelayEndDeviceConfReq(ctx, dev, tc.MaxDownlinkLength, tc.MaxUplinkLength)
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
				st.QueuedEvents = tc.State.QueuedEvents
				a.So(st, should.Resemble, tc.State)
			},
		})
	}
}

func TestHandleRelayEndDeviceConfAns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RelayEndDeviceConfAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "reject served",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayEndDeviceConfReq{
							Configuration: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
								Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always{
									Always: &ttnpb.RelayEndDeviceAlwaysMode{},
								},
								ServingDeviceId: "foo",
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayEndDeviceConfAns{
				SecondChannelFrequencyAck: true,
				// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
				SecondChannelDataRateIndexAck: false,
				SecondChannelIndexAck:         true,
				BackoffAck:                    false,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayEndDeviceConfReject.With(events.WithData(&ttnpb.MACCommand_RelayEndDeviceConfAns{
					SecondChannelFrequencyAck: true,
					// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
					SecondChannelDataRateIndexAck: false,
					SecondChannelIndexAck:         true,
					BackoffAck:                    false,
				})),
			},
		},
		{
			Name: "enable served",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayEndDeviceConfReq{
							Configuration: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration{
								Mode: &ttnpb.MACCommand_RelayEndDeviceConfReq_Configuration_Always{
									Always: &ttnpb.RelayEndDeviceAlwaysMode{},
								},
								ServingDeviceId: "foo",
							},
						}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayEndDeviceConfAns{
				SecondChannelFrequencyAck: true,
				// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         true,
				BackoffAck:                    true,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayEndDeviceConfAccept.With(events.WithData(&ttnpb.MACCommand_RelayEndDeviceConfAns{
					SecondChannelFrequencyAck: true,
					// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
					SecondChannelDataRateIndexAck: true,
					SecondChannelIndexAck:         true,
					BackoffAck:                    true,
				})),
			},
		},
		{
			Name: "disable served",
			Device: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Served{
								Served: &ttnpb.ServedRelayParameters{
									Mode: &ttnpb.ServedRelayParameters_Always{
										Always: &ttnpb.RelayEndDeviceAlwaysMode{},
									},
									ServingDeviceId: "foo",
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayEndDeviceConfReq{}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Payload: &ttnpb.MACCommand_RelayEndDeviceConfAns{
				SecondChannelFrequencyAck: true,
				// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
				SecondChannelDataRateIndexAck: true,
				SecondChannelIndexAck:         true,
				BackoffAck:                    true,
			},
			Events: events.Builders{
				mac.EvtReceiveRelayEndDeviceConfAccept.With(events.WithData(&ttnpb.MACCommand_RelayEndDeviceConfAns{
					SecondChannelFrequencyAck: true,
					// NOTE: SecondChannelAckOffsetAck is not defined in the specification.
					SecondChannelDataRateIndexAck: true,
					SecondChannelIndexAck:         true,
					BackoffAck:                    true,
				})),
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
				evs, err := mac.HandleRelayEndDeviceConfAns(ctx, dev, tc.Payload)
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
