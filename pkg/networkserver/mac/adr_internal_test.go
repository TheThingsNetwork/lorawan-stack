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
	"math"
	"math/rand"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	ADRLossRate          = adrLossRate
	ADRUplinks           = adrUplinks
	ADRDataRateRange     = adrDataRateRange
	ADRTxPowerRange      = adrTxPowerRange
	ADRMargin            = adrMargin
	ADRAdaptDataRate     = adrAdaptDataRate
	ADRAdaptTxPowerIndex = adrAdaptTxPowerIndex

	ClampDataRateRange = clampDataRateRange
	ClampTxPowerRange  = clampTxPowerRange
	ClampNbTrans       = clampNbTrans

	TxPowerStep = txPowerStep
)

func NewADRUplink(
	fCnt uint32,
	maxSNR float32,
	gtwCount uint,
	confirmed bool,
	tx *ttnpb.MACState_UplinkMessage_TxSettings,
) *ttnpb.MACState_UplinkMessage {
	if gtwCount == 0 {
		gtwCount = 1 + uint(rand.Int()%100)
	}
	mds := make([]*ttnpb.MACState_UplinkMessage_RxMetadata, 0, gtwCount)
	for i := uint(0); i < gtwCount; i++ {
		mds = append(mds, &ttnpb.MACState_UplinkMessage_RxMetadata{
			Snr: float32(-rand.Int31n(math.MaxInt32+int32(maxSNR)-1)) - rand.Float32() + maxSNR,
		})
	}
	mds[rand.Intn(len(mds))].Snr = maxSNR

	mType := ttnpb.MType_UNCONFIRMED_UP
	if confirmed {
		mType = ttnpb.MType_CONFIRMED_UP
	}

	return &ttnpb.MACState_UplinkMessage{
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
		Settings:   tx,
	}
}

type ADRMatrixRow struct {
	FCnt         uint32
	MaxSNR       float32
	GtwDiversity uint
	Confirmed    bool
	TxSettings   *ttnpb.TxSettings
}

func ADRMatrixToUplinks(m []ADRMatrixRow) []*ttnpb.MACState_UplinkMessage {
	ups := make([]*ttnpb.MACState_UplinkMessage, 0, len(m))
	for _, r := range m {
		ups = append(ups,
			NewADRUplink(r.FCnt, r.MaxSNR, r.GtwDiversity, r.Confirmed, &ttnpb.MACState_UplinkMessage_TxSettings{
				DataRate: r.TxSettings.GetDataRate(),
			}),
		)
	}
	return ups
}
