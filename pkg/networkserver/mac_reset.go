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

	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func handleResetInd(ctx context.Context, dev *ttnpb.EndDevice, pld *ttnpb.MACCommand_ResetInd, fps *frequencyplans.Store) error {
	if dev.SupportsJoin {
		return nil
	}

	if pld == nil {
		return errMissingPayload
	}

	if err := resetMACState(fps, dev); err != nil {
		return err
	}

	dev.MACState.QueuedResponses = append(
		dev.MACState.QueuedResponses,
		(&ttnpb.MACCommand_ResetConf{
			MinorVersion: pld.MinorVersion,
		}).MACCommand(),
	)
	return nil
}
