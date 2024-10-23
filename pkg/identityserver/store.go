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
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	bunstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/bunstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

func (is *IdentityServer) setupStore() error {
	store.SetPaginationDefaults(store.PaginationDefaults{
		DefaultLimit: is.config.Pagination.DefaultLimit,
	})
	return is.setupBunStore()
}

func (is *IdentityServer) setupBunStore() (err error) {
	is.db, err = storeutil.OpenDB(is.Context(), is.config.DatabaseURI)
	if err != nil {
		return err
	}
	bunDB := bun.NewDB(is.db, pgdialect.New())
	if is.LogDebug() {
		bunDB.AddQueryHook(storeutil.NewLoggerHook(log.FromContext(is.Context()).WithField("namespace", "db")))
	}
	bunStore, err := bunstore.NewStore(is.Context(), bunDB)
	if err != nil {
		return err
	}
	is.store = bunStore

	return nil
}
