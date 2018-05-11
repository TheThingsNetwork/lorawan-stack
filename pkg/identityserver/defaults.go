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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// DefaultSettings are the default settings loaded in the Identity Server
// when the database is created for the first time.
var DefaultSettings = ttnpb.IdentityServerSettings{
	BlacklistedIDs: []string{
		"admin",
		"administrator",
		"applicationserver",
		"as",
		"broker",
		"console",
		"dashboard",
		"ga",
		"gatewayagent",
		"gatewayserver",
		"gs",
		"handler",
		"identityserver",
		"is",
		"joinserver",
		"js",
		"me",
		"myself",
		"networkserver",
		"ns",
		"root",
		"router",
		"self",
		"staff",
		"support",
		"sysadmin",
		"this",
		"tti",
		"ttn",
		"webui",
	},
	IdentityServerSettings_UserRegistrationFlow: ttnpb.IdentityServerSettings_UserRegistrationFlow{
		InvitationOnly: false,
		SkipValidation: false,
		AdminApproval:  false,
	},
	AllowedEmails:      []string{"*"},
	ValidationTokenTTL: time.Hour,
	InvitationTokenTTL: time.Hour * time.Duration(24*7),
}
