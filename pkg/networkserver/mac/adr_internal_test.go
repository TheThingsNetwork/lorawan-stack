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

package mac

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func NewADRUplink(fCnt uint32, maxSNR float32, gtwCount uint, confirmed bool, tx ttnpb.TxSettings) *ttnpb.UplinkMessage {
	if gtwCount == 0 {
		gtwCount = 1 + uint(rand.Int()%100)
	}
	mds := make([]*ttnpb.RxMetadata, 0, gtwCount)
	for i := uint(0); i < gtwCount; i++ {
		mds = append(mds, &ttnpb.RxMetadata{
			Snr: float32(-rand.Int31n(math.MaxInt32+int32(maxSNR)-1)) - rand.Float32() + maxSNR,
		})
	}
	mds[rand.Intn(len(mds))].Snr = maxSNR

	mType := ttnpb.MType_UNCONFIRMED_UP
	if confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}

	return &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHdr: &ttnpb.MHDR{
				MType: mType,
			},
			Payload: &ttnpb.Message_MacPayload{
				MacPayload: &ttnpb.MACPayload{
					FHdr: &ttnpb.FHDR{
						FCtrl: &ttnpb.FCtrl{
							Adr: true,
						},
						FCnt: fCnt & 0xffff,
					},
					FullFCnt: fCnt,
				},
			},
		},
		RxMetadata: mds,
		Settings:   &tx,
	}
}

type ADRMatrixRow struct {
	FCnt         uint32
	MaxSNR       float32
	GtwDiversity uint
	Confirmed    bool
	TxSettings   ttnpb.TxSettings
}

func ADRMatrixToUplinks(m []ADRMatrixRow) []*ttnpb.UplinkMessage {
	ups := make([]*ttnpb.UplinkMessage, 0, len(m))
	for _, r := range m {
		ups = append(ups, NewADRUplink(r.FCnt, r.MaxSNR, r.GtwDiversity, r.Confirmed, r.TxSettings))
	}
	return ups
}

func TestADRLossRate(t *testing.T) {
	for _, tc := range []struct {
		Name    string
		Uplinks []*ttnpb.UplinkMessage
		Rate    float32
	}{
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 13},
			}),
			Rate: 1. / 3.,
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 14},
			}),
			Rate: 2. / 4.,
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 13},
				{FCnt: 15},
			}),
			Rate: 2. / 5.,
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 12},
				{FCnt: 13},
				{FCnt: 13},
				{FCnt: 13},
			}),
		},
		{
			Uplinks: ADRMatrixToUplinks([]ADRMatrixRow{
				{FCnt: 11},
				{FCnt: 12},
				{FCnt: 1},
				{FCnt: 1},
				{FCnt: 3},
				{FCnt: 3},
			}),
			Rate: 1. / 3.,
		},
	} {
		tc := tc
		test.RunSubtest(t, test.SubtestConfig{
			Name: strings.Join(func() (ss []string) {
				for _, up := range tc.Uplinks {
					ss = append(ss, fmt.Sprintf("%d", up.Payload.GetMacPayload().FHdr.FCnt))
				}
				return ss
			}(), ","),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				a.So(adrLossRate(tc.Uplinks...), should.Equal, tc.Rate)
			},
		})
	}
}

func TestClampDataRateRange(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings

		InputMinDataRateIndex ttnpb.DataRateIndex
		InputMaxDataRateIndex ttnpb.DataRateIndex

		ExpectedMinDataRateIndex ttnpb.DataRateIndex
		ExpectedMaxDataRateIndex ttnpb.DataRateIndex
	}{
		{
			Name: "no device",

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_5,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_12,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_12,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "maximum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
		},
		{
			Name: "maximum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_5,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_1,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
		},
		{
			Name: "maximum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_12,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum+maximum;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_2,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_3,
		},
		{
			Name: "minimum+maximum;left-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_3,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_6,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_6,
		},
		{
			Name: "minimum+maximum;inside of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_7,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_9,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_7,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_9,
		},
		{
			Name: "minimum+maximum;right-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_7,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_11,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_7,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
		{
			Name: "minimum+maximum;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_12,
								},
								MaxDataRateIndex: &ttnpb.DataRateIndexValue{
									Value: ttnpb.DataRateIndex_DATA_RATE_15,
								},
							},
						},
					},
				},
			},

			InputMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_5,
			InputMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,

			// min > max implies that no common data rate range has been found.
			ExpectedMinDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_12,
			ExpectedMaxDataRateIndex: ttnpb.DataRateIndex_DATA_RATE_10,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			min, max := clampDataRateRange(tc.Device, tc.Defaults, tc.InputMinDataRateIndex, tc.InputMaxDataRateIndex)
			a.So(min, should.Equal, tc.ExpectedMinDataRateIndex)
			a.So(max, should.Equal, tc.ExpectedMaxDataRateIndex)
		})
	}
}
func TestClampTxPowerRange(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings

		InputMinTxPowerIndex uint8
		InputMaxTxPowerIndex uint8

		ExpectedMinTxPowerIndex uint8
		ExpectedMaxTxPowerIndex uint8
	}{
		{
			Name: "no device",

			InputMinTxPowerIndex: 1,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 1,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 5,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 1,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 12,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 12,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "maximum only;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 3,
		},
		{
			Name: "maximum only;inside provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 5,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 1,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 1,
			ExpectedMaxTxPowerIndex: 5,
		},
		{
			Name: "maximum only;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 12,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum+maximum;left of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 2,
								},
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 3,
		},
		{
			Name: "minimum+maximum;left-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 3,
								},
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 6,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 5,
			ExpectedMaxTxPowerIndex: 6,
		},
		{
			Name: "minimum+maximum;inside of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 7,
								},
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 9,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 7,
			ExpectedMaxTxPowerIndex: 9,
		},
		{
			Name: "minimum+maximum;right-joined of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 7,
								},
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 11,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			ExpectedMinTxPowerIndex: 7,
			ExpectedMaxTxPowerIndex: 10,
		},
		{
			Name: "minimum+maximum;right of provided interval",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinTxPowerIndex: &types.UInt32Value{
									Value: 12,
								},
								MaxTxPowerIndex: &types.UInt32Value{
									Value: 15,
								},
							},
						},
					},
				},
			},

			InputMinTxPowerIndex: 5,
			InputMaxTxPowerIndex: 10,

			// min > max implies that no common TX power range has been found.
			ExpectedMinTxPowerIndex: 12,
			ExpectedMaxTxPowerIndex: 10,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			min, max := clampTxPowerRange(tc.Device, tc.Defaults, tc.InputMinTxPowerIndex, tc.InputMaxTxPowerIndex)
			a.So(min, should.Equal, tc.ExpectedMinTxPowerIndex)
			a.So(max, should.Equal, tc.ExpectedMaxTxPowerIndex)
		})
	}
}

func TestClampNbTrans(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		Device   *ttnpb.EndDevice
		Defaults *ttnpb.MACSettings

		InputNbTrans uint32

		ExpectedNbTrans uint32
	}{
		{
			Name: "no device",

			InputNbTrans: 1,

			ExpectedNbTrans: 1,
		},
		{
			Name: "minimum only;left of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &types.UInt32Value{
									Value: 1,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 2,

			ExpectedNbTrans: 2,
		},
		{
			Name: "minimum only;right of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &types.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 2,

			ExpectedNbTrans: 3,
		},
		{
			Name: "maximum only;left of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxNbTrans: &types.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 5,

			ExpectedNbTrans: 3,
		},
		{
			Name: "maximum only;right of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MaxNbTrans: &types.UInt32Value{
									Value: 7,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 5,

			ExpectedNbTrans: 5,
		},
		{
			Name: "minimum+maximum;left of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &types.UInt32Value{
									Value: 2,
								},
								MaxNbTrans: &types.UInt32Value{
									Value: 3,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 5,

			ExpectedNbTrans: 3,
		},
		{
			Name: "minimum+maximum;inside of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &types.UInt32Value{
									Value: 7,
								},
								MaxNbTrans: &types.UInt32Value{
									Value: 9,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 8,

			ExpectedNbTrans: 8,
		},
		{
			Name: "minimum+maximum;right of provided value",

			Device: &ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Adr: &ttnpb.ADRSettings{
						Mode: &ttnpb.ADRSettings_Dynamic{
							Dynamic: &ttnpb.ADRSettings_DynamicMode{
								MinNbTrans: &types.UInt32Value{
									Value: 12,
								},
								MaxNbTrans: &types.UInt32Value{
									Value: 15,
								},
							},
						},
					},
				},
			},

			InputNbTrans: 8,

			ExpectedNbTrans: 12,
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			value := clampNbTrans(tc.Device, tc.Defaults, tc.InputNbTrans)
			a.So(value, should.Equal, tc.ExpectedNbTrans)
		})
	}
}
