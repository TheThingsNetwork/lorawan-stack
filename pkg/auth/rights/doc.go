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

/*
Package rights implements a gRPC Unary hook that preload rights in the context.

In order to preload the rights the following steps are taken:

1. The hook looks if the request message implements one of either interfaces:

	type applicationIdentifiersGetters interface {
		GetApplicationID() string
	}

	type gatewayIdentifiersGetters interface {
		GetGatewayID() string
		GetEUI() *types.EUI64
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
		"go.thethings.network/lorawan-stack/pkg/auth"
		"go.thethings.network/lorawan-stack/pkg/auth/rights"
		"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
