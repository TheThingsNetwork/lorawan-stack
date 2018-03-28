// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

/*
Package rights implements a gRPC Unary hook that preload rights in the context.

In order to preload the rights the following steps are taken:

1. The hook looks if the request message implements one of either interfaces:

	type applicationIdentifiersGetters interface {
		GetApplicationID() string
	}

	type gatewayIdentifiersGetters interface {
		GetGatewayID() string
	}

If the message implements both interfaces only applicationIdentifiers is taken
into account.

2. The hook gets a gRPC connection to an Identity Server through the
IdentityServerConnector interface and then calls either the ListApplicationRights
or ListGatewayRights method of the Identity Server using the authorization
value of the original request.

Optionally the hook can set up a TTL cache whenever the Config.TTL value is
different to its zero value.

3. The resulting rights are put in the context.

Lastly, the way to check the rights in the protected Unary method is a matter of:

	import (
		"github.com/TheThingsNetwork/ttn/pkg/auth"
		"github.com/TheThingsNetwork/ttn/pkg/auth/rights"
		"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	)

	func (a *ApplicationServer) MyMethod(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
		if !ttnpb.IncludesRights(rights.FromContext(ctx), ttnpb.RIGHT_APPLICATION_TRAFFIC_READ) {
			return nil, auth.ErrNotAuthorized.New(nil)
		}

		...
		....
	}
*/
package rights
