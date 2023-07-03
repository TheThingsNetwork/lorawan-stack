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

package mac_test

import (
	"context"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestNeedsTxParamSetupReq(t *testing.T) {
	type TestCase struct {
		Name        string
		InputDevice *ttnpb.EndDevice
		Band        *band.Band
		Needs       bool
	}
	tcs := []TestCase{
		{
			Name:        "no MAC state",
			InputDevice: &ttnpb.EndDevice{},
			Band:        LoRaWANBands[band.EU_863_870][ttnpb.PHYVersion_RP001_V1_1_REV_B],
		},
	}
	for _, conf := range []struct {
		Suffix                               string
		CurrentParameters, DesiredParameters *ttnpb.MACParameters
		RecentMacCommandIdentifiers          []ttnpb.MACCommandIdentifier
		Needs                                bool
	}{
		{
			Suffix: "current(EIRP:26,downlink:nil,uplink:nil),desired(EIRP:26,downlink:nil,uplink:nil)",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp: 26,
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp: 26,
			},
		},
		{
			Suffix: "current(EIRP:26,downlink:true,uplink:false),desired(EIRP:26,downlink:nil,uplink:nil)",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp: 26,
			},
		},
		{
			Suffix: "current(EIRP:26,downlink:nil,uplink:nil),desired(EIRP:26,downlink:true,uplink:true)",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp: 26,
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
			},
			Needs: true,
		},
		{
			Suffix: "current(EIRP:26,downlink:nil,uplink:nil),desired(EIRP:26,downlink:false,uplink:false)",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp: 26,
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: false},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			Needs: true,
		},
		{
			Suffix: "current(EIRP:26,downlink:true,uplink:nil),desired(EIRP:26,downlink:false,uplink:false)",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: false},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			Needs: true,
		},
		{
			Suffix: "current(EIRP:24,downlink:true,uplink:false),desired(EIRP:26,downlink:true,uplink:false)",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp:           24,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			Needs: true,
		},
		{
			Suffix: "current(EIRP:24,downlink:true,uplink:false),desired(EIRP:26,downlink:true,uplink:false),recent",
			CurrentParameters: &ttnpb.MACParameters{
				MaxEirp:           24,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			DesiredParameters: &ttnpb.MACParameters{
				MaxEirp:           26,
				DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
				UplinkDwellTime:   &ttnpb.BoolValue{Value: false},
			},
			RecentMacCommandIdentifiers: []ttnpb.MACCommandIdentifier{
				ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP,
			},
		},
	} {
		ForEachBandMACVersion(t, func(makeName func(parts ...string) string, phy *band.Band, phyVersion ttnpb.PHYVersion, macVersion ttnpb.MACVersion) {
			tcs = append(tcs,
				TestCase{
					Name: makeName(conf.Suffix),
					InputDevice: &ttnpb.EndDevice{
						LorawanVersion:    macVersion,
						LorawanPhyVersion: phyVersion,
						MacState: &ttnpb.MACState{
							LorawanVersion:              macVersion,
							CurrentParameters:           conf.CurrentParameters,
							DesiredParameters:           conf.DesiredParameters,
							RecentMacCommandIdentifiers: conf.RecentMacCommandIdentifiers,
						},
					},
					Band:  phy,
					Needs: phy.TxParamSetupReqSupport && conf.Needs && macspec.UseTxParamSetupReq(macVersion),
				},
			)
		})
	}

	for _, tc := range tcs {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)
				res := DeviceNeedsTxParamSetupReq(dev, tc.Band)
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

func TestEnqueueTxParamSetupReq(t *testing.T) {
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		MaxDownlinkLength           uint16
		MaxUplinkLength             uint16
		State                       EnqueueState
	}{
		{
			Name: "payload fits/EIRP 26/dwell time both",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_TxParamSetupReq{
							MaxEirpIndex:      ttnpb.DeviceEIRP_DEVICE_EIRP_26,
							DownlinkDwellTime: true,
							UplinkDwellTime:   true,
						}).MACCommand(),
					},
				},
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: EnqueueState{
				MaxDownLen: 40,
				MaxUpLen:   23,
				Ok:         true,
				QueuedEvents: events.Builders{
					EvtEnqueueTxParamSetupRequest.With(events.WithData(&ttnpb.MACCommand_TxParamSetupReq{
						MaxEirpIndex:      ttnpb.DeviceEIRP_DEVICE_EIRP_26,
						DownlinkDwellTime: true,
						UplinkDwellTime:   true,
					})),
				},
			},
		},
		{
			Name: "payload fits/EIRP 26/no dwell time limitations",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
				},
			},
			MaxDownlinkLength: 42,
			MaxUplinkLength:   24,
			State: EnqueueState{
				MaxDownLen: 42,
				MaxUpLen:   24,
				Ok:         true,
			},
		},
		{
			Name: "downlink does not fit/EIRP 26/dwell time both",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
				},
			},
			MaxDownlinkLength: 1,
			MaxUplinkLength:   24,
			State: EnqueueState{
				MaxDownLen: 1,
				MaxUpLen:   24,
			},
		},
		{
			Name: "uplink does not fit/EIRP 26/dwell time both",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					LorawanVersion: ttnpb.MACVersion_MAC_V1_1,
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp: 26,
					},
					DesiredParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
				},
			},
			MaxDownlinkLength: 42,
			State: EnqueueState{
				MaxDownLen: 42,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)

				st := EnqueueTxParamSetupReq(ctx, dev, tc.MaxDownlinkLength, tc.MaxUplinkLength, LoRaWANBands[band.AS_923][ttnpb.PHYVersion_RP001_V1_1_REV_B])
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(st.QueuedEvents, should.ResembleEventBuilders, tc.State.QueuedEvents)
				st.QueuedEvents = tc.State.QueuedEvents
				a.So(st, should.Resemble, tc.State)
			},
		})
	}
}

func TestHandleTxParamSetupAns(t *testing.T) {
	for _, tc := range []struct {
		Name                        string
		InputDevice, ExpectedDevice *ttnpb.EndDevice
		Events                      events.Builders
		Error                       error
	}{
		{
			Name: "no request",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			Events: events.Builders{
				EvtReceiveTxParamSetupAnswer,
			},
			Error: ErrRequestNotFound.WithAttributes("cid", ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP),
		},
		{
			Name: "EIRP 26, dwell time both",
			InputDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					PendingRequests: []*ttnpb.MACCommand{
						(&ttnpb.MACCommand_TxParamSetupReq{
							MaxEirpIndex:      ttnpb.DeviceEIRP_DEVICE_EIRP_26,
							DownlinkDwellTime: true,
							UplinkDwellTime:   true,
						}).MACCommand(),
					},
					CurrentParameters: &ttnpb.MACParameters{},
					DesiredParameters: &ttnpb.MACParameters{},
				},
			},
			ExpectedDevice: &ttnpb.EndDevice{
				MacState: &ttnpb.MACState{
					CurrentParameters: &ttnpb.MACParameters{
						MaxEirp:           26,
						DownlinkDwellTime: &ttnpb.BoolValue{Value: true},
						UplinkDwellTime:   &ttnpb.BoolValue{Value: true},
					},
					DesiredParameters: &ttnpb.MACParameters{},
					PendingRequests:   []*ttnpb.MACCommand{},
				},
			},
			Events: events.Builders{
				EvtReceiveTxParamSetupAnswer,
			},
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name:     tc.Name,
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				dev := ttnpb.Clone(tc.InputDevice)

				evs, err := HandleTxParamSetupAns(ctx, dev)
				if tc.Error != nil && !a.So(err, should.EqualErrorOrDefinition, tc.Error) ||
					tc.Error == nil && !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(dev, should.Resemble, tc.ExpectedDevice)
				a.So(evs, should.ResembleEventBuilders, tc.Events)
			},
		})
	}
}
