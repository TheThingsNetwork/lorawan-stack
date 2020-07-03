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

package identityserver

import (
	"path"
	"time"

	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/v3/pkg/accountapp"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// DefaultIdentityServerConfig is the default configuration for the Identity Server.
var DefaultIdentityServerConfig = identityserver.Config{
	DatabaseURI: "postgresql://root@localhost:26257/ttn_lorawan_dev?sslmode=disable",
	AccountApp: accountapp.Config{
		UI: accountapp.UIConfig{
			TemplateData: webui.TemplateData{
				SiteName:      "The Things Stack for LoRaWAN",
				Title:         "Account",
				Language:      "en",
				CanonicalURL:  shared.DefaultAccountAppPublicURL,
				AssetsBaseURL: shared.DefaultAssetsBaseURL,
				IconPrefix:    "account-",
				CSSFiles:      []string{"account.css"},
				JSFiles:       []string{"account.js"},
			},
			FrontendConfig: accountapp.FrontendConfig{
				StackConfig: accountapp.StackConfig{
					IS: webui.APIConfig{Enabled: true, BaseURL: shared.DefaultPublicURL + "/api/v3"},
				},
			},
		},
	},
}

func init() {
	DefaultIdentityServerConfig.AuthCache.MembershipTTL = 10 * time.Minute
	DefaultIdentityServerConfig.UserRegistration.Invitation.TokenTTL = 7 * 24 * time.Hour
	DefaultIdentityServerConfig.UserRegistration.PasswordRequirements.MinLength = 8
	DefaultIdentityServerConfig.UserRegistration.PasswordRequirements.MaxLength = 1000
	DefaultIdentityServerConfig.UserRegistration.PasswordRequirements.MinUppercase = 1
	DefaultIdentityServerConfig.UserRegistration.PasswordRequirements.MinDigits = 1
	DefaultIdentityServerConfig.Email.Network.Name = DefaultIdentityServerConfig.AccountApp.UI.SiteName
	DefaultIdentityServerConfig.Email.Network.IdentityServerURL = shared.DefaultAccountAppPublicURL
	DefaultIdentityServerConfig.Email.Network.ConsoleURL = shared.DefaultConsolePublicURL
	DefaultIdentityServerConfig.ProfilePicture.Bucket = "profile_pictures"
	DefaultIdentityServerConfig.ProfilePicture.BucketURL = path.Join(shared.DefaultAssetsBaseURL, "blob", "profile_pictures")
	DefaultIdentityServerConfig.ProfilePicture.UseGravatar = true
	DefaultIdentityServerConfig.EndDevicePicture.Bucket = "end_device_pictures"
	DefaultIdentityServerConfig.EndDevicePicture.BucketURL = path.Join(shared.DefaultAssetsBaseURL, "blob", "end_device_pictures")
}
