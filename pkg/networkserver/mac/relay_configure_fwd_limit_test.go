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

func TestDeviceNeedsRelayConfigureFwdLimitReq(t *testing.T) {
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
			Name: "no limits",
			InputDevice: &ttnpb.EndDevice{
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
									Limits: &ttnpb.ServingRelayForwardingLimits{},
								},
							},
						},
					},
				},
			},
			Needs: true,
		},
		{
			Name: "some limits",
			InputDevice: &ttnpb.EndDevice{
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
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
			Name: "no change",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
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
				res := mac.DeviceNeedsRelayConfigureFwdLimitReq(dev)
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

func TestEnqueueRelayConfigureFwdLimitReq(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       mac.EnqueueState
	}{
		{
			Name: "no limits",
			InputDevice: &ttnpb.EndDevice{
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
									Limits: &ttnpb.ServingRelayForwardingLimits{},
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
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
									Limits: &ttnpb.ServingRelayForwardingLimits{},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfigureFwdLimitReq{}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 44,
				MaxUpLen:   49,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayConfigureFwdLimitRequest.With(events.WithData(
						&ttnpb.MACCommand_RelayConfigureFwdLimitReq{},
					)),
				},
				Ok: true,
			},
		},
		{
			Name: "some limits",
			InputDevice: &ttnpb.EndDevice{
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfigureFwdLimitReq{
							JoinRequestLimits: &ttnpb.RelayForwardLimits{
								BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
								ReloadRate: 12,
							},
							OverallLimits: &ttnpb.RelayForwardLimits{
								BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
								ReloadRate: 23,
							},
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 50,
			MaxUplinkLength:   50,
			State: mac.EnqueueState{
				MaxDownLen: 44,
				MaxUpLen:   49,
				QueuedEvents: events.Builders{
					mac.EvtEnqueueRelayConfigureFwdLimitRequest.With(events.WithData(
						&ttnpb.MACCommand_RelayConfigureFwdLimitReq{
							JoinRequestLimits: &ttnpb.RelayForwardLimits{
								BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
								ReloadRate: 12,
							},
							OverallLimits: &ttnpb.RelayForwardLimits{
								BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
								ReloadRate: 23,
							},
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
				st := mac.EnqueueRelayConfigureFwdLimitReq(ctx, dev, tc.MaxDownlinkLength, tc.MaxUplinkLength)
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
				st.QueuedEvents = tc.State.QueuedEvents
				a.So(st, should.Resemble, tc.State)
			},
		})
	}
}

func TestHandleRelayConfigureFwdLimitAns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		Name             string
		Device, Expected *ttnpb.EndDevice
		Payload          *ttnpb.MACCommand_RelayConfigureFwdLimitAns
		Events           events.Builders
		Error            error
	}{
		{
			Name: "no limits",
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
									Limits: &ttnpb.ServingRelayForwardingLimits{},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfigureFwdLimitReq{}).MACCommand(),
					},
				},
			},
			Expected: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									Limits: &ttnpb.ServingRelayForwardingLimits{},
								},
							},
						},
					},
					DesiredParameters: &ttnpb.MACParameters{
						Relay: &ttnpb.RelayParameters{
							Mode: &ttnpb.RelayParameters_Serving{
								Serving: &ttnpb.ServingRelayParameters{
									Limits: &ttnpb.ServingRelayForwardingLimits{},
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfigureFwdLimitAns{},
			Events: events.Builders{
				mac.EvtReceiveRelayConfigureFwdLimitAnswer.With(events.WithData(
					&ttnpb.MACCommand_RelayConfigureFwdLimitAns{},
				)),
			},
		},
		{
			Name: "some limits",
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
										},
									},
								},
							},
						},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_RelayConfigureFwdLimitReq{
							JoinRequestLimits: &ttnpb.RelayForwardLimits{
								BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
								ReloadRate: 12,
							},
							OverallLimits: &ttnpb.RelayForwardLimits{
								BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
								ReloadRate: 23,
							},
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
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
									Limits: &ttnpb.ServingRelayForwardingLimits{
										JoinRequests: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_2,
											ReloadRate: 12,
										},
										Overall: &ttnpb.RelayForwardLimits{
											BucketSize: ttnpb.RelayLimitBucketSize_RELAY_LIMIT_BUCKET_SIZE_4,
											ReloadRate: 23,
										},
									},
								},
							},
						},
					},
				},
			},
			Payload: &ttnpb.MACCommand_RelayConfigureFwdLimitAns{},
			Events: events.Builders{
				mac.EvtReceiveRelayConfigureFwdLimitAnswer.With(events.WithData(
					&ttnpb.MACCommand_RelayConfigureFwdLimitAns{},
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
				evs, err := mac.HandleRelayConfigureFwdLimitAns(ctx, dev, tc.Payload)
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
