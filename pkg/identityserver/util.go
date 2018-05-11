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

package identityserver

import (
	"time"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func organizationOrUserIDsUserIDs(ids ttnpb.UserIdentifiers) ttnpb.OrganizationOrUserIdentifiers {
	return ttnpb.OrganizationOrUserIdentifiers{
		ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{
			UserID: &ids,
		},
	}
}

func organizationOrUserIDsOrganizationIDs(ids ttnpb.OrganizationIdentifiers) ttnpb.OrganizationOrUserIdentifiers {
	return ttnpb.OrganizationOrUserIdentifiers{
		ID: &ttnpb.OrganizationOrUserIdentifiers_OrganizationID{
			OrganizationID: &ids,
		},
	}
}

func timeValue(t time.Time) *time.Time {
	return &t
}
