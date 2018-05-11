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
	"fmt"
	"time"

	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver/oauth"
)

// DefaultIdentityServerConfig is the default configuration for the IdentityServer.
var DefaultIdentityServerConfig = identityserver.Config{
	DatabaseURI:      "postgres://root@localhost:26257/is_development?sslmode=disable",
	PublicURL:        "http://localhost:1885",
	OrganizationName: "The Things Network",
	OAuth: oauth.Config{
		AuthorizationCodeTTL: time.Minute * 5,
		AccessTokenTTL:       time.Hour,
		PublicURL:            shared.DefaultOAuthPublicURL,
	},
}

// DefaultIdentityServerInitialData is the default initial data when running the `init` command.
var DefaultIdentityServerInitialData = identityserver.InitialData{
	Settings: identityserver.DefaultSettings,
	Admin: identityserver.InitialAdminData{
		UserID:   "admin",
		Email:    "admin@localhost",
		Password: "admin",
	},
	Console: identityserver.InitialConsoleData{
		ClientSecret: "console",
		RedirectURI:  fmt.Sprintf("%s/oauth/callback", shared.DefaultConsolePublicURL),
	},
}
