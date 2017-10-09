// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package factory

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// DefaultUser implements UserFactory interface.
type DefaultUser struct{}

// BuildUser returns a ttnpb.User.
func (f DefaultUser) BuildUser() types.User {
	return &ttnpb.User{}
}

// DefaultApplication implements ApplicationFactory interface.
type DefaultApplication struct{}

// BuildApplication returns a ttnpb.Application.
func (f DefaultApplication) BuildApplication() types.Application {
	return &ttnpb.Application{}
}

// DefaultGateway implements GatewayFactory interface.
type DefaultGateway struct{}

// BuildGateway returns a ttnpb.Gateway.
func (f DefaultGateway) BuildGateway() types.Gateway {
	return &ttnpb.Gateway{}
}

// DefaultClient implement ClientFactory interface.
type DefaultClient struct{}

// BuildClient returns a ttnpb.Client.
func (f DefaultClient) BuildClient() types.Client {
	return &ttnpb.Client{}
}
