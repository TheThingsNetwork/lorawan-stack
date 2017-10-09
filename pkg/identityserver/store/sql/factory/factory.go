// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package factory

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/types"

// UserFactory is the interface that provides a method to construct the result
// types in the User store.
type UserFactory interface {
	BuildUser() types.User
}

// ApplicationFactory is the interface that provides a method to construct the
// result types in the Application store.
type ApplicationFactory interface {
	BuildApplication() types.Application
}

// GatewayFactory is the interface that provides a method to construct the
// result types in the Gateway store.
type GatewayFactory interface {
	BuildGateway() types.Gateway
}

// ClientFactory is the interface that provides a method to construct the
// result types in the Client store.
type ClientFactory interface {
	BuildClient() types.Client
}
