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

package rights

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Rights for the request.
type Rights struct {
	ApplicationRights  map[ttnpb.ApplicationIdentifiers][]ttnpb.Right
	GatewayRights      map[ttnpb.GatewayIdentifiers][]ttnpb.Right
	OrganizationRights map[ttnpb.OrganizationIdentifiers][]ttnpb.Right
}

// IncludesApplicationRights returns whether the given rights are included for the given application.
func (r Rights) IncludesApplicationRights(appID ttnpb.ApplicationIdentifiers, rights ...ttnpb.Right) bool {
	return ttnpb.IncludesRights(r.ApplicationRights[appID], rights...)
}

// IncludesGatewayRights returns whether the given rights are included for the given gateway.
func (r Rights) IncludesGatewayRights(gtwID ttnpb.GatewayIdentifiers, rights ...ttnpb.Right) bool {
	return ttnpb.IncludesRights(r.GatewayRights[gtwID], rights...)
}

// IncludesOrganizationRights returns whether the given rights are included for the given organization.
func (r Rights) IncludesOrganizationRights(orgID ttnpb.OrganizationIdentifiers, rights ...ttnpb.Right) bool {
	return ttnpb.IncludesRights(r.OrganizationRights[orgID], rights...)
}

type rightsKeyType struct{}

var rightsKey rightsKeyType

// FromContext returns the request rights from the context.
func FromContext(ctx context.Context) (Rights, bool) {
	if rights, ok := ctx.Value(rightsKey).(Rights); ok {
		return rights, true
	}
	return Rights{}, false
}

func newContext(ctx context.Context, rights Rights) context.Context {
	return context.WithValue(ctx, rightsKey, rights)
}
