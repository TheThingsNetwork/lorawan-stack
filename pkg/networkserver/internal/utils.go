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
	"time"

	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var LoRaWANVersionPairs = map[ttnpb.MACVersion]map[ttnpb.PHYVersion]struct{}{
	ttnpb.MAC_V1_0: {
		ttnpb.PHY_V1_0: struct{}{},
	},
	ttnpb.MAC_V1_0_1: {
		ttnpb.PHY_V1_0_1: struct{}{},
	},
	ttnpb.MAC_V1_0_2: {
		ttnpb.PHY_V1_0_2_REV_A: struct{}{},
		ttnpb.PHY_V1_0_2_REV_B: struct{}{},
	},
	ttnpb.MAC_V1_0_3: {
		ttnpb.PHY_V1_0_3_REV_A: struct{}{},
	},
	ttnpb.MAC_V1_1: {
		ttnpb.PHY_V1_1_REV_A: struct{}{},
		ttnpb.PHY_V1_1_REV_B: struct{}{},
	},
}

var LoRaWANBands = func() map[string]map[ttnpb.PHYVersion]*band.Band {
	bands := make(map[string]map[ttnpb.PHYVersion]*band.Band, len(band.All))
	for _, b := range band.All {
		vers := b.Versions()
		m := make(map[ttnpb.PHYVersion]*band.Band, len(vers))
		for _, ver := range vers {
			b, err := b.Version(ver)
			if err != nil {
				panic(fmt.Errorf("failed to obtain %s band of version %s", b.ID, ver))
			}
			m[ver] = &b
		}
		bands[b.ID] = m
	}
	return bands
}()

var errNoBandVersion = errors.DefineInvalidArgument("no_band_version", "specified version `{ver}` of band `{id}` does not exist")

func DeviceFrequencyPlanAndBand(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*frequencyplans.FrequencyPlan, *band.Band, error) {
	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return nil, nil, err
	}
	b, ok := LoRaWANBands[fp.BandID][dev.LoRaWANPHYVersion]
	if !ok || b == nil {
		return nil, nil, errNoBandVersion.WithAttributes(
			"ver", dev.LoRaWANPHYVersion,
			"id", fp.BandID,
		)
	}
	return fp, b, nil
}

func DeviceBand(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*band.Band, error) {
	_, phy, err := DeviceFrequencyPlanAndBand(dev, fps)
	return phy, err
}

func LastUplink(ups ...*ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return ups[len(ups)-1]
}

func LastDownlink(downs ...*ttnpb.DownlinkMessage) *ttnpb.DownlinkMessage {
	return downs[len(downs)-1]
}

func RXMetadataStats(ctx context.Context, mds []*ttnpb.RxMetadata) (gateways int, maxSNR float32) {
	if len(mds) == 0 {
		return 0, 0
	}
	gtws := make(map[string]struct{}, len(mds))
	maxSNR = mds[0].SNR
	for _, md := range mds {
		if md.PacketBroker != nil {
			gtws[fmt.Sprintf("%s@%s/%s", md.PacketBroker.ForwarderID, md.PacketBroker.ForwarderNetID, md.PacketBroker.ForwarderTenantID)] = struct{}{}
		} else {
			gtws[unique.ID(ctx, md.GatewayIdentifiers)] = struct{}{}
		}
		if md.SNR > maxSNR {
			maxSNR = md.SNR
		}
	}
	return len(gtws), maxSNR
}

func TimePtr(t time.Time) *time.Time {
	return &t
}

// CopyEndDevice returns a deep copy of ttnpb.EndDevice pb.
func CopyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

// CopyUplinkMessage returns a deep copy of ttnpb.UplinkMessage pb.
func CopyUplinkMessage(pb *ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return deepcopy.Copy(pb).(*ttnpb.UplinkMessage)
}
