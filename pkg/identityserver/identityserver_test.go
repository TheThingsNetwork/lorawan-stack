// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	store "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
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

func getIdentityServer(t *testing.T) (*IdentityServer, *grpc.ClientConn) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	setup.Do(func() {
		dbAddress := os.Getenv("SQL_DB_ADDRESS")
		if dbAddress == "" {
			dbAddress = "localhost:5432"
		}
		dbName := os.Getenv("TEST_DATABASE_NAME")
		if dbName == "" {
			dbName = "ttn_lorawan_is_test"
		}
		dbAuth := os.Getenv("SQL_DB_AUTH")
		if dbAuth == "" {
			dbAuth = "root:root"
		}
		dbConnString = fmt.Sprintf("postgresql://%s@%s/%s?sslmode=disable", dbAuth, dbAddress, dbName)
		db, err := store.Open(ctx, dbConnString)
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
		if err = store.Clear(db); err != nil {
			panic(err)
		}
		if err = population.Populate(test.Context(), db); err != nil {
			panic(err)
		}
	})
	c := componenttest.NewComponent(t, &component.Config{ServiceBase: config.ServiceBase{
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
	}})
	conf := &Config{
		DatabaseURI: dbConnString,
	}
	conf.UserRegistration.Enabled = true
	conf.UserRegistration.PasswordRequirements.MinLength = 10
	conf.UserRegistration.PasswordRequirements.MaxLength = 1000
	conf.Email.Templates.Static = map[string][]byte{
		"overridden.subject.txt": []byte("Overridden subject {{.User.Name}}"),
		"overridden.html":        []byte("Overridden HTML {{.User.Name}} {{.User.Email}}"),
		"overridden.txt":         []byte("Overridden text {{.User.Name}} {{.User.Email}}"),
	}
	conf.UserRights.CreateApplications = true
	conf.UserRights.CreateClients = true
	conf.UserRights.CreateGateways = true
	conf.UserRights.CreateOrganizations = true
	conf.AdminRights.All = true
	var euiBlock types.EUI64Prefix
	euiBlock.UnmarshalConfigString("70B3D57ED0000000/36")
	conf.DevEUIBlock.Enabled = true
	conf.DevEUIBlock.Prefix = euiBlock
	conf.DevEUIBlock.ApplicationLimit = 3
	conf.Network.NetID = test.DefaultNetID
	conf.Network.TenantID = "test"
	is, err := New(c, conf)
	if err != nil {
		t.Fatal(err)
	}
	componenttest.StartComponent(t, c)
	return is, is.LoopbackConn()
}

func testWithIdentityServer(t *testing.T, f func(is *IdentityServer, cc *grpc.ClientConn)) {
	f(getIdentityServer(t))
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
