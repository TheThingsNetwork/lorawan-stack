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

package store

import (
	"context"
	"database/sql"
	"net/url"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver.
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

type testStore struct {
	db *bun.DB

	*applicationStore
	*clientStore
	// deviceStore
	*gatewayStore
	*organizationStore
	*userStore
	// userSessionStore
	// apiKeyStore
	*membershipStore
	// contactInfoStore
	// invitationStore
	// loginTokenStore
	// oauthStore
	// euiStore
	// entitySearch
	// notificationStore

	// authenticationProviderStore
	// externalUserStore
	// tenantStore
}

func (t testStore) Init(ctx context.Context) error {
	return Migrate(ctx, t.db.DB)
}

func (t testStore) Close() error {
	return t.db.Close()
}

func newTestStore(t *testing.T, dsn *url.URL) storetest.Store {
	t.Helper()

	t.Logf("Connecting to %s", dsn.String())
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn.String())))
	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(NewLoggerHook(test.GetLogger(t)))

	baseStore := newStore(db)
	return &testStore{
		db: db,

		applicationStore: newApplicationStore(baseStore),
		clientStore:      newClientStore(baseStore),
		// deviceStore:       deviceStore{baseStore: baseStore},
		gatewayStore:      newGatewayStore(baseStore),
		organizationStore: newOrganizationStore(baseStore),
		userStore:         newUserStore(baseStore),
		// userSessionStore:  userSessionStore{baseStore: baseStore},
		// apiKeyStore:       apiKeyStore{baseStore: baseStore},
		membershipStore: newMembershipStore(baseStore),
		// contactInfoStore:  contactInfoStore{baseStore: baseStore},
		// invitationStore:   invitationStore{baseStore: baseStore},
		// loginTokenStore:   loginTokenStore{baseStore: baseStore},
		// oauthStore:        oauthStore{baseStore: baseStore},
		// euiStore:          euiStore{baseStore: baseStore},
		// entitySearch:      entitySearch{baseStore: baseStore},
		// notificationStore: notificationStore{baseStore: baseStore},

		// authenticationProviderStore: authenticationProviderStore{baseStore: baseStore},
		// externalUserStore:           externalUserStore{baseStore: baseStore},
		// tenantStore:                 tenantStore{baseStore: baseStore},
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
