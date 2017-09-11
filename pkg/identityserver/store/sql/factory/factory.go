// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package factory

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/types"

// UserFactory is the interface that provides a method to construct the result
// types in the User store.
type UserFactory interface {
	User() types.User
}

// ApplicationFactory is the interface that provides a method to construct the
// result types in the Application store.
type ApplicationFactory interface {
	Application() types.Application
}

// GatewayFactory is the interface that provides a method to construct the
// result types in the Gateway store.
type GatewayFactory interface {
	Gateway() types.Gateway
}

// ComponentFactory is the interface that provides a method to construct the
// result types in the Component store.
type ComponentFactory interface {
	Component() types.Component
}

// ClientFactory is the interface that provides a method to construct the
// result types in the Client store.
type ClientFactory interface {
	Client() types.Client
}
