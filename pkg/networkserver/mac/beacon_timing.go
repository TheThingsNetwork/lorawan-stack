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

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// NOTE: This command is deprecated in LoRaWAN 1.0.3

func DeviceNeedsBeaconTimingReq(dev *ttnpb.EndDevice) bool {
	// TODO: Support BeaconTimingReq. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2431)
	return !dev.GetMulticast() && dev.GetMacState().GetDeviceClass() == ttnpb.Class_CLASS_B && false
}

func HandleBeaconTimingReq(ctx context.Context, dev *ttnpb.EndDevice) (events.Builders, error) {
	_ = DeviceNeedsBeaconTimingReq(dev)
	// TODO: Support BeaconTimingReq. (https://github.com/TheThingsNetwork/lorawan-stack/issues/2431)
	return nil, nil
}
