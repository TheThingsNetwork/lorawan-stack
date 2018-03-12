// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// ErrNoNetworkServerFound is returned if no network server was found for a passed DevAddr.
var ErrNoNetworkServerFound = &errors.ErrDescriptor{
	MessageFormat:  "No network server found for DevAddr `{dev_addr}`",
	SafeAttributes: []string{"dev_addr"},
	Code:           1,
	Type:           errors.NotFound,
}

func init() {
	ErrNoNetworkServerFound.Register()
}
