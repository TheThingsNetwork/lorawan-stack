// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

// Internal package contains various Network Server utilities
package internal

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var LoRaWANVersionPairs = map[ttnpb.MACVersion]map[ttnpb.PHYVersion]struct{}{
	ttnpb.MACVersion_MAC_V1_0: {
		ttnpb.PHYVersion_TS001_V1_0: struct{}{},
	},
	ttnpb.MACVersion_MAC_V1_0_1: {
		ttnpb.PHYVersion_TS001_V1_0_1: struct{}{},
	},
	ttnpb.MACVersion_MAC_V1_0_2: {
		ttnpb.PHYVersion_RP001_V1_0_2:       struct{}{},
		ttnpb.PHYVersion_RP001_V1_0_2_REV_B: struct{}{},
	},
	ttnpb.MACVersion_MAC_V1_0_3: {
		ttnpb.PHYVersion_RP001_V1_0_3_REV_A: struct{}{},
	},
	ttnpb.MACVersion_MAC_V1_1: {
		ttnpb.PHYVersion_RP001_V1_1_REV_A: struct{}{},
		ttnpb.PHYVersion_RP001_V1_1_REV_B: struct{}{},
	},
}

var LoRaWANBands = func() map[string]map[ttnpb.PHYVersion]*band.Band {
	bands := make(map[string]map[ttnpb.PHYVersion]*band.Band, len(band.All))
	for id, vers := range band.All {
		m := make(map[ttnpb.PHYVersion]*band.Band, len(vers))
		for ver, b := range vers {
			b := b
			m[ver] = &b
		}
		bands[id] = m
	}
	return bands
}()

var errNoBandVersion = errors.DefineInvalidArgument("no_band_version", "specified version `{ver}` of band `{id}` does not exist")

func FrequencyPlanAndBand(frequencyPlanID string, phyVersion ttnpb.PHYVersion, fps *frequencyplans.Store) (*frequencyplans.FrequencyPlan, *band.Band, error) {
	fp, err := fps.GetByID(frequencyPlanID)
	if err != nil {
		return nil, nil, err
	}
	b, ok := LoRaWANBands[fp.BandID][phyVersion]
	if !ok || b == nil {
		return nil, nil, errNoBandVersion.WithAttributes(
			"ver", phyVersion,
			"id", fp.BandID,
		)
	}
	return fp, b, nil
}

func DeviceFrequencyPlanAndBand(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*frequencyplans.FrequencyPlan, *band.Band, error) {
	return FrequencyPlanAndBand(dev.FrequencyPlanId, dev.LorawanPhyVersion, fps)
}

func DeviceBand(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*band.Band, error) {
	_, phy, err := DeviceFrequencyPlanAndBand(dev, fps)
	return phy, err
}

func LastUplink(ups ...*ttnpb.MACState_UplinkMessage) *ttnpb.MACState_UplinkMessage {
	return ups[len(ups)-1]
}

func LastDownlink(downs ...*ttnpb.MACState_DownlinkMessage) *ttnpb.MACState_DownlinkMessage {
	return downs[len(downs)-1]
}

func RXMetadataStats(ctx context.Context, mds []*ttnpb.RxMetadata) (gateways int, maxSNR float32) {
	if len(mds) == 0 {
		return 0, 0
	}
	gtws := make(map[string]struct{}, len(mds))
	maxSNR = mds[0].Snr
	for _, md := range mds {
		switch {
		case md.PacketBroker != nil:
			gtws[fmt.Sprintf(
				"%s@%s/%s",
				md.PacketBroker.ForwarderClusterId,
				md.PacketBroker.ForwarderNetId,
				md.PacketBroker.ForwarderTenantId,
			)] = struct{}{}
		case md.Relay != nil:
			gtws[fmt.Sprintf("relay:%s", md.Relay.DeviceId)] = struct{}{}
		case md.GatewayIds != nil:
			gtws[unique.ID(ctx, md.GatewayIds)] = struct{}{}
		default:
			continue // Metadata without PB, Relay or Gateway IDs should be invalid so skipping.
		}
		if md.Snr > maxSNR {
			maxSNR = md.Snr
		}
	}
	return len(gtws), maxSNR
}

func TimePtr(v time.Time) *time.Time {
	return &v
}

// FullFCnt returns full FCnt given fCnt, lastFCnt and whether or not 32-bit FCnts are supported.
func FullFCnt(fCnt uint16, lastFCnt uint32, supports32BitFCnt bool) uint32 {
	switch {
	case fCnt == 0 && lastFCnt == 0:
		return 0
	case !supports32BitFCnt:
		return uint32(fCnt)
	case uint32(fCnt) >= lastFCnt&0xffff:
		return lastFCnt&^0xffff | uint32(fCnt)
	case lastFCnt < 0xffff0000:
		return (lastFCnt+0x10000)&^0xffff | uint32(fCnt)
	default:
		return uint32(fCnt)
	}
}
