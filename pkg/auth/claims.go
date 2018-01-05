// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Entity is an enum that defines the valids entity types for claims.
type Entity string

const (
	// EntityUser is an user.
	EntityUser Entity = "user"

	// EntityApplication is an application.
	EntityApplication = "application"

	// EntityGateway is a gateway.
	EntityGateway = "gateway"
)

// String implements fmt.Stringer.
func (e Entity) String() string {
	return string(e)
}

// Claims is the type that represents a claims to do something on the network.
type Claims struct {
	// EntityID is the ID of the entity this claims are intended for.
	EntityID string

	// EntityType is the type of entity this claims are intended for.
	EntityType Entity

	// Source is the source of this claims, either an API key or a token.
	Source string

	// Rights is the list of actions this token has access to.
	Rights []ttnpb.Right
}

// UserID returns the user ID of the user profile this claims are for, or the
// empty string if it is not for a user.
func (c *Claims) UserID() (id string) {
	if c.EntityType == EntityUser {
		id = c.EntityID
	}
	return
}

// ApplicationID returns the application ID  of the application this claims are
// for, or the empty string if it is not for an application.
func (c *Claims) ApplicationID() (id string) {
	if c.EntityType == EntityApplication {
		id = c.EntityID
	}
	return
}

// GatewayID returns the gateway ID of the gateway this claims are for, or the
// empty string if it is not for a gateway.
func (c *Claims) GatewayID() (id string) {
	if c.EntityType == EntityGateway {
		id = c.EntityID
	}
	return
}

// HasRights checks whether or not the provided rights are included in the claims.
// It will only return true if all the provided rights are included in the claims.
func (c *Claims) HasRights(rights ...ttnpb.Right) bool {
	ok := true
	for _, right := range rights {
		ok = ok && c.hasRight(right)
	}

	return ok
}

// hasRight checks whether or not the right is included in this claims.
func (c *Claims) hasRight(right ttnpb.Right) bool {
	for _, r := range c.Rights {
		if r == right {
			return true
		}
	}
	return false
}
