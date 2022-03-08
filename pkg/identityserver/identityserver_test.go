// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"database/sql"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/iancoleman/strcase"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	store "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

var (
	setup        sync.Once
	dbConnString string
	population   = store.NewPopulator(16, 42)
)

var (
	userIndex                                               int
	newUser, newUserIdx                                     = getTestUser()
	rejectedUser, rejectedUserIdx                           = getTestUser()
	defaultUser, defaultUserIdx                             = getTestUser()
	suspendedUser, suspendedUserIdx                         = getTestUser()
	adminUser, adminUserIdx                                 = getTestUser()
	userTestUser, userTestUserIdx                           = getTestUser()
	collaboratorUser, collaboratorUserIdx                   = getTestUser()
	applicationAccessUser, applicationAccessUserIdx         = getTestUser()
	appAccessCollaboratorUser, appAccessCollaboratorUserIdx = getTestUser()
	clientAccessUser, clientAccessUserIdx                   = getTestUser()
	gatewayAccessUser, gatewayAccessUserIdx                 = getTestUser()
	gtwAccessCollaboratorUser, gtwAccessCollaboratorUserIdx = getTestUser()
	organizationAccessUser, organizationAccessUserIdx       = getTestUser()
	orgAccessCollaboratorUser, orgAccessCollaboratorUserIdx = getTestUser()
	userAccessUser, userAccessUserIdx                       = getTestUser()
	paginationUser, paginationUserIdx                       = getTestUser()
)

var now = time.Now()

func init() {
	newUser.Admin = false
	newUser.PrimaryEmailAddressValidatedAt = nil
	newUser.State = ttnpb.State_STATE_REQUESTED

	rejectedUser.Admin = false
	rejectedUser.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
	rejectedUser.State = ttnpb.State_STATE_REJECTED

	defaultUser.Admin = false
	defaultUser.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
	defaultUser.State = ttnpb.State_STATE_APPROVED

	defaultUser.TemporaryPassword = ""
	defaultUser.TemporaryPasswordCreatedAt = nil
	defaultUser.TemporaryPasswordExpiresAt = nil

	userTestUser.Admin = false
	userTestUser.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
	userTestUser.State = ttnpb.State_STATE_APPROVED

	userTestUser.TemporaryPassword = ""
	userTestUser.TemporaryPasswordCreatedAt = nil
	userTestUser.TemporaryPasswordExpiresAt = nil

	for id, apiKeys := range population.APIKeys {
		if id.GetUserIds().GetUserId() == defaultUser.GetIds().GetUserId() || id.GetUserIds().GetUserId() == userTestUser.GetIds().GetUserId() {
			expiredTime := time.Now().Add(-1 * time.Hour)
			population.APIKeys[id] = append(
				apiKeys,
				&ttnpb.APIKey{
					Name:   "key without rights",
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_SEND_INVITES},
				},
				&ttnpb.APIKey{
					Name:      "expired key",
					Rights:    []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
					ExpiresAt: ttnpb.ProtoTimePtr(expiredTime),
				},
			)
		}
	}

	suspendedUser.Admin = false
	suspendedUser.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
	suspendedUser.State = ttnpb.State_STATE_SUSPENDED

	adminUser.Admin = true
	adminUser.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
	adminUser.State = ttnpb.State_STATE_APPROVED

	paginationUser.Admin = false
	paginationUser.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTimePtr(now)
	paginationUser.State = ttnpb.State_STATE_APPROVED
}

func getTestUser() (*ttnpb.User, int) {
	defer func() { userIndex++ }()

	return population.Users[userIndex], userIndex
}

func userCreds(idx int, preferredNames ...string) grpc.CallOption {
	for id, apiKeys := range population.APIKeys {
		if id.GetUserIds().GetUserId() == population.Users[idx].GetIds().GetUserId() {
			selectedIdx := 0
			if len(preferredNames) == 0 {
				preferredNames = []string{"default key"}
			}
		findPreferred:
			for _, name := range preferredNames {
				for i, apiKey := range apiKeys {
					if apiKey.Name == name {
						selectedIdx = i
						break findPreferred
					}
				}
			}
			return grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "bearer",
				AuthValue:     apiKeys[selectedIdx].Key,
				AllowInsecure: true,
			})
		}
	}
	return nil
}

func userAPIKeys(userID *ttnpb.UserIdentifiers) ttnpb.APIKeys {
	for id, apiKeys := range population.APIKeys {
		if id.GetUserIds().GetUserId() == userID.GetUserId() {
			return ttnpb.APIKeys{
				ApiKeys: apiKeys,
			}
		}
	}

	return ttnpb.APIKeys{
		ApiKeys: []*ttnpb.APIKey{},
	}
}

func applicationAPIKeys(applicationID *ttnpb.ApplicationIdentifiers) ttnpb.APIKeys {
	for id, apiKeys := range population.APIKeys {
		if id.GetApplicationIds().GetApplicationId() == applicationID.GetApplicationId() {
			return ttnpb.APIKeys{
				ApiKeys: apiKeys,
			}
		}
	}

	return ttnpb.APIKeys{
		ApiKeys: []*ttnpb.APIKey{},
	}
}

func gatewayAPIKeys(gatewayID *ttnpb.GatewayIdentifiers) ttnpb.APIKeys {
	for id, apiKeys := range population.APIKeys {
		if id.GetGatewayIds().GetGatewayId() == gatewayID.GetGatewayId() {
			return ttnpb.APIKeys{
				ApiKeys: apiKeys,
			}
		}
	}

	return ttnpb.APIKeys{
		ApiKeys: []*ttnpb.APIKey{},
	}
}

func organizationAPIKeys(organizationID *ttnpb.OrganizationIdentifiers) ttnpb.APIKeys {
	for id, apiKeys := range population.APIKeys {
		if id.GetOrganizationIds().GetOrganizationId() == organizationID.GetOrganizationId() {
			return ttnpb.APIKeys{
				ApiKeys: apiKeys,
			}
		}
	}

	return ttnpb.APIKeys{
		ApiKeys: []*ttnpb.APIKey{},
	}
}

func userApplications(userID *ttnpb.UserIdentifiers) ttnpb.Applications {
	applications := []*ttnpb.Application{}
	for _, app := range population.Applications {
		for id, collaborators := range population.Memberships {
			if app.IDString() == id.IDString() {
				for _, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						applications = append(applications, app)
					}
				}
			}
		}
	}

	return ttnpb.Applications{
		Applications: applications,
	}
}

func userClients(userID *ttnpb.UserIdentifiers) ttnpb.Clients {
	clients := []*ttnpb.Client{}
	for _, client := range population.Clients {
		for id, collaborators := range population.Memberships {
			if client.IDString() == id.IDString() {
				for _, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						clients = append(clients, client)
					}
				}
			}
		}
	}

	return ttnpb.Clients{
		Clients: clients,
	}
}

func userGateways(userID *ttnpb.UserIdentifiers) ttnpb.Gateways {
	gateways := []*ttnpb.Gateway{}
	for _, gateway := range population.Gateways {
		for id, collaborators := range population.Memberships {
			if gateway.IDString() == id.IDString() {
				for _, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						gateways = append(gateways, gateway)
					}
				}
			}
		}
	}

	return ttnpb.Gateways{
		Gateways: gateways,
	}
}

func userOrganizations(userID *ttnpb.UserIdentifiers) ttnpb.Organizations {
	organizations := []*ttnpb.Organization{}
	for _, organization := range population.Organizations {
		for id, collaborators := range population.Memberships {
			if organization.IDString() == id.IDString() {
				for _, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						organizations = append(organizations, organization)
					}
				}
			}
		}
	}

	return ttnpb.Organizations{
		Organizations: organizations,
	}
}

type testOptions struct {
	privateDatabase bool
	population      *storetest.Population
	componentConfig *component.Config
	isConfig        *Config
}

type TestOption func(*testOptions)

func withPrivateTestDatabase(p *storetest.Population) TestOption {
	return func(opts *testOptions) {
		opts.privateDatabase = true
		opts.population = p
	}
}

func defaultTestOptions() *testOptions {
	testOptions := &testOptions{}
	testOptions.componentConfig = &component.Config{
		ServiceBase: config.ServiceBase{
			Base: config.Base{
				Log: config.Log{
					Level: log.DebugLevel,
				},
			},
			KeyVault: config.KeyVault{
				Provider: "static",
				Static: map[string][]byte{
					"is-test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
			},
		},
	}
	testOptions.isConfig = &Config{}
	testOptions.isConfig.UserRegistration.Enabled = true
	testOptions.isConfig.UserRegistration.PasswordRequirements.MinLength = 10
	testOptions.isConfig.UserRegistration.PasswordRequirements.MaxLength = 1000
	testOptions.isConfig.Email.Templates.Static = map[string][]byte{
		"overridden.subject.txt": []byte("Overridden subject {{.User.Name}}"),
		"overridden.html":        []byte("Overridden HTML {{.User.Name}} {{.User.Email}}"),
		"overridden.txt":         []byte("Overridden text {{.User.Name}} {{.User.Email}}"),
	}
	testOptions.isConfig.UserRights.CreateApplications = true
	testOptions.isConfig.UserRights.CreateClients = true
	testOptions.isConfig.UserRights.CreateGateways = true
	testOptions.isConfig.UserRights.CreateOrganizations = true
	testOptions.isConfig.AdminRights.All = true
	testOptions.isConfig.DevEUIBlock.Enabled = true
	testOptions.isConfig.DevEUIBlock.ApplicationLimit = 3
	testOptions.isConfig.Network.NetID = test.DefaultNetID
	testOptions.isConfig.Network.TenantID = "test"
	return testOptions
}

func testWithIdentityServer(t *testing.T, f func(*IdentityServer, *grpc.ClientConn), options ...TestOption) {
	testOptions := defaultTestOptions()
	for _, option := range options {
		option(testOptions)
	}

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	var err error
	baseDSN := storetest.GetDSN("ttn_lorawan_is_test")
	testName := t.Name()
	if i := strings.IndexRune(testName, '/'); i != -1 {
		testName = testName[:i]
	}
	schemaName := strcase.ToSnake(testName)
	schemaDSN := baseDSN
	setup := &setup

	var baseDB *sql.DB
	if testOptions.privateDatabase {
		baseDB, err = sql.Open("postgres", baseDSN.String())
		if err != nil {
			t.Fatal(err)
		}
		defer baseDB.Close()
		if err = storetest.CreateSchema(baseDB, schemaName); err != nil {
			t.Fatal(err)
		}
		schemaDSN = storetest.GetSchemaDSN(baseDSN, schemaName)
		setup = &sync.Once{}
	}

	setup.Do(func() {
		db, err := store.Open(ctx, schemaDSN.String())
		if err != nil {
			panic(err)
		}
		defer db.Close()
		if err = store.Initialize(db); err != nil {
			panic(err)
		}
		if err = store.AutoMigrate(db).Error; err != nil {
			panic(err)
		}
		if !testOptions.privateDatabase {
			if err = store.Clear(db); err != nil {
				panic(err)
			}
			if err = population.Populate(test.Context(), db); err != nil {
				panic(err)
			}
		}
		if testOptions.population != nil {
			if err = testOptions.population.Populate(test.Context(), store.NewCombinedStore(db)); err != nil {
				t.Fatal(err)
			}
		}
	})

	c := componenttest.NewComponent(t, testOptions.componentConfig)
	testOptions.isConfig.DatabaseURI = schemaDSN.String()
	is, err := New(c, testOptions.isConfig)
	if err != nil {
		t.Fatal(err)
	}

	componenttest.StartComponent(t, c)

	f(is, is.LoopbackConn())

	if testOptions.privateDatabase {
		if t.Failed() {
			t.Logf("Keeping database %q", schemaDSN)
		} else {
			if err = storetest.DropSchema(baseDB, schemaName); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func reverse(s string) string {
	b := []byte(s)
	first := 0
	last := len(b) - 1
	for first < last {
		b[first], b[last] = b[last], b[first]
		first++
		last--
	}
	return string(b)
}
