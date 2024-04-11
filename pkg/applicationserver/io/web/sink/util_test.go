// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package sink_test

import (
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

var (
	registeredApplicationID = &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}
	registeredDeviceID      = &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: registeredApplicationID,
		DeviceId:       "foo-device",
	}
	registeredWebhookID  = "foo-hook"
	registeredWebhookIDs = &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIds: registeredApplicationID,
		WebhookId:      registeredWebhookID,
	}

	timeout = (1 << 5) * test.Delay
)
