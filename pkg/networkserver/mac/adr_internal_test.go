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
			SNR: float32(-rand.Int31n(math.MaxInt32+int32(maxSNR)-1)) - rand.Float32() + maxSNR,
		})
	}
	mds[rand.Intn(len(mds))].SNR = maxSNR

	mType := ttnpb.MType_UNCONFIRMED_UP
	if confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}

	return &ttnpb.UplinkMessage{
		Payload: &ttnpb.Message{
			MHDR: ttnpb.MHDR{
				MType: mType,
			},
			Payload: &ttnpb.Message_MACPayload{
				MACPayload: &ttnpb.MACPayload{
					FHDR: ttnpb.FHDR{
						FCtrl: ttnpb.FCtrl{
							ADR: true,
						},
						FCnt: fCnt & 0xffff,
					},
					FullFCnt: fCnt,
				},
			},
		},
		RxMetadata: mds,
		Settings:   tx,
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
	if len(m) > 20 {
		panic("ADR matrix contains more than 20 rows")
	}

	ups := make([]*ttnpb.UplinkMessage, 0, 20)
	for _, r := range m {
		ups = append(ups, NewADRUplink(r.FCnt, r.MaxSNR, r.GtwDiversity, r.Confirmed, r.TxSettings))
	}
	return ups
}

func TestLossRate(t *testing.T) {
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
					ss = append(ss, fmt.Sprintf("%d", up.Payload.GetMACPayload().FHDR.FCnt))
				}
				return ss
			}(), ","),
			Parallel: true,
			Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
				a.So(lossRate(tc.Uplinks...), should.Equal, tc.Rate)
			},
		})
	}
}
