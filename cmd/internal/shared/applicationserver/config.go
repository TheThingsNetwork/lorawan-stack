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

package shared

import (
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
)

// DefaultApplicationServerConfig is the default configuration for the Application Server.
var DefaultApplicationServerConfig = applicationserver.Config{
	LinkMode: "all",
	MQTT: applicationserver.MQTTConfig{
		Listen:    ":1883",
		ListenTLS: ":8883",
		Public:    fmt.Sprintf("mqtt://%s:1883", shared.DefaultPublicHost),
		PublicTLS: fmt.Sprintf("mqtts://%s:8883", shared.DefaultPublicHost),
	},
	Webhooks: applicationserver.WebhooksConfig{
		Target:    "direct",
		Timeout:   5 * time.Second,
		QueueSize: 16,
		Workers:   16,
	},
}
