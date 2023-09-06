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

type mockISOrganizationRegistry struct {
	ttnpb.UnimplementedOrganizationRegistryServer
	ttnpb.UnimplementedOrganizationAccessServer

	orgs sync.Map
}

func newOrganizationRegistry() *mockISOrganizationRegistry {
	return &mockISOrganizationRegistry{}
}

func (m *mockISOrganizationRegistry) Create(
	ctx context.Context, req *ttnpb.CreateOrganizationRequest,
) (*ttnpb.Organization, error) {
	m.orgs.Store(unique.ID(ctx, req.Organization.Ids), req.Organization)
	return req.Organization, nil
}

func (m *mockISOrganizationRegistry) Get(
	ctx context.Context, req *ttnpb.GetOrganizationRequest,
) (*ttnpb.Organization, error) {
	loadedOrganization, ok := m.orgs.Load(unique.ID(ctx, req.OrganizationIds))
	if !ok {
		return nil, errNotFound
	}
	org, ok := loadedOrganization.(*ttnpb.Organization)
	if !ok {
		panic("stored value is not of type *ttnpb.Organization")
	}

	return org, nil
}
