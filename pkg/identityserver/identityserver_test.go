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
	"context"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
)

var (
	testConfig = Config{
		DatabaseURI:      "postgres://root@localhost:26257/is_development_tests?sslmode=disable",
		Hostname:         "localhost",
		OrganizationName: "The Things Network",
		PublicURL:        "https://www.thethingsnetwork.org",
	}
	initialData = InitialData{
		Settings: testSettings(),
		Admin: InitialAdminData{
			UserID:   "admin",
			Email:    "admin@localhost",
			Password: "12345678",
		},
		Console: InitialConsoleData{
			ClientSecret: "secret",
			RedirectURI:  "http://localhost/oauth/callback",
		},
	}
	testIS *IdentityServer
)

func getIS(t testing.TB) *IdentityServer {
	if testIS == nil {
		logger := test.GetLogger(t)
		comp := component.MustNew(logger, &component.Config{})

		is, err := New(comp, testConfig)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create an Identity Server instance")
		}

		// drop the database before initializing the IS
		err = is.store.Clean()
		if err != nil {
			logger.WithError(err).Fatal("Failed to drop database")
		}

		err = is.Init(initialData)
		if err != nil {
			logger.WithError(err).Fatal("Failed to initialize the Identity Server instance")
		}

		for _, user := range testUsers() {
			err = is.store.Users.Create(user)
			if err != nil {
				logger.WithError(err).Fatal("Failed to feed the database with test users")
			}
		}

		err = is.store.Clients.Create(testClient())
		if err != nil {
			logger.WithError(err).Fatal("Failed to create test client")
		}

		testIS = is
	}

	return testIS
}

func testCtx(ids ttnpb.UserIdentifiers) context.Context {
	return newContextWithAuthorizationData(context.Background(), &authorizationData{
		EntityIdentifiers: ids,
		Source:            auth.Token,
		Rights:            ttnpb.AllRights(),
	})
}

func testSettings() ttnpb.IdentityServerSettings {
	return ttnpb.IdentityServerSettings{
		BlacklistedIDs:     []string{"blacklisted-id", "admin"},
		AllowedEmails:      []string{"*@bar.com"},
		ValidationTokenTTL: time.Duration(time.Hour),
		InvitationTokenTTL: time.Duration(time.Hour),
	}
}

func testClient() *ttnpb.Client {
	cli := &ttnpb.Client{
		ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-client"},
		Description:       "foo description",
		CreatorIDs:        testUsers()["alice"].UserIdentifiers,
		Secret:            "secret",
		RedirectURI:       "localhost",
		Rights:            make([]ttnpb.Right, 0, 50),
		State:             ttnpb.STATE_APPROVED,
	}

	cli.Rights = append(cli.Rights, ttnpb.AllUserRights()...)
	cli.Rights = append(cli.Rights, ttnpb.AllApplicationRights()...)
	cli.Rights = append(cli.Rights, ttnpb.AllGatewayRights()...)
	cli.Rights = append(cli.Rights, ttnpb.AllOrganizationRights()...)

	return cli
}

func testUsers() map[string]*ttnpb.User {
	return map[string]*ttnpb.User{
		"alice": {
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: "alice",
				Email:  "alice@alice.com",
			},
			Password: "123456",
			Admin:    true,
			State:    ttnpb.STATE_APPROVED,
		},
		"bob": {
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: "bob",
				Email:  "bob@bob.com",
			},
			Password: "1234567",
			Admin:    true,
		},
		"john-doe": {
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: "john-doe",
				Email:  "john@doe.com",
			},
			Password: "123456",
		},
	}
}
