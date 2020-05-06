// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package rightsutil

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EventIsVisible returns whether evt is visible given rights in the context.
func EventIsVisible(ctx context.Context, evt events.Event) (bool, error) {
	visibility := evt.Visibility().Union(ttnpb.RightsFrom(
		ttnpb.RIGHT_APPLICATION_ALL,
		ttnpb.RIGHT_CLIENT_ALL,
		ttnpb.RIGHT_GATEWAY_ALL,
		ttnpb.RIGHT_ORGANIZATION_ALL,
		ttnpb.RIGHT_USER_ALL,
	))
	for _, entityIDs := range evt.Identifiers() {
		switch ids := entityIDs.Identifiers().(type) {
		case *ttnpb.ApplicationIdentifiers:
			rights, err := rights.ListApplication(ctx, *ids)
			if err != nil {
				return false, err
			}
			if len(rights.Implied().Intersect(visibility).GetRights()) > 0 {
				return true, nil
			}
		case *ttnpb.ClientIdentifiers:
			rights, err := rights.ListClient(ctx, *ids)
			if err != nil {
				return false, err
			}
			if len(rights.Implied().Intersect(visibility).GetRights()) > 0 {
				return true, nil
			}
		case *ttnpb.EndDeviceIdentifiers:
			rights, err := rights.ListApplication(ctx, ids.ApplicationIdentifiers)
			if err != nil {
				return false, err
			}
			if len(rights.Implied().Intersect(visibility).GetRights()) > 0 {
				return true, nil
			}
		case *ttnpb.GatewayIdentifiers:
			rights, err := rights.ListGateway(ctx, *ids)
			if err != nil {
				return false, err
			}
			if len(rights.Implied().Intersect(visibility).GetRights()) > 0 {
				return true, nil
			}
		case *ttnpb.OrganizationIdentifiers:
			rights, err := rights.ListOrganization(ctx, *ids)
			if err != nil {
				return false, err
			}
			if len(rights.Implied().Intersect(visibility).GetRights()) > 0 {
				return true, nil
			}
		case *ttnpb.UserIdentifiers:
			rights, err := rights.ListUser(ctx, *ids)
			if err != nil {
				return false, err
			}
			if len(rights.Implied().Intersect(visibility).GetRights()) > 0 {
				return true, nil
			}
		}
	}
	return false, nil
}
