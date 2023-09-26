// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package mockis

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type mockISUserRegistry struct {
	ttnpb.UnimplementedUserRegistryServer
	ttnpb.UnimplementedUserAccessServer

	users sync.Map
}

func newUserRegistry() *mockISUserRegistry {
	return &mockISUserRegistry{}
}

func (m *mockISUserRegistry) Create(ctx context.Context, req *ttnpb.CreateUserRequest) (*ttnpb.User, error) {
	m.users.Store(unique.ID(ctx, req.User.Ids), req.User)
	return req.User, nil
}

func (m *mockISUserRegistry) Get(ctx context.Context, req *ttnpb.GetUserRequest) (*ttnpb.User, error) {
	loadedUser, ok := m.users.Load(unique.ID(ctx, req.UserIds))
	if !ok {
		return nil, errNotFound
	}
	usr, ok := loadedUser.(*ttnpb.User)
	if !ok {
		panic("stored value is not of type *ttnpb.User")
	}

	return usr, nil
}
