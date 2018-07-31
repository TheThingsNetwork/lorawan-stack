// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

var evtMACRekey = events.Define("ns.mac.rekey_ind", "handled device rekey indication")

func handleRekeyInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_RekeyInd) error {
	if !dev.SupportsJoin {
		return nil
	}

	if pld == nil {
		return errMissingPayload
	}

	conf := &ttnpb.MACCommand_RekeyConf{
		MinorVersion: pld.MinorVersion,
	}

	dev.SessionFallback = nil
	dev.MACState.QueuedResponses = append(
		dev.MACState.QueuedResponses,
		conf.MACCommand(),
	)

	events.Publish(evtMACRekey(ctx, dev.EndDeviceIdentifiers, conf))
	return nil
}
