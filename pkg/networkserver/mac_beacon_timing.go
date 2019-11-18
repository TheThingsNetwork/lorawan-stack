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
	"context"

	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func deviceNeedsBeaconTimingReq(dev *ttnpb.EndDevice) bool {
	return dev.MACState != nil &&
		dev.MACState.DeviceClass == ttnpb.CLASS_B &&
		false // TODO: Support Class B (https://github.com/TheThingsNetwork/lorawan-stack/issues/19)
}

func handleBeaconTimingReq(ctx context.Context, dev *ttnpb.EndDevice) ([]events.DefinitionDataClosure, error) {
	_ = deviceNeedsBeaconTimingReq(dev)
	// TODO: Support Class B (https://github.com/TheThingsNetwork/lorawan-stack/issues/19)
	// NOTE: This command is deprecated in LoRaWAN 1.0.3
	return nil, nil
}
