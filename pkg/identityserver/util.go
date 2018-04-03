// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

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
