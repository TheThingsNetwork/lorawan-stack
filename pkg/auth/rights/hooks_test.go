// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rights

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Make sure in compile time that the identifiers implements interfaces the hook uses.
var (
	_ applicationIdentifiersGetters = new(ttnpb.ApplicationIdentifiers)
	_ applicationIdentifiersGetters = new(ttnpb.EndDeviceIdentifiers)
	_ gatewayIdentifiersGetters     = new(ttnpb.GatewayIdentifiers)
)
