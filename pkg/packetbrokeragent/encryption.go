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

package packetbrokeragent

import (
	"context"

	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errNoPHYPayload = errors.DefineFailedPrecondition("no_phy_payload", "no PHYPayload in message")

func (a *Agent) encryptUplink(ctx context.Context, msg *packetbroker.UplinkMessage) error {
	// TODO: Obtain KEK, encrypt PHYPayload and gateway metadata (https://github.com/TheThingsIndustries/lorawan-stack/issues/1919).
	if msg.PhyPayload.GetPlain() == nil {
		return errNoPHYPayload.New()
	}
	return nil
}

func (a *Agent) decryptUplink(ctx context.Context, msg *packetbroker.UplinkMessage) error {
	// TODO: Obtain KEK, decrypt PHYPayload and gateway metadata (https://github.com/TheThingsIndustries/lorawan-stack/issues/1919).
	if msg.PhyPayload.GetPlain() == nil {
		return errNoPHYPayload.New()
	}
	return nil
}
