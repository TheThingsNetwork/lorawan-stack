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

package identityserver

import (
	"time"

	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/oauth"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// DefaultIdentityServerConfig is the default configuration for the Identity Server.
var DefaultIdentityServerConfig = identityserver.Config{
	DatabaseURI: "postgresql://root@localhost:26257/ttn_lorawan_dev?sslmode=disable",
	OAuth: oauth.Config{
		Mount: "/oauth",
		UI: oauth.UIConfig{
			TemplateData: webui.TemplateData{
				SiteName:      "The Things Network Stack for LoRaWAN",
				Language:      "en",
				CanonicalURL:  shared.DefaultOAuthPublicURL,
				AssetsBaseURL: shared.DefaultAssetsBaseURL,
				IconPrefix:    "oauth-",
				CSSFiles:      []string{"oauth.css"},
				JSFiles:       []string{"oauth.js"},
			},
		},
	},
}

func init() {
	DefaultIdentityServerConfig.AuthCache.MembershipTTL = 10 * time.Minute
}
