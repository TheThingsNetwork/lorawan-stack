// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// DefaultSettings are the default settings loaded in the Identity Server
// when the database is created for the first time.
var DefaultSettings = &ttnpb.IdentityServerSettings{
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

// DefaultSpecializers contains the specializers used in the store by default.
var DefaultSpecializers = Specializers{
	User:         func(base ttnpb.User) store.User { return &base },
	Application:  func(base ttnpb.Application) store.Application { return &base },
	Gateway:      func(base ttnpb.Gateway) store.Gateway { return &base },
	Client:       func(base ttnpb.Client) store.Client { return &base },
	Organization: func(base ttnpb.Organization) store.Organization { return &base },
}
