// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// defaultSettings are the default settings loaded in the Identity Server
// when the database is created for the first time.
var defaultSettings = &ttnpb.IdentityServerSettings{
	BlacklistedIDs: []string{
		"admin",
		"root",
		"ttn",
		"tti",
		"sysadmin",
		"console",
		"webui",
		"dashboard",
		"handler",
		"ns",
		"broker",
		"router",
		"is",
		"as",
		"js",
	},
	IdentityServerSettings_UserRegistrationFlow: ttnpb.IdentityServerSettings_UserRegistrationFlow{
		SelfRegistration: true,
		SkipValidation:   false,
		AdminApproval:    false,
	},
	AllowedEmails:      []string{"*"},
	ValidationTokenTTL: time.Duration(time.Hour),
}
