// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
		DefaultSettings:  testSettings(),
		Specializers:     DefaultSpecializers,
	}
	testIS *IdentityServer
	rights []ttnpb.Right
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
		err = is.store.DropDatabase()
		if err != nil {
			logger.WithError(err).Fatal("Failed to drop database")
		}

		err = is.Init()
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

func init() {
	rights = append(rights, ttnpb.AllUserRights()...)
	rights = append(rights, ttnpb.AllApplicationRights()...)
	rights = append(rights, ttnpb.AllGatewayRights()...)
	rights = append(rights, ttnpb.AllOrganizationRights()...)
}

func allRights() []ttnpb.Right { return rights }

func testCtx(userID string) context.Context {
	return newContextWithClaims(context.Background(), &claims{
		EntityIdentifiers: ttnpb.UserIdentifiers{UserID: userID},
		Source:            auth.Token,
		Rights:            rights,
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
		Creator:           testUsers()["alice"].UserIdentifiers,
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
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: "alice"},
			Password:        "123456",
			Admin:           true,
			Email:           "alice@alice.com",
			State:           ttnpb.STATE_APPROVED,
		},
		"bob": {
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: "bob"},
			Password:        "1234567",
			Email:           "bob@bob.com",
			Admin:           true,
		},
		"john-doe": {
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: "john-doe"},
			Password:        "123456",
			Email:           "john@doe.com",
		},
	}
}
