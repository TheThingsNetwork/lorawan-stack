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
	"net/url"
	"testing"

	_ "github.com/lib/pq" // PostgreSQL driver.
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

type testStore struct {
	*Store
}

func (t testStore) Init(ctx context.Context) error {
	return Migrate(ctx, t.Store.baseStore.baseDB.DB.DB)
}

func (t testStore) Close() error {
	return t.Store.baseStore.baseDB.DB.Close()
}

func newTestStore(t *testing.T, dsn *url.URL) storetest.Store {
	t.Helper()

	ctx := test.Context()
	if dl, ok := t.Deadline(); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(test.Context(), dl)
		defer cancel()
	}

	t.Logf("Connecting to %s", dsn.String())
	sqldb, err := store.OpenDB(ctx, dsn.String())
	if err != nil {
		t.Fatal(err)
	}
	db := bun.NewDB(sqldb, pgdialect.New())

	db.AddQueryHook(NewLoggerHook(test.GetLogger(t)))

	store, err := NewStore(ctx, db)
	if err != nil {
		t.Fatal(err)
	}

	return &testStore{Store: store}
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
	st.TestEndDeviceCAC(t)
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
