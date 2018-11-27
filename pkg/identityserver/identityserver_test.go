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
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

var (
	setup        sync.Once
	dbConnString string
	population   = store.NewPopulator(10, 42)
)

func init() {
	population.Users[0].Admin = false
	population.Users[0].State = ttnpb.STATE_APPROVED
}

func userCreds() grpc.CallOption {
	for id, apiKeys := range population.APIKeys {
		if id.GetUserIDs().GetUserID() == population.Users[0].GetUserID() {
			return grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "bearer",
				AuthValue:     apiKeys[0].Key,
				AllowInsecure: true,
			})
		}
	}
	return nil
}

func getIdentityServer(t *testing.T) (*IdentityServer, *grpc.ClientConn) {
	setup.Do(func() {
		dbName := os.Getenv("TEST_DB_NAME")
		if dbName == "" {
			dbName = "is_integration_test"
		}
		dbConnString = fmt.Sprintf("postgresql://root@localhost:26257/%s?sslmode=disable", dbName)
		db, err := gorm.Open("postgres", dbConnString)
		if err != nil {
			panic(err)
		}
		defer db.Close()
		if err = db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", dbName)).Error; err != nil {
			panic(err)
		}
		store.AutoMigrate(db)
		if err = store.Clear(db); err != nil {
			panic(err)
		}
		if err = population.Populate(test.Context(), db); err != nil {
			panic(err)
		}
	})
	c := component.MustNew(test.GetLogger(t), &component.Config{ServiceBase: config.ServiceBase{
		Base: config.Base{
			Log: config.Log{
				Level: log.DebugLevel,
			},
		},
	}})
	is, err := New(c, &Config{
		DatabaseURI: dbConnString,
	})
	if err != nil {
		panic(err)
	}
	err = is.Start()
	if err != nil {
		panic(err)
	}
	return is, is.LoopbackConn()
}

func testWithIdentityServer(t *testing.T, f func(is *IdentityServer, cc *grpc.ClientConn)) {
	f(getIdentityServer(t))
}
