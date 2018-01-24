// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// defaultSettings are the default settings loaded in the Identity Server
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
		SelfRegistration: true,
		SkipValidation:   false,
		AdminApproval:    false,
	},
	AllowedEmails:      []string{"*"},
	ValidationTokenTTL: time.Hour,
	InvitationTokenTTL: time.Hour * time.Duration(24*7),
}

var DefaultFactories = Factories{
	User:        func() store.User { return &ttnpb.User{} },
	Application: func() store.Application { return &ttnpb.Application{} },
	Gateway:     func() store.Gateway { return &ttnpb.Gateway{} },
	Client:      func() store.Client { return &ttnpb.Client{} },
}
