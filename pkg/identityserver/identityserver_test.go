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
	"context"
	"database/sql"
	"strings"
	"sync"
	"testing"

	"github.com/iancoleman/strcase"
	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	store "go.thethings.network/lorawan-stack/v3/pkg/identityserver/bunstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func rpcCreds(key *ttnpb.APIKey) grpc.CallOption {
	return grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "bearer",
		AuthValue:     key.Key,
		AllowInsecure: true,
	})
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
			Cluster: cluster.Config{
				Keys: []string{
					"11111111111111111111111111111111",
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
	testOptions.isConfig.Network.NSID = &test.DefaultNSID
	testOptions.isConfig.Network.TenantID = "test"
	return testOptions
}

var (
	baseDBName = "ttn_lorawan_is_test"
	baseDSN    = storetest.GetDSN(baseDBName)
	baseDB     = func() *sql.DB {
		baseDB, err := sql.Open("postgres", baseDSN.String())
		if err != nil {
			panic(err)
		}
		return baseDB
	}()
	setupBaseDBOnce sync.Once
	setupBaseDBErr  error
)

func testWithIdentityServer(t *testing.T, f func(*IdentityServer, *grpc.ClientConn), options ...TestOption) {
	testOptions := defaultTestOptions()
	for _, option := range options {
		option(testOptions)
	}

	ctx := test.Context()

	setupBaseDBOnce.Do(func() {
		var db *bun.DB
		db, setupBaseDBErr = store.Open(context.Background(), baseDSN.String())
		if setupBaseDBErr != nil {
			return
		}
		defer db.Close()
		if setupBaseDBErr = store.Initialize(ctx, db, baseDBName); setupBaseDBErr != nil {
			return
		}
		if setupBaseDBErr = store.Migrate(ctx, db); setupBaseDBErr != nil {
			return
		}
	})
	if setupBaseDBErr != nil {
		t.Fatal(setupBaseDBErr)
	}

	testName := t.Name()
	if i := strings.IndexRune(testName, '/'); i != -1 {
		testName = testName[:i]
	}
	schemaName := strcase.ToSnake(testName)
	schemaDSN := baseDSN

	if testOptions.privateDatabase {
		if err := storetest.CreateSchema(baseDB, schemaName); err != nil {
			t.Fatal(err)
		}
		schemaDSN = storetest.GetSchemaDSN(baseDSN, schemaName)
	}

	db, err := store.Open(context.Background(), schemaDSN.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if testOptions.privateDatabase {
		if err = store.Initialize(ctx, db, baseDBName); err != nil {
			t.Fatal(err)
		}
		if err = store.Migrate(ctx, db); err != nil {
			t.Fatal(err)
		}
	} else {
		if err = store.Clear(ctx, db); err != nil {
			t.Fatal(err)
		}
	}

	if testOptions.population != nil {
		st, err := store.NewStore(ctx, db)
		if err != nil {
			t.Fatal(err)
		}
		if err = testOptions.population.Populate(test.Context(), st); err != nil {
			t.Fatal(err)
		}
	}

	c := componenttest.NewComponent(t, testOptions.componentConfig)
	testOptions.isConfig.DatabaseURI = schemaDSN.String()
	is, err := New(c, testOptions.isConfig)
	if err != nil {
		t.Fatal(err)
	}

	componenttest.StartComponent(t, c)

	f(is, is.LoopbackConn())

	is.Close()

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
