// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "is_development_tests"
)

var testIS *IdentityServer

func getIS(t testing.TB) *IdentityServer {
	if testIS == nil {
		logger := test.GetLogger(t)
		comp := component.New(logger, &component.Config{ServiceBase: shared.DefaultServiceBase})

		is, err := New(comp, testConfig(), WithDefaultSettings(testSettings()))
		if err != nil {
			logger.WithError(err).Fatal("Failed to create an Identity Server instance")
		}

		err = is.start()
		if err != nil {
			logger.WithError(err).Fatal("Failed to initialize the Identity Server instance")
		}

		for _, user := range testUsers() {
			err := is.store.Users.Create(user)
			if err != nil {
				logger.WithError(err).Fatal("Failed to feed the database with test users")
			}
		}

		testIS = is
	}

	return testIS
}

func testConfig() *Config {
	return &Config{
		Hostname:         "development.identityserver.ttn",
		DSN:              fmt.Sprintf(address, database),
		RecreateDatabase: true,
	}
}

func testSettings() *ttnpb.IdentityServerSettings {
	return &ttnpb.IdentityServerSettings{
		BlacklistedIDs:     []string{"blacklisted-id", "admin"},
		AllowedEmails:      []string{"*@bar.com"},
		ValidationTokenTTL: time.Duration(time.Hour),
	}
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
		},
		"john-doe": {
			UserIdentifier: ttnpb.UserIdentifier{"john-doe"},
			Password:       "123456",
			Email:          "john@doe.com",
		},
	}
}
