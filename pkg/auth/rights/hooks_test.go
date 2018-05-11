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

package rights_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-identity-server/commands"
	"go.thethings.network/lorawan-stack/pkg/auth"
	. "go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store/sql"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const databaseURI = "postgres://root@localhost:26257/is_rightshook_test?sslmode=disable"

var testHandler = func(t *testing.T, expected []ttnpb.Right) grpc.UnaryHandler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		assertions.New(t).So(FromContext(ctx), should.Resemble, expected)
		return nil, nil
	}
}

// isProvider implements rights.IdentityServerConnector.
type isProvider struct {
	conn *grpc.ClientConn
}

func (p isProvider) Get(ctx context.Context) *grpc.ClientConn {
	return p.conn
}

func TestUnaryHook(t *testing.T) {
	a := assertions.New(t)

	// The test database needs to drop, and then, recreated.
	s, err := sql.Open(databaseURI)
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to create a store instance")
	}

	err = s.Clean()
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to clean store")
	}
	defer s.Close()

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	isConfig := commands.DefaultConfig.IS
	isConfig.DatabaseURI = databaseURI
	is, err := identityserver.New(c, isConfig)
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to create an Identity Server instance")
	}

	err = is.Init(identityserver.InitialData{
		Settings: identityserver.DefaultSettings,
		Admin: identityserver.InitialAdminData{
			UserID:   "admin",
			Email:    "admin@localhost",
			Password: "admin",
		},
		Console: identityserver.InitialConsoleData{
			ClientSecret: "console",
			RedirectURI:  "http://foo.bar",
		},
	})
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to initialize Identity Server")
	}

	srv := rpcserver.New(context.Background())
	is.RegisterServices(srv.Server)

	conn, err := rpcserver.StartLoopback(context.Background(), srv.Server)
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to start gRPC services of the Identity Server")
	}
	defer srv.Stop()

	// Feed database with an organization plus an organization API key.
	org := &ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{
			OrganizationID: "org",
		},
	}
	err = s.Organizations.Create(org)
	a.So(err, should.BeNil)

	orgKeyStr, err := auth.GenerateOrganizationAPIKey("issuer")
	a.So(err, should.BeNil)
	orgKey := ttnpb.APIKey{
		Key:    orgKeyStr,
		Name:   "Key",
		Rights: []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS},
	}
	err = s.Organizations.SaveAPIKey(org.OrganizationIdentifiers, orgKey)
	a.So(err, should.BeNil)

	// Feed database with an application plus an application API key.
	app := &ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
			ApplicationID: "app",
		},
	}
	err = s.Applications.Create(app)
	a.So(err, should.BeNil)

	appKeyStr, err := auth.GenerateApplicationAPIKey("issuer")
	a.So(err, should.BeNil)
	appKey := ttnpb.APIKey{
		Key:    appKeyStr,
		Name:   "Key",
		Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS},
	}
	err = s.Applications.SaveAPIKey(app.ApplicationIdentifiers, appKey)
	a.So(err, should.BeNil)

	// Feed database with a gateway plus a gateway API key.
	gtw := &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{
			GatewayID: "gtw",
			EUI:       new(types.EUI64),
		},
	}
	err = s.Gateways.Create(gtw)
	a.So(err, should.BeNil)

	gtwKeyStr, err := auth.GenerateGatewayAPIKey("issuer")
	a.So(err, should.BeNil)
	gtwKey := ttnpb.APIKey{
		Key:    gtwKeyStr,
		Name:   "Key",
		Rights: []ttnpb.Right{ttnpb.RIGHT_GATEWAY_DELETE},
	}
	err = s.Gateways.SaveAPIKey(gtw.GatewayIdentifiers, gtwKey)
	a.So(err, should.BeNil)

	hook, err := New(
		context.Background(),
		&isProvider{
			conn: conn,
		},
		Config{
			TTL:           0,
			AllowInsecure: true,
		},
	)
	if !a.So(err, should.BeNil) {
		t.Fatal("Failed to create the hook type")
	}

	for _, tc := range []struct {
		tcName    string
		authValue string
		req       interface{}
		expected  []ttnpb.Right
		errored   bool
	}{
		{
			// Skips the hook because there is no auth value.
			"NoOrganizationAuthValue",
			"",
			&org.OrganizationIdentifiers,
			[]ttnpb.Right{},
			false,
		},
		{
			// Skips the hook because there is no auth value.
			"NoApplicationAuthValue",
			"",
			&app.ApplicationIdentifiers,
			[]ttnpb.Right{},
			false,
		},
		{
			// Skips the hook because there is no auth value.
			"NoGatewayAuthValue",
			"",
			&gtw.GatewayIdentifiers,
			[]ttnpb.Right{},
			false,
		},
		{
			// It fails because the auth value have wrong format and can not be decoded.
			"InvalidOrganizationAuthValue",
			"---",
			&ttnpb.OrganizationIdentifiers{
				OrganizationID: "non-existent",
			},
			[]ttnpb.Right{},
			true,
		},
		{
			// It fails because the auth value have wrong format and can not be decoded.
			"InvalidApplicationAuthValue",
			"---",
			&ttnpb.ApplicationIdentifiers{
				ApplicationID: "non-existent",
			},
			[]ttnpb.Right{},
			true,
		},
		{
			// It fails because the auth value have wrong format and can not be decoded.
			"InvalidGatewayAuthValue",
			"---",
			&ttnpb.GatewayIdentifiers{
				GatewayID: "non-existent",
			},
			[]ttnpb.Right{},
			true,
		},
		{
			// The hook does not make any call because the request message does not implement any interface.
			"NoImplementedInterface",
			appKeyStr,
			nil,
			[]ttnpb.Right{},
			false,
		},
		{
			// Returns not authorized because the API key does not have rights for this application.
			"NotAuthorizedForOrganizationAPIKey",
			orgKeyStr,
			&ttnpb.OrganizationIdentifiers{
				OrganizationID: "random-application",
			},
			[]ttnpb.Right{},
			true,
		},
		{
			// It returns the rights of the application API key.
			"AuthorizedForOrganizationAPIKey",
			orgKeyStr,
			&org.OrganizationIdentifiers,
			orgKey.Rights,
			false,
		},
		{
			// Returns not authorized because the API key does not have rights for this application.
			"NotAuthorizedForApplicationAPIKey",
			appKeyStr,
			&ttnpb.ApplicationIdentifiers{
				ApplicationID: "random-application",
			},
			[]ttnpb.Right{},
			true,
		},
		{
			// It returns the rights of the application API key.
			"AuthorizedForApplicationAPIKey",
			appKeyStr,
			&app.ApplicationIdentifiers,
			appKey.Rights,
			false,
		},
		{
			// Returns not authorized because the API key does not have rights for this gateway.
			"NotAuthorizedForGatewayAPIKey",
			gtwKeyStr,
			&ttnpb.GatewayIdentifiers{
				GatewayID: "random-gtw",
			},
			[]ttnpb.Right{},
			true,
		},
		{
			// It returns the rights of the gateway API key.
			"AuthorizedForGatewayAPIKey",
			gtwKeyStr,
			&gtw.GatewayIdentifiers,
			gtwKey.Rights,
			false,
		},
	} {
		t.Run(tc.tcName, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs("authorization", fmt.Sprintf("Bearer %s", tc.authValue)),
			)

			_, err = hook.UnaryHook()(testHandler(t, tc.expected))(ctx, tc.req)
			if tc.errored {
				a.So(err, should.NotBeNil)
			} else {
				a.So(err, should.BeNil)
			}
		})
	}
}
