// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package band

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func TestGetPhyVersions(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := context.Background()

	for _, tc := range []struct {
		Name           string
		BandID         string
		Expected       *ttnpb.GetPhyVersionsResponse
		ErrorAssertion func(err error) bool
	}{
		{
			Name:   "Unknown",
			BandID: "AS_925",
			ErrorAssertion: func(err error) bool {
				return errors.IsNotFound(err)
			},
		},
		{
			Name:   "EU868",
			BandID: "EU_863_870",
			Expected: &ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "EU_863_870",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
				},
			},
		},
		{
			Name:   "AU915",
			BandID: "AU_915_928",
			Expected: &ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "AU_915_928",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
						},
					},
				},
			},
		},
		{
			Name:   "AS923",
			BandID: "AS_923",
			Expected: &ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "AS_923",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
						},
					},
				},
			},
		},
		{
			Name: "All",
			Expected: &ttnpb.GetPhyVersionsResponse{
				VersionInfo: []*ttnpb.GetPhyVersionsResponse_VersionInfo{
					{
						BandId: "AS_923",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
						},
					},
					{
						BandId: "AS_923_2",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
						},
					},
					{
						BandId: "AS_923_3",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
						},
					},
					{
						BandId: "AS_923_4",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
						},
					},
					{
						BandId: "AU_915_928",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
						},
					},
					{
						BandId: "CN_470_510",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
						},
					},
					{
						BandId: "CN_470_510_20_A",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_470_510_20_B",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_470_510_26_A",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_470_510_26_B",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
						},
					},
					{
						BandId: "CN_779_787",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "EU_433",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "EU_863_870",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "IN_865_867",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
						},
					},
					{
						BandId: "ISM_2400",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "KR_920_923",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
						},
					},
					{
						BandId: "MA_869_870_DRAFT",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
					{
						BandId: "RU_864_870",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
						},
					},
					{
						BandId: "US_902_928",
						PhyVersions: []ttnpb.PHYVersion{
							ttnpb.PHYVersion_RP002_V1_0_3,
							ttnpb.PHYVersion_RP002_V1_0_2,
							ttnpb.PHYVersion_RP002_V1_0_1,
							ttnpb.PHYVersion_RP002_V1_0_0,
							ttnpb.PHYVersion_RP001_V1_0_3_REV_A,
							ttnpb.PHYVersion_RP001_V1_1_REV_B,
							ttnpb.PHYVersion_RP001_V1_1_REV_A,
							ttnpb.PHYVersion_RP001_V1_0_2_REV_B,
							ttnpb.PHYVersion_RP001_V1_0_2,
							ttnpb.PHYVersion_TS001_V1_0_1,
							ttnpb.PHYVersion_TS001_V1_0,
						},
					},
				},
			},
		},
	} {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			res, err := GetPhyVersions(ctx, &ttnpb.GetPhyVersionsRequest{
				BandId: tc.BandID,
			})

			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(err), should.BeTrue)
			} else {
				if !a.So(res, should.NotBeNil) {
					t.Fatalf("Nil value received. Expected :%v", tc.Expected)
				}
				sort.Slice(res.VersionInfo, func(i, j int) bool { return res.VersionInfo[i].BandId <= res.VersionInfo[j].BandId })
				for _, vi := range res.VersionInfo {
					sort.Slice(vi.PhyVersions, func(i, j int) bool { return vi.PhyVersions[i] >= vi.PhyVersions[j] })
				}
				if !a.So(res, should.Resemble, tc.Expected) {
					t.Fatalf("Unexpected value: %v", res)
				}
			}
		})
	}
}

func TestBand_convertToBandDescription(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	for _, tc := range []struct {
		Name           string
		Band           Band
		Expected       *ttnpb.BandDescription
		ErrorAssertion func(err error) bool
	}{
		{
			Name: "All",
			Band: Band{
				ID: "All",

				Beacon: Beacon{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					CodingRate:    "default",
					Frequencies:   nil,
				},
				PingSlotFrequencies: nil,

				MaxUplinkChannels: 1,
				UplinkChannels: []Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				MaxDownlinkChannels: 1,
				DownlinkChannels: []Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				SubBands: []SubBandParameters{
					{
						MinFrequency: 1,
						MaxFrequency: 2,
						DutyCycle:    3.0,
						MaxEIRP:      4.0,
					},
				},

				DataRates: make(map[ttnpb.DataRateIndex]DataRate),

				FreqMultiplier:   1,
				ImplementsCFList: true,
				CFListType:       ttnpb.CFListType_CHANNEL_MASKS,

				ReceiveDelay1: 1 * time.Second,
				ReceiveDelay2: 2 * time.Second,

				JoinAcceptDelay1: 3 * time.Second,
				JoinAcceptDelay2: 4 * time.Second,
				MaxFCntGap:       5,

				SupportsDynamicADR:   true,
				ADRAckLimit:          ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1,
				MinRetransmitTimeout: 1 * time.Second,
				MaxRetransmitTimeout: 2 * time.Second,

				TxOffset:            []float32{1.0, 2.0},
				MaxADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,

				TxParamSetupReqSupport: true,

				DefaultMaxEIRP: 1.0,

				DefaultRx2Parameters: Rx2Parameters{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
				},

				BootDwellTime: DwellTime{
					Uplinks:   boolPtr(true),
					Downlinks: boolPtr(true),
				},
			},
			Expected: &ttnpb.BandDescription{
				Id: "All",

				Beacon: &ttnpb.BandDescription_Beacon{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					CodingRate:    "default",
				},
				PingSlotFrequency: nil,

				MaxUplinkChannels: 1,
				UplinkChannels: []*ttnpb.BandDescription_Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				MaxDownlinkChannels: 1,
				DownlinkChannels: []*ttnpb.BandDescription_Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				SubBands: []*ttnpb.BandDescription_SubBandParameters{
					{
						MinFrequency: 1,
						MaxFrequency: 2,
						DutyCycle:    3.0,
						MaxEirp:      4.0,
					},
				},

				DataRates: make(map[uint32]*ttnpb.BandDescription_BandDataRate),

				FreqMultiplier:   1,
				ImplementsCfList: true,
				CfListType:       ttnpb.CFListType_CHANNEL_MASKS,

				ReceiveDelay_1: &types.Duration{Seconds: 1},
				ReceiveDelay_2: &types.Duration{Seconds: 2},

				JoinAcceptDelay_1: &types.Duration{Seconds: 3},
				JoinAcceptDelay_2: &types.Duration{Seconds: 4},
				MaxFcntGap:        5,

				SupportsDynamicAdr:   true,
				AdrAckLimit:          ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1,
				MinRetransmitTimeout: &types.Duration{Seconds: 1},
				MaxRetransmitTimeout: &types.Duration{Seconds: 2},

				TxOffset:            []float32{1.0, 2.0},
				MaxAdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,

				TxParamSetupReqSupport: true,

				DefaultMaxEirp: 1.0,

				DefaultRx2Parameters: &ttnpb.BandDescription_Rx2Parameters{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
				},

				BootDwellTime: &ttnpb.BandDescription_DwellTime{
					Uplinks:   &types.BoolValue{Value: true},
					Downlinks: &types.BoolValue{Value: true},
				},
			},
		},
		{
			Name: "Nullable",
			Band: Band{
				ID: "Nullable band",

				Beacon: Beacon{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					CodingRate:    "default",
				},

				MaxUplinkChannels: 1,
				UplinkChannels: []Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				MaxDownlinkChannels: 1,
				DownlinkChannels: []Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				SubBands: []SubBandParameters{
					{
						MinFrequency: 1,
						MaxFrequency: 2,
						DutyCycle:    3.0,
						MaxEIRP:      4.0,
					},
				},

				DataRates: make(map[ttnpb.DataRateIndex]DataRate),

				FreqMultiplier:   1,
				ImplementsCFList: true,
				CFListType:       ttnpb.CFListType_CHANNEL_MASKS,

				ReceiveDelay1: 1 * time.Second,
				ReceiveDelay2: 2 * time.Second,

				JoinAcceptDelay1: 3 * time.Second,
				JoinAcceptDelay2: 4 * time.Second,
				MaxFCntGap:       5,

				SupportsDynamicADR:   true,
				ADRAckLimit:          ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1,
				MinRetransmitTimeout: 1 * time.Second,
				MaxRetransmitTimeout: 2 * time.Second,

				TxOffset:            []float32{1.0, 2.0},
				MaxADRDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,

				TxParamSetupReqSupport: true,

				DefaultMaxEIRP: 1.0,

				DefaultRx2Parameters: Rx2Parameters{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
				},

				BootDwellTime: DwellTime{},
			},
			Expected: &ttnpb.BandDescription{
				Id: "Nullable band",

				Beacon: &ttnpb.BandDescription_Beacon{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_0,
					CodingRate:    "default",
				},
				PingSlotFrequency: nil,

				MaxUplinkChannels: 1,
				UplinkChannels: []*ttnpb.BandDescription_Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				MaxDownlinkChannels: 1,
				DownlinkChannels: []*ttnpb.BandDescription_Channel{
					{
						Frequency:   1,
						MinDataRate: 2,
						MaxDataRate: 3,
					},
				},

				SubBands: []*ttnpb.BandDescription_SubBandParameters{
					{
						MinFrequency: 1,
						MaxFrequency: 2,
						DutyCycle:    3.0,
						MaxEirp:      4.0,
					},
				},

				DataRates: make(map[uint32]*ttnpb.BandDescription_BandDataRate),

				FreqMultiplier:   1,
				ImplementsCfList: true,
				CfListType:       ttnpb.CFListType_CHANNEL_MASKS,

				ReceiveDelay_1: &types.Duration{Seconds: 1},
				ReceiveDelay_2: &types.Duration{Seconds: 2},

				JoinAcceptDelay_1: &types.Duration{Seconds: 3},
				JoinAcceptDelay_2: &types.Duration{Seconds: 4},
				MaxFcntGap:        5,

				SupportsDynamicAdr:   true,
				AdrAckLimit:          ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1,
				MinRetransmitTimeout: &types.Duration{Seconds: 1},
				MaxRetransmitTimeout: &types.Duration{Seconds: 2},

				TxOffset:            []float32{1.0, 2.0},
				MaxAdrDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,

				TxParamSetupReqSupport: true,

				DefaultMaxEirp: 1.0,

				DefaultRx2Parameters: &ttnpb.BandDescription_Rx2Parameters{
					DataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
				},

				BootDwellTime: &ttnpb.BandDescription_DwellTime{},
			},
		},
	} {
		tc := tc

		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			res := tc.Band.BandDescription()

			if !a.So(res, should.NotBeNil) {
				t.Fatalf("Nil value received. Expected :%v", tc.Expected)
			}
			if !a.So(res, should.Resemble, tc.Expected) {
				t.Fatalf("Unexpected value: %v", res)
			}
		})
	}
}

func convertBands(input map[string]map[ttnpb.PHYVersion]Band) map[string]*ttnpb.ListBandsResponse_VersionedBandDescription { //nolint:lll
	output := make(map[string]*ttnpb.ListBandsResponse_VersionedBandDescription)
	for bandID, versions := range input {
		versionedBandDescription := &ttnpb.ListBandsResponse_VersionedBandDescription{
			Band: make(map[string]*ttnpb.BandDescription),
		}

		for PHYVersion, band := range versions {
			versionedBandDescription.Band[PHYVersion.String()] = band.BandDescription()
		}

		output[bandID] = versionedBandDescription
	}

	return output
}

func TestListBands(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := context.Background()

	for _, tc := range []struct {
		Name           string
		BandID         string
		PhyVersion     ttnpb.PHYVersion
		Expected       *ttnpb.ListBandsResponse
		ErrorAssertion func(err error) bool
	}{
		{
			Name:   "Unknown",
			BandID: "AS_925",
			ErrorAssertion: func(err error) bool {
				return errors.IsNotFound(err)
			},
		},
		{
			Name:       "All",
			BandID:     "",
			PhyVersion: ttnpb.PHYVersion_PHY_UNKNOWN,
			Expected: &ttnpb.ListBandsResponse{
				Descriptions: convertBands(All),
			},
		},
		{
			Name:       "Band",
			BandID:     AS_923,
			PhyVersion: ttnpb.PHYVersion_PHY_UNKNOWN,
			Expected: &ttnpb.ListBandsResponse{
				Descriptions: convertBands(map[string]map[ttnpb.PHYVersion]Band{
					AS_923: All[AS_923],
				}),
			},
		},
		{
			Name:       "PhyVersion",
			BandID:     "",
			PhyVersion: ttnpb.PHYVersion_TS001_V1_0_1,
			Expected: &ttnpb.ListBandsResponse{
				Descriptions: convertBands(map[string]map[ttnpb.PHYVersion]Band{
					AU_915_928: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[AU_915_928][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					CN_470_510: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[CN_470_510][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					CN_779_787: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[CN_779_787][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					EU_433: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[EU_433][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					EU_863_870: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[EU_863_870][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					ISM_2400: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[ISM_2400][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					US_902_928: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[US_902_928][ttnpb.PHYVersion_TS001_V1_0_1],
					},
					MA_869_870_DRAFT: {
						ttnpb.PHYVersion_TS001_V1_0_1: All[MA_869_870_DRAFT][ttnpb.PHYVersion_TS001_V1_0_1],
					},
				}),
			},
		},
		{
			Name:       "Band and PhyVersion",
			BandID:     AS_923,
			PhyVersion: ttnpb.PHYVersion_RP001_V1_0_2,
			Expected: &ttnpb.ListBandsResponse{
				Descriptions: convertBands(map[string]map[ttnpb.PHYVersion]Band{
					AS_923: {
						ttnpb.PHYVersion_RP001_V1_0_2: All[AS_923][ttnpb.PHYVersion_RP001_V1_0_2],
					},
				}),
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			res, err := ListBands(ctx, &ttnpb.ListBandsRequest{
				BandId:     tc.BandID,
				PhyVersion: tc.PhyVersion,
			})

			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				a.So(tc.ErrorAssertion(err), should.BeTrue)
			} else {
				if !a.So(res, should.NotBeNil) {
					t.Fatalf("Nil value received. Expected :%v", tc.Expected)
				}
				if !a.So(res, should.Resemble, tc.Expected) {
					t.Fatalf("Unexpected value: %v", res)
				}
			}
		})
	}
}
