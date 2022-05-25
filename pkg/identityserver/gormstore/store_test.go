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
	"context"
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
	userSessionStore
	apiKeyStore
	membershipStore
	contactInfoStore
	invitationStore
	loginTokenStore
	oauthStore
	euiStore
	entitySearch
	notificationStore
}

func (t testStore) Init(ctx context.Context) error {
	return Migrate(ctx, t.db)
}

func (t testStore) Close() error {
	return t.db.Close()
}

func newTestStore(t *testing.T, dsn *url.URL) storetest.Store {
	t.Helper()

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
		userSessionStore:  userSessionStore{baseStore: &baseStore},
		apiKeyStore:       apiKeyStore{baseStore: &baseStore},
		membershipStore:   membershipStore{baseStore: &baseStore},
		contactInfoStore:  contactInfoStore{baseStore: &baseStore},
		invitationStore:   invitationStore{baseStore: &baseStore},
		loginTokenStore:   loginTokenStore{baseStore: &baseStore},
		oauthStore:        oauthStore{baseStore: &baseStore},
		euiStore:          euiStore{baseStore: &baseStore},
		entitySearch:      entitySearch{baseStore: &baseStore},
		notificationStore: notificationStore{baseStore: &baseStore},
	}
}

func TestApplicationStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestApplicationStoreCRUD(t)
	st.TestApplicationStorePagination(t)
}

func TestClientStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestClientStoreCRUD(t)
	st.TestClientStorePagination(t)
}

func TestEndDeviceStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestEndDeviceStoreCRUD(t)
	st.TestEndDeviceStorePagination(t)
	st.TestEndDeviceBatchUpdate(t)
}

func TestGatewayStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestGatewayStoreCRUD(t)
	st.TestGatewayStorePagination(t)
}

func TestOrganizationStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestOrganizationStoreCRUD(t)
	st.TestOrganizationStorePagination(t)
}

func TestUserStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestUserStoreCRUD(t)
	st.TestUserStorePagination(t)
}

func TestUserSessionStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestUserSessionStore(t)
	st.TestUserSessionStorePagination(t)
}

func TestAPIKeyStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestAPIKeyStoreCRUD(t)
	st.TestAPIKeyStorePagination(t)
}

func TestMembershipStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestMembershipStoreCRUD(t)
	st.TestMembershipStorePagination(t)
}

func TestContactInfoStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestContactInfoStoreCRUD(t)
}

func TestInvitationStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestInvitationStore(t)
	st.TestInvitationStorePagination(t)
}

func TestLoginTokenStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestLoginTokenStore(t)
}

func TestOAuthStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestOAuthStore(t)
	st.TestOAuthStorePagination(t)
}

func TestEUIStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestEUIStore(t)
}

func TestDeletedEntities(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestDeletedEntities(t)
}

func TestEntitySearch(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestEntitySearch(t)
	st.TestEntitySearchPagination(t)
}

func TestNotificationStore(t *testing.T) {
	t.Parallel()

	st := storetest.New(t, newTestStore)
	st.TestNotificationStore(t)
}
