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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func handleADRParamSetupAns(ctx context.Context, dev *ttnpb.EndDevice) error {
	cmds := dev.GetQueuedMACCommands()
	for i, cmd := range cmds {
		if cmd.CID() != ttnpb.CID_ADR_PARAM_SETUP {
			continue
		}

		req := cmd.GetADRParamSetupReq()
		_ = req
		// TODO: Handle ADR parameters (https://github.com/TheThingsIndustries/ttn/issues/292)

		dev.QueuedMACCommands = append(cmds[:i], cmds[i+1:]...)
		return nil
	}
	return ErrMACRequestNotFound.New(nil)
}
