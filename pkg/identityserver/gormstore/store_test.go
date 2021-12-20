// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"net/url"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq" // PostgreSQL driver.
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
)

type testStore struct {
	db *gorm.DB
	applicationStore
	clientStore
	deviceStore
	gatewayStore
	organizationStore
	userStore
}

func (t testStore) Init() error {
	if err := AutoMigrate(t.db).Error; err != nil {
		return err
	}
	return nil
}

func newTestStore(t *testing.T, dsn url.URL) storetest.Store {
	t.Logf("Connecting to %s", dsn.String())
	db, err := gorm.Open("postgres", dsn.String())
	if err != nil {
		t.Fatal(err)
	}
	testDB := db.Debug()
	baseStore := baseStore{DB: testDB}
	return &testStore{
		db:                db,
		applicationStore:  applicationStore{baseStore: &baseStore},
		clientStore:       clientStore{baseStore: &baseStore},
		deviceStore:       deviceStore{baseStore: &baseStore},
		gatewayStore:      gatewayStore{baseStore: &baseStore},
		organizationStore: organizationStore{baseStore: &baseStore},
		userStore:         userStore{baseStore: &baseStore},
	}
}

func TestApplicationStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestApplicationStoreCRUD(t)
}

func TestClientStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestClientStoreCRUD(t)
}

func TestEndDeviceStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestEndDeviceStoreCRUD(t)
}

func TestGatewayStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestGatewayStoreCRUD(t)
}

func TestOrganizationStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestOrganizationStoreCRUD(t)
}

func TestUserStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestUserStoreCRUD(t)
}
