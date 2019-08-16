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

package console

import (
	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/console"
	"go.thethings.network/lorawan-stack/pkg/webui"
)

// DefaultConsoleConfig is the default configuration for the Console.
var DefaultConsoleConfig = console.Config{
	OAuth: console.OAuth{
		AuthorizeURL: shared.DefaultOAuthPublicURL + "/authorize",
		TokenURL:     shared.DefaultOAuthPublicURL + "/token",
		ClientID:     "console",
		ClientSecret: "console",
	},
	UI: console.UIConfig{
		TemplateData: webui.TemplateData{
			SiteName:      "The Things Stack for LoRaWAN",
			Title:         "Console",
			SubTitle:      "Management platform for The Things Stack for LoRaWAN",
			Language:      "en",
			CanonicalURL:  shared.DefaultConsolePublicURL,
			AssetsBaseURL: shared.DefaultAssetsBaseURL,
			IconPrefix:    "console-",
			CSSFiles:      []string{"console.css"},
			JSFiles:       []string{"console.js"},
		},
		FrontendConfig: console.FrontendConfig{
			IS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
			GS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
			NS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
			AS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
			JS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
		},
	},
}
