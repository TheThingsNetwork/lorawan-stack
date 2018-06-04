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

package oauth

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/assets"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store/sql"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "is_test_oauth"
	issuer   = "issuer.test.local"
	userID   = "john-doe"
	password = "123456"
)

var (
	client = &ttnpb.Client{
		ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo"},
		RedirectURI:       "http://example.com/oauth/callback",
		Secret:            "secret",
		Grants: []ttnpb.GrantType{
			ttnpb.GRANT_AUTHORIZATION_CODE,
			ttnpb.GRANT_REFRESH_TOKEN,
		},
		State: ttnpb.STATE_APPROVED,
		Rights: []ttnpb.Right{
			ttnpb.RIGHT_USER_INFO,
			ttnpb.RIGHT_USER_APPLICATIONS_LIST,
		},
		CreatorIDs: ttnpb.UserIdentifiers{UserID: userID},
	}
	server *Server
	s      *store.Store
)

func cleanStore(logger log.Interface, database string) *store.Store {
	st, err := sql.Open(fmt.Sprintf(address, database))
	if err != nil {
		logger.WithError(err).Fatal("Failed to establish a connection with the CockroachDB instance")
	}

	err = st.Clean()
	if err != nil {
		logger.WithError(err).Fatalf("Failed to delete database `%s`", database)
	}

	err = st.Init()
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize store")
	}

	return st
}

func testServer(t *testing.T) *Server {
	if server == nil {
		logger := test.GetLogger(t)

		a := assertions.New(t)

		s = cleanStore(logger, database)

		p, err := auth.Hash(password)
		a.So(err, should.BeNil)

		err = s.Users.Create(&ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: userID,
			},
			Password: string(p),
		})
		a.So(err, should.BeNil)

		err = s.Clients.Create(client)
		a.So(err, should.BeNil)

		comp, err := component.New(logger, &component.Config{})
		a.So(err, should.BeNil)

		assets := assets.New(comp, assets.Config{
			Directory: "../../webui",
		})

		config := Config{
			AuthorizationCodeTTL: time.Minute * 5,
			AccessTokenTTL:       time.Hour,
			Store:                s,
			Assets:               assets,
			Specializers: SpecializersConfig{
				User:   func(base ttnpb.User) store.User { return &base },
				Client: func(base ttnpb.Client) store.Client { return &base },
			},
		}
		server, err = New(comp, config)
		a.So(err, should.BeNil)

		go server.Component.Start()

		// Wait component to be started.
		time.Sleep(time.Duration(5) * time.Second)
	}

	return server
}
