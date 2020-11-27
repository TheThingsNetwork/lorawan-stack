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

	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
)

// DefaultWebhookTemplatesConfig is the default configuration for the Webhook templates.
var DefaultWebhookTemplatesConfig = web.TemplatesConfig{
	URL: "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-webhook-templates/master",
}

// DefaultApplicationServerConfig is the default configuration for the Application Server.
var DefaultApplicationServerConfig = applicationserver.Config{
	MQTT: config.MQTT{
		Listen:           ":1883",
		ListenTLS:        ":8883",
		PublicAddress:    fmt.Sprintf("%s:1883", shared.DefaultPublicHost),
		PublicTLSAddress: fmt.Sprintf("%s:8883", shared.DefaultPublicHost),
	},
	Webhooks: applicationserver.WebhooksConfig{
		Templates: DefaultWebhookTemplatesConfig,
		Target:    "direct",
		Timeout:   5 * time.Second,
		QueueSize: 16,
		Workers:   16,
		Downlinks: web.DownlinksConfig{PublicAddress: shared.DefaultPublicURL + "/api/v3"},
	},
	EndDeviceFetcher: applicationserver.EndDeviceFetcherConfig{
		Cache: applicationserver.EndDeviceFetcherCacheConfig{
			Enable: true,
			TTL:    5 * time.Minute,
		},
	},
	Distribution: applicationserver.DistributionConfig{
		Timeout: time.Minute,
	},
}
