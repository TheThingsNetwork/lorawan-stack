// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// defaultSettings are the default settings loaded in the Identity Server
// when the database is created for the first time.
var defaultSettings = &ttnpb.IdentityServerSettings{
	BlacklistedIDs: []string{
		"me",
		"self",
		"this",
		"myself",
		"admin",
		"administrator",
		"root",
		"staff",
		"ttn",
		"tti",
		"support",
		"sysadmin",
		"console",
		"webui",
		"dashboard",
		"handler",
		"broker",
		"router",
		"ns",
		"is",
		"as",
		"js",
		"gs",
		"ga",
		"networkserver",
		"identityserver",
		"applicationserver",
		"joinserver",
		"gatewayserver",
		"gatewayagent",
	},
	IdentityServerSettings_UserRegistrationFlow: ttnpb.IdentityServerSettings_UserRegistrationFlow{
		SelfRegistration: true,
		SkipValidation:   false,
		AdminApproval:    false,
	},
	AllowedEmails:      []string{"*"},
	ValidationTokenTTL: time.Hour,
}
