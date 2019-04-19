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

package networkserver

import (
	"testing"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNewMACState(t *testing.T) {
	for _, tc := range []struct {
		Name               string
		Device             *ttnpb.EndDevice
		MACState           *ttnpb.MACState
		FrequencyPlanStore *frequencyplans.Store
		ErrorAssertion     func(*testing.T, error) bool
	}{
		{
			Name: "1.1/EU868",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.EUFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.MACSettings_RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
			},
			MACState: func() *ttnpb.MACState {
				fp := test.Must(frequencyplans.NewStore(test.FrequencyPlansFetcher).GetByID(test.EUFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				phy := test.Must(band.GetByID(fp.BandID)).(band.Band)
				return &ttnpb.MACState{
					DeviceClass:         ttnpb.CLASS_A,
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					PingSlotPeriodicity: ttnpb.PING_EVERY_1S,
					CurrentParameters: ttnpb.MACParameters{
						ADRAckDelay:            uint32(phy.ADRAckDelay),
						ADRAckLimit:            uint32(phy.ADRAckLimit),
						ADRDataRateIndex:       0,
						ADRNbTrans:             1,
						ADRTxPowerIndex:        0,
						BeaconFrequency:        0,
						DownlinkDwellTime:      false,
						MaxDutyCycle:           ttnpb.DUTY_CYCLE_1,
						MaxEIRP:                phy.DefaultMaxEIRP,
						PingSlotDataRateIndex:  ttnpb.DATA_RATE_0,
						PingSlotFrequency:      0,
						RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_16,
						RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_0,
						Rx1DataRateOffset:      0,
						Rx1Delay:               ttnpb.RxDelay(phy.ReceiveDelay1.Seconds()),
						Rx2DataRateIndex:       phy.DefaultRx2Parameters.DataRateIndex,
						Rx2Frequency:           phy.DefaultRx2Parameters.Frequency,
						UplinkDwellTime:        false,
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:   868100000,
								DownlinkFrequency: 868100000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   868300000,
								DownlinkFrequency: 868300000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   868500000,
								DownlinkFrequency: 868500000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
						},
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRAckDelay:            uint32(phy.ADRAckDelay),
						ADRAckLimit:            uint32(phy.ADRAckLimit),
						ADRDataRateIndex:       0,
						ADRNbTrans:             1,
						ADRTxPowerIndex:        0,
						BeaconFrequency:        0,
						DownlinkDwellTime:      false,
						MaxDutyCycle:           ttnpb.DUTY_CYCLE_1,
						MaxEIRP:                phy.DefaultMaxEIRP,
						PingSlotDataRateIndex:  ttnpb.DATA_RATE_0,
						PingSlotFrequency:      0,
						RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_16,
						RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_0,
						Rx1DataRateOffset:      0,
						Rx1Delay:               ttnpb.RX_DELAY_13,
						Rx2DataRateIndex:       phy.DefaultRx2Parameters.DataRateIndex,
						Rx2Frequency:           phy.DefaultRx2Parameters.Frequency,
						UplinkDwellTime:        false,
						Channels: []*ttnpb.MACParameters_Channel{
							{
								UplinkFrequency:   868100000,
								DownlinkFrequency: 868100000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   868300000,
								DownlinkFrequency: 868300000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   868500000,
								DownlinkFrequency: 868500000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   867100000,
								DownlinkFrequency: 867100000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   867300000,
								DownlinkFrequency: 867300000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   867500000,
								DownlinkFrequency: 867500000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   867700000,
								DownlinkFrequency: 867700000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
							{
								UplinkFrequency:   867900000,
								DownlinkFrequency: 867900000,
								MinDataRateIndex:  ttnpb.DATA_RATE_0,
								MaxDataRateIndex:  ttnpb.DATA_RATE_5,
								EnableUplink:      true,
							},
						},
					},
				}
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
		{
			Name: "1.1/US915",
			Device: &ttnpb.EndDevice{
				FrequencyPlanID:   test.USFrequencyPlanID,
				LoRaWANVersion:    ttnpb.MAC_V1_1,
				LoRaWANPHYVersion: ttnpb.PHY_V1_1_REV_B,
				MACSettings: &ttnpb.MACSettings{
					DesiredRx1Delay: &ttnpb.MACSettings_RxDelayValue{
						Value: ttnpb.RX_DELAY_13,
					},
				},
			},
			MACState: func() *ttnpb.MACState {
				var bandChannels []*ttnpb.MACParameters_Channel
				for i := 0; i < 64; i++ {
					bandChannels = append(bandChannels, &ttnpb.MACParameters_Channel{
						UplinkFrequency:  uint64(902300000 + 200000*i),
						MinDataRateIndex: ttnpb.DATA_RATE_0,
						MaxDataRateIndex: ttnpb.DATA_RATE_3,
						EnableUplink:     true,
					})
				}
				for i := 0; i < 8; i++ {
					bandChannels = append(bandChannels, &ttnpb.MACParameters_Channel{
						UplinkFrequency:  uint64(903000000 + 1600000*i),
						MinDataRateIndex: ttnpb.DATA_RATE_4,
						MaxDataRateIndex: ttnpb.DATA_RATE_4,
						EnableUplink:     true,
					})
				}
				for i := 0; i < 72; i++ {
					bandChannels[i].DownlinkFrequency = uint64(923300000 + 600000*(i%8))
				}

				fp := test.Must(frequencyplans.NewStore(test.FrequencyPlansFetcher).GetByID(test.USFrequencyPlanID)).(*frequencyplans.FrequencyPlan)
				phy := test.Must(band.GetByID(fp.BandID)).(band.Band)
				return &ttnpb.MACState{
					DeviceClass:         ttnpb.CLASS_A,
					LoRaWANVersion:      ttnpb.MAC_V1_1,
					PingSlotPeriodicity: ttnpb.PING_EVERY_1S,
					CurrentParameters: ttnpb.MACParameters{
						ADRAckDelay:            uint32(phy.ADRAckDelay),
						ADRAckLimit:            uint32(phy.ADRAckLimit),
						ADRDataRateIndex:       0,
						ADRNbTrans:             1,
						ADRTxPowerIndex:        0,
						BeaconFrequency:        0,
						DownlinkDwellTime:      false,
						MaxDutyCycle:           ttnpb.DUTY_CYCLE_1,
						MaxEIRP:                phy.DefaultMaxEIRP,
						PingSlotDataRateIndex:  ttnpb.DATA_RATE_0,
						PingSlotFrequency:      0,
						RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_16,
						RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_0,
						Rx1DataRateOffset:      0,
						Rx1Delay:               ttnpb.RxDelay(phy.ReceiveDelay1.Seconds()),
						Rx2DataRateIndex:       phy.DefaultRx2Parameters.DataRateIndex,
						Rx2Frequency:           phy.DefaultRx2Parameters.Frequency,
						UplinkDwellTime:        false,
						Channels:               deepcopy.Copy(bandChannels).([]*ttnpb.MACParameters_Channel),
					},
					DesiredParameters: ttnpb.MACParameters{
						ADRAckDelay:            uint32(phy.ADRAckDelay),
						ADRAckLimit:            uint32(phy.ADRAckLimit),
						ADRDataRateIndex:       0,
						ADRNbTrans:             1,
						ADRTxPowerIndex:        0,
						BeaconFrequency:        0,
						DownlinkDwellTime:      false,
						MaxDutyCycle:           ttnpb.DUTY_CYCLE_1,
						MaxEIRP:                phy.DefaultMaxEIRP,
						PingSlotDataRateIndex:  ttnpb.DATA_RATE_0,
						PingSlotFrequency:      0,
						RejoinCountPeriodicity: ttnpb.REJOIN_COUNT_16,
						RejoinTimePeriodicity:  ttnpb.REJOIN_TIME_0,
						Rx1DataRateOffset:      0,
						Rx1Delay:               ttnpb.RX_DELAY_13,
						Rx2DataRateIndex:       phy.DefaultRx2Parameters.DataRateIndex,
						Rx2Frequency:           phy.DefaultRx2Parameters.Frequency,
						UplinkDwellTime:        false,
						Channels: func() []*ttnpb.MACParameters_Channel {
							ret := deepcopy.Copy(bandChannels).([]*ttnpb.MACParameters_Channel)
							for _, ch := range ret {
								switch ch.UplinkFrequency {
								case 903900000,
									904100000,
									904300000,
									904500000,
									904700000,
									904900000,
									905100000,
									905300000:
									continue
								}
								ch.EnableUplink = false
							}
							return ret
						}(),
					},
				}
			}(),
			FrequencyPlanStore: frequencyplans.NewStore(test.FrequencyPlansFetcher),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			pb := CopyEndDevice(tc.Device)

			macState, err := newMACState(pb, tc.FrequencyPlanStore, ttnpb.MACSettings{})
			if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(t, err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}
			a.So(macState, should.Resemble, tc.MACState)
			a.So(pb, should.Resemble, tc.Device)
		})
	}
}
