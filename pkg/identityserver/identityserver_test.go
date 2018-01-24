// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/auth/oauth"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"google.golang.org/grpc/metadata"
)

var (
	testConfig = &Config{
		DatabaseURI:      "postgres://root@localhost:26257/is_development_tests?sslmode=disable",
		Hostname:         "localhost",
		OrganizationName: "The Things Network",
		PublicURL:        "https://www.thethingsnetwork.org",
	}
	testIS      *IdentityServer
	accessToken string
)

func init() {
	token, err := auth.GenerateAccessToken("")
	if err != nil {
		panic(err)
	}
	accessToken = token
}

func getIS(t testing.TB) *IdentityServer {
	if testIS == nil {
		logger := test.GetLogger(t)
		comp := component.New(logger, &component.Config{})

		is, err := New(comp, testConfig, WithDefaultSettings(testSettings()))
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
			err := is.store.Users.Create(user)
			if err != nil {
				logger.WithError(err).Fatal("Failed to feed the database with test users")
			}
		}

		err = is.store.Clients.Create(testClient())
		if err != nil {
			logger.WithError(err).Fatal("Failed to create test client")
		}

		err = is.store.OAuth.SaveAccessToken(testAccessData())
		if err != nil {
			logger.WithError(err).Fatal("Failed to save test access data")
		}

		testIS = is
	}

	return testIS
}

func testCtx() context.Context {
	return metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", testAccessData().AccessToken)),
	)
}

func testSettings() *ttnpb.IdentityServerSettings {
	return &ttnpb.IdentityServerSettings{
		BlacklistedIDs:     []string{"blacklisted-id", "admin"},
		AllowedEmails:      []string{"*@bar.com"},
		ValidationTokenTTL: time.Duration(time.Hour),
	}
}

func testAccessData() *store.AccessData {
	cli := testClient()

	return &store.AccessData{
		AccessToken: accessToken,
		RedirectURI: cli.RedirectURI,
		Scope:       oauth.Scope(cli.Rights),
		ExpiresIn:   time.Duration(time.Hour),
		CreatedAt:   time.Now(),
		ClientID:    cli.ClientID,
		UserID:      testUsers()["bob"].UserID,
	}
}

func testClient() *ttnpb.Client {
	cli := &ttnpb.Client{
		ClientIdentifier: ttnpb.ClientIdentifier{"test-client"},
		Description:      "foo description",
		Creator:          testUsers()["alice"].UserIdentifier,
		Secret:           "secret",
		RedirectURI:      "localhost",
		Rights:           make([]ttnpb.Right, 0, 50),
	}

	cli.Rights = append(cli.Rights, ttnpb.AllUserRights...)
	cli.Rights = append(cli.Rights, ttnpb.AllApplicationRights...)
	cli.Rights = append(cli.Rights, ttnpb.AllGatewayRights...)

	return cli
}

func testUsers() map[string]*ttnpb.User {
	return map[string]*ttnpb.User{
		"alice": {
			UserIdentifier: ttnpb.UserIdentifier{"alice"},
			Password:       "123456",
			Admin:          true,
			Email:          "alice@alice.com",
		},
		"bob": {
			UserIdentifier: ttnpb.UserIdentifier{"bob"},
			Password:       "1234567",
			Email:          "bob@bob.com",
			Admin:          true,
		},
		"john-doe": {
			UserIdentifier: ttnpb.UserIdentifier{"john-doe"},
			Password:       "123456",
			Email:          "john@doe.com",
		},
	}
}
