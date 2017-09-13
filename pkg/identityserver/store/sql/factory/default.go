// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package factory

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/types"

// DefaultUser implements UserFactory interface.
type DefaultUser struct{}

// User returns a DefaultUser.
func (f DefaultUser) User() types.User {
	return &types.DefaultUser{}
}

// DefaultApplication implements ApplicationFactory interface.
type DefaultApplication struct{}

// Application returns a DefaultApplication.
func (f DefaultApplication) Application() types.Application {
	return &types.DefaultApplication{}
}

// DefaultGateway implements GatewayFactory interface.
type DefaultGateway struct{}

// Gateway returns a DefaultGateway.
func (f DefaultGateway) Gateway() types.Gateway {
	return &types.DefaultGateway{}
}

// DefaultClient implement ClientFactory interface.
type DefaultClient struct{}

// Client returns a DefaultClient.
func (f DefaultClient) Client() types.Client {
	return &types.DefaultClient{}
}
