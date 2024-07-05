// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package managed

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// requireProfileRights requires the caller to have the required rights to manage the profile.
// This is tested by checking whether the caller is authorized to create gateways.
func requireProfileRights(ctx context.Context, ids *ttnpb.OrganizationOrUserIdentifiers) error {
	if usrIDs := ids.GetUserIds(); usrIDs != nil {
		if err := rights.RequireUser(ctx, usrIDs, ttnpb.Right_RIGHT_USER_GATEWAYS_CREATE); err != nil {
			return err
		}
	} else if orgIDs := ids.GetOrganizationIds(); orgIDs != nil {
		if err := rights.RequireOrganization(ctx, orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_GATEWAYS_CREATE); err != nil {
			return err
		}
	}
	return nil
}

func group(ids *ttnpb.OrganizationOrUserIdentifiers) string {
	if usrIDs := ids.GetUserIds(); usrIDs != nil {
		return usrIDs.UserId
	}
	if orgIDs := ids.GetOrganizationIds(); orgIDs != nil {
		return orgIDs.OrganizationId
	}
	return ""
}

var errNoProfileName = errors.DefineInvalidArgument("no_profile_name", "no profile name set")
